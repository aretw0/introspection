package introspection

import (
	"fmt"
	"reflect"
	"strings"
)

// MermaidOption configures the rendering behavior.
type MermaidOption func(*MermaidOptions)

type MermaidOptions struct {
	Styles string // Custom Mermaid class definitions
}

// DefaultStyles returns the standard Mermaid class definitions for lifecycle diagrams.
func DefaultStyles() string {
	return `    classDef created fill:#f8f9fa,stroke:#dee2e6,color:#6c757d;
    classDef pending fill:#eef2ff,stroke:#c7d2fe,color:#4338ca;
    classDef starting fill:#cfe2ff,stroke:#b8d4ff,color:#004085;
    classDef running fill:#d1ecf1,stroke:#bee5eb,color:#0c5460;
    classDef suspended fill:#fff3cd,stroke:#ffe69c,color:#856404;
    classDef stopping fill:#f8d7da,stroke:#f5c6cb,color:#721c24;
    classDef stopped fill:#e9ecef,stroke:#adb5bd,color:#495057;
    classDef finished fill:#d4edda,stroke:#c3e6cb,color:#155724;
    classDef killed fill:#343a40,stroke:#212529,color:#ffffff;
    classDef failed fill:#f8d7da,stroke:#f5c6cb,color:#721c24;
    classDef container stroke-width:3px,stroke-dasharray: 0;
    classDef process stroke-width:1px;
    classDef goroutine stroke-dasharray: 5 5;
    classDef supervisor stroke-width:2px,stroke-dasharray: 0;
    classDef signal stroke-width:2px,stroke-dasharray: 0;
    classDef active fill:#eef2ff,stroke:#4338ca,stroke-width:2px;
`
}

// WithStyles allows custom Mermaid class definitions.
func WithStyles(styles string) MermaidOption {
	return func(o *MermaidOptions) {
		o.Styles = styles
	}
}

// DiagramConfig holds configuration for customizing diagram rendering.
type DiagramConfig struct {
	// Primary component configuration
	PrimaryID        string // Node ID for primary component (default: "primary")
	PrimaryLabel     string // Subgraph label for primary component (default: "Primary Component")
	PrimaryNodeLabel string // Label prefix for primary node (default: "⚡ Component")
	
	// Secondary component configuration  
	SecondaryID        string // Root node ID for secondary component (default: "secondary")
	SecondaryLabel     string // Subgraph label for secondary component (default: "Secondary Component")
	
	// Connection configuration
	ConnectionLabel string // Label for edge between components (default: "manages")
	
	// Node style customization
	NodeStyler NodeStyleFunc // Custom function to style nodes based on metadata
	NodeLabeler NodeLabelFunc // Custom function to build node labels
}

// NodeStyleFunc is a function that returns icon, shape start, shape end, and CSS class for a node.
type NodeStyleFunc func(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string)

// NodeLabelFunc is a function that builds the label for a node.
type NodeLabelFunc func(name, status string, pid int, metadata map[string]string, icon string) string

// DefaultDiagramConfig returns a generic configuration with no domain-specific terms.
func DefaultDiagramConfig() *DiagramConfig {
	return &DiagramConfig{
		PrimaryID:        "primary",
		PrimaryLabel:     "Primary Component",
		PrimaryNodeLabel: "⚡ Component",
		SecondaryID:      "secondary",
		SecondaryLabel:   "Secondary Component",
		ConnectionLabel:  "manages",
		NodeStyler:       defaultNodeStyler,
		NodeLabeler:      defaultNodeLabeler,
	}
}

// defaultNodeStyler provides default node styling based on metadata type field.
func defaultNodeStyler(metadata map[string]string) (icon, shapeStart, shapeEnd, cssClass string) {
	nodeType := "process"
	if metadata != nil {
		if t, ok := metadata["type"]; ok {
			nodeType = t
		}
	}
	
	switch nodeType {
	case "supervisor", "manager", "coordinator":
		return "🧠", "{{", "}}", "supervisor"
	case "process", "task":
		return "⚙️", "[", "]", "process"
	case "container", "pod":
		return "📦", "[[", "]]", "container"
	case "func", "goroutine", "function":
		return "λ", "(", ")", "goroutine"
	default:
		return "⚙️", "[", "]", "process"
	}
}

