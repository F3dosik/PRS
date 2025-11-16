package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/F3dosik/PRS.git/internal/models/api"
	"github.com/F3dosik/PRS.git/internal/repository"
	"go.uber.org/zap"
)

func HandleTeamAdd(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamAdd(w, r, storage, logger)
	}
}

func teamAdd(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	var team api.Team
	if err := DecodeJSON(r, &team); err != nil {
		logger.Warn("cannot decode team JSON", zap.Error(err))
		RespondError(w, err)
		return
	}

	if team.TeamName == "" || len(team.Members) == 0 {
		logger.Warn("team name or members are invalid")
		apiErr := api.NewAPIError(api.ErrInvalidTeam, "team name or members are invalid")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := storage.UpdateTeam(ctx, &team); err != nil {
		logger.Warn("cannot update team", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 201 response")
	RespondJSON(w, http.StatusCreated, api.TeamResponse{Team: &team})
}

func HandleTeamGet(storage *repository.Storage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamGet(w, r, storage, logger)
	}
}

func teamGet(w http.ResponseWriter, r *http.Request, storage *repository.Storage, logger *zap.SugaredLogger) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		apiErr := api.NewAPIError(api.ErrInvalidParameter, "team_name query parameter is required")
		RespondError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	team, err := storage.GetTeam(ctx, teamName)
	if err != nil {
		logger.Warn("cannot get team", zap.Error(err))
		RespondError(w, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondJSON(w, http.StatusOK, api.TeamResponse{Team: team})
}
