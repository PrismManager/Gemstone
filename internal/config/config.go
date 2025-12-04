package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = "/etc/gemstone"
	// DefaultDataDir is the default data directory
	DefaultDataDir = "/var/lib/gemstone"
	// DefaultLogDir is the default log directory
	DefaultLogDir = "/var/log/gemstone"
	// DefaultRunDir is the default runtime directory
	DefaultRunDir = "/run/gemstone"
	// DefaultAPIPort is the default API port
	DefaultAPIPort = 9876
	// DefaultSocketPath is the default Unix socket path
	DefaultSocketPath = "/run/gemstone/gemstone.sock"
)

// Config represents the main configuration
type Config struct {
	API       APIConfig `yaml:"api"`
	Logging   LogConfig `yaml:"logging"`
	Processes []Process `yaml:"processes,omitempty"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Port       int    `yaml:"port"`
	Host       string `yaml:"host"`
	AuthToken  string `yaml:"auth_token,omitempty"`
	EnableCORS bool   `yaml:"enable_cors"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	MaxSize    int    `yaml:"max_size"`    // Max size in MB
	MaxBackups int    `yaml:"max_backups"` // Max number of backups
	MaxAge     int    `yaml:"max_age"`     // Max age in days
	Compress   bool   `yaml:"compress"`
	Directory  string `yaml:"directory"`
}

// Process represents a managed process configuration
type Process struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args,omitempty"`
	WorkDir     string            `yaml:"work_dir,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	AutoStart   bool              `yaml:"auto_start"`
	AutoRestart bool              `yaml:"auto_restart"`
	MaxRestarts int               `yaml:"max_restarts"`
	User        string            `yaml:"user,omitempty"`
	Group       string            `yaml:"group,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Enabled:    true,
			Port:       DefaultAPIPort,
			Host:       "127.0.0.1",
			EnableCORS: false,
		},
		Logging: LogConfig{
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
			Directory:  DefaultLogDir,
		},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	if p := os.Getenv("GEMSTONE_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(DefaultConfigDir, "config.yaml")
}

// GetDataPath returns the data directory path
func GetDataPath() string {
	if p := os.Getenv("GEMSTONE_DATA"); p != "" {
		return p
	}
	return DefaultDataDir
}

// GetLogPath returns the log directory path
func GetLogPath() string {
	if p := os.Getenv("GEMSTONE_LOG"); p != "" {
		return p
	}
	return DefaultLogDir
}

// GetSocketPath returns the Unix socket path
func GetSocketPath() string {
	if p := os.Getenv("GEMSTONE_SOCKET"); p != "" {
		return p
	}
	return DefaultSocketPath
}
