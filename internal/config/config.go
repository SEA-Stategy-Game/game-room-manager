package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Port     int
	Env      string
	LogLevel string
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	v.SetDefault("port", 8080)
	v.SetDefault("env", "local")
	v.SetDefault("log_level", "info")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	_ = v.ReadInConfig()

	cfg := &Config{
		Port:     v.GetInt("port"),
		Env:      v.GetString("env"),
		LogLevel: v.GetString("log_level"),
	}
	// Allow conventional PORT env var to override, but still validate.
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			cfg.Port = parsed
		} else {
			return nil, fmt.Errorf("invalid PORT env value: %q", p)
		}
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Port)
	}

	return cfg, nil
}

