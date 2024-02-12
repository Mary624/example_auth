package storage

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("user not found")
)

type User struct {
	Guid     string    `json:"guid"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Key          string    `json:"key"`
}
