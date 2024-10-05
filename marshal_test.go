package jprops

import (
	"testing"
)

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

func TestMarshalSimple(t *testing.T) {
	config := SimpleConfig{
		AppName: "TestApp",
		Port:    3000,
		Debug:   false,
	}

	expected := "app.debug=false\napp.name=TestApp\napp.port=3000\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalNested(t *testing.T) {
	config := NestedConfig{
		AppName: "MyApp",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "admin",
			Password: "secret",
		},
	}

	expected := "app.name=MyApp\ndatabase.host=localhost\ndatabase.password=secret\ndatabase.port=5432\ndatabase.username=admin\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalOptionalFields(t *testing.T) {
	appName := "MyApp"
	debug := true
	config := OptionalConfig{
		AppName: &appName,
		Port:    nil, // Optional field not set
		Debug:   &debug,
	}

	expected := "app.debug=true\napp.name=MyApp\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalUnsupportedType(t *testing.T) {
	type UnsupportedConfig struct {
		Data []string `property:"data"`
	}

	config := UnsupportedConfig{
		Data: []string{"one", "two", "three"},
	}

	_, err := Marshal(config)
	if err == nil {
		t.Fatal("Expected error for unsupported type, got nil")
	}
}

func TestMarshalNilPointer(t *testing.T) {
	config := OptionalConfig{
		AppName: nil,
		Port:    nil,
		Debug:   nil,
	}

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty string, got:\n%s", data)
	}
}

func TestMarshalEmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	config := EmptyStruct{}

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty string, got:\n%s", data)
	}
}

func TestMarshalNonStruct(t *testing.T) {
	nonStruct := "I am not a struct"

	_, err := Marshal(nonStruct)
	if err == nil {
		t.Fatal("Expected error for non-struct input, got nil")
	}
}