// defaultNodeLabeler provides default node label formatting.
func defaultNodeLabeler(name, status string, pid int, metadata map[string]string, icon string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("<b>%s %s</b>", icon, name))
	if status != "" {
		parts = append(parts, fmt.Sprintf("Status: %s", status))
	}
	if pid > 0 {
		parts = append(parts, fmt.Sprintf("PID: %d", pid))
	}
	if metadata != nil {
		if image, ok := metadata["image"]; ok && image != "" {
			parts = append(parts, fmt.Sprintf("Image: %s", image))
		}
		if restarts, ok := metadata["restarts"]; ok && restarts != "0" {
			parts = append(parts, fmt.Sprintf("🔄 Restarts: %s", restarts))
		}
	}
	return strings.Join(parts, "<br/>")
}

// ComponentDiagram renders a customizable topology diagram with two components.
// This is a generic version that allows full customization of labels and styling.
func ComponentDiagram(primary, secondary any, config *DiagramConfig, opts ...MermaidOption) string {
	if config == nil {
		config = DefaultDiagramConfig()
	}
	
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}
	
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	
	// 1. Primary Component Subgraph
	sb.WriteString(fmt.Sprintf("    subgraph %s_graph [%s]\n", config.PrimaryID, config.PrimaryLabel))
	renderGenericFragment(&sb, primary, config.PrimaryID, config.PrimaryNodeLabel, "        ")
	sb.WriteString("    end\n\n")
	
	// 2. Secondary Component Subgraph
	sb.WriteString(fmt.Sprintf("    subgraph %s_graph [%s]\n", config.SecondaryID, config.SecondaryLabel))
	renderGenericTree(&sb, secondary, config.SecondaryID, config.NodeStyler, config.NodeLabeler, "        ")
	sb.WriteString("    end\n\n")
	
	// 3. Connection
	sb.WriteString(fmt.Sprintf("    %s -- %s --> %s\n", config.PrimaryID, config.ConnectionLabel, config.SecondaryID))
	
	// 4. Styles
	sb.WriteString(options.Styles)
	
	return sb.String()
}

// TreeDiagram returns a generic Mermaid diagram representing a hierarchical tree structure.
// The structure is introspected via reflection using common field names (Name, Status, PID, Metadata, Children).
func TreeDiagram(root any, config *DiagramConfig, opts ...MermaidOption) string {
	if config == nil {
		config = DefaultDiagramConfig()
	}
	
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}
	
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	sb.WriteString(options.Styles)
	renderGenericTree(&sb, root, config.SecondaryID, config.NodeStyler, config.NodeLabeler, "    ")
	return sb.String()
}

// StateMachineDiagram renders a generic Mermaid state diagram.
// It introspects the state object via reflection to find relevant fields.
type StateMachineConfig struct {
	// State names
	InitialState  string // Default: "Running"
	GracefulState string // Default: "Graceful"
	ForcedState   string // Default: "ForceExit"
	
	// Transition labels
	InitialToGraceful string // Default: "Interrupt"
	GracefulToForced  string // Default: "Force"
	GracefulToFinal   string // Default: "Complete"
	
	// Note content generator
	NoteGenerator func(state any) string
}

// DefaultStateMachineConfig returns a generic state machine configuration.
func DefaultStateMachineConfig() *StateMachineConfig {
	return &StateMachineConfig{
		InitialState:      "Running",
		GracefulState:     "Graceful",
		ForcedState:       "ForceExit",
		InitialToGraceful: "Interrupt",
		GracefulToForced:  "Force",
		GracefulToFinal:   "Complete",
	}
}

