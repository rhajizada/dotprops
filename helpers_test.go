package dotprops

import (
	"errors"
	"fmt"
	"strconv"
)

// Shared Test Structures

type SimpleConfig struct {
	AppName string `property:"app.name"`
	Port    int    `property:"app.port"`
	Debug   bool   `property:"app.debug"`
}

type NestedConfig struct {
	AppName  string         `property:"app.name"`
	Database DatabaseConfig `property:"database"`
}

type DatabaseConfig struct {
	Host     string `property:"host"`
	Port     int    `property:"port"`
	Username string `property:"username"`
	Password string `property:"password"`
}

type OptionalConfig struct {
	AppName *string `property:"app.name"`
	Port    *int    `property:"app.port"`
	Debug   *bool   `property:"app.debug"`
}

type ConfigWithPointer struct {
	AppName  string          `property:"app.name"`
	Database *DatabaseConfig `property:"database"`
}

// Custom Types Implementing Interfaces for Testing

// CustomString implements TextMarshaler and TextUnmarshaler
type CustomString string

func (cs CustomString) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("custom_%s", cs)), nil
}

func (cs *CustomString) UnmarshalText(text []byte) error {
	if len(text) < 8 || string(text[:7]) != "custom_" {
		return errors.New("invalid prefix")
	}
	*cs = CustomString(string(text[7:]))
	return nil
}

// CustomInt implements TextMarshaler and TextUnmarshaler
type CustomInt int

func (ci CustomInt) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("custom_%d", ci)), nil
}

func (ci *CustomInt) UnmarshalText(text []byte) error {
	if len(text) < 7 || string(text[:7]) != "custom_" {
		return errors.New("invalid prefix")
	}
	numStr := string(text[7:])
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return err
	}
	*ci = CustomInt(num)
	return nil
}
