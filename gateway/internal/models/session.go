package models

import "time"

type SessionData struct {
	AccessToken   string    `json:"access_token"`
	IDToken       string    `json:"id_token"`
	RefreshToken  string    `json:"refresh_token"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Roles         string    `json:"roles"`
	ExpiresAt     time.Time `json:"expires_at"`
}
