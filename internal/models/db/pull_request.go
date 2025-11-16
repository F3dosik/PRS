package db

import (
	"time"

	"github.com/google/uuid"
)

type PrStatus string

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

type PullRequest struct {
	ID                uuid.UUID
	Title             string
	AuthorID          uuid.UUID // Не указатель, чтобы всегда иметь автора
	Status            PrStatus
	Reviewer1ID       *uuid.UUID
	Reviewer2ID       *uuid.UUID
	NeedMoreReviewers bool
	CreatedAt         time.Time
	MergedAt          time.Time
}
