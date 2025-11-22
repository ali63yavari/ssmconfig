package ssmconfig

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type cacheEntry struct {
	values *atomic.Pointer[map[string]string]
	once   sync.Once
}

type Loader struct {
	ssmClient *ssm.Client
	strict    bool
	logger    func(format string, args ...interface{})
	cache     sync.Map // map[string]*cacheEntry
}

type LoaderOption func(*Loader)

// WithStrictMode enables strict mode where missing required fields will cause a panic.
func WithStrictMode(strict bool) LoaderOption {
	return func(l *Loader) {
		l.strict = strict
	}
}

// WithLogger sets a custom logger function for logging missing required fields.
// This allows integration with logging libraries like Sentry, zap, logrus, etc.
// The logger function receives a format string and variadic arguments.
func WithLogger(logger func(format string, args ...interface{})) LoaderOption {
	return func(l *Loader) {
		l.logger = logger
	}
}

func NewLoader(ctx context.Context, opts ...LoaderOption) (*Loader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	loader := &Loader{
		ssmClient: ssm.NewFromConfig(cfg),
		strict:    false,
		logger:    nil,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// Load loads configuration from AWS SSM Parameter Store and returns a typed struct.
// Environment variables (specified via "env" tags) will override SSM parameter values.
func Load[T any](ctx context.Context, prefix string, opts ...LoaderOption) (*T, error) {
	loader, err := NewLoader(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return LoadWithLoader[T](loader, ctx, prefix)
}

// LoadWithLoader loads configuration using an existing Loader instance.
func LoadWithLoader[T any](loader *Loader, ctx context.Context, prefix string) (*T, error) {
	values, err := loader.loadByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var result T
	if err := mapToStruct(values, &result, loader.strict, loader.logger); err != nil {
		return nil, fmt.Errorf("mapping to struct: %w", err)
	}

	return &result, nil
}

func (l *Loader) loadByPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	return l.loadByPrefixWithCache(ctx, prefix, true)
}

// loadByPrefixWithCache loads parameters with optional cache bypass.
func (l *Loader) loadByPrefixWithCache(ctx context.Context, prefix string, useCache bool) (map[string]string, error) {
	// If not using cache, load fresh and update cache
	if !useCache {
		result, err := l.loadFromSSM(ctx, prefix)
		if err != nil {
			return nil, err
		}

		// Update cache with fresh values
		entryPtr, _ := l.cache.Load(prefix)
		if entryPtr != nil {
			entry := entryPtr.(*cacheEntry)
			// Make a copy for the cache
			cachedValues := make(map[string]string, len(result))
			for k, v := range result {
				cachedValues[k] = v
			}
			entry.values.Store(&cachedValues)
		}

		// Return a copy
		resultCopy := make(map[string]string, len(result))
		for k, v := range result {
			resultCopy[k] = v
		}
		return resultCopy, nil
	}

	// Use cache - get or create cache entry for this prefix
	entryPtr, _ := l.cache.Load(prefix)
	var entry *cacheEntry

	if entryPtr == nil {
		// Create new cache entry with atomic pointer for values
		newEntry := &cacheEntry{
			values: &atomic.Pointer[map[string]string]{},
		}
		actual, _ := l.cache.LoadOrStore(prefix, newEntry)
		entry = actual.(*cacheEntry)
	} else {
		entry = entryPtr.(*cacheEntry)
	}

	// Check if already cached
	cachedValues := entry.values.Load()
	if cachedValues != nil {
		// Return a copy to avoid race conditions
		result := make(map[string]string, len(*cachedValues))
		for k, v := range *cachedValues {
			result[k] = v
		}
		return result, nil
	}

	// Cache miss - load from SSM using sync.Once to ensure only one load per prefix
	var result map[string]string
	var loadErr error

	entry.once.Do(func() {
		result, loadErr = l.loadFromSSM(ctx, prefix)
		if loadErr == nil {
			// Make a copy for the cache
			cachedValues := make(map[string]string, len(result))
			for k, v := range result {
				cachedValues[k] = v
			}
			// Store in cache using atomic pointer
			entry.values.Store(&cachedValues)
		}
	})

	if loadErr != nil {
		return nil, loadErr
	}

	// If we loaded successfully, result is already set
	// Otherwise, try to get from cache (another goroutine might have loaded it)
	if result == nil {
		cachedValues := entry.values.Load()
		if cachedValues != nil {
			result = make(map[string]string, len(*cachedValues))
			for k, v := range *cachedValues {
				result[k] = v
			}
			return result, nil
		}
		return nil, fmt.Errorf("failed to load parameters for prefix: %s", prefix)
	}

	// Return a copy
	resultCopy := make(map[string]string, len(result))
	for k, v := range result {
		resultCopy[k] = v
	}

	return resultCopy, nil
}

// loadFromSSM performs the actual SSM API call to load parameters.
func (l *Loader) loadFromSSM(ctx context.Context, prefix string) (map[string]string, error) {
	out := make(map[string]string)

	var nextToken *string
	for {
		resp, err := l.ssmClient.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           &prefix,
			Recursive:      ToPointerValue(true),
			WithDecryption: ToPointerValue(true),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching parameters: %w", err)
		}

		for _, p := range resp.Parameters {
			name := strings.TrimPrefix(*p.Name, prefix)
			// Remove leading slash if present
			name = strings.TrimPrefix(name, "/")
			out[name] = *p.Value
		}

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return out, nil
}

// InvalidateCache clears the cache for a specific prefix.
// If prefix is empty, clears all cached entries.
// After invalidation, the next call to loadByPrefix will reload from SSM.
func (l *Loader) InvalidateCache(prefix string) {
	if prefix == "" {
		// Clear all cache entries
		l.cache.Range(func(key, value interface{}) bool {
			entry := value.(*cacheEntry)
			entry.values.Store(nil)
			// Reset sync.Once by creating a new entry
			newEntry := &cacheEntry{
				values: &atomic.Pointer[map[string]string]{},
			}
			l.cache.Store(key, newEntry)
			return true
		})
	} else {
		// Clear specific prefix
		if entryPtr, ok := l.cache.Load(prefix); ok {
			entry := entryPtr.(*cacheEntry)
			entry.values.Store(nil)
			// Reset sync.Once by creating a new entry
			newEntry := &cacheEntry{
				values: &atomic.Pointer[map[string]string]{},
			}
			l.cache.Store(prefix, newEntry)
		}
	}
}
