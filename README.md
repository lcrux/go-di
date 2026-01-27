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
- **Keyed Services**: Register and resolve services by custom keys.
- **Factory Injection**: Factories can receive `Container` and/or `LifecycleContext`.
- **Validation**: Validate registrations after setup.

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
    "github.com/lcrux/go-di/di"
)

type MyService struct {
    Name string
}

func main() {
    container := di.NewContainer()
    defer container.Shutdown()

    // Register the service
    di.Register[MyService](container, di.Singleton, func() *MyService {
        return &MyService{Name: "Hello, go-di!"}
    })

    // Resolve the service
    service := di.Resolve[MyService](container, nil)

    fmt.Println(service.Name)
}
```

### Registering Services

To register a service, use the `Register` function with a container instance:

```go
import "github.com/lcrux/go-di/di"

type MyService struct {}

container := di.NewContainer()
defer container.Shutdown()

di.Register[MyService](container, di.Singleton, func() *MyService {
    return &MyService{}
})
```

### Resolving Services

To resolve a registered service, use the `Resolve` function with a container instance:

```go
service := di.Resolve[MyService](container, nil)
```

### Keyed Registrations and Resolution

Register services with explicit keys and resolve them by key:

```go
di.RegisterWithKey[*MyService](container, "my-service.primary", di.Singleton, func() *MyService {
    return &MyService{Name: "Primary"}
})

svc := di.ResolveWithKey[*MyService](container, "my-service.primary", nil)
```

### Resolving Keyed Instances in Custom Factories

If you need a specific key inside a factory, request `Container` and/or `LifecycleContext` and resolve manually:

```go
di.RegisterWithKey[*MyService](container, "my-service.primary", di.Singleton, func() *MyService {
    return &MyService{Name: "Primary"}
})

di.Register[*Consumer](container, di.Transient, func(c di.Container, ctx di.LifecycleContext) *Consumer {
    primary := di.ResolveWithKey[*MyService](c, "my-service.primary", ctx)
    return &Consumer{Service: primary}
})
```

### Wrapper Types to Select Instances by Type

You can create wrapper types to distinguish multiple instances of the same underlying type:

```go
type RepoPrimary struct{ Repo }
type RepoReplica struct{ Repo }

di.Register[RepoPrimary](container, di.Singleton, func() RepoPrimary {
    return RepoPrimary{Repo: NewRepo("primary")}
})
di.Register[RepoReplica](container, di.Singleton, func() RepoReplica {
    return RepoReplica{Repo: NewRepo("replica")}
})

di.Register[*Service](container, di.Transient, func(p RepoPrimary, r RepoReplica) *Service {
    return NewService(p.Repo, r.Repo)
})
```

### Lifecycle Scopes

`go-di` supports three lifecycle scopes:

- **Transient**: A new instance is created every time the service is resolved.
- **Singleton**: A single instance is shared across the container’s lifetime.
- **Scoped**: A single instance is shared within a specific lifecycle context.

### Container Lifecycle

- `NewContainer()` creates a new container with its own background lifecycle context.
- `Resolve(..., nil)` uses the container’s background context automatically.
- `CloseContext(ctx)` triggers lifecycle cleanup for scoped instances and returns any errors.
- `Shutdown()` closes all contexts and returns any errors from lifecycle cleanup.

### Using Scoped Contexts

To use scoped instances, create a new lifecycle context from the container:

```go
ctx := container.NewContext()
defer container.CloseContext(ctx)

scopedService := di.Resolve[MyService](container, ctx)
```

### Services with Dependencies

You can register and resolve services that depend on other services. Here’s an example:

```go
package main

import (
    "fmt"
    "github.com/lcrux/go-di/di"
)

type Database struct {
    ConnectionString string
}

type UserService struct {
    DB *Database
}

func main() {
    container := di.NewContainer()
    defer container.Shutdown()

    // Register the Database service
    di.Register[Database](container, di.Singleton, func() *Database {
        return &Database{ConnectionString: "postgres://user:password@localhost/db"}
    })

    // Register the UserService with a dependency on Database
    di.Register[UserService](container, di.Singleton, func(db *Database) *UserService {
        return &UserService{DB: db}
    })

    // Resolve the UserService
    userService := di.Resolve[UserService](container, nil)

    fmt.Printf("UserService is using database with connection string: %s\n", userService.DB.ConnectionString)
}
```

### Lifecycle Cleanup

Any resolved instance that implements `LifecycleListener` will have its `EndLifecycle()` method
invoked when a lifecycle context is closed or the container is shut down.

```go
type Worker struct{}

func (w *Worker) EndLifecycle() error {
    // release resources here
    return nil
}

container := di.NewContainer()
defer container.Shutdown()

di.Register[*Worker](container, di.Scoped, func() *Worker {
    return &Worker{}
})

ctx := container.NewContext()
_ = di.Resolve[*Worker](container, ctx)
_ = container.CloseContext(ctx) // triggers EndLifecycle on scoped instances
```

### Validation

You can validate all registrations after setup to detect missing dependencies early:

```go
if err := container.Validate(); err != nil {
    panic(err)
}
```

## Project Structure

- **di/container.go**: Service registration, dependency resolution, and container lifecycle.
- **di/lifecycle_context.go**: Lifecycle scopes, contexts, and shutdown behavior.
- **di/di-utils/debug_logger.go**: Debug logging utilities and environment flag handling.
- **di/di-utils/semaphore.go**: Concurrency helpers used during lifecycle shutdown.
- **di/di-utils/types.go**: Generic type utilities for reflection.

## Running Tests

To run the tests, use the following commands:

```bash
go test ./di/...
go test ./demo/...
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
