# sparkplugb-agent

`sparkplugb-agent` is an independent edge daemon that bridges internal gRPC clients to external MQTT brokers for Eclipse Sparkplug B v1.0.

## What it does

- Exposes a gRPC bidirectional stream over a Unix domain socket for local edge applications.
- Maintains multiple independent Sparkplug client instances.
- Publishes edge telemetry to external MQTT brokers.
- Receives Sparkplug command traffic from MQTT and forwards it to connected gRPC streams.
- Handles clean shutdown and NBIRTH/NDEATH lifecycle events per client.

## Layout

- `cmd/sparkplugb-agent`: daemon entry point
- `config`: example runtime config
- `internal/config`: config loading and validation
- `internal/grpc`: gRPC Unix socket server
- `internal/mqtt`: MQTT client wrapper
- `internal/sparkplug`: client manager and lifecycle orchestration
- `proto`: internal gRPC contract
- `pkg/pb`: generated protobuf code target

## Build

```bash
go build ./...
```
