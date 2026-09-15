package otlp_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/standards-lab/go-observability"
	"github.com/standards-lab/go-observability/otlp"
)

// The tests build their own providers over the exporters rather than
// installing otel globals, so they run in parallel.

const testTimeout = 5 * time.Second

// testContext returns a context bounded by testTimeout, so a hang in the
// exporter fails the test instead of the whole run.
func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	t.Cleanup(cancel)
	return ctx
}

// traceCollector and metricsCollector are the two halves of an in-process
// OTLP collector: each implements one collector service, records what it
// receives, and keeps the request metadata so a test can see the headers the
// exporter sent. They are two types because both service interfaces name
// their method Export with different signatures.
type traceCollector struct {
	coltracepb.UnimplementedTraceServiceServer
	mu      sync.Mutex
	spans   []*tracepb.ResourceSpans
	headers metadata.MD
}

func (c *traceCollector) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spans = append(c.spans, req.GetResourceSpans()...)
	c.headers, _ = metadata.FromIncomingContext(ctx)
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

func (c *traceCollector) spanNames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var names []string
	for _, rs := range c.spans {
		for _, ss := range rs.GetScopeSpans() {
			for _, s := range ss.GetSpans() {
				names = append(names, s.GetName())
			}
		}
	}
	return names
}

type metricsCollector struct {
	colmetricpb.UnimplementedMetricsServiceServer
	mu      sync.Mutex
	metrics []*metricpb.ResourceMetrics
	headers metadata.MD
}

func (c *metricsCollector) Export(ctx context.Context, req *colmetricpb.ExportMetricsServiceRequest) (*colmetricpb.ExportMetricsServiceResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, req.GetResourceMetrics()...)
	c.headers, _ = metadata.FromIncomingContext(ctx)
	return &colmetricpb.ExportMetricsServiceResponse{}, nil
}

func (c *metricsCollector) metricNames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var names []string
	for _, rm := range c.metrics {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				names = append(names, m.GetName())
			}
		}
	}
	return names
}

// startCollector serves both collector services on a loopback listener and
// returns its address, the form Config.Endpoint takes, alongside the two
// recorders. The server stops at cleanup.
func startCollector(t *testing.T) (string, *traceCollector, *metricsCollector) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	tc, mc := &traceCollector{}, &metricsCollector{}
	srv := grpc.NewServer()
	coltracepb.RegisterTraceServiceServer(srv, tc)
	colmetricpb.RegisterMetricsServiceServer(srv, mc)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String(), tc, mc
}

func wantHeader(t *testing.T, md metadata.MD, key, want string) {
	t.Helper()
	got := md.Get(key)
	if len(got) != 1 || got[0] != want {
		t.Errorf("request header %s = %q, want [%q]", key, got, want)
	}
}

func TestNewTraceExporter_RequiresEndpoint(t *testing.T) {
	t.Parallel()
	exporter, err := otlp.NewTraceExporter(testContext(t), observability.Config{})
	if !errors.Is(err, otlp.ErrEndpointRequired) {
		t.Errorf("err = %v, want ErrEndpointRequired", err)
	}
	if exporter != nil {
		t.Errorf("exporter = %v, want nil", exporter)
	}
}

func TestNewMetricExporter_RequiresEndpoint(t *testing.T) {
	t.Parallel()
	exporter, err := otlp.NewMetricExporter(testContext(t), observability.Config{})
	if !errors.Is(err, otlp.ErrEndpointRequired) {
		t.Errorf("err = %v, want ErrEndpointRequired", err)
	}
	if exporter != nil {
		t.Errorf("exporter = %v, want nil", exporter)
	}
}

// The constructors dial lazily, so construction and shutdown both succeed
// against an endpoint nothing listens on.
func TestNewTraceExporter_ConstructsAndShutsDownWithoutCollector(t *testing.T) {
	t.Parallel()
	ctx := testContext(t)
	exporter, err := otlp.NewTraceExporter(ctx, observability.Config{Endpoint: "127.0.0.1:1"})
	if err != nil {
		t.Fatalf("new trace exporter: %v", err)
	}
	if exporter == nil {
		t.Fatal("exporter is nil")
	}
	if err := exporter.Shutdown(ctx); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

func TestNewMetricExporter_ConstructsAndShutsDownWithoutCollector(t *testing.T) {
	t.Parallel()
	ctx := testContext(t)
	exporter, err := otlp.NewMetricExporter(ctx, observability.Config{Endpoint: "127.0.0.1:1"})
	if err != nil {
		t.Fatalf("new metric exporter: %v", err)
	}
	if exporter == nil {
		t.Fatal("exporter is nil")
	}
	if err := exporter.Shutdown(ctx); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

// A span ended under a synchronous processor reaches the in-process
// collector over the wire, carrying the configured headers.
func TestNewTraceExporter_ExportsToCollector(t *testing.T) {
	t.Parallel()
	ctx := testContext(t)
	addr, tc, _ := startCollector(t)

	exporter, err := otlp.NewTraceExporter(ctx, observability.Config{
		Endpoint: addr,
		Headers:  map[string]string{"x-test-token": "secret"},
	})
	if err != nil {
		t.Fatalf("new trace exporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	_, span := tp.Tracer("test").Start(ctx, "op")
	span.End()
	if err := tp.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown tracer provider: %v", err)
	}

	if names := tc.spanNames(); len(names) != 1 || names[0] != "op" {
		t.Fatalf("collector received spans %q, want [op]", names)
	}
	wantHeader(t, tc.headers, "x-test-token", "secret")
}

// A measurement flushed through a periodic reader over the exporter reaches
// the in-process collector over the wire, carrying the configured headers.
func TestNewMetricExporter_ExportsToCollector(t *testing.T) {
	t.Parallel()
	ctx := testContext(t)
	addr, _, mc := startCollector(t)

	exporter, err := otlp.NewMetricExporter(ctx, observability.Config{
		Endpoint: addr,
		Headers:  map[string]string{"x-test-token": "secret"},
	})
	if err != nil {
		t.Fatalf("new metric exporter: %v", err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))

	counter, err := mp.Meter("test").Int64Counter("requests")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	counter.Add(ctx, 1)
	// Shutdown collects and exports once more before releasing the reader;
	// that final collect is the one export here. A ForceFlush ahead of it
	// would export the cumulative counter a second time.
	if err := mp.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown meter provider: %v", err)
	}

	if names := mc.metricNames(); len(names) != 1 || names[0] != "requests" {
		t.Fatalf("collector received metrics %q, want [requests]", names)
	}
	wantHeader(t, mc.headers, "x-test-token", "secret")
}
