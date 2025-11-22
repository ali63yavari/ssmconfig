package ssmconfig

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// RefreshingConfig holds a configuration that automatically refreshes from Parameter Store.
type RefreshingConfig[T any] struct {
	mu              sync.RWMutex
	config          *T
	loader          *Loader
	prefix          string
	refreshInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	onChange        func(oldConfig, newConfig *T)
}

// RefreshingConfigOption configures a RefreshingConfig.
type RefreshingConfigOption[T any] func(*RefreshingConfig[T])

// WithRefreshInterval sets the interval for auto-refreshing the configuration.
// Default is 5 minutes if not specified.
func WithRefreshInterval[T any](interval time.Duration) RefreshingConfigOption[T] {
	return func(rc *RefreshingConfig[T]) {
		rc.refreshInterval = interval
	}
}

// WithOnChange sets a callback function that is called when the configuration changes.
func WithOnChange[T any](callback func(oldConfig, newConfig *T)) RefreshingConfigOption[T] {
	return func(rc *RefreshingConfig[T]) {
		rc.onChange = callback
	}
}

// LoadWithAutoRefresh loads configuration and starts auto-refreshing it periodically.
func LoadWithAutoRefresh[T any](
	ctx context.Context, prefix string, opts ...LoaderOption) (*RefreshingConfig[T], error) {
	loader, err := NewLoader(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return LoadWithAutoRefreshAndLoader[T](loader, ctx, prefix)
}

// LoadWithAutoRefreshAndLoader loads configuration with auto-refresh using an existing Loader.
func LoadWithAutoRefreshAndLoader[T any](
	loader *Loader, ctx context.Context, prefix string,
	opts ...RefreshingConfigOption[T]) (*RefreshingConfig[T], error) {
	// Initial load
	config, err := LoadWithLoader[T](loader, ctx, prefix)
	if err != nil {
		return nil, err
	}

	refreshCtx, cancel := context.WithCancel(ctx)

	rc := &RefreshingConfig[T]{
		config:          config,
		loader:          loader,
		prefix:          prefix,
		refreshInterval: 5 * time.Minute, // Default 5 minutes
		ctx:             refreshCtx,
		cancel:          cancel,
	}

	// Apply options
	for _, opt := range opts {
		opt(rc)
	}

	// Start auto-refresh
	rc.start()

	return rc, nil
}

// Get returns a thread-safe copy of the current configuration.
// The returned pointer points to the same underlying config, so modifications
// should be avoided. For safe modifications, use GetCopy.
func (rc *RefreshingConfig[T]) Get() *T {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.config
}

// GetCopy returns a deep copy of the current configuration.
// This is safe to modify without affecting the original.
func (rc *RefreshingConfig[T]) GetCopy() (*T, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return deepCopy(rc.config)
}

// deepCopy creates a deep copy of a struct using reflection.
func deepCopy[T any](src *T) (*T, error) {
	if src == nil {
		return nil, nil
	}

	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.New(srcVal.Type())

	if err := copyValue(srcVal, dstVal.Elem()); err != nil {
		return nil, err
	}

	result, ok := dstVal.Interface().(*T)
	if !ok {
		return nil, fmt.Errorf("failed to convert to type %T", result)
	}
	return result, nil
}

//nolint:gocyclo,funlen // Complex function due to multiple reflect.Kind cases and deep copying logic
func copyValue(src, dst reflect.Value) error {
	switch src.Kind() {
	case reflect.Invalid:
		return fmt.Errorf("invalid source value")
	case reflect.Bool:
		dst.SetBool(src.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetInt(src.Int())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.SetUint(src.Uint())
		return nil
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(src.Float())
		return nil
	case reflect.String:
		dst.SetString(src.String())
		return nil
	case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported kind for copying: %v", src.Kind())
	case reflect.Ptr:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.New(src.Elem().Type()))
		return copyValue(src.Elem(), dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return nil
		}
		originalValue := src.Elem()
		copiedValue := reflect.New(originalValue.Type()).Elem()
		if err := copyValue(originalValue, copiedValue); err != nil {
			return err
		}
		dst.Set(copiedValue)

	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if err := copyValue(src.Field(i), dst.Field(i)); err != nil {
				return err
			}
		}

	case reflect.Slice:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			if err := copyValue(src.Index(i), dst.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Map:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copiedValue := reflect.New(originalValue.Type()).Elem()
			if err := copyValue(originalValue, copiedValue); err != nil {
				return err
			}
			copiedKey := reflect.New(key.Type()).Elem()
			if err := copyValue(key, copiedKey); err != nil {
				return err
			}
			dst.SetMapIndex(copiedKey, copiedValue)
		}

	default:
		dst.Set(src)
	}

	return nil
}

// Refresh manually triggers a refresh of the configuration.
// This bypasses the cache to ensure fresh values are loaded from SSM.
func (rc *RefreshingConfig[T]) Refresh() error {
	// Invalidate cache first to ensure we get fresh values
	rc.loader.InvalidateCache(rc.prefix)

	newConfig, err := LoadWithLoader[T](rc.loader, rc.ctx, rc.prefix)
	if err != nil {
		return err
	}

	rc.mu.Lock()
	oldConfig := rc.config
	hasChanged := !reflect.DeepEqual(oldConfig, newConfig)
	rc.config = newConfig
	rc.mu.Unlock()

	// Notify of change if callback is set and config actually changed
	if rc.onChange != nil && hasChanged {
		rc.onChange(oldConfig, newConfig)
	}

	return nil
}

// Stop stops the auto-refresh goroutine.
func (rc *RefreshingConfig[T]) Stop() {
	rc.cancel()
	rc.wg.Wait()
}

// start begins the auto-refresh goroutine.
func (rc *RefreshingConfig[T]) start() {
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		ticker := time.NewTicker(rc.refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-rc.ctx.Done():
				return
			case <-ticker.C:
				if err := rc.Refresh(); err != nil && rc.loader.logger != nil {
					rc.loader.logger("Error refreshing config: %v", err)
				}
			}
		}
	}()
}
