package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/F3dosik/PRS.git/internal/models/api"
)

const (
	ContentTypePlainText = "text/plain; charset=utf-8"
	ContentTypeJSON      = "application/json; charset=utf-8"
	ErrInternal          = "INTERNAL_ERROR"
)

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // При лишних полях возвращает ошибку
	if err := dec.Decode(&dst); err != nil {
		return api.NewAPIError(api.ErrInvalidJSON, "cannot decode request body")
	}

	return nil
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, err error) {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		var status int
		switch apiErr.Code {
		case api.ErrTeamExist, api.ErrInvalidJSON, api.ErrInvalidParameter,
			api.ErrInvalidUser, api.ErrInvalidPR:
			status = http.StatusBadRequest
		case api.ErrNotFound:
			status = http.StatusNotFound
		case api.ErrPRExist, api.ErrPRMerged, api.ErrNotAssigned, api.ErrNoCandidate:
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
		}
		RespondJSON(w, status, api.NewErrorResponse(*apiErr))
		return
	}

	internalErr := api.NewAPIError(ErrInternal, "internal server error")
	RespondJSON(w, http.StatusInternalServerError, api.NewErrorResponse(*internalErr))
}
