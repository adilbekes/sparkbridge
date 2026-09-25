# Stage 3 Requirements: Input Adapters & Output Sinks

## Sources

- Task request from the active feature intent
- `AGENTS.md`
- `docs/architecture/design_patterns.md`
- `docs/specs/sparkplug_b_spec.md`

## Requirements

### FR-1: JSON File Input

The system shall provide a `json` input adapter that can consume JSON payloads from a watched file or stdin and convert them into `domain.Event` values.

Acceptance Criteria:

- The adapter compiles only when `input_json` or `all_inputs` is set.
- The adapter auto-registers itself with the adapter registry.
- The adapter can decode JSON with `group_id`, `node_id`, `device_id`, `msg_type`, `timestamp`, and `metrics` fields.

### FR-2: HTTP Webhook Input

The system shall provide an `http` input adapter that accepts webhook POST requests and converts them into `domain.Event` values.

Acceptance Criteria:

- The adapter compiles only when `input_http` or `all_inputs` is set.
- The adapter auto-registers itself with the adapter registry.
- The adapter exposes a `net/http` handler suitable for webhook ingestion.

### FR-3: Raw MQTT Input

The system shall provide a `mqtt` input adapter that subscribes to raw MQTT topics and converts JSON payloads into `domain.Event` values.

Acceptance Criteria:

- The adapter compiles only when `input_mqtt` or `all_inputs` is set.
- The adapter auto-registers itself with the adapter registry.
- The adapter can extract a topic and JSON body into a domain event.

### FR-4: Stdout Sink

The system shall provide a `stdout` sink for local debugging that prints encoded payloads in a human-readable representation.

Acceptance Criteria:

- The sink compiles only when `sink_stdout` or `all_sinks` is set.
- The sink auto-registers itself with the adapter registry.
- The sink writes to standard output without requiring external services.

### FR-5: MQTT Sink

The system shall provide an `mqtt` sink that can publish Sparkplug payloads and support Last Will and Testament configuration.

Acceptance Criteria:

- The sink compiles only when `sink_mqtt` or `all_sinks` is set.
- The sink auto-registers itself with the adapter registry.
- The sink accepts an LWT payload configuration.

### FR-6: NATS JetStream Sink

The system shall provide a `nats` sink that can publish encoded messages to JetStream subjects.

Acceptance Criteria:

- The sink compiles only when `sink_nats` or `all_sinks` is set.
- The sink auto-registers itself with the adapter registry.

### FR-7: Kafka Sink

The system shall provide a `kafka` sink that can publish encoded messages to Kafka topics.

Acceptance Criteria:

- The sink compiles only when `sink_kafka` or `all_sinks` is set.
- The sink auto-registers itself with the adapter registry.

## Non-Functional Requirements

- All adapters and sinks must respect `context.Context` for cancellation.
- Adapter and sink registration must be thread-safe.
- Build-tag coverage must remain explicit and modular.
- The new packages must not violate Clean Architecture boundaries.
