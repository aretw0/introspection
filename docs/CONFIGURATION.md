# Configuration Philosophy: Introspection

This document explains the configuration approach used in the introspection package.

## Core Principle

**Configuration Over Convention**: Explicit configuration is preferred over implicit conventions or magic behavior.

## Why Configuration Matters

Different domains and use cases have different needs:

- A **task scheduler** talks about "tasks" and "schedulers"
- A **web server** talks about "handlers" and "middleware"  
- A **worker pool** talks about "workers" and "supervisors"
- A **state machine** talks about "states" and "transitions"

Rather than picking one and forcing all users to adapt, we let **you** define your terminology.

## Configuration Structures

### DiagramConfig

The main configuration for diagram generation:

```go
type DiagramConfig struct {
    // Identity Configuration
    PrimaryID        string  // ID for primary component type
    SecondaryID      string  // ID for secondary component type
    
    // Label Configuration
    PrimaryLabel     string  // Label for primary components section
    SecondaryLabel   string  // Label for secondary components section
    PrimaryNodeLabel string  // Template for primary node labels
    ConnectionLabel  string  // Label for connections between components
    
    // Styling Configuration
    NodeStyler  NodeStylerFunc   // Custom node styling
    NodeLabeler NodeLabelerFunc  // Custom node labeling
}
```

### StateMachineConfig

Configuration for state machine diagrams:

```go
type StateMachineConfig struct {
    // State Names
    InitialState      string  // Name of the initial/running state
    GracefulState     string  // Name of the graceful shutdown state
    ForcedState       string  // Name of the forced termination state
    
    // Transition Labels
    InitialToGraceful string  // Label for graceful shutdown transition
    GracefulToForced  string  // Label for forced termination transition
}
```

## Configuration Patterns

### 1. Required vs. Optional

**Required Fields**: Must be provided for the function to work correctly.

```go
config := &DiagramConfig{
    SecondaryID: "workers",  // REQUIRED for TreeDiagram
}
```

**Optional Fields**: Have sensible defaults if not provided.

```go
config := &DiagramConfig{
    SecondaryID:    "workers",
    SecondaryLabel: "Worker Pool",  // OPTIONAL, defaults to SecondaryID
}
```

### 2. Functional Options

For additional configuration, we use functional options:

```go
type DiagramOption func(*diagramOptions)

func WithStyles(css string) DiagramOption {
    return func(o *diagramOptions) {
        o.styles = css
    }
}

// Usage
diagram := TreeDiagram(state, config,
    WithStyles("classDef custom fill:#fff;"),
)
```

**Benefits**:
- Backward compatible (can add new options)
- Self-documenting
- Optional configuration
- Chainable

### 3. Function Configuration

For maximum flexibility, some configuration uses functions:

```go
type NodeStylerFunc func(
    metadata map[string]string,
) (icon, shapeStart, shapeEnd, cssClass string)

config := &DiagramConfig{
    NodeStyler: func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
        switch metadata["status"] {
        case "active":
            return "🟢", "[", "]", "running"
        case "stopped":
            return "🔴", "[", "]", "stopped"
        default:
            return "⚪", "[", "]", "pending"
        }
    },
}
```

**Benefits**:
- Runtime customization
- Access to component metadata
- Full control over output

## Default Behavior

### When Configuration is Minimal

If you provide minimal configuration, sensible defaults are used:

```go
config := &DiagramConfig{
    SecondaryID: "components",
}

// Defaults applied:
// - SecondaryLabel: "components" (same as ID)
// - NodeStyler: default styler (based on status field)
// - NodeLabeler: default labeler (shows name and status)
```

### Default Node Styling

The default `NodeStyler` recognizes these status values:

- `"Running"` → 🔵 Blue (running state)
- `"Stopped"` → 🟢 Green (stopped/completed)
- `"Failed"` → 🔴 Red (error state)
- `"Pending"` → 🟣 Purple (not started)
- Default → 📋 Gray (unknown)

### Default Node Labeling

The default `NodeLabeler` formats nodes as:

```
{icon} {name} (PID: {pid}) - {status}
```

Example: `🔵 worker-1 (PID: 1234) - Running`

## Configuration Examples

### Example 1: Task Scheduler

