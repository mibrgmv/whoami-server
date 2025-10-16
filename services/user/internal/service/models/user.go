package models

import userv1 "github.com/mibrgmv/whoami-server/user/internal/protogen/user/v1"

type User struct {
	ID            string
	Username      string
	Email         string
	FirstName     string
	LastName      string
	Enabled       bool
	EmailVerified bool
	CreatedAt     string
}

type UpdateUserData struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
}

func (u *User) ToProto() *userv1.User {
	return &userv1.User{
		Id:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Enabled:       u.Enabled,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}
