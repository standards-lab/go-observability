// Package observability is the OpenTelemetry infrastructure library: the
// configuration block that names the collector, the process lifecycle
// service that builds the resource and the providers, the trace-correlating
// log handler, the HTTP server middleware, and the request-ID source
// function (the last three land in later stages). The package depends on the
// standard library, go-core, and the stable v1 OpenTelemetry API and SDK; the
// gRPC OTLP exporters and their dependency weight live in the otlp
// sub-module, which this package never imports.
//
// # Configuration
//
// [Config] implements the config package's Merge and Finalize contract, so it
// loads as part of an application's configuration. Endpoint is the one
// required field and has no default. SampleRatio is a pointer: nil is unset
// and takes the default of 1, while an explicit 0 survives the load and means
// never sample. Headers and ResourceAttributes merge key-wise across layers.
// Finalize composes the standard override names from the prefix it receives
// (via [NewEnv], recorded on [Env] for introspection); an empty prefix
// disables the overrides.
//
// # Telemetry
//
// [Telemetry] is the process's OpenTelemetry lifecycle service, built from a
// finalized [Config] and an [Exporters] pair the composition root supplies —
// this package never constructs an exporter itself, since the gRPC OTLP
// exporters belong to the otlp sub-module. Start builds the resource from
// the configured attributes merged over the SDK's own defaults, constructs
// the TracerProvider and MeterProvider, and installs both and the W3C
// TraceContext propagator as process globals; Shutdown flushes and releases
// them. It registers at lifecycle stage 0 with no readiness check:
// observability never gates readiness.
package observability
