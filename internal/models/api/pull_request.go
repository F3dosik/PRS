package api

import (
	"time"

	"github.com/google/uuid"
)

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     uuid.UUID   `json:"pull_request_id"`
	PullRequestName   string      `json:"pull_request_name"`
	AuthorID          uuid.UUID   `json:"author_id"`
	Status            PRStatus    `json:"status"`
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
	CreatedAt         time.Time   `json:"createdAt,omitempty"`
	MergedAt          *time.Time  `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   uuid.UUID `json:"pull_request_id"`
	PullRequestName string    `json:"pull_request_name"`
	AuthorID        uuid.UUID `json:"author_id"`
	Status          PRStatus  `json:"status"`
}

type PullRequestResponse struct {
	PullRequest PullRequest `json:"pr"`
}

type PullRequestReassignResponse struct {
	PullRequest PullRequest `json:"pr"`
	ReplacedBy  uuid.UUID   `json:"replaced_by"`
}
