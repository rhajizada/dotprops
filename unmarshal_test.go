package dotprops

import (
	"testing"
)

func TestUnmarshalSimple(t *testing.T) {
	data := []byte(`
app.name=TestApp
app.port=3000
app.debug=true
`)

	var config SimpleConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName != "TestApp" {
		t.Errorf("Expected AppName 'TestApp', got '%s'", config.AppName)
	}
	if config.Port != 3000 {
		t.Errorf("Expected Port 3000, got %d", config.Port)
	}
	if config.Debug != true {
		t.Errorf("Expected Debug true, got %v", config.Debug)
	}
}

func TestUnmarshalNested(t *testing.T) {
	data := []byte(`
app.name=MyApp
database.host=localhost
database.port=5432
database.username=admin
database.password=secret
`)

	var config NestedConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName != "MyApp" {
		t.Errorf("Expected AppName 'MyApp', got '%s'", config.AppName)
	}
	if config.Database.Host != "localhost" {
		t.Errorf("Expected Database.Host 'localhost', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected Database.Port 5432, got %d", config.Database.Port)
	}
	if config.Database.Username != "admin" {
		t.Errorf("Expected Database.Username 'admin', got '%s'", config.Database.Username)
	}
	if config.Database.Password != "secret" {
		t.Errorf("Expected Database.Password 'secret', got '%s'", config.Database.Password)
	}
}

func TestUnmarshalOptionalFields(t *testing.T) {
	data := []byte(`
app.name=MyApp
app.port=8080
`)

	var config OptionalConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName == nil || *config.AppName != "MyApp" {
		t.Errorf("Expected AppName 'MyApp', got '%v'", config.AppName)
	}

	if config.Port == nil || *config.Port != 8080 {
		t.Errorf("Expected Port 8080, got '%v'", config.Port)
	}

	if config.Debug != nil {
		t.Errorf("Expected Debug to be nil, got '%v'", config.Debug)
	}
}

func TestUnmarshalUnsupportedFieldType(t *testing.T) {
	type UnsupportedConfig struct {
		Data []string `property:"data"`
	}

	var config UnsupportedConfig

	data := []byte("data=one,two,three")

	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to unsupported field type, but it did not")
	}

	// Since 'Data' is an unsupported type, it should remain at zero value (nil)
	if config.Data != nil {
		t.Errorf("Expected Data to be nil, got %v", config.Data)
	}
}

func TestUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`
app.name=MyApp
app.port=8080
app.debug=not_a_boolean
`)

	var config SimpleConfig
	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to invalid boolean value, but it did not")
	}

	// Since 'app.debug' couldn't be set due to type mismatch, it should remain at zero value (false)
	if config.Debug != false {
		t.Errorf("Expected Debug false, got %v", config.Debug)
	}
}

func TestUnmarshalPointerNestedStruct(t *testing.T) {
	data := []byte(`
app.name=PointerApp
database.host=127.0.0.1
database.port=3306
database.username=root
database.password=toor
`)

	var config ConfigWithPointer
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName != "PointerApp" {
		t.Errorf("Expected AppName 'PointerApp', got '%s'", config.AppName)
	}

	if config.Database == nil {
		t.Fatal("Expected Database to be initialized, got nil")
	}

	if config.Database.Host != "127.0.0.1" {
		t.Errorf("Expected Database.Host '127.0.0.1', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 3306 {
		t.Errorf("Expected Database.Port 3306, got %d", config.Database.Port)
	}
	if config.Database.Username != "root" {
		t.Errorf("Expected Database.Username 'root', got '%s'", config.Database.Username)
	}
	if config.Database.Password != "toor" {
		t.Errorf("Expected Database.Password 'toor', got '%s'", config.Database.Password)
	}
}

func TestUnmarshalNestedTypeMismatch(t *testing.T) {
	data := []byte(`
app.name=MyApp
database.host=localhost
database.port=invalid_port
`)

	var config NestedConfig
	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to type mismatch in nested struct, but it did not")
	}

	// Since 'database.port' couldn't be set due to type mismatch, it should remain at zero value (0)
	if config.Database.Port != 0 {
		t.Errorf("Expected Database.Port to be 0, got %d", config.Database.Port)
	}
}

// New Tests for TextUnmarshaler Interface

func TestUnmarshalWithTextUnmarshaler(t *testing.T) {
	type CustomConfig struct {
		Name  CustomString `property:"name"`
		Count CustomInt    `property:"count"`
	}

	data := []byte(`
name=custom_example
count=custom_42
`)

	var config CustomConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "example" {
		t.Errorf("Expected Name 'example', got '%s'", config.Name)
	}
	if config.Count != 42 {
		t.Errorf("Expected Count 42, got '%d'", config.Count)
	}
}

func TestUnmarshalWithTextUnmarshalerInvalidPrefix(t *testing.T) {
	type CustomConfig struct {
		Name  CustomString `property:"name"`
		Count CustomInt    `property:"count"`
	}

	data := []byte(`
name=invalid_example
count=custom_42
`)

	var config CustomConfig
	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to invalid prefix in 'name', but it did not")
	}

	// Even though 'count' is valid, the entire Unmarshal should fail
	if config.Count != 0 {
		t.Errorf("Expected Count to be 0 due to failure, got '%d'", config.Count)
	}
}
