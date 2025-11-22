package ssmconfig

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

type cacheEntry struct {
	values *atomic.Pointer[map[string]string]
	once   sync.Once
}

type Loader struct {
	ssmClient       *ssm.Client
	strict          bool
	logger          func(format string, args ...interface{})
	cache           sync.Map // map[string]*cacheEntry
	useStrongTyping bool     // If true, use strongly-typed conversion; if false, prefer JSON decoding
	configFiles     []string // List of config file paths (YAML, JSON, TOML)
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

// WithStrongTyping controls whether to use strongly-typed conversion or prefer JSON decoding.
// If true (default), uses strongly-typed conversion for simple types (int, string, bool, etc.).
// If false, prefers JSON decoding for all types. The json:"true" tag on fields always takes precedence.
func WithStrongTyping(useStrongTyping bool) LoaderOption {
	return func(l *Loader) {
		l.useStrongTyping = useStrongTyping
	}
}

// WithConfigFiles adds configuration file paths to load from.
// Files are loaded using Viper in order, with later files overriding earlier ones.
// Supported formats: .yaml, .yml, .json, .toml
// Priority: ENV > File > SSM
func WithConfigFiles(filePaths ...string) LoaderOption {
	return func(l *Loader) {
		l.configFiles = append(l.configFiles, filePaths...)
	}
}

func NewLoader(ctx context.Context, opts ...LoaderOption) (*Loader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	loader := &Loader{
		ssmClient:       ssm.NewFromConfig(cfg),
		strict:          false,
		logger:          nil,
		useStrongTyping: true, // Default to strongly-typed conversion
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
	// Load from SSM Parameter Store
	ssmValues, err := loader.loadByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Load from config files using Viper (if configured)
	fileValues, err := loader.loadFromFiles()
	if err != nil {
		return nil, fmt.Errorf("loading config files: %w", err)
	}

	// Merge: Start with SSM values, then overlay file values
	// File values override SSM values (but ENV will override both in mapToStruct)
	mergedValues := make(map[string]string)
	// First add SSM values
	for k, v := range ssmValues {
		mergedValues[k] = v
	}
	// Then overlay file values (file values take precedence over SSM)
	for k, v := range fileValues {
		mergedValues[k] = v
	}

	var result T
	if err := mapToStruct(mergedValues, &result, loader.strict, loader.logger, loader.useStrongTyping); err != nil {
		return nil, fmt.Errorf("mapping to struct: %w", err)
	}

	return &result, nil
}

// loadFromFiles loads configuration from YAML, JSON, and TOML files using Viper.
// Returns a flat map[string]string compatible with SSM parameter format.
func (l *Loader) loadFromFiles() (map[string]string, error) {
	if len(l.configFiles) == 0 {
		return make(map[string]string), nil
	}

	v := viper.New()
	firstFile := true
	
	// Load each file
	for _, filePath := range l.configFiles {
		if filePath == "" {
			continue
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		// Set file path
		v.SetConfigFile(filePath)
		
		if firstFile {
			// Read first config file
			if err := v.ReadInConfig(); err != nil {
				if l.logger != nil {
					l.logger("WARNING: Failed to read config file %s: %v", filePath, err)
				}
				continue
			}
			firstFile = false
		} else {
			// Merge subsequent files (later files override earlier ones)
			if err := v.MergeInConfig(); err != nil {
				if l.logger != nil {
					l.logger("WARNING: Failed to merge config file %s: %v", filePath, err)
				}
				continue
			}
		}
	}

	// Convert Viper's nested config to flat map[string]string
	// Viper uses dot notation (e.g., "database.host"), which matches our SSM format
	result := make(map[string]string)
	
	// Get all keys from Viper and convert values to strings
	keys := v.AllKeys()
	for _, key := range keys {
		// Convert Viper's dot notation to SSM slash notation
		ssmKey := strings.ReplaceAll(key, ".", "/")
		
		// Get value and convert to string
		value := v.Get(key)
		if value != nil {
			// Convert to string representation
			result[ssmKey] = fmt.Sprintf("%v", value)
		}
	}

	return result, nil
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
