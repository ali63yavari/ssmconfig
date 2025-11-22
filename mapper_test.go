package ssmconfig

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMapToStruct_BasicTypes(t *testing.T) {
	t.Run("maps string field", func(t *testing.T) {
		type Config struct {
			Name string `ssm:"name"`
		}

		values := map[string]string{"name": "test"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Name)
	})

	t.Run("maps int field", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port"`
		}

		values := map[string]string{"port": "8080"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, 8080, result.Port)
	})

	t.Run("maps bool field", func(t *testing.T) {
		type Config struct {
			Debug bool `ssm:"debug"`
		}

		values := map[string]string{"debug": "true"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.True(t, result.Debug)
	})

	t.Run("maps float field", func(t *testing.T) {
		type Config struct {
			Ratio float64 `ssm:"ratio"`
		}

		values := map[string]string{"ratio": "3.14"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result.Ratio)
	})

	t.Run("maps string slice", func(t *testing.T) {
		type Config struct {
			Hosts []string `ssm:"hosts"`
		}

		values := map[string]string{"hosts": "host1,host2,host3"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, []string{"host1", "host2", "host3"}, result.Hosts)
	})

	t.Run("maps int8 field", func(t *testing.T) {
		type Config struct {
			Value int8 `ssm:"value"`
		}

		values := map[string]string{"value": "127"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, int8(127), result.Value)
	})

	t.Run("maps int16 field", func(t *testing.T) {
		type Config struct {
			Value int16 `ssm:"value"`
		}

		values := map[string]string{"value": "32767"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, int16(32767), result.Value)
	})

	t.Run("maps int32 field", func(t *testing.T) {
		type Config struct {
			Value int32 `ssm:"value"`
		}

		values := map[string]string{"value": "2147483647"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, int32(2147483647), result.Value)
	})

	t.Run("maps int64 field", func(t *testing.T) {
		type Config struct {
			Value int64 `ssm:"value"`
		}

		values := map[string]string{"value": "9223372036854775807"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, int64(9223372036854775807), result.Value)
	})

	t.Run("maps uint field", func(t *testing.T) {
		type Config struct {
			Value uint `ssm:"value"`
		}

		values := map[string]string{"value": "42"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, uint(42), result.Value)
	})

	t.Run("maps uint8 field", func(t *testing.T) {
		type Config struct {
			Value uint8 `ssm:"value"`
		}

		values := map[string]string{"value": "255"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, uint8(255), result.Value)
	})

	t.Run("maps uint16 field", func(t *testing.T) {
		type Config struct {
			Value uint16 `ssm:"value"`
		}

		values := map[string]string{"value": "65535"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, uint16(65535), result.Value)
	})

	t.Run("maps uint32 field", func(t *testing.T) {
		type Config struct {
			Value uint32 `ssm:"value"`
		}

		values := map[string]string{"value": "4294967295"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, uint32(4294967295), result.Value)
	})

	t.Run("maps uint64 field", func(t *testing.T) {
		type Config struct {
			Value uint64 `ssm:"value"`
		}

		values := map[string]string{"value": "18446744073709551615"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, uint64(18446744073709551615), result.Value)
	})

	t.Run("maps float32 field", func(t *testing.T) {
		type Config struct {
			Value float32 `ssm:"value"`
		}

		values := map[string]string{"value": "3.14"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, float32(3.14), result.Value)
	})

	t.Run("maps bool false", func(t *testing.T) {
		type Config struct {
			Debug bool `ssm:"debug"`
		}

		values := map[string]string{"debug": "false"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.False(t, result.Debug)
	})
}

func TestMapToStruct_EnvironmentOverrides(t *testing.T) {
	t.Run("env var overrides SSM value", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database_url" env:"DB_URL"`
		}

		os.Setenv("DB_URL", "env-override")
		defer os.Unsetenv("DB_URL")

		values := map[string]string{"database_url": "ssm-value"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "env-override", result.DatabaseURL)
	})

	t.Run("falls back to SSM when env var not set", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database_url" env:"DB_URL"`
		}

		values := map[string]string{"database_url": "ssm-value"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "ssm-value", result.DatabaseURL)
	})

	t.Run("empty env var falls back to SSM", func(t *testing.T) {
		type Config struct {
			DatabaseURL string `ssm:"database_url" env:"DB_URL"`
		}

		os.Setenv("DB_URL", "")
		defer os.Unsetenv("DB_URL")

		values := map[string]string{"database_url": "ssm-value"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "ssm-value", result.DatabaseURL)
	})
}

