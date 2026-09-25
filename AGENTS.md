# SparkBridge AI Context

## Project Scope & Identity

- Project name: SparkBridge
- Primary language: Go 1.27+
- Architecture: Clean Architecture with the standard Go project layout (`cmd/`, `pkg/`, `internal/`)

## Domain & Protocol Rules

- Always use `google.golang.org/protobuf/proto` for Protobuf operations.
- Use `docs/specs/` as the primary Sparkplug protocol reference location.
- Use `docs/specs/sparkplug_b_spec.pdf` and `docs/specs/sparkplug_b_spec.md` as the authoritative Sparkplug B sources.
- Use the official Sparkplug B schema (`sparkplug_b.proto`, proto2) for protocol work.
- Treat optional proto fields strictly as pointers such as `*uint64`, `*string`, and `*float32` so field presence is explicit.
- `seq` is strictly in-memory, implemented with `atomic.Uint64` or `sync.Mutex`, constrained to `0..255`.
- Reset `seq` to `0` immediately after emitting an `NBIRTH` payload.
- Never persist `seq` to disk.
- `bdSeq` must be loaded and incremented through the `StateStore` interface during node initialization.
- `bdSeq` persists as a full session counter across process restarts, while the emitted Sparkplug wire value is constrained to `0..255` via modulo normalization.

## Concurrency & Performance

- Use Go channels such as `chan domain.Event` and `chan EncodedMessage` to decouple pipeline stages.
- Prefer modern Go 1.27+ concurrency primitives for worker pools, cancellation, and coordination.
- All I/O and pipeline workers must accept and respect `context.Context` for cancellation and timeouts.
- Avoid global mutable state.
- Inject dependencies explicitly through constructors such as `NewEngine(...)` and `NewPipeline(...)`.
- Generated code and runtime code must be safe under `go test -race ./...`.
- Aim for zero memory leaks and strict thread safety.

## Build Tags & Adapter Registry

- Every adapter package under `pkg/adapters/inputs/<name>/`, `pkg/adapters/sinks/<name>/`, `pkg/adapters/storage/<name>/`, and `pkg/adapters/crypto/<name>/` must include a build tag header that matches its capability tag plus the matching all-inclusive tag.
- Required build tag forms:
  - `//go:build input_<name> || all_inputs`
  - `//go:build sink_<name> || all_sinks`
  - `//go:build storage_<name> || all_storage`
  - `//go:build crypto_<name> || all_crypto`
- Adapter implementations must auto-register through `pkg/adapters/registry` in package `init()` when compiled.
- The registry must be thread-safe and the CLI/daemon must instantiate inputs and sinks dynamically from runtime YAML configuration.
- If config requests an adapter that was not compiled into the binary, return a clear build-tag error that names the adapter and the missing tag.
- Support modular binaries that omit unneeded adapter dependencies by compiling only the requested build tags.

## Architecture & Design Patterns

- Maintain the consolidated pattern reference in `docs/architecture/design_patterns.md`.
- Builder Pattern (`pkg/engine/builder.go`): use the Builder pattern for constructing `spbproto.Payload` objects step by step before marshaling.
- Factory Method Pattern (`pkg/adapters/registry/`): use Factory Methods to instantiate concrete adapters from configuration strings.
- Multiton Pattern (`pkg/manager/` or `pkg/bridge/`): implement a thread-safe `BridgeManager` using `map[string]*SparkBridge` for named bridge instances.
- Singleton Pattern (`pkg/adapters/registry/`): use a thread-safe Singleton for the global adapter registry.
- Prototype Pattern (`pkg/domain/`): implement deep-copy clone methods on `domain.Metric` and `domain.Event`.
- Bridge Pattern (`pkg/pipeline/`): keep engine/pipeline abstraction separate from network transports.
- Adapter Pattern (`pkg/adapters/`): wrap third-party drivers in `InputAdapter` or `OutputSink` implementations.
- Composite Pattern (`pkg/adapters/sinks/multisink`): support multi-sink dispatch.
- Decorator Pattern (`pkg/adapters/sinks/`): add encryption, metrics, and logging wrappers around sinks.
- Facade Pattern (`pkg/sdk/`): provide a simplified `sparkbridge.Client` SDK entrypoint.
- Proxy Pattern (`pkg/adapters/sinks/buffered`): provide offline disk-buffering for sinks.
- Mediator Pattern (`pkg/pipeline/`): keep inputs and outputs decoupled through the pipeline.
- State Pattern (`pkg/engine/state_machine.go`): model node lifecycle transitions.
- Command Pattern (`pkg/domain/`): encapsulate telemetry events as commands for workers.
- Strategy Pattern (`pkg/engine/strategy/`): support interchangeable encryption and reconnect strategies.
- Observer Pattern (`pkg/engine/events/`): provide lifecycle notifications via an internal event bus.
- Memento Pattern (`pkg/engine/memento/`): restore `bdSeq` and alias state after restart.
- Chain of Responsibility (`pkg/pipeline/middleware/`): process event middleware in sequence.
- Template Method Pattern (`pkg/adapters/base/`): standardize adapter lifecycle hooks.
- Visitor Pattern (`pkg/domain/visitor.go`): traverse Sparkplug metric types for validation and serialization.
- Flyweight Pattern (`pkg/engine/alias`): reuse metric alias mappings to reduce allocations and payload size.

## Code Quality & Testing

- Write table-driven unit tests for domain logic, encoders, and state counters.
- Minimize external dependencies.
- Prefer the Go standard library wherever feasible, especially `net/http`, `sync`, `context`, `time`, and `crypto/aes`.
- Add meaningful Go doc comments to all exported types, functions, and interfaces.

## Repository Notes

- The module name is `sparkbridge`.
- Public packages should remain small and explicit.
- Keep code aligned with Sparkplug B lifecycle rules and deterministic state handling.
- The adapter registry lives at `pkg/adapters/registry`.
