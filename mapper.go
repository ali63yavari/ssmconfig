package ssmconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func mapToStruct(values map[string]string, dest interface{}, strict bool, logger func(format string, args ...interface{}), useStrongTyping bool) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	var missingRequired []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		ssmTag := field.Tag.Get("ssm")
		envTag := field.Tag.Get("env")
		requiredTag := field.Tag.Get("required")
		jsonTag := field.Tag.Get("json")
		validateTag := field.Tag.Get("validate")

		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}

		// Handle nested structs (with or without tags)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			// Check if this nested struct should be decoded from JSON
			if jsonTag == "true" || jsonTag == "1" || jsonTag == "yes" {
				// Decode nested struct from JSON string
				var val string
				var hasValue bool

				// Check environment variable first (override)
				if envTag != "" {
					val = os.Getenv(envTag)
					if val != "" {
						hasValue = true
					}
				}

				// Fall back to SSM parameter if env var not set or empty
				if !hasValue && ssmTag != "" {
					if ssmVal, exists := values[ssmTag]; exists && ssmVal != "" {
						val = ssmVal
						hasValue = true
					}
				}

				// Only validate required fields - skip optional fields silently
				if !hasValue {
					if isRequiredField(requiredTag) {
						missingInfo := fmt.Sprintf("field '%s' (ssm:'%s', env:'%s')", field.Name, ssmTag, envTag)
						missingRequired = append(missingRequired, missingInfo)
						if logger != nil {
							logger("WARNING: Required field missing: %s", missingInfo)
						}
					}
					continue
				}

				// Decode JSON into nested struct
				var nestedPtr interface{}
				if fv.Kind() == reflect.Ptr {
					if fv.IsNil() {
						fv.Set(reflect.New(fieldType))
					}
					nestedPtr = fv.Interface()
					// For pointer, decode directly
					if err := json.Unmarshal([]byte(val), nestedPtr); err != nil {
						return fmt.Errorf("decoding JSON for nested struct field %s: %w", field.Name, err)
					}
				} else {
					// For value type, decode into address
					nestedPtr = fv.Addr().Interface()
					if err := json.Unmarshal([]byte(val), nestedPtr); err != nil {
						return fmt.Errorf("decoding JSON for nested struct field %s: %w", field.Name, err)
					}
				}

				// Run custom validators for nested struct if specified
				if validateTag != "" {
					ensureBuiltinValidators() // Ensure built-in validators are available
					if err := validateField(fv, validateTag, field.Name); err != nil {
						return err
					}
				}
				continue
			}

			// Nested struct - recursively map it from multiple SSM parameters
			var nestedPtr interface{}
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					// Create new instance if pointer is nil
					fv.Set(reflect.New(fieldType))
				}
				nestedPtr = fv.Interface()
			} else {
				// Get address of struct field for recursive call
				nestedPtr = fv.Addr().Interface()
			}

			// Recursively map nested struct with prefix
			prefix := ""
			if ssmTag != "" {
				prefix = ssmTag
			} else if envTag != "" {
				// For nested structs without ssm tag, use field name as prefix
				prefix = strings.ToLower(field.Name)
			} else {
				// No tags - use field name as prefix for nested struct
				prefix = strings.ToLower(field.Name)
			}

			// Filter values with the prefix for nested struct
			nestedValues := filterValuesByPrefix(values, prefix)

			// Check if nested struct itself is required
			isNestedRequired := isRequiredField(requiredTag)

			// If nested struct is required, check if it has any values
			if isNestedRequired && len(nestedValues) == 0 {
				missingInfo := fmt.Sprintf("nested struct field '%s' (ssm:'%s', env:'%s')", field.Name, ssmTag, envTag)
				missingRequired = append(missingRequired, missingInfo)
				if logger != nil {
					logger("WARNING: Required nested struct missing: %s", missingInfo)
				}
				continue
			}

			if err := mapToStruct(nestedValues, nestedPtr, strict, logger, useStrongTyping); err != nil {
				return fmt.Errorf("mapping nested struct field %s: %w", field.Name, err)
			}

			// Run custom validators for nested struct if specified
			if validateTag != "" {
				ensureBuiltinValidators() // Ensure built-in validators are available
				if err := validateField(fv, validateTag, field.Name); err != nil {
					return err
				}
			}
			continue
		}

		// Handle regular (non-struct) fields
		if ssmTag == "" && envTag == "" {
			continue
		}

		isRequired := isRequiredField(requiredTag)

		var val string
		var hasValue bool

		// Check environment variable first (override)
		if envTag != "" {
			val = os.Getenv(envTag)
			if val != "" {
				hasValue = true
			}
		}

		// Fall back to SSM parameter if env var not set or empty
		if !hasValue && ssmTag != "" {
			if ssmVal, exists := values[ssmTag]; exists && ssmVal != "" {
				val = ssmVal
				hasValue = true
			}
		}

		// Only validate required fields - skip optional fields silently
		if !hasValue {
			if isRequired {
				missingInfo := fmt.Sprintf("field '%s' (ssm:'%s', env:'%s')", field.Name, ssmTag, envTag)
				missingRequired = append(missingRequired, missingInfo)
				if logger != nil {
					logger("WARNING: Required field missing: %s", missingInfo)
				}
			}
			continue
		}

		// Determine whether to use JSON decoding or strongly-typed conversion
		// Priority: json tag > loader preference
		useJSON := jsonTag == "true" || jsonTag == "1" || jsonTag == "yes"

		if !useJSON {
			// No explicit JSON tag - use loader's preference
			useJSON = !useStrongTyping
		}

		if useJSON {
			// Use JSON decoding - requires valid JSON format
			if err := setFieldValueJSON(fv, val); err != nil {
				return fmt.Errorf("decoding JSON for field %s: %w", field.Name, err)
			}
		} else {
			// Use strongly typed conversion for simple types
			// For complex types (non-string slices, maps), JSON decoding is required
			if err := setFieldValue(fv, val); err != nil {
				// If strongly typed conversion fails and it's a complex type,
				// suggest using json:"true" tag or setting useStrongTyping=false
				kind := fv.Kind()
				if kind == reflect.Slice && fv.Type().Elem().Kind() != reflect.String {
					return fmt.Errorf("setting field %s: %w (hint: use json:\"true\" tag or set useStrongTyping=false)", field.Name, err)
				}
				if kind == reflect.Map {
					return fmt.Errorf("setting field %s: %w (hint: use json:\"true\" tag or set useStrongTyping=false)", field.Name, err)
				}
				return fmt.Errorf("setting field %s: %w", field.Name, err)
			}
		}

		// Run custom validators if specified
		if validateTag != "" {
			ensureBuiltinValidators() // Ensure built-in validators are available
			if err := validateField(fv, validateTag, field.Name); err != nil {
				return err
			}
		}
	}

	// Validate and report missing required fields
	if len(missingRequired) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", strings.Join(missingRequired, ", "))
		if strict {
			panic(fmt.Sprintf("ssmconfig: %s", msg))
		}
		// In non-strict mode, we still log but don't panic
		// The error is already logged per field above
	}

	return nil
}

