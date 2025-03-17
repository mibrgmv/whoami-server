package models

import (
	"time"
)

type User struct {
	ID        int64
	Name      string
	Password  string
	CreatedAt time.Time
	LastLogin time.Time
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
