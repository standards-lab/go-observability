// Package observability is the OpenTelemetry infrastructure library: the
// configuration block that names the collector, the headers the exporter
// sends, the trace sampling ratio, and the resource attributes. The package
// depends on the standard library and go-core's config package alone; the
// gRPC OTLP exporters and their dependency weight live in the otlp
// sub-module, which the base module never imports.
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
package observability
