package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration file.
type Config struct {
	Server   ServerConfig  `yaml:"server"`
	Metrics  MetricsConfig `yaml:"metrics"`
	Storage  StorageConfig `yaml:"storage"`
	Policies []Policy      `yaml:"policies"`
}

// ServerConfig configures the HTTP listener.
type ServerConfig struct {
	Address      string   `yaml:"address"`
	GRPCAddr     string   `yaml:"grpc_address"`
	ReadTimeout  Duration `yaml:"read_timeout"`
	WriteTimeout Duration `yaml:"write_timeout"`
	IdleTimeout  Duration `yaml:"idle_timeout"`
}

// GRPCAddress returns the configured gRPC listener.
func (s ServerConfig) GRPCAddress() string {
	return s.GRPCAddr
}

// MetricsConfig controls Prometheus exposure.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// StorageConfig describes the storage driver.
type StorageConfig struct {
	Driver string      `yaml:"driver"`
	Redis  RedisConfig `yaml:"redis"`
}

// RedisConfig holds redis specific settings.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Policy binds a limiter to routes/methods.
type Policy struct {
	Name      string          `yaml:"name"`
	Routes    []string        `yaml:"routes"`
	Methods   []string        `yaml:"methods"`
	Identity  IdentityConfig  `yaml:"identity"`
	Algorithm AlgorithmConfig `yaml:"algorithm"`
}

// IdentityConfig defines how to extract an identity key.
type IdentityConfig struct {
	Type     string `yaml:"type"`
	Key      string `yaml:"key"`
	Fallback string `yaml:"fallback"`
}

// AlgorithmConfig configures an algorithm instance.
type AlgorithmConfig struct {
	Type       string   `yaml:"type"`
	Limit      int      `yaml:"limit"`
	Burst      int      `yaml:"burst"`
	RefillRate int      `yaml:"refill_rate"`
	Interval   Duration `yaml:"interval"`
	LeakRate   float64  `yaml:"leak_rate"`
	Window     Duration `yaml:"window"`
}

// Duration allows YAML parsing of Go duration strings.
type Duration time.Duration

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// UnmarshalYAML parses duration strings such as "1s" or integers representing seconds.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var asString string
	if err := value.Decode(&asString); err == nil {
		parsed, err := time.ParseDuration(asString)
		if err != nil {
			return err
		}
		*d = Duration(parsed)
		return nil
	}

	var asInt int64
	if err := value.Decode(&asInt); err == nil {
		*d = Duration(time.Duration(asInt) * time.Second)
		return nil
	}

	return fmt.Errorf("duration must be a string like \"1s\" or an integer representing seconds")
}

// Load parses the YAML configuration file.
func Load(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	cfg.setDefaults()
	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Server.Address == "" {
		c.Server.Address = ":8080"
	}
	if c.Server.GRPCAddr == "" {
		c.Server.GRPCAddr = ":9090"
	}
	if c.Server.ReadTimeout.Duration() == 0 {
		c.Server.ReadTimeout = Duration(5 * time.Second)
	}
	if c.Server.WriteTimeout.Duration() == 0 {
		c.Server.WriteTimeout = Duration(10 * time.Second)
	}
	if c.Server.IdleTimeout.Duration() == 0 {
		c.Server.IdleTimeout = Duration(60 * time.Second)
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Storage.Driver == "" {
		c.Storage.Driver = "memory"
	}
}
