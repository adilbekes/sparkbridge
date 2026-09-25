# SparkBridge Design Patterns

This document is the consolidated architecture reference for mandatory design patterns in SparkBridge. The root `AGENTS.md` is the instruction source; this file provides the readable companion reference.

## Creational Patterns

### Builder Pattern

Location: `pkg/engine/builder.go`

Why it is needed:
Sparkplug B Protobuf payloads (`spbproto.Payload`) contain dozens of optional fields, metrics, datasets, and properties. The Builder prevents error-prone struct initializations and ensures valid payload state before marshaling.

Requirements:

- build payloads incrementally
- set timestamps, sequence numbers, metrics, and properties explicitly
- enforce validation rules prior to binary encoding
- keep the construction flow ordered and testable

### Factory Method Pattern

Location: `pkg/adapters/registry/`

Why it is needed:
Decouples configuration parsing from concrete driver structs. The daemon can instantiate `http`, `mqtt`, `kafka`, or `nats` drivers dynamically at runtime based on string identifiers without hardcoding package dependencies.

Requirements:

- separate config parsing from concrete adapter types
- support runtime selection of inputs, sinks, storage, and crypto implementations
- keep adapter creation isolated from business logic

### Multiton Pattern

Location: `pkg/manager/` or `pkg/bridge/`

Why it is needed:
Enables a single daemon process to manage multiple isolated `SparkBridge` instances in a `map[string]*SparkBridge` registry. Each instance can talk to a different MQTT broker or tenant with its own sequence numbers (`seq`, `bdSeq`) and pipeline workers.

Requirements:

- support multiple independent bridge instances in one binary
- isolate `seq`, `bdSeq`, workers, and adapter configuration per instance
- allow multi-broker and multi-tenant topologies

Example:

- Instance `A` talks to Broker 1 with `GroupID="FactoryA"`
- Instance `B` talks to Broker 2 with `GroupID="FactoryB"`

### Singleton Pattern

Location: `pkg/adapters/registry/`

Why it is needed:
Guarantees a single, thread-safe global adapter registry where build-tagged packages register their factory constructors during Go package `init()`.

Requirements:

- one runtime registry instance (`sync.Once`)
- safe registration from package `init()` functions
- safe concurrent lookup and mutation during startup

### Prototype Pattern

Location: `pkg/domain/`

Why it is needed:
High-frequency telemetry updates (`NDATA`) reuse metric definitions created during `NBIRTH`. Deep-cloning metric prototypes avoids redundant memory allocations on every telemetry pulse.

Requirements:

- deep copy domain objects (`domain.Metric` and `domain.Event`)
- preserve field presence, maps, and pointer semantics
- support high-performance DATA event generation

---

## Structural Patterns

### Bridge Pattern

Location: `pkg/pipeline/` and `pkg/interfaces/`

Why it is needed:
The foundational pattern of the project. Decouples core Sparkplug B encoding, validation, and sequence management (Abstraction) from network transport mechanisms like MQTT, NATS, Kafka, or HTTP (Implementation).

Requirements:

- keep abstraction separate from implementation
- prevent pipeline code from depending on concrete MQTT, Kafka, NATS, or HTTP drivers

### Adapter Pattern

Location: `pkg/adapters/`

Why it is needed:
Wraps third-party client drivers (Eclipse Paho, NATS JetStream, Franz-Kafka) inside standard project interfaces (`InputAdapter`, `OutputSink`).

Requirements:

- isolate Paho, NATS, Kafka, and Gin/HTTP integration details
- keep external APIs behind project interfaces
- translate external protocol events into standardized `domain.Event` structs

### Composite Pattern

Location: `pkg/adapters/sinks/multisink`

Why it is needed:
Allows treating multiple output sinks as a single unified `MultiSink`. The engine treats it as a single output, but messages are automatically fan-out published to all registered child sinks (e.g., MQTT + Kafka simultaneously).

Requirements:

- a composite sink must present the same interface as a leaf sink
- the composite must fan out to child sinks
- support dual-publishing and deterministic failure handling

### Decorator Pattern

Location: `pkg/adapters/sinks/`

Why it is needed:
Dynamically attaches cross-cutting concerns (AES-GCM encryption, Prometheus metrics, structured logging) around output sinks without modifying sink implementations.

Requirements:

- support AES-GCM encryption wrappers
- support metrics collection wrappers
- support structured logging wrappers
- preserve the underlying sink contract

### Facade Pattern

Location: `pkg/sdk/`

Why it is needed:
Provides external Go consumers with a simplified `sparkbridge.Client` entry point, hiding internal Go channels, worker pools, and Protobuf marshaling complexity.

Requirements:

- hide internal channels
- hide protobuf marshaling details
- provide a minimal, stable Go SDK surface

### Proxy Pattern

Location: `pkg/adapters/sinks/buffered`

Why it is needed:
Acts as an offline buffering proxy. Intercepts outbound telemetry during network drops, persists payloads to local disk (SQLite/BadgerDB), and flushes them when connectivity recovers.

Requirements:

- buffer payloads during connectivity outages
- persist buffered payloads to disk
- forward payloads in order when connectivity returns

### Flyweight Pattern

Location: `pkg/engine/alias/`

Why it is needed:
Sparkplug B uses metric aliasing (short integer IDs) to replace long metric name strings in `NDATA` payloads, drastically reducing network bandwidth and RAM usage.

