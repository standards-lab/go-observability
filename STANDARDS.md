# go-observability standards

The judgement calls the standards-reviewer applies to go-observability.

- A change that alters documented behavior updates the README and the affected `doc.go` in the same change.
- `architecture/standards/go-elemental/principles/dependencies.md`: the bottom-up line and no provider in a base, held by review; the base `go.mod` takes go-core, the OpenTelemetry API and SDK, and `otelhttp` alone, and every exporter enters through `otlp/go.mod`, which the base module never imports.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: the doc.go inventory of `observability` and `otlp`, held by review, and `otlp`'s unit tests against an in-test collector on a loopback port 0.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: the base module at the root and the `otlp` sub-module, tagged `v*` and `otlp/v*`.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: the check and currency over `GO_MODULES`, and a `CHANGELOG.md` per module.
- `architecture/standards/go-elemental/principles/baseline-standards.md`: `Config` takes the endpoint, headers, sample ratio, and resource attributes from the application, and `NewMiddleware` and `RequestIDSource` follow OpenTelemetry's conventions alone.
- `architecture/principles/service-tiers.md`: OpenTelemetry is the standard tier and there is no native tier; `otlp` speaks OTLP, and no backend enters either module's graph.
- `architecture/principles/context-architecture.md`: the README and each `doc.go` are the homes; `context/` records only what they do not express.
