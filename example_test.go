package introspection_test

import (
	"context"
	"fmt"
	"time"

	introspection "github.com/aretw0/instrospection"
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
	for snapshot := range snapshots {
		count++
		fmt.Printf("Snapshot from: %s\n", snapshot.ComponentID)
		if count >= 2 {
			cancel()
		}
	}

	// Output:
	// Snapshot from: service-1
	// Snapshot from: service-2
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
