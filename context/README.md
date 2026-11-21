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

- **observability** (built): configuration, the telemetry lifecycle, the trace-correlating log
  handler, the HTTP middleware, and the request-ID source ([doc.go](../doc.go)).
- **otlp** (built): the gRPC trace and metric exporters ([otlp/doc.go](../otlp/doc.go)).
- **The logs pipeline** (planned): `otel/log`, `otel/sdk/log`, the `slog` bridge, and an OTLP log
  exporter in `otlp` ([logs-pipeline.md](logs-pipeline.md)).
