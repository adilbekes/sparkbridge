# Stage 3 User Stories: Input Adapters & Output Sinks

## Sources

- `aidlc/spaces/default/intents/construction/input-adapters-output-sinks/requirements.md`
- `AGENTS.md`

## Stories

### Story 1: Ingest JSON from a File

As a developer, I want to ingest JSON from a file or stdin so that I can quickly feed test data into SparkBridge.

Acceptance Criteria:

- Given valid JSON input, when the adapter reads it, then it produces a `domain.Event`.
- Given invalid JSON input, when the adapter reads it, then it returns an error.

### Story 2: Accept HTTP Webhooks

As an operator, I want to send telemetry to an HTTP endpoint so that I can integrate existing services without an MQTT client.

Acceptance Criteria:

- Given a valid webhook POST, when the handler receives it, then it converts the body into a `domain.Event`.
- Given an unsupported method, when the handler receives it, then it rejects the request.

### Story 3: Translate Raw MQTT Messages

As an integrator, I want to subscribe to raw MQTT topics and translate their payloads into domain events so that I can ingest non-Sparkplug telemetry.

Acceptance Criteria:

- Given a raw MQTT topic and JSON payload, when the adapter receives it, then it emits a `domain.Event`.

### Story 4: Debug Encoded Payloads Locally

As a developer, I want a stdout sink so that I can inspect encoded payloads during local development.

Acceptance Criteria:

- Given an encoded payload, when the sink publishes it, then it prints a readable representation.

### Story 5: Publish to MQTT, NATS, and Kafka

As an operator, I want dedicated output sinks for MQTT, NATS JetStream, and Kafka so that I can route encoded messages into my downstream infrastructure.

Acceptance Criteria:

- Given a sink configuration, when the sink starts, then it registers itself and accepts publish calls.
- Given a publish request, when the sink processes it, then it returns success or a clear error.
