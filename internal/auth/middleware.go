package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
	"whoami-server-gateway/internal/auth/keycloak"
)

type KeycloakJWK struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type KeycloakJWKS struct {
	Keys []KeycloakJWK `json:"keys"`
}

// todo options
type Middleware struct {
	keycloakBaseURL string
	realm           string
	publicKeys      map[string]*rsa.PublicKey
	keysMutex       sync.RWMutex
	httpClient      *http.Client
	lastKeyRefresh  time.Time
	keyRefreshTTL   time.Duration
}

// todo pointer
func NewMiddleware(keycloakBaseURL, realm string) Middleware {
	return Middleware{
		keycloakBaseURL: keycloakBaseURL,
		realm:           realm,
		publicKeys:      make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keyRefreshTTL: 1 * time.Hour,
	}
}

// todo naming
func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			writeAuthError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := m.validateToken(tokenString)
		if err != nil {
			writeAuthError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.Subject)
		ctx = context.WithValue(ctx, "username", claims.PreferredUsername)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "claims", claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewJWTMiddleware(keycloakBaseURL, realm string) func(http.Handler) http.Handler {
	am := NewMiddleware(keycloakBaseURL, realm)
	return am.Middleware
}

func (m *Middleware) validateToken(tokenString string) (*keycloak.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &keycloak.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key ID in token header")
		}

		publicKey, err := m.getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*keycloak.Claims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	if claims.Issuer != fmt.Sprintf("%s/realms/%s", m.keycloakBaseURL, m.realm) {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

func (m *Middleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	m.keysMutex.RLock()
	if publicKey, exists := m.publicKeys[kid]; exists && time.Since(m.lastKeyRefresh) < m.keyRefreshTTL {
		m.keysMutex.RUnlock()
		return publicKey, nil
	}
	m.keysMutex.RUnlock()

	if err := m.refreshPublicKeys(); err != nil {
		return nil, fmt.Errorf("failed to refresh public keys: %w", err)
	}

	m.keysMutex.RLock()
	defer m.keysMutex.RUnlock()

	publicKey, exists := m.publicKeys[kid]
	if !exists {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}

	return publicKey, nil
}

func (m *Middleware) refreshPublicKeys() error {
	m.keysMutex.Lock()
	defer m.keysMutex.Unlock()

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", m.keycloakBaseURL, m.realm)

	resp, err := m.httpClient.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks KeycloakJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	m.publicKeys = make(map[string]*rsa.PublicKey)

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			publicKey, err := m.jwkToRSAPublicKey(key)
			if err != nil {
				continue
			}
			m.publicKeys[key.Kid] = publicKey
		}
	}

	m.lastKeyRefresh = time.Now()
	return nil
}

func (m *Middleware) jwkToRSAPublicKey(jwk KeycloakJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

func writeAuthError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":             "authentication_failed",
		"error_description": message,
	})
}

func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func GetUsername(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}

func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value("email").(string); ok {
		return email
	}
	return ""
}

func GetClaims(ctx context.Context) *keycloak.Claims {
	if claims, ok := ctx.Value("claims").(*keycloak.Claims); ok {
		return claims
	}
	return nil
}

func HasRole(ctx context.Context, role string) bool {
	claims := GetClaims(ctx)
	if claims == nil {
		return false
	}

	if realmAccess, ok := claims.RealmAccess["roles"].([]interface{}); ok {
		for _, r := range realmAccess {
			if roleStr, ok := r.(string); ok && roleStr == role {
				return true
			}
		}
	}

	return false
}

func GinJWTMiddleware(baseURL, realm string) gin.HandlerFunc {
	jwtMiddleware := NewJWTMiddleware(baseURL, realm)
	return gin.WrapH(jwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will be handled by the next handler in Gin
	})))
}
