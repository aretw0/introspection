package introspection

import "time"

// StateChange represents a typed state transition.
// The generic parameter S allows type-safe access to state without assertions.
type StateChange[S any] struct {
	ComponentID   string
	ComponentType string // Component type identifier (e.g., "processor", "controller", "manager")
	OldState      S
	NewState      S
	Timestamp     time.Time
}

// StateSnapshot is the envelope for cross-domain aggregation.
// It unifies different state types via a common wrapper.
type StateSnapshot struct {
	ComponentID   string
	ComponentType string // Component type identifier (e.g., "processor", "controller", "manager")
	Timestamp     time.Time
	Payload       any // Component state as any type
}

// ComponentEvent is the interface for event sourcing.
// Every event must provide identification and timing metadata.
type ComponentEvent interface {
	ComponentID() string
	ComponentType() string
	Timestamp() time.Time
	EventType() string
}
