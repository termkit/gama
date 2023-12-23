package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	configName = ".gama"
	configType = "yaml"
)

type Config struct {
	Github Github `mapstructure:"github"`
}

type Github struct {
	Token string `mapstructure:"token"`
}

func LoadConfig() (*Config, error) {
	configPath, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(".gama")
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.BindEnv("github.token", "GITHUB_TOKEN")
	viper.AutomaticEnv()

	// Read the config file first
	if err := viper.ReadInConfig(); err == nil {
		config := new(Config)
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
		return config, nil
	}

	// If config file is not found, try to unmarshal from environment variables
	config := new(Config)
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return config, nil
}

func CheckConfig() error {
	configPath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configFile := fmt.Sprintf("%s/%s.%s", configPath, configName, configType)

	file, err := os.Stat(configFile)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if file != nil {
		return nil
	}
	return fmt.Errorf("config file does not exist")
}

// SaveConfig saves the configuration to the config file.
func SaveConfig(config *Config) error {
	configPath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	viper.Set("github.token", config.Github.Token)

	if err := viper.SafeWriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