// StateMachineDiagram renders a customizable Mermaid state diagram.
func StateMachineDiagram(state any, config *StateMachineConfig, opts ...MermaidOption) string {
	if config == nil {
		config = DefaultStateMachineConfig()
	}
	
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}
	
	var sb strings.Builder
	v := reflect.ValueOf(state)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	forceExitThreshold := getIntField(v, "ForceExitThreshold")
	stopping := getBoolField(v, "Stopping")
	stopped := getBoolField(v, "Stopped")
	
	sb.WriteString("stateDiagram-v2\n")
	sb.WriteString(fmt.Sprintf("    [*] --> %s\n", config.InitialState))
	sb.WriteString(fmt.Sprintf("    %s --> %s: %s\n", config.InitialState, config.GracefulState, config.InitialToGraceful))
	
	// Add note if provided
	if config.NoteGenerator != nil {
		note := config.NoteGenerator(state)
		if note != "" {
			sb.WriteString(fmt.Sprintf("    note right of %s\n", config.GracefulState))
			sb.WriteString(note)
			sb.WriteString("    end note\n")
		}
	}
	
	if forceExitThreshold > 0 {
		sb.WriteString(fmt.Sprintf("    %s --> %s: %s x%d\n", config.GracefulState, config.ForcedState, config.GracefulToForced, forceExitThreshold))
		sb.WriteString(fmt.Sprintf("    %s --> [*]: Exit\n", config.ForcedState))
	}
	
	sb.WriteString(fmt.Sprintf("    %s --> [*]: %s\n", config.GracefulState, config.GracefulToFinal))
	
	// Apply state classes
	if stopped {
		sb.WriteString("    class [*] stopped\n")
	} else if stopping {
		sb.WriteString(fmt.Sprintf("    class %s stopping\n", config.GracefulState))
	} else {
		sb.WriteString(fmt.Sprintf("    class %s running\n", config.InitialState))
	}
	
	sb.WriteString(options.Styles)
	return sb.String()
}

// renderGenericFragment renders a single component node (for primary/controller type components).
func renderGenericFragment(sb *strings.Builder, comp any, id, labelPrefix, indent string) {
	v := reflect.ValueOf(comp)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	enabled := getBoolField(v, "Enabled")
	stopping := getBoolField(v, "Stopping")
	stopped := getBoolField(v, "Stopped")
	reason := getStringField(v, "Reason")
	
	statusMode := "Running"
	statusClass := "running"
	
	if !enabled {
		statusMode = "Disabled"
		statusClass = "stopped"
	} else if stopped {
		statusMode = "Stopped"
		statusClass = "stopped"
	} else if stopping {
		statusMode = "Stopping"
		statusClass = "pending"
	}
	
	if reason == "" {
		reason = "None"
	}
	
	label := fmt.Sprintf("<b>%s</b><br/>Mode: %s", labelPrefix, statusMode)
	if reason != "None" {
		label += fmt.Sprintf("<br/>Reason: %s", reason)
	}
	
	sb.WriteString(fmt.Sprintf("%s%s[\"%s\"]:::signal\n", indent, id, label))
	sb.WriteString(fmt.Sprintf("%sclass %s %s\n", indent, id, statusClass))
}

// renderGenericTree renders a hierarchical tree structure.
func renderGenericTree(sb *strings.Builder, root any, rootID string, styler NodeStyleFunc, labeler NodeLabelFunc, indent string) {
	renderGenericNode(sb, root, rootID, styler, labeler, indent)
}

// renderGenericNode renders a single node and recursively renders its children.
func renderGenericNode(sb *strings.Builder, node any, id string, styler NodeStyleFunc, labeler NodeLabelFunc, indent string) {
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	name := getStringField(v, "Name")
	status := getStringField(v, "Status")
	pid := getIntField(v, "PID")
	metadata := getMapField(v, "Metadata")
	children := getSliceField(v, "Children")
	
	icon, shapeStart, shapeEnd, idClass := styler(metadata)
	label := labeler(name, status, pid, metadata, icon)
	
	sb.WriteString(fmt.Sprintf("%s%s%s\"%s\"%s:::%s\n", indent, id, shapeStart, label, shapeEnd, idClass))
	
	statusClass := strings.ToLower(status)
	if statusClass == "" {
		statusClass = "pending"
	}
	sb.WriteString(fmt.Sprintf("%sclass %s %s\n", indent, id, statusClass))
	
	for i, child := range children {
		childID := fmt.Sprintf("%s_%d", id, i)
		renderGenericNode(sb, child, childID, styler, labeler, indent)
		sb.WriteString(fmt.Sprintf("%s%s --> %s\n", indent, id, childID))
	}
}

// SystemDiagram renders a full system topology diagram combining signal context and worker tree.
// Deprecated: Use ComponentDiagram with custom DiagramConfig for domain-agnostic diagrams.
// Accepts signal.State and worker.State (or pointers to them) as any.
func SystemDiagram(sig, work any, opts ...MermaidOption) string {
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}

	var sb strings.Builder

	sb.WriteString("graph TD\n")

	// 1. Signal Context Subgraph
	sb.WriteString("    subgraph ControlPlane [Signal Context]\n")
	renderSignalFragment(&sb, sig, "S", "        ")
	sb.WriteString("    end\n\n")

	// 2. Worker Subgraph (The worker tree)
	sb.WriteString("    subgraph DataPlane [Supervision Tree]\n")
	renderWorkerFragment(&sb, work, "root", "        ")
	sb.WriteString("    end\n\n")

	// 3. Connection
	sb.WriteString("    S -- governs --> root\n")

	// 4. Styles
	sb.WriteString(options.Styles)

	return sb.String()
}

