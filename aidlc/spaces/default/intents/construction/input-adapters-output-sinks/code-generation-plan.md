# Stage 3 Code Generation Plan

## Sources

- `aidlc/spaces/default/intents/construction/input-adapters-output-sinks/requirements.md`
- `aidlc/spaces/default/intents/construction/input-adapters-output-sinks/stories.md`
- `AGENTS.md`
- `pkg/interfaces/interfaces.go`
- `pkg/domain/event.go`
- `pkg/adapters/registry/registry.go`

## Plan

1. Scaffold the missing Stage 3 artifact files.
2. Implement the input adapters with build tags and registry init hooks.
3. Implement the output sinks with build tags and registry init hooks.
4. Add focused unit tests for JSON decoding and adapter lookup.
5. Run race-enabled test commands across the target build-tag combinations.
6. Run `make arch-check` to confirm import boundary compliance.

## Unit Breakdown

- `pkg/adapters/inputs/json`: JSON parsing and event conversion
- `pkg/adapters/inputs/http`: HTTP handler and webhook parsing
- `pkg/adapters/inputs/mqtt`: topic + JSON payload translation
- `pkg/adapters/sinks/stdout`: human-readable payload output
- `pkg/adapters/sinks/mqtt`: MQTT publish and LWT wiring
- `pkg/adapters/sinks/nats`: JetStream subject publishing
- `pkg/adapters/sinks/kafka`: topic publishing
- `pkg/adapters/registry`: lookup and registration paths
