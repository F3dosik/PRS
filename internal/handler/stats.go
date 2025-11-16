package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/F3dosik/PRS.git/internal/repository"
	"go.uber.org/zap"
)

func HandlerStats(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats(w, r, storage, logger)
	}
}

func stats(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := storage.GetStats(ctx)
	if err != nil {
		logger.Warn("cannot get stats", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, stats)
}