// SignalStateMachine renders a Mermaid state diagram for the signal context.
// Deprecated: Use StateMachineDiagram with custom StateMachineConfig for domain-agnostic diagrams.
func SignalStateMachine(sig any, opts ...MermaidOption) string {
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}

	var sb strings.Builder

	v := reflect.ValueOf(sig)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	forceExitThreshold := getIntField(v, "ForceExitThreshold")
	hookTimeout := getField(v, "HookTimeout")
	stopping := getBoolField(v, "Stopping")
	received := getField(v, "Received")

	sb.WriteString("stateDiagram-v2\n")
	sb.WriteString("    [*] --> Running\n")

	signals := "SIGTERM"
	if forceExitThreshold == 1 {
		signals = "SIGINT/SIGTERM"
	}
	sb.WriteString(fmt.Sprintf("    Running --> Graceful: %s\n", signals))

	sb.WriteString("    note right of Graceful\n")
	sb.WriteString("        Context Cancelled\n")
	sb.WriteString("        Hooks Running (LIFO)\n")
	sb.WriteString(fmt.Sprintf("        Timeout: %v\n", hookTimeout))
	sb.WriteString("    end note\n")

	if forceExitThreshold > 0 {
		sb.WriteString(fmt.Sprintf("    Graceful --> ForceExit: Signal x%d\n", forceExitThreshold))
		sb.WriteString("    ForceExit --> [*]: os.Exit(1)\n")
	}

	sb.WriteString("    Graceful --> [*]: Hooks Complete\n")

	stoppedState := getBoolField(v, "Stopped")
	if stoppedState {
		sb.WriteString("    class [*] stopped\n")
	} else if stopping {
		sb.WriteString("    class Graceful stopping\n")
	} else if received == nil {
		sb.WriteString("    class Running running\n")
	} else {
		rv := reflect.ValueOf(received)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				sb.WriteString("    class Running running\n")
			}
		}
	}

	if received != nil {
		rv := reflect.ValueOf(received)
		isNil := false
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface ||
			rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map ||
			rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
			isNil = rv.IsNil()
		}
		if !isNil {
			sb.WriteString(fmt.Sprintf("    note left of Graceful: Received %v\n", received))
		}
	}

	sb.WriteString(options.Styles)
	return sb.String()
}

// WorkerTreeDiagram returns a Mermaid diagram string representing the worker hierarchy.
// Deprecated: Use TreeDiagram with custom DiagramConfig for domain-agnostic diagrams.
func WorkerTreeDiagram(s any, opts ...MermaidOption) string {
	options := &MermaidOptions{Styles: DefaultStyles()}
	for _, opt := range opts {
		opt(options)
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")
	sb.WriteString(options.Styles)
	renderWorkerNode(&sb, s, "root", "    ")
	return sb.String()
}

func renderSignalFragment(sb *strings.Builder, sig any, id, indent string) {
	v := reflect.ValueOf(sig)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	enabled := getBoolField(v, "Enabled")
	stopping := getBoolField(v, "Stopping")
	stoppedState := getBoolField(v, "Stopped")
	received := getField(v, "Received")
	reason := getStringField(v, "Reason")

	statusMode := "Running"
	statusClass := "running"

	if !enabled {
		statusMode = "Disabled"
		statusClass = "stopped"
	} else if stoppedState {
		statusMode = "Stopped"
		statusClass = "stopped"
	} else if stopping {
		statusMode = "Stopping"
		statusClass = "pending"

		if reason == "Signal:Terminate" {
			statusClass = "failed"
		}
	}

	receivedStr := "None"
	if received != nil {
		rv := reflect.ValueOf(received)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface ||
			rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map ||
			rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
			if !rv.IsNil() {
				receivedStr = fmt.Sprintf("%v", received)
			}
		} else {
			receivedStr = fmt.Sprintf("%v", received)
		}
	}

	if reason == "" {
		reason = "None"
	}

	label := fmt.Sprintf("<b>⚡ Lifecycle Controller</b><br/>Mode: %s", statusMode)
	if receivedStr != "None" {
		label += fmt.Sprintf("<br/>Received: %s", receivedStr)
	}
	if reason != "None" {
		label += fmt.Sprintf("<br/>Reason: %s", reason)
	}

	sb.WriteString(fmt.Sprintf("%s%s[\"%s\"]:::signal\n", indent, id, label))
	sb.WriteString(fmt.Sprintf("%sclass %s %s\n", indent, id, statusClass))
}

