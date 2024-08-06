package config

import "time"

func fillDefaultSettings(cfg *Config) *Config {
	if cfg.Settings.LiveMode.Interval == time.Duration(0) {
		cfg.Settings.LiveMode.Interval = 15 * time.Second
	}

	return cfg
}
