package observability_test

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/standards-lab/go-observability"
)

// The tests that call Start install process globals in the otel package, so
// none of them run in parallel and each restores the globals it replaced.

func finalizedConfig(t *testing.T, ratio float64) observability.Config {
	t.Helper()
	cfg := observability.Config{
		Endpoint:    "localhost:4317",
		SampleRatio: &ratio,
		ResourceAttributes: map[string]string{
			"service.name":    "app",
			"service.version": "1.2.3",
		},
	}
	if err := cfg.Finalize(""); err != nil {
		t.Fatalf("finalize config: %v", err)
	}
	return cfg
}

// keepExporter is the SDK's in-memory exporter with Shutdown made a no-op.
// The batch processor calls the exporter's Shutdown right after its final
// flush, and InMemoryExporter.Shutdown clears the spans it holds, which
// would hide the very flush the tests observe.
type keepExporter struct {
	*tracetest.InMemoryExporter
}

func (keepExporter) Shutdown(context.Context) error { return nil }

// testExporters returns the SDK's in-memory doubles: an exporter that keeps
// the spans it receives and a reader that collects on demand.
func testExporters() (observability.Exporters, keepExporter, *sdkmetric.ManualReader) {
	exporter := keepExporter{tracetest.NewInMemoryExporter()}
	reader := sdkmetric.NewManualReader()
	return observability.Exporters{Trace: exporter, Metric: reader}, exporter, reader
}

// restoreGlobals captures the otel globals before a test installs its own
// and puts them back when the test ends, so no installation leaks into the
// next test's assertions.
func restoreGlobals(t *testing.T) {
	t.Helper()
	tp := otel.GetTracerProvider()
	mp := otel.GetMeterProvider()
	prop := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(tp)
		otel.SetMeterProvider(mp)
		otel.SetTextMapPropagator(prop)
	})
}

// startTelemetry starts a Telemetry over the in-memory doubles and shuts it
// down at cleanup; a test that shuts down itself just sees a second, clean
// no-op.
func startTelemetry(t *testing.T, ratio float64) (*observability.Telemetry, keepExporter, *sdkmetric.ManualReader) {
	t.Helper()
	restoreGlobals(t)
	exporters, exporter, reader := testExporters()
	tel := observability.New(finalizedConfig(t, ratio), exporters)
	if err := tel.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = tel.Shutdown(context.Background()) })
	return tel, exporter, reader
}

func wantPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic, want one containing %q", want)
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, want) {
			t.Errorf("panic = %v, want it to contain %q", r, want)
		}
	}()
	fn()
}

func wantAttribute(t *testing.T, res *resource.Resource, key, want string) {
	t.Helper()
	if res == nil {
		t.Fatal("resource is nil")
	}
	got, ok := res.Set().Value(attribute.Key(key))
	if !ok {
		t.Errorf("resource lacks %s", key)
		return
	}
	if got.AsString() != want {
		t.Errorf("resource %s = %q, want %q", key, got.AsString(), want)
	}
}

func TestNew_PanicsOnUnfinalizedConfig(t *testing.T) {
	exporters, _, _ := testExporters()
	wantPanic(t, "Config not finalized", func() {
		observability.New(observability.Config{Endpoint: "localhost:4317"}, exporters)
	})
}

func TestNew_PanicsOnNilExporters(t *testing.T) {
	exporters, _, _ := testExporters()

	wantPanic(t, "nil trace exporter", func() {
		observability.New(finalizedConfig(t, 1), observability.Exporters{Metric: exporters.Metric})
	})
	wantPanic(t, "nil metric reader", func() {
		observability.New(finalizedConfig(t, 1), observability.Exporters{Trace: exporters.Trace})
	})
}

func TestTelemetry_StartInstallsGlobals(t *testing.T) {
	startTelemetry(t, 1)

	if _, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); !ok {
		t.Errorf("global TracerProvider = %T, want *sdktrace.TracerProvider", otel.GetTracerProvider())
	}
	if _, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); !ok {
		t.Errorf("global MeterProvider = %T, want *sdkmetric.MeterProvider", otel.GetMeterProvider())
	}
	if _, ok := otel.GetTextMapPropagator().(propagation.TraceContext); !ok {
		t.Errorf("global propagator = %T, want propagation.TraceContext", otel.GetTextMapPropagator())
	}
}

func TestTelemetry_SpansCarryResourceAndFlushOnShutdown(t *testing.T) {
	tel, exporter, _ := startTelemetry(t, 1)

	// A span through the global provider reaches the exporter only once the
	// batcher flushes, which Shutdown forces.
	_, span := otel.Tracer("test").Start(context.Background(), "op")
	if !span.IsRecording() {
		t.Fatal("span not recording under a ratio of 1")
	}
	span.End()
	if got := len(exporter.GetSpans()); got != 0 {
		t.Fatalf("exported %d spans before shutdown, want 0 (batched)", got)
	}

	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans after shutdown, want 1", len(spans))
	}
	if spans[0].Name != "op" {
		t.Errorf("span name = %q, want op", spans[0].Name)
	}
	// The configured attributes ride the resource, merged over the SDK's
	// defaults, which contribute the telemetry SDK attributes.
	wantAttribute(t, spans[0].Resource, "service.name", "app")
	wantAttribute(t, spans[0].Resource, "service.version", "1.2.3")
	wantAttribute(t, spans[0].Resource, "telemetry.sdk.language", "go")
}

func TestTelemetry_SamplerHonorsRatio(t *testing.T) {
	tel, exporter, _ := startTelemetry(t, 0)

	// A ratio of 0 means never sample: the root span records nothing and the
	// exporter sees nothing after the flush.
	_, span := otel.Tracer("test").Start(context.Background(), "op")
	if span.IsRecording() {
		t.Error("span recording under a ratio of 0")
	}
	span.End()

	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if got := len(exporter.GetSpans()); got != 0 {
		t.Errorf("exported %d spans under a ratio of 0, want 0", got)
	}
}

func TestTelemetry_MetricsCarryResource(t *testing.T) {
	_, _, reader := startTelemetry(t, 1)

	counter, err := otel.Meter("test").Int64Counter("requests")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	counter.Add(context.Background(), 1)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	wantAttribute(t, rm.Resource, "service.name", "app")
	wantAttribute(t, rm.Resource, "service.version", "1.2.3")
	if len(rm.ScopeMetrics) != 1 || len(rm.ScopeMetrics[0].Metrics) != 1 {
		t.Fatalf("collected %+v, want one scope with one metric", rm.ScopeMetrics)
	}
	if got := rm.ScopeMetrics[0].Metrics[0].Name; got != "requests" {
		t.Errorf("metric name = %q, want requests", got)
	}
}

func TestTelemetry_ShutdownReleasesReader(t *testing.T) {
	tel, _, reader := startTelemetry(t, 1)

	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// The reader is shut down with its provider, so a later collection
	// fails rather than returning stale data.
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err == nil {
		t.Error("collect after shutdown succeeded, want an error")
	}
}
