package api

import "github.com/google/uuid"

type TeamMember struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsActive bool      `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamResponse struct {
	Team *Team `json:"team"`
}