Requirements:

- maintain internal mappings between string names and `uint64` aliases
- reuse alias mappings across consecutive payload builds

---

## Behavioral Patterns

### Mediator Pattern

Location: `pkg/pipeline/`

Why it is needed:
The pipeline acts as a central mediator routing events asynchronously via channels so input adapters and output sinks never communicate directly.

Requirements:

- input adapters must not talk directly to sinks
- pipeline channels should decouple transport from processing

### State Pattern

Location: `pkg/engine/state_machine.go`

Why it is needed:
Models the Sparkplug B Node lifecycle (`OFFLINE`, `NBIRTH`, `ONLINE`, `NDEATH`), dynamically altering payload dispatch behavior based on the current state.

Requirements:

- payload dispatch must depend on current state
- state transitions must be explicit and testable

### Command Pattern

Location: `pkg/domain/`

Why it is needed:
Encapsulates telemetry events as standalone command objects (`Event`, `EncodedMessage`) that can be queued, buffered, and executed asynchronously by worker pools.

Requirements:

- `Event` should represent a unit of work
- telemetry should be queueable and asynchronously processed

### Strategy Pattern

Location: `pkg/engine/strategy/`

Why it is needed:
Allows swapping algorithms at runtime, such as encryption strategies (AES-GCM, RSA, No-Op) or reconnection backoff policies.

Requirements:

- support AES-GCM, RSA, and No-Op encryption strategies
- support pluggable reconnection policies

### Observer Pattern

Location: `pkg/engine/events/`

Why it is needed:
Internal event bus that broadcasts lifecycle events (e.g., connection loss) to trigger automatic reactions like `NDEATH` (Last Will) generation.

Requirements:

- support subscription and notification for lifecycle events
- allow automatic reactions such as `NDEATH` generation on disconnect

### Memento Pattern

Location: `pkg/engine/memento/`

Why it is needed:
Captures snapshots of critical engine state (`bdSeq` sequence counter, active alias tables) to allow state restoration via `StateStore` after service crashes or restarts.

Requirements:

- snapshot `bdSeq`
- snapshot active alias mappings
- restore state via `StateStore`

### Template Method Pattern

Location: `pkg/adapters/base/`

Why it is needed:
Standardizes the execution lifecycle skeleton (`Init` -> `Connect` -> `Process` -> `Close`) for all input adapters and output sinks, ensuring consistent behavior across all drivers.

Requirements:

- standardize lifecycle steps for all input/sink drivers
- allow concrete adapters to override only specific execution hooks

### Visitor Pattern

Location: `pkg/domain/visitor.go`

Why it is needed:
Enables traversing complex nested Sparkplug B metric types (`DataSet`, `Template`, primitives) for validation, printing, or formatting without altering domain struct definitions.

Requirements:

- traverse metric value structures cleanly
- separate algorithms from domain data structures

---

## Concurrency & Resilience Patterns

### Functional Options Pattern

Location: `pkg/bridge/options.go`, `pkg/engine/options.go`

Why it is needed:
Provides an idiomatic Go mechanism for constructing `SparkBridge` and `Engine` instances with flexible optional configurations without breaking backward compatibility or requiring bloated config structs.

Requirements:

- use `type Option func(*SparkBridge)` signature
- provide sensible default values for unconfigured options

### Worker Pool Pattern

Location: `pkg/pipeline/worker_pool.go`

Why it is needed:
Controls concurrency by capping the number of parallel background processing goroutines, preventing Out-Of-Memory (OOM) panics during heavy telemetry spikes.

Requirements:

- bounded goroutine worker pool executing background jobs
- graceful shutdown waiting for active jobs to finish before exiting

### Circuit Breaker Pattern

Location: `pkg/adapters/sinks/resilience/`

Why it is needed:
Prevents cascading failures when downstream brokers crash by failing fast and immediately routing messages to the offline disk proxy instead of hanging on network timeouts.

Requirements:

- implement three states: `Closed` (normal), `Open` (failing fast), `Half-Open` (testing recovery)
- configurable error thresholds and timeout reset intervals

### Fan-Out / Fan-In Pattern

Location: `pkg/pipeline/fan.go`

Why it is needed:
`Fan-Out` duplicates incoming events to multiple internal processors in parallel. `Fan-In` multiplexes events from dozens of input adapters into a single pipeline channel.

Requirements:

- safe channel multiplexing and demultiplexing without race conditions
- non-blocking channel operations to prevent backpressure stalls

### Rate Limiting & Backpressure Pattern

Location: `pkg/pipeline/ratelimit/`

Why it is needed:
Protects downstream brokers and local RAM from traffic spikes by controlling message throughput (Token Bucket) or signaling inputs to slow down reading.

Requirements:

- implement configurable rate limiting algorithms (Token Bucket / Leaky Bucket)
- graceful backpressure policies (drop non-critical metrics vs. buffer vs. block)

---

## Architecture Rules

- keep the consolidated pattern guidance aligned with `AGENTS.md`
- keep Sparkplug protocol-sensitive behavior aligned with `docs/specs/sparkplug_b_spec.pdf` and `docs/specs/sparkplug_b_spec.md`
- keep adapter and engine responsibilities isolated
- keep named bridge instances independent unless explicit sharing is required