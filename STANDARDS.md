# go-observability standards

The judgement calls the standards-reviewer applies to go-observability, beyond what `mise run check` enforces.

- A change that alters documented behavior updates the README and the affected `doc.go` in the same change.
- `architecture/standards/go-elemental/principles/dependencies.md`: the bottom-up line and no provider in a base, across the base module's and `otlp`'s `go.mod`.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: the doc.go inventory of `observability` and `otlp`, and `otlp`'s unit tests against an in-test collector on a loopback port 0.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: the base module and the `otlp` sub-module, and their tags.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: the check and currency over `GO_MODULES`, and a `CHANGELOG.md` per module.
- `architecture/standards/go-elemental/principles/lifecycle-and-context.md`: `Telemetry`'s `Start` and `Shutdown`, and `New`'s panics on an unfinalized `Config` or a missing exporter.
- `architecture/standards/go-elemental/principles/baseline-standards.md`: `Config` takes the endpoint, headers, sample ratio, and resource attributes from the application, and `NewMiddleware` and `RequestIDSource` follow OpenTelemetry's conventions alone.
- `architecture/principles/service-tiers.md`: the `observability` package and the `otlp` exporters.
- `architecture/principles/context-architecture.md`: the README and each `doc.go` are the homes.
