package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Github    Github    `mapstructure:"github"`
	Shortcuts Shortcuts `mapstructure:"keys"`
}

type Github struct {
	Token string `mapstructure:"token"`
}

type Shortcuts struct {
	SwitchTabRight string `mapstructure:"switch_tab_right"`
	SwitchTabLeft  string `mapstructure:"switch_tab_left"`
	Quit           string `mapstructure:"quit"`
	Refresh        string `mapstructure:"refresh"`
	Enter          string `mapstructure:"enter"`
	Tab            string `mapstructure:"tab"`
}

func LoadConfig() (*Config, error) {
	var config = new(Config)
	defer func() {
		config = fillDefaultShortcuts(config)
	}()

	setConfig()

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.BindEnv("github.token", "GITHUB_TOKEN")
	viper.AutomaticEnv()

	// Read the config file first
	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
		return config, nil
	}

	// If config file is not found, try to unmarshal from environment variables
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func setConfig() {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "gama", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "gama"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		return
	}

	oldConfigPath := filepath.Join(os.Getenv("HOME"), ".gama.yaml")
	if _, err := os.Stat(oldConfigPath); err == nil {
		viper.AddConfigPath(os.Getenv("HOME"))
		viper.SetConfigName(".gama")
		viper.SetConfigType("yaml")
		return
	}
}
