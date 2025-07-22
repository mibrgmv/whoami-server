package users

import (
	"whoami-server-gateway/internal/auth/keycloak/client"
	"whoami-server-gateway/protogen/golang/user"
)

type Handler struct {
	keycloak *client.Client
	user     *user.UserServiceClient
}

func NewHandler(keycloak *client.Client, user *user.UserServiceClient) *Handler {
	return &Handler{
		keycloak: keycloak,
		user:     user,
	}
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}
