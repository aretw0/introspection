package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aretw0/instrospection"
)

// WorkerState represents the state of a worker component.
type WorkerState struct {
	Name     string
	Status   string
	PID      int
	Metadata map[string]string
	Children []WorkerState
}

// SignalState represents the state of a signal handler component.
type SignalState struct {
	Enabled            bool
	Received           any
	ForceExitThreshold int
	SignalCount        int
	Stopping           bool
	Stopped            bool
	Reason             string
	HookTimeout        time.Duration
}

// Worker is a simple component that implements TypedWatcher.
type Worker struct {
	id    string
	state WorkerState
	ch    chan introspection.StateChange[WorkerState]
}

func NewWorker(id, name string) *Worker {
	return &Worker{
		id: id,
		state: WorkerState{
			Name:     name,
			Status:   "Running",
			PID:      12345,
			Metadata: map[string]string{"type": "process"},
		},
		ch: make(chan introspection.StateChange[WorkerState], 10),
	}
}

func (w *Worker) ComponentType() string {
	return "worker"
}

func (w *Worker) State() WorkerState {
	return w.state
}

func (w *Worker) Watch(ctx context.Context) <-chan introspection.StateChange[WorkerState] {
	out := make(chan introspection.StateChange[WorkerState])
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

func (w *Worker) UpdateState(newStatus string) {
	oldState := w.state
	w.state.Status = newStatus
	w.ch <- introspection.StateChange[WorkerState]{
		ComponentID:   w.id,
		ComponentType: "worker",
		OldState:      oldState,
		NewState:      w.state,
		Timestamp:     time.Now(),
	}
}

// SignalHandler is a component that implements TypedWatcher.
type SignalHandler struct {
	id    string
	state SignalState
	ch    chan introspection.StateChange[SignalState]
}

func NewSignalHandler(id string) *SignalHandler {
	return &SignalHandler{
		id: id,
		state: SignalState{
			Enabled:            true,
			ForceExitThreshold: 2,
			HookTimeout:        5 * time.Second,
		},
		ch: make(chan introspection.StateChange[SignalState], 10),
	}
}

func (s *SignalHandler) ComponentType() string {
	return "signal"
}

func (s *SignalHandler) State() SignalState {
	return s.state
}

func (s *SignalHandler) Watch(ctx context.Context) <-chan introspection.StateChange[SignalState] {
	out := make(chan introspection.StateChange[SignalState])
	go func() {
		defer close(out)
		for {
			select {
			case change := <-s.ch:
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

func (s *SignalHandler) Stop() {
	oldState := s.state
	s.state.Stopping = true
	s.state.Reason = "Signal:Terminate"
	s.ch <- introspection.StateChange[SignalState]{
		ComponentID:   s.id,
		ComponentType: "signal",
		OldState:      oldState,
		NewState:      s.state,
		Timestamp:     time.Now(),
	}
}

func main() {
	fmt.Println("Introspection Package Demo")
	fmt.Println("===========================\n")

	// Create components
	worker := NewWorker("worker-1", "MainWorker")
	signal := NewSignalHandler("signal-1")

	// 1. Demonstrate Introspectable interface
	fmt.Println("1. Introspectable Interface")
	fmt.Println("   Worker state:", worker.State())
	fmt.Println("   Signal state:", signal.State())
	fmt.Println()

	// 2. Demonstrate TypedWatcher
	fmt.Println("2. TypedWatcher Interface")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Watch worker state changes
	go func() {
		workerChanges := worker.Watch(ctx)
		for change := range workerChanges {
			fmt.Printf("   [WORKER] State changed: %s -> %s at %s\n",
				change.OldState.Status,
				change.NewState.Status,
				change.Timestamp.Format("15:04:05"))
		}
	}()

	// Watch signal state changes
	go func() {
		signalChanges := signal.Watch(ctx)
		for change := range signalChanges {
			fmt.Printf("   [SIGNAL] Stopping changed: %v -> %v\n",
				change.OldState.Stopping,
				change.NewState.Stopping)
		}
	}()

	// Trigger some state changes
	time.Sleep(100 * time.Millisecond)
	worker.UpdateState("Processing")
	time.Sleep(100 * time.Millisecond)
	worker.UpdateState("Idle")
	time.Sleep(100 * time.Millisecond)
	signal.Stop()

	time.Sleep(200 * time.Millisecond)
	fmt.Println()

	// 3. Demonstrate Aggregator
	fmt.Println("3. State Aggregation")
	aggCtx, aggCancel := context.WithCancel(context.Background())
	defer aggCancel()

	snapshots := introspection.AggregateWatchers(aggCtx, worker, signal)

	// Collect aggregated snapshots
	go func() {
		for snapshot := range snapshots {
			fmt.Printf("   [AGGREGATE] Component: %s (%s) at %s\n",
				snapshot.ComponentID,
				snapshot.ComponentType,
				snapshot.Timestamp.Format("15:04:05"))
		}
	}()

	worker.UpdateState("Stopping")
	time.Sleep(200 * time.Millisecond)
	fmt.Println()

	// 4. Demonstrate Mermaid Diagrams
	fmt.Println("4. Mermaid Diagram Generation")
	fmt.Println()

	// Create a more complex worker tree
	rootState := WorkerState{
		Name:     "supervisor",
		Status:   "Running",
		Metadata: map[string]string{"type": "supervisor"},
		Children: []WorkerState{
			{
				Name:     "worker-1",
				Status:   "Running",
				PID:      1001,
				Metadata: map[string]string{"type": "process"},
			},
			{
				Name:     "worker-2",
				Status:   "Failed",
				PID:      1002,
				Metadata: map[string]string{"type": "container", "image": "app:v1.0"},
			},
		},
	}

	// Generate worker tree diagram
	fmt.Println("   Worker Tree Diagram:")
	fmt.Println("   " + repeatString("─", 50))
	diagram := introspection.WorkerTreeDiagram(rootState)
	for _, line := range splitLines(diagram) {
		fmt.Println("   " + line)
	}
	fmt.Println()

	// Generate signal state machine
	fmt.Println("   Signal State Machine:")
	fmt.Println("   " + repeatString("─", 50))
	stateMachine := introspection.SignalStateMachine(signal.State())
	for _, line := range splitLines(stateMachine) {
		fmt.Println("   " + line)
	}
	fmt.Println()

	// Generate system diagram
	fmt.Println("   System Diagram:")
	fmt.Println("   " + repeatString("─", 50))
	systemDiagram := introspection.SystemDiagram(signal.State(), rootState)
	for _, line := range splitLines(systemDiagram) {
		fmt.Println("   " + line)
	}

	// Cleanup
	cancel()
	aggCancel()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n✅ Demo completed successfully!")
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
