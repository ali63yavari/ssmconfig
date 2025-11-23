# ssmconfig

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/ali63yavari/ssmconfig.svg)](https://pkg.go.dev/github.com/ali63yavari/ssmconfig)
[![Go Report Card](https://goreportcard.com/badge/github.com/ali63yavari/ssmconfig)](https://goreportcard.com/report/github.com/ali63yavari/ssmconfig)
[![CI](https://github.com/ali63yavari/ssmconfig/actions/workflows/ci.yml/badge.svg)](https://github.com/ali63yavari/ssmconfig/actions/workflows/ci.yml)

A powerful, type-safe Go library for loading configuration from AWS Systems Manager (SSM) Parameter Store, with support for environment variable overrides, file-based configuration, auto-refresh, custom validators, and more.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Features in Detail](#features-in-detail)
  - [1. Basic Configuration Loading](#1-basic-configuration-loading)
  - [2. Environment Variable Overrides](#2-environment-variable-overrides)
  - [3. Nested Structs](#3-nested-structs)
  - [4. Required Fields](#4-required-fields)
  - [5. Custom Logging](#5-custom-logging)
  - [6. Custom Validators](#6-custom-validators)
  - [7. JSON Decoding](#7-json-decoding)
  - [8. File-Based Configuration](#8-file-based-configuration)
  - [9. Auto-Refresh Configuration](#9-auto-refresh-configuration)
  - [10. Strong Typing vs JSON Decoding](#10-strong-typing-vs-json-decoding)
  - [11. Caching](#11-caching)
  - [12. Viper Integration](#12-viper-integration)
- [Struct Tags Reference](#struct-tags-reference)
- [Loader Options](#loader-options)
- [RefreshingConfig Options](#refreshingconfig-options)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [AWS Setup](#aws-setup)
- [Type Conversion](#type-conversion)
- [Error Handling](#error-handling)
- [Thread Safety](#thread-safety)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)
- [Support](#support)

## Features

- 🔒 **Type-Safe**: Uses Go generics to return strongly-typed configuration structs
- 🌍 **Environment Overrides**: Environment variables automatically override SSM parameters
- 📁 **File Support**: Load from YAML, JSON, and TOML files (via Viper)
- 🔄 **Auto-Refresh**: Automatically refresh configuration at configurable intervals
- ✅ **Validation**: Built-in and custom validators for field validation
- 🏗️ **Nested Structs**: Full support for nested configuration structures
- 📦 **JSON Decoding**: Decode complex JSON strings from SSM into structs
- 🔐 **Required Fields**: Mark fields as required with optional strict mode
- ⚡ **Caching**: Built-in caching to reduce SSM API calls
- 🔌 **Viper Integration**: Seamless integration with Viper configuration library
- 🎯 **Priority System**: ENV > File > SSM (configurable priority order)

## Installation

```bash
go get github.com/ali63yavari/ssmconfig
```

## Quick Start

### 1. Define Your Configuration Struct

```go
type Config struct {
    DatabaseURL string `ssm:"database_url" env:"DB_URL" required:"true"`
    Port        int    `ssm:"port" env:"PORT"`
    Debug       bool   `ssm:"debug" env:"DEBUG"`
}
```

### 2. Load Configuration

```go
package main

import (
    "context"
    "log"
    
    "github.com/ali63yavari/ssmconfig"
)

func main() {
    ctx := context.Background()
    
    cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Database: %s", cfg.DatabaseURL)
    log.Printf("Port: %d", cfg.Port)
}
```

### 3. Set Up AWS SSM Parameters

```bash
aws ssm put-parameter --name "/myapp/database_url" --value "postgres://localhost:5432/mydb" --type "String"
aws ssm put-parameter --name "/myapp/port" --value "8080" --type "String"
aws ssm put-parameter --name "/myapp/debug" --value "true" --type "String"
```

That's it! Your configuration is loaded and ready to use.

## Core Concepts

### Configuration Priority

The library follows a priority order when resolving configuration values:

1. **Environment Variables** (highest priority)
2. **File-based Configuration** (YAML, JSON, TOML)
3. **AWS SSM Parameter Store** (lowest priority)

This allows you to:
- Override any SSM parameter with an environment variable
- Use local config files for development
- Store production configs in SSM

### Struct Tags

The library uses struct tags to define how fields are mapped:

- `ssm:"parameter_name"` - SSM parameter path (relative to prefix)
- `env:"ENV_VAR_NAME"` - Environment variable name
- `required:"true"` - Mark field as required
- `json:"true"` - Decode value as JSON string
- `validate:"validator1,validator2:param"` - Custom validators

## Features in Detail

### 1. Basic Configuration Loading

Load configuration from SSM Parameter Store with automatic type conversion.

```go
type Config struct {
    DatabaseURL string `ssm:"database_url"`
    Port        int    `ssm:"port"`
    Debug       bool   `ssm:"debug"`
}

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
```

**Supported Types:**
- `string`, `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `[]string` (comma-separated values)

### 2. Environment Variable Overrides

Environment variables automatically override SSM parameters.

```go
type Config struct {
    DatabaseURL string `ssm:"database_url" env:"DB_URL"`
    Port        int    `ssm:"port" env:"PORT"`
}

// Set environment variable
os.Setenv("DB_URL", "postgres://override:5432/mydb")

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
// cfg.DatabaseURL will be "postgres://override:5432/mydb" (from ENV)
```

### 3. Nested Structs

Full support for nested configuration structures with automatic prefix handling.

```go
type DatabaseConfig struct {
    Host string `ssm:"host" env:"DB_HOST"`
    Port int    `ssm:"port" env:"DB_PORT"`
}

type ServerConfig struct {
    Host string `ssm:"host" env:"SERVER_HOST"`
    Port int    `ssm:"port" env:"SERVER_PORT"`
}

type Config struct {
    Database DatabaseConfig `ssm:"database"`
    Server   ServerConfig   `ssm:"server"`
}

// SSM Parameters:
// /myapp/database/host = "localhost"
// /myapp/database/port = "5432"
// /myapp/server/host = "0.0.0.0"
// /myapp/server/port = "8080"

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
```

**Deep Nesting:**
```go
type Config struct {
    App struct {
        Database struct {
            Host string `ssm:"host"`
            Port int    `ssm:"port"`
        } `ssm:"database"`
    } `ssm:"app"`
}
```

### 4. Required Fields

Mark fields as required. Missing required fields will be logged, and optionally cause a panic in strict mode.

```go
type Config struct {
    APIKey    string `ssm:"api_key" env:"API_KEY" required:"true"`
    DatabaseURL string `ssm:"database_url" env:"DB_URL" required:"true"`
    Port        int    `ssm:"port"` // Optional
}

// Non-strict mode (default): logs warnings
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")

// Strict mode: panics on missing required fields
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithStrictMode(true))
```

### 5. Custom Logging

Integrate with your logging library (Sentry, zap, logrus, etc.) without adding dependencies.

```go
import "github.com/sirupsen/logrus"

logger := logrus.New()

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithLogger(func(format string, args ...interface{}) {
        logger.Warnf(format, args...)
    }))
```

### 6. Custom Validators

Register custom validators for field validation.

```go
// Register a custom validator
ssmconfig.RegisterValidator("alphanumeric", func(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return fmt.Errorf("alphanumeric validator requires string type")
    }
    for _, r := range str {
        if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
            return fmt.Errorf("string must be alphanumeric")
        }
    }
    return nil
})

type Config struct {
    Username string `ssm:"username" validate:"alphanumeric"`
}
```

**Built-in Validators:**
- `email` - Validates email format
- `url` - Validates URL format
- `minlen:N` - Minimum string length (e.g., `minlen:5`)
- `maxlen:N` - Maximum string length (e.g., `maxlen:100`)
- `min:N` - Minimum numeric value (e.g., `min:0`)
- `max:N` - Maximum numeric value (e.g., `max:100`)

**Parameterized Validators:**
```go
type Config struct {
    Password string `ssm:"password" validate:"minlen:8,maxlen:128"`
    Port     int    `ssm:"port" validate:"min:1,max:65535"`
}
```

**Multiple Validators:**
```go
type Config struct {
    Email string `ssm:"email" validate:"email,minlen:5,maxlen:100"`
}
```

### 7. JSON Decoding

Decode complex JSON strings from SSM into structs, slices, or maps.

```go
type DatabaseConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
    SSL  bool   `json:"ssl"`
}

type Config struct {
    Database DatabaseConfig `ssm:"database" json:"true"`
    Servers  []ServerConfig `ssm:"servers" json:"true"`
}

// SSM Parameter: /myapp/database = '{"host":"localhost","port":5432,"ssl":true}'
// SSM Parameter: /myapp/servers = '[{"host":"api","port":8080},{"host":"web","port":80}]'

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
```

**JSON from Environment Variables:**
```go
type Config struct {
    Database DatabaseConfig `ssm:"database" env:"DB_CONFIG" json:"true"`
}

os.Setenv("DB_CONFIG", `{"host":"localhost","port":5432}`)
```

### 8. File-Based Configuration

Load configuration from YAML, JSON, and TOML files using Viper.

```go
type Config struct {
    Database struct {
        URL  string `ssm:"database/url" env:"DB_URL"`
        Port int    `ssm:"database/port" env:"DB_PORT"`
    } `ssm:"database"`
    
    Server struct {
        Host string `ssm:"server/host" env:"SERVER_HOST"`
        Port int    `ssm:"server/port" env:"SERVER_PORT"`
    } `ssm:"server"`
}

// config.yaml
// database:
//   url: "postgres://localhost:5432/mydb"
//   port: 5432
// server:
//   host: "0.0.0.0"
//   port: 8080

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithConfigFiles("config.yaml", "config.local.yaml"))
```

**Supported Formats:**
- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- TOML (`.toml`)

**Multiple Files:**
Later files override earlier ones. Useful for:
- Base config: `config.yaml`
- Environment-specific: `config.prod.yaml`
- Local overrides: `config.local.yaml`

### 9. Auto-Refresh Configuration

Automatically refresh configuration at configurable intervals.

```go
type Config struct {
    DatabaseURL string `ssm:"database_url" env:"DB_URL"`
    LastUpdated string `ssm:"last_updated"`
}

// Auto-refresh every 5 minutes
refreshingConfig, err := ssmconfig.LoadWithAutoRefresh[Config](ctx, "/myapp/",
    ssmconfig.WithRefreshInterval[Config](5*time.Minute),
    ssmconfig.WithOnChange[Config](func(old, new *Config) {
        log.Printf("Config changed! Old DB: %s, New DB: %s", 
            old.DatabaseURL, new.DatabaseURL)
    }))
defer refreshingConfig.Stop()

// Thread-safe access
cfg := refreshingConfig.Get()

// Get a safe copy (for long-running operations)
cfgCopy := refreshingConfig.GetCopy()
```

**Manual Refresh:**
```go
err := refreshingConfig.Refresh()
```

### 10. Strong Typing vs JSON Decoding

Control whether to use strongly-typed conversion or JSON decoding.

```go
// Default: Strong typing (for simple types)
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")

// Prefer JSON decoding for all types
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithStrongTyping(false))

// Per-field control (always takes precedence)
type Config struct {
    SimpleValue string `ssm:"simple"`                    // Uses strong typing
    ComplexData MyStruct `ssm:"complex" json:"true"`    // Always uses JSON
}
```

### 11. Caching

Built-in caching reduces SSM API calls. Cache is per-prefix and thread-safe.

```go
loader, err := ssmconfig.NewLoader(ctx)
if err != nil {
    log.Fatal(err)
}

// First call: fetches from SSM
cfg1, err := ssmconfig.LoadWithLoader[Config](loader, ctx, "/myapp/")

// Second call: uses cache
cfg2, err := ssmconfig.LoadWithLoader[Config](loader, ctx, "/myapp/")

// Invalidate cache for a specific prefix
loader.InvalidateCache("/myapp/")

// Invalidate all caches
loader.InvalidateCache("")
```

### 12. Viper Integration

Use ssmconfig as a remote provider for Viper.

```go
import (
    "github.com/spf13/viper"
    "github.com/ali63yavari/ssmconfig"
)

ctx := context.Background()

// Create Viper remote provider
provider, err := ssmconfig.NewViperRemoteProvider(ctx, "/myapp/")
if err != nil {
    log.Fatal(err)
}
defer provider.Stop()

// Add to Viper
viper.RemoteConfig = provider

// Read configuration
err = viper.ReadRemoteConfig()
if err != nil {
    log.Fatal(err)
}

// Use Viper as normal
dbURL := viper.GetString("database_url")
```

## Struct Tags Reference

| Tag | Description | Example |
|-----|-------------|---------|
| `ssm` | SSM parameter path (relative to prefix) | `ssm:"database_url"` |
| `env` | Environment variable name | `env:"DB_URL"` |
| `required` | Mark field as required | `required:"true"` |
| `json` | Decode value as JSON | `json:"true"` |
| `validate` | Custom validators | `validate:"email,minlen:5"` |

## Loader Options

| Option | Description |
|-------|-------------|
| `WithStrictMode(bool)` | Enable strict mode (panic on missing required fields) |
| `WithLogger(func)` | Custom logger function |
| `WithStrongTyping(bool)` | Control strong typing vs JSON decoding |
| `WithConfigFiles(...string)` | Add config file paths (YAML, JSON, TOML) |

## RefreshingConfig Options

| Option | Description |
|-------|-------------|
| `WithRefreshInterval[T](time.Duration)` | Set refresh interval |
| `WithOnChange[T](func(old, new *T))` | Change notification callback |

## Best Practices

### 1. Use Environment-Specific Prefixes

```go
prefix := os.Getenv("CONFIG_PREFIX")
if prefix == "" {
    prefix = "/myapp/dev/"
}
cfg, err := ssmconfig.Load[Config](ctx, prefix)
```

### 2. Validate Required Fields in Production

```go
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithStrictMode(true), // Panic on missing required fields
    ssmconfig.WithLogger(logger.Warnf))
```

### 3. Use File Configs for Local Development

```go
var configFiles []string
if os.Getenv("ENV") == "local" {
    configFiles = []string{"config.local.yaml"}
}

cfg, err := ssmconfig.Load[Config](ctx, "/myapp/",
    ssmconfig.WithConfigFiles(configFiles...))
```

### 4. Reuse Loader Instances

```go
loader, err := ssmconfig.NewLoader(ctx)
if err != nil {
    log.Fatal(err)
}

// Reuse for multiple configs
appConfig, _ := ssmconfig.LoadWithLoader[AppConfig](loader, ctx, "/app/")
dbConfig, _ := ssmconfig.LoadWithLoader[DBConfig](loader, ctx, "/db/")
```

### 5. Use Auto-Refresh for Long-Running Services

```go
refreshingConfig, err := ssmconfig.LoadWithAutoRefresh[Config](ctx, "/myapp/",
    ssmconfig.WithRefreshInterval[Config](5*time.Minute))

defer refreshingConfig.Stop()

// Use in your service
for {
    cfg := refreshingConfig.Get()
    // Use config...
    time.Sleep(1 * time.Second)
}
```

## Examples

Comprehensive examples are available in the [`example/`](./example/) directory:

- [`basic.go`](./example/basic.go) - Basic usage
- [`nested_structs.go`](./example/nested_structs.go) - Nested structures
- [`json_decoding.go`](./example/json_decoding.go) - JSON decoding
- [`required_fields.go`](./example/required_fields.go) - Required fields
- [`custom_validators.go`](./example/custom_validators.go) - Custom validators
- [`auto_refresh.go`](./example/auto_refresh.go) - Auto-refresh
- [`file_config.go`](./example/file_config.go) - File-based config
- [`viper_integration.go`](./example/viper_integration.go) - Viper integration
- [`edge_cases.go`](./example/edge_cases.go) - Edge cases
- [`advanced_nested.go`](./example/advanced_nested.go) - Advanced nesting

Run examples:
```bash
go run example/basic.go
```

## AWS Setup

### IAM Permissions

Your AWS credentials need the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:GetParameter",
                "ssm:GetParameters",
                "ssm:GetParametersByPath"
            ],
            "Resource": "arn:aws:ssm:*:*:parameter/myapp/*"
        }
    ]
}
```

### Creating Parameters

```bash
# Basic parameters
aws ssm put-parameter --name "/myapp/database_url" \
    --value "postgres://localhost:5432/mydb" \
    --type "String"

aws ssm put-parameter --name "/myapp/port" \
    --value "8080" \
    --type "String"

# Secure parameters (encrypted)
aws ssm put-parameter --name "/myapp/api_key" \
    --value "secret-key" \
    --type "SecureString" \
    --key-id "alias/aws/ssm"
```

## Type Conversion

The library automatically converts string values from SSM to Go types:

| SSM Value | Go Type | Result |
|-----------|---------|--------|
| `"123"` | `int` | `123` |
| `"true"` | `bool` | `true` |
| `"3.14"` | `float64` | `3.14` |
| `"a,b,c"` | `[]string` | `["a", "b", "c"]` |

## Error Handling

The library returns errors for:
- AWS configuration issues
- Missing SSM parameters (if required)
- Type conversion failures
- Validation failures
- JSON decoding errors

Always check errors:
```go
cfg, err := ssmconfig.Load[Config](ctx, "/myapp/")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```

## Thread Safety

- `Loader` is thread-safe and can be used concurrently
- `RefreshingConfig.Get()` and `RefreshingConfig.GetCopy()` are thread-safe
- Validator registry is thread-safe
- Cache operations are thread-safe

## Testing

The module includes comprehensive unit tests with 70%+ coverage.

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests (requires AWS credentials)
go test -tags=integration ./...
```
