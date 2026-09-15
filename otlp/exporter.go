package otlp

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/standards-lab/go-observability"
)

// ErrEndpointRequired is the error both constructors return when the Config
// they receive has an empty Endpoint. Config.Finalize enforces the same rule
// in the base module, but a Config can reach this package without having
// been finalized, so the check repeats here rather than letting the gRPC
// client fail later with a less direct message.
var ErrEndpointRequired = errors.New("otlp: observability endpoint required")

// NewTraceExporter builds the gRPC OTLP span exporter over cfg: it dials
// cfg.Endpoint, sends cfg.Headers with every request, and connects in plain
// text (see the package comment on transport security). The exporter dials
// lazily, so construction succeeds with no collector listening; an
// unreachable collector surfaces as export errors, never here. The result is
// what [observability.Exporters].Trace takes. It returns
// [ErrEndpointRequired] when cfg.Endpoint is empty.
func NewTraceExporter(ctx context.Context, cfg observability.Config) (sdktrace.SpanExporter, error) {
	if cfg.Endpoint == "" {
		return nil, ErrEndpointRequired
	}
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithHeaders(cfg.Headers),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp: new trace exporter: %w", err)
	}
	return exporter, nil
}

// NewMetricExporter builds the gRPC OTLP metric exporter over cfg, with the
// same endpoint, headers, plain-text transport, and lazy dial as
// [NewTraceExporter]. It returns the raw exporter, not a reader:
// [observability.Exporters].Metric takes a reader, and the composition root
// decides how the exporter becomes one, in production by wrapping it in
// sdkmetric.NewPeriodicReader. It returns [ErrEndpointRequired] when
// cfg.Endpoint is empty.
func NewMetricExporter(ctx context.Context, cfg observability.Config) (sdkmetric.Exporter, error) {
	if cfg.Endpoint == "" {
		return nil, ErrEndpointRequired
	}
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithHeaders(cfg.Headers),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp: new metric exporter: %w", err)
	}
	return exporter, nil
}
