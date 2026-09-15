package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler is the trace-correlating decorator [NewTraceHandler] returns.
// It holds only the handler it wraps; every record's trace attributes come
// from the context the record is logged under, never from stored state.
type traceHandler struct {
	next slog.Handler
}

// NewTraceHandler wraps next in a handler that correlates log records with
// the request's span. When a record is logged under a context carrying a
// valid OpenTelemetry span context, the handler appends the trace_id and
// span_id attributes (the W3C Trace Context identifiers, hex-encoded) to the
// record before handing it to next. Under a context with no valid span, the
// usual case outside a request, the record passes through unchanged. Enabled,
// WithAttrs, and WithGroup delegate to next unchanged, so a logger built on
// the wrapper keeps correlating after With or WithGroup. The intended next is
// the handler behind go-core's logging.New, obtained from the logger's Handler
// method, but any slog.Handler works.
func NewTraceHandler(next slog.Handler) slog.Handler {
	return &traceHandler{next: next}
}

// Enabled reports whether the wrapped handler handles records at level.
func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle appends trace_id and span_id to record when ctx carries a valid
// span context, then hands the record to the wrapped handler.
func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.next.Handle(ctx, record)
}

// WithAttrs returns a trace-correlating handler around the wrapped handler
// with attrs bound.
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{next: h.next.WithAttrs(attrs)}
}

// WithGroup returns a trace-correlating handler around the wrapped handler
// with the group name opened. trace_id and span_id then nest inside that
// group with every other attribute added afterward, standard slog.Handler
// semantics; nothing in this codebase opens a group on the correlated logger
// today, so the nesting has no observed consequence yet.
func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{next: h.next.WithGroup(name)}
}
