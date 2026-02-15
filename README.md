# Introspection

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/introspection)](https://goreportcard.com/report/github.com/aretw0/introspection)
[![Go Reference](https://pkg.go.dev/badge/github.com/aretw0/introspection.svg)](https://pkg.go.dev/github.com/aretw0/introspection)
[![License](https://img.shields.io/github/license/aretw0/introspection.svg?color=red)](./LICENSE)
[![Release](https://img.shields.io/github/release/aretw0/introspection.svg?branch=main)](https://github.com/aretw0/introspection/releases)

A **domain-agnostic** Go package for component state introspection, monitoring, and visualization.

## Overview

**Introspection** solves the generic problem of "how any Go component publishes its internal state for visualization and monitoring". Originally extracted from the [lifecycle](https://github.com/aretw0/lifecycle) project, this package provides:

- **Domain-Agnostic Design**: No hardcoded terminology like "worker" or "signal" - works with any domain
- **Introspectable Interface**: Components can expose their internal state
- **TypedWatcher[S]**: Type-safe state watching with generics
- **State Aggregation**: Combine state changes from multiple components
- **Customizable Mermaid Diagrams**: Automatic visualization with full control over labels, styles, and structure

This package is useful for any project that wants to auto-document its topology at runtime and enable live monitoring and debugging **without being tied to specific domain terminology**.

## Key Principle: Composability Over Context

The package emphasizes **composability** - you define your domain, we provide the observation layer. No assumptions about "workers", "signals", or any other specific domain concepts.

## Features

### 1. Introspectable Components

Components implement the `Introspectable` interface to expose their state:

```go
type Introspectable interface {
    State() any
}
```

### 2. Type-Safe State Watching

The `TypedWatcher[S]` interface provides type-safe state change notifications:

```go
type TypedWatcher[S any] interface {
    State() S
    Watch(ctx context.Context) <-chan StateChange[S]
}
```

### 3. State Aggregation

Aggregate state changes from multiple components into a unified stream:

```go
snapshots := introspection.AggregateWatchers(ctx, component1, component2, component3)
for snapshot := range snapshots {
    fmt.Printf("Component %s changed state\n", snapshot.ComponentID)
}
```

### 4. Generic Mermaid Diagram Generation (Domain-Agnostic)

Generate Mermaid diagrams with **full customization** - no hardcoded labels or terminology:

```go
// Generic TreeDiagram - works with any hierarchical structure
config := &introspection.DiagramConfig{
    SecondaryID: "root",
}
diagram := introspection.TreeDiagram(hierarchyState, config)

// ComponentDiagram - fully customizable labels
config := &introspection.DiagramConfig{
    PrimaryID:        "controller",
    PrimaryLabel:     "Control Layer",
    PrimaryNodeLabel: "🎮 Controller",
    SecondaryID:      "workers",
    SecondaryLabel:   "Worker Pool",
    ConnectionLabel:  "manages",
}
diagram := introspection.ComponentDiagram(controllerState, workerState, config)

// StateMachineDiagram - custom state names and transitions
smConfig := &introspection.StateMachineConfig{
    InitialState:      "Active",
    GracefulState:     "Draining",
    ForcedState:       "Terminated",
    InitialToGraceful: "SHUTDOWN",
    GracefulToForced:  "KILL",
}
stateMachine := introspection.StateMachineDiagram(state, smConfig)
```

**Backward Compatible**: Legacy functions (`WorkerTreeDiagram`, `SignalStateMachine`, `SystemDiagram`) remain available but are deprecated in favor of the generic versions.

## Installation

```bash
go get github.com/aretw0/introspection
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/aretw0/introspection"
)

// Define your component's state
type MyComponentState struct {
    Name   string
    Status string
}

// Implement TypedWatcher
type MyComponent struct {
    state MyComponentState
    changes chan introspection.StateChange[MyComponentState]
}

func (c *MyComponent) ComponentType() string {
    return "processor"  // Use your own domain terminology
}

func (c *MyComponent) State() MyComponentState {
    return c.state
}

func (c *MyComponent) Watch(ctx context.Context) <-chan introspection.StateChange[MyComponentState] {
    // Return a channel that sends state changes
    // (implementation details omitted for brevity)
}

func main() {
    component := &MyComponent{
        state: MyComponentState{Name: "processor-1", Status: "Running"},
    }
    
    // Watch state changes
    ctx := context.Background()
    changes := component.Watch(ctx)
    
    for change := range changes {
        fmt.Printf("State changed: %v -> %v\n", 
            change.OldState, change.NewState)
    }
}
```

## Examples

### Generic Example (Custom Task Scheduler Domain)

See the [examples/generic](examples/generic) directory for a complete working example demonstrating **domain-agnostic** usage with a Task Scheduler domain (no worker/signal terminology):
- Custom domain types (Scheduler, Tasks)
- Generic `TreeDiagram`, `ComponentDiagram`, `StateMachineDiagram`
- Custom node styling and labeling
- Fully customized labels and terminology

Run the example:
```bash
cd examples/generic
go run main.go
```

### Basic Example (Legacy Worker/Signal Domain)

See the [examples/basic](examples/basic) directory for a complete working example using the original worker/signal domain that demonstrates:
- Implementing the `Introspectable` interface
- Using `TypedWatcher` for type-safe state watching
- Aggregating state changes from multiple components
- Generating Mermaid diagrams (legacy functions)

Run the example:
```bash
cd examples/basic
go run main.go
```

## Use Cases

- **Observability**: Monitor the state of distributed system components **in any domain**
- **Debugging**: Track state transitions in real-time
- **Documentation**: Auto-generate system topology diagrams **with your terminology**
- **Testing**: Verify component behavior through state inspection
- **Monitoring**: Build dashboards that visualize component states
- **Domain Modeling**: Express your system's architecture without framework constraints

## Key Interfaces

### Introspectable
```go
type Introspectable interface {
    State() any
}
```

### Component
```go
type Component interface {
    ComponentType() string
}
```

### TypedWatcher[S]
```go
type TypedWatcher[S any] interface {
    State() S
    Watch(ctx context.Context) <-chan StateChange[S]
}
```

### EventSource
```go
type EventSource interface {
    Events(ctx context.Context) <-chan ComponentEvent
}
```

## Core Types

### StateChange[S]
```go
type StateChange[S any] struct {
    ComponentID   string
    ComponentType string
    OldState      S
    NewState      S
    Timestamp     time.Time
}
```

### StateSnapshot
```go
type StateSnapshot struct {
    ComponentID   string
    ComponentType string
    Timestamp     time.Time
    Payload       any
}
```

## Visualization

The package includes powerful Mermaid diagram generation capabilities:

### Domain-Agnostic Configuration

Control every aspect of diagram generation through configuration:

```go
// Custom node styling based on your domain
config := &introspection.DiagramConfig{
    NodeStyler: func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
        switch metadata["type"] {
        case "database":
            return "🗄️", "[(", ")]", "container"
        case "api":
            return "🌐", "[/", "/]", "process"
        default:
            return "📋", "[", "]", "process"
        }
    },
    NodeLabeler: func(name, status string, pid int, metadata map[string]string, icon string) string {
        return fmt.Sprintf("%s %s [%s]", icon, name, status)
    },
}
```

### Default Styles
The package comes with pre-defined Mermaid styles for common component states:
- Running (blue)
- Stopped (gray)
- Failed (red)
- Pending (purple)
- And more...

### Custom Styles
You can customize the appearance with your own Mermaid class definitions:

```go
diagram := introspection.TreeDiagram(
    state,
    config,
    introspection.WithStyles("classDef custom fill:#fff;"),
)
```

## Design Philosophy

### Composability Over Context

The package is designed around the principle that **generic observation mechanisms should not dictate domain terminology**. Instead of hardcoding terms like "worker", "signal", or "supervisor":

- **You define your domain** (tasks, services, processors, etc.)
- **You configure the visualization** (labels, icons, connections)
- **You compose the observation layer** (watchers, aggregators, diagrams)

This approach enables:
- ✅ **True reusability** across different domains
- ✅ **No conceptual coupling** to specific architectures
- ✅ **Full control** over terminology and presentation
- ✅ **Backward compatibility** with legacy code

## Testing

Run tests:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -v -cover
```

## Documentation

For more detailed information:
- [Release Process](docs/RELEASING.md) - How to create releases
- [Technical Design](docs/TECHNICAL.md) - Architecture and design
- [Product Vision](docs/PRODUCT.md) - Vision and use cases
- [Configuration](docs/CONFIGURATION.md) - Configuration philosophy
- [Recipes](docs/RECIPES.md) - Common usage patterns
- [Decisions](docs/DECISIONS.md) - Design decisions

## Origin

This package was extracted from the [lifecycle](https://github.com/aretw0/lifecycle) project and made domain-agnostic to serve as a standalone, reusable component for any Go project that needs state introspection and monitoring capabilities.

## License

See [LICENSE](LICENSE) file for details.