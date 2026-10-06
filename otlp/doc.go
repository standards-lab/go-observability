// Package otlp builds the gRPC OTLP exporters that the base observability
// module's [observability.Telemetry] wires into its providers. It is a
// separate module because every OTLP exporter pulls go.opentelemetry.io/proto/otlp
// and, behind it, grpc, protobuf, grpc-gateway, and genproto: a dependency
// footprint the base module never imports, so an application that wants only
// the tracer and meter interfaces never compiles it. The repository README's
// Design section records the split and its reasoning.
//
// The package exports:
//
//   - [NewTraceExporter], which builds the span exporter for
//     [observability.Exporters]
//   - [NewMetricExporter], which builds the metric exporter the composition
//     root wraps in a reader for [observability.Exporters]
//   - [ErrEndpointRequired], the error both constructors return for a Config
//     with no Endpoint
//
// # Transport security
//
// Both exporters connect in plain text (the exporters' WithInsecure option),
// unconditionally. [observability.Config] has no TLS field: the OTLP target
// the library is built against is the local compose collector stack, which
// nothing secures with TLS. TLS support is planned for when a deployed or
// managed backend needs it: Config gains the field, and these constructors
// honor it.
package otlp
