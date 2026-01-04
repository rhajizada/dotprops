package dotprops

import (
	"bufio"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// parseProperties reads properties data and returns a map of key-value pairs.
func parseProperties(data []byte) (map[string]any, error) {
	props := make(map[string]any)
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
			for i := range len(keyList) - 1 {
				k := keyList[i]
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]any)
				}
				// Type assertion to navigate deeper into the nested map
				if nextMap, ok := current[k].(map[string]any); ok {
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
func getNestedProperty(props map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	var current any = props
	for _, part := range parts {
		switch currMap := current.(type) {
		case map[string]any:
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
func setStructFields(structVal reflect.Value, props map[string]any) error {
	structType := structVal.Type()

	for i := range structVal.NumField() {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		if err := setStructField(field, fieldType, props); err != nil {
			return err
		}
	}

	return nil
}

func setStructField(field reflect.Value, fieldType reflect.StructField, props map[string]any) error {
	if fieldType.Anonymous {
		return setEmbeddedStructField(field, props)
	}

	propertyKey := fieldType.Tag.Get("property")
	if propertyKey == "" {
		propertyKey = fieldType.Name
	}

	value, ok := getNestedProperty(props, propertyKey)
	if !ok {
		return nil
	}

	return applyFieldValue(field, propertyKey, value)
}

func setEmbeddedStructField(field reflect.Value, props map[string]any) error {
	switch field.Kind() {
	case reflect.Struct:
		return setStructFields(field, props)
	case reflect.Ptr:
		if field.Type().Elem().Kind() != reflect.Struct {
			return nil
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setStructFields(field.Elem(), props)
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.String,
		reflect.UnsafePointer:
		return nil
	}
	return nil
}

func applyFieldValue(field reflect.Value, propertyKey string, value any) error {
	if handled, err := applyPropUnmarshaller(field, propertyKey, value); handled {
		return err
	}
	if handled, err := applyNestedStruct(field, propertyKey, value); handled {
		return err
	}
	if handled, err := applyTextUnmarshaller(field, propertyKey, value); handled {
		return err
	}

	valueStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value for field '%s', got %T", propertyKey, value)
	}

	if err := setFieldValue(field, valueStr); err != nil {
		return fmt.Errorf("error setting field '%s': %w", propertyKey, err)
	}
	return nil
}

func applyPropUnmarshaller(field reflect.Value, propertyKey string, value any) (bool, error) {
	propUnmarshallerType := reflect.TypeFor[PropUnmarshaller]()
	target, ok := interfaceValue(field, propUnmarshallerType)
	if !ok {
		return false, nil
	}

	unmarshaler, ok := target.(PropUnmarshaller)
	if !ok {
		return false, nil
	}

	key, valStr, extractErr := extractKeyValue(propertyKey, value)
	if extractErr != nil {
		return true, fmt.Errorf("error extracting key-value for field '%s': %w", propertyKey, extractErr)
	}

	if unmarshalErr := unmarshaler.UnmarshalProp(key, valStr); unmarshalErr != nil {
		return true, fmt.Errorf("error unmarshaling field '%s': %w", propertyKey, unmarshalErr)
	}
	return true, nil
}

func applyNestedStruct(field reflect.Value, propertyKey string, value any) (bool, error) {
	switch field.Kind() {
	case reflect.Struct:
		subProps, ok := value.(map[string]any)
		if !ok {
			return true, fmt.Errorf("expected map for nested struct field '%s', got %T", propertyKey, value)
		}
		return true, setStructFields(field, subProps)
	case reflect.Ptr:
		if field.Type().Elem().Kind() != reflect.Struct {
			return false, nil
		}
		subProps, ok := value.(map[string]any)
		if !ok {
			return true, fmt.Errorf("expected map for nested struct pointer field '%s', got %T", propertyKey, value)
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return true, setStructFields(field.Elem(), subProps)
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.String,
		reflect.UnsafePointer:
		return false, nil
	}
	return false, nil
}

func applyTextUnmarshaller(field reflect.Value, propertyKey string, value any) (bool, error) {
	textUnmarshallerType := reflect.TypeFor[TextUnmarshaler]()
	target, ok := interfaceValue(field, textUnmarshallerType)
	if !ok {
		return false, nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return true, fmt.Errorf("expected string value for field '%s', got %T", propertyKey, value)
	}

	unmarshaler, ok := target.(TextUnmarshaler)
	if !ok {
		return false, nil
	}

	if err := unmarshaler.UnmarshalText([]byte(valueStr)); err != nil {
		return true, fmt.Errorf("error unmarshaling field '%s': %w", propertyKey, err)
	}
	return true, nil
}

func interfaceValue(field reflect.Value, iface reflect.Type) (any, bool) {
	if !field.IsValid() {
		return nil, false
	}

	if field.Kind() == reflect.Ptr {
		if !field.Type().Implements(iface) {
			return nil, false
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		if field.CanInterface() {
			return field.Interface(), true
		}
		return nil, false
	}

	if field.Type().Implements(iface) && field.CanInterface() {
		return field.Interface(), true
	}

	if field.CanAddr() && field.Addr().Type().Implements(iface) {
		return field.Addr().Interface(), true
	}

	return nil, false
}

// Helper function to extract key-value pair for PropUnmarshaler.
func extractKeyValue(propertyKey string, value any) (string, string, error) {
	valueStr, ok := value.(string)
	if !ok {
		return "", "", fmt.Errorf("expected string value for property '%s', got %T", propertyKey, value)
	}
	return propertyKey, valueStr, nil
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
	case reflect.Uintptr:
		uintVal, err := strconv.ParseUint(valueStr, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value '%s' for field", valueStr)
		}
		field.SetUint(uintVal)
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
		reflect.Struct,
		reflect.UnsafePointer:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
