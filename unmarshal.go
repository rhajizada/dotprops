package dotprops

import (
	"errors"
	"reflect"
)

func Unmarshal(data []byte, v interface{}) error {
	val := reflect.ValueOf(v)

	// Ensure v is a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("unmarshal expects a pointer to a struct")
	}

	// Parse the properties
	props, err := parseProperties(data)
	if err != nil {
		return err
	}

	// Set the struct fields
	return setStructFields(val.Elem(), props)
}
