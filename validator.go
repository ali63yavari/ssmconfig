package ssmconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// ValidatorFunc is a function that validates a field value.
// It receives the field value and returns an error if validation fails.
type ValidatorFunc func(value interface{}) error

// ParameterizedValidatorFunc is a function that validates a field value with parameters.
// The params string contains the parameters from the validate tag (e.g., "5" for minlen:5).
type ParameterizedValidatorFunc func(value interface{}, params string) error

var (
	validators              = make(map[string]ValidatorFunc)
	parameterizedValidators = make(map[string]ParameterizedValidatorFunc)
	validatorsMu            sync.RWMutex
)

// RegisterValidator registers a custom validator function that can be used via the validate tag.
// The name should match the value in the validate tag (e.g., validate:"myvalidator").
func RegisterValidator(name string, validator ValidatorFunc) {
	validatorsMu.Lock()
	defer validatorsMu.Unlock()
	validators[name] = validator
}

// RegisterParameterizedValidator registers a custom validator function that accepts parameters.
// The name should match the value in the validate tag (e.g., validate:"minlen:5").
func RegisterParameterizedValidator(name string, validator ParameterizedValidatorFunc) {
	validatorsMu.Lock()
	defer validatorsMu.Unlock()
	parameterizedValidators[name] = validator
}

// UnregisterValidator removes a registered validator.
func UnregisterValidator(name string) {
	validatorsMu.Lock()
	defer validatorsMu.Unlock()
	delete(validators, name)
	delete(parameterizedValidators, name)
}

// GetValidator retrieves a registered validator by name.
func GetValidator(name string) (ValidatorFunc, bool) {
	validatorsMu.RLock()
	defer validatorsMu.RUnlock()
	validator, ok := validators[name]
	return validator, ok
}

// GetParameterizedValidator retrieves a registered parameterized validator by name.
func GetParameterizedValidator(name string) (ParameterizedValidatorFunc, bool) {
	validatorsMu.RLock()
	defer validatorsMu.RUnlock()
	validator, ok := parameterizedValidators[name]
	return validator, ok
}

// validateField validates a field value using the specified validator(s).
// The validatorName can be:
// - A simple name (e.g., "email")
// - A parameterized validator (e.g., "minlen:5")
// - Multiple validators comma-separated (e.g., "email,minlen:5,maxlen:100")
// 
// For nested structs, this validates the entire struct object.
// Validators on fields within nested structs are processed recursively.
func validateField(fv reflect.Value, validatorName string, fieldName string) error {
	if validatorName == "" {
		return nil
	}

	// Get the actual value from the field
	var value interface{}
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			// For optional nested structs, nil is valid unless required
			// But if a validator is specified, we should validate it
			return fmt.Errorf("field '%s' is nil, cannot validate", fieldName)
		}
		value = fv.Elem().Interface()
	} else {
		value = fv.Interface()
	}

	// Handle struct types - validators receive the struct value
	// This allows validating the entire nested struct object

	// Support multiple validators separated by commas
	validators := strings.Split(validatorName, ",")
	for _, validatorSpec := range validators {
		validatorSpec = strings.TrimSpace(validatorSpec)
		if validatorSpec == "" {
			continue
		}

		// Check if it's a parameterized validator (e.g., "minlen:5")
		parts := strings.SplitN(validatorSpec, ":", 2)
		validatorKey := parts[0]
		params := ""
		if len(parts) > 1 {
			params = parts[1]
		}

		// Try parameterized validator first
		if params != "" {
			if paramValidator, ok := GetParameterizedValidator(validatorKey); ok {
				if err := paramValidator(value, params); err != nil {
					return fmt.Errorf("validation failed for field '%s' using validator '%s': %w", fieldName, validatorSpec, err)
				}
				continue
			}
		}

		// Try simple validator
		if validator, ok := GetValidator(validatorKey); ok {
			if err := validator(value); err != nil {
				return fmt.Errorf("validation failed for field '%s' using validator '%s': %w", fieldName, validatorSpec, err)
			}
			continue
		}

		return fmt.Errorf("validator '%s' not found for field '%s'", validatorSpec, fieldName)
	}

	return nil
}

var builtinValidatorsRegistered = false
var builtinValidatorsMu sync.Mutex

