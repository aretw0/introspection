# Recipes: Common Usage Patterns

This document provides practical examples and patterns for using the introspection package.

## Table of Contents

1. [Basic State Exposure](#1-basic-state-exposure)
2. [Type-Safe State Watching](#2-type-safe-state-watching)
3. [Multi-Component Aggregation](#3-multi-component-aggregation)
4. [Custom Visualization](#4-custom-visualization)
5. [Real-Time Monitoring](#5-real-time-monitoring)
6. [Testing with Introspection](#6-testing-with-introspection)
7. [Integration Patterns](#7-integration-patterns)

---

## 1. Basic State Exposure

### Recipe: Making a Component Introspectable

```go
package main

import "github.com/aretw0/introspection"

type MyComponentState struct {
    Name   string
    Status string
    Count  int
}

type MyComponent struct {
    state MyComponentState
}

// Implement Introspectable
func (c *MyComponent) State() any {
    return c.state
}

// Implement Component (optional but recommended)
func (c *MyComponent) ComponentType() string {
    return "processor"
}
```

**Use Case**: Basic state visibility without active watching.

---

## 2. Type-Safe State Watching

### Recipe: Implementing a TypedWatcher

```go
package main

import (
    "context"
    "time"
    "github.com/aretw0/introspection"
)

type TaskState struct {
    ID       string
    Status   string
    Progress int
}

type Task struct {
    state   TaskState
    changes chan introspection.StateChange[TaskState]
}

func NewTask(id string) *Task {
    return &Task{
        state: TaskState{
            ID:     id,
            Status: "pending",
        },
        changes: make(chan introspection.StateChange[TaskState], 10),
    }
}

// TypedWatcher implementation
func (t *Task) State() TaskState {
    return t.state
}

func (t *Task) Watch(ctx context.Context) <-chan introspection.StateChange[TaskState] {
    output := make(chan introspection.StateChange[TaskState])
    
    go func() {
        defer close(output)
        for {
            select {
            case <-ctx.Done():
                return
            case change := <-t.changes:
                select {
                case output <- change:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    
    return output
}

func (t *Task) ComponentType() string {
    return "task"
}

// Update state and notify watchers
func (t *Task) UpdateState(newStatus string, progress int) {
    oldState := t.state
    t.state.Status = newStatus
    t.state.Progress = progress
    
    t.changes <- introspection.StateChange[TaskState]{
        ComponentID:   t.state.ID,
        ComponentType: "task",
        OldState:      oldState,
        NewState:      t.state,
        Timestamp:     time.Now(),
    }
}
```

**Use Case**: Real-time state change notifications with type safety.

---

## 3. Multi-Component Aggregation

### Recipe: Monitoring Multiple Components

```go
package main

import (
    "context"
    "fmt"
    "github.com/aretw0/introspection"
)

func monitorTasks(ctx context.Context, tasks ...*Task) {
    // Convert to TypedWatcher slice
    watchers := make([]introspection.TypedWatcher[TaskState], len(tasks))
    for i, task := range tasks {
        watchers[i] = task
    }
    
    // Aggregate all state changes
    snapshots := introspection.AggregateWatchers(ctx, watchers...)
    
    for snapshot := range snapshots {
        fmt.Printf("[%s] Component %s changed state\n",
            snapshot.Timestamp.Format("15:04:05"),
            snapshot.ComponentID)
        
        // Access state through Payload
        if state, ok := snapshot.Payload.(TaskState); ok {
            fmt.Printf("  Status: %s, Progress: %d%%\n",
                state.Status, state.Progress)
        }
    }
}

// Usage
func main() {
    ctx := context.Background()
    
    task1 := NewTask("task-1")
    task2 := NewTask("task-2")
    task3 := NewTask("task-3")
    
    go monitorTasks(ctx, task1, task2, task3)
    
    // Simulate state changes
    task1.UpdateState("running", 25)
    task2.UpdateState("running", 10)
    task3.UpdateState("completed", 100)
}
```

**Use Case**: Centralized monitoring of multiple related components.

---

## 4. Custom Visualization

### Recipe: Task Scheduler Visualization

```go
package main

import (
    "fmt"
    "github.com/aretw0/introspection"
)

type SchedulerState struct {
    Name   string
    Status string
}

type TaskPoolState struct {
    Tasks []TaskState
}

func visualizeScheduler(schedulerState SchedulerState, taskState TaskPoolState) {
    // Customize for your domain
    config := &introspection.DiagramConfig{
        PrimaryID:        "scheduler",
        PrimaryLabel:     "Task Scheduler",
        PrimaryNodeLabel: "🗓️ Scheduler",
        SecondaryID:      "tasks",
        SecondaryLabel:   "Active Tasks",
        ConnectionLabel:  "schedules",
        
        // Custom node styling
        NodeStyler: func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
            status := metadata["status"]
            priority := metadata["priority"]
            
            // Icon based on priority
            if priority == "high" {
                icon = "⚠️"
            } else {
                icon = "📋"
            }
            
            // Style based on status
            switch status {
            case "running":
                cssClass = "active"
            case "completed":
                cssClass = "done"
            case "failed":
                cssClass = "error"
            default:
                cssClass = "pending"
            }
            
            shapeStart = "["
            shapeEnd = "]"
            
            return icon, shapeStart, shapeEnd, cssClass
        },
    }
    
    diagram := introspection.ComponentDiagram(schedulerState, taskState, config)
    fmt.Println(diagram)
}
```

**Use Case**: Domain-specific visualization with custom styling.

---

## 5. Real-Time Monitoring

### Recipe: Live Dashboard Updates

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/aretw0/introspection"
)

type Dashboard struct {
    components []introspection.Introspectable
}

func (d *Dashboard) RefreshDiagram() string {
    // Collect current states
    var states []any
    for _, comp := range d.components {
        states = append(states, comp.State())
    }
    
    config := &introspection.DiagramConfig{
        SecondaryID: "components",
    }
    
    // Generate diagram from current state
    return introspection.TreeDiagram(states, config)
}

func (d *Dashboard) StartLiveMonitoring(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            diagram := d.RefreshDiagram()
            fmt.Println("\033[2J\033[H") // Clear screen
            fmt.Println(diagram)
        }
    }
}
```

**Use Case**: Periodic diagram updates showing current system state.

---

## 6. Testing with Introspection

### Recipe: Verify Component State in Tests

```go
package main

import (
    "testing"
    "time"
)

func TestTaskProgression(t *testing.T) {
    task := NewTask("test-task")
    
    // Verify initial state
    state := task.State()
    if state.Status != "pending" {
        t.Errorf("Expected initial status 'pending', got '%s'", state.Status)
    }
    
    // Update state
    task.UpdateState("running", 50)
    
    // Verify updated state
    state = task.State()
    if state.Status != "running" {
        t.Errorf("Expected status 'running', got '%s'", state.Status)
    }
    if state.Progress != 50 {
        t.Errorf("Expected progress 50, got %d", state.Progress)
    }
}

func TestStateChangeNotification(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    task := NewTask("test-task")
    changes := task.Watch(ctx)
    
    // Trigger state change
    go func() {
        time.Sleep(100 * time.Millisecond)
        task.UpdateState("completed", 100)
    }()
    
    // Wait for notification
    select {
    case change := <-changes:
        if change.NewState.Status != "completed" {
            t.Errorf("Expected status 'completed', got '%s'", change.NewState.Status)
        }
    case <-ctx.Done():
        t.Error("Timeout waiting for state change notification")
    }
}
```

**Use Case**: Behavior verification through state inspection.

---

## 7. Integration Patterns

### Recipe: Integrating with HTTP Server

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/aretw0/introspection"
)

type Server struct {
    components []introspection.Introspectable
}

// Expose state via HTTP endpoint
func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
    states := make(map[string]any)
    
    for _, comp := range s.components {
        if c, ok := comp.(introspection.Component); ok {
            states[c.ComponentType()] = comp.State()
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(states)
}

// Expose diagram via HTTP endpoint
func (s *Server) handleDiagram(w http.ResponseWriter, r *http.Request) {
    var states []any
    for _, comp := range s.components {
        states = append(states, comp.State())
    }
    
    config := &introspection.DiagramConfig{
        SecondaryID: "components",
    }
    
    diagram := introspection.TreeDiagram(states, config)
    
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(diagram))
}

func (s *Server) RegisterRoutes() {
    http.HandleFunc("/api/state", s.handleState)
    http.HandleFunc("/api/diagram", s.handleDiagram)
}
```

**Use Case**: Exposing introspection data over HTTP for monitoring tools.

---

## Best Practices

### 1. State Immutability

Return copies of state, not references to internal fields:

```go
// ✅ Good
func (c *Component) State() ComponentState {
    return c.state // Returns copy
}

// ❌ Bad
func (c *Component) State() *ComponentState {
    return &c.state // Returns reference
}
```

### 2. Channel Cleanup

Always close channels when done:

```go
func (c *Component) Watch(ctx context.Context) <-chan StateChange {
    ch := make(chan StateChange)
    go func() {
        defer close(ch) // ✅ Always close
        // ... send state changes
    }()
    return ch
}
```

### 3. Context Usage

Always respect context cancellation:

```go
for {
    select {
    case <-ctx.Done():
        return // ✅ Respect cancellation
    case change := <-changes:
        // Process change
    }
}
```

### 4. Buffered Channels

Use buffered channels to prevent blocking:

```go
// ✅ Good for high-frequency updates
changes := make(chan StateChange, 100)

// ❌ May block if consumer is slow
changes := make(chan StateChange)
```

### 5. Type Assertions

Always check type assertions:

```go
if state, ok := payload.(MyState); ok {
    // ✅ Safe to use state
} else {
    // ❌ Handle unexpected type
}
```

---

## Common Pitfalls

### Pitfall 1: Goroutine Leaks

```go
// ❌ Bad: goroutine never exits
func (c *Component) Watch(ctx context.Context) <-chan StateChange {
    ch := make(chan StateChange)
    go func() {
        for change := range c.changes {
            ch <- change // Blocks if nobody reads
        }
    }()
    return ch
}

// ✅ Good: respects context
func (c *Component) Watch(ctx context.Context) <-chan StateChange {
    ch := make(chan StateChange)
    go func() {
        defer close(ch)
        for {
            select {
            case <-ctx.Done():
                return
            case change := <-c.changes:
                select {
                case ch <- change:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return ch
}
```

### Pitfall 2: Race Conditions

```go
// ❌ Bad: concurrent access to shared state
func (c *Component) UpdateState(newState State) {
    c.state = newState // Race condition!
}

// ✅ Good: use mutex or channels
func (c *Component) UpdateState(newState State) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.state = newState
}
```

### Pitfall 3: Nil Channel Reads

```go
// ❌ Bad: nil channel blocks forever
var changes chan StateChange
for change := range changes { // Blocks forever!
    // ...
}

// ✅ Good: check before ranging
if changes != nil {
    for change := range changes {
        // ...
    }
}
```

---

## Additional Resources

- See [examples/basic](../examples/basic) for a complete working example
- See [examples/generic](../examples/generic) for domain-agnostic patterns
- See [TECHNICAL.md](TECHNICAL.md) for architecture details
- See [CONFIGURATION.md](CONFIGURATION.md) for advanced configuration
