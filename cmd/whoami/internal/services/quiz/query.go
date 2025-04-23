package quiz

import "github.com/google/uuid"

type Query struct {
	Ids       []uuid.UUID
	PageSize  int32
	PageToken string
}
