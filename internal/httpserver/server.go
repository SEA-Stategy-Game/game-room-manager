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
	"github.com/SEA-Stategy-Game/game-room-manager/internal/rooms"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
}

func zapRequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	r := chi.NewRouter()
	r.Use(zapRequestLogger(logger))

	dbPath := os.Getenv("DB_PATH") // Get path from environment
	if dbPath == "" {
		dbPath = "manager.db" // Default to current directory for local dev
	}
	logger.Info("initializing database", zap.String("path", dbPath))

	roomRepo, err := rooms.NewSQLiteRepository(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize room repository: %w", err)
	}
	roomSvc := rooms.NewService(roomRepo, cfg.GameImage)
	roomHandler := rooms.NewHandler(roomSvc, logger)
	r.Get("/rooms", roomHandler.GetRooms)
	r.Get("/room/{roomId}", roomHandler.GetRoom)
	r.Post("/rooms/{roomId}/players/{playerId}/join", roomHandler.JoinRoom)
	r.Post("/rooms/create", roomHandler.CreateGame)
	r.Post("/rooms/{roomId}/status", roomHandler.SetStatus)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("readiness check", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	addr := fmt.Sprintf(":%d", cfg.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		server: srv,
	}, nil
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
