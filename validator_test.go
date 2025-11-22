package ssmconfig

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterValidator(t *testing.T) {
	t.Run("registers and retrieves validator", func(t *testing.T) {
		validator := func(value interface{}) error {
			return nil
		}

		RegisterValidator("test", validator)
		defer UnregisterValidator("test")

		retrieved, ok := GetValidator("test")
		assert.True(t, ok)
		assert.NotNil(t, retrieved)
	})

	t.Run("validator is thread-safe", func(t *testing.T) {
		validator := func(value interface{}) error {
			return nil
		}

		RegisterValidator("concurrent", validator)
		defer UnregisterValidator("concurrent")

		// Test concurrent access
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, ok := GetValidator("concurrent")
				assert.True(t, ok)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestRegisterParameterizedValidator(t *testing.T) {
	t.Run("registers parameterized validator", func(t *testing.T) {
		validator := func(value interface{}, params string) error {
			return nil
		}

		RegisterParameterizedValidator("test", validator)
		defer UnregisterValidator("test")

		retrieved, ok := GetParameterizedValidator("test")
		assert.True(t, ok)
		assert.NotNil(t, retrieved)
	})
}

func TestUnregisterValidator(t *testing.T) {
	t.Run("unregisters validator", func(t *testing.T) {
		validator := func(value interface{}) error {
			return nil
		}

		RegisterValidator("test", validator)
		_, ok := GetValidator("test")
		assert.True(t, ok)

		UnregisterValidator("test")
		_, ok = GetValidator("test")
		assert.False(t, ok)
	})
}

func TestBuiltinValidators(t *testing.T) {
	t.Run("email validator", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetValidator("email")
		require.True(t, ok)

		err := validator("test@example.com")
		assert.NoError(t, err)

		err = validator("invalid-email")
		assert.Error(t, err)
	})

	t.Run("url validator", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetValidator("url")
		require.True(t, ok)

		err := validator("https://example.com")
		assert.NoError(t, err)

		err = validator("invalid-url")
		assert.Error(t, err)
	})

	t.Run("minlen validator", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetParameterizedValidator("minlen")
		require.True(t, ok)

		err := validator("hello", "3")
		assert.NoError(t, err)

		err = validator("hi", "3")
		assert.Error(t, err)
	})

	t.Run("maxlen validator", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetParameterizedValidator("maxlen")
		require.True(t, ok)

		err := validator("hello", "10")
		assert.NoError(t, err)

		err = validator("this is too long", "10")
		assert.Error(t, err)
	})

	t.Run("min validator for numbers", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetParameterizedValidator("min")
		require.True(t, ok)

		err := validator(5, "3")
		assert.NoError(t, err)

		err = validator(2, "3")
		assert.Error(t, err)
	})

	t.Run("max validator for numbers", func(t *testing.T) {
		ensureBuiltinValidators()

		validator, ok := GetParameterizedValidator("max")
		require.True(t, ok)

		err := validator(5, "10")
		assert.NoError(t, err)

		err = validator(15, "10")
		assert.Error(t, err)
	})
}

func TestValidateField(t *testing.T) {
	t.Run("validates with simple validator", func(t *testing.T) {
		RegisterValidator("test", func(value interface{}) error {
			if value.(string) != "valid" {
				return errors.New("invalid")
			}
			return nil
		})
		defer UnregisterValidator("test")

		// Create a reflect.Value for testing
		fv := reflect.ValueOf("valid")
		err := validateField(fv, "test", "testField")
		assert.NoError(t, err)
	})

	t.Run("validates with parameterized validator", func(t *testing.T) {
		RegisterParameterizedValidator("test", func(value interface{}, params string) error {
			if value.(string) != params {
				return errors.New("mismatch")
			}
			return nil
		})
		defer UnregisterValidator("test")

		fv := reflect.ValueOf("expected")
		err := validateField(fv, "test:expected", "testField")
		assert.NoError(t, err)
	})

	t.Run("handles multiple validators", func(t *testing.T) {
		RegisterValidator("v1", func(value interface{}) error {
			return nil
		})
		RegisterValidator("v2", func(value interface{}) error {
			return nil
		})
		defer UnregisterValidator("v1")
		defer UnregisterValidator("v2")

		fv := reflect.ValueOf("test")
		err := validateField(fv, "v1,v2", "testField")
		assert.NoError(t, err)
	})

	t.Run("fails on unknown validator", func(t *testing.T) {
		fv := reflect.ValueOf("test")
		err := validateField(fv, "unknown", "testField")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCustomValidators(t *testing.T) {
	t.Run("custom alphanumeric validator", func(t *testing.T) {
		RegisterValidator("alphanumeric", func(value interface{}) error {
			str := value.(string)
			for _, r := range str {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
					return errors.New("non-alphanumeric character")
				}
			}
			return nil
		})
		defer UnregisterValidator("alphanumeric")

		validator, ok := GetValidator("alphanumeric")
		require.True(t, ok)

		err := validator("abc123")
		assert.NoError(t, err)

		err = validator("abc-123")
		assert.Error(t, err)
	})

	t.Run("custom regex validator", func(t *testing.T) {
		RegisterParameterizedValidator("regex", func(value interface{}, params string) error {
			str := value.(string)
			matched, err := regexp.MatchString(params, str)
			if err != nil {
				return err
			}
			if !matched {
				return errors.New("pattern mismatch")
			}
			return nil
		})
		defer UnregisterValidator("regex")

		validator, ok := GetParameterizedValidator("regex")
		require.True(t, ok)

		err := validator("test@example.com", "^[a-z]+@[a-z]+\\.[a-z]+$")
		assert.NoError(t, err)

		err = validator("invalid", "^[a-z]+@[a-z]+\\.[a-z]+$")
		assert.Error(t, err)
	})
}
