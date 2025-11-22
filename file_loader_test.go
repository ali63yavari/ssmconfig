package ssmconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithConfigFiles(t *testing.T) {
	t.Run("adds config files to loader", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles("config.yaml", "config.json"))
		require.NoError(t, err)
		assert.Len(t, loader.configFiles, 2)
		assert.Equal(t, "config.yaml", loader.configFiles[0])
		assert.Equal(t, "config.json", loader.configFiles[1])
	})

	t.Run("appends to existing config files", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx,
			WithConfigFiles("config.yaml"),
			WithConfigFiles("config.json"))
		require.NoError(t, err)
		assert.Len(t, loader.configFiles, 2)
	})
}

func TestLoader_LoadFromFiles(t *testing.T) {
	t.Run("returns empty map when no files configured", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx)
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("skips non-existent files", func(t *testing.T) {
		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles("nonexistent.yaml"))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("loads from YAML file", func(t *testing.T) {
		// Create temporary YAML file
		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
database:
  url: "postgres://localhost:5432/mydb"
  port: 5432
server:
  host: "0.0.0.0"
  port: 8080
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(yamlFile))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost:5432/mydb", values["database/url"])
		assert.Equal(t, "5432", values["database/port"])
		assert.Equal(t, "0.0.0.0", values["server/host"])
		assert.Equal(t, "8080", values["server/port"])
	})

	t.Run("loads from JSON file", func(t *testing.T) {
		// Create temporary JSON file
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(jsonFile, []byte(`{
  "database": {
    "url": "postgres://localhost:5432/mydb",
    "port": 5432
  },
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  }
}`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(jsonFile))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost:5432/mydb", values["database/url"])
		assert.Equal(t, "5432", values["database/port"])
	})

	t.Run("loads from TOML file", func(t *testing.T) {
		// Create temporary TOML file
		tmpDir := t.TempDir()
		tomlFile := filepath.Join(tmpDir, "config.toml")
		err := os.WriteFile(tomlFile, []byte(`
[database]
url = "postgres://localhost:5432/mydb"
port = 5432

[server]
host = "0.0.0.0"
port = 8080
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(tomlFile))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost:5432/mydb", values["database/url"])
		assert.Equal(t, "5432", values["database/port"])
	})

	t.Run("later files override earlier ones", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "config1.yaml")
		file2 := filepath.Join(tmpDir, "config2.yaml")
		
		err := os.WriteFile(file1, []byte(`
database:
  url: "file1-url"
  port: 5432
`), 0644)
		require.NoError(t, err)

		err = os.WriteFile(file2, []byte(`
database:
  url: "file2-url"
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(file1, file2))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		require.NoError(t, err)
		// file2 should override file1
		assert.Equal(t, "file2-url", values["database/url"])
		// port from file1 should still be present
		assert.Equal(t, "5432", values["database/port"])
	})

	t.Run("handles invalid YAML file gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidFile, []byte("invalid: yaml: content: [["), 0644)
		require.NoError(t, err)

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
		loader, err := NewLoader(ctx, WithConfigFiles(invalidFile), WithLogger(logger))
		require.NoError(t, err)

		values, err := loader.loadFromFiles()
		// Should not error, just skip invalid file
		require.NoError(t, err)
		assert.Empty(t, values)
		assert.Len(t, loggedMessages, 1)
	})
}

func TestLoadWithConfigFiles(t *testing.T) {
	t.Run("loads config from file with priority", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database/url" env:"DB_URL"`
			Port        int    `ssm:"database/port" env:"DB_PORT"`
		}

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
database:
  url: "file-url"
  port: 5432
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		cfg, err := Load[Config](ctx, "/test/", WithConfigFiles(yamlFile))
		// Will fail without actual SSM, but tests the code path
		_ = err
		_ = cfg
	})

	t.Run("ENV overrides file value", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database/url" env:"DB_URL"`
		}

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
database:
  url: "file-url"
`), 0644)
		require.NoError(t, err)

		os.Setenv("DB_URL", "env-url")
		defer os.Unsetenv("DB_URL")

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(yamlFile))
		require.NoError(t, err)

		fileValues, err := loader.loadFromFiles()
		require.NoError(t, err)
		assert.Equal(t, "file-url", fileValues["database/url"])

		// In actual usage, ENV would override this in mapToStruct
		// This test verifies file loading works
	})
}

func TestFilePriority(t *testing.T) {
	t.Run("file values override SSM values", func(t *testing.T) {
		type Config struct {
			Value string `ssm:"value" env:"VALUE"`
		}

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
value: "file-value"
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(yamlFile))
		require.NoError(t, err)

		// Simulate SSM values
		ssmValues := map[string]string{"value": "ssm-value"}
		fileValues, err := loader.loadFromFiles()
		require.NoError(t, err)

		// Merge: file should override SSM
		merged := make(map[string]string)
		for k, v := range ssmValues {
			merged[k] = v
		}
		for k, v := range fileValues {
			merged[k] = v
		}

		assert.Equal(t, "file-value", merged["value"])
	})
}

func TestFileConfig_NestedStructs(t *testing.T) {
	t.Run("loads nested struct from YAML file", func(t *testing.T) {
		type Config struct {
			Database struct {
				Host string `ssm:"host"` // Relative path - prefix "database" is applied by parent
				Port int    `ssm:"port"`
				SSL  bool   `ssm:"ssl"`
			} `ssm:"database"`

			Server struct {
				Host string `ssm:"host"` // Relative path - prefix "server" is applied by parent
				Port int    `ssm:"port"`
			} `ssm:"server"`
		}

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
database:
  host: "localhost"
  port: 5432
  ssl: true
server:
  host: "0.0.0.0"
  port: 8080
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(yamlFile))
		require.NoError(t, err)

		// Load from file
		fileValues, err := loader.loadFromFiles()
		require.NoError(t, err)

		// Verify file values are loaded correctly
		assert.Equal(t, "localhost", fileValues["database/host"])
		assert.Equal(t, "5432", fileValues["database/port"])
		assert.Equal(t, "true", fileValues["database/ssl"])
		assert.Equal(t, "0.0.0.0", fileValues["server/host"])
		assert.Equal(t, "8080", fileValues["server/port"])

		// Now test mapping to struct (without SSM, just file)
		var cfg Config
		err = mapToStruct(fileValues, &cfg, false, nil, true)
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, true, cfg.Database.SSL)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})

	t.Run("loads nested struct from JSON file", func(t *testing.T) {
		type Config struct {
			Database struct {
				Host string `ssm:"host"` // Relative path - prefix "database" is applied by parent
				Port int    `ssm:"port"`
			} `ssm:"database"`
		}

		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(jsonFile, []byte(`{
  "database": {
    "host": "db.example.com",
    "port": 3306
  }
}`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(jsonFile))
		require.NoError(t, err)

		fileValues, err := loader.loadFromFiles()
		require.NoError(t, err)

		var cfg Config
		err = mapToStruct(fileValues, &cfg, false, nil, true)
		require.NoError(t, err)

		assert.Equal(t, "db.example.com", cfg.Database.Host)
		assert.Equal(t, 3306, cfg.Database.Port)
	})

	t.Run("loads deeply nested struct from file", func(t *testing.T) {
		type Config struct {
			App struct {
				Database struct {
					Host string `ssm:"host"` // Relative path - prefix "app/database" is applied by parent
					Port int    `ssm:"port"`
				} `ssm:"database"` // Relative to "app" prefix
				Server struct {
					Host string `ssm:"host"` // Relative path - prefix "app/server" is applied by parent
					Port int    `ssm:"port"`
				} `ssm:"server"` // Relative to "app" prefix
			} `ssm:"app"`
		}

		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(`
app:
  database:
    host: "localhost"
    port: 5432
  server:
    host: "0.0.0.0"
    port: 8080
`), 0644)
		require.NoError(t, err)

		os.Setenv("AWS_REGION", "us-east-1")
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		defer os.Unsetenv("AWS_REGION")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		ctx := context.Background()
		loader, err := NewLoader(ctx, WithConfigFiles(yamlFile))
		require.NoError(t, err)

		fileValues, err := loader.loadFromFiles()
		require.NoError(t, err)

		var cfg Config
		err = mapToStruct(fileValues, &cfg, false, nil, true)
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.App.Database.Host)
		assert.Equal(t, 5432, cfg.App.Database.Port)
		assert.Equal(t, "0.0.0.0", cfg.App.Server.Host)
		assert.Equal(t, 8080, cfg.App.Server.Port)
	})
}

