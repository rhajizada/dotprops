# dotprops

![ci](https://github.com/rhajizada/dotprops/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.22-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)


dotprops is a `Go` package for marshalling and unmarshalling `Java` `.properties`
files into structs, similar to how the `encoding/json` package works.

## Features

- **Marshal Go structs** into `.properties` file format.
- **Unmarshal `.properties` files** into Go structs.
- Supports **nested structures**.
- Handles **optional fields** via pointers.
- Custom property marshaling and unmarshaling via `PropMarshaler` and `
PropUnmarshaler` interfaces.
- Custom text marshaling and unmarshaling via `TextMarshaler` and
  `TextUnmarshaler` interfaces.

## Installation

```bash
go get -u github.com/rhajizada/dotprops
```

## Usage

### Importing the package

```go
import "github.com/rhajizada/dotprops"
```

### Marshalling

Convert a struct into a `.properties` formatted byte slice.

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

type Config struct {
    AppName string `property:"app.name"`
    Port    int    `property:"app.port"`
    Debug   bool   `property:"app.debug"`
}

func main() {
    config := Config{
        AppName: "MyApp",
        Port:    8080,
        Debug:   true,
    }

    data, err := dotprops.Marshal(config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data))
}
```

### Unmarshalling

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

type Config struct {
    AppName string `property:"app.name"`
    Port    int    `property:"app.port"`
    Debug   bool   `property:"app.debug"`
}

func main() {
    data := []byte(`
app.name=MyApp
app.port=8080
app.debug=true
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", config)
}
```

### Optional fields using pointers

Fields that are optional can be represented as pointers in your struct. If the
property is missing in the file, the pointer will be `nil`.

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

type Config struct {
    AppName *string `property:"app.name"`
    Port    int     `property:"app.port"`
    Debug   *bool   `property:"app.debug"`
}

func main() {
    data := []byte(`
app.name=MyApp
app.port=8080
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    if config.AppName != nil {
        fmt.Println("AppName:", *config.AppName)
    }
    fmt.Println("Port:", config.Port)
    if config.Debug != nil {
        fmt.Println("Debug:", *config.Debug)
    } else {
        fmt.Println("Debug is not set")
    }
}
```

### Nested structures

`dotprops` supports nested structs to represent hierarchical properties using
dot notation.

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

type DatabaseConfig struct {
    Host     string `property:"host"`
    Port     int    `property:"port"`
    Username string `property:"username"`
    Password string `property:"password"`
}

type Config struct {
    AppName  string         `property:"app.name"`
    Database DatabaseConfig `property:"database"`
}

func main() {
    data := []byte(`
app.name=MyApp
database.host=localhost
database.port=5432
database.username=admin
database.password=secret
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("AppName: %s\n", config.AppName)
    fmt.Printf("Database Host: %s\n", config.Database.Host)
    fmt.Printf("Database Port: %d\n", config.Database.Port)
}
```

### Nested structures with optional fields

You can also have optional nested structs by using pointers.

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

type DatabaseConfig struct {
    Host     *string `property:"host"`
    Port     *int    `property:"port"`
    Username string  `property:"username"`
    Password string  `property:"password"`
}

type Config struct {
    AppName  string          `property:"app.name"`
    Database *DatabaseConfig `property:"database"`
}

func main() {
    data := []byte(`
app.name=MyApp
database.username=admin
database.password=secret
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("AppName: %s\n", config.AppName)
    if config.Database != nil {
        if config.Database.Host != nil {
            fmt.Printf("Database Host: %s\n", *config.Database.Host)
        } else {
            fmt.Println("Database Host is not set")
        }
        if config.Database.Port != nil {
            fmt.Printf("Database Port: %d\n", *config.Database.Port)
        } else {
            fmt.Println("Database Port is not set")
        }
        fmt.Printf("Database Username: %s\n", config.Database.Username)
    } else {
        fmt.Println("Database config is not set")
    }
}
```

### Custom Marshaling and Unmarshaling Interfaces

`dotprops` provides two sets of interfaces to allow for custom serialization and
deserialization behaviors:

- `TextMarshaler` and `TextUnmarshaler`
- `PropMarshaler` and `PropUnmarshaler`

#### TextMarshaler and TextUnmarshaler

These interfaces are similar to those provided by Go's encoding packages and
allow you to define custom text-based serialization for individual fields.

**Marshalling Example:**

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

// CustomString implements TextMarshaler and TextUnmarshaler
type CustomString string

func (cs CustomString) MarshalText() ([]byte, error) {
    return []byte(fmt.Sprintf("custom_%s", cs)), nil
}

func (cs *CustomString) UnmarshalText(text []byte) error {
    if len(text) < 8 || string(text[:7]) != "custom_" {
        return fmt.Errorf("invalid prefix in value: %s", text)
    }
    *cs = CustomString(string(text[7:]))
    return nil
}

type Config struct {
    Name CustomString `property:"name"`
}

func main() {
    config := Config{
        Name: "example",
    }

    data, err := dotprops.Marshal(&config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data))
}
```

