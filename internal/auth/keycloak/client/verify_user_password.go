package client

import (
	"context"
	"fmt"
)

func (c *Client) VerifyUserPassword(ctx context.Context, userID, password string) error {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	_, err = c.ExchangeCredentialsForTokens(ctx, user.Username, password)
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}
