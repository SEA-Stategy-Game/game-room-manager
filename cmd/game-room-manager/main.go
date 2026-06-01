package main

import (
	"log"
	"os"

	"github.com/SEA-Stategy-Game/game-room-manager/internal/config"
	"github.com/SEA-Stategy-Game/game-room-manager/internal/httpserver"
	"github.com/SEA-Stategy-Game/game-room-manager/internal/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("failed to load config", zap.Error(err))
	}

	log, err := logger.New(cfg)
	if err != nil {
		zap.L().Fatal("failed to init logger", zap.Error(err))
	}
	defer log.Sync() //nolint:errcheck

	zap.ReplaceGlobals(log)

	log.Info("starting game-room-manager service",
		zap.Int("port", cfg.Port),
		zap.String("env", cfg.Env),
		zap.String("log_level", cfg.LogLevel),
	)

	srv := httpserver.New(cfg, log)

	if err := srv.Run(); err != nil {
		log.Error("server exited with error", zap.Error(err))
		os.Exit(1)
	}
}
