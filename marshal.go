package dotprops

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Marshal returns the properties encoding of v.
// v must be a struct or a pointer to a struct.
func Marshal(v any) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.Elem().Kind() != reflect.Struct {
			return nil, errors.New("marshal expects a pointer to a struct")
		}
		val = val.Elem()
	} else if val.Kind() != reflect.Struct {
		return nil, errors.New("marshal expects a struct or a pointer to a struct")
	}

	// Ensure the value is addressable
	if !val.CanAddr() {
		return nil, errors.New("marshal requires an addressable struct to handle TextMarshaler")
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

// encodeStruct encodes a struct into the props map with proper key prefixes.
func encodeStruct(prefix string, val reflect.Value, props map[string]string) error {
	valType := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := valType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fullKey, _ := fieldKey(prefix, fieldType)

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue // Skip nil pointers
			}
			field = field.Elem()
		}

		if handled, err := marshalPropField(field, fullKey, props); handled {
			if err != nil {
				return err
			}
			continue
		}

		if handled, err := marshalTextField(field, fullKey, props); handled {
			if err != nil {
				return err
			}
			continue
		}

		if err := encodeFieldValue(fullKey, field, props); err != nil {
			return err
		}
	}

	return nil
}

func fieldKey(prefix string, fieldType reflect.StructField) (string, bool) {
	isEmbedded := fieldType.Anonymous
	propertyKey := fieldType.Tag.Get("property")
	if propertyKey == "" && !isEmbedded {
		propertyKey = fieldType.Name
	}

	switch {
	case isEmbedded:
		return prefix, true
	case prefix != "":
		return prefix + "." + propertyKey, false
	default:
		return propertyKey, false
	}
}

func marshalPropField(field reflect.Value, fullKey string, props map[string]string) (bool, error) {
	if !field.CanInterface() {
		return false, nil
	}

	if pm, ok := field.Addr().Interface().(PropMarshaler); ok {
		key, value, err := pm.MarshalProp()
		if err != nil {
			return true, fmt.Errorf("error marshaling field '%s': %w", fullKey, err)
		}
		props[key] = value
		return true, nil
	}

	return false, nil
}

func marshalTextField(field reflect.Value, fullKey string, props map[string]string) (bool, error) {
	if !field.CanInterface() {
		return false, nil
	}

	if marshaler, ok := field.Addr().Interface().(TextMarshaler); ok {
		text, err := marshaler.MarshalText()
		if err != nil {
			return true, fmt.Errorf("error marshaling field '%s': %w", fullKey, err)
		}
		props[fullKey] = string(text)
		return true, nil
	}

	return false, nil
}

func encodeFieldValue(fullKey string, field reflect.Value, props map[string]string) error {
	switch field.Kind() {
	case reflect.Struct:
		return encodeStruct(fullKey, field, props)
	case reflect.String:
		props[fullKey] = field.String()
	case reflect.Bool:
		props[fullKey] = strconv.FormatBool(field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		props[fullKey] = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		props[fullKey] = strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		props[fullKey] = fmt.Sprintf("%f", field.Float())
	case reflect.Invalid,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.UnsafePointer:
		return fmt.Errorf("unsupported field type: %s for field %s", field.Kind(), fullKey)
	}
	return nil
}