func TestMapToStruct_RequiredFields(t *testing.T) {
	t.Run("logs warning for missing required field in non-strict mode", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"true"`
		}

		var loggedMessages []string
		logger := func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		}

		values := map[string]string{}
		var result Config
		err := mapToStruct(values, &result, false, logger, true)
		require.NoError(t, err)
		assert.Len(t, loggedMessages, 1)
		loggedStr := loggedMessages[0]
		assert.Contains(t, loggedStr, "WARNING")
		// Check that the logged message contains field information (either api_key or field name)
		assert.True(t,
			contains(loggedStr, "api_key") || contains(loggedStr, "field"),
			"Logged message should contain field information: %s", loggedStr)
	})

	t.Run("panics for missing required field in strict mode", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"true"`
		}

		values := map[string]string{}
		var result Config

		assert.Panics(t, func() {
			_ = mapToStruct(values, &result, true, nil, true)
		})
	})

	t.Run("does not panic when required field is present", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"true"`
		}

		values := map[string]string{"api_key": "secret"}
		var result Config
		err := mapToStruct(values, &result, true, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "secret", result.APIKey)
	})

	t.Run("validates required field from env var", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" env:"API_KEY" required:"true"`
		}

		os.Setenv("API_KEY", "env-secret")
		defer os.Unsetenv("API_KEY")

		values := map[string]string{}
		var result Config
		err := mapToStruct(values, &result, true, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "env-secret", result.APIKey)
	})

	t.Run("handles required field with value 1", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"1"`
		}

		values := map[string]string{}
		var result Config

		assert.Panics(t, func() {
			_ = mapToStruct(values, &result, true, nil, true)
		})
	})

	t.Run("handles required field with value yes", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"yes"`
		}

		values := map[string]string{}
		var result Config

		assert.Panics(t, func() {
			_ = mapToStruct(values, &result, true, nil, true)
		})
	})
}

func TestMapToStruct_NestedStructs(t *testing.T) {
	t.Run("maps nested struct from SSM parameters", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `ssm:"host"`
			Port int    `ssm:"port"`
		}

		type Config struct {
			Database DatabaseConfig `ssm:"database"`
		}

		values := map[string]string{
			"database/host": "localhost",
			"database/port": "5432",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Database.Host)
		assert.Equal(t, 5432, result.Database.Port)
	})

	t.Run("maps nested struct with pointer", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `ssm:"host"`
			Port int    `ssm:"port"`
		}

		type Config struct {
			Database *DatabaseConfig `ssm:"database"`
		}

		values := map[string]string{
			"database/host": "localhost",
			"database/port": "5432",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		require.NotNil(t, result.Database)
		assert.Equal(t, "localhost", result.Database.Host)
		assert.Equal(t, 5432, result.Database.Port)
	})

	t.Run("handles required nested struct", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `ssm:"host"`
		}

		type Config struct {
			Database DatabaseConfig `ssm:"database" required:"true"`
		}

		var loggedMessages []string
		logger := func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		}

		values := map[string]string{}
		var result Config
		err := mapToStruct(values, &result, false, logger, true)
		require.NoError(t, err)
		assert.Len(t, loggedMessages, 1)
	})

	t.Run("maps nested struct without ssm tag", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `ssm:"host"`
			Port int    `ssm:"port"`
		}

		type Config struct {
			Database DatabaseConfig // No ssm tag, uses field name
		}

		values := map[string]string{
			"database/host": "localhost",
			"database/port": "5432",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Database.Host)
		assert.Equal(t, 5432, result.Database.Port)
	})

	t.Run("maps nested struct with env tag only", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `ssm:"host"`
		}

		type Config struct {
			Database DatabaseConfig `env:"DB_CONFIG"` // Only env tag
		}

		values := map[string]string{
			"database/host": "localhost",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Database.Host)
	})
}

