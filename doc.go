/*
Package introspection provides a domain-agnostic observation layer for Go applications.
It enables components to expose their internal state for visualization, monitoring, and debugging
without coupling to specific domain concepts.

# Core Concepts

The library is built around three pillars:

 1. State Exposure:
    Components implement simple interfaces to expose their state without domain-specific terminology.

    type Introspectable interface {
        State() any
    }

    type Component interface {
        ComponentType() string
    }

 2. Type-Safe Watching:
    Components can publish state changes through type-safe channels using Go generics.

    type TypedWatcher[S any] interface {
        State() S
        Watch(ctx context.Context) <-chan StateChange[S]
    }

 3. Automatic Visualization:
    Generate Mermaid diagrams from component state with full customization.

    config := &DiagramConfig{
        SecondaryID: "components",
    }
    diagram := TreeDiagram(state, config)

# Key Interfaces

The package defines minimal interfaces for maximum flexibility:

  - Introspectable: Exposes component state
  - Component: Identifies component type
  - TypedWatcher[S]: Type-safe state change notifications
  - EventSource: Event-based notifications

# Visualization

The package includes powerful Mermaid diagram generation:

  - TreeDiagram: Hierarchical component structures
  - ComponentDiagram: Relationships between component types
  - StateMachineDiagram: Component lifecycle and transitions

All visualization is fully customizable through configuration:

	config := &DiagramConfig{
		PrimaryID:        "scheduler",
		PrimaryLabel:     "Task Scheduler",
		PrimaryNodeLabel: "🗓️ Scheduler",
		SecondaryID:      "tasks",
		SecondaryLabel:   "Active Tasks",
		ConnectionLabel:  "schedules",
	}

# State Aggregation

Combine state changes from multiple components:

	watchers := []TypedWatcher[MyState]{component1, component2, component3}
	snapshots := AggregateWatchers(ctx, watchers...)
	for snapshot := range snapshots {
		// Process state changes
	}

# Design Philosophy

The package emphasizes:

  - Domain Agnostic: No hardcoded terminology - you define your domain
  - Type Safety: Leverage Go generics for compile-time safety
  - Composability: Small, focused interfaces that compose well
  - Zero Dependencies: Standard library only
  - Backward Compatible: Legacy APIs remain available

# Examples

See the examples directory for complete working examples:

  - examples/basic: Legacy worker/signal domain
  - examples/generic: Domain-agnostic task scheduler

# Documentation

For detailed documentation:

  - docs/TECHNICAL.md: Architecture and design
  - docs/PRODUCT.md: Vision and use cases
  - docs/DECISIONS.md: Design rationale
  - docs/CONFIGURATION.md: Configuration philosophy
  - docs/RECIPES.md: Common usage patterns
*/
package introspection

