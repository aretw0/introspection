# Product Vision: Introspection

## The Problem Space

Modern Go applications, especially those managing concurrent components (workers, services, tasks, processes), face several observability challenges:

1. **Runtime Topology Blindness**: The architecture exists only in code and documentation, making it hard to understand the actual runtime structure.

2. **State Visibility Gap**: Internal component states are hidden, making debugging and monitoring difficult.

3. **Domain Coupling**: Existing introspection tools are often tightly coupled to specific domain concepts (e.g., "workers", "supervisors"), making them hard to reuse.

4. **Manual Visualization**: Creating diagrams and documentation is a manual, error-prone process that quickly becomes outdated.

## The Solution

**introspection** provides a **domain-agnostic** observation layer that solves these problems through:

### 1. Generic State Exposure

Components implement simple interfaces to expose their state without committing to specific domain terminology:

```go
type Introspectable interface {
    State() any
}

type TypedWatcher[S any] interface {
    State() S
    Watch(ctx context.Context) <-chan StateChange[S]
}
```

### 2. Type-Safe State Watching

Using Go generics, components can publish type-safe state changes that consumers can watch in real-time:

```go
type StateChange[S any] struct {
    ComponentID   string
    ComponentType string
    OldState      S
    NewState      S
    Timestamp     time.Time
}
```

### 3. Automatic Visualization

Generate live Mermaid diagrams that reflect the actual runtime topology:

- **Tree Diagrams**: Hierarchical component structures
- **Component Diagrams**: Relationships between different component types
- **State Machine Diagrams**: Component lifecycle and transitions

### 4. Domain Agnosticism

Full customization of all labels, icons, and terminology through configuration:

```go
config := &DiagramConfig{
    PrimaryID:        "your-component",
    PrimaryLabel:     "Your Label",
    PrimaryNodeLabel: "🎯 Your Icon",
    // ... fully customizable
}
```

## Use Cases

1. **Observability**: Monitor distributed system components in real-time
2. **Debugging**: Track state transitions and identify issues
3. **Documentation**: Auto-generate always-current architecture diagrams
4. **Testing**: Verify component behavior through state inspection
5. **Monitoring**: Build dashboards visualizing component health
6. **Onboarding**: Help new developers understand system architecture

## Key Differentiators

✅ **Domain Agnostic**: Works with any domain model
✅ **Zero Dependencies**: Lightweight, no external dependencies
✅ **Type Safe**: Leverages Go generics for compile-time safety
✅ **Composable**: Small, focused interfaces that compose well
✅ **Backward Compatible**: Legacy APIs remain supported

## Vision

To be the **standard introspection layer** for Go applications that need to expose, monitor, and visualize their internal state—regardless of domain.

## Success Metrics

- Adoption across different domains (not just worker pools)
- Zero coupling to specific architectural patterns
- Useful visualizations generated from minimal code
- Easy integration into existing codebases
