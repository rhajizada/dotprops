# dotprops

![ci](https://github.com/rhajizada/dotprops/actions/workflows/ci.yml/badge.svg)

dotprops is a `Go` package for marshalling and unmarshalling `Java` `.properties`
files into structs, similar to how the `encoding/json` package works.

## Features

- **Marshal Go structs** into `.properties` file format.
- **Unmarshal `.properties` files** into Go structs.
- Supports **nested structures**.
- Handles **optional fields** via pointers.

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

`dotprops` supports nested structs to represent hierarchical properties using dot notation.

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
