# go-observability context

go-observability is the observability infrastructure library of Go Elemental: the OpenTelemetry
configuration and process lifecycle, the trace-correlating log handler, the HTTP server
middleware, and the request-ID source function, with the OTLP exporters isolated in the `otlp`
sub-module.

The README, including its Design section, and each package's `doc.go` document this repository.
The [Go Elemental](https://github.com/standards-lab/architecture/blob/main/standards/go-elemental/README.md)
standard states the principles it follows. This context records only working knowledge the code
and the README do not express.

## Capability map

Every package is built, and its code and `doc.go` are authoritative. Detail for what is unbuilt
is added when it is about to be built.

- **`Config`** has no protocol field, because `otlp` ships gRPC exporters alone and a field with
  one valid value would be dead configuration. The field arrives with HTTP/protobuf exporters.
- **`otlp`** connects in plain text. TLS arrives, with its `Config` field, when a deployed or
  managed backend needs it.
- **`NewMiddleware`** passes `service.name` as `otelhttp`'s operation name. The default span
  formatter ignores it; a custom formatter would use it.
- **The logs pipeline** (`otel/log`, `otel/sdk/log`, the `slog` bridge, and an OTLP log exporter)
  is planned for when `otel/log` reaches a stable v1, with the log exporter joining the others in
  `otlp`.
