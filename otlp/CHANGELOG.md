# Changelog

All notable changes to the OTLP exporters (`github.com/standards-lab/go-observability/otlp`) are
documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). This changelog
covers this sub-module only; the base module keeps its own.

## [Unreleased]

## [v0.4.0] - 2026-11-11

### Changed

- **Breaking:** The `go-observability` requirement is v0.4.0, which requires `go-core` v0.12.0
  and pulls go-core's v0.12.0 breaks into an importer's build: `reactor.New` and
  `event.NewRecorder` take a trailing `reactor.Options` and `event.Options` in place of variadic
  option functions, `reactor.CancelWithin` returns a `time.Duration`, and the
  `reactor.CancelAfter`, `event.Clock`, and `event.IDs` option functions are removed. The
  exporters' code is unchanged, since the base module's `Config` and `Exporters`, which they use,
  did not change in v0.4.0.

## [v0.3.0] - 2026-11-11

### Changed

- **Breaking:** The `go-observability` requirement is v0.3.0, which requires `go-core` v0.11.0
  and pulls go-core's v0.8.0 and v0.11.0 breaks into an importer's build: `graph.Scope.After` is
  removed in favor of `graph.After` on the `Define` line, `graph.Graph.Observe` takes a
  `graph.Ref`, a name containing a comma or a newline panics, `graph.Dependency` gains a `Ref`
  field, a `Monitored` value's runtime failure is labelled with its node's name, and
  `lifecycle.Coordinator.Checks` holds a `Check` only for a `Starter` that is also a
  `ReadinessChecker`. The exporters' code is unchanged: `Config` and `Exporters`, the base
  module's types they use, are unchanged in v0.3.0.
- The `google.golang.org/grpc` requirement is v1.85.0 (was v1.84.0), for its bug fixes.

## [v0.2.0] - 2026-10-09

### Changed

- **Breaking:** The `go-observability` requirement is v0.2.0, which requires `go-core` v0.7.0.
  An importer still on go-core's `lifecycle.Service`, `Add`, or stages breaks, since the
  requirement pulls go-core v0.6.0's lifecycle into its build. The exporters' code is unchanged:
  `Config` and `Exporters`, the base module's types they use, are unchanged in v0.2.0.

## [v0.1.0] - 2026-09-15

The first release of the gRPC OTLP exporters, against `github.com/standards-lab/go-observability v0.1.0`.

### Added

- `NewTraceExporter` and `NewMetricExporter` — the span and metric exporters over `Config`'s
  `Endpoint` and `Headers`, connecting in plain text unconditionally: `Config` has no TLS field
  yet, and the only OTLP target in the workspace so far is the local compose collector stack,
  which nothing secures with TLS. TLS support is deferred until a deployed or managed backend
  needs it. Both dial lazily, so construction never blocks on a collector that isn't up yet.

[Unreleased]: https://github.com/standards-lab/go-observability/compare/otlp/v0.4.0...HEAD
[v0.4.0]: https://github.com/standards-lab/go-observability/compare/otlp/v0.3.0...otlp/v0.4.0
[v0.3.0]: https://github.com/standards-lab/go-observability/compare/otlp/v0.2.0...otlp/v0.3.0
[v0.2.0]: https://github.com/standards-lab/go-observability/compare/otlp/v0.1.0...otlp/v0.2.0
[v0.1.0]: https://github.com/standards-lab/go-observability/releases/tag/otlp/v0.1.0
