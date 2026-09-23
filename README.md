# go-observability

go-observability is the observability infrastructure library of Go Elemental, the Standards Lab
organization's Go implementation of the Elemental Architecture. It holds the OpenTelemetry
configuration and process lifecycle, the trace-correlating log handler, the HTTP server
middleware, and the request-ID source function every service instruments against; the OTLP
exporters are isolated in the `otlp` sub-module.

`github.com/standards-lab/go-observability` is the base module. `github.com/standards-lab/go-observability/otlp`
is a nested sub-module that pins the exporters and is released on its own tags.

## Standard

`go-observability` is an infrastructure library of
[Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md), the
minimal-dependency Go standard. This README and each package's `doc.go` document the repository;
the standard's principles it enhances are stated below. Its repository-level principles:

- The base module depends on the standard library, `go-core`, and the stable v1 OpenTelemetry API
  and SDK (`otel`, `otel/trace`, `otel/metric`, `otel/sdk`, `otel/sdk/metric`), plus `otelhttp` for
  the HTTP server middleware — admitted as a stated v0 exception: the contrib repository has never
  released its instrumentation modules past v0, and `otelhttp` passes every other standard-library
  marker: stdlib types at its boundary (`func(http.Handler) http.Handler`), one non-OpenTelemetry
  transitive dependency (`github.com/felixge/httpsnoop`), and maintenance by the project that
  defines the ecosystem.
- The `otlp` sub-module isolates the exporters and the gRPC/protobuf weight every one of them
  pulls in (`go.opentelemetry.io/proto/otlp`, `google.golang.org/grpc`, `google.golang.org/protobuf`,
  and their kin); a consumer that needs only the tracer and meter interfaces never compiles it.
- Observability never gates readiness: `Telemetry` registers with the process lifecycle through a
  `Start` and a `Shutdown`, and no health check.

## Design

**Standard tier only, no provider.** The standard is OpenTelemetry: its API and SDK, its semantic
conventions, W3C Trace Context propagation, and OTLP on the wire. A backend sits behind the OTLP
exporter, so nothing about the backend enters the dependency graph, and the library declares no
provider. The service reaches every backend through an OpenTelemetry collector, locally in front of
Loki, Tempo, and Mimir and in deployment in front of a managed OTLP endpoint. Naming a backend in
the collector's exporter configuration is operational configuration, like a connection string, so
this layer has no native tier.

**The `otlp` sub-module isolates dependency weight; it does not make the exporter swappable.** Every
OTLP exporter pulls gRPC, protobuf, grpc-gateway, and genproto, which a consumer of the interfaces
alone should never compile. The split follows the export boundary, where that weight sits, rather
than the signal: traces and metrics share one exporter graph and one version line. If every consumer
of the base module ends up importing `otlp` too, merging the two back is cheap.

**The request id is the trace id.** OpenTelemetry defines no generic request-id attribute and uses
the trace context instead. `NewMiddleware` starts the server span, and `RequestIDSource` supplies
the span's trace id to go-web-sdk's `middleware.RequestID`, which generates its own id for a service
that composes no tracing. Neither module imports the other. The log handler stamps `trace_id` and
`span_id` on each record, so log-to-trace links, exemplars, and trace-to-log queries in the backends
need nothing product-specific from the service.

**Logs go to stdout, not OTLP.** A line written to stdout persists without a live connection to the
collector, and that connection is exactly what is missing when a service is failing. The collector's
`filelog` receiver ingests the line. The OpenTelemetry logs pipeline is also still pre-1.0. The
choice reverses when `otel/log` and its `slog` bridge reach a stable v1.

**Telemetry starts first and stops last.** The composition root registers `Telemetry`'s `Start` and
`Shutdown` as the lifecycle's startup and shutdown hooks rather than in a numbered stage, so it is
installed before anything it would instrument starts and flushed after everything drains. It never
gates readiness: a service that cannot reach its collector still serves traffic.

**Moving to a managed backend** changes only the collector's exporter section. The move needs a
review of metric temporality (cumulative against delta), attribute cardinality limits, and, once
logs move to OTLP, the backend's logs ingestion.

**Rejected alternatives:**

- A backend-named sub-module (`loki`, `tempo`, `grafana`): there is no Go-side interface for it
  to adapt, because the backend never enters the dependency graph.
- A per-signal split (`traces`, `metrics`, `logs`): it divides what moves together and leaves the
  weight undivided.
- Hand-rolled HTTP instrumentation: span naming, low-cardinality route attribution, and the
  duration histogram's buckets answer to a specification, where a wrong answer looks right.
- A stdlib-only event abstraction in front of OpenTelemetry: this library exists to take the
  OpenTelemetry dependency, so the abstraction would have nothing left to do.

## Packages

- `observability` (the base module's root package) — `Config`, on go-core's Merge-and-Finalize
  contract; `Telemetry`, the process lifecycle service that builds the resource and the
  providers; `NewTraceHandler`, the trace-correlating `slog.Handler`; `NewMiddleware`, the HTTP
  server instrumentation; and `RequestIDSource`, the trace-id source `go-web-sdk`'s
  `middleware.WithIDSource` takes.
- `otlp` — the gRPC trace and metric exporter constructors, connecting in plain text (TLS is
  deferred until a deployed or managed backend needs it — `otlp`'s own `doc.go` states why).

## Development

The repository uses a Go workspace and [mise](https://mise.jdx.dev):

```
mise run build   # build every module standalone, with the workspace off
mise run test    # test every module
```

## License

[Apache License 2.0](LICENSE).
