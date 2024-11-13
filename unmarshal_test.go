package dotprops

import (
	"testing"
)

// Existing Tests

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

func TestUnmarshalWithCustomFloat(t *testing.T) {
	type FloatConfig struct {
		Rate CustomFloat `property:"rate"`
	}

	data := []byte(`
rate=custom_99.99
`)

	var config FloatConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expected := CustomFloat(99.99)
	if config.Rate != expected {
		t.Errorf("Expected Rate 99.99, got %.2f", config.Rate)
	}
}

func TestUnmarshalWithCustomFloatInvalidPrefix(t *testing.T) {
	type FloatConfig struct {
		Rate CustomFloat `property:"rate"`
	}

	data := []byte(`
rate=invalid_99.99
`)

	var config FloatConfig
	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to invalid prefix in 'rate', but it did not")
	}

	// Since 'rate' couldn't be set due to type mismatch, it should remain at zero value (0)
	if config.Rate != 0 {
		t.Errorf("Expected Rate to be 0 due to failure, got %.2f", config.Rate)
	}
}

// Additional Tests to Increase Coverage

func TestUnmarshalWithUintAndFloatFields(t *testing.T) {
	type NumericConfig struct {
		MaxUsers  uint    `property:"max.users"`
		Threshold float64 `property:"threshold"`
	}

	data := []byte(`
max.users=1000
threshold=75.5
`)

	var config NumericConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.MaxUsers != 1000 {
		t.Errorf("Expected MaxUsers 1000, got %d", config.MaxUsers)
	}
	if config.Threshold != 75.5 {
		t.Errorf("Expected Threshold 75.5, got %f", config.Threshold)
	}
}

func TestUnmarshalWithMultiLevelNestedStruct(t *testing.T) {
	data := []byte(`
service.name=AuthService
service.endpoint.url=https://auth.example.com
service.endpoint.port=443
service.endpoint.active=true
`)

	var config MultiLevelNestedConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Service.Name != "AuthService" {
		t.Errorf("Expected Service.Name 'AuthService', got '%s'", config.Service.Name)
	}
	if config.Service.Endpoint.URL != "https://auth.example.com" {
		t.Errorf("Expected Service.Endpoint.URL 'https://auth.example.com', got '%s'", config.Service.Endpoint.URL)
	}
	if config.Service.Endpoint.Port != 443 {
		t.Errorf("Expected Service.Endpoint.Port 443, got %d", config.Service.Endpoint.Port)
	}
	if config.Service.Endpoint.Active != true {
		t.Errorf("Expected Service.Endpoint.Active true, got %v", config.Service.Endpoint.Active)
	}
}

func TestUnmarshalWithEmbeddedStruct(t *testing.T) {
	type BaseConfig struct {
		Version string `property:"version"`
	}

	type EmbeddedConfig struct {
		BaseConfig
		Name string `property:"name"`
	}

	data := []byte(`
version=1.0.0
name=EmbeddedService
`)

	var config EmbeddedConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", config.Version)
	}
	if config.Name != "EmbeddedService" {
		t.Errorf("Expected Name 'EmbeddedService', got '%s'", config.Name)
	}
}

func TestUnmarshalWithUnsupportedNestedStruct(t *testing.T) {
	type InnerUnsupported struct {
		Data []string `property:"data"`
	}

	type OuterConfig struct {
		Name  string           `property:"name"`
		Inner InnerUnsupported `property:"inner"`
	}

	data := []byte(`
name=Outer
inner.data=one,two
`)

	var config OuterConfig
	err := Unmarshal(data, &config)
	if err == nil {
		t.Fatal("Expected Unmarshal to fail due to unsupported nested struct field type, but it did not")
	}

	// Since 'inner.data' is unsupported, it should not be set
	if config.Inner.Data != nil {
		t.Errorf("Expected Inner.Data to be nil, got %v", config.Inner.Data)
	}
}

func TestUnmarshalMissingKeys(t *testing.T) {
	type CompleteConfig struct {
		Name    string `property:"name"`
		Version string `property:"version"`
	}

	data := []byte(`
name=IncompleteService
`)

	var config CompleteConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "IncompleteService" {
		t.Errorf("Expected Name 'IncompleteService', got '%s'", config.Name)
	}
	if config.Version != "" {
		t.Errorf("Expected Version '', got '%s'", config.Version)
	}
}

func TestUnmarshalWithExtraKeys(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
	}

	data := []byte(`
name=ExtraService
extra.key=extra_value
another.key=another_value
`)

	var config Config
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "ExtraService" {
		t.Errorf("Expected Name 'ExtraService', got '%s'", config.Name)
	}
	// Extra keys should be ignored, no error
}

func TestUnmarshalWithEmptyValues(t *testing.T) {
	type Config struct {
		Name    string `property:"name"`
		Version string `property:"version"`
	}

	data := []byte(`
name=
version=1.2.3
`)

	var config Config
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "" {
		t.Errorf("Expected Name '', got '%s'", config.Name)
	}
	if config.Version != "1.2.3" {
		t.Errorf("Expected Version '1.2.3', got '%s'", config.Version)
	}
}

func TestUnmarshalWithWhitespace(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
		Port int    `property:"port"`
	}

	data := []byte(`
name = WhitespaceService 
port =  8080  
`)

	var config Config
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "WhitespaceService" {
		t.Errorf("Expected Name 'WhitespaceService', got '%s'", config.Name)
	}
	if config.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", config.Port)
	}
}

func TestUnmarshalWithDuplicateKeys(t *testing.T) {
	type Config struct {
		Name string `property:"name"`
	}

	data := []byte(`
name=First
name=Second
`)

	var config Config
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// The last occurrence should take precedence
	if config.Name != "Second" {
		t.Errorf("Expected Name 'Second', got '%s'", config.Name)
	}
}
