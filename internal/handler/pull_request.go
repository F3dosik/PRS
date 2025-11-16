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

func HandlerPullRequestCreate(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pullRequestCreate(w, r, storage, logger)
	}
}

func pullRequestCreate(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	var pr api.PullRequestShort
	if err := DecodeJSON(r, &pr); err != nil {
		logger.Warn("cannot decode JSON", zap.Error(err))
		RespondError(w, err)
		return
	}

	if pr.PullRequestID == uuid.Nil || pr.PullRequestName == "" || pr.AuthorID == uuid.Nil {
		logger.Warn("pull request is invalid")
		apiErr := api.NewAPIError(api.ErrInvalidPR, "invalid pull request")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	pullRequest, err := storage.PullRequestCreate(ctx, pr.PullRequestID, pr.AuthorID, pr.PullRequestName)
	if err != nil {
		logger.Warn("cannot create pull request", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 201 response")
	RespondJSON(w, http.StatusCreated, api.PullRequestResponse{PullRequest: *pullRequest})
}

type mergeRequest struct {
	PullRequestID uuid.UUID `json:"pull_request_id"`
}

func HandlerPullRequestMerge(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pullRequestMerge(w, r, storage, logger)
	}
}

func pullRequestMerge(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	var req mergeRequest
	if err := DecodeJSON(r, &req); err != nil {
		logger.Warn("invalid JSON", zap.Error(err))
		RespondError(w, err)
		return
	}

	if req.PullRequestID == uuid.Nil {
		apiErr := api.NewAPIError(api.ErrInvalidPR, "pull_request_id is required")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	pr, err := storage.PullRequestMerge(ctx, req.PullRequestID)
	if err != nil {
		logger.Warn("cannot pull request merge", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, api.PullRequestResponse{PullRequest: *pr})
}

type reassignRequest struct {
	PullRequestID uuid.UUID `json:"pull_request_id"`
	OldUserID     uuid.UUID `json:"old_user_id"`
}

func HandlerPullRequestReassign(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pullRequestReassign(w, r, storage, logger)
	}
}

func pullRequestReassign(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	var req reassignRequest
	if err := DecodeJSON(r, &req); err != nil {
		logger.Warn("invalid JSON", zap.Error(err))
		RespondError(w, err)
		return
	}

	if req.PullRequestID == uuid.Nil {
		apiErr := api.NewAPIError(api.ErrInvalidPR, "pull request is required")
		RespondError(w, apiErr)
		return
	}
	if req.OldUserID == uuid.Nil {
		apiErr := api.NewAPIError(api.ErrInvalidUser, "user id is required")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	prReassignResponse, err := storage.PullRequestReassign(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		logger.Warn("cannot pull requestreassign", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, prReassignResponse)
}
