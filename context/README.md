# go-observability context

go-observability is the observability infrastructure library of Go Elemental: the OpenTelemetry
configuration and process lifecycle, the trace-correlating log handler, the HTTP server
middleware, and the request-ID source function, with the OTLP exporters isolated in the `otlp`
sub-module.

The README and each package's `doc.go` document this repository. The
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md)
standard states the principles it follows. The settled design this repository builds to —
the standard-versus-native split, the module shape, and the correlation model — is
`standards-lab/context/design/observability-strategy.md`, the workspace coordinator's record;
this context records only working knowledge the code and the README do not express.

## Capability map

The built packages are authoritative through their code and `doc.go`. Detail for what is unbuilt
is added when it is about to be built.

- **`Config`** — the collector endpoint, headers, trace sampling ratio, and resource attributes,
  on `go-core`'s Merge-and-Finalize contract. Built. No protocol field: the `otlp` sub-module
  ships gRPC exporters alone for the first release, so a field with one valid value would be
  dead configuration; the field returns if HTTP/protobuf is ever added — a narrowing of
  `observability-strategy.md`'s "endpoint and protocol," recorded here since the design note
  itself is promoted, not restated, per `context-architecture.md`.
- **`Telemetry`** — the `lifecycle.Service` at stage 0: builds the `resource.Resource`, constructs
  the `TracerProvider` and `MeterProvider`, installs them and the W3C `TraceContext` propagator as
  process globals, flushes on shutdown. Built. Takes an `Exporters{Trace, Metric}` pair from the
  composition root rather than constructing either itself — the trace exporter and the metric
  reader are already-built SDK values, so this package stays free of the `otlp` sub-module's
  weight. No readiness check, per the design's posture rule.
- **The correlating `slog.Handler`** (`NewTraceHandler`) — wraps any `slog.Handler`, appending
  `trace_id` and `span_id` from a request's span context when valid; a record with no active
  span passes through unchanged. Built. Under `WithGroup`, the two attributes nest inside the
  opened group like any attribute added afterward — standard `slog` semantics, and nothing in the
  workspace opens a group on a correlated logger yet.
- **The HTTP server middleware** (`NewMiddleware`) — a constructor over `Config` wrapping
  `otelhttp.NewMiddleware`, structurally a `func(http.Handler) http.Handler` with no
  `go-web-sdk` import. Built. Resolves its tracer/meter provider from the `otel` globals rather
  than taking them explicitly, which is safe even when the middleware is constructed before
  `Telemetry.Start` runs — the pre-installation globals are delegating proxies. The operation
  name passed to `otelhttp.NewMiddleware` (`ResourceAttributes["service.name"]`) doesn't
  currently affect span naming under the default formatter; kept anyway since it costs nothing
  and is the right input if a custom formatter is ever added.
- **`RequestIDSource`** — `func(*http.Request) string`, returning the request's current trace id
  or `""` to decline, the shape `go-web-sdk`'s `middleware.RequestID(WithIDSource(...))` takes.
  Built. It only reads a span already on the request; the composition root reaching one at all
  depends on wiring `NewMiddleware` ahead of `RequestID` in the chain — a later task's concern
  (`v1.observability.tasks.instrumentation`), not this one's.
- **`otlp`** — the gRPC trace and metric exporter constructors (`NewTraceExporter`, `NewMetricExporter`). Built. Both connect in plain text unconditionally: `Config` has no TLS field yet, and the only OTLP target in the workspace so far is the local compose collector stack, which nothing secures with TLS. TLS support is deferred until a deployed or managed backend needs it.