func TestMapToStruct_JSONDecoding(t *testing.T) {
	t.Run("decodes JSON string to struct", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}

		type Config struct {
			Database DatabaseConfig `ssm:"database" json:"true"`
		}

		values := map[string]string{
			"database": `{"host":"localhost","port":5432}`,
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Database.Host)
		assert.Equal(t, 5432, result.Database.Port)
	})

	t.Run("decodes JSON string to slice", func(t *testing.T) {
		type Config struct {
			Hosts []string `ssm:"hosts" json:"true"`
		}

		values := map[string]string{
			"hosts": `["host1","host2","host3"]`,
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, []string{"host1", "host2", "host3"}, result.Hosts)
	})

	t.Run("decodes JSON string to map", func(t *testing.T) {
		type Config struct {
			Metadata map[string]string `ssm:"metadata" json:"true"`
		}

		values := map[string]string{
			"metadata": `{"key1":"value1","key2":"value2"}`,
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "value1", result.Metadata["key1"])
		assert.Equal(t, "value2", result.Metadata["key2"])
	})

	t.Run("uses JSON decoding when useStrongTyping is false", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port"`
		}

		values := map[string]string{"port": "8080"}
		var result Config
		err := mapToStruct(values, &result, false, nil, false)
		require.NoError(t, err)
		assert.Equal(t, 8080, result.Port)
	})

	t.Run("decodes JSON with json tag value 1", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port" json:"1"`
		}

		values := map[string]string{"port": "8080"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, 8080, result.Port)
	})

	t.Run("decodes JSON with json tag value yes", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port" json:"yes"`
		}

		values := map[string]string{"port": "8080"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, 8080, result.Port)
	})

	t.Run("decodes JSON nested struct with pointer", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `json:"host"`
		}

		type Config struct {
			Database *DatabaseConfig `ssm:"database" json:"true"`
		}

		values := map[string]string{
			"database": `{"host":"localhost"}`,
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		require.NotNil(t, result.Database)
		assert.Equal(t, "localhost", result.Database.Host)
	})

	t.Run("decodes JSON nested struct with env override", func(t *testing.T) {
		type DatabaseConfig struct {
			Host string `json:"host"`
		}

		type Config struct {
			Database DatabaseConfig `ssm:"database" env:"DB_CONFIG" json:"true"`
		}

		os.Setenv("DB_CONFIG", `{"host":"env-host"}`)
		defer os.Unsetenv("DB_CONFIG")

		values := map[string]string{
			"database": `{"host":"ssm-host"}`,
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "env-host", result.Database.Host)
	})
}

