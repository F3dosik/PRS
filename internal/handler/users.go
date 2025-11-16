package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/F3dosik/PRS.git/internal/models/api"
	"github.com/F3dosik/PRS.git/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SetIsActiveRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	IsActive bool      `json:"is_active"`
}

func HandlerSetIsActive(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setIsActive(w, r, storage, logger)
	}
}

func setIsActive(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	var req SetIsActiveRequest
	if err := DecodeJSON(r, &req); err != nil {
		logger.Warn("cannot decode json", zap.Error(err))
		RespondError(w, err)
		return
	}

	if req.UserID == uuid.Nil {
		logger.Warn("user_id is invalid")
		apiErr := api.NewAPIError(api.ErrInvalidUser, "user_id is required")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := storage.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		logger.Warn("cannot set isActive", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, api.UserResponse{User: *user})
}

func HandlerGetReview(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getReview(w, r, storage, logger)
	}
}

func getReview(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		logger.Warn("user_id is missing")
		RespondError(w, api.NewAPIError(api.ErrInvalidUser, "user_id is required"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Warn("invalid user_id format", zap.Error(err))
		apiErr := api.NewAPIError(api.ErrInvalidUser, "invalid user_id format")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	resp, err := storage.GetReview(ctx, userID)
	if err != nil {
		logger.Warn("cannot get review", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, resp)
}
