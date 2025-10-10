package models

import userpb "github.com/mibrgmv/whoami-server/users/internal/protogen/user"

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

func (u *User) ToProto() *userpb.User {
	return &userpb.User{
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
