# Logs pipeline

The OpenTelemetry logs pipeline: `otel/log`, `otel/sdk/log`, the `slog` bridge (`otelslog`), and
an OTLP log exporter (`otlploggrpc`). Logs go to stdout until it is built, as the README's Design
section says.

**Trigger.** The pipeline is built when `otel/log` and its `slog` bridge reach a stable v1, the
point at which the README's "Logs go to stdout, not OTLP" choice reverses.

**Placement.** The log exporter joins the other OTLP exporters in the `otlp` sub-module, so the
base module keeps no exporter.

## Open at the logs pipeline plan

- `otel/log` is now v1.47.0, while the `slog` bridge (`otelslog`, v0.21.0) and `otlploggrpc`
  (v0.23.0) remain pre-v1; whether the trigger waits for both or admits them as v0 exceptions,
  as `otelhttp` is admitted.
