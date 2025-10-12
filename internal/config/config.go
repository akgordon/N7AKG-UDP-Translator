package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Listen struct {
		Address string `yaml:"address" mapstructure:"address"`
		Port    int    `yaml:"port" mapstructure:"port"`
	} `yaml:"listen" mapstructure:"listen"`

	Target struct {
		Address string `yaml:"address" mapstructure:"address"`
		Port    int    `yaml:"port" mapstructure:"port"`
	} `yaml:"target" mapstructure:"target"`

	Verbose bool `yaml:"verbose" mapstructure:"verbose"`

	// Message formatting options
	Formatting struct {
		// Source format detection
		AutoDetect bool   `yaml:"auto_detect" mapstructure:"auto_detect"`
		SourceType string `yaml:"source_type" mapstructure:"source_type"` // e.g., "wsjt-x", "fldigi", "js8call"

		// N1MM formatting options
		N1MM struct {
			Station  string `yaml:"station" mapstructure:"station"`
			Operator string `yaml:"operator" mapstructure:"operator"`
			Contest  string `yaml:"contest" mapstructure:"contest"`
		} `yaml:"n1mm" mapstructure:"n1mm"`
	} `yaml:"formatting" mapstructure:"formatting"`
}

// Load loads the configuration from file or creates default configuration
func Load(configFile string) (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Listen.Address = "0.0.0.0"
	cfg.Listen.Port = 2333
	cfg.Target.Address = "127.0.0.1"
	cfg.Target.Port = 12060
	cfg.Verbose = false
	cfg.Formatting.AutoDetect = true
	cfg.Formatting.SourceType = "auto"
	cfg.Formatting.N1MM.Station = "UDP-RELAY"
	cfg.Formatting.N1MM.Operator = "OP"
	cfg.Formatting.N1MM.Contest = "GENERAL"

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Look for config in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil // Return defaults if can't find home
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".udp-logger-relay")
	}

	// Environment variable support
	viper.SetEnvPrefix("UDP_LOGGER")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return cfg, nil
}

// SaveDefault saves a default configuration file to the user's home directory
func SaveDefault() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to find home directory: %w", err)
	}

	configPath := filepath.Join(home, ".udp-logger-relay.yaml")

	defaultConfig := `# UDP Logger Relay Configuration
listen:
  address: "0.0.0.0"
  port: 2333

target:
  address: "127.0.0.1"
  port: 12060    # N1MM Logger Plus default UDP port

verbose: false

formatting:
  auto_detect: true
  source_type: "auto"  # Options: auto, wsjt-x, fldigi, js8call, etc.
  
  n1mm:
    station: "UDP-RELAY"
    operator: "OP"
    contest: "GENERAL"
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
