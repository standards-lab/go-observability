package observability_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/standards-lab/go-observability"
)

// The tests that start Telemetry install process globals, so none of them
// run in parallel; see telemetry_test.go.

// serveThrough sends a GET for path through mw around a mux that serves the
// "GET /users/{id}" pattern with a 202 and a fixed body, so the span name
// carries the route, and returns the recorded response.
func serveThrough(t *testing.T, mw func(http.Handler) http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "user "+r.PathValue("id"))
	})
	rec := httptest.NewRecorder()
	mw(mux).ServeHTTP(rec, req)
	return rec
}

func wantResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Body.String(); got != "user 42" {
		t.Errorf("body = %q, want %q", got, "user 42")
	}
}

func TestNewMiddleware_WrapsHandler(t *testing.T) {
	cfg := finalizedConfig(t, 1)
	mw := observability.NewMiddleware(cfg)

	rec := serveThrough(t, mw, httptest.NewRequest(http.MethodGet, "/users/42", nil))
	wantResponse(t, rec)
}

func TestNewMiddleware_InstrumentsRequests(t *testing.T) {
	tel, exporter, reader := startTelemetry(t, 1)
	mw := observability.NewMiddleware(finalizedConfig(t, 1))

	// The incoming traceparent names the trace the server span continues.
	const traceID = "0af7651916cd43dd8448eb211c80319c"
	const parentSpanID = "b7ad6b7169203331"
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req.Header.Set("traceparent", "00-"+traceID+"-"+parentSpanID+"-01")

	rec := serveThrough(t, mw, req)
	wantResponse(t, rec)

	// Metrics collect through the reader, which Shutdown releases, so they
	// are read first; spans reach the exporter only on the flush Shutdown
	// forces.
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	if !hasMetric(rm, "http.server.request.duration") {
		t.Errorf("collected %s, want http.server.request.duration among them", metricNames(rm))
	}

	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans, want 1", len(spans))
	}
	span := spans[0]
	if span.Name != "GET /users/{id}" {
		t.Errorf("span name = %q, want %q", span.Name, "GET /users/{id}")
	}
	if span.SpanKind != trace.SpanKindServer {
		t.Errorf("span kind = %v, want server", span.SpanKind)
	}
	if got := span.SpanContext.TraceID().String(); got != traceID {
		t.Errorf("trace id = %s, want %s (the incoming traceparent)", got, traceID)
	}
	if got := span.Parent.SpanID().String(); got != parentSpanID {
		t.Errorf("parent span id = %s, want %s (the incoming traceparent)", got, parentSpanID)
	}
	wantAttribute(t, span.Resource, "service.name", "app")
}

func TestNewMiddleware_ConstructedBeforeStartStillInstruments(t *testing.T) {
	// The premise: no real provider is installed when the middleware is
	// built, only the otel package's delegating placeholder.
	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		t.Fatal("a real TracerProvider is installed before construction; the test would prove nothing")
	}
	mw := observability.NewMiddleware(finalizedConfig(t, 1))

	// startTelemetry restores the globals it installs at cleanup.
	tel, exporter, _ := startTelemetry(t, 1)

	rec := serveThrough(t, mw, httptest.NewRequest(http.MethodGet, "/users/42", nil))
	wantResponse(t, rec)

	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans from a middleware built before Start, want 1", len(spans))
	}
	if spans[0].Name != "GET /users/{id}" {
		t.Errorf("span name = %q, want %q", spans[0].Name, "GET /users/{id}")
	}
}

func hasMetric(rm metricdata.ResourceMetrics, name string) bool {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return true
			}
		}
	}
	return false
}

func metricNames(rm metricdata.ResourceMetrics) []string {
	var names []string
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			names = append(names, m.Name)
		}
	}
	return names
}
