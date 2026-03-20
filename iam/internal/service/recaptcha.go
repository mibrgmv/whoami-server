package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gordle/iam/internal/config"
)

type recaptchaVerifier struct {
	config     config.RecaptchaConfig
	httpClient *http.Client
}

func newRecaptchaVerifier(cfg config.RecaptchaConfig) *recaptchaVerifier {
	return &recaptchaVerifier{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (v *recaptchaVerifier) enabled() bool {
	return v.config.SecretKey != ""
}

type recaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func (v *recaptchaVerifier) verify(ctx context.Context, token string) error {
	if !v.enabled() {
		return nil
	}

	if token == "" {
		return ErrRecaptchaRequired
	}

	data := url.Values{
		"secret":   {v.config.SecretKey},
		"response": {token},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.config.VerifyURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create recaptcha request: %w", err)
	}
	req.URL.RawQuery = data.Encode()

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify recaptcha: %w", err)
	}
	defer resp.Body.Close()

	var result recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse recaptcha response: %w", err)
	}

	if !result.Success {
		return ErrRecaptchaFailed
	}

	return nil
}
