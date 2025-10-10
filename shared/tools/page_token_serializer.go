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

func ParsePageToken(token string) (string, error) {
	if len(token) == 0 {
		return "", nil
	}
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid page token: %w", err)
	}
	decodedToken, err := uuid.Parse(string(decoded))
	if err != nil {
		return "", fmt.Errorf("could not parse uuid: %w", err)
	}
	return decodedToken.String(), nil
}
