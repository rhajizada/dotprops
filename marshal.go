package jprops

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Marshal returns the properties encoding of v.
// v must be a struct.
func Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil, errors.New("marshal expects a struct")
	}

	var sb strings.Builder
	structType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := structType.Field(i)
		propertyKey := fieldType.Tag.Get("property")

		if propertyKey == "" {
			continue
		}

		var valueStr string

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue // Skip nil pointers
			}
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.String:
			valueStr = field.String()
		case reflect.Bool:
			valueStr = strconv.FormatBool(field.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueStr = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			valueStr = strconv.FormatUint(field.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			valueStr = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		default:
			continue // Skip unsupported types
		}

		sb.WriteString(fmt.Sprintf("%s=%s\n", propertyKey, valueStr))
	}

	return []byte(sb.String()), nil
}
