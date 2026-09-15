package observability_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/standards-lab/go-observability"
)

// validConfig satisfies the one required field; everything else exercises
// defaults under test.
func validConfig() observability.Config {
	return observability.Config{Endpoint: "localhost:4317"}
}

func TestConfig_MergeOverlaysSetFields(t *testing.T) {
	base := observability.Config{
		Endpoint:    "localhost:4317",
		SampleRatio: new(0.5),
	}
	overlay := observability.Config{
		SampleRatio: new(0.25),
	}

	base.Merge(&overlay)

	if base.SampleRatio == nil || *base.SampleRatio != 0.25 {
		t.Errorf("SampleRatio = %v, want 0.25", base.SampleRatio)
	}
	// Fields the overlay leaves unset keep the base values.
	if base.Endpoint != "localhost:4317" {
		t.Errorf("Endpoint = %s, want localhost:4317", base.Endpoint)
	}
}

func TestConfig_MergeExplicitZeroRatioWins(t *testing.T) {
	// A set zero is a value, not an absence: the overlay's 0 replaces the
	// base's 1.
	base := observability.Config{
		Endpoint:    "localhost:4317",
		SampleRatio: new(1.0),
	}
	overlay := observability.Config{SampleRatio: new(0.0)}

	base.Merge(&overlay)

	if base.SampleRatio == nil || *base.SampleRatio != 0 {
		t.Errorf("SampleRatio = %v, want 0", base.SampleRatio)
	}
}

func TestConfig_MergeHeadersKeyWise(t *testing.T) {
	base := observability.Config{
		Endpoint: "localhost:4317",
		Headers:  map[string]string{"authorization": "old", "x-tenant": "acme"},
	}
	overlay := observability.Config{
		Headers: map[string]string{"authorization": "new"},
	}

	base.Merge(&overlay)

	if got := base.Headers["authorization"]; got != "new" {
		t.Errorf("Headers[authorization] = %s, want new", got)
	}
	// An overlay key overrides without dropping the others.
	if got := base.Headers["x-tenant"]; got != "acme" {
		t.Errorf("Headers[x-tenant] = %s, want acme", got)
	}
}

func TestConfig_MergeHeadersOntoNilMap(t *testing.T) {
	base := validConfig()
	overlay := observability.Config{Headers: map[string]string{"authorization": "tok"}}

	base.Merge(&overlay)

	if got := base.Headers["authorization"]; got != "tok" {
		t.Errorf("Headers[authorization] = %s, want tok", got)
	}
}

func TestConfig_MergeResourceAttributesKeyWise(t *testing.T) {
	base := observability.Config{
		Endpoint: "localhost:4317",
		ResourceAttributes: map[string]string{
			"service.name":           "svc",
			"deployment.environment": "dev",
		},
	}
	overlay := observability.Config{
		ResourceAttributes: map[string]string{"deployment.environment": "prod"},
	}

	base.Merge(&overlay)

	if got := base.ResourceAttributes["deployment.environment"]; got != "prod" {
		t.Errorf("ResourceAttributes[deployment.environment] = %s, want prod", got)
	}
	if got := base.ResourceAttributes["service.name"]; got != "svc" {
		t.Errorf("ResourceAttributes[service.name] = %s, want svc", got)
	}
}

func TestConfig_MergeResourceAttributesOntoNilMap(t *testing.T) {
	base := validConfig()
	overlay := observability.Config{
		ResourceAttributes: map[string]string{"service.name": "svc"},
	}

	base.Merge(&overlay)

	if got := base.ResourceAttributes["service.name"]; got != "svc" {
		t.Errorf("ResourceAttributes[service.name] = %s, want svc", got)
	}
}

