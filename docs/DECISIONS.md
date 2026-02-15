# Design Decisions: Introspection

This document captures key architectural and design decisions made for the introspection package.

## Table of Contents

1. [Domain Agnosticism](#1-domain-agnosticism)
2. [Go Generics for Type Safety](#2-go-generics-for-type-safety)
3. [Channel-Based State Watching](#3-channel-based-state-watching)
4. [Configuration Over Convention](#4-configuration-over-convention)
5. [Mermaid for Visualization](#5-mermaid-for-visualization)
6. [Zero Dependencies](#6-zero-dependencies)
7. [Backward Compatibility](#7-backward-compatibility)
8. [Ubuntu-Only CI Testing](#8-ubuntu-only-ci-testing)
9. [CI Workflow Optimization - Avoid Duplicate Runs](#9-ci-workflow-optimization---avoid-duplicate-runs)

---

## 1. Domain Agnosticism

### Decision

All core APIs must be domain-agnostic. No hardcoded terminology like "worker", "supervisor", "task", etc.

### Rationale

**Problem**: The original implementation (extracted from lifecycle) was tightly coupled to worker/signal terminology, limiting reusability.

**Solution**: 
- Generic interfaces (`Introspectable`, `TypedWatcher[S]`)
- Configurable labels and terminology via `DiagramConfig`
- Type parameters instead of concrete types

### Trade-offs

✅ **Pros**:
- Reusable across different domains
- No conceptual coupling
- Easier to understand for new users

❌ **Cons**:
- More verbose (requires configuration)
- Legacy APIs needed for backward compatibility

### Examples

**Before (Domain-Specific)**:
```go
diagram := WorkerTreeDiagram(workerState)
```

**After (Domain-Agnostic)**:
```go
config := &DiagramConfig{
    SecondaryID: "components",
    SecondaryLabel: "Component Pool",
}
diagram := TreeDiagram(componentState, config)
```

---

## 2. Go Generics for Type Safety

### Decision

Use Go generics (introduced in Go 1.18) for type-safe state watching.

### Rationale

**Problem**: Using `interface{}` for states loses type information and requires runtime type assertions.

**Solution**: Generic `TypedWatcher[S]` interface where `S` is the state type.

```go
type TypedWatcher[S any] interface {
    State() S
    Watch(ctx context.Context) <-chan StateChange[S]
}
```

### Trade-offs

✅ **Pros**:
- Compile-time type safety
- No runtime type assertions
- Better IDE support and documentation
- More explicit contracts

❌ **Cons**:
- Requires Go 1.18+
- Slightly more complex type signatures
- Generic constraints can be verbose

### Impact

Minimum Go version: **1.18** (currently using 1.24.13)

---

## 3. Channel-Based State Watching

### Decision

Use channels for state change notifications instead of callbacks.

### Rationale

**Problem**: Need a concurrency-safe way to notify subscribers of state changes.

**Options Considered**:
1. **Callbacks**: Simple but not idiomatic Go, hard to compose
2. **Channels**: Idiomatic Go, composable, context-aware
3. **Polling**: Inefficient, adds latency

**Decision**: Channels (Option 2)

### Trade-offs

✅ **Pros**:
- Idiomatic Go pattern
- Natural integration with `select` statements
- Easy to compose with other channels
- Context-aware cancellation

❌ **Cons**:
- Need to handle channel cleanup
- Potential for goroutine leaks if not used carefully
- Buffering considerations

### Best Practices

```go
// Always use context for cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

changes := watcher.Watch(ctx)
for change := range changes {
    // Process change
}
```

---

## 4. Configuration Over Convention

### Decision

Prefer explicit configuration over implicit conventions.

### Rationale

**Problem**: Different domains have different needs for labeling, styling, and organization.

**Solution**: Functional options and configuration structs.

```go
type DiagramConfig struct {
    PrimaryID        string
    PrimaryLabel     string
    PrimaryNodeLabel string
    NodeStyler       func(...) (...)
    NodeLabeler      func(...) string
}
```

### Trade-offs

✅ **Pros**:
- Maximum flexibility
- No magic behavior
- Self-documenting
- Easy to test different configurations

❌ **Cons**:
- More verbose for simple cases
- Requires understanding of configuration options

### Mitigation

Provide sensible defaults and common presets:

```go
// Simple case: use defaults
diagram := TreeDiagram(state, &DiagramConfig{
    SecondaryID: "components",
})

// Advanced case: full customization
config := &DiagramConfig{
    SecondaryID: "components",
    NodeStyler: myCustomStyler,
    NodeLabeler: myCustomLabeler,
}
diagram := TreeDiagram(state, config)
```

---

## 5. Mermaid for Visualization

### Decision

Use Mermaid diagram syntax for visualization output.

### Rationale

**Problem**: Need a way to visualize component topology and state.

**Options Considered**:
1. **GraphViz/DOT**: Powerful but requires external tools
2. **ASCII Art**: Limited expressiveness
3. **Mermaid**: Markdown-compatible, widely supported
4. **Custom Format**: Requires building entire ecosystem

**Decision**: Mermaid (Option 3)

### Trade-offs

✅ **Pros**:
- Renders in GitHub, GitLab, many documentation tools
- Human-readable text format
- No binary dependencies
- Wide adoption
- Multiple diagram types (flowchart, state machine, sequence)

❌ **Cons**:
- Syntax can be verbose
- Rendering quality varies by tool
- Limited customization in some renderers

### Future Considerations

Could add output adapters for other formats (GraphViz, PlantUML) without changing core APIs.

---

## 6. Zero Dependencies

### Decision

Depend only on the Go standard library.

### Rationale

**Problem**: External dependencies add maintenance burden, version conflicts, and security concerns.

**Solution**: Use only `stdlib` packages.

### Trade-offs

✅ **Pros**:
- Minimal footprint
- No dependency conflicts
- Faster compilation
- Easier to audit
- Long-term stability

❌ **Cons**:
- May need to implement some utilities ourselves
- Cannot leverage specialized libraries

### Exceptions

None currently. This is a hard requirement for the core package.

Users can build on top with their own dependencies (e.g., metrics libraries, visualization tools).

---

## 7. Backward Compatibility

### Decision

Maintain backward compatibility with legacy APIs while deprecating them.

### Rationale

**Problem**: The package was extracted from lifecycle with specific worker/signal APIs. Breaking existing users is not acceptable.

**Solution**: 
- Keep legacy functions (`WorkerTreeDiagram`, `SignalStateMachine`, `SystemDiagram`)
- Mark as deprecated in documentation
- Internally delegate to new generic APIs
- Plan removal in v1.0

### Migration Path

**v0.1.x**: Legacy APIs available, marked deprecated
**v0.x**: Legacy APIs still available with deprecation warnings
**v1.0**: Legacy APIs removed (with clear migration guide)

### Trade-offs

✅ **Pros**:
- Smooth migration path
- No immediate breaking changes
- Time to update dependent code

❌ **Cons**:
- Maintenance of two API surfaces
- Larger codebase
- Potential confusion for new users

### Deprecation Strategy

```go
// WorkerTreeDiagram generates a Mermaid diagram for worker hierarchies.
//
// Deprecated: Use TreeDiagram with a custom DiagramConfig instead.
// This function will be removed in v1.0.
func WorkerTreeDiagram(state any, options ...DiagramOption) string {
    // Delegate to generic implementation
}
```

---

## 8. Ubuntu-Only CI Testing

### Decision

Run continuous integration tests only on `ubuntu-latest`, not across multiple operating systems.

### Rationale

**Problem**: Initially copied multi-OS CI workflow from lifecycle and procio projects, but those projects have fundamentally different needs.

**Analysis**:
- **lifecycle/procio**: Multi-OS testing is **essential**
  - Platform-specific syscalls (Pdeathsig on Linux, Job Objects on Windows)
  - OS-specific signal handling (SIGINT/SIGTERM vs Ctrl+C)
  - Platform-specific terminal I/O (CONIN$ on Windows, stdin on Unix)
  - Different process management primitives per OS

- **introspection**: Multi-OS testing provides **no value**
  - 100% platform-agnostic code (Go stdlib only)
  - No OS-specific syscalls or primitives
  - Interfaces, channels, reflection, string generation
  - Works identically on all platforms

**Solution**: Use only `ubuntu-latest` for CI testing.

### Trade-offs

✅ **Pros**:
- 3x faster CI execution (~1 minute vs ~3 minutes)
- 67% less GitHub Actions minutes consumption
- Simpler workflow (no matrix strategy)
- Easier to debug CI failures
- Reflects actual project needs
- Faster feedback for developers

❌ **Cons**:
- Won't detect hypothetical platform-specific issues
- Different from sibling projects (but appropriately so)

### Impact

**CI Workflow Before**:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
runs-on: ${{ matrix.os }}
# Runs 3 parallel jobs
```

**CI Workflow After**:
```yaml
runs-on: ubuntu-latest
# Runs 1 job - sufficient for platform-agnostic code
```

### Lessons Learned

Don't blindly copy infrastructure from other projects. Understand the **actual needs** of each project:

- ✅ **lifecycle/procio**: Multi-OS necessary (platform-specific code)
- ✅ **introspection**: Ubuntu sufficient (platform-agnostic code)

Each project should have CI appropriate for its nature, not a one-size-fits-all approach.

---

## 9. CI Workflow Optimization - Avoid Duplicate Runs

### Decision

Configure CI workflow to run on `pull_request` events for all branches, but only on `push` events for the `main` branch.

### Rationale

**Problem**: The original workflow configuration triggered on both `push` and `pull_request` for all branches:

```yaml
on:
  push:
    branches: ["**"]
  pull_request:
    branches: ["**"]
```

This caused **duplicate pipeline runs** when:
- Developer pushes to a feature branch that has an open PR
- Result: 2 identical test runs (push event + PR event)
- Waste of GitHub Actions minutes and processing resources

**Analysis**:
- **Feature branches with PRs**: Only need PR pipeline (already tests the code)
- **Main branch**: Needs push pipeline (no PRs, direct protection)
- **Feature branches without PRs**: Developers can open draft PR for testing

**Solution**: Optimize triggers to run tests exactly once per code change:

```yaml
on:
  push:
    branches: ["main"]
  pull_request:
    branches: ["**"]
```

### Trade-offs

✅ **Pros**:
- ~50% reduction in CI runs for feature branches with PRs
- More efficient use of GitHub Actions minutes
- Faster feedback (no queue waiting for duplicate jobs)
- Encourages PR-based workflow (best practice)
- Main branch still protected with automatic tests

❌ **Cons**:
- Feature branches without PRs won't run CI automatically on push
- Developers must open PR (even draft) to trigger CI

### Impact

**Before**:
```
Feature branch push → CI runs (1)
Open PR → CI runs again (2)
Push to PR branch → CI runs twice (3 + 4)
Total: 4 runs for 2 actual code changes
```

**After**:
```
Feature branch push → No CI
Open PR → CI runs (1)
Push to PR branch → CI runs (2)
Total: 2 runs for 2 actual code changes
```

**Main branch**:
```
Push to main → CI runs (unchanged, still protected)
```

### Best Practices Alignment

This aligns with GitHub's recommended practices:
- Use PRs for code review and testing
- Protect main branch with status checks
- Avoid redundant CI runs
- Optimize for cost and speed

### Resource Savings

For a typical development cycle:
- **Before**: ~8-12 CI runs per feature (push + PR events)
- **After**: ~4-6 CI runs per feature (PR events only)
- **Savings**: ~50% reduction in CI minutes

---

## Future Decisions

These are questions we'll need to answer in future versions:

1. **Persistence**: Should we support persisting state history? If so, what format?
2. **Distribution**: Should we support distributed component introspection?
3. **Performance**: What are acceptable performance characteristics? Need benchmarks.
4. **Metrics**: Should we integrate with standard metrics libraries? How?

## Decision Process

New design decisions should:

1. Align with core principles (domain agnostic, composable, type-safe)
2. Be documented in this file
3. Include rationale and trade-offs
4. Consider backward compatibility
5. Have test coverage

## Feedback

Design decisions are not immutable. If you disagree with a decision or have a better approach, please open an issue or discussion.
