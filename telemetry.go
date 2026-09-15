package observability

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Exporters carries the exporter-side components the composition root builds
// and [Telemetry] wires into its providers. Trace is the span exporter; Start
// wraps it in the SDK's batch span processor. Metric is a reader rather than
// a raw exporter, because the composition root decides how the metric
// exporter becomes one: a periodic reader around the OTLP exporter in
// production, a manual reader in a hermetic test. Both arrive from outside
// because the gRPC exporters live in the otlp sub-module, which the base
// module never imports.
type Exporters struct {
	Trace  sdktrace.SpanExporter
	Metric sdkmetric.Reader
}

// Telemetry is the process's OpenTelemetry lifecycle service. Construction
// performs no work beyond recording the configuration and exporters; Start
// builds the resource and the providers and installs them as process
// globals; Shutdown flushes and releases them. Start and Shutdown carry the
// lifecycle hook signature, so the composition root registers the bare
// method values at stage 0, ahead of everything else it starts. There is no
// readiness check: observability never gates readiness, so a collector the
// service cannot reach is never a reason to fail a probe.
type Telemetry struct {
	sampleRatio float64
	attributes  []attribute.KeyValue
	exporters   Exporters

	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

// New records cfg's sampling ratio and resource attributes and the exporters
// Start will wire in. It panics if cfg was not finalized or if either
// exporter is nil: each is a wiring defect at the composition root, not a
// runtime condition.
func New(cfg Config, exporters Exporters) *Telemetry {
	if !cfg.finalized() {
		panic("observability: Config not finalized: call Finalize before New")
	}
	if exporters.Trace == nil {
		panic("observability: nil trace exporter")
	}
	if exporters.Metric == nil {
		panic("observability: nil metric reader")
	}

	attrs := make([]attribute.KeyValue, 0, len(cfg.ResourceAttributes))
	for k, v := range cfg.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	return &Telemetry{
		sampleRatio: *cfg.SampleRatio,
		attributes:  attrs,
		exporters:   exporters,
	}
}

// Start builds the resource, constructs the providers, and installs them as
// process globals. The resource merges the SDK's defaults (the fallback
// service name, the telemetry SDK attributes, and OTEL_RESOURCE_ATTRIBUTES)
// with the configured resource attributes, the configured ones winning. The
// TracerProvider batches spans to the trace exporter and samples by a
// parent-based ratio sampler at the configured ratio; the MeterProvider
// collects through the metric reader. Both install as the otel globals
// alongside the W3C TraceContext propagator. The context exists for the
// lifecycle hook signature; nothing here performs I/O.
func (t *Telemetry) Start(ctx context.Context) error {
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(t.attributes...))
	if err != nil {
		return fmt.Errorf("build resource: %w", err)
	}

	t.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(t.exporters.Trace),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(t.sampleRatio))),
	)
	t.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(t.exporters.Metric),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(t.tracerProvider)
	otel.SetMeterProvider(t.meterProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return nil
}

// Shutdown shuts down both providers under ctx, which flushes pending spans
// and metrics to the exporters and releases them, and joins the errors. The
// lifecycle calls it only after a successful Start, so the providers are
// always present. The globals stay installed; tracers and meters obtained
// from them keep working but record nothing.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	if err := t.tracerProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown tracer provider: %w", err))
	}
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown meter provider: %w", err))
	}
	return errors.Join(errs...)
}
