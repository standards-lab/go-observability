// Package otlp builds the gRPC OTLP exporters that the base observability
// module's [observability.Telemetry] wires into its providers. It is a
// separate module because every OTLP exporter pulls go.opentelemetry.io/proto/otlp
// and, behind it, grpc, protobuf, grpc-gateway, and genproto: a dependency
// footprint the base module never imports, so an application that wants only
// the tracer and meter interfaces never compiles it. The repository README's
// Design section records the split and its reasoning.
//
// [NewTraceExporter] and [NewMetricExporter] each take the base module's
// [observability.Config] and build one exporter over its Endpoint and
// Headers. The explicit endpoint takes precedence over the SDK's
// OTEL_EXPORTER_OTLP_* environment variables, so Config alone decides where
// the exporters send. Both dial lazily: construction succeeds with no
// collector listening, and an unreachable collector surfaces as export
// errors rather than a construction failure, which keeps a collector that is
// down from ever blocking a service's startup. The metric constructor returns
// the raw exporter rather than a reader, because [observability.Exporters]
// takes a reader and the composition root decides how the exporter becomes
// one, in production by wrapping it in sdkmetric.NewPeriodicReader.
//
// # Transport security
//
// Both exporters connect in plain text (the exporters' WithInsecure option),
// unconditionally. [observability.Config] has no TLS field, and the only
// OTLP target in the workspace so far is the local compose collector stack,
// which nothing secures with TLS, so plain-text gRPC is the correct default
// for that target and the only one Config currently expresses. TLS support
// is deferred until a deployed or managed backend needs it, at which point
// Config gains the field and these constructors honor it.
package otlp