// ValidateRequiredFields validates that all required fields are present.
// This can be called separately to check validation without loading.
// Returns an error listing all missing required fields.
func ValidateRequiredFields[T any](values map[string]string, logger func(format string, args ...interface{})) error {
	var result T
	// Use a temporary struct to validate without actually setting values
	// We'll use strict=false to collect all missing fields
	var missingRequired []string

	// Create a validation mapper that only checks for required fields
	v := reflect.ValueOf(&result)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("type must be a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		ssmTag := field.Tag.Get("ssm")
		envTag := field.Tag.Get("env")
		requiredTag := field.Tag.Get("required")

		if !isRequiredField(requiredTag) {
			continue
		}

		// Check if value exists
		hasValue := false
		if envTag != "" {
			if os.Getenv(envTag) != "" {
				hasValue = true
			}
		}
		if !hasValue && ssmTag != "" {
			if val, exists := values[ssmTag]; exists && val != "" {
				hasValue = true
			}
		}

		if !hasValue {
			missingInfo := fmt.Sprintf("field '%s' (ssm:'%s', env:'%s')", field.Name, ssmTag, envTag)
			missingRequired = append(missingRequired, missingInfo)
			if logger != nil {
				logger("WARNING: Required field missing: %s", missingInfo)
			}
		}
	}

	if len(missingRequired) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingRequired, ", "))
	}

	return nil
}

