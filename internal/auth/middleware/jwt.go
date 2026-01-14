package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ClerkClaims represents the claims in a Clerk JWT
type ClerkClaims struct {
	jwt.RegisteredClaims
	SessionID       string `json:"sid"`
	AuthorizedParty string `json:"azp,omitempty"`
	Email           string `json:"email,omitempty"`
	Status          string `json:"sts,omitempty"` // For org membership status
}

// ContextKey type for context values
type ContextKey string

const (
	// UserIDKey is the context key for user ID (from sub claim)
	UserIDKey ContextKey = "user_id"
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session_id"
	// ClaimsKey is the context key for full claims
	ClaimsKey ContextKey = "claims"
)

// JWTMiddleware handles Clerk JWT verification
type JWTMiddleware struct {
	publicKey         *rsa.PublicKey
	authorizedParties []string
}

// NewJWTMiddleware creates a new JWT middleware instance
// pemPublicKey: Your Clerk PEM public key (from CLERK_PEM_PUBLIC_KEY env var)
// authorizedParties: List of permitted origins (e.g., ["http://localhost:3000", "https://devarena.dev"])
func NewJWTMiddleware(pemPublicKey string, authorizedParties []string) (*JWTMiddleware, error) {
	// Parse the PEM public key
	block, _ := pem.Decode([]byte(pemPublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return &JWTMiddleware{
		publicKey:         rsaPublicKey,
		authorizedParties: authorizedParties,
	}, nil
}

// Authenticate returns a Gin middleware function for JWT authentication
func (m *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Authorization header first (cross-origin)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		// If no Authorization header, try __session cookie (same-origin)
		if tokenString == "" {
			tokenString, _ = c.Cookie("__session")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No session token found. Provide Authorization header or __session cookie.",
			})
			return
		}

		// Parse and validate the token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("Invalid token: %v", err),
			})
			return
		}

		fmt.Println("JWT Middleware: User ID: ", claims.Subject)
		// Set claims in context
		c.Set(string(UserIDKey), claims.Subject) // sub claim contains user ID
		c.Set(string(SessionIDKey), claims.SessionID)
		c.Set(string(ClaimsKey), claims)

		c.Next()
	}
}

// validateToken validates the JWT token using Clerk's PEM public key
func (m *JWTMiddleware) validateToken(tokenString string) (*ClerkClaims, error) {
	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenString, &ClerkClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*ClerkClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	// Validate expiration and not-before claims
	now := time.Now()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		return nil, fmt.Errorf("token is expired")
	}
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return nil, fmt.Errorf("token is not yet valid")
	}

	// Validate authorized party (azp) claim if present and if we have authorized parties configured
	if claims.AuthorizedParty != "" && len(m.authorizedParties) > 0 {
		isAuthorized := false
		for _, party := range m.authorizedParties {
			if claims.AuthorizedParty == party {
				isAuthorized = true
				break
			}
		}
		if !isAuthorized {
			return nil, fmt.Errorf("unauthorized party: %s", claims.AuthorizedParty)
		}
	}

	// Optional: Check for pending organization status
	if claims.Status == "pending" {
		return nil, fmt.Errorf("user organization membership is pending")
	}

	return claims, nil
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetSessionID extracts the session ID from the Gin context
func GetSessionID(c *gin.Context) (string, bool) {
	sessionID, exists := c.Get(string(SessionIDKey))
	if !exists {
		return "", false
	}
	return sessionID.(string), true
}

// GetClaims extracts the full claims from the Gin context
func GetClaims(c *gin.Context) (*ClerkClaims, bool) {
	claims, exists := c.Get(string(ClaimsKey))
	if !exists {
		return nil, false
	}
	return claims.(*ClerkClaims), true
}
