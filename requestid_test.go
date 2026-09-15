package observability_test

import (
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace"

	"github.com/standards-lab/go-observability"
)

func TestRequestIDSource_ReturnsTraceIDUnderValidSpan(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil).WithContext(spanContext())

	if got, want := observability.RequestIDSource(req), testTraceID.String(); got != want {
		t.Errorf("RequestIDSource = %q, want %q", got, want)
	}
}

func TestRequestIDSource_ReturnsEmptyWithoutSpan(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	if got := observability.RequestIDSource(req); got != "" {
		t.Errorf("RequestIDSource = %q, want empty", got)
	}
}

func TestRequestIDSource_ReturnsEmptyUnderInvalidSpanContext(t *testing.T) {
	// A span context can be present but invalid, e.g. a zero TraceID from a
	// malformed traceparent header the propagator declined to extract.
	sc := trace.NewSpanContext(trace.SpanContextConfig{})
	req := httptest.NewRequest("GET", "/", nil).WithContext(
		trace.ContextWithSpanContext(t.Context(), sc),
	)

	if got := observability.RequestIDSource(req); got != "" {
		t.Errorf("RequestIDSource = %q, want empty", got)
	}
}
