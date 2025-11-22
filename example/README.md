# SSM Config Examples

This directory contains comprehensive examples demonstrating various ways to use the `ssmconfig` module.

## Examples

### 1. `basic.go` - Basic Usage
Demonstrates the simplest way to load configuration from AWS SSM Parameter Store.

**Key Features:**
- Simple struct with basic types
- Environment variable overrides
- Required fields

**Run:**
```bash
go run example/basic.go
```

### 2. `nested_structs.go` - Nested Structures
Shows how to work with nested configuration structs.

**Key Features:**
- Multiple levels of nesting
- Automatic prefix handling
- Required nested structs

**Run:**
```bash
go run example/nested_structs.go
```

### 3. `json_decoding.go` - JSON Decoding
Demonstrates decoding complex data structures from JSON strings in SSM.

**Key Features:**
- JSON-decoded nested structs
- JSON arrays (slices)
- JSON maps
- Using `json:"true"` tag

**Run:**
```bash
go run example/json_decoding.go
```

### 4. `required_fields.go` - Required Field Validation
Shows different modes of required field validation.

**Key Features:**
- Strict mode (panics on missing required fields)
- Non-strict mode (logs warnings)
- Custom logger integration

**Run:**
```bash
go run example/required_fields.go
```

### 5. `custom_validators.go` - Custom Validators
Demonstrates registering and using custom validators.

**Key Features:**
- Built-in validators (email, url, minlen, maxlen, min, max)
- Custom validators
- Parameterized validators
- Multiple validators per field

**Run:**
```bash
go run example/custom_validators.go
```

### 6. `auto_refresh.go` - Auto-Refreshing Config
Shows how to use auto-refreshing configuration.

**Key Features:**
- Periodic auto-refresh
- Change notifications
- Thread-safe access
- Manual refresh
- Safe copying

**Run:**
```bash
go run example/auto_refresh.go
```

### 7. `viper_integration.go` - Viper Integration
Demonstrates integrating with Viper configuration library.

**Key Features:**
- Reading SSM config into Viper
- Using ViperRemoteProvider
- Key format conversion (slashes to dots)

**Prerequisites:**
```bash
go get github.com/spf13/viper
```

**Run:**
```bash
go run example/viper_integration.go
```

**Note:** This example requires Viper as an external dependency. The example file includes a build tag to prevent it from being built with other examples.

### 8. `edge_cases.go` - Edge Cases
Covers edge cases and advanced scenarios.

**Key Features:**
- Pointer to struct fields
- Reusing loader instances
- Cache invalidation
- Strong typing vs JSON decoding
- Mixed configurations

**Run:**
```bash
go run example/edge_cases.go
```

### 9. `advanced_nested.go` - Advanced Nested Structures
Shows complex nested structures with mixed approaches.

**Key Features:**
- Deep nesting (3+ levels)
- Mixed JSON and SSM parameter mapping
- Validators on nested fields
- Complex data types

**Run:**
```bash
go run example/advanced_nested.go
```

### 10. `file_config.go` - File-Based Configuration
Demonstrates loading configuration from YAML, JSON, and TOML files using Viper.

**Key Features:**
- Load from YAML, JSON, and TOML files
- Multiple files support (later files override earlier ones)
- Priority: ENV > File > SSM
- Automatic format detection
- Nested structure support

**Run:**
```bash
go run example/file_config.go
```

**Note:** This example requires Viper as a dependency (already included in the module).

## Prerequisites

1. AWS credentials configured (via AWS CLI, environment variables, or IAM role)
2. SSM parameters created in AWS Parameter Store
3. Appropriate IAM permissions to read SSM parameters

## Setting Up Test Parameters

You can create test parameters using AWS CLI:

```bash
# Basic parameters
aws ssm put-parameter --name "/myapp/database_url" --value "postgres://localhost:5432/mydb" --type "String"
aws ssm put-parameter --name "/myapp/port" --value "8080" --type "String"
aws ssm put-parameter --name "/myapp/debug" --value "true" --type "String"

# Nested parameters
aws ssm put-parameter --name "/myapp/database/host" --value "localhost" --type "String"
aws ssm put-parameter --name "/myapp/database/port" --value "5432" --type "String"
aws ssm put-parameter --name "/myapp/server/host" --value "0.0.0.0" --type "String"
aws ssm put-parameter --name "/myapp/server/port" --value "8080" --type "String"

# JSON parameters
aws ssm put-parameter --name "/myapp/database" --value '{"host":"localhost","port":5432,"ssl":true}' --type "String"
aws ssm put-parameter --name "/myapp/servers" --value '[{"name":"api","port":8080},{"name":"web","port":80}]' --type "String"
```

## Common Patterns

### Pattern 1: Simple Config with Environment Overrides
```go
type Config struct {
    DatabaseURL string `ssm:"database_url" env:"DB_URL" required:"true"`
    Port        int    `ssm:"port" env:"PORT"`
}

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
```

### Pattern 2: Nested Config with Validation
```go
type Config struct {
    Database struct {
        Host string `ssm:"host" validate:"minlen:3"`
        Port int    `ssm:"port" validate:"min:1,max:65535"`
    } `ssm:"database" required:"true"`
}

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithStrictMode(true))
```

### Pattern 3: Auto-Refreshing Config
```go
refreshingConfig, err := ssmconfig.LoadWithAutoRefresh[Config](ctx, "/myapp/",
    ssmconfig.WithRefreshInterval[Config](1*time.Minute),
    ssmconfig.WithOnChange[Config](func(old, new *Config) {
        log.Printf("Config changed!")
    }))
defer refreshingConfig.Stop()
```

### Pattern 4: Custom Validators
```go
ssmconfig.RegisterValidator("custom", func(value interface{}) error {
    // Validation logic
    return nil
})

type Config struct {
    Field string `ssm:"field" validate:"custom"`
}
```

### Pattern 5: File-Based Configuration
```go
type Config struct {
    Database struct {
        URL  string `ssm:"database/url" env:"DB_URL"`
        Port int    `ssm:"database/port" env:"DB_PORT"`
    } `ssm:"database"`
}

// Load from YAML, JSON, or TOML files
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithConfigFiles("config.yaml", "config.local.yaml"))
```

**Priority Order:**
1. Environment variables (highest priority)
2. Config files (middle priority)
3. SSM Parameter Store (lowest priority)

## Notes

- All examples assume AWS credentials are properly configured
- Replace `/myapp/` prefix with your actual SSM parameter path
- **Priority order:** Environment variables > Config files > SSM parameters
- Required fields are validated based on `required:"true"` tag
- Validators run after value assignment
- Cache is per-prefix and thread-safe
- Config files support YAML, JSON, and TOML formats (via Viper)
- Multiple config files can be specified; later files override earlier ones