func renderWorkerFragment(sb *strings.Builder, s any, rootID string, indent string) {
	renderWorkerNode(sb, s, rootID, indent)
}

func renderWorkerNode(sb *strings.Builder, s any, id, indent string) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	name := getStringField(v, "Name")
	status := getStringField(v, "Status")
	pid := getIntField(v, "PID")
	metadata := getMapField(v, "Metadata")
	children := getSliceField(v, "Children")

	icon, shapeStart, shapeEnd, idClass := getWorkerNodeStyle(metadata)
	label := buildWorkerNodeLabel(name, status, pid, metadata, icon)

	sb.WriteString(fmt.Sprintf("%s%s%s\"%s\"%s:::%s\n", indent, id, shapeStart, label, shapeEnd, idClass))

	statusClass := strings.ToLower(status)
	if statusClass == "" {
		statusClass = "pending"
	}
	sb.WriteString(fmt.Sprintf("%sclass %s %s\n", indent, id, statusClass))

	for i, child := range children {
		childID := fmt.Sprintf("%s_%d", id, i)
		renderWorkerNode(sb, child, childID, indent)
		sb.WriteString(fmt.Sprintf("%s%s --> %s\n", indent, id, childID))
	}
}

func getWorkerNodeStyle(metadata map[string]string) (icon, shapeStart, shapeEnd, idClass string) {
	workerType := "process"
	if metadata != nil {
		if t, ok := metadata["type"]; ok {
			workerType = t
		}
	}

	switch workerType {
	case "supervisor":
		icon, shapeStart, shapeEnd, idClass = "🧠", "{{", "}}", "supervisor"
	case "process":
		icon, shapeStart, shapeEnd, idClass = "⚙️", "[", "]", "process"
	case "container":
		icon, shapeStart, shapeEnd, idClass = "📦", "[[", "]]", "container"
	case "func", "goroutine":
		icon, shapeStart, shapeEnd, idClass = "λ", "(", ")", "goroutine"
	default:
		icon, shapeStart, shapeEnd, idClass = "⚙️", "[", "]", "process"
	}
	return
}

func buildWorkerNodeLabel(name string, status string, pid int, metadata map[string]string, icon string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("<b>%s %s</b>", icon, name))
	if status != "" {
		parts = append(parts, fmt.Sprintf("Status: %s", status))
	}
	if pid > 0 {
		parts = append(parts, fmt.Sprintf("PID: %d", pid))
	}
	if metadata != nil {
		if image, ok := metadata["image"]; ok && image != "" {
			parts = append(parts, fmt.Sprintf("Image: %s", image))
		}
		if restarts, ok := metadata["restarts"]; ok && restarts != "0" {
			parts = append(parts, fmt.Sprintf("🔄 Restarts: %s", restarts))
		}
	}
	return strings.Join(parts, "<br/>")
}

// Reflection helpers
func getIntField(v reflect.Value, name string) int {
	field := v.FieldByName(name)
	if field.IsValid() && field.CanInt() {
		return int(field.Int())
	}
	return 0
}

func getBoolField(v reflect.Value, name string) bool {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Bool {
		return field.Bool()
	}
	return false
}

func getStringField(v reflect.Value, name string) string {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.String {
		return field.String()
	}
	return ""
}

func getField(v reflect.Value, name string) any {
	field := v.FieldByName(name)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}
	return nil
}

func getMapField(v reflect.Value, name string) map[string]string {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Map {
		result := make(map[string]string)
		iter := field.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			if k.Kind() == reflect.String && v.Kind() == reflect.String {
				result[k.String()] = v.String()
			}
		}
		return result
	}
	return nil
}

func getSliceField(v reflect.Value, name string) []any {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Slice {
		result := make([]any, field.Len())
		for i := 0; i < field.Len(); i++ {
			result[i] = field.Index(i).Interface()
		}
		return result
	}
	return nil
}
