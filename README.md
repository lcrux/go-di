# go-di

`go-di` is a lightweight and flexible dependency injection library for Go. It simplifies the management of dependencies in your Go applications, making your code more modular, testable, and maintainable.

## Features

- **Simple API**: Easy to use and integrate into existing projects.
- **Flexible Dependency Management**: Supports constructor injection and lazy initialization.
- **Modular Design**: Encourages clean and modular code architecture.
- **Testable**: Simplifies mocking and testing by managing dependencies effectively.

## Installation

To install the library, use:

```bash
go get github.com/username/go-di
```

## Getting Started

Here’s a quick example to demonstrate how to use `go-di` in your project:

### 1. Define Your Services

```go
package services

type UserService struct {
    // Dependencies can be injected here
}

func (u *UserService) GetUser(id int) string {
    return "User Name"
}
```

### 2. Register Dependencies

```go
package registry

import (
    "github.com/username/go-di"
    "your_project/services"
)

func RegisterDependencies(container *di.Container) {
    container.Register(func() *services.UserService {
        return &services.UserService{}
    })
}
```

### 3. Resolve and Use Dependencies

```go
package main

import (
    "fmt"
    "github.com/username/go-di"
    "your_project/registry"
    "your_project/services"
)

func main() {
    container := di.NewContainer()
    registry.RegisterDependencies(container)

    userService := container.Resolve[*services.UserService]()
    fmt.Println(userService.GetUser(1))
}
```

## Debugging

Enable debug logs to troubleshoot the library's behavior by setting the `GODI_DEBUG` environment variable to `true`:

```bash
export GODI_DEBUG=true
```

This will enable detailed logs for operations like service registration, resolution, and context management.

---

## Scoped Contexts

`go-di` supports scoped contexts for managing service instances within specific lifetimes. This is useful for scenarios like request-based lifetimes in web applications.

### Example: Using Scoped Contexts

```go
package main

import (
    "fmt"
    "github.com/username/go-di/v1/registry"
)

func main() {
    container := registry.NewRegistryContext()

    // Register a service
    container.SetInstance(reflect.TypeOf("example"), reflect.ValueOf("Scoped Instance"))

    // Resolve the service
    instance, _ := container.GetInstance(reflect.TypeOf("example"))
    fmt.Println(instance.String())

    // Clean up scoped instances
    container.Close()
}
```

## Examples

### Registering Services

Register services with different lifetimes (e.g., `Transient`, `Scoped`, `Singleton`):

```go
import "github.com/username/go-di/v1/registry"

func init() {
    registry.Register[MyService](NewMyService, registry.Transient)
    registry.Register[MyRepository](NewMyRepository, registry.Singleton)
}
```

### Resolving Services

Resolve services from the registry and use them:

```go
import "github.com/username/go-di/v1/registry"

func main() {
    myService, err := registry.Resolve[MyService]()
    if err != nil {
        log.Fatalf("Failed to resolve MyService: %v", err)
    }

    myService.DoSomething()
}
```

### Using Scoped Contexts

Create a scoped context for managing service lifetimes:

```go
import "github.com/username/go-di/v1/registry"

func main() {
    regCtx := registry.NewRegistryContext()
    defer regCtx.Close()

    myService, err := registry.ResolveWithContext[MyService](regCtx)
    if err != nil {
        log.Fatalf("Failed to resolve MyService with context: %v", err)
    }

    myService.DoSomething()
}
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests to improve the library.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Start building modular and testable Go applications with `go-di` today!