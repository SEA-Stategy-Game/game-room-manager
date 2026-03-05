package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SEA-Stategy-Game/game-room-manager/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("readiness check", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	addr := fmt.Sprintf(":%d", cfg.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		server: srv,
	}
}

func (s *Server) Run() error {
	s.logger.Info("starting HTTP server", zap.String("addr", s.server.Addr))

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		s.logger.Info("received shutdown signal", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("server shutdown error", zap.Error(err))
			return err
		}
		s.logger.Info("server shut down gracefully")
		return nil
	case err := <-errCh:
		s.logger.Error("server error", zap.Error(err))
		return err
	}
}

