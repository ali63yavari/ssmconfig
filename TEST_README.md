# Test Suite Documentation

This document describes the comprehensive test suite for the `ssmconfig` module.

## Test Files

### 1. `loader_test.go`
Tests for the `Loader` struct and loading functionality.

**Coverage:**
- `NewLoader` with various options
- `WithStrictMode`, `WithLogger`, `WithStrongTyping` options
- Cache invalidation (`InvalidateCache`)
- Loader configuration

**Key Tests:**
- `TestNewLoader` - Tests loader creation with different options
- `TestLoader_InvalidateCache` - Tests cache invalidation for specific and all prefixes
- `TestWithStrictMode` - Tests strict mode option
- `TestWithLogger` - Tests custom logger option
- `TestWithStrongTyping` - Tests strong typing option

### 2. `mapper_test.go`
Tests for struct mapping functionality.

**Coverage:**
- Basic type mapping (string, int, bool, float, slices)
- Environment variable overrides
- Required field validation (strict and non-strict modes)
- Nested structs (with and without pointers)
- JSON decoding (structs, slices, maps)
- Custom validators
- Edge cases (unexported fields, empty values, invalid types)

**Key Tests:**
- `TestMapToStruct_BasicTypes` - Tests mapping of basic Go types
- `TestMapToStruct_EnvironmentOverrides` - Tests env var precedence
- `TestMapToStruct_RequiredFields` - Tests required field validation
- `TestMapToStruct_NestedStructs` - Tests nested struct mapping
- `TestMapToStruct_JSONDecoding` - Tests JSON string decoding
- `TestMapToStruct_Validators` - Tests custom validators
- `TestMapToStruct_EdgeCases` - Tests edge cases
- `TestFilterValuesByPrefix` - Tests prefix filtering utility

### 3. `validator_test.go`
Tests for custom validator functionality.

**Coverage:**
- Validator registration and retrieval
- Parameterized validators
- Built-in validators (email, url, minlen, maxlen, min, max)
- Thread-safe validator registry
- Multiple validators per field
- Custom validator examples

**Key Tests:**
- `TestRegisterValidator` - Tests validator registration
- `TestRegisterParameterizedValidator` - Tests parameterized validators
- `TestUnregisterValidator` - Tests validator removal
- `TestBuiltinValidators` - Tests all built-in validators
- `TestValidateField` - Tests field validation logic
- `TestCustomValidators` - Tests custom validator examples

### 4. `refresh_test.go`
Tests for auto-refreshing configuration functionality.

**Coverage:**
- Auto-refresh configuration creation
- Thread-safe config access (`Get`, `GetCopy`)
- Refresh interval configuration
- Change notification callbacks
- Deep copying of configs
- Stop functionality

**Key Tests:**
- `TestLoadWithAutoRefresh` - Tests auto-refresh setup
- `TestRefreshingConfig_Get` - Tests thread-safe access
- `TestRefreshingConfig_GetCopy` - Tests safe copying
- `TestRefreshingConfig_Stop` - Tests graceful shutdown
- `TestWithRefreshInterval` - Tests interval configuration
- `TestWithOnChange` - Tests change callbacks
- `TestDeepCopy` - Tests deep copying functionality

### 5. `integration_test.go`
Integration tests that require actual AWS credentials.

**Coverage:**
- Real AWS SSM Parameter Store integration
- End-to-end configuration loading
- Environment variable overrides
- Nested structs
- JSON decoding
- Auto-refresh

**Note:** These tests are tagged with `// +build integration` and require:
- AWS credentials configured
- SSM parameters created in AWS
- Run with: `go test -tags=integration`

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Specific Test
```bash
go test -v -run TestMapToStruct_BasicTypes
```

### Run with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Integration Tests
```bash
go test -tags=integration ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

## Test Coverage

The test suite covers:

### Core Functionality
- ✅ Basic configuration loading
- ✅ Type conversion (string, int, bool, float, slices)
- ✅ Environment variable overrides
- ✅ Required field validation
- ✅ Nested structs
- ✅ JSON decoding
- ✅ Custom validators
- ✅ Caching
- ✅ Auto-refresh
- ✅ Thread safety

### Edge Cases
- ✅ Unexported fields
- ✅ Empty values
- ✅ Invalid type conversions
- ✅ Missing required fields
- ✅ Pointer fields
- ✅ Nested pointers
- ✅ Multiple validators
- ✅ Cache invalidation
- ✅ Concurrent access

### Options and Configuration
- ✅ Strict mode
- ✅ Custom logger
- ✅ Strong typing vs JSON
- ✅ Refresh intervals
- ✅ Change callbacks

## Test Dependencies

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/require` - Requirements
- `github.com/golang/mock/gomock` - Mocking (for future use)

## Adding New Tests

When adding new features, ensure you:

1. Add unit tests for the new functionality
2. Test both success and failure cases
3. Test edge cases
4. Update this documentation
5. Ensure tests pass: `go test ./...`

## Mocking AWS Services

For unit tests that don't require actual AWS credentials, consider:

1. Using environment variables to mock AWS config
2. Creating a mock SSM client interface
3. Using dependency injection for testability

## Continuous Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run tests
  run: go test -v -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

## Performance Testing

For performance-critical paths, consider adding benchmarks:

```go
func BenchmarkMapToStruct(b *testing.B) {
    // Benchmark code
}
```

Run with: `go test -bench=.`

