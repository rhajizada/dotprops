package dotprops_test

import (
	"testing"

	"github.com/rhajizada/dotprops"
)

// TestParseProperties exercises the parser through Unmarshal with nested keys.
func TestParseProperties(t *testing.T) {
	data := []byte(`
# This is a comment
key1=value1
key2=value2

key3.subkey1=value3
key3.subkey2=value4
`)

	type Key3Config struct {
		Subkey1 string `property:"subkey1"`
		Subkey2 string `property:"subkey2"`
	}
	type Config struct {
		Key1 string     `property:"key1"`
		Key2 string     `property:"key2"`
		Key3 Key3Config `property:"key3"`
	}

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Key1 != "value1" {
		t.Errorf("Expected Key1 'value1', got '%s'", config.Key1)
	}
	if config.Key2 != "value2" {
		t.Errorf("Expected Key2 'value2', got '%s'", config.Key2)
	}
	if config.Key3.Subkey1 != "value3" {
		t.Errorf("Expected Key3.Subkey1 'value3', got '%s'", config.Key3.Subkey1)
	}
	if config.Key3.Subkey2 != "value4" {
		t.Errorf("Expected Key3.Subkey2 'value4', got '%s'", config.Key3.Subkey2)
	}
}

// TestParsePropertiesInvalidLine ensures invalid lines are ignored.
func TestParsePropertiesInvalidLine(t *testing.T) {
	data := []byte(`
key1=value1
invalid_line_without_equals
key2=value2
`)

	type Config struct {
		Key1 string `property:"key1"`
		Key2 string `property:"key2"`
	}

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Key1 != "value1" {
		t.Errorf("Expected Key1 'value1', got '%s'", config.Key1)
	}
	if config.Key2 != "value2" {
		t.Errorf("Expected Key2 'value2', got '%s'", config.Key2)
	}
}

// TestSetStructFields_Simple tests Unmarshal with a simple struct and correct property values.
func TestSetStructFields_Simple(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Age  int    `property:"age"`
	}

	data := []byte("name=Alice\nage=30\n")

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "Alice" {
		t.Errorf("Expected Name 'Alice', got '%s'", config.Name)
	}
	if config.Age != 30 {
		t.Errorf("Expected Age 30, got %d", config.Age)
	}
}

// TestSetStructFields_Nested tests Unmarshal with nested structs and correct property values.
func TestSetStructFields_Nested(t *testing.T) {
	type InnerConfig struct {
		SubName string `property:"sub.name"`
		Value   int    `property:"value"`
	}

	type OuterConfig struct {
		Name  string      `property:"name"`
		Inner InnerConfig `property:"inner"`
	}

	data := []byte("name=Outer\ninner.sub.name=Inner\ninner.value=100\n")

	var config OuterConfig
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
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

// TestSetStructFields_PointerNested tests Unmarshal with a pointer to a nested struct.
func TestSetStructFields_PointerNested(t *testing.T) {
	type InnerConfig struct {
		SubName string `property:"sub.name"`
		Value   int    `property:"value"`
	}

	type OuterConfig struct {
		Name  string       `property:"name"`
		Inner *InnerConfig `property:"inner"`
	}

	data := []byte("name=Outer\ninner.sub.name=Inner\ninner.value=100\n")

	var config OuterConfig
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
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

// TestSetStructFields_TypeMismatch tests Unmarshal with a type mismatch.
func TestSetStructFields_TypeMismatch(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Age  int    `property:"age"`
	}

	data := []byte("name=Bob\nage=not_an_int\n")

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to type mismatch, but it did not")
	}

	if config.Age != 0 {
		t.Errorf("Expected Age to be 0, got %d", config.Age)
	}
}

// TestSetStructFields_UnsupportedType tests Unmarshal with an unsupported field type.
func TestSetStructFields_UnsupportedType(t *testing.T) {
	type Config struct {
		Channel chan int `property:"channel"`
	}

	data := []byte("channel=data\n")

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to unsupported field type, but it did not")
	}
}

// TestSetStructFields_UnsupportedNestedType tests Unmarshal with an unsupported nested field type.
func TestSetStructFields_UnsupportedNestedType(t *testing.T) {
	type InnerConfig struct {
		Channel chan int `property:"channel"`
	}

	type OuterConfig struct {
		Name  string      `property:"name"`
		Inner InnerConfig `property:"inner"`
	}

	data := []byte("name=Outer\ninner.channel=data\n")

	var config OuterConfig
	err := dotprops.Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to unsupported nested field type, but it did not")
	}
}

// TestSetStructFields_PartialData tests Unmarshal with partial data (missing some fields).
func TestSetStructFields_PartialData(t *testing.T) {
	type Config struct {
		Name    string `property:"name"`
		Age     int    `property:"age"`
		Address string `property:"address"`
	}

	data := []byte("name=Charlie\nage=25\n")

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
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

// TestSetStructFields_ExtraProperties tests Unmarshal with extra properties not present in the struct.
func TestSetStructFields_ExtraProperties(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
	}

	data := []byte("name=Dana\nunknown=value\nanother=value\nextra.key=extra_value\n")

	var config Config
	err := dotprops.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "Dana" {
		t.Errorf("Expected Name 'Dana', got '%s'", config.Name)
	}
}
