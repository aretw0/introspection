package introspection_test

import (
	"context"
	"fmt"
	"time"

	introspection "github.com/aretw0/introspection"
)

// Example demonstrates basic usage of the introspection package
// for observing and visualizing component state.
func Example() {
	// Define a simple component state
	type ServiceState struct {
		Name   string
		Status string
	}

	// Create a component that implements TypedWatcher
	service := &simpleWatcher[ServiceState]{
		state: ServiceState{Name: "API", Status: "Running"},
		ch:    make(chan introspection.StateChange[ServiceState], 1),
	}

	// Watch for state changes
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	changes := service.Watch(ctx)

	// Trigger a state change
	service.UpdateState(ServiceState{Name: "API", Status: "Stopped"})

	// Observe the change
	select {
	case change := <-changes:
		fmt.Printf("State changed: %s -> %s\n", change.OldState.Status, change.NewState.Status)
	case <-time.After(50 * time.Millisecond):
	}

	// Output:
	// State changed: Running -> Stopped
}

// ExampleTreeDiagram demonstrates generating a Mermaid diagram
// from a hierarchical data structure.
func ExampleTreeDiagram() {
	// Define a tree structure for tasks
	type Task struct {
		Name     string
		Status   string
		Children []Task
	}

	// Create a task hierarchy
	root := Task{
		Name:   "Project",
		Status: "Active",
		Children: []Task{
			{Name: "Backend", Status: "Running"},
			{Name: "Frontend", Status: "Running"},
		},
	}

	// Generate diagram with configuration
	config := introspection.DefaultDiagramConfig()
	config.SecondaryID = "tasks"

	diagram := introspection.TreeDiagram(root, config)

	// The diagram contains Mermaid markup
	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleComponentDiagram demonstrates creating a diagram showing
// relationships between two components.
func ExampleComponentDiagram() {
	type Controller struct {
		Name string
	}

	type Worker struct {
		ID     int
		Status string
	}

	controller := Controller{Name: "MainController"}
	worker := Worker{ID: 1, Status: "Active"}

	config := introspection.DefaultDiagramConfig()
	config.PrimaryID = "controller"
	config.PrimaryLabel = "Controller"
	config.SecondaryID = "worker"
	config.SecondaryLabel = "Worker"
	config.ConnectionLabel = "manages"

	diagram := introspection.ComponentDiagram(controller, worker, config)

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleStateMachineDiagram demonstrates generating a state machine
// visualization for a component.
func ExampleStateMachineDiagram() {
	type ProcessState struct {
		Running bool
		Stopped bool
	}

	state := ProcessState{Running: true}

	config := introspection.DefaultStateMachineConfig()
	config.InitialState = "Idle"
	config.GracefulState = "Stopping"
	config.ForcedState = "Killed"
	config.InitialToGraceful = "STOP"
	config.GracefulToForced = "KILL"
	config.GracefulToFinal = "Exit"

	diagram := introspection.StateMachineDiagram(state, config)

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleAggregateWatchers demonstrates combining state changes
// from multiple components into a single stream.
func ExampleAggregateWatchers() {
	type ServiceState struct {
		Name   string
		Status string
	}

	// Create multiple watchers
	service1 := &simpleWatcher[ServiceState]{
		id:    "service-1",
		state: ServiceState{Name: "API", Status: "Running"},
		ch:    make(chan introspection.StateChange[ServiceState], 1),
	}

	service2 := &simpleWatcher[ServiceState]{
		id:    "service-2",
		state: ServiceState{Name: "Worker", Status: "Running"},
		ch:    make(chan introspection.StateChange[ServiceState], 1),
	}

	// Aggregate state changes from both services
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	snapshots := introspection.AggregateWatchers(ctx, service1, service2)

	// Trigger state changes
	service1.UpdateState(ServiceState{Name: "API", Status: "Stopped"})
	service2.UpdateState(ServiceState{Name: "Worker", Status: "Idle"})

	// Collect snapshots
	count := 0
	for range snapshots {
		count++
		if count >= 2 {
			cancel()
		}
	}

	fmt.Printf("Received %d snapshots\n", count)

	// Output:
	// Received 2 snapshots
}

// ExampleNewWatcherAdapter demonstrates wrapping a TypedWatcher
// to convert it to StateSnapshot stream for aggregation.
func ExampleNewWatcherAdapter() {
	type ServiceState struct {
		Name   string
		Status string
	}

	service := &simpleWatcher[ServiceState]{
		id:    "api",
		state: ServiceState{Name: "API", Status: "Running"},
		ch:    make(chan introspection.StateChange[ServiceState], 1),
	}

	// Create an adapter
	adapter := introspection.NewWatcherAdapter("service", service)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	snapshots := adapter.Snapshots(ctx)

	// Trigger a state change
	service.UpdateState(ServiceState{Name: "API", Status: "Stopped"})

	// Observe the snapshot
	select {
	case snapshot := <-snapshots:
		fmt.Printf("Snapshot from: %s, Type: %s\n", snapshot.ComponentID, snapshot.ComponentType)
	case <-time.After(50 * time.Millisecond):
	}

	// Output:
	// Snapshot from: api, Type: service
}

// simpleWatcher is a minimal TypedWatcher implementation for examples.
type simpleWatcher[S any] struct {
	id    string
	state S
	ch    chan introspection.StateChange[S]
}

func (w *simpleWatcher[S]) ComponentType() string {
	return "service"
}

func (w *simpleWatcher[S]) State() S {
	return w.state
}

func (w *simpleWatcher[S]) Watch(ctx context.Context) <-chan introspection.StateChange[S] {
	out := make(chan introspection.StateChange[S])
	go func() {
		defer close(out)
		for {
			select {
			case change := <-w.ch:
				select {
				case out <- change:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func (w *simpleWatcher[S]) UpdateState(newState S) {
	oldState := w.state
	w.state = newState
	w.ch <- introspection.StateChange[S]{
		ComponentID:   w.id,
		ComponentType: "service",
		OldState:      oldState,
		NewState:      newState,
		Timestamp:     time.Now(),
	}
}

// ExampleWorkerTreeDiagram demonstrates the legacy WorkerTreeDiagram function
// for backward compatibility with the worker/signal domain.
func ExampleWorkerTreeDiagram() {
	// Define worker state (legacy domain)
	type WorkerState struct {
		Name     string
		Status   string
		PID      int
		Metadata map[string]string
		Children []WorkerState
	}

	root := WorkerState{
		Name:   "supervisor",
		Status: "Running",
		Children: []WorkerState{
			{Name: "worker-1", Status: "Running", PID: 1001},
			{Name: "worker-2", Status: "Idle", PID: 1002},
		},
	}

	diagram := introspection.WorkerTreeDiagram(root)

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleSignalStateMachine demonstrates the legacy SignalStateMachine function
// for backward compatibility with the signal domain.
func ExampleSignalStateMachine() {
	type SignalState struct {
		Enabled            bool
		Stopping           bool
		ForceExitThreshold int
	}

	state := SignalState{
		Enabled:            true,
		ForceExitThreshold: 2,
	}

	diagram := introspection.SignalStateMachine(state)

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleSystemDiagram demonstrates the legacy SystemDiagram function
// that combines signal context and worker tree.
func ExampleSystemDiagram() {
	type SignalState struct {
		Enabled  bool
		Stopping bool
	}

	type WorkerState struct {
		Name     string
		Status   string
		Children []WorkerState
	}

	signal := SignalState{Enabled: true}
	worker := WorkerState{
		Name:   "root",
		Status: "Running",
		Children: []WorkerState{
			{Name: "child-1", Status: "Running"},
		},
	}

	diagram := introspection.SystemDiagram(signal, worker)

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// ExampleAggregateEvents demonstrates combining events from multiple
// event sources into a single stream.
func ExampleAggregateEvents() {
	// Create mock event sources
	source1 := &mockEventSource{
		id: "source-1",
		ch: make(chan introspection.ComponentEvent, 1),
	}
	source2 := &mockEventSource{
		id: "source-2",
		ch: make(chan introspection.ComponentEvent, 1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	events := introspection.AggregateEvents(ctx, source1, source2)

	// Send events from both sources
	source1.SendEvent(&mockEvent{id: "source-1", eventType: "started"})
	source2.SendEvent(&mockEvent{id: "source-2", eventType: "connected"})

	// Collect events
	count := 0
	for range events {
		count++
		if count >= 2 {
			cancel()
		}
	}

	fmt.Printf("Received %d events\n", count)

	// Output:
	// Received 2 events
}

// ExampleWithStyles demonstrates customizing Mermaid diagram styles.
func ExampleWithStyles() {
	type Task struct {
		Name     string
		Status   string
		Children []Task
	}

	root := Task{
		Name:   "Main",
		Status: "Running",
	}

	customStyles := `
    classDef running fill:#90EE90
    classDef failed fill:#FFB6C1
`

	config := introspection.DefaultDiagramConfig()
	diagram := introspection.TreeDiagram(root, config, introspection.WithStyles(customStyles))

	fmt.Println(len(diagram) > 0)

	// Output:
	// true
}

// mockEventSource is a simple EventSource implementation for examples.
type mockEventSource struct {
	id string
	ch chan introspection.ComponentEvent
}

func (m *mockEventSource) Events(ctx context.Context) <-chan introspection.ComponentEvent {
	out := make(chan introspection.ComponentEvent)
	go func() {
		defer close(out)
		for {
			select {
			case event := <-m.ch:
				select {
				case out <- event:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func (m *mockEventSource) SendEvent(event introspection.ComponentEvent) {
	m.ch <- event
}

// mockEvent is a simple ComponentEvent implementation for examples.
type mockEvent struct {
	id        string
	eventType string
}

func (e *mockEvent) ComponentID() string {
	return e.id
}

func (e *mockEvent) ComponentType() string {
	return "service"
}

func (e *mockEvent) Timestamp() time.Time {
	return time.Now()
}

func (e *mockEvent) EventType() string {
	return e.eventType
}