// ensureBuiltinValidators ensures built-in validators are registered.
func ensureBuiltinValidators() {
	builtinValidatorsMu.Lock()
	defer builtinValidatorsMu.Unlock()
	if !builtinValidatorsRegistered {
		RegisterBuiltinValidators()
		builtinValidatorsRegistered = true
	}
}

// RegisterBuiltinValidators registers common built-in validators.
func RegisterBuiltinValidators() {
	// Email validator
	RegisterValidator("email", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("email validator requires string type")
		}
		if !isValidEmail(str) {
			return fmt.Errorf("invalid email format: %s", str)
		}
		return nil
	})

	// URL validator
	RegisterValidator("url", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("url validator requires string type")
		}
		if !isValidURL(str) {
			return fmt.Errorf("invalid URL format: %s", str)
		}
		return nil
	})

	// Min length validator (usage: validate:"minlen:5")
	RegisterParameterizedValidator("minlen", func(value interface{}, params string) error {
		minLen, err := strconv.Atoi(params)
		if err != nil {
			return fmt.Errorf("invalid minlen parameter: %s", params)
		}
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("minlen validator requires string type")
		}
		if len(str) < minLen {
			return fmt.Errorf("string length %d is less than minimum %d", len(str), minLen)
		}
		return nil
	})

	// Max length validator (usage: validate:"maxlen:100")
	RegisterParameterizedValidator("maxlen", func(value interface{}, params string) error {
		maxLen, err := strconv.Atoi(params)
		if err != nil {
			return fmt.Errorf("invalid maxlen parameter: %s", params)
		}
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("maxlen validator requires string type")
		}
		if len(str) > maxLen {
			return fmt.Errorf("string length %d exceeds maximum %d", len(str), maxLen)
		}
		return nil
	})

	// Min value validator for numbers (usage: validate:"min:0")
	RegisterParameterizedValidator("min", func(value interface{}, params string) error {
		minVal, err := strconv.ParseFloat(params, 64)
		if err != nil {
			return fmt.Errorf("invalid min parameter: %s", params)
		}
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			val := reflect.ValueOf(v).Int()
			if float64(val) < minVal {
				return fmt.Errorf("value %d is less than minimum %v", val, minVal)
			}
		case uint, uint8, uint16, uint32, uint64:
			val := reflect.ValueOf(v).Uint()
			if float64(val) < minVal {
				return fmt.Errorf("value %d is less than minimum %v", val, minVal)
			}
		case float32, float64:
			val := reflect.ValueOf(v).Float()
			if val < minVal {
				return fmt.Errorf("value %v is less than minimum %v", val, minVal)
			}
		default:
			return fmt.Errorf("min validator requires numeric type")
		}
		return nil
	})

	// Max value validator for numbers (usage: validate:"max:100")
	RegisterParameterizedValidator("max", func(value interface{}, params string) error {
		maxVal, err := strconv.ParseFloat(params, 64)
		if err != nil {
			return fmt.Errorf("invalid max parameter: %s", params)
		}
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			val := reflect.ValueOf(v).Int()
			if float64(val) > maxVal {
				return fmt.Errorf("value %d exceeds maximum %v", val, maxVal)
			}
		case uint, uint8, uint16, uint32, uint64:
			val := reflect.ValueOf(v).Uint()
			if float64(val) > maxVal {
				return fmt.Errorf("value %d exceeds maximum %v", val, maxVal)
			}
		case float32, float64:
			val := reflect.ValueOf(v).Float()
			if val > maxVal {
				return fmt.Errorf("value %v exceeds maximum %v", val, maxVal)
			}
		default:
			return fmt.Errorf("max validator requires numeric type")
		}
		return nil
	})
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}
	atIndex := -1
	dotIndex := -1
	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Multiple @ symbols
			}
			atIndex = i
		} else if char == '.' && atIndex != -1 {
			dotIndex = i
		}
	}
	return atIndex > 0 && dotIndex > atIndex && dotIndex < len(email)-1
}

// isValidURL performs basic URL validation.
func isValidURL(url string) bool {
	if len(url) < 4 {
		return false
	}
	// Check for http:// or https://
	return (len(url) >= 7 && url[0:7] == "http://") ||
		(len(url) >= 8 && url[0:8] == "https://")
}

