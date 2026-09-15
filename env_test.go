package observability_test

import (
	"testing"

	"github.com/standards-lab/go-observability"
)

// Every name derives from the prefix through config.EnvName with the
// "observability" segment; this pins the full set a consumer binds to.
func TestNewEnv(t *testing.T) {
	env := observability.NewEnv("app")

	cases := []struct {
		field string
		got   string
		want  string
	}{
		{"Endpoint", env.Endpoint, "APP_OBSERVABILITY_ENDPOINT"},
		{"SampleRatio", env.SampleRatio, "APP_OBSERVABILITY_SAMPLE_RATIO"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %s, want %s", tc.field, tc.got, tc.want)
		}
	}
}

func TestNewEnv_EmptyPrefixReturnsZeroEnv(t *testing.T) {
	if env := observability.NewEnv(""); env != (observability.Env{}) {
		t.Errorf("NewEnv(\"\") = %+v, want the zero Env (overrides disabled)", env)
	}
}
