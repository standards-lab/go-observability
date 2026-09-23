// Package observability is the OpenTelemetry infrastructure library: the
// configuration block that names the collector, the process lifecycle
// service that builds the resource and the providers, the trace-correlating
// log handler, the HTTP server middleware, and the request-ID source
// function. The package depends on the standard library, go-core, and the
// stable v1 OpenTelemetry API and SDK; the gRPC OTLP exporters and their
// dependency weight live in the otlp sub-module, which this package never
// imports.
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
// them. The composition root registers Start as a lifecycle startup hook
// and Shutdown as a shutdown hook, so telemetry is installed before the
// first numbered stage starts and flushed after the last drains, without
// holding a stage of its own. It has no readiness check: observability
// never gates readiness.
//
// # Log correlation
//
// [NewTraceHandler] wraps a slog.Handler — the intended one is the handler
// behind go-core's logging.New, via the logger's Handler method — and
// appends trace_id and span_id to a record logged under a context carrying
// a valid span context, so a log line and the span it happened inside share
// a key every other signal joins on. A record logged with no active span
// passes through unchanged.
//
// # HTTP server middleware
//
// [NewMiddleware] wraps otelhttp.NewMiddleware as a plain
// func(http.Handler) http.Handler, structurally go-web-sdk's Middleware type
// without importing that module. The wrapped handler runs inside a server
// span carrying the HTTP semantic-convention attributes, continues a trace
// an incoming traceparent header names, and records the request-duration
// histogram. It resolves its tracer and meter provider from the otel
// globals rather than taking them as parameters, which is safe to construct
// before [Telemetry.Start] installs the real ones: the pre-installation
// globals are delegating proxies that route to the real providers the
// moment Start installs them.
//
// # Request-ID source
//
// [RequestIDSource] is the func(*http.Request) string go-web-sdk's
// middleware.WithIDSource takes: it returns the request's current trace id
// when one is on the request's context, or "" to decline and let
// middleware.RequestID fall through to its next source. It reads a span
// already on the request; reaching one depends on the composition root
// wiring [NewMiddleware] ahead of go-web-sdk's RequestID middleware in the
// chain. Neither module imports the other.
package observability
