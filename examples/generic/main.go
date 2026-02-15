package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aretw0/introspection"
)

// Example using a Task Scheduling domain (not worker/signal)
// This demonstrates the domain-agnostic nature of the package

// SchedulerState represents a task scheduler component
type SchedulerState struct {
	Enabled  bool
	Stopping bool
	Stopped  bool
	Reason   string
	ForceExitThreshold int
	HookTimeout        time.Duration
}

// TaskState represents a task in the system
type TaskState struct {
	Name     string
	Status   string
	PID      int
	Metadata map[string]string
	Children []TaskState
}

// Scheduler manages tasks
type Scheduler struct {
	id    string
	state SchedulerState
	ch    chan introspection.StateChange[SchedulerState]
}

func NewScheduler(id string) *Scheduler {
	return &Scheduler{
		id: id,
		state: SchedulerState{
			Enabled:            true,
			ForceExitThreshold: 2,
			HookTimeout:        10 * time.Second,
		},
		ch: make(chan introspection.StateChange[SchedulerState], 10),
	}
}

func (s *Scheduler) ComponentType() string {
	return "scheduler"
}

func (s *Scheduler) State() SchedulerState {
	return s.state
}

func (s *Scheduler) Watch(ctx context.Context) <-chan introspection.StateChange[SchedulerState] {
	out := make(chan introspection.StateChange[SchedulerState])
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

func (s *Scheduler) Shutdown() {
	oldState := s.state
	s.state.Stopping = true
	s.state.Reason = "Scheduled Shutdown"
	s.ch <- introspection.StateChange[SchedulerState]{
		ComponentID:   s.id,
		ComponentType: "scheduler",
		OldState:      oldState,
		NewState:      s.state,
		Timestamp:     time.Now(),
	}
}

func main() {
	fmt.Println("Generic Introspection Demo - Task Scheduler Domain")
	fmt.Println("===================================================")
	fmt.Println()

	// Create scheduler
	scheduler := NewScheduler("scheduler-1")

	// Create task hierarchy
	taskTree := TaskState{
		Name:     "MainScheduler",
		Status:   "Running",
		Metadata: map[string]string{"type": "manager"},
		Children: []TaskState{
			{
				Name:     "DataProcessor",
				Status:   "Running",
				PID:      5001,
				Metadata: map[string]string{"type": "task"},
			},
			{
				Name:     "ReportGenerator",
				Status:   "Running",
				PID:      5002,
				Metadata: map[string]string{"type": "task"},
			},
			{
				Name:     "EmailSender",
				Status:   "Failed",
				PID:      5003,
				Metadata: map[string]string{"type": "task", "restarts": "3"},
			},
		},
	}

	// 1. Demonstrate Generic Tree Diagram
	fmt.Println("1. Generic TreeDiagram (Domain-Agnostic)")
	fmt.Println(repeatString("-", 50))
	
	config := &introspection.DiagramConfig{
		SecondaryID: "task_root",
		NodeStyler:  introspection.DefaultDiagramConfig().NodeStyler,
		NodeLabeler: introspection.DefaultDiagramConfig().NodeLabeler,
	}
	
	treeDiagram := introspection.TreeDiagram(taskTree, config)
	printDiagram(treeDiagram)
	fmt.Println()

	// 2. Demonstrate Custom ComponentDiagram
	fmt.Println("2. ComponentDiagram with Custom Labels")
	fmt.Println(repeatString("-", 50))
	
	componentConfig := &introspection.DiagramConfig{
		PrimaryID:        "scheduler",
		PrimaryLabel:     "Task Scheduler",
		PrimaryNodeLabel: "📅 Scheduler",
		SecondaryID:      "tasks",
		SecondaryLabel:   "Task Hierarchy",
		ConnectionLabel:  "orchestrates",
		NodeStyler:       customTaskStyler,
		NodeLabeler:      customTaskLabeler,
	}
	
	componentDiagram := introspection.ComponentDiagram(
		scheduler.State(),
		taskTree,
		componentConfig,
	)
	printDiagram(componentDiagram)
	fmt.Println()

	// 3. Demonstrate Custom State Machine
	fmt.Println("3. StateMachineDiagram with Custom States")
	fmt.Println(repeatString("-", 50))
	
	smConfig := &introspection.StateMachineConfig{
		InitialState:      "Active",
		GracefulState:     "Draining",
		ForcedState:       "Terminated",
		InitialToGraceful: "SHUTDOWN",
		GracefulToForced:  "KILL",
		GracefulToFinal:   "Completed",
		NoteGenerator: func(s any) string {
			state := s.(SchedulerState)
			return fmt.Sprintf("        Tasks Draining\n        Timeout: %v\n", state.HookTimeout)
		},
	}
	
	stateMachine := introspection.StateMachineDiagram(
		scheduler.State(),
		smConfig,
	)
	printDiagram(stateMachine)
	fmt.Println()

	// 4. Demonstrate TypedWatcher with custom domain
	fmt.Println("4. TypedWatcher with Custom Domain")
	fmt.Println(repeatString("-", 50))
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		changes := scheduler.Watch(ctx)
		for change := range changes {
			fmt.Printf("   [SCHEDULER] %s: Stopping=%v, Reason=%s\n",
				change.Timestamp.Format("15:04:05"),
				change.NewState.Stopping,
				change.NewState.Reason)
		}
	}()
	
	time.Sleep(100 * time.Millisecond)
	scheduler.Shutdown()
	time.Sleep(100 * time.Millisecond)
	
	fmt.Println()
	fmt.Println("✅ Demo completed - Package is domain-agnostic!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  • No hardcoded 'worker' or 'signal' terminology")
	fmt.Println("  • Fully customizable labels and styling")
	fmt.Println("  • Works with any domain model (tasks, services, etc.)")
	fmt.Println("  • Composable through configuration")
}

// customTaskStyler provides custom styling for task nodes
func customTaskStyler(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
	if metadata == nil {
		return "📋", "[", "]", "process"
	}
	
	switch metadata["type"] {
	case "manager":
		return "👔", "{{", "}}", "supervisor"
	case "task":
		return "📋", "[", "]", "process"
	default:
		return "⚙️", "[", "]", "process"
	}
}

// customTaskLabeler provides custom label formatting
func customTaskLabeler(name, status string, pid int, metadata map[string]string, icon string) string {
	label := fmt.Sprintf("<b>%s %s</b>", icon, name)
	if status != "" {
		label += fmt.Sprintf("<br/>State: %s", status)
	}
	if pid > 0 {
		label += fmt.Sprintf("<br/>Task ID: %d", pid)
	}
	if metadata != nil {
		if restarts, ok := metadata["restarts"]; ok && restarts != "0" {
			label += fmt.Sprintf("<br/>⚠️ Restarts: %s", restarts)
		}
	}
	return label
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func printDiagram(diagram string) {
	lines := splitLines(diagram)
	for _, line := range lines {
		if line != "" {
			fmt.Println("   " + line)
		}
	}
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
