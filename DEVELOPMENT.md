# Development Guide

This document provides guidelines for developers contributing to the `go-di` library.

## Code Guidelines

### Naming Conventions

- Use `CamelCase` for exported identifiers.
- Use `lowerCamelCase` for unexported identifiers.
- Avoid abbreviations unless they are well-known (e.g., `URL`, `ID`).

### Package Structure

- Organize code into logical packages:
  - `di/container.go`: Service registration, dependency resolution, and container lifecycle.
  - `di/lifecycle_context.go`: Lifecycle scopes, context management, and cleanup.
  - `di/di-utils/debug_logger.go`: Debug logging utilities.
  - `di/di-utils/semaphore.go`: Concurrency helpers used for shutdown operations.
  - `di/di-utils/types.go`: Generic type utilities for reflection.
  - `demo/`: Example application demonstrating container usage.
- Keep package responsibilities focused and cohesive.

### Commenting Standards

- Use `godoc`-style comments for all exported functions, types, and methods.
- Example:

  ```go
  // Register adds a service to a container with the specified lifecycle scope.
  func Register[T any](c Container, factoryFn interface{}, scope LifecycleScope) error {
      // Implementation...
  }
  ```

### Error Handling

- Avoid using `panic` in library code.
- Prefer returning descriptive errors and log with `utils.DebugLog` where helpful.
- Use `errors.Is` and `errors.As` for error wrapping and unwrapping.
- Return descriptive error messages.

### Thread Safety

- Use `sync.Mutex` or `sync.RWMutex` for managing shared resources.
- Ensure all public methods are thread-safe.

## Versioning Best Practices

This library follows [Semantic Versioning](https://semver.org/):

- **Major Versions**: Increment for breaking changes.
- **Minor Versions**: Increment for new features that are backward-compatible.
- **Patch Versions**: Increment for backward-compatible bug fixes.

### Deprecation Policies

- Mark deprecated features with clear warnings in the documentation.
- Provide alternatives for deprecated features.
- Avoid removing deprecated features until the next major version.

### Release Management

- Use tools like `goreleaser` to automate releases.
- Maintain a `CHANGELOG.md` to document changes in each release.
- Tag each release in Git with the version number (e.g., `v1.0.0`).

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration. The pipeline is configured to:

1. Run on every push and pull request to the `main` branch.
2. Perform the following steps:
   - Checkout the code.
   - Set up Go (version 1.25).
   - Install dependencies using `go mod tidy`.
   - Run lint checks using `golangci-lint`.
   - Run the test suite.

### Viewing Pipeline Results

You can view the results of the CI pipeline in the [Actions tab](https://github.com/lcrux/go-di/actions) of the repository.

## Linting

This project uses `golangci-lint` for linting and static analysis. Follow these steps to set it up:

### Installation

Install `golangci-lint` using the following command:

```bash
brew install golangci-lint # macOS
# OR
sudo apt install golangci-lint # Linux
```

### Running Lint Checks

To run lint checks locally, use:

```bash
golangci-lint run
```

### Configuration

The linting rules are defined in the `.golangci.yml` file. Ensure your code adheres to these rules before submitting changes.

## Workspace and Modules

This repository uses a Go workspace for local development.

- `go.work` includes `./lib` and `./demo`.
- The library module is rooted at `lib/go.mod`.
- The demo app module is rooted at `demo/go.mod`.

## Running Tests

Run module tests from the repository root:

```bash
go test ./lib/...
go test ./demo/...
```
