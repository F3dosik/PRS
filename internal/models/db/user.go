package db

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID
	Name string
	IsActive bool
	TeamID *uuid.UUID
	CreatedAt time.Time
}