func TestConfig_JSONZeroRatioIsSet(t *testing.T) {
	// The pointer distinguishes an explicit 0 in a file from an absent key,
	// so a layer that says "never sample" is not mistaken for one that says
	// nothing.
	var set, unset observability.Config
	if err := json.Unmarshal([]byte(`{"sample_ratio": 0}`), &set); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"endpoint": "localhost:4317"}`), &unset); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if set.SampleRatio == nil || *set.SampleRatio != 0 {
		t.Errorf("SampleRatio = %v after explicit 0, want 0", set.SampleRatio)
	}
	if unset.SampleRatio != nil {
		t.Errorf("SampleRatio = %v after absent key, want nil", unset.SampleRatio)
	}
}

func TestConfig_FinalizeDefaults(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Finalize(""); err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	if cfg.SampleRatio == nil || *cfg.SampleRatio != 1 {
		t.Errorf("SampleRatio = %v, want 1", cfg.SampleRatio)
	}
	// The maps keep no default: nil means no headers and no attributes.
	if cfg.Headers != nil {
		t.Errorf("Headers = %v, want nil", cfg.Headers)
	}
	if cfg.ResourceAttributes != nil {
		t.Errorf("ResourceAttributes = %v, want nil", cfg.ResourceAttributes)
	}
}

func TestConfig_FinalizeRequiresEndpoint(t *testing.T) {
	cfg := observability.Config{SampleRatio: new(0.5)}
	err := cfg.Finalize("")
	if err == nil {
		t.Fatal("Finalize accepted a config with no endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint required") {
		t.Errorf("error = %v, want it to name the missing field", err)
	}
}

func TestConfig_FinalizeEnvOverrides(t *testing.T) {
	cfg := validConfig()

	t.Setenv("TEST_OBSERVABILITY_ENDPOINT", "collector.internal:4317")
	t.Setenv("TEST_OBSERVABILITY_SAMPLE_RATIO", "0.1")

	if err := cfg.Finalize("test"); err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	if cfg.Endpoint != "collector.internal:4317" {
		t.Errorf("Endpoint = %s, want collector.internal:4317", cfg.Endpoint)
	}
	if cfg.SampleRatio == nil || *cfg.SampleRatio != 0.1 {
		t.Errorf("SampleRatio = %v, want 0.1", cfg.SampleRatio)
	}
}

func TestConfig_FinalizeEnvSuppliesRequiredEndpoint(t *testing.T) {
	// An override satisfies the required field when no file set it.
	var cfg observability.Config
	t.Setenv("TEST_OBSERVABILITY_ENDPOINT", "collector.internal:4317")

	if err := cfg.Finalize("test"); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if cfg.Endpoint != "collector.internal:4317" {
		t.Errorf("Endpoint = %s, want collector.internal:4317", cfg.Endpoint)
	}
}

func TestConfig_FinalizeMalformedEnvFails(t *testing.T) {
	cfg := validConfig()
	t.Setenv("TEST_OBSERVABILITY_SAMPLE_RATIO", "half")

	err := cfg.Finalize("test")
	if err == nil {
		t.Fatal(`Finalize accepted TEST_OBSERVABILITY_SAMPLE_RATIO="half"`)
	}
	if !strings.Contains(err.Error(), "TEST_OBSERVABILITY_SAMPLE_RATIO") {
		t.Errorf("error = %v, want it to name the variable", err)
	}
}

func TestConfig_FinalizeZeroEnvDisablesOverrides(t *testing.T) {
	// The zero Env names no variables, so ambient values cannot leak in.
	t.Setenv("TEST_OBSERVABILITY_ENDPOINT", "collector.internal:4317")
	t.Setenv("TEST_OBSERVABILITY_SAMPLE_RATIO", "0.1")

	cfg := validConfig()
	if err := cfg.Finalize(""); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if cfg.Endpoint != "localhost:4317" {
		t.Errorf("Endpoint = %s with zero Env, want localhost:4317", cfg.Endpoint)
	}
	if cfg.SampleRatio == nil || *cfg.SampleRatio != 1 {
		t.Errorf("SampleRatio = %v with zero Env, want 1", cfg.SampleRatio)
	}
}

func TestConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*observability.Config)
		wantErr string
	}{
		{
			"negative sample ratio",
			func(c *observability.Config) { c.SampleRatio = new(-0.1) },
			"invalid sample_ratio: -0.1",
		},
		{
			"sample ratio above one",
			func(c *observability.Config) { c.SampleRatio = new(1.5) },
			"invalid sample_ratio: 1.5",
		},
		{
			"sample ratio NaN",
			func(c *observability.Config) { c.SampleRatio = new(math.NaN()) },
			"invalid sample_ratio: NaN",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.mutate(&cfg)

			err := cfg.Finalize("")
			if err == nil {
				t.Fatal("Finalize accepted an invalid config")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestConfig_ValidateAllowsRatioBounds(t *testing.T) {
	// Both ends of [0, 1] are valid: never sample, and sample everything.
	for _, ratio := range []float64{0, 1} {
		cfg := validConfig()
		cfg.SampleRatio = new(ratio)

		if err := cfg.Finalize(""); err != nil {
			t.Errorf("Finalize rejected sample_ratio %g: %v", ratio, err)
		}
	}
}
