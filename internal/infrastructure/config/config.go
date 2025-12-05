package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config holds all configuration for the agent
type Config struct {
	Agent    AgentConfig    `mapstructure:"agent"`
	Libvirt  LibvirtConfig  `mapstructure:"libvirt"`
	Resources ResourceConfig `mapstructure:"resources"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Health   HealthConfig   `mapstructure:"health"`
}

type AgentConfig struct {
	Name              string        `mapstructure:"name" validate:"required"`
	APIURL            string        `mapstructure:"api_url" validate:"required,url"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval" validate:"required"`
	Version           string        `mapstructure:"version"`
}

type LibvirtConfig struct {
	URI         string `mapstructure:"uri" validate:"required"`
	StoragePool string `mapstructure:"storage_pool" validate:"required"`
	Network     string `mapstructure:"network" validate:"required"`
	ImageCache  string `mapstructure:"image_cache" validate:"required"`
}

type ResourceConfig struct {
	ReservedCPU    int `mapstructure:"reserved_cpu" validate:"min=0"`
	ReservedRAMGB  int `mapstructure:"reserved_ram_gb" validate:"min=0"`
	ReservedDiskGB int `mapstructure:"reserved_disk_gb" validate:"min=0"`
}

type GRPCConfig struct {
	ListenAddr string `mapstructure:"listen_addr" validate:"required"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	TLSCert    string `mapstructure:"tls_cert"`
	TLSKey     string `mapstructure:"tls_key"`
	TLSCA      string `mapstructure:"tls_ca"`
}

type LoggingConfig struct {
	Level   string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Output  string `mapstructure:"output" validate:"required,oneof=stdout file both"`
	File    string `mapstructure:"file"`
	Format  string `mapstructure:"format" validate:"required,oneof=json text"`
	MaxSize int    `mapstructure:"max_size" validate:"min=1"`
	MaxAge  int    `mapstructure:"max_age" validate:"min=1"`
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ListenAddr string `mapstructure:"listen_addr"`
	Path       string `mapstructure:"path"`
}

type HealthConfig struct {
	ListenAddr string `mapstructure:"listen_addr"`
	Path       string `mapstructure:"path"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Override with environment variables
	viper.SetEnvPrefix("GHOST")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
