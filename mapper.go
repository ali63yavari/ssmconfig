package ssmconfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func mapToStruct(values map[string]string, dest interface{}, strict bool, logger func(format string, args ...interface{})) error {
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
			// Nested struct - recursively map it
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
			if err := mapToStruct(nestedValues, nestedPtr, strict, logger); err != nil {
				return fmt.Errorf("mapping nested struct field %s: %w", field.Name, err)
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

		if err := setFieldValue(fv, val); err != nil {
			return fmt.Errorf("setting field %s: %w", field.Name, err)
		}
	}

	// Only panic if there are missing required fields and strict mode is enabled
	if len(missingRequired) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", strings.Join(missingRequired, ", "))
		if strict {
			panic(fmt.Sprintf("ssmconfig: %s", msg))
		}
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
