# Changelog

All notable changes to `github.com/standards-lab/go-observability` are documented here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the module adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). This changelog covers the base
module only; the `otlp` sub-module keeps its own.

## [Unreleased]

## [v0.4.0] - 2026-11-11

### Changed

- **Breaking:** The `go-core` requirement is v0.12.0, which pulls go-core's v0.12.0 breaks into
  an importer's build: `reactor.New` takes a trailing `reactor.Options` in place of variadic
  `reactor.Option` values, `reactor.CancelWithin` returns a `time.Duration` for
  `Options.CancelAfter`, `event.NewRecorder` takes a trailing `event.Options` in place of
  variadic `event.Option` values, and the `reactor.CancelAfter`, `event.Clock`, and `event.IDs`
  option functions are removed.
- The documentation names go-web-sdk v0.17.0's call form for `RequestIDSource`:
  `middleware.RequestID(middleware.RequestIDOptions{Source: observability.RequestIDSource})`, in
  place of the removed `middleware.WithIDSource` option. `RequestIDSource` is unchanged.

## [v0.3.0] - 2026-11-11

### Changed

- **Breaking:** The `go-core` requirement is v0.11.0, which pulls go-core's v0.8.0 and v0.11.0
  breaks into an importer's build: `graph.Scope.After` is removed, so an order-only edge is
  declared with `graph.After` on the node's `Define` line; `graph.Graph.Observe` takes
  `func(graph.Ref)` in place of `func(string)`; a node or group name containing a comma or a
  newline panics; `graph.Dependency` gains a `Ref` field, so an unkeyed composite literal no
  longer compiles; a `Monitored` value's runtime failure ends the run as
  `run: <node>: <error>`; and `lifecycle.Coordinator.Checks` holds a `Check` only for a value
  that is both a `Starter` and a `ReadinessChecker`.
- The documentation describes a node ordering itself after `Telemetry` with `graph.After` on its
  `Define` line, in place of the removed `Scope.After`.

## [v0.2.0] - 2026-10-09

### Changed

- **Breaking:** The `go-core` requirement is v0.7.0, whose logger writes each record's time in
  `time.UTC`, whatever `time.Local` is. An importer still on go-core's `lifecycle.Service`,
  `Add`, or stages breaks, since the requirement pulls go-core v0.6.0's lifecycle into its
  build.
- `*Telemetry` joins a graph-backed `lifecycle.Coordinator` as a node's value: the Coordinator
  infers from its methods that it is a `Starter` and a `Stopper`, so it needs no adapter. A node
  it instruments uses its node, or orders itself after it with `Scope.After`, so it starts first
  and shuts down last.
- `Telemetry.Shutdown` releases the exporters itself when `Start` failed or never ran, since the
  Coordinator shuts down a participant whose `Start` failed. It no longer dereferences the
  providers `Start` never built.

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

[Unreleased]: https://github.com/standards-lab/go-observability/compare/v0.4.0...HEAD
[v0.4.0]: https://github.com/standards-lab/go-observability/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/standards-lab/go-observability/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/standards-lab/go-observability/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/standards-lab/go-observability/releases/tag/v0.1.0
