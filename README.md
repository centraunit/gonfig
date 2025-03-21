# GoNfig - Flexible Configuration Management for Go

A high-performance configuration management library for Go with support for environment variables, schema validation, and multiple configuration loaders.

[![Go Tests](https://github.com/centraunit/gonfig/actions/workflows/tests.yml/badge.svg)](https://github.com/centraunit/gonfig/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/centraunit/gonfig)](https://goreportcard.com/report/github.com/centraunit/gonfig)
[![GoDoc](https://godoc.org/github.com/centraunit/gonfig?status.svg)](https://godoc.org/github.com/centraunit/gonfig)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Table of Contents
- [Performance Benchmarks](#performance-benchmarks)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration Schema](#configuration-schema)
- [Environment Variables](#environment-variables)
- [Type-Safe Configuration Access](#type-safe-configuration-access)
- [Struct Unmarshaling](#struct-unmarshaling)
- [Dynamic Configuration Updates](#dynamic-configuration-updates)
- [Custom Configuration Loaders](#custom-configuration-loaders)
- [Thread Safety](#thread-safety)
- [Contributing](#contributing)
- [License](#license)

## Performance Benchmarks

| Operation | Time (ns/op) | Memory (B/op) | Allocs/op |
|-----------|-------------|---------------|-----------|
| Get Simple | 84.01 | 32 | 1 |
| Get Deep | 176.9 | 80 | 1 |
| GetString Simple | 90.49 | 32 | 1 |
| GetInt Direct | 99.86 | 32 | 1 |
| GetBool Direct | 93.04 | 32 | 1 |
| GetFloat Direct | 121.5 | 32 | 1 |
| GetStringArray Direct | 106.5 | 32 | 1 |
| Set Simple | 148.8 | 39 | 1 |
| Set Deep | 308.4 | 87 | 1 |
| Refresh | 2434 | 2832 | 20 |

## Features

- Dot notation path access (e.g., "database.host")
- Type-safe configuration access
- Environment variable support with defaults
- Schema validation
- Multiple configuration loaders
- High-performance path caching
- Thread-safe operations
- Automatic type conversion
- Default value support

## Installation

```bash
go get github.com/centraunit/gonfig
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/centraunit/gonfig"
    "github.com/centraunit/gonfig/contracts"
)

func main() {
    // Initialize the config registry with environment
    // Supported environments: "development", "staging", "production", "testing"
    config, err := gonfig.GetConfigRegistry("development")
    if err != nil {
        log.Fatal(err)
    }

    // Register a configuration loader
    config.Register("app", func(registry contracts.ConfigRegistry) map[string]interface{} {
        return map[string]interface{}{
            "database": map[string]interface{}{
                "host": "localhost",
                "port": 5432,
                "credentials": map[string]interface{}{
                    "username": "admin",
                    "password": "secret",
                },
            },
        }
    })

    // Access configuration values
    host, err := config.GetString("app.database.host")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Database Host: %s\n", host)

    port, err := config.GetInt("app.database.port")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Database Port: %d\n", port)

    // Use environment variables with defaults
    apiKey := config.GetEnvString("API_KEY", "default-key")
    fmt.Printf("API Key: %s\n", apiKey)
}
```

## Environment Files

GoNfig automatically loads environment files based on the environment:
- `.env` file for "development", "staging", or "production" environments
- `.env.testing` file for "testing" environment

## Configuration Schema

```go
package main

import (
    "fmt"
    "log"
    "reflect"
    
    "github.com/centraunit/gonfig"
    "github.com/centraunit/gonfig/contracts"
)

func main() {
    // Initialize config registry
    config, err := gonfig.GetConfigRegistry("development")
    if err != nil {
        log.Fatal(err)
    }
    
    schema := gonfig.NewConfigSchema()

    // Define schema fields
    schema.AddField("app.database.host", contracts.ConfigSchemaField{
        Type:     reflect.String,
        Required: true,
        Validator: func(value interface{}) error {
            host, ok := value.(string)
            if !ok || host == "" {
                return fmt.Errorf("invalid database host")
            }
            return nil
        },
    })

    schema.AddField("app.database.port", contracts.ConfigSchemaField{
        Type:     reflect.Int,
        Required: true,
        Default:  5432,
    })

    // Validate configuration
    err = schema.Validate(config.GetAll())
    if err != nil {
        log.Fatal(err)
    }
}
```

## Environment Variables

Access environment variables with type safety:

```go
// String with default
dbHost := config.GetEnvString("DB_HOST", "localhost")

// Integer with default
dbPort := config.GetEnvInt("DB_PORT", 5432)

// Boolean with default
debug := config.GetEnvBool("DEBUG_MODE", false)

// String array with default
allowedHosts := config.GetEnvStringArray("ALLOWED_HOSTS", []string{"localhost"})
```

## Type-Safe Configuration Access

```go
// String access
host, err := config.GetString("app.database.host")
hostWithDefault, err := config.GetString("app.database.host", "localhost")

// Integer access
port, err := config.GetInt("app.database.port")
portWithDefault, err := config.GetInt("app.database.port", 5432)

// Boolean access
enabled, err := config.GetBool("app.feature.enabled")
enabledWithDefault, err := config.GetBool("app.feature.enabled", false)

// Float access
timeout, err := config.GetFloat("app.api.timeout")
timeoutWithDefault, err := config.GetFloat("app.api.timeout", 30.0)

// String array access
hosts, err := config.GetStringArray("app.allowed.hosts")
hostsWithDefault, err := config.GetStringArray("app.allowed.hosts", []string{"localhost"})
```

## Struct Unmarshaling

Unmarshal configuration sections into structs:

```go
type DatabaseConfig struct {
    Host     string `config:"host"`
    Port     int    `config:"port"`
    Username string `config:"credentials.username"`
    Password string `config:"credentials.password"`
}

var dbConfig DatabaseConfig
err := config.Unmarshal("app.database", &dbConfig)
if err != nil {
    log.Fatal(err)
}
```

## Dynamic Configuration Updates

```go
// Set a configuration value
config.Set("app.api.timeout", 60.0)

// Refresh configuration from all loaders
config.Refresh()
```

## Custom Configuration Loaders

```go
config.Register("custom", func(registry contracts.ConfigRegistry) map[string]interface{} {
    return map[string]interface{}{
        "settings": map[string]interface{}{
            "value": "test",
            "nested": map[string]interface{}{
                "key": "value",
            },
        },
    }
})

// Access custom config
value, err := config.GetString("custom.settings.value")
```

## Thread Safety

All operations in GoNfig are thread-safe and can be used in concurrent environments:

```go
go func() {
    value, _ := config.GetString("key")
    // Use value
}()

go func() {
    config.Set("key", "new value")
}()
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
