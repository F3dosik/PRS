package api

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	TeamName *string    `json:"team_name"`
	IsActive bool      `json:"is_active"`
}

type UserResponse struct {
	User User `json:"user"`
}

type GetReviewResponse struct {
	UserID uuid.UUID `json:"user_id"`
	PullRequests []PullRequestShort
}