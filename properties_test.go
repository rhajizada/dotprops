package jprops

import (
	"reflect"
	"testing"
)

// TestParseProperties tests the parseProperties function to ensure it correctly parses the properties data.
func TestParseProperties(t *testing.T) {
	data := []byte(`
# This is a comment
key1=value1
key2=value2

key3.subkey1=value3
key3.subkey2=value4
`)

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]interface{}{
			"subkey1": "value3",
			"subkey2": "value4",
		},
	}

	props, err := parseProperties(data)
	if err != nil {
		t.Fatalf("parseProperties failed: %v", err)
	}

	if !reflect.DeepEqual(props, expected) {
		t.Errorf("Expected props to be %+v, got %+v", expected, props)
	}
}

// TestParsePropertiesInvalidLine ensures that invalid lines are ignored or handled appropriately.
func TestParsePropertiesInvalidLine(t *testing.T) {
	data := []byte(`
key1=value1
invalid_line_without_equals
key2=value2
`)

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	props, err := parseProperties(data)
	if err != nil {
		t.Fatalf("parseProperties failed: %v", err)
	}

	if !reflect.DeepEqual(props, expected) {
		t.Errorf("Expected props to be %+v, got %+v", expected, props)
	}
}

// TestSetStructFields_Simple tests setStructFields with a simple struct and correct property values.
func TestSetStructFields_Simple(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Age  int    `property:"age"`
	}

	props := map[string]interface{}{
		"name": "Alice",
		"age":  "30",
	}

	var config Config
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err != nil {
		t.Fatalf("setStructFields failed: %v", err)
	}

	if config.Name != "Alice" {
		t.Errorf("Expected Name 'Alice', got '%s'", config.Name)
	}
	if config.Age != 30 {
		t.Errorf("Expected Age 30, got %d", config.Age)
	}
}

// TestSetStructFields_Nested tests setStructFields with a nested struct and correct property values.
func TestSetStructFields_Nested(t *testing.T) {
	type InnerConfig struct {
		SubName string `property:"sub.name"`
		Value   int    `property:"value"`
	}

	type OuterConfig struct {
		Name  string      `property:"name"`
		Inner InnerConfig `property:"inner"`
	}

	props := map[string]interface{}{
		"name": "Outer",
		"inner": map[string]interface{}{
			"sub": map[string]interface{}{
				"name": "Inner",
			},
			"value": "100",
		},
	}

	var config OuterConfig
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err != nil {
		t.Fatalf("setStructFields failed: %v", err)
	}

	if config.Name != "Outer" {
		t.Errorf("Expected Name 'Outer', got '%s'", config.Name)
	}
	if config.Inner.SubName != "Inner" {
		t.Errorf("Expected Inner.SubName 'Inner', got '%s'", config.Inner.SubName)
	}
	if config.Inner.Value != 100 {
		t.Errorf("Expected Inner.Value 100, got %d", config.Inner.Value)
	}
}

// TestSetStructFields_PointerNested tests setStructFields with a pointer to a nested struct.
func TestSetStructFields_PointerNested(t *testing.T) {
	type InnerConfig struct {
		SubName string `property:"sub.name"`
		Value   int    `property:"value"`
	}

	type OuterConfig struct {
		Name  string       `property:"name"`
		Inner *InnerConfig `property:"inner"`
	}

	props := map[string]interface{}{
		"name": "Outer",
		"inner": map[string]interface{}{
			"sub": map[string]interface{}{
				"name": "Inner",
			},
			"value": "100",
		},
	}

	var config OuterConfig
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err != nil {
		t.Fatalf("setStructFields failed: %v", err)
	}

	if config.Name != "Outer" {
		t.Errorf("Expected Name 'Outer', got '%s'", config.Name)
	}
	if config.Inner == nil {
		t.Fatal("Expected Inner to be initialized, got nil")
	}
	if config.Inner.SubName != "Inner" {
		t.Errorf("Expected Inner.SubName 'Inner', got '%s'", config.Inner.SubName)
	}
	if config.Inner.Value != 100 {
		t.Errorf("Expected Inner.Value 100, got %d", config.Inner.Value)
	}
}

// TestSetStructFields_TypeMismatch tests setStructFields with a type mismatch.
func TestSetStructFields_TypeMismatch(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Age  int    `property:"age"`
	}

	props := map[string]interface{}{
		"name": "Bob",
		"age":  "not_an_int",
	}

	var config Config
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err == nil {
		t.Fatal("Expected setStructFields to fail due to type mismatch, but it did not")
	}

	// Since 'age' couldn't be set due to type mismatch, it should remain at zero value
	if config.Age != 0 {
		t.Errorf("Expected Age to be 0, got %d", config.Age)
	}
}

// TestSetStructFields_UnsupportedType tests setStructFields with an unsupported field type.
func TestSetStructFields_UnsupportedType(t *testing.T) {
	type Config struct {
		Channel chan int `property:"channel"`
	}

	props := map[string]interface{}{
		"channel": "data",
	}

	var config Config
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err == nil {
		t.Fatal("Expected setStructFields to fail due to unsupported field type, but it did not")
	}
}

// TestSetStructFields_UnsupportedNestedType tests setStructFields with an unsupported nested field type.
func TestSetStructFields_UnsupportedNestedType(t *testing.T) {
	type InnerConfig struct {
		Channel chan int `property:"channel"`
	}

	type OuterConfig struct {
		Name  string      `property:"name"`
		Inner InnerConfig `property:"inner"`
	}

	props := map[string]interface{}{
		"name": "Outer",
		"inner": map[string]interface{}{
			"channel": "data",
		},
	}

	var config OuterConfig
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err == nil {
		t.Fatal("Expected setStructFields to fail due to unsupported nested field type, but it did not")
	}
}

// TestSetStructFields_PartialData tests setStructFields with partial data (missing some fields).
func TestSetStructFields_PartialData(t *testing.T) {
	type Config struct {
		Name    string `property:"name"`
		Age     int    `property:"age"`
		Address string `property:"address"`
	}

	props := map[string]interface{}{
		"name": "Charlie",
		"age":  "25",
		// 'address' is missing
	}

	var config Config
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err != nil {
		t.Fatalf("setStructFields failed: %v", err)
	}

	if config.Name != "Charlie" {
		t.Errorf("Expected Name 'Charlie', got '%s'", config.Name)
	}
	if config.Age != 25 {
		t.Errorf("Expected Age 25, got %d", config.Age)
	}
	if config.Address != "" {
		t.Errorf("Expected Address '', got '%s'", config.Address)
	}
}

// TestSetStructFields_ExtraProperties tests setStructFields with extra properties not present in the struct.
func TestSetStructFields_ExtraProperties(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
	}

	props := map[string]interface{}{
		"name":      "Dana",
		"unknown":   "value",
		"another":   "value",
		"extra.key": "extra_value",
	}

	var config Config
	err := setStructFields(reflect.ValueOf(&config).Elem(), props)
	if err != nil {
		t.Fatalf("setStructFields failed: %v", err)
	}

	if config.Name != "Dana" {
		t.Errorf("Expected Name 'Dana', got '%s'", config.Name)
	}
}