**Unmarshalling Example:**

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

// CustomString implements TextMarshaler and TextUnmarshaler
type CustomString string

func (cs CustomString) MarshalText() ([]byte, error) {
    return []byte(fmt.Sprintf("custom_%s", cs)), nil
}

func (cs *CustomString) UnmarshalText(text []byte) error {
    if len(text) < 8 || string(text[:7]) != "custom_" {
        return fmt.Errorf("invalid prefix in value: %s", text)
    }
    *cs = CustomString(string(text[7:]))
    return nil
}

type Config struct {
    Name CustomString `property:"name"`
}

func main() {
    data := []byte(`
name=custom_example
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Name: %s\n", config.Name)
}
```

#### PropMarshaler and PropUnmarshaler

These custom interfaces allow for more granular control over how individual
properties are marshaled and unmarshaled, especially useful for complex or
specialized data formats.

**Defining Custom Interfaces:**

First, define the `PropMarshaler` and `PropUnmarshaler` interfaces within your package:

```go
package dotprops

// PropMarshaler allows custom marshaling of a single property.
type PropMarshaler interface {
    MarshalProp() (key string, value string, err error)
}

// PropUnmarshaler allows custom unmarshaling of a single property.
type PropUnmarshaler interface {
    UnmarshalProp(key string, value string) error
}
```

**Marshalling with PropMarshaler:**

Implement the PropMarshaler interface in your custom type to control how a
specific field is marshaled.

```go
package main

import (
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

// CustomPropMarshaller implements PropMarshaler
type CustomPropMarshaller struct {
    Field1 string
    Field2 int
}

func (c CustomPropMarshaller) MarshalProp() (string, string, error) {
    key := "custom.field"
    value := fmt.Sprintf("%s_%d", c.Field1, c.Field2)
    return key, value, nil
}

type Config struct {
    CustomField CustomPropMarshaller `property:"custom.field"`
    Name        string               `property:"name"`
}

func main() {
    config := Config{
        CustomField: CustomPropMarshaller{
            Field1: "value1",
            Field2: 42,
        },
        Name: "TestService",
    }

    data, err := dotprops.Marshal(&config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data))
}

```

**Handling Errors in PropMarshaler:**

If your PropMarshaler implementation encounters an error during marshaling, it
should return it to ensure that the Marshal function can handle it appropriately.

```go
package main

import (
    "errors"
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

// FaultyPropMarshaller implements PropMarshaler and always returns an error
type FaultyPropMarshaller struct{}

func (f FaultyPropMarshaller) MarshalProp() (string, string, error) {
    return "", "", errors.New("intentional marshaling error")
}

type Config struct {
    FaultyField FaultyPropMarshaller `property:"faulty.field"`
    Name        string               `property:"name"`
}

func main() {
    config := Config{
        FaultyField: FaultyPropMarshaller{},
        Name:        "FaultyService",
    }

    _, err := dotprops.Marshal(&config)
    if err != nil {
        log.Fatalf("Marshal failed: %v", err)
    }
}
```

**Unmarshalling with PropUnmarshaler:**

Implement the PropUnmarshaler interface in your custom type to control how
a specific field is unmarshaled.

```go
package main

import (
    "fmt"
    "log"
    "strconv"
    "strings"

    "github.com/rhajizada/dotprops"
)

// CustomPropUnmarshaller implements PropUnmarshaler
type CustomPropUnmarshaller struct {
    Field1 string
    Field2 int
}

func (c *CustomPropUnmarshaller) UnmarshalProp(key string, value string) error {
    if key != "custom.field" {
        return fmt.Errorf("unexpected key: %s", key)
    }
    parts := strings.Split(value, "_")
    if len(parts) != 2 {
        return fmt.Errorf("invalid value format: %s", value)
    }
    c.Field1 = parts[0]
    var err error
    c.Field2, err = strconv.Atoi(parts[1])
    if err != nil {
        return err
    }
    return nil
}

type Config struct {
    CustomField CustomPropUnmarshaller `property:"custom.field"`
    Name        string                  `property:"name"`
}

func main() {
    data := []byte(`
custom.field=value1_42
name=TestService
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("CustomField: %+v\n", config.CustomField)
    fmt.Printf("Name: %s\n", config.Name)
}
```

**Handling Errors in PropUnmarshaler:**

If your PropUnmarshaler implementation encounters an error during unmarshaling,
it should return it to ensure that the Unmarshal function can handle it
appropriately.

```go
package main

import (
    "errors"
    "fmt"
    "log"

    "github.com/rhajizada/dotprops"
)

// FaultyPropUnmarshaller implements PropUnmarshaler and always returns an error
type FaultyPropUnmarshaller struct{}

func (f *FaultyPropUnmarshaller) UnmarshalProp(key string, value string) error {
    return errors.New("intentional unmarshaling error")
}

type Config struct {
    FaultyField FaultyPropUnmarshaller `property:"faulty.field"`
    Name        string                  `property:"name"`
}

func main() {
    data := []byte(`
faulty.field=invalid_value
name=FaultyService
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatalf("Unmarshal failed: %v", err)
    }

    fmt.Printf("FaultyField: %+v\n", config.FaultyField)
    fmt.Printf("Name: %s\n", config.Name)
}
```

**Combining TextMarshaler/Unmarshaler with PropMarshaler/Unmarshaler:**

You can seamlessly integrate custom marshaling and unmarshaling interfaces
within the same struct, allowing for flexible and granular control over how each
field is processed.

```go
package main

import (
    "fmt"
    "log"
    "strconv"
    "strings"

    "github.com/rhajizada/dotprops"
)

// CustomString implements TextMarshaler and TextUnmarshaler
type CustomString string

func (cs CustomString) MarshalText() ([]byte, error) {
    return []byte(fmt.Sprintf("custom_%s", cs)), nil
}

func (cs *CustomString) UnmarshalText(text []byte) error {
    if len(text) < 8 || string(text[:7]) != "custom_" {
        return fmt.Errorf("invalid prefix in value: %s", text)
    }
    *cs = CustomString(string(text[7:]))
    return nil
}

// CustomPropMarshaller implements PropMarshaler
type CustomPropMarshaller struct {
    Field1 string
    Field2 int
}

func (c CustomPropMarshaller) MarshalProp() (string, string, error) {
    key := "custom.field"
    value := fmt.Sprintf("%s_%d", c.Field1, c.Field2)
    return key, value, nil
}

// CustomPropUnmarshaller implements PropUnmarshaler
type CustomPropUnmarshaller struct {
    Field1 string
    Field2 int
}

func (c *CustomPropUnmarshaller) UnmarshalProp(key string, value string) error {
    if key != "custom.field" {
        return fmt.Errorf("unexpected key: %s", key)
    }
    parts := strings.Split(value, "_")
    if len(parts) != 2 {
        return fmt.Errorf("invalid value format: %s", value)
    }
    c.Field1 = parts[0]
    var err error
    c.Field2, err = strconv.Atoi(parts[1])
    if err != nil {
        return err
    }
    return nil
}

type Config struct {
    Name        CustomString           `property:"name"`
    CustomField CustomPropUnmarshaller `property:"custom.field"`
}

func main() {
    data := []byte(`
name=custom_example
custom.field=value1_42
`)

    var config Config
    err := dotprops.Unmarshal(data, &config)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Name: %s\n", config.Name)
    fmt.Printf("CustomField: %+v\n", config.CustomField)
}
```

