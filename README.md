# Introspection

A generic Go package for component state introspection, monitoring, and visualization.

## Overview

**Introspection** solves the generic problem of "how any Go component publishes its internal state for visualization and monitoring". Originally extracted from the [lifecycle](https://github.com/aretw0/lifecycle) project, this package provides:

- **Introspectable Interface**: Components can expose their internal state
- **TypedWatcher[S]**: Type-safe state watching with generics
- **State Aggregation**: Combine state changes from multiple components
- **Mermaid Diagram Generation**: Automatic visualization of component topology and state machines

This package is useful for any project that wants to auto-document its topology at runtime and enable live monitoring and debugging.

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
snapshots := introspection.AggregateWatchers(ctx, worker, supervisor, signal)
for snapshot := range snapshots {
    fmt.Printf("Component %s changed state\n", snapshot.ComponentID)
}
```

### 4. Mermaid Diagram Generation

Generate Mermaid diagrams for visualization:

- **Worker Tree Diagram**: Hierarchical view of workers
- **Signal State Machine**: Lifecycle state machine
- **System Diagram**: Complete system topology

```go
// Generate a worker tree diagram
diagram := introspection.WorkerTreeDiagram(workerState)

// Generate a signal state machine
stateMachine := introspection.SignalStateMachine(signalState)

// Generate a complete system diagram
systemDiagram := introspection.SystemDiagram(signalState, workerState)
```

## Installation

```bash
go get github.com/aretw0/instrospection
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/aretw0/instrospection"
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
    return "worker"
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
        state: MyComponentState{Name: "worker-1", Status: "Running"},
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

See the [examples/basic](examples/basic) directory for a complete working example that demonstrates:
- Implementing the `Introspectable` interface
- Using `TypedWatcher` for type-safe state watching
- Aggregating state changes from multiple components
- Generating Mermaid diagrams

Run the example:
```bash
cd examples/basic
go run main.go
```

## Use Cases

- **Observability**: Monitor the state of distributed system components
- **Debugging**: Track state transitions in real-time
- **Documentation**: Auto-generate system topology diagrams
- **Testing**: Verify component behavior through state inspection
- **Monitoring**: Build dashboards that visualize component states

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
diagram := introspection.WorkerTreeDiagram(
    state,
    introspection.WithStyles("classDef custom fill:#fff;"),
)
```

## Testing

Run tests:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -v -cover
```

## Origin

This package was extracted from the [lifecycle](https://github.com/aretw0/lifecycle) project to make it available as a standalone, reusable component for any Go project that needs state introspection and monitoring capabilities.

## License

See [LICENSE](LICENSE) file for details.