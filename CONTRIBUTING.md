# Contributing to SparkBridge

Thank you for contributing to SparkBridge.

## Before You Start

- Read `AGENTS.md` at the repository root.
- Follow the SparkBridge rules for Go version, Sparkplug B protocol handling, and state management.
- Keep changes small, reviewable, and test-backed.

## Working Principles

- Use `google.golang.org/protobuf/proto` for protobuf operations.
- Keep `seq` in memory only and never persist it.
- Persist `bdSeq` through the `StateStore` interface.
- Respect `context.Context` in all I/O and pipeline code.
- Prefer standard library packages unless an external dependency is clearly justified.

## Testing

- Add or update table-driven tests for domain logic, encoders, and state counters.
- Run the relevant package tests before submitting changes.
- Preserve race-safety and thread safety.

## Documentation

- Update README or related docs when the change affects public behavior.
- Add Go doc comments to exported APIs when you introduce them.
- Keep architecture-pattern updates aligned with `docs/architecture/design_patterns.md`.

## Repository Layout

- `cmd/`: binaries and entrypoints.
- `internal/`: application and infrastructure code.
- `pkg/`: reusable packages intended for external use.
