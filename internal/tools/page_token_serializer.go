package tools

import (
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
)

func CreatePageToken(userID uuid.UUID) string {
	encoded := base64.URLEncoding.EncodeToString([]byte(userID.String()))
	return encoded
}

func ParsePageToken(token string) (uuid.UUID, error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid page token: %w", err)
	}
	return uuid.Parse(string(decoded))
}
