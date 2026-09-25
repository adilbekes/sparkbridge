# Sparkplug B Specification Reference

Source: `docs/specs/sparkplug_b_spec.pdf` 
Version: Sparkplug 3.0.0 / 2022-11-16

This document extracts the core Sparkplug B rules that SparkBridge must follow. For any Sparkplug B protocol question, the PDF in `docs/specs/` is the primary source of truth.

## Protocol Scope

- Sparkplug B uses MQTT topics and Google Protocol Buffers.
- Sparkplug A is deprecated and is not supported here.
- Sparkplug B messages are binary protobuf payloads carried over MQTT.
- All timestamps in Sparkplug are UTC.

## Topic Namespace

Sparkplug topic structure:

```text
namespace/group_id/message_type/edge_node_id/[device_id]
```

Namespace rules:

- Sparkplug B namespace MUST be `spBv1.0`.
- `namespace` identifies the payload encoding version.

Identifier rules:

- `group_id` MUST be a valid UTF-8 string and MUST NOT contain `+`, `/`, or `#`.
- `edge_node_id` MUST be a valid UTF-8 string and MUST NOT contain `+`, `/`, or `#`.
- `device_id` MUST be a valid UTF-8 string and MUST NOT contain `+`, `/`, or `#`.
- `group_id` + `edge_node_id` MUST be unique across the infrastructure.
- `device_id` MUST be unique within a given Edge Node.

Message type rules:

- `NBIRTH`, `NDEATH`, `NDATA`, and `NCMD` are node-level topics and MUST NOT include `device_id`.
- `DBIRTH`, `DDEATH`, `DDATA`, and `DCMD` are device-level topics and MUST include `device_id`.
- `STATE` is used by Host Applications and is published on `spBv1.0/STATE/[sparkplug_host_id]`.

Topic shapes:

- Node birth: `spBv1.0/{group_id}/NBIRTH/{edge_node_id}`
- Node data: `spBv1.0/{group_id}/NDATA/{edge_node_id}`
- Node death: `spBv1.0/{group_id}/NDEATH/{edge_node_id}`
- Node command: `spBv1.0/{group_id}/NCMD/{edge_node_id}`
- Device birth: `spBv1.0/{group_id}/DBIRTH/{edge_node_id}/{device_id}`
- Device data: `spBv1.0/{group_id}/DDATA/{edge_node_id}/{device_id}`
- Device death: `spBv1.0/{group_id}/DDEATH/{edge_node_id}/{device_id}`
- Device command: `spBv1.0/{group_id}/DCMD/{edge_node_id}/{device_id}`
- Host state: `spBv1.0/STATE/{sparkplug_host_id}`

## Payload Rules

Sparkplug B payload structure includes:

- `timestamp` in UTC milliseconds since epoch.
- `metrics` as an ordered list.
- `seq` as the message sequence number where applicable.
- `uuid` and `body` are optional payload fields.

Metric rules:

- `name` is required unless aliases are being used.
- `alias` is required in NBIRTH and DBIRTH when aliases are used.
- `timestamp` MUST be included in NBIRTH, DBIRTH, NDATA, and DDATA metrics.
- `datatype` MUST be included in NBIRTH and DBIRTH metric definitions.
- `datatype` SHOULD NOT be included in NDATA, NCMD, DDATA, and DCMD metric definitions.

Metric value types supported by Sparkplug B include:

- `uint32`
- `uint64`
- `float`
- `double`
- `bool`
- `string`
- `bytes`
- `DataSet`
- `Template`

Property and complex type rules:

- `PropertySet` stores paired keys and values with matching array lengths.
- `PropertyValue` carries a type and optional null flag.
- `DataSet` uses column/type metadata plus rows of `DataSetValue` entries.
- `Template` definitions are only allowed in NBIRTH.
- `Template` instances must reference their definition and may omit subset members in NDATA/DDATA.

## Data Type Reference

Sparkplug B data types include the following core values:

- `Unknown = 0`
- `Int8 = 1`
- `Int16 = 2`
- `Int32 = 3`
- `Int64 = 4`
- `UInt8 = 5`
- `UInt16 = 6`
- `UInt32 = 7`
- `UInt64 = 8`
- `Float = 9`
- `Double = 10`
- `Boolean = 11`
- `String = 12`
- `DateTime = 13`
- `Text = 14`
- `UUID = 15`
- `DataSet = 16`
- `Bytes = 17`
- `File = 18`
- `Template = 19`
- `PropertySet = 20`
- `PropertySetList = 21`
- `Int8Array = 22`
- `Int16Array = 23`
- `Int32Array = 24`
- `Int64Array = 25`
- `UInt8Array = 26`
- `UInt16Array = 27`
- `UInt32Array = 28`
- `UInt64Array = 29`
- `FloatArray = 30`
- `DoubleArray = 31`
- `BooleanArray = 32`
- `StringArray = 33`
- `DateTimeArray = 34`

