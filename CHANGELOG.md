# Changelog

All notable changes to `github.com/standards-lab/go-observability` are documented here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the module adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). This changelog covers the base
module only; the `otlp` sub-module keeps its own.

## [Unreleased]

## [v0.1.0] - 2026-09-15

The first release of the observability infrastructure library: the OpenTelemetry configuration
and process lifecycle, the trace-correlating log handler, the HTTP server middleware, and the
request-ID source function. The module depends on the standard library,
`github.com/standards-lab/go-core v0.4.1`, and the stable v1 OpenTelemetry API and SDK, plus
`otelhttp` admitted as a stated v0 exception. The `otlp` sub-module pins the gRPC exporters and
is released on its own tags.

### Added

- `Config` — the collector endpoint, headers, trace sampling ratio, and resource attributes, on
  go-core's Merge-and-Finalize contract. No protocol field: the `otlp` sub-module ships gRPC
  exporters alone for this release, so a field with one valid value would be dead configuration.
- `Telemetry` — the `lifecycle.Service` at stage 0: builds the resource from the configured
  attributes merged over the SDK's defaults, constructs the `TracerProvider` and
  `MeterProvider`, installs them and the W3C `TraceContext` propagator as process globals, and
  flushes on shutdown. No readiness check: observability never gates readiness.
- `NewTraceHandler` — a `slog.Handler` decorator appending `trace_id` and `span_id` to a record
  logged under a context carrying a valid span context; a record with no active span passes
  through unchanged.
- `NewMiddleware` — the HTTP server middleware, wrapping `otelhttp.NewMiddleware` as a plain
  `func(http.Handler) http.Handler`, structurally `go-web-sdk`'s `Middleware` type without
  importing that module.
- `RequestIDSource` — the `func(*http.Request) string` `go-web-sdk`'s
  `middleware.WithIDSource` takes, returning the request's current trace id or `""` to decline.

[Unreleased]: https://github.com/standards-lab/go-observability/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/standards-lab/go-observability/releases/tag/v0.1.0
