package jprops

import (
	"errors"
	"reflect"
)

// Unmarshal parses the properties-encoded data and stores the result
// in the value pointed to by v, which must be a pointer to a struct.

// Unmarshal parses the properties-encoded data and stores the result
// in the value pointed to by v, which must be a pointer to a struct.
func Unmarshal(data []byte, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("unmarshal(non-struct pointer)")
	}

	// Parse the data into key-value pairs
	props, err := parseProperties(data)
	if err != nil {
		return err
	}

	// Set struct fields based on the properties
	return setStructFields(val.Elem(), props)
}
