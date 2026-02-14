package introspection

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Test generic TreeDiagram with custom domain
func TestTreeDiagram_GenericDomain(t *testing.T) {
	// Use a generic task/job domain instead of worker
	type TaskState struct {
		Name     string
		Status   string
		PID      int
		Metadata map[string]string
		Children []TaskState
	}

	root := TaskState{
		Name:   "scheduler",
		Status: "Running",
		Metadata: map[string]string{
			"type": "manager",
		},
		Children: []TaskState{
			{
				Name:     "task-1",
				Status:   "Running",
				PID:      2001,
				Metadata: map[string]string{"type": "task"},
			},
			{
				Name:     "task-2",
				Status:   "Failed",
				PID:      2002,
				Metadata: map[string]string{"type": "task"},
			},
		},
	}

	config := DefaultDiagramConfig()
	config.SecondaryID = "scheduler_root"

	diagram := TreeDiagram(root, config)

	expectedStrings := []string{
		"graph TD",
		"scheduler_root{{",    // Manager uses {{ shape
		"⚙️ task-1",           // Task label with icon
		"⚙️ task-2",
		"class scheduler_root running",
		"class scheduler_root_0 running",
		"class scheduler_root_1 failed",
	}

	for _, want := range expectedStrings {
		if !strings.Contains(diagram, want) {
			t.Errorf("TreeDiagram() missing expected string %q", want)
		}
	}
}

// Test ComponentDiagram with custom labels
func TestComponentDiagram_CustomLabels(t *testing.T) {
	type ControllerState struct {
		Enabled  bool
		Stopping bool
		Stopped  bool
		Reason   string
	}

	type ProcessState struct {
		Name     string
		Status   string
		PID      int
		Metadata map[string]string
		Children []ProcessState
	}

	controller := ControllerState{Enabled: true}
	process := ProcessState{
		Name:   "main",
		Status: "Running",
	}

	config := &DiagramConfig{
		PrimaryID:        "ctrl",
		PrimaryLabel:     "Control Layer",
		PrimaryNodeLabel: "🎮 Controller",
		SecondaryID:      "proc",
		SecondaryLabel:   "Process Layer",
		ConnectionLabel:  "coordinates",
		NodeStyler:       defaultNodeStyler,
		NodeLabeler:      defaultNodeLabeler,
	}

	diagram := ComponentDiagram(controller, process, config)

	expectedStrings := []string{
		"subgraph ctrl_graph [Control Layer]",
		"subgraph proc_graph [Process Layer]",
		"🎮 Controller",
		"ctrl -- coordinates --> proc",
	}

	for _, want := range expectedStrings {
		if !strings.Contains(diagram, want) {
			t.Errorf("ComponentDiagram() missing expected string %q", want)
		}
	}
}

// Test StateMachineDiagram with custom configuration
func TestStateMachineDiagram_CustomConfig(t *testing.T) {
	type ServiceState struct {
		ForceExitThreshold int
		Stopping           bool
		Stopped            bool
	}

	state := ServiceState{
		ForceExitThreshold: 2,
		Stopping:           false,
		Stopped:            false,
	}

	config := &StateMachineConfig{
		InitialState:      "Active",
		GracefulState:     "Draining",
		ForcedState:       "Terminated",
		InitialToGraceful: "Shutdown Signal",
		GracefulToForced:  "Kill",
		GracefulToFinal:   "Drained",
	}

	diagram := StateMachineDiagram(state, config)

	expectedStrings := []string{
		"stateDiagram-v2",
		"[*] --> Active",
		"Active --> Draining: Shutdown Signal",
		"Draining --> Terminated: Kill x2",
		"Draining --> [*]: Drained",
		"class Active running",
	}

	for _, want := range expectedStrings {
		if !strings.Contains(diagram, want) {
			t.Errorf("StateMachineDiagram() missing expected string %q", want)
		}
	}
}

