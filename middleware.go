package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewMiddleware returns the HTTP server middleware: otelhttp.NewMiddleware
// configured from cfg. The handler it wraps runs inside a server span that
// carries the HTTP semantic-convention attributes, continues the trace an
// incoming traceparent header names, and records the request-duration
// histogram; the wrapped handler sees the span in the request context, so a
// logger built on [NewTraceHandler] correlates its records with it.
//
// The operation name otelhttp takes is cfg.ResourceAttributes["service.name"]
// when set and empty otherwise. Under otelhttp's default span-name formatter
// the operation does not appear in the span name: spans are named by method
// and route ("GET /users/{id}" when the mux sets a request pattern, "GET"
// when it does not), and the service name reaches the backend through the
// resource that [Telemetry] builds. The operation only surfaces through a
// custom span-name formatter, which this constructor does not install, so an
// empty operation loses nothing and no substitute name is invented.
//
// The return type is the plain func(http.Handler) http.Handler rather than
// go-web-sdk's Middleware, which has that as its underlying type: the value
// is assignable to it without this module importing the SDK.
//
// No tracer or meter provider is passed, so otelhttp resolves both from the
// otel globals. The composition root usually builds this middleware before
// [Telemetry.Start] installs the real providers, and that order is safe by
// design: the pre-installation globals are delegating proxies that route to
// the real providers retroactively the moment Start installs them, so a
// tracer, meter, or propagator obtained early starts recording then. There
// is no ordering problem here to fix.
func NewMiddleware(cfg Config) func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware(cfg.ResourceAttributes["service.name"])
}