func TestMapToStruct_Validators(t *testing.T) {
	t.Run("runs validator on field", func(t *testing.T) {
		RegisterValidator("test", func(value interface{}) error {
			str := value.(string)
			if str != "valid" {
				return errors.New("invalid value")
			}
			return nil
		})
		defer UnregisterValidator("test")

		type Config struct {
			Field string `ssm:"field" validate:"test"`
		}

		values := map[string]string{"field": "valid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "valid", result.Field)
	})

	t.Run("fails validation", func(t *testing.T) {
		RegisterValidator("test", func(value interface{}) error {
			str := value.(string)
			if str != "valid" {
				return errors.New("invalid value")
			}
			return nil
		})
		defer UnregisterValidator("test")

		type Config struct {
			Field string `ssm:"field" validate:"test"`
		}

		values := map[string]string{"field": "invalid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("runs multiple validators", func(t *testing.T) {
		RegisterValidator("minlen3", func(value interface{}) error {
			str := value.(string)
			if len(str) < 3 {
				return errors.New("too short")
			}
			return nil
		})
		RegisterValidator("maxlen10", func(value interface{}) error {
			str := value.(string)
			if len(str) > 10 {
				return errors.New("too long")
			}
			return nil
		})
		defer UnregisterValidator("minlen3")
		defer UnregisterValidator("maxlen10")

		type Config struct {
			Field string `ssm:"field" validate:"minlen3,maxlen10"`
		}

		values := map[string]string{"field": "valid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
	})

	t.Run("runs validator on nested struct", func(t *testing.T) {
		RegisterValidator("test", func(value interface{}) error {
			return nil
		})
		defer UnregisterValidator("test")

		type DatabaseConfig struct {
			Host string `ssm:"host"`
		}

		type Config struct {
			Database DatabaseConfig `ssm:"database" validate:"test"`
		}

		values := map[string]string{
			"database/host": "localhost",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
	})
}

func TestMapToStruct_EdgeCases(t *testing.T) {
	t.Run("handles unexported fields", func(t *testing.T) {
		type Config struct {
			Public  string `ssm:"public"`
			private string `ssm:"private"` // unexported
		}

		values := map[string]string{
			"public":  "value1",
			"private": "value2",
		}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "value1", result.Public)
		// private field should remain zero value
	})

	t.Run("handles fields without tags", func(t *testing.T) {
		type Config struct {
			WithTag    string `ssm:"with_tag"`
			WithoutTag string
		}

		values := map[string]string{"with_tag": "value"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Equal(t, "value", result.WithTag)
		assert.Empty(t, result.WithoutTag)
	})

	t.Run("handles empty values", func(t *testing.T) {
		type Config struct {
			Field string `ssm:"field"`
		}

		values := map[string]string{"field": ""}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.Empty(t, result.Field)
	})

	t.Run("handles invalid int value", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port"`
		}

		values := map[string]string{"port": "invalid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid int value")
	})

	t.Run("handles invalid bool value", func(t *testing.T) {
		type Config struct {
			Debug bool `ssm:"debug"`
		}

		values := map[string]string{"debug": "invalid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bool value")
	})

	t.Run("handles invalid float value", func(t *testing.T) {
		type Config struct {
			Ratio float64 `ssm:"ratio"`
		}

		values := map[string]string{"ratio": "invalid"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid float value")
	})

	t.Run("handles int8 overflow", func(t *testing.T) {
		type Config struct {
			Value int8 `ssm:"value"`
		}

		values := map[string]string{"value": "1000"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("handles int16 overflow", func(t *testing.T) {
		type Config struct {
			Value int16 `ssm:"value"`
		}

		values := map[string]string{"value": "100000"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("handles int32 overflow", func(t *testing.T) {
		type Config struct {
			Value int32 `ssm:"value"`
		}

		values := map[string]string{"value": "3000000000"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("handles unsupported slice type", func(t *testing.T) {
		type Config struct {
			Values []int `ssm:"values"`
		}

		values := map[string]string{"values": "1,2,3"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported slice type")
	})

	t.Run("handles unsupported field type", func(t *testing.T) {
		type Config struct {
			Value chan int `ssm:"value"`
		}

		values := map[string]string{"value": "test"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported field type")
	})

	t.Run("handles invalid dest type", func(t *testing.T) {
		var notStruct string
		err := mapToStruct(map[string]string{}, notStruct, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer to struct")
	})

	t.Run("handles non-pointer dest", func(t *testing.T) {
		type Config struct {
			Value string `ssm:"value"`
		}

		var result Config
		err := mapToStruct(map[string]string{}, result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer to struct")
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		type Config struct {
			Database struct {
				Host string `json:"host"`
			} `ssm:"database" json:"true"`
		}

		values := map[string]string{"database": "invalid-json"}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON")
	})

	t.Run("handles empty JSON string", func(t *testing.T) {
		type Config struct {
			Port int `ssm:"port" json:"true"`
		}

		values := map[string]string{"port": "   "} // Whitespace only
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty JSON string")
	})

	t.Run("handles nested pointer in JSON", func(t *testing.T) {
		type Config struct {
			Value **string `ssm:"value" json:"true"`
		}

		values := map[string]string{"value": `"test"`}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nested pointers not supported")
	})

	t.Run("handles interface{} type with JSON", func(t *testing.T) {
		type Config struct {
			Value interface{} `ssm:"value" json:"true"`
		}

		values := map[string]string{"value": `{"key":"value"}`}
		var result Config
		err := mapToStruct(values, &result, false, nil, true)
		require.NoError(t, err)
		assert.NotNil(t, result.Value)
	})
}

func TestValidateRequiredFields(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"true"`
			Port   int    `ssm:"port"`
		}

		var loggedMessages []string
		logger := func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		}

		values := map[string]string{"port": "8080"}
		err := ValidateRequiredFields[Config](values, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
		assert.Len(t, loggedMessages, 1)
	})

	t.Run("passes when all required fields present", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" required:"true"`
			Port   int    `ssm:"port"`
		}

		values := map[string]string{"api_key": "secret"}
		err := ValidateRequiredFields[Config](values, nil)
		require.NoError(t, err)
	})

	t.Run("validates from env var", func(t *testing.T) {
		type Config struct {
			APIKey string `ssm:"api_key" env:"API_KEY" required:"true"`
		}

		os.Setenv("API_KEY", "env-secret")
		defer os.Unsetenv("API_KEY")

		values := map[string]string{}
		err := ValidateRequiredFields[Config](values, nil)
		require.NoError(t, err)
	})

	t.Run("handles non-struct type", func(t *testing.T) {
		err := ValidateRequiredFields[string](map[string]string{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a struct")
	})
}

func TestFilterValuesByPrefix(t *testing.T) {
	t.Run("filters values by prefix", func(t *testing.T) {
		values := map[string]string{
			"database/host": "localhost",
			"database/port": "5432",
			"server/host":   "0.0.0.0",
			"server/port":   "8080",
		}

		result := filterValuesByPrefix(values, "database")
		assert.Equal(t, map[string]string{
			"host": "localhost",
			"port": "5432",
		}, result)
	})

	t.Run("handles empty prefix", func(t *testing.T) {
		values := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		result := filterValuesByPrefix(values, "")
		assert.Equal(t, values, result)
	})

	t.Run("handles exact match", func(t *testing.T) {
		values := map[string]string{
			"database": "value",
		}

		result := filterValuesByPrefix(values, "database")
		assert.Equal(t, map[string]string{"": "value"}, result)
	})

	t.Run("handles prefix without trailing slash", func(t *testing.T) {
		values := map[string]string{
			"database/host": "localhost",
		}

		result := filterValuesByPrefix(values, "database")
		assert.Equal(t, map[string]string{"host": "localhost"}, result)
	})
}

func TestSetFieldValue_ErrorCases(t *testing.T) {
	t.Run("handles unsettable field", func(t *testing.T) {
		type Config struct {
			value string // unexported
		}

		config := &Config{}
		fv := reflect.ValueOf(config).Elem().Field(0)
		err := setFieldValue(fv, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be set")
	})
}

func TestSetFieldValueJSON_ErrorCases(t *testing.T) {
	t.Run("handles unsettable field", func(t *testing.T) {
		type Config struct {
			value string // unexported
		}

		config := &Config{}
		fv := reflect.ValueOf(config).Elem().Field(0)
		err := setFieldValueJSON(fv, `"test"`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be set")
	})

	t.Run("handles empty JSON string", func(t *testing.T) {
		type Config struct {
			Value string
		}

		config := &Config{}
		fv := reflect.ValueOf(config).Elem().Field(0)
		err := setFieldValueJSON(fv, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty JSON string")
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		type Config struct {
			Value string
		}

		config := &Config{}
		fv := reflect.ValueOf(config).Elem().Field(0)
		err := setFieldValueJSON(fv, "invalid-json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshaling JSON")
	})
}
