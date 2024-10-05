package jprops

import (
	"bufio"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// parseProperties reads properties data and returns a map of key-value pairs.
func parseProperties(data []byte) (map[string]interface{}, error) {
	props := make(map[string]interface{})
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	pattern := regexp.MustCompile("^([^#][^=]*)=(.*)")

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		match := pattern.FindStringSubmatch(line)
		if len(match) > 0 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])

			// Split the key into parts for nested maps
			keyList := strings.Split(key, ".")

			current := props
			for i := 0; i < len(keyList)-1; i++ {
				k := keyList[i]
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				// Type assertion to navigate deeper into the nested map
				if nextMap, ok := current[k].(map[string]interface{}); ok {
					current = nextMap
				} else {
					// Handle type mismatch if the existing key is not a map
					return nil, fmt.Errorf("type mismatch at key: %s", k)
				}
			}

			// Assign the value to the last key
			lastKey := keyList[len(keyList)-1]
			current[lastKey] = value
		} else {
			// Line didn't match the pattern, handle as needed (skip or log)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return props, nil
}

// getNestedProperty traverses the nested map to retrieve the value for a dot-separated key.
func getNestedProperty(props map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	var current interface{} = props
	for _, part := range parts {
		switch currMap := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = currMap[part]
			if !ok {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return current, true
}

// setStructFields sets the fields of the struct based on the provided properties.
func setStructFields(structVal reflect.Value, props map[string]interface{}) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the property key from the struct tag or use the field name
		propertyKey := fieldType.Tag.Get("property")
		if propertyKey == "" {
			propertyKey = fieldType.Name
		}

		// Retrieve the value using the helper function
		value, ok := getNestedProperty(props, propertyKey)
		if !ok {
			continue // Property not found in data
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct {
			if valueMap, ok := value.(map[string]interface{}); ok {
				// Recursively set fields of the nested struct
				if err := setStructFields(field, valueMap); err != nil {
					return fmt.Errorf("error setting nested struct '%s': %v", fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected map for nested struct '%s', got %T", fieldType.Name, value)
			}
		} else if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			if valueMap, ok := value.(map[string]interface{}); ok {
				// Initialize the pointer to a new struct if nil
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
				// Recursively set fields of the nested struct
				if err := setStructFields(field.Elem(), valueMap); err != nil {
					return fmt.Errorf("error setting nested struct '%s': %v", fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected map for nested struct pointer '%s', got %T", fieldType.Name, value)
			}
		} else {
			// Set the field value
			if valueStr, ok := value.(string); ok {
				if err := setFieldValue(field, valueStr); err != nil {
					return fmt.Errorf("error setting field '%s': %v", fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected string value for field '%s', got %T", fieldType.Name, value)
			}
		}
	}

	return nil
}

// setFieldValue sets a single field value based on the provided string.
func setFieldValue(field reflect.Value, valueStr string) error {
	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	// Trim whitespace from valueStr
	valueStr = strings.TrimSpace(valueStr)

	switch field.Kind() {
	case reflect.String:
		field.SetString(valueStr)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(valueStr)
		if err != nil {
			return fmt.Errorf("invalid boolean value '%s' for field", valueStr)
		}
		field.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(valueStr, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer value '%s' for field", valueStr)
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(valueStr, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value '%s' for field", valueStr)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(valueStr, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value '%s' for field", valueStr)
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
