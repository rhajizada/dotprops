package dotprops

import (
	"errors"
	"testing"
)

func TestMarshalSimple(t *testing.T) {
	config := &SimpleConfig{ // Pass a pointer
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
	config := &NestedConfig{ // Pass a pointer
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
	config := &OptionalConfig{
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
	config := &OptionalConfig{
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

	config := &EmptyStruct{}

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

// New Tests for TextMarshaler Interface

func TestMarshalWithTextMarshaler(t *testing.T) {
	type CustomConfig struct {
		Name  CustomString `property:"name"`
		Count CustomInt    `property:"count"`
	}

	config := &CustomConfig{
		Name:  "example",
		Count: 42,
	}

	expected := "count=custom_42\nname=custom_example\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

type FaultyCustomString string

// Implement TextMarshaler that returns an error
func (fcs FaultyCustomString) MarshalText() ([]byte, error) {
	return nil, errors.New("marshal error")
}

type ErrorConfig struct {
	Name FaultyCustomString `property:"name"`
}

func TestMarshalWithTextMarshalerError(t *testing.T) {
	config := &ErrorConfig{
		Name: "faulty",
	}
	_, err := Marshal(config)
	if err == nil {
		t.Fatal("Expected Marshal to fail due to TextMarshaler error, but it did not")
	}
}

// Additional Tests to Increase Coverage

func TestMarshalWithUintAndFloatFields(t *testing.T) {
	type NumericConfig struct {
		MaxUsers  uint    `property:"max.users"`
		Threshold float64 `property:"threshold"`
	}

	config := &NumericConfig{
		MaxUsers:  1000,
		Threshold: 75.5,
	}

	expected := "max.users=1000\nthreshold=75.500000\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithMultiLevelNestedStruct(t *testing.T) {
	config := &MultiLevelNestedConfig{
		Service: ServiceConfig{
			Name: "AuthService",
			Endpoint: EndpointConfig{
				URL:    "https://auth.example.com",
				Port:   443,
				Active: true,
			},
		},
	}

	expected := "service.endpoint.active=true\nservice.endpoint.port=443\nservice.endpoint.url=https://auth.example.com\nservice.name=AuthService\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithEmbeddedStruct(t *testing.T) {
	type BaseConfig struct {
		Version string `property:"version"`
	}

	type EmbeddedConfig struct {
		BaseConfig
		Name string `property:"name"`
	}

	config := &EmbeddedConfig{
		BaseConfig: BaseConfig{
			Version: "1.0.0",
		},
		Name: "EmbeddedService",
	}

	expected := "name=EmbeddedService\nversion=1.0.0\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithUnsupportedNestedStruct(t *testing.T) {
	type InnerUnsupported struct {
		Data []string `property:"data"`
	}

	type OuterConfig struct {
		Name  string           `property:"name"`
		Inner InnerUnsupported `property:"inner"`
	}

	config := &OuterConfig{
		Name: "Outer",
		Inner: InnerUnsupported{
			Data: []string{"one", "two"},
		},
	}

	_, err := Marshal(config)
	if err == nil {
		t.Fatal("Expected Marshal to fail due to unsupported nested struct field type, but it did not")
	}
}

func TestMarshalWithMultipleEmbeddedStructs(t *testing.T) {
	type BaseConfig struct {
		Version string `property:"version"`
	}
	type SecurityConfig struct {
		Enabled bool `property:"enabled"`
	}
	type EmbeddedConfig struct {
		BaseConfig
		SecurityConfig
		Name string `property:"name"`
	}

	config := &EmbeddedConfig{
		BaseConfig: BaseConfig{
			Version: "2.0.0",
		},
		SecurityConfig: SecurityConfig{
			Enabled: true,
		},
		Name: "MultiEmbeddedService",
	}

	expected := "enabled=true\nname=MultiEmbeddedService\nversion=2.0.0\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithPointerEmbeddedStructs(t *testing.T) {
	type BaseConfig struct {
		Version string `property:"version"`
	}
	type EmbeddedConfig struct {
		*BaseConfig
		Name string `property:"name"`
	}

	config := &EmbeddedConfig{
		BaseConfig: &BaseConfig{
			Version: "3.1.4",
		},
		Name: "PointerEmbeddedService",
	}

	expected := "name=PointerEmbeddedService\nversion=3.1.4\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}
func TestMarshalWithCustomTextMarshaler(t *testing.T) {
	type CustomConfig struct {
		Name  CustomString `property:"name"`
		Count CustomInt    `property:"count"`
	}

	config := &CustomConfig{
		Name:  "example", // Changed from "custom_example" to "example"
		Count: 100,
	}

	expected := "count=custom_100\nname=custom_example\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithUnsupportedNestedStructTypes(t *testing.T) {
	type InnerUnsupported struct {
		Data []string `property:"data"`
	}
	type OuterConfig struct {
		Name  string           `property:"name"`
		Inner InnerUnsupported `property:"inner"`
	}

	config := &OuterConfig{
		Name: "OuterService",
		Inner: InnerUnsupported{
			Data: []string{"one", "two"},
		},
	}

	_, err := Marshal(config)
	if err == nil {
		t.Fatal("Expected Marshal to fail due to unsupported nested struct field type, but it did not")
	}
}

func TestMarshalWithMultipleLevelsNestedStructs(t *testing.T) {
	type Level3 struct {
		Key string `property:"key"`
	}
	type Level2 struct {
		Level3 Level3 `property:"level3"`
	}
	type Level1 struct {
		Level2 Level2 `property:"level2"`
	}
	type OuterConfig struct {
		Name   string `property:"name"`
		Level1 Level1 `property:"level1"`
		Active bool   `property:"active"`
	}

	config := &OuterConfig{
		Name: "DeepNestedService",
		Level1: Level1{
			Level2: Level2{
				Level3: Level3{
					Key: "deep_value",
				},
			},
		},
		Active: true,
	}

	expected := "active=true\nlevel1.level2.level3.key=deep_value\nname=DeepNestedService\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithMissingEmbeddedStructFields(t *testing.T) {
	type BaseConfig struct {
		Version string `property:"version"`
	}
	type EmbeddedConfig struct {
		BaseConfig
		Name string `property:"name"`
	}

	config := &EmbeddedConfig{
		BaseConfig: BaseConfig{
			Version: "", // Zero value
		},
		Name: "EmbeddedService",
	}

	expected := "name=EmbeddedService\nversion=\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}

func TestMarshalWithExtraProperties(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Age  int    `property:"age"`
		// Extra field without a property tag should be ignored
		unexportedField string
	}

	config := &Config{
		Name:            "ExtraService",
		Age:             45,
		unexportedField: "should_be_ignored",
	}

	expected := "age=45\nname=ExtraService\n"

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, data)
	}
}
