package main

import (
	"log"

	cfg "github.com/F3dosik/PRS.git/internal/config/server"
	"github.com/F3dosik/PRS.git/internal/server"
	"github.com/F3dosik/PRS.git/pkg/logger"
)

func main() {
	cfg, err := cfg.LoadServerConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}
	mode := logger.Mode(cfg.LogMode)
	baseLogger, sugarLogger := logger.NewLogger(mode)
	defer func() { _ = baseLogger.Sync() }()
	server, err := server.NewServer(cfg, sugarLogger)
	if err != nil {
		sugarLogger.Fatalw("failed to init server", "err", err)
	}
	if err := server.Run(); err != nil {
		sugarLogger.Fatalw("server stopped with error", "err", err)
	}
}
