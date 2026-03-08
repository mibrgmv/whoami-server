package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"whoami-server/libs/keycloak"
)

type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	UsernameKey      contextKey = "username"
	EmailKey         contextKey = "email"
	EmailVerifiedKey contextKey = "email_verified"
	RolesKey         contextKey = "roles"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type AuthFailureRecorder interface {
	RecordAuthFailure(reason string)
}

type JWTConfig struct {
	KeycloakBaseURL string
	Realm           string
	KeyRefreshTTL   time.Duration
	HTTPTimeout     time.Duration
	Metrics         AuthFailureRecorder
}

func JWT(cfg JWTConfig) gin.HandlerFunc {
	validator := &jwtValidator{
		config:     cfg,
		publicKeys: make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}

	return validator.handler
}

// JWTOptional creates a middleware that extracts JWT claims if present but doesn't fail if not.
// Use this for routes that work for both authenticated and unauthenticated users.
func JWTOptional(cfg JWTConfig) gin.HandlerFunc {
	validator := &jwtValidator{
		config:     cfg,
		publicKeys: make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}

	return validator.optionalHandler
}

type jwtValidator struct {
	config      JWTConfig
	publicKeys  map[string]*rsa.PublicKey
	keysMutex   sync.RWMutex
	lastRefresh time.Time
	httpClient  *http.Client
}

func (v *jwtValidator) recordAuthFailure(reason string) {
	if v.config.Metrics != nil {
		v.config.Metrics.RecordAuthFailure(reason)
	}
}

func (v *jwtValidator) handler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		v.recordAuthFailure("missing_token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		v.recordAuthFailure("invalid_format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
		c.Abort()
		return
	}

	claims, err := v.validateToken(tokenString)
	if err != nil {
		v.recordAuthFailure("invalid_token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
		c.Abort()
		return
	}

	v.setClaimsInContext(c, claims)
	c.Next()
}

// optionalHandler extracts JWT claims if present but continues even if not authenticated
func (v *jwtValidator) optionalHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// No token provided - continue as guest
		c.Next()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		// Invalid format - continue as guest
		c.Next()
		return
	}

	claims, err := v.validateToken(tokenString)
	if err != nil {
		// Invalid token - continue as guest
		c.Next()
		return
	}

	v.setClaimsInContext(c, claims)
	c.Next()
}

func (v *jwtValidator) setClaimsInContext(c *gin.Context, claims *keycloak.Claims) {
	c.Set("user_id", claims.Subject)
	c.Set("username", claims.PreferredUsername)
	c.Set("email", claims.Email)
	c.Set("email_verified", claims.EmailVerified)
	c.Set("claims", claims)

	roles := extractRoles(claims)
	rolesStr := strings.Join(roles, ",")

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, UserIDKey, claims.Subject)
	ctx = context.WithValue(ctx, UsernameKey, claims.PreferredUsername)
	ctx = context.WithValue(ctx, EmailKey, claims.Email)
	ctx = context.WithValue(ctx, EmailVerifiedKey, claims.EmailVerified)
	ctx = context.WithValue(ctx, RolesKey, rolesStr)

	c.Request = c.Request.WithContext(ctx)
}

func extractRoles(claims *keycloak.Claims) []string {
	if claims == nil || claims.RealmAccess == nil {
		return nil
	}

	rolesInterface, ok := claims.RealmAccess["roles"].([]interface{})
	if !ok {
		return nil
	}

	roles := make([]string, 0, len(rolesInterface))
	for _, r := range rolesInterface {
		if roleStr, ok := r.(string); ok {
			roles = append(roles, roleStr)
		}
	}
	return roles
}

func (v *jwtValidator) validateToken(tokenString string) (*keycloak.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &keycloak.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key ID in token header")
		}

		publicKey, err := v.getPublicKey(kid)
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

	if claims.Issuer != fmt.Sprintf("%s/realms/%s", v.config.KeycloakBaseURL, v.config.Realm) {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

func (v *jwtValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	v.keysMutex.RLock()
	if publicKey, exists := v.publicKeys[kid]; exists && time.Since(v.lastRefresh) < v.config.KeyRefreshTTL {
		v.keysMutex.RUnlock()
		return publicKey, nil
	}
	v.keysMutex.RUnlock()

	if err := v.refreshPublicKeys(); err != nil {
		return nil, fmt.Errorf("failed to refresh public keys: %w", err)
	}

	v.keysMutex.RLock()
	defer v.keysMutex.RUnlock()

	publicKey, exists := v.publicKeys[kid]
	if !exists {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}

	return publicKey, nil
}

func (v *jwtValidator) refreshPublicKeys() error {
	v.keysMutex.Lock()
	defer v.keysMutex.Unlock()

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", v.config.KeycloakBaseURL, v.config.Realm)

	resp, err := v.httpClient.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			publicKey, err := v.jwkToRSAPublicKey(key)
			if err != nil {
				continue
			}
			v.publicKeys[key.Kid] = publicKey
		}
	}

	v.lastRefresh = time.Now()
	return nil
}

func (v *jwtValidator) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
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
