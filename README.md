# SparkBridge

SparkBridge is an open-source, high-performance Go SDK and edge gateway for the Eclipse Sparkplug B specification.

It is designed to operate in two modes:

- Go SDK: an embeddable library for applications that want Sparkplug B encoding, routing, and lifecycle management inside their own process.
- Standalone Daemon / CLI: an "Any-to-Sparkplug" declarative gateway that accepts upstream inputs, normalizes them into Sparkplug events, and forwards encoded payloads to downstream transports.

The repository currently ships a daemon-oriented entrypoint and internal packages that provide the foundation for the SDK/gateway split.

## Core Architecture

SparkBridge uses an asynchronous, decoupled pipeline built around Go channels and worker pools.

```mermaid
flowchart LR
  A[Input Adapters\nJSON | HTTP | Raw MQTT | gRPC] --> B[chan Event]
  B --> C[Engine Worker Pool\nProto Marshal | AES-GCM | seq increments]
  C --> D[chan EncodedMessage]
  D --> E[Output Sinks\nMQTT | NATS | Kafka | Disk]
```

ASCII fallback:

```text
Input Adapters (JSON, HTTP, Raw MQTT, gRPC)
  -> chan Event
  -> Engine Worker Pool (Proto Marshal, Encryption, seq increments)
  -> chan EncodedMessage
  -> Output Sinks (MQTT, NATS, Kafka, Disk)
```

## Key Features & Design Principles

- Decoupled input/output pipeline using Go channels and worker pools.
- Strict compliance with the Eclipse Sparkplug B specification, using Protobuf v2 as the wire format basis.
- Clear state management rules for Sparkplug lifecycle counters.
- Optional AES-GCM payload encryption before dispatch.
- Architecture and design-pattern guidance is consolidated in `docs/architecture/design_patterns.md`.

### Core State Rules

- `seq`: thread-safe, stored strictly in memory, constrained to `0..255`, and automatically reset to `0` on `NBIRTH`.
- `bdSeq`: persisted across restarts through a `StateStore` interface, constrained to `0..255`.

These rules keep Sparkplug lifecycle behavior deterministic while allowing the runtime to remain stateless where appropriate and durable where required.

## Project Roadmap

### Stage 1: Core Architecture, Canonical Domain Models & Asynchronous Pipeline

- Define the canonical domain model for events, messages, lifecycle transitions, and sink payloads.
- Establish the channel-based processing pipeline and worker pool boundaries.
- Finalize the separation between SDK APIs and daemon orchestration.

### Stage 2: Engine State Management, Protobuf Encoding & Topic Builder

- Implement the `seq` and `bdSeq` state model.
- Add Sparkplug B Protobuf encoding.
- Build deterministic topic construction for outbound MQTT-compatible routing.

### Stage 3: MVP CLI, JSON Ingestion & Local File/Stdout Sink

- Ship a minimal CLI for local execution.
- Accept JSON input payloads.
- Support local file and stdout sinks for development and verification.

### Stage 4: MQTT Network Sink & Native Last Will and Testament (LWT) Support

- Add an MQTT sink for network delivery.
- Implement native LWT support for edge presence semantics.
- Validate reconnect and publish behavior under broker interruptions.

### Stage 5: Input Adapters (HTTP Webhooks, Raw MQTT Translator, gRPC)

- Add HTTP webhook ingestion.
- Add a raw MQTT translator for inbound Sparkplug-compatible traffic.
- Add gRPC ingestion for embedded and service-to-service use cases.

### Stage 6: Multi-Sink Outputs (NATS JetStream, Apache Kafka) & LWT Heartbeat Emulation

- Add NATS JetStream output support.
- Add Apache Kafka output support.
- Emulate heartbeat behavior for deployments that need derived LWT-like monitoring.

### Stage 7: Production Readiness (AES Encryption, Disk Offline Buffering, Metric Aliasing)

- Enable AES-GCM payload encryption.
- Add disk-backed offline buffering for disconnected operation.
- Implement metric aliasing for production-scale topic and payload optimization.

## Getting Started

This section is intentionally left as a contributor-facing placeholder until the CLI, SDK surface, and packaging story are finalized.

### Installation

TODO: document package installation and versioning.

### CLI Usage

TODO: document gateway startup, config flags, and runtime examples.

### Go Module Imports

TODO: document the public Go module path and package usage patterns.

### Build Commands

```bash
make init
make gen-proto
```

TODO: document the full build, test, and release workflow.

## Repository Layout

- `cmd/sparkbridge`: current daemon entrypoint.
- `internal/config`: configuration loading and validation.
- `internal/grpc`: gRPC server implementation.
- `internal/mqtt`: MQTT client wrapper.
- `internal/sparkplug`: client manager and lifecycle orchestration.
- `pkg/pb`: generated protobuf output.

## Contribution Guidelines

- Keep changes aligned with the documented architecture and state rules.
- Prefer small, reviewable changes that preserve Sparkplug semantics.
- Add or update tests when touching state management, encoding, or transport behavior.
- Treat the roadmap as phased guidance; do not introduce later-stage dependencies into earlier-stage work unless explicitly needed.
- Adapter packages are expected to use build tags and auto-register through `pkg/adapters/registry` so the CLI/daemon can compose modular binaries from YAML config.

## Vision

SparkBridge aims to become a dependable Sparkplug B foundation for both application developers and edge operators:

- as a library, it should make correct Sparkplug behavior easy to embed;
- as a gateway, it should make declarative, high-throughput Sparkplug bridging easy to run;
- as a platform, it should keep the control plane simple while preserving protocol correctness, durability, and performance.
