package observability

import "github.com/standards-lab/go-core/config"

// Env names the environment variables [Config.Finalize] reads, composed from
// the prefix it receives under the "observability" segment:
// OBSERVABILITY_ENDPOINT and OBSERVABILITY_SAMPLE_RATIO, prefixed by whatever
// [config.EnvName] produces. Headers and ResourceAttributes are maps and have
// no override. An empty name disables that one override, and the zero Env
// (an empty prefix) disables all of them. Populated by Finalize and exposed
// for introspection.
type Env struct {
	Endpoint    string
	SampleRatio string
}

// NewEnv composes the standard override names from a prefix under the
// "observability" segment: OBSERVABILITY_ENDPOINT and
// OBSERVABILITY_SAMPLE_RATIO, prefixed by whatever [config.EnvName] produces.
// An empty prefix returns the zero Env, disabling the overrides.
func NewEnv(prefix string) Env {
	if prefix == "" {
		return Env{}
	}
	return Env{
		Endpoint: config.EnvName(
			prefix, "observability", "endpoint",
		),
		SampleRatio: config.EnvName(
			prefix, "observability", "sample", "ratio",
		),
	}
}
