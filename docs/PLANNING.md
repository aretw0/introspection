# Planning & Roadmap: Introspection

## Current Version: 0.1.2

This document outlines the current state and future direction of the introspection package.

## Project Status

### ✅ Implemented (v0.1.0)

#### Core Interfaces

- [x] `Introspectable` - Basic state exposure
- [x] `Component` - Component type identification
- [x] `TypedWatcher[S]` - Type-safe state watching with generics
- [x] `EventSource` - Event-based notifications
- [x] `StateChange[S]` - Generic state change representation
- [x] `StateSnapshot` - Point-in-time state capture

#### State Management

- [x] State aggregation from multiple components
- [x] Channel-based state change propagation
- [x] Context-aware cancellation

#### Visualization

- [x] Generic `TreeDiagram` - Hierarchical structures
- [x] Generic `ComponentDiagram` - Component relationships
- [x] Generic `StateMachineDiagram` - State transitions
- [x] Full configuration support via `DiagramConfig`
- [x] Custom node styling and labeling
- [x] Default Mermaid CSS classes for common states

#### Backward Compatibility

- [x] Legacy `WorkerTreeDiagram` (deprecated)
- [x] Legacy `SignalStateMachine` (deprecated)
- [x] Legacy `SystemDiagram` (deprecated)
- [x] Adapter layer for smooth migration

#### Examples

- [x] Basic example (worker/signal domain)
- [x] Generic example (task scheduler domain)

### ✅ Implemented (v0.1.1)

- [x] Add pkg.go.dev example_test.go.

### 🔄 In Progress (v0.1.2)

- [x] Fix module path typo: `instrospection` → `introspection` (PR #6)
- [x] Document post-merge steps for re-tagging and pkg.go.dev (`docs/MODULE_RENAME.md`)
- [x] Add decision record for module rename (`docs/DECISIONS.md` #10)
- [ ] Delete old tags and releases (manual, post-merge)
- [ ] Re-tag as `v0.1.2` (manual, post-merge)
- [ ] Verify pkg.go.dev indexing (manual, post-merge)

## Roadmap

### 🎯 v0.2.0 - Enhanced Visualization

**Focus**: Expand diagram capabilities and customization options.

#### Planned Features

- [ ] **Sequence Diagrams**: Visualize component interactions over time
- [ ] **Graph Layouts**: Support for different Mermaid graph directions (TB, LR, BT, RL)
- [ ] **Conditional Styling**: Style nodes based on runtime conditions
- [ ] **Rich Metadata**: Support for tooltips and extended node information
- [ ] **Diagram Composition**: Combine multiple diagram types

#### Examples to Add

- [ ] Sequence diagram example
- [ ] Multi-domain example (combining different component types)
- [ ] Real-time visualization example (web dashboard)

### 🎯 v0.3.0 - Metrics & Analytics

**Focus**: Add quantitative observation capabilities.

#### Planned Features

- [ ] **State Duration Tracking**: How long components spend in each state
- [ ] **Transition Counting**: Frequency of state transitions
- [ ] **Health Metrics**: Aggregate component health indicators
- [ ] **Anomaly Detection**: Identify unusual state patterns
- [ ] **Metrics Export**: Prometheus/OpenMetrics format support

#### Integration Points

- [ ] Metrics collection interface
- [ ] Pluggable metrics backends
- [ ] Time-series state snapshots

### 🎯 v0.4.0 - Advanced Patterns

**Focus**: Support for complex observation scenarios.

#### Planned Features

- [ ] **State Filtering**: Filter state changes by criteria
- [ ] **State Transformation**: Map/reduce over state changes
- [ ] **State Replay**: Record and replay state change sequences
- [ ] **State Diffing**: Compare states across time or components
- [ ] **Conditional Watching**: Watch only when certain conditions are met

#### Advanced Use Cases

- [ ] A/B testing different component configurations
- [ ] Historical state analysis
- [ ] Distributed system correlation

### 🎯 v1.0.0 - Stability & Production Readiness

**Focus**: Polish, documentation, and production-grade reliability.

#### Requirements for 1.0

- [ ] **Comprehensive Documentation**
  - [ ] Complete API documentation
  - [ ] Architecture guide
  - [ ] Best practices guide
  - [ ] Migration guide from legacy APIs
- [ ] **Performance Optimization**
  - [ ] Benchmark suite
  - [ ] Memory profiling
  - [ ] Channel buffering strategies
- [ ] **Error Handling**
  - [ ] Graceful degradation
  - [ ] Error propagation patterns
  - [ ] Recovery strategies
- [ ] **Testing**
  - [ ] >90% code coverage
  - [ ] Integration tests
  - [ ] Stress tests
  - [ ] Cross-platform validation (Linux, Windows, macOS)
- [ ] **Ecosystem Integration**
  - [ ] Integration examples with lifecycle
  - [ ] Integration examples with procio
  - [ ] Third-party observability tools

## Non-Goals

Things we explicitly **do not** plan to support:

❌ **Domain-Specific Terminology**: We will never hardcode domain terms like "worker", "task", "service" in core APIs.

❌ **Heavy Dependencies**: We will not introduce dependencies on logging frameworks, metrics libraries, or visualization tools.

❌ **Opinionated Architectures**: We will not prescribe specific architectural patterns (microservices, actors, etc.).

❌ **Network Protocols**: We will not implement distributed tracing or network protocols. Users can build these on top.

❌ **UI Components**: We generate diagram definitions, not visual components. Rendering is external.

## Contributing

We welcome contributions that align with the project's philosophy:

1. **Domain Agnostic**: Keep generic, no hardcoded terminology
2. **Minimal Dependencies**: Standard library preferred
3. **Type Safe**: Leverage Go generics where appropriate
4. **Well Tested**: Include tests with race detection
5. **Documented**: Update docs with new features

## Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

## Migration Path

### From Legacy APIs (v0.1.x → v0.2.x)

Legacy functions will remain available through v0.x releases but will be removed in v1.0.

**Recommended Migration**:

1. Replace `WorkerTreeDiagram` with `TreeDiagram` + custom config
2. Replace `SignalStateMachine` with `StateMachineDiagram` + custom config
3. Replace `SystemDiagram` with `ComponentDiagram` + custom config

**Timeline**:

- v0.1.x: Legacy functions available, deprecated
- v0.x: Legacy functions still available with deprecation warnings
- v1.0: Legacy functions removed

## Feedback

We value feedback! Please open issues for:

- Feature requests aligned with our goals
- Bug reports
- Documentation improvements
- Example suggestions

For discussions, see [GitHub Discussions](https://github.com/aretw0/introspection/discussions).
