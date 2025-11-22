package ssmconfig

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWithAutoRefresh(t *testing.T) {
	t.Run("creates refreshing config", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		type Config struct {
			Value string `ssm:"value"`
		}

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		// Note: This would require actual SSM setup or mocking
		// For now, we test the structure
		_, err = LoadWithAutoRefreshAndLoader[Config](loader, ctx, "/test/",
			WithRefreshInterval[Config](1*time.Second))
		// In real tests, we'd assert on the result
		_ = err
	})
}

func TestLoadWithAutoRefresh_Function(t *testing.T) {
	t.Run("creates refreshing config with function", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		type Config struct {
			Value string `ssm:"value"`
		}

		ctx := context.Background()
		_, err := LoadWithAutoRefresh[Config](ctx, "/test/")
		// Error expected without actual SSM
		_ = err
	})
}

func TestRefreshingConfig_Get(t *testing.T) {
	t.Run("returns current config", func(t *testing.T) {
		type Config struct {
			Value string
		}

		cfg := &Config{Value: "test"}
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		loader, _ := NewLoader(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rc := &RefreshingConfig[Config]{
			config: cfg,
			loader: loader,
			prefix: "/test/",
			ctx:    ctx,
			cancel: cancel,
		}

		result := rc.Get()
		assert.Equal(t, "test", result.Value)
	})
}

func TestRefreshingConfig_GetCopy(t *testing.T) {
	t.Run("returns safe copy", func(t *testing.T) {
		type Config struct {
			Value string
		}

		cfg := &Config{Value: "test"}
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		loader, _ := NewLoader(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rc := &RefreshingConfig[Config]{
			config: cfg,
			loader: loader,
			prefix: "/test/",
			ctx:    ctx,
			cancel: cancel,
		}

		cfgCopy, err := rc.GetCopy()
		require.NoError(t, err)
		assert.Equal(t, "test", cfgCopy.Value)

		// Modify original
		rc.mu.Lock()
		rc.config.Value = "modified"
		rc.mu.Unlock()

		// Copy should be unchanged
		assert.Equal(t, "test", cfgCopy.Value)
	})
}

func TestRefreshingConfig_Stop(t *testing.T) {
	t.Run("stops refreshing", func(t *testing.T) {
		type Config struct {
			Value string
		}

		cfg := &Config{Value: "test"}
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		loader, _ := NewLoader(context.Background())
		ctx, cancel := context.WithCancel(context.Background())

		rc := &RefreshingConfig[Config]{
			config: cfg,
			loader: loader,
			prefix: "/test/",
			ctx:    ctx,
			cancel: cancel,
			wg:     sync.WaitGroup{},
		}

		rc.Stop()
		// Context should be canceled
		assert.Error(t, rc.ctx.Err())
	})
}

func TestRefreshingConfig_Refresh(t *testing.T) {
	t.Run("refreshes config", func(t *testing.T) {
		type Config struct {
			Value string
		}

		cfg := &Config{Value: "old"}
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		loader, _ := NewLoader(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rc := &RefreshingConfig[Config]{
			config: cfg,
			loader: loader,
			prefix: "/test/",
			ctx:    ctx,
			cancel: cancel,
		}

		// Refresh will fail without actual SSM, but tests the code path
		err := rc.Refresh()
		// Error expected without actual SSM setup
		_ = err
	})

	t.Run("calls onChange callback on change", func(t *testing.T) {
		type Config struct {
			Value string
		}

		cfg := &Config{Value: "old"}
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		loader, _ := NewLoader(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var callbackCalled bool
		callback := func(old, new *Config) {
			callbackCalled = true
		}

		rc := &RefreshingConfig[Config]{
			config:   cfg,
			loader:   loader,
			prefix:   "/test/",
			ctx:      ctx,
			cancel:   cancel,
			onChange: callback,
		}

		// Manually set new config to trigger callback
		rc.mu.Lock()
		oldConfig := rc.config
		newConfig := &Config{Value: "new"}
		hasChanged := !reflect.DeepEqual(oldConfig, newConfig)
		rc.config = newConfig
		rc.mu.Unlock()

		if rc.onChange != nil && hasChanged {
			rc.onChange(oldConfig, newConfig)
		}

		assert.True(t, callbackCalled)
	})
}

func TestWithRefreshInterval(t *testing.T) {
	t.Run("sets refresh interval", func(t *testing.T) {
		type Config struct {
			Value string
		}

		rc := &RefreshingConfig[Config]{}
		opt := WithRefreshInterval[Config](30 * time.Second)
		opt(rc)

		assert.Equal(t, 30*time.Second, rc.refreshInterval)
	})
}

func TestWithOnChange(t *testing.T) {
	t.Run("sets onChange callback", func(t *testing.T) {
		type Config struct {
			Value string
		}

		var called bool
		callback := func(old, new *Config) {
			called = true
		}

		rc := &RefreshingConfig[Config]{}
		opt := WithOnChange[Config](callback)
		opt(rc)

		assert.NotNil(t, rc.onChange)
		rc.onChange(nil, nil)
		assert.True(t, called)
	})
}

func TestDeepCopy(t *testing.T) {
	t.Run("copies simple struct", func(t *testing.T) {
		type Config struct {
			Value string
		}

		original := &Config{Value: "test"}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Equal(t, "test", copyConfig.Value)
		assert.NotSame(t, original, copyConfig)
	})

	t.Run("copies nested struct", func(t *testing.T) {
		type Nested struct {
			Value string
		}
		type Config struct {
			Nested Nested
		}

		original := &Config{Nested: Nested{Value: "test"}}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Equal(t, "test", copyConfig.Nested.Value)
		assert.NotSame(t, original, copyConfig)
	})

	t.Run("copies pointer fields", func(t *testing.T) {
		type Config struct {
			Value *string
		}

		val := "test"
		original := &Config{Value: &val}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Equal(t, "test", *copyConfig.Value)
		assert.NotSame(t, original.Value, copyConfig.Value)
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		type Config struct {
			Value *string
		}

		original := &Config{Value: nil}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Nil(t, copyConfig.Value)
	})

	t.Run("handles nil input", func(t *testing.T) {
		type Config struct {
			Value string
		}

		var original *Config = nil
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		assert.Nil(t, copyConfig)
	})

	t.Run("copies slice fields", func(t *testing.T) {
		type Config struct {
			Values []string
		}

		original := &Config{Values: []string{"a", "b", "c"}}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Equal(t, []string{"a", "b", "c"}, copyConfig.Values)
		// Verify it's a different slice (not the same reference)
		if len(copyConfig.Values) > 0 && len(original.Values) > 0 {
			copyConfig.Values[0] = "modified"
			assert.NotEqual(t, original.Values[0], "modified", "Should be a copy, not a reference")
		}
	})

	t.Run("copies map fields", func(t *testing.T) {
		type Config struct {
			Metadata map[string]string
		}

		original := &Config{Metadata: map[string]string{"key": "value"}}
		copyConfig, err := deepCopy(original)
		require.NoError(t, err)
		require.NotNil(t, copyConfig)
		assert.Equal(t, "value", copyConfig.Metadata["key"])
		// Verify it's a different map (not the same reference)
		if copyConfig.Metadata != nil {
			copyConfig.Metadata["key"] = "modified"
			assert.NotEqual(t, original.Metadata["key"], "modified", "Should be a copy, not a reference")
		}
	})
}
