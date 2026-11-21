# Logs pipeline

The planned OpenTelemetry logs pipeline adds `otel/log`, `otel/sdk/log`, the `slog` bridge
(`otelslog`), and an OTLP log exporter (`otlploggrpc`). Until it is built, logs go to stdout, as
the README's Design section explains.

**Trigger.** The pipeline is built when `otel/log` and its `slog` bridge reach a stable v1, the
point at which the README's "Logs go to stdout, not OTLP" choice reverses.

**Placement.** The log exporter joins the other OTLP exporters in the `otlp` sub-module, so the
base module keeps no exporter.

## Open questions for the plan

- `otel/log` is v1; the `slog` bridge and `otlploggrpc` are pre-v1. Does the trigger wait for both
  to reach v1, or admit them as v0 exceptions, as `otelhttp` is admitted?
