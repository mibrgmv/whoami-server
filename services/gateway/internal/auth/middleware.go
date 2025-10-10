package auth

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
	"github.com/mibrgmv/whoami-server/services/gateway/internal/auth/keycloak"
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

// todo options
type Middleware struct {
	keycloakBaseURL string
	realm           string
	publicKeys      map[string]*rsa.PublicKey
	keysMutex       sync.RWMutex
	lastKeyRefresh  time.Time
	keyRefreshTTL   time.Duration
	httpClient      *http.Client
}

func NewMiddleware(keycloakBaseURL, realm string) *Middleware {
	return &Middleware{
		keycloakBaseURL: keycloakBaseURL,
		realm:           realm,
		publicKeys:      make(map[string]*rsa.PublicKey),
		keyRefreshTTL:   1 * time.Hour, // todo config
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (m *Middleware) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := m.validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("username", claims.PreferredUsername)
		c.Set("email", claims.Email)
		c.Set("email_verified", claims.EmailVerified)
		c.Set("claims", claims)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", claims.Subject)
		ctx = context.WithValue(ctx, "username", claims.PreferredUsername)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "email_verified", claims.EmailVerified)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
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

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

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

func (m *Middleware) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
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

func (m *Middleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Claims not found"})
			c.Abort()
			return
		}

		keycloakClaims, ok := claims.(*keycloak.Claims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		if realmAccess, ok := keycloakClaims.RealmAccess["roles"].([]interface{}); ok {
			for _, r := range realmAccess {
				if roleStr, ok := r.(string); ok && roleStr == role {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