```go
config := &DiagramConfig{
    PrimaryID:        "scheduler",
    PrimaryLabel:     "Task Scheduler",
    PrimaryNodeLabel: "🗓️ Scheduler",
    SecondaryID:      "tasks",
    SecondaryLabel:   "Active Tasks",
    ConnectionLabel:  "schedules",
}
```

### Example 2: Web Server

```go
config := &DiagramConfig{
    PrimaryID:        "server",
    PrimaryLabel:     "HTTP Server",
    PrimaryNodeLabel: "🌐 Server",
    SecondaryID:      "handlers",
    SecondaryLabel:   "Request Handlers",
    ConnectionLabel:  "routes to",
}
```

### Example 3: State Machine

```go
config := &StateMachineConfig{
    InitialState:      "Active",
    GracefulState:     "Draining",
    ForcedState:       "Terminated",
    InitialToGraceful: "SHUTDOWN",
    GracefulToForced:  "KILL",
}
```

### Example 4: Custom Styling

```go
config := &DiagramConfig{
    SecondaryID: "components",
    NodeStyler: func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
        componentType := metadata["type"]
        status := metadata["status"]
        
        // Custom icons based on type
        switch componentType {
        case "database":
            icon = "🗄️"
        case "cache":
            icon = "⚡"
        case "api":
            icon = "🌐"
        default:
            icon = "📦"
        }
        
        // Custom shapes based on status
        if status == "critical" {
            shapeStart = "{{"
            shapeEnd = "}}"
            cssClass = "critical"
        } else {
            shapeStart = "["
            shapeEnd = "]"
            cssClass = "normal"
        }
        
        return icon, shapeStart, shapeEnd, cssClass
    },
}
```

## Configuration Best Practices

### 1. Start Simple

Begin with minimal configuration and add customization as needed:

```go
// Start simple
config := &DiagramConfig{
    SecondaryID: "components",
}

// Add customization when needed
config.SecondaryLabel = "Component Pool"
config.NodeStyler = myCustomStyler
```

### 2. Use Descriptive IDs

Choose IDs that reflect your domain:

```go
// ✅ Good
SecondaryID: "tasks"
SecondaryID: "handlers"
SecondaryID: "workers"

// ❌ Bad
SecondaryID: "things"
SecondaryID: "items"
SecondaryID: "stuff"
```

### 3. Leverage Defaults

Only configure what you need to change:

```go
// If default styling works, don't override it
config := &DiagramConfig{
    SecondaryID:    "components",
    SecondaryLabel: "My Components",
    // NodeStyler: (using default)
    // NodeLabeler: (using default)
}
```

### 4. Document Custom Stylers

If using custom styling, document the expectations:

```go
// CustomStyler expects metadata fields:
// - "priority": "high" | "medium" | "low"
// - "health": "healthy" | "degraded" | "down"
func CustomStyler(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
    // Implementation
}
```

## Migration from Legacy APIs

Legacy APIs had hardcoded configuration. Here's how to migrate:

### WorkerTreeDiagram → TreeDiagram

```go
// Before
diagram := WorkerTreeDiagram(state)

// After
config := &DiagramConfig{
    SecondaryID:    "workers",
    SecondaryLabel: "Worker Pool",
}
diagram := TreeDiagram(state, config)
```

### SignalStateMachine → StateMachineDiagram

```go
// Before
diagram := SignalStateMachine(state)

// After
config := &StateMachineConfig{
    InitialState:      "Running",
    GracefulState:     "Stopping",
    ForcedState:       "Stopped",
    InitialToGraceful: "SIGINT",
    GracefulToForced:  "SIGTERM",
}
diagram := StateMachineDiagram(state, config)
```

### SystemDiagram → ComponentDiagram

```go
// Before
diagram := SystemDiagram(signalState, workerState)

// After
config := &DiagramConfig{
    PrimaryID:        "signal",
    PrimaryLabel:     "Signal Handler",
    SecondaryID:      "workers",
    SecondaryLabel:   "Worker Pool",
    ConnectionLabel:  "manages",
}
diagram := ComponentDiagram(signalState, workerState, config)
```

## Future Configuration

Future versions may add:

- **Diagram Options**: Control diagram direction, themes, etc.
- **Export Options**: Different output formats (GraphViz, PlantUML)
- **Filtering Options**: Show/hide certain nodes or edges
- **Animation Options**: For real-time diagrams

All additions will follow the same principles:
- Explicit over implicit
- Functional options for extensibility
- Sensible defaults
- Backward compatibility
