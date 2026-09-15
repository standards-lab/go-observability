# Changelog

All notable changes to the OTLP exporters (`github.com/standards-lab/go-observability/otlp`) are
documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). This changelog
covers this sub-module only; the base module keeps its own.

## [Unreleased]

## [v0.1.0] - 2026-09-15

The first release of the gRPC OTLP exporters, against `github.com/standards-lab/go-observability v0.1.0`.

### Added

- `NewTraceExporter` and `NewMetricExporter` — the span and metric exporters over `Config`'s
  `Endpoint` and `Headers`, connecting in plain text unconditionally: `Config` has no TLS field
  yet, and the only OTLP target in the workspace so far is the local compose collector stack,
  which nothing secures with TLS. TLS support is deferred until a deployed or managed backend
  needs it. Both dial lazily, so construction never blocks on a collector that isn't up yet.

[Unreleased]: https://github.com/standards-lab/go-observability/compare/otlp/v0.1.0...HEAD
[v0.1.0]: https://github.com/standards-lab/go-observability/releases/tag/otlp/v0.1.0
