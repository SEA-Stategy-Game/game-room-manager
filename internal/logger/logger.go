package logger

import (
	"strings"

	"github.com/SEA-Stategy-Game/game-room-manager/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	var cfgZap zap.Config
	if cfg.Env == "local" || cfg.Env == "dev" {
		cfgZap = zap.NewDevelopmentConfig()
	} else {
		cfgZap = zap.NewProductionConfig()
	}

	if cfg.LogLevel != "" {
		level := new(zapcore.Level)
		if err := level.UnmarshalText([]byte(strings.ToLower(cfg.LogLevel))); err == nil {
			cfgZap.Level = zap.NewAtomicLevelAt(*level)
		}
	}

	logger, err := cfgZap.Build()
	if err != nil {
		return nil, err
	}

	return logger.With(zap.String("service", "game-room-manager"), zap.String("env", cfg.Env)), nil
}
