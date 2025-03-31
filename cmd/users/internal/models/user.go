package models

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	pb "whoami-server/protogen/golang/user"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Password  string
	CreatedAt time.Time
	LastLogin time.Time
}

func (user *User) ToProto() *pb.User {
	return &pb.User{
		UserId:    user.ID.String(),
		Username:  user.Name,
		Password:  user.Password,
		CreatedAt: timestamppb.New(user.CreatedAt),
		LastLogin: timestamppb.New(user.LastLogin),
	}
}