Operational mapping notes:

- `Int32` maps to a protobuf `uint32` representation.
- `Int64` maps to a protobuf `uint64` representation.
- `Float` maps to `float`.
- `Double` maps to `double`.
- `Boolean` maps to `bool`.
- `String` maps to `string`.
- `Bytes` maps to `bytes`.

## Sequence Management

Sequence rules:

- `seq` MUST be included in every Sparkplug Edge Node MQTT message except `NDEATH`.
- `NBIRTH` MUST include `seq = 0`.
- `NDATA`, `DBIRTH`, `DDATA`, and `DDEATH` MUST include `seq` unless the message type is explicitly exempted by the spec.
- `seq` MUST increment by one per message and wrap from `255` back to `0`.
- `seq` is an in-memory session counter and MUST NOT be persisted to disk.

Birth/reset behavior:

- `NBIRTH` resets the Edge Node message flow and starts a new `seq` sequence at `0`.
- `DBIRTH` follows `NBIRTH` in the same MQTT session and continues sequence progression.
- `NDATA` and `DDATA` continue sequence progression while the session remains valid.

## bdSeq Rules

Birth/death correlation rules:

- `bdSeq` MUST be included in `NBIRTH` and `NDEATH`.
- `bdSeq` MUST match between the `NBIRTH` and the associated `NDEATH`.
- `bdSeq` MUST start at `0` and increment by one on every new MQTT CONNECT packet.
- `bdSeq` MUST wrap from `255` back to `0`.
- `bdSeq` is tied to the MQTT CONNECT Will Message payload.
- `bdSeq` is used by host applications to correlate a `NDEATH` to the matching `NBIRTH`.

## Message-Specific Constraints

### NBIRTH

- Must be the first MQTT message published by a Sparkplug Edge Node in a session.
- Must use QoS 0.
- Must use retain false.
- Must include payload timestamp.
- Must include all metrics the Edge Node will ever report in that session.
- Must include `seq`.
- Must include `bdSeq` as a metric.
- Must include `Node Control/Rebirth` with boolean false.

### NDATA

- Must use QoS 0.
- Must use retain false.
- Must include payload timestamp.
- Must include `seq`.
- Must only publish metrics that changed since the previous NBIRTH or NDATA.
- Should use report-by-exception and avoid periodic publishing when possible.

### DBIRTH

- Must be published immediately after NBIRTH and before any NDATA/DDATA.
- Must use QoS 0.
- Must use retain false.
- Must include payload timestamp.
- Must include `seq`.
- Must include all metrics the Device will ever report in that session.

### DDATA

- Must use QoS 0.
- Must use retain false.
- Must include payload timestamp.
- Must include `seq`.
- Must only publish metrics that changed since the previous DBIRTH or DDATA.

### NDEATH

- Must not include `seq`.
- Must include `bdSeq`.
- Must be registered as the MQTT CONNECT Will Message.
- Must use Will QoS 1.
- Must use Will retain false.
- Should be published before intentional disconnect.

### DDEATH

- Must include payload timestamp.
- Must include `seq`.
- Must increment sequence like other data messages.
- Must be published when a Device becomes unavailable.

### NCMD and DCMD

- Must include payload timestamp.
- Must not include `seq`.
- Must use QoS 0.
- Must use retain false.

### STATE

- Sparkplug Host Applications MUST publish STATE messages for online/offline state.
- STATE payload is JSON, not Sparkplug protobuf.
- STATE birth uses `online=true` and STATE death uses `online=false`.
- STATE topic is `spBv1.0/STATE/{sparkplug_host_id}`.
- Host Applications must subscribe to their own STATE topic after connecting.

## Session Behavior

- Host Applications and Edge Nodes use MQTT session awareness to synchronize online/offline state.
- Birth certificates establish state.
- Death certificates mark components offline and stale.
- Primary Host Applications influence Edge Node connection behavior in multi-server topologies.
- Clean Session / Clean Start must be enabled for Sparkplug clients.

## Conformance Reminders

- Sparkplug Edge Nodes use the `spBv1.0/#` namespace.
- Host Applications MUST publish STATE birth and death certificates.
- Sparkplug Compliant MQTT Servers MUST support QoS 0, QoS 1, retained messages, and Will Messages.
- Sparkplug Aware MQTT Servers MUST retain NBIRTH and DBIRTH messages as certificates.

## Implementation Guidance for SparkBridge

- Treat this reference as authoritative for Sparkplug behavior.
- Prefer exact payload and topic conformance over shortcuts in engine logic.
- Keep `seq` in memory only.
- Persist `bdSeq` via the project `StateStore`.
- Use `docs/specs/sparkplug_b_spec.pdf` for any ambiguities not captured here.
