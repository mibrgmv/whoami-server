package keycloak

type LoginResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// todo
type CreateUserRequest struct {
	Username      string                 `json:"username"`
	Email         string                 `json:"email"`
	FirstName     string                 `json:"firstName,omitempty"`
	LastName      string                 `json:"lastName,omitempty"`
	Enabled       bool                   `json:"enabled"`
	EmailVerified bool                   `json:"emailVerified"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	Credentials   []UserCredential       `json:"credentials,omitempty"`
}

type UserCredential struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

type CreateUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type DeleteUserResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