func isRequiredField(requiredTag string) bool {
	return requiredTag == "true" || requiredTag == "1" || requiredTag == "yes"
}

// filterValuesByPrefix filters the values map to only include keys that start with the given prefix.
// The prefix is removed from the keys in the returned map.
// Example: prefix="database", key="database/host" -> "host" in result
func filterValuesByPrefix(values map[string]string, prefix string) map[string]string {
	if prefix == "" {
		return values
	}

	result := make(map[string]string)
	prefixWithSlash := prefix + "/"

	for key, value := range values {
		// Check if key starts with prefix (with or without slash)
		if strings.HasPrefix(key, prefixWithSlash) {
			// Remove prefix and leading slash
			newKey := strings.TrimPrefix(key, prefixWithSlash)
			result[newKey] = value
		} else if key == prefix {
			// Exact match - include as empty key (root level)
			result[""] = value
		}
	}

	return result
}

func setFieldValue(fv reflect.Value, val string) error {
	if !fv.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	kind := fv.Kind()

	switch kind {
	case reflect.String:
		fv.SetString(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value: %w", err)
		}
		// Check bounds for specific int types
		switch kind {
		case reflect.Int8:
			if intVal > 127 || intVal < -128 {
				return fmt.Errorf("value %d out of range for int8", intVal)
			}
		case reflect.Int16:
			if intVal > 32767 || intVal < -32768 {
				return fmt.Errorf("value %d out of range for int16", intVal)
			}
		case reflect.Int32:
			if intVal > 2147483647 || intVal < -2147483648 {
				return fmt.Errorf("value %d out of range for int32", intVal)
			}
		}
		fv.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value: %w", err)
		}
		fv.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %w", err)
		}
		fv.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid bool value: %w", err)
		}
		fv.SetBool(boolVal)

	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			// Handle string slices (comma-separated)
			parts := strings.Split(val, ",")
			slice := reflect.MakeSlice(fv.Type(), len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			fv.Set(slice)
		} else {
			return fmt.Errorf("unsupported slice type: %v", fv.Type().Elem().Kind())
		}

	default:
		return fmt.Errorf("unsupported field type: %v", kind)
	}

	return nil
}

// setFieldValueJSON decodes a JSON string and sets it to the field value.
// Supports structs, slices, maps, and other JSON-serializable types.
func setFieldValueJSON(fv reflect.Value, val string) error {
	if !fv.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	// Trim whitespace
	val = strings.TrimSpace(val)
	if val == "" {
		return fmt.Errorf("empty JSON string")
	}

	kind := fv.Kind()
	typ := fv.Type()

	// Handle pointer types
	if kind == reflect.Ptr {
		if typ.Elem().Kind() == reflect.Ptr {
			return fmt.Errorf("nested pointers not supported for JSON decoding")
		}

		// Create new instance if pointer is nil
		if fv.IsNil() {
			fv.Set(reflect.New(typ.Elem()))
		}

		// Decode into the pointed-to value
		return json.Unmarshal([]byte(val), fv.Interface())
	}

	// Handle interface{} type
	if kind == reflect.Interface {
		var result interface{}
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return fmt.Errorf("unmarshaling JSON: %w", err)
		}
		fv.Set(reflect.ValueOf(result))
		return nil
	}

	// For non-pointer types, create a temporary pointer to unmarshal into
	ptr := reflect.New(typ)
	if err := json.Unmarshal([]byte(val), ptr.Interface()); err != nil {
		return fmt.Errorf("unmarshaling JSON: %w", err)
	}

	// Set the value from the pointer
	fv.Set(ptr.Elem())
	return nil
}
