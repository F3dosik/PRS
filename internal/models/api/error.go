package api

type ErrorCode string

const (
	ErrTeamExist        ErrorCode = "TEAM_EXISTS"
	ErrPRExist          ErrorCode = "PR_EXISTS"
	ErrPRMerged         ErrorCode = "PR_MERGED"
	ErrNotAssigned      ErrorCode = "NOT_ASSIGNED"
	ErrNoCandidate      ErrorCode = "NO_CANDIDATE"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrInvalidJSON      ErrorCode = "INVALID_JSON"
	ErrInvalidTeam      ErrorCode = "INVALID_TEAM"
	ErrInvalidParameter ErrorCode = "INVALID_PARAMETER"
	ErrInvalidUser      ErrorCode = "INVALID_USER"

	ErrInvalidPR ErrorCode = "INVALID_PULL_REQUEST"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func (e *APIError) Error() string {
	return string(e.Code) + ": " + e.Message
}

type ErrorResponse struct {
	Err APIError `json:"error"`
}

func NewErrorResponse(apiErr APIError) *ErrorResponse {
	return &ErrorResponse{Err: apiErr}
}
