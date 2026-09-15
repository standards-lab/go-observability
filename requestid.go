package observability

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// RequestIDSource is the request-ID source function go-web-sdk's
// middleware.WithIDSource takes: it returns the request's current trace id
// (hex-encoded, the same form NewTraceHandler appends as trace_id) when r's
// context carries a valid OpenTelemetry span context, or "" otherwise. An
// empty return declines the request, so middleware.RequestID falls through
// to its next source rather than getting a trace id.
//
// This function only reads a span already on the request; it never starts
// one. It reaches a valid span context only when the composition root wires
// [NewMiddleware] ahead of go-web-sdk's RequestID middleware in the chain,
// since that ordering is what puts a span on the request's context in the
// first place. Neither module imports the other — this is the seam
// go-web-sdk's WithIDSource exists for.
func RequestIDSource(r *http.Request) string {
	sc := trace.SpanContextFromContext(r.Context())
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}
