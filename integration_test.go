// +build integration

package ssmconfig

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require actual AWS credentials and SSM parameters
// Run with: go test -tags=integration

func TestIntegration_BasicLoad(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping integration test: AWS_REGION not set")
	}

	type Config struct {
		DatabaseURL string `ssm:"database_url" required:"true"`
		Port        int    `ssm:"port"`
		Debug       bool   `ssm:"debug"`
	}

	ctx := context.Background()
	cfg, err := Load[Config](ctx, "/test/")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	assert.NotNil(t, cfg)
}

func TestIntegration_EnvironmentOverride(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping integration test: AWS_REGION not set")
	}

	type Config struct {
		DatabaseURL string `ssm:"database_url" env:"DB_URL"`
	}

	os.Setenv("DB_URL", "env-override")
	defer os.Unsetenv("DB_URL")

	ctx := context.Background()
	cfg, err := Load[Config](ctx, "/test/")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	assert.Equal(t, "env-override", cfg.DatabaseURL)
}

func TestIntegration_NestedStructs(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping integration test: AWS_REGION not set")
	}

	type DatabaseConfig struct {
		Host string `ssm:"host"`
		Port int    `ssm:"port"`
	}

	type Config struct {
		Database DatabaseConfig `ssm:"database"`
	}

	ctx := context.Background()
	cfg, err := Load[Config](ctx, "/test/")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	require.NotNil(t, cfg)
}

func TestIntegration_JSONDecoding(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping integration test: AWS_REGION not set")
	}

	type Config struct {
		Database struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `ssm:"database" json:"true"`
	}

	ctx := context.Background()
	cfg, err := Load[Config](ctx, "/test/")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	require.NotNil(t, cfg)
}

func TestIntegration_AutoRefresh(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("Skipping integration test: AWS_REGION not set")
	}

	type Config struct {
		Value string `ssm:"value"`
	}

	ctx := context.Background()
	loader, err := NewLoader(ctx)
	require.NoError(t, err)

	refreshingConfig, err := LoadWithAutoRefreshAndLoader[Config](loader, ctx, "/test/",
		WithRefreshInterval[Config](5*time.Second))
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer refreshingConfig.Stop()

	cfg := refreshingConfig.Get()
	assert.NotNil(t, cfg)
}

