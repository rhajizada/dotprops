package jprops

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Marshal returns the properties encoding of v.
// v must be a struct.
func Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshal expects a struct")
	}

	props := make(map[string]string)
	err := encodeStruct("", val, props)
	if err != nil {
		return nil, err
	}

	// Sort the keys for consistent output
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the properties string
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, props[k]))
	}

	return []byte(sb.String()), nil
}

// encodeStruct encodes a struct into the props map with proper key prefixes
func encodeStruct(prefix string, val reflect.Value, props map[string]string) error {
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := valType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the property key from the struct tag or use the field name
		propertyKey := fieldType.Tag.Get("property")
		if propertyKey == "" {
			propertyKey = fieldType.Name
		}

		// Build the full key
		fullKey := propertyKey
		if prefix != "" {
			fullKey = prefix + "." + propertyKey
		}

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue // Skip nil pointers
			}
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			// Recursively encode nested structs
			err := encodeStruct(fullKey, field, props)
			if err != nil {
				return err
			}
		case reflect.String:
			props[fullKey] = field.String()
		case reflect.Bool:
			props[fullKey] = fmt.Sprintf("%v", field.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			props[fullKey] = fmt.Sprintf("%d", field.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			props[fullKey] = fmt.Sprintf("%d", field.Uint())
		case reflect.Float32, reflect.Float64:
			props[fullKey] = fmt.Sprintf("%f", field.Float())
		default:
			return fmt.Errorf("unsupported field type: %s for field %s", field.Kind(), fullKey)
		}
	}

	return nil
}
