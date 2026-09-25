# Stage 3 Unit Test Instructions

## Sources

- `aidlc/spaces/default/intents/construction/input-adapters-output-sinks/code-generation-plan.md`
- `pkg/adapters/registry/registry.go`

## Commands

Run the following test sets:

1. Base race suite:

```bash
go test -race ./...
```

2. Inputs + sinks focus:

```bash
go test -race -tags "input_json,input_http,input_mqtt,sink_stdout,sink_mqtt,sink_nats,sink_kafka" ./...
```

3. Adapter lookup verification:

```bash
go test -race -tags "input_http,sink_mqtt" ./pkg/adapters/registry ./pkg/adapters/inputs/http ./pkg/adapters/sinks/mqtt
```

## Expected Outcomes

- JSON decoding tests should pass.
- Registry lookup tests should resolve compiled adapters.
- Architecture boundaries should pass under `make arch-check`.
