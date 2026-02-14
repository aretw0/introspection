# AGENTS.md

## Project Overview

**introspection** is a domain-agnostic Go package for component state introspection, monitoring, and visualization.

Originally extracted from the [lifecycle](https://github.com/aretw0/lifecycle) project, this package provides a generic observation layer for any Go application that needs to expose and visualize its internal state.

For a detailed breakdown of the vision and use cases, see **[docs/PRODUCT.md](docs/PRODUCT.md)**.

## Project Structure & Documentation

* **[docs/TECHNICAL.md](docs/TECHNICAL.md)**: Architecture and technical design.
* **[docs/PLANNING.md](docs/PLANNING.md)**: Roadmap and future enhancements.
* **[docs/PRODUCT.md](docs/PRODUCT.md)**: Vision, use cases, and problem space.
* **[docs/DECISIONS.md](docs/DECISIONS.md)**: Design decisions and rationale.
* **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)**: Configuration philosophy.
* **[docs/RECIPES.md](docs/RECIPES.md)**: Common usage patterns and examples.
* **[examples/](examples/)**: Runnable examples (`basic`, `generic`).

## Key Commands

Ensure dependencies are synced:

```bash
go mod tidy
# or
make tidy
```

### Running Tests

```bash
go test -race -timeout 60s -v ./...
# or
make test
```

### Running Coverage

```bash
go test -race -timeout 60s -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
# or
make coverage
```

### Running Examples

```bash
# Generic example (domain-agnostic)
cd examples/generic
go run main.go

# Basic example (legacy worker/signal domain)
cd examples/basic
go run main.go
```

## Development Philosophy

* **Domain Agnostic**: No hardcoded terminology. Users define their own domain concepts.
* **Composability Over Context**: Focus on providing generic observation primitives that compose well.
* **Type Safety**: Leverage Go generics for type-safe state watching.
* **Zero Dependencies**: Keep the package lightweight and dependency-free.
* **Observability First**: Every component should be introspectable and visualizable.

## Design Principles

1. **Generic by Default**: All new features should be domain-agnostic.
2. **Backward Compatible**: Legacy functions remain available but deprecated.
3. **Configuration Over Convention**: Use functional options for customization.
4. **Testing**: Comprehensive tests with race detection enabled.
