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
func parseProperties(data []byte) (map[string]string, error) {
	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	pattern := regexp.MustCompile("^([^#][^=]*)=(.*)")

	for scanner.Scan() {
		line := scanner.Text()
		match := pattern.FindStringSubmatch(line)
		if len(match) > 0 {
			key := match[1]
			value := match[2]
			props[key] = value
		} else {
			continue
		}
	}

	return props, scanner.Err()
}

// setStructFields sets the fields of the struct based on the provided properties.
func setStructFields(structVal reflect.Value, props map[string]string) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)
		propertyKey := fieldType.Tag.Get("property")

		if propertyKey == "" {
			continue
		}

		valueStr, ok := props[propertyKey]
		if !ok {
			continue // Property not found in data
		}

		// Set the field value
		if err := setFieldValue(field, valueStr); err != nil {
			return fmt.Errorf("error setting field '%s': %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue sets a single field value based on the provided string.

// setFieldValue sets a single field value based on the provided string.
func setFieldValue(field reflect.Value, valueStr string) error {
	// Handle pointer types by dereferencing to the base type
	isPointer := false
	if field.Kind() == reflect.Ptr {
		isPointer = true
		if field.IsNil() {
			// Initialize the pointer to a new value unless valueStr is empty
			if valueStr != "" {
				field.Set(reflect.New(field.Type().Elem()))
			} else {
				// If valueStr is empty, keep the pointer as nil
				return nil
			}
		}
		field = field.Elem()
	}

	// Trim whitespace from valueStr
	valueStr = strings.TrimSpace(valueStr)

	// If valueStr is empty, set the field to its zero value or nil for pointers
	if valueStr == "" {
		if isPointer {
			// Set pointer field to nil
			fieldAddr := field.Addr()
			fieldAddr.Set(reflect.Zero(fieldAddr.Type()))
		} else {
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

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
