package observability

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

const defaultSampleRatio = 1.0

// Config holds the collector address, the headers the exporter sends, the
// trace sampling ratio, and the resource attributes for OpenTelemetry.
// Endpoint is the collector's OTLP/gRPC address ("localhost:4317") and is
// the one required field: the collector location is always
// deployment-specific, so it has no default. There is no protocol field
// because the otlp sub-module ships gRPC exporters alone; a protocol field
// returns if HTTP/protobuf is ever added. Headers carries extra headers the
// gRPC exporter sends with every request, such as an auth token for a
// managed backend; a sensitive value rides the secrets layer of
// [config.Load] rather than a committed file. SampleRatio is the trace
// sampler's ratio in [0, 1], a tri-state pointer: nil is unset and takes the
// default of 1 (sample everything), while an explicit 0 survives the load
// and means never sample. ResourceAttributes carries the resource attributes
// the composition root supplies (service.name, service.version,
// deployment.environment) as plain key-value data; building the resource
// from them is the Telemetry service's job. Env records the
// environment-variable names Finalize composed and read; it is excluded from
// JSON.
type Config struct {
	Endpoint           string            `json:"endpoint"`
	Headers            map[string]string `json:"headers"`
	SampleRatio        *float64          `json:"sample_ratio"`
	ResourceAttributes map[string]string `json:"resource_attributes"`
	Env                Env               `json:"-"`
}

// Merge overlays src's set fields onto the receiver. Headers and
// ResourceAttributes merge key-wise, so an overlay can set one header or one
// attribute without dropping the rest.
func (c *Config) Merge(src *Config) {
	if src.Endpoint != "" {
		c.Endpoint = src.Endpoint
	}
	if src.SampleRatio != nil {
		c.SampleRatio = src.SampleRatio
	}

	for k, v := range src.Headers {
		if c.Headers == nil {
			c.Headers = make(map[string]string, len(src.Headers))
		}
		c.Headers[k] = v
	}
	for k, v := range src.ResourceAttributes {
		if c.ResourceAttributes == nil {
			c.ResourceAttributes = make(map[string]string, len(src.ResourceAttributes))
		}
		c.ResourceAttributes[k] = v
	}
}

// Finalize composes the environment override names from envPrefix (an empty
// prefix disables overrides), applies defaults, applies the overrides, and
// validates. Endpoint is the one required field; a malformed override fails
// with an error naming its variable.
func (c *Config) Finalize(envPrefix string) error {
	c.Env = NewEnv(envPrefix)
	c.applyDefaults()
	if err := c.applyEnv(); err != nil {
		return err
	}
	return c.validate()
}

func (c *Config) applyDefaults() {
	if c.SampleRatio == nil {
		c.SampleRatio = new(defaultSampleRatio)
	}
}

func (c *Config) applyEnv() error {
	if v := os.Getenv(c.Env.Endpoint); v != "" {
		c.Endpoint = v
	}
	if v := os.Getenv(c.Env.SampleRatio); v != "" {
		ratio, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("%s: %w", c.Env.SampleRatio, err)
		}
		c.SampleRatio = &ratio
	}
	return nil
}

func (c *Config) validate() error {
	if c.Endpoint == "" {
		return errors.New("observability endpoint required")
	}
	// The negated conjunction rejects NaN, which every ordered comparison
	// would otherwise let through.
	if ratio := *c.SampleRatio; !(ratio >= 0 && ratio <= 1) {
		return fmt.Errorf("invalid sample_ratio: %g", ratio)
	}
	return nil
}
