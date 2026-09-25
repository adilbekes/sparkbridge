ARCHIFY_CONFIG ?= ARCHIFY.yaml
ARCHIFY_BIN ?= ./bin/archify-local

.PHONY: init fetch-proto gen-proto arch-check test build-grpc-mqtt build-all

init:
	@mkdir -p pkg/domain pkg/interfaces pkg/spbproto pkg/engine pkg/pipeline \
		pkg/adapters/inputs/json pkg/adapters/inputs/http pkg/adapters/inputs/mqtt pkg/adapters/inputs/grpc \
		pkg/adapters/sinks/stdout pkg/adapters/sinks/file pkg/adapters/sinks/mqtt pkg/adapters/sinks/nats pkg/adapters/sinks/kafka \
		pkg/adapters/storage/filestore pkg/adapters/storage/memorystore pkg/adapters/crypto/aesgcm \
		internal/config internal/cli internal/logger api/proto api/asyncapi api/openapi configs scripts
	@echo "go mod init github.com/<user>/sparkbridge"

fetch-proto:
	@mkdir -p api/proto
	curl -fsSL https://raw.githubusercontent.com/eclipse/tahu/master/sparkplug_b/sparkplug_b.proto -o api/proto/sparkplug_b.proto

gen-proto:
	@mkdir -p pkg/spbproto
	protoc --proto_path=api/proto --go_out=pkg/spbproto --go_opt=paths=source_relative api/proto/sparkplug_b.proto

arch-check:
	@"$(ARCHIFY_BIN)" check --config "$(ARCHIFY_CONFIG)"

test:
	go test -race ./...

build-grpc-mqtt:
	go build -tags "input_grpc,sink_mqtt" -o bin/sparkbridge ./cmd/sparkbridge

build-all:
	go build -tags "all_inputs all_sinks all_storage all_crypto" -o bin/sparkbridge ./cmd/sparkbridge
