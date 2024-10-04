package jprops

import (
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal parses the properties-encoded data and stores the result
// in the value pointed to by v, which must be a pointer to a struct.
func Unmarshal(data []byte, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("unmarshal(non-struct pointer)")
	}

	structVal := val.Elem()
	structType := structVal.Type()

	// Check for nested structs
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i).Type

		// Handle pointer fields
		fieldKind := fieldType.Kind()
		if fieldKind == reflect.Ptr {
			fieldType = fieldType.Elem()
			fieldKind = fieldType.Kind()
		}

		if fieldKind == reflect.Struct {
			return fmt.Errorf("unmarshal(nested structs are not supported): field '%s' is a nested struct", structType.Field(i).Name)
		}
	}

	// Parse the data into key-value pairs
	props, err := parseProperties(data)
	if err != nil {
		return err
	}

	// Set struct fields based on the properties
	return setStructFields(val.Elem(), props)
}
