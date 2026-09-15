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

## Development

The repository uses a Go workspace and [mise](https://mise.jdx.dev):

```
mise run build   # build every module standalone, with the workspace off
mise run test    # test every module
```

## License

[Apache License 2.0](LICENSE).
