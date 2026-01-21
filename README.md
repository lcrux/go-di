# Go Dependency Injection (go-di)

![GitHub stars](https://img.shields.io/github/stars/lcrux/go-di)
![GitHub forks](https://img.shields.io/github/forks/lcrux/go-di)
![GitHub issues](https://img.shields.io/github/issues/lcrux/go-di)
![GitHub license](https://img.shields.io/github/license/lcrux/go-di)

## Overview

The `go-di` project is a lightweight and flexible dependency injection library for Go. It provides a simple way to manage service lifetimes and resolve dependencies in Go applications. The library supports three instance scopes: `Transient`, `Singleton`, and `Scoped`.

### Why Use go-di?

- Simplifies dependency management in Go applications.
- Provides thread-safe service resolution.
- Supports scoped contexts for better resource management.
- Lightweight and easy to integrate.

## Features

- **Service Registration**: Register services with factory functions and specify their lifetime scopes.
- **Dependency Resolution**: Automatically resolve dependencies and manage their lifetimes.
- **Scoped Contexts**: Create and manage scoped instances for specific contexts.
- **Thread-Safe**: Built with concurrency in mind, ensuring thread safety.

## Installation

To use `go-di`, add it to your Go module:

```bash
go get github.com/lcrux/go-di
```

## Getting Started

### Quick Start Example

Here’s a complete example to get started:

```go
package main

import (
    "fmt"
    "github.com/lcrux/go-di/v0/di"
)

type MyService struct {
    Name string
}

func main() {
    // Register the service
    di.Register[MyService](func() *MyService {
        return &MyService{Name: "Hello, go-di!"}
    }, di.Singleton)

    // Resolve the service
    service, err := di.Resolve[MyService]()
    if err != nil {
        panic(err)
    }

    fmt.Println(service.Name)
}
```

### Registering Services

To register a service, use the `Register` function:

```go
import "github.com/lcrux/go-di/v0/di"

type MyService struct {}

di.Register[MyService](func() *MyService {
    return &MyService{}
}, di.Singleton)
```

### Resolving Services

To resolve a registered service, use the `Resolve` function:

```go
service, err := di.Resolve[MyService]()
if err != nil {
    log.Fatalf("Failed to resolve service: %v", err)
}
```

### Using Scoped Contexts

To use scoped instances, create a new `RegistryContext`:

```go
ctx := di.NewRegistryContext()
defer ctx.Close()

scopedService, err := di.ResolveWithContext[MyService](ctx)
if err != nil {
    log.Fatalf("Failed to resolve scoped service: %v", err)
}
```

### Services with Dependencies

You can register and resolve services that depend on other services. Here’s an example:

```go
package main

import (
    "fmt"
    "github.com/lcrux/go-di/v0/di"
)

type Database struct {
    ConnectionString string
}

type UserService struct {
    DB *Database
}

func main() {
    // Register the Database service
    di.Register[Database](func() *Database {
        return &Database{ConnectionString: "postgres://user:password@localhost/db"}
    }, di.Singleton)

    // Register the UserService with a dependency on Database
    di.Register[UserService](func(db *Database) *UserService {
        return &UserService{DB: db}
    }, di.Singleton)

    // Resolve the UserService
    userService, err := di.Resolve[UserService]()
    if err != nil {
        panic(err)
    }

    fmt.Printf("UserService is using database with connection string: %s\n", userService.DB.ConnectionString)
}
```

## Project Structure

- **context.go**: Manages scoped instances within a registry context.
- **registry.go**: Handles service registration and resolution.
- **utils.go**: Contains utility functions for debugging and logging.
- **tests/**: Unit tests for the library.

## Running Tests

To run the tests, use the following command:

```bash
go test ./tests/...
```

## Debugging

The library includes `DebugLog` functions to help with debugging. Ensure that debugging is enabled in your environment to see detailed logs.

```bash
# Enable debugging for go-di

# Linux / macOS
export GODI_DEBUG=true

# Windows PowerShell
$env:GODI_DEBUG="true"

# Windows Command Prompt
set GODI_DEBUG=true
```

## Development Guide

For information on contributing, setting up the development environment, code guidelines, versioning, and more, see the [DEVELOPMENT.md](DEVELOPMENT.md) file.
