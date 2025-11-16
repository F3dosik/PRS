package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/F3dosik/PRS.git/internal/config/server"
	"github.com/F3dosik/PRS.git/internal/handler"
	"github.com/F3dosik/PRS.git/internal/middleware"
	"github.com/F3dosik/PRS.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	config  *cfg.ServerConfig
	storage *repository.Storage
	router  chi.Router
	logger  *zap.SugaredLogger
}

func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) (*Server, error) {
	storage, err := repository.NewStorage(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}
	r := chi.NewRouter()

	server := &Server{
		config:  cfg,
		storage: storage,
		router:  r,
		logger:  logger,
	}
	server.routes()

	return server, nil
}

func (s *Server) routes() {
	s.router.Use(middleware.WithLogging(s.logger))

	s.router.Route("/team", func(r chi.Router) {
		r.Post("/add", handler.HandleTeamAdd(s.storage, s.logger))
		r.Get("/get", handler.HandleTeamGet(s.storage, s.logger))
	})

	s.router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", handler.HandlerSetIsActive(s.storage, s.logger))
		r.Get("/getReview", handler.HandlerGetReview(s.storage, s.logger))
	})

	s.router.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", handler.HandlerPullRequestCreate(s.storage, s.logger))
		r.Post("/merge", handler.HandlerPullRequestMerge(s.storage, s.logger))
		r.Post("/reassign", handler.HandlerPullRequestReassign(s.storage, s.logger))
	})

	s.router.Get("/stats", handler.HandlerStats(s.storage, s.logger))

}

func (s *Server) Run() error {
	s.logger.Infow("starting server",
		"port", s.config.Port,
		"log_mode", s.config.LogMode,
	)

	srv := &http.Server{
		Addr:    s.config.Port,
		Handler: s.router,
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		s.logger.Infow("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Errorw("graceful shutdown failed", "err", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server listen failed: %w", err)
	}

	s.logger.Infow("server stopped")
	return nil
}
