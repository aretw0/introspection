# Technical Architecture: Introspection

## Overview

**introspection** is a domain-agnostic observation layer for Go applications. It provides primitives for state exposure, monitoring, and visualization without coupling to specific domain concepts.

## Module Structure

```text
introspection/
├── interfaces.go      # Core interfaces (Introspectable, Component, TypedWatcher, EventSource)
├── types.go           # Core types (StateChange, StateSnapshot, ComponentEvent)
├── adapter.go         # WatcherAdapter for cross-domain aggregation
├── aggregator.go      # Multi-component state aggregation
├── mermaid.go         # Generic Mermaid diagram generation (TreeDiagram, ComponentDiagram, StateMachineDiagram)
├── mermaid_legacy.go  # Deprecated Mermaid functions (WorkerTreeDiagram, SignalStateMachine, SystemDiagram)
├── reflect.go         # Reflection helpers for struct field extraction
├── doc.go             # Package documentation
├── version.go         # Version embedding
└── examples/          # Runnable examples
    ├── basic/         # Legacy worker/signal domain example
    └── generic/       # Domain-agnostic example
```

## Core Concepts

### 1. Introspectable Interface

The most basic interface - any component that can expose its state:

```go
type Introspectable interface {
    State() any
}
```

**Design Rationale**: Minimal contract. Components decide what state to expose and how to represent it.

### 2. Component Interface

Identifies the type of a component:

```go
type Component interface {
    ComponentType() string
}
```

**Design Rationale**: Enables generic categorization without domain-specific terminology.

### 3. TypedWatcher[S]

Type-safe state change notifications using Go generics:

```go
type TypedWatcher[S any] interface {
    State() S
    Watch(ctx context.Context) <-chan StateChange[S]
}
```

**Design Rationale**: 
- Type safety at compile time
- Context-aware for proper cancellation
- Channel-based for natural Go concurrency patterns

### 4. StateChange[S]

Captures state transitions with metadata:

```go
type StateChange[S any] struct {
    ComponentID   string
    ComponentType string
    OldState      S
    NewState      S
    Timestamp     time.Time
}
```

**Design Rationale**: Complete change history with identity and timing information.

### 5. State Aggregation

Combines state changes from multiple components:

```go
func AggregateWatchers[S any](
    ctx context.Context,
    watchers ...TypedWatcher[S],
) <-chan StateSnapshot
```

**Design Rationale**: Fan-in pattern for monitoring multiple components through a single channel.

## Visualization System

### Mermaid Diagram Generation

Three main diagram types, all fully customizable:

#### 1. Tree Diagram

Visualizes hierarchical structures:

```go
func TreeDiagram(
    state any,
    config *DiagramConfig,
    options ...DiagramOption,
) string
```

**Use Case**: Component hierarchies, supervision trees, organizational structures.

#### 2. Component Diagram

Shows relationships between different component types:

```go
func ComponentDiagram(
    primaryState any,
    secondaryState any,
    config *DiagramConfig,
    options ...DiagramOption,
) string
```

**Use Case**: System architectures, component interactions, data flows.

#### 3. State Machine Diagram

Visualizes state transitions:

```go
func StateMachineDiagram(
    state any,
    config *StateMachineConfig,
) string
```

**Use Case**: Lifecycle management, workflow states, process stages.

### Configuration Philosophy

**Full Customization Through Configuration**:

```go
type DiagramConfig struct {
    // Identity
    PrimaryID        string
    SecondaryID      string
    
    // Labels (customize all text)
    PrimaryLabel     string
    SecondaryLabel   string
    PrimaryNodeLabel string
    ConnectionLabel  string
    
    // Styling
    NodeStyler  func(map[string]string) (icon, shapeStart, shapeEnd, cssClass string)
    NodeLabeler func(name, status string, pid int, metadata map[string]string, icon string) string
}
```

**Design Rationale**: 
- No hardcoded terminology
- Full control over visual representation
- Sensible defaults for common cases
- Functional options for advanced customization

## Backward Compatibility

Legacy functions remain available but deprecated:

- `WorkerTreeDiagram()` → use `TreeDiagram()` with custom config
- `SignalStateMachine()` → use `StateMachineDiagram()` with custom config
- `SystemDiagram()` → use `ComponentDiagram()` with custom config

**Migration Path**: All legacy functions internally delegate to generic versions with predefined configurations.

## Design Principles

### 1. Domain Agnosticism

**Principle**: No hardcoded domain terminology.

**Implementation**:
- All labels configurable
- Generic type parameters
- Functional options for customization

### 2. Type Safety

**Principle**: Leverage Go's type system for correctness.

**Implementation**:
- Generic `TypedWatcher[S]` interface
- Compile-time type checking
- No `interface{}` unless necessary

### 3. Composability

**Principle**: Small, focused interfaces that compose well.

**Implementation**:
- Single-responsibility interfaces
- Channel-based communication
- Context-aware cancellation

### 4. Zero Dependencies

**Principle**: Keep the package lightweight.

**Implementation**:
- Standard library only
- No external dependencies
- Minimal footprint

### 5. Observability by Default

**Principle**: Components should be introspectable without extra work.

**Implementation**:
- Simple interfaces
- Automatic aggregation
- Built-in visualization

## Testing Strategy

### Test Coverage

- Unit tests for all public functions
- Race detection enabled (`-race`)
- Example-based tests for documentation

### Test Organization

```text
*_test.go files alongside implementation:
- introspection_test.go      # Core types, interfaces, aggregation
- adapter_test.go            # WatcherAdapter
- mermaid_generic_test.go    # Generic diagram functions
- mermaid_legacy_test.go     # Legacy diagram functions
```

### Testing Philosophy

1. **Behavioral Testing**: Test what components do, not how they do it
2. **Race Detection**: Always run with `-race` flag
3. **Example Tests**: Serve as both tests and documentation

## Future Enhancements

See [PLANNING.md](PLANNING.md) for roadmap and future features.
