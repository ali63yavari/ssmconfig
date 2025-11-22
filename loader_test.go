package ssmconfig

import (
	"context"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with default options", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)
		assert.NotNil(t, loader)
		assert.False(t, loader.strict)
		assert.Nil(t, loader.logger)
		assert.True(t, loader.useStrongTyping)
	})

	t.Run("creates loader with strict mode", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrictMode(true))
		require.NoError(t, err)
		assert.True(t, loader.strict)
	})

	t.Run("creates loader with logger", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		logged := false
		logger := func(format string, args ...interface{}) {
			logged = true
		}
		loader, err := NewLoader(ctx, WithLogger(logger))
		require.NoError(t, err)
		assert.NotNil(t, loader.logger)
		loader.logger("test")
		assert.True(t, logged)
	})

	t.Run("creates loader with strong typing disabled", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrongTyping(false))
		require.NoError(t, err)
		assert.False(t, loader.useStrongTyping)
	})
}

func TestLoad(t *testing.T) {
	t.Run("loads basic config", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database_url" required:"true"`
			Port        int    `ssm:"port"`
			Debug       bool   `ssm:"debug"`
		}

		// Set up environment to mock SSM
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()

		// Note: This test requires actual AWS credentials or a localstack setup
		// For unit testing, we'd typically use a mock SSM client
		// This is a placeholder that shows the structure
		_, err := Load[Config](ctx, "/test/")
		// In real tests, we'd assert on the result
		_ = err
	})
}

func TestLoader_InvalidateCache(t *testing.T) {
	t.Run("invalidates specific prefix", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		// Create a cache entry
		entry := &cacheEntry{
			values: &atomic.Pointer[map[string]string]{},
		}
		values := map[string]string{"key": "value"}
		entry.values.Store(&values)
		loader.cache.Store("/test/", entry)

		// Verify cache exists and has values
		entryPtr, ok := loader.cache.Load("/test/")
		assert.True(t, ok)
		assert.NotNil(t, entryPtr)
		cachedEntry := entryPtr.(*cacheEntry)
		cachedValues := cachedEntry.values.Load()
		assert.NotNil(t, cachedValues)
		assert.Equal(t, "value", (*cachedValues)["key"])

		// Invalidate
		loader.InvalidateCache("/test/")

		// Verify cache entry is reset (values cleared, but entry still exists)
		entryPtr, ok = loader.cache.Load("/test/")
		assert.True(t, ok, "Cache entry should still exist after invalidation")
		cachedEntry = entryPtr.(*cacheEntry)
		cachedValues = cachedEntry.values.Load()
		assert.Nil(t, cachedValues, "Cache values should be cleared after invalidation")
	})

	t.Run("invalidates all cache", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		// Create multiple cache entries with values
		entry1 := &cacheEntry{values: &atomic.Pointer[map[string]string]{}}
		entry2 := &cacheEntry{values: &atomic.Pointer[map[string]string]{}}
		values1 := map[string]string{"key1": "value1"}
		values2 := map[string]string{"key2": "value2"}
		entry1.values.Store(&values1)
		entry2.values.Store(&values2)
		loader.cache.Store("/test1/", entry1)
		loader.cache.Store("/test2/", entry2)

		// Verify entries exist and have values
		entryPtr1, ok1 := loader.cache.Load("/test1/")
		entryPtr2, ok2 := loader.cache.Load("/test2/")
		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.NotNil(t, entryPtr1.(*cacheEntry).values.Load())
		assert.NotNil(t, entryPtr2.(*cacheEntry).values.Load())

		// Invalidate all (empty string means all)
		loader.InvalidateCache("")

		// Verify all values are cleared (entries still exist but values are nil)
		entryPtr1, ok1 = loader.cache.Load("/test1/")
		entryPtr2, ok2 = loader.cache.Load("/test2/")
		assert.True(t, ok1, "Cache entry /test1/ should still exist")
		assert.True(t, ok2, "Cache entry /test2/ should still exist")
		assert.Nil(t, entryPtr1.(*cacheEntry).values.Load(), "Cache values for /test1/ should be cleared")
		assert.Nil(t, entryPtr2.(*cacheEntry).values.Load(), "Cache values for /test2/ should be cleared")
	})

	t.Run("invalidates non-existent prefix", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		// Invalidate non-existent prefix should not panic
		loader.InvalidateCache("/nonexistent/")
	})
}

func TestWithStrictMode(t *testing.T) {
	t.Run("sets strict mode", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrictMode(true))
		require.NoError(t, err)
		assert.True(t, loader.strict)
	})

	t.Run("disables strict mode", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrictMode(false))
		require.NoError(t, err)
		assert.False(t, loader.strict)
	})
}

func TestWithLogger(t *testing.T) {
	t.Run("sets custom logger", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		var loggedMessages []string
		logger := func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		}

		loader, err := NewLoader(ctx, WithLogger(logger))
		require.NoError(t, err)
		assert.NotNil(t, loader.logger)

		loader.logger("test message")
		assert.Len(t, loggedMessages, 1)
		assert.Equal(t, "test message", loggedMessages[0])
	})
}

func TestWithStrongTyping(t *testing.T) {
	t.Run("enables strong typing", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrongTyping(true))
		require.NoError(t, err)
		assert.True(t, loader.useStrongTyping)
	})

	t.Run("disables strong typing", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithStrongTyping(false))
		require.NoError(t, err)
		assert.False(t, loader.useStrongTyping)
	})
}

func TestLoadWithLoader(t *testing.T) {
	t.Run("loads config with existing loader", func(t *testing.T) {
		type Config struct {
			Value string `ssm:"value"`
		}

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		// This will fail without actual SSM, but tests the code path
		_, err = LoadWithLoader[Config](loader, ctx, "/test/")
		// Error is expected without actual SSM setup
		_ = err
	})
}