// Test custom NodeStyler function
func TestCustomNodeStyler(t *testing.T) {
	customStyler := func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
		if metadata != nil {
			if nodeType, ok := metadata["type"]; ok {
				switch nodeType {
				case "database":
					return "🗄️", "[(", ")]", "container"
				case "api":
					return "🌐", "[/", "/]", "process"
				default:
					return "📋", "[", "]", "process"
				}
			}
		}
		return "📋", "[", "]", "process"
	}

	// Test database style
	icon, start, end, class := customStyler(map[string]string{"type": "database"})
	if icon != "🗄️" || start != "[(" || end != ")]" || class != "container" {
		t.Errorf("customStyler database got (%s, %s, %s, %s)", icon, start, end, class)
	}

	// Test api style
	icon, start, end, class = customStyler(map[string]string{"type": "api"})
	if icon != "🌐" || start != "[/" || end != "/]" || class != "process" {
		t.Errorf("customStyler api got (%s, %s, %s, %s)", icon, start, end, class)
	}
}

// Test custom NodeLabeler function
func TestCustomNodeLabeler(t *testing.T) {
	customLabeler := func(name, status string, pid int, metadata map[string]string, icon string) string {
		label := fmt.Sprintf("%s %s", icon, name)
		if status != "" {
			label += fmt.Sprintf(" [%s]", status)
		}
		return label
	}

	label := customLabeler("service-a", "healthy", 0, nil, "✨")
	expected := "✨ service-a [healthy]"
	if label != expected {
		t.Errorf("customLabeler got %q, want %q", label, expected)
	}
}

// Test TreeDiagram with nil config uses defaults
func TestTreeDiagram_NilConfig(t *testing.T) {
	type NodeState struct {
		Name     string
		Status   string
		Metadata map[string]string
		Children []NodeState
	}

	root := NodeState{
		Name:   "root",
		Status: "Running",
	}

	diagram := TreeDiagram(root, nil)

	if !strings.Contains(diagram, "graph TD") {
		t.Error("TreeDiagram with nil config should still produce valid diagram")
	}
}

// Test ComponentDiagram with nil config uses defaults
func TestComponentDiagram_NilConfig(t *testing.T) {
	type SimpleState struct {
		Enabled bool
	}

	type TreeState struct {
		Name   string
		Status string
	}

	primary := SimpleState{Enabled: true}
	secondary := TreeState{Name: "test", Status: "Running"}

	diagram := ComponentDiagram(primary, secondary, nil)

	expectedStrings := []string{
		"graph TD",
		"subgraph primary_graph [Primary Component]",
		"subgraph secondary_graph [Secondary Component]",
		"primary -- manages --> secondary",
	}

	for _, want := range expectedStrings {
		if !strings.Contains(diagram, want) {
			t.Errorf("ComponentDiagram() with nil config missing %q", want)
		}
	}
}

// Test StateMachineDiagram with nil config uses defaults  
func TestStateMachineDiagram_NilConfig(t *testing.T) {
	type BasicState struct {
		ForceExitThreshold int
		Stopping           bool
		Stopped            bool
	}

	state := BasicState{}

	diagram := StateMachineDiagram(state, nil)

	expectedStrings := []string{
		"stateDiagram-v2",
		"[*] --> Running",
		"Running --> Graceful: Interrupt",
	}

	for _, want := range expectedStrings {
		if !strings.Contains(diagram, want) {
			t.Errorf("StateMachineDiagram() with nil config missing %q", want)
		}
	}
}

// Test that custom note generator works
func TestStateMachineDiagram_WithNoteGenerator(t *testing.T) {
	type StateWithTimeout struct {
		ForceExitThreshold int
		Stopping           bool
		Stopped            bool
		HookTimeout        time.Duration
	}

	state := StateWithTimeout{
		HookTimeout: 5 * time.Second,
	}

	config := DefaultStateMachineConfig()
	config.NoteGenerator = func(s any) string {
		st := s.(StateWithTimeout)
		return fmt.Sprintf("        Timeout: %v\n", st.HookTimeout)
	}

	diagram := StateMachineDiagram(state, config)

	if !strings.Contains(diagram, "Timeout: 5s") {
		t.Error("StateMachineDiagram should include custom note content")
	}
}
