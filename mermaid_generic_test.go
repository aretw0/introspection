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

// Test custom PrimaryNodeStyler function
func TestCustomPrimaryNodeStyler(t *testing.T) {
	type CustomState struct {
		Signal   string
		IsFailed bool
	}

	// Custom styler that checks for failure state
	customStyler := func(state any) string {
		if st, ok := state.(CustomState); ok {
			if st.IsFailed {
				return "failed"
			}
			if st.Signal == "SIGTERM" {
				return "stopping"
			}
		}
		return "running"
	}

	// Test failed state
	state1 := CustomState{IsFailed: true}
	cssClass := customStyler(state1)
	if cssClass != "failed" {
		t.Errorf("customStyler for failed state got %q, want %q", cssClass, "failed")
	}

	// Test stopping state (with signal)
	state2 := CustomState{Signal: "SIGTERM"}
	cssClass = customStyler(state2)
	if cssClass != "stopping" {
		t.Errorf("customStyler for SIGTERM got %q, want %q", cssClass, "stopping")
	}

	// Test running state
	state3 := CustomState{}
	cssClass = customStyler(state3)
	if cssClass != "running" {
		t.Errorf("customStyler for normal state got %q, want %q", cssClass, "running")
	}
}

// Test custom PrimaryNodeLabeler function
func TestCustomPrimaryNodeLabeler(t *testing.T) {
	type CustomState struct {
		Signal   string
		Received int
	}

	// Custom labeler that shows signal information
	customLabeler := func(state any) string {
		if st, ok := state.(CustomState); ok {
			if st.Signal != "" {
				return fmt.Sprintf("Signal: %s<br/>Received: %d times", st.Signal, st.Received)
			}
		}
		return "No signal received"
	}

	// Test with signal
	state1 := CustomState{Signal: "SIGINT", Received: 3}
	label := customLabeler(state1)
	expected := "Signal: SIGINT<br/>Received: 3 times"
	if label != expected {
		t.Errorf("customLabeler got %q, want %q", label, expected)
	}

	// Test without signal
	state2 := CustomState{}
	label = customLabeler(state2)
	expected = "No signal received"
	if label != expected {
		t.Errorf("customLabeler got %q, want %q", label, expected)
	}
}

// Test ComponentDiagram with custom primary node styling
func TestComponentDiagram_CustomPrimaryNodeStyler(t *testing.T) {
	type SignalState struct {
		Signal   string
		Received int
		IsFailed bool
	}

	type ProcessState struct {
		Name   string
		Status string
	}

	signalState := SignalState{Signal: "SIGTERM", Received: 2}
	processState := ProcessState{Name: "worker", Status: "Running"}

	// Custom styler that uses Signal field
	customStyler := func(state any) string {
		if st, ok := state.(SignalState); ok {
			if st.IsFailed {
				return "failed"
			}
			if st.Signal != "" {
				return "stopping"
			}
		}
		return "running"
	}

	// Custom labeler that displays signal info
	customLabeler := func(state any) string {
		if st, ok := state.(SignalState); ok {
			if st.Signal != "" {
				return fmt.Sprintf("Signal: %s<br/>Received: %d", st.Signal, st.Received)
			}
		}
		return "Active"
	}

	config := DefaultDiagramConfig()
	config.PrimaryNodeStyler = customStyler
	config.PrimaryNodeLabeler = customLabeler
	config.PrimaryID = "signal"
	config.PrimaryLabel = "Signal Handler"
	config.PrimaryNodeLabel = "📡 Handler"

	diagram := ComponentDiagram(signalState, processState, config)

	// Verify custom label is present
	if !strings.Contains(diagram, "Signal: SIGTERM") {
		t.Error("ComponentDiagram should contain custom signal label")
	}
	if !strings.Contains(diagram, "Received: 2") {
		t.Error("ComponentDiagram should contain received count")
	}

	// Verify custom CSS class is applied
	if !strings.Contains(diagram, "class signal stopping") {
		t.Error("ComponentDiagram should apply custom CSS class 'stopping'")
	}
}

// Test backward compatibility - default behavior with reflection
func TestComponentDiagram_DefaultPrimaryNodeBehavior(t *testing.T) {
	type LegacyControllerState struct {
		Enabled  bool
		Stopping bool
		Stopped  bool
		Reason   string
	}

	type ProcessState struct {
		Name   string
		Status string
	}

	// Test various states with default behavior
	testCases := []struct {
		name          string
		state         LegacyControllerState
		expectedClass string
		expectedLabel string
	}{
		{
			name:          "Running state",
			state:         LegacyControllerState{Enabled: true},
			expectedClass: "running",
			expectedLabel: "Mode: Running",
		},
		{
			name:          "Stopping state",
			state:         LegacyControllerState{Enabled: true, Stopping: true, Reason: "Shutdown"},
			expectedClass: "pending",
			expectedLabel: "Mode: Stopping<br/>Reason: Shutdown",
		},
		{
			name:          "Stopped state",
			state:         LegacyControllerState{Enabled: true, Stopped: true, Reason: "Complete"},
			expectedClass: "stopped",
			expectedLabel: "Mode: Stopped<br/>Reason: Complete",
		},
		{
			name:          "Disabled state",
			state:         LegacyControllerState{Enabled: false},
			expectedClass: "stopped",
			expectedLabel: "Mode: Disabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processState := ProcessState{Name: "test", Status: "Running"}
			config := DefaultDiagramConfig()
			diagram := ComponentDiagram(tc.state, processState, config)

			if !strings.Contains(diagram, tc.expectedLabel) {
				t.Errorf("Expected label %q not found in diagram", tc.expectedLabel)
			}
			if !strings.Contains(diagram, fmt.Sprintf("class primary %s", tc.expectedClass)) {
				t.Errorf("Expected CSS class %q not found in diagram", tc.expectedClass)
			}
		})
	}
}

// Test that nil config still works (backward compatibility)
func TestComponentDiagram_NilConfigWithCustomState(t *testing.T) {
	type ControllerState struct {
		Enabled  bool
		Stopping bool
		Stopped  bool
		Reason   string
	}

	type ProcessState struct {
		Name   string
		Status string
	}

	controller := ControllerState{Enabled: true, Stopping: true, Reason: "Graceful"}
	process := ProcessState{Name: "main", Status: "Running"}

	diagram := ComponentDiagram(controller, process, nil)

	// Should use default config with reflection-based behavior
	if !strings.Contains(diagram, "Mode: Stopping") {
		t.Error("ComponentDiagram with nil config should use default primary labeler")
	}
	if !strings.Contains(diagram, "Reason: Graceful") {
		t.Error("ComponentDiagram with nil config should include reason")
	}
	if !strings.Contains(diagram, "class primary pending") {
		t.Error("ComponentDiagram with nil config should apply correct CSS class")
	}
}
