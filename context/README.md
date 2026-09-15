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
- **The correlating `slog.Handler`** — wraps `go-core/logging.New`'s handler, appending `trace_id`
  and `span_id` from a request's span context when valid. Not yet built.
- **The HTTP server middleware** — a constructor over `Config` wrapping `otelhttp.NewMiddleware`,
  structurally a `func(http.Handler) http.Handler` with no `go-web-sdk` import. Not yet built.
- **The request-ID source function** — `func(*http.Request) string`, returning the request's
  current trace id, the shape `go-web-sdk`'s `web.RequestID(WithIDSource(...))` takes. Not yet
  built.
- **`otlp`** — the gRPC trace and metric exporter constructors. Not yet built.
