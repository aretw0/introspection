package introspection

import "context"

// Introspectable is an interface for components that can report their internal state.
// This is used for generating visualization and status reports.
//
// Note: For type-safe state watching, use TypedWatcher[S] instead.
type Introspectable interface {
	// State returns a serializable DTO (Data Transfer Object) representing the component's state.
	State() any
}

// Component identifies the type of a system component.
// Implementing this interface allows the introspection system to correctly classify
// the component (e.g., "processor", "controller", "manager") without relying on package paths.
type Component interface {
	// ComponentType returns the type of the component.
	ComponentType() string
}

// TypedWatcher provides type-safe state watching for a specific state type S.
// Implementations can return their domain-specific state without any type assertions.
type TypedWatcher[S any] interface {
	// State returns the current state snapshot
	State() S

	// Watch returns a channel of type-safe state changes.
	// The channel is closed when the provided context is cancelled.
	Watch(ctx context.Context) <-chan StateChange[S]
}

// EventSource provides an event stream for observability.
type EventSource interface {
	// Events returns a channel of component events.
	// The channel is closed when the provided context is cancelled.
	Events(ctx context.Context) <-chan ComponentEvent
}
