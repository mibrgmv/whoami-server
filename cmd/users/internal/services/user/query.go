package user

import "github.com/google/uuid"

type Query struct {
	UserIDs   []uuid.UUID
	Username  *string
	PageSize  int32
	PageToken string
}
