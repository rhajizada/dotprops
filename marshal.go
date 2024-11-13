package dotprops

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// TextMarshaler interface as defined in the encoding package
type TextMarshaler interface {
	MarshalText() (text []byte, err error)
}

// Marshal returns the properties encoding of v.
// v must be a struct.
// Marshal returns the properties encoding of v.
// v must be a struct or a pointer to a struct.
func Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("marshal expects a pointer to a struct")
		}
		val = val.Elem()
	} else if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshal expects a struct or a pointer to a struct")
	}

	// Ensure the value is addressable
	if !val.CanAddr() {
		return nil, fmt.Errorf("marshal requires an addressable struct to handle TextMarshaler")
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

		// Check if the field is embedded
		isEmbedded := fieldType.Anonymous

		// Get the property key from the struct tag or use the field name
		propertyKey := fieldType.Tag.Get("property")
		if propertyKey == "" && !isEmbedded {
			propertyKey = fieldType.Name
		}

		var fullKey string
		if isEmbedded {
			// Do not add propertyKey as prefix; use the current prefix
			fullKey = prefix
		} else if prefix != "" {
			fullKey = prefix + "." + propertyKey
		} else {
			fullKey = propertyKey
		}

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue // Skip nil pointers
			}
			field = field.Elem()
		}

		// Check if the field implements TextMarshaler
		if field.CanInterface() {
			if marshaler, ok := field.Addr().Interface().(TextMarshaler); ok {
				text, err := marshaler.MarshalText()
				if err != nil {
					return fmt.Errorf("error marshaling field '%s': %v", fullKey, err)
				}
				props[fullKey] = string(text)
				continue
			}
		}

		switch field.Kind() {
		case reflect.Struct:
			if isEmbedded {
				// For embedded structs, continue with the same prefix
				err := encodeStruct(fullKey, field, props)
				if err != nil {
					return err
				}
			} else {
				// For nested structs, use the new prefix
				err := encodeStruct(fullKey, field, props)
				if err != nil {
					return err
				}
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
