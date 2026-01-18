package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ClerkWebhookHandler handles Clerk webhook events
type ClerkWebhookHandler struct {
	db            *db.Database
	signingSecret string
}

// NewClerkWebhookHandler creates a new webhook handler
func NewClerkWebhookHandler(database *db.Database, signingSecret string) *ClerkWebhookHandler {
	return &ClerkWebhookHandler{
		db:            database,
		signingSecret: signingSecret,
	}
}

// ClerkWebhookEvent represents a Clerk webhook payload
type ClerkWebhookEvent struct {
	Data   json.RawMessage `json:"data"`
	Object string          `json:"object"`
	Type   string          `json:"type"`
}

// ClerkUserData represents user data from Clerk webhook
type ClerkUserData struct {
	ID               string            `json:"id"`
	FirstName        string            `json:"first_name"`
	LastName         string            `json:"last_name"`
	Username         *string           `json:"username"`
	ImageURL         string            `json:"image_url"`
	EmailAddresses   []EmailAddress    `json:"email_addresses"`
	PrimaryEmailID   string            `json:"primary_email_address_id"`
	ExternalAccounts []ExternalAccount `json:"external_accounts"`
	CreatedAt        int64             `json:"created_at"`
	UpdatedAt        int64             `json:"updated_at"`
}

// EmailAddress represents an email in Clerk
type EmailAddress struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
}

// ExternalAccount represents linked OAuth accounts
type ExternalAccount struct {
	Provider string `json:"provider"`
	Username string `json:"username"`
}

// HandleWebhook processes incoming Clerk webhooks
func (h *ClerkWebhookHandler) HandleWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify the webhook signature
	if !h.verifySignature(c.Request.Header, body) {
		log.Printf("Webhook signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse the webhook event
	var event ClerkWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Printf("Received Clerk webhook: type=%s", event.Type)

	// Handle different event types
	switch event.Type {
	case "user.created":
		h.handleUserCreated(c, event.Data)
	case "user.updated":
		h.handleUserUpdated(c, event.Data)
	case "user.deleted":
		h.handleUserDeleted(c, event.Data)
	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
		c.JSON(http.StatusOK, gin.H{"message": "Event type not handled"})
	}
}

// verifySignature verifies the Svix webhook signature
func (h *ClerkWebhookHandler) verifySignature(headers http.Header, payload []byte) bool {
	// Get Svix headers
	svixID := headers.Get("svix-id")
	svixTimestamp := headers.Get("svix-timestamp")
	svixSignature := headers.Get("svix-signature")

	if svixID == "" || svixTimestamp == "" || svixSignature == "" {
		log.Printf("Missing Svix headers")
		return false
	}

	// Check timestamp to prevent replay attacks (5 minute tolerance)
	timestamp, err := parseUnixTimestamp(svixTimestamp)
	if err != nil {
		log.Printf("Invalid timestamp: %v", err)
		return false
	}
	if time.Since(timestamp) > 5*time.Minute {
		log.Printf("Webhook timestamp too old")
		return false
	}

	// Create the signed payload
	signedPayload := fmt.Sprintf("%s.%s.%s", svixID, svixTimestamp, string(payload))

	// The signing secret is prefixed with "whsec_"
	secret := strings.TrimPrefix(h.signingSecret, "whsec_")
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		log.Printf("Error decoding signing secret: %v", err)
		return false
	}

	// Calculate expected signature
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedPayload))
	expectedSig := mac.Sum(nil)
	expectedSigBase64 := base64.StdEncoding.EncodeToString(expectedSig)

	// Parse the signature versions (v1,signature format)
	signatures := strings.Split(svixSignature, " ")
	for _, sig := range signatures {
		parts := strings.Split(sig, ",")
		if len(parts) == 2 && parts[0] == "v1" {
			if parts[1] == expectedSigBase64 {
				return true
			}
		}
	}

	return false
}

// handleUserCreated handles user.created events
func (h *ClerkWebhookHandler) handleUserCreated(c *gin.Context, data json.RawMessage) {
	var userData ClerkUserData
	if err := json.Unmarshal(data, &userData); err != nil {
		log.Printf("Error parsing user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Extract primary email
	primaryEmail := ""
	for _, email := range userData.EmailAddresses {
		if email.ID == userData.PrimaryEmailID {
			primaryEmail = email.EmailAddress
			break
		}
	}

	// Extract GitHub username if connected
	githubUsername := ""
	for _, account := range userData.ExternalAccounts {
		if account.Provider == "oauth_github" {
			githubUsername = account.Username
			break
		}
	}

	// Build display name
	displayName := strings.TrimSpace(fmt.Sprintf("%s %s", userData.FirstName, userData.LastName))
	if displayName == "" {
		displayName = primaryEmail
	}

	// Get username - use pointer to handle NULL properly
	var username *string
	if userData.Username != nil && *userData.Username != "" {
		username = userData.Username
	}

	// Get github username as pointer (NULL if empty)
	var githubUsernamePtr *string
	if githubUsername != "" {
		githubUsernamePtr = &githubUsername
	}

	// Insert user into database (or update if created just-in-time during onboarding)
	// Use NULLIF to convert empty strings to NULL for UNIQUE constraint compatibility
	userID := uuid.New().String()
	query := `
		INSERT INTO users (id, clerk_user_id, email, username, display_name, avatar_url, github_username, github_connected, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (clerk_user_id) DO UPDATE SET
			email = EXCLUDED.email,
			username = COALESCE(NULLIF(EXCLUDED.username, ''), users.username),
			display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
			avatar_url = COALESCE(NULLIF(EXCLUDED.avatar_url, ''), users.avatar_url),
			github_username = COALESCE(NULLIF(EXCLUDED.github_username, ''), users.github_username),
			github_connected = EXCLUDED.github_connected,
			updated_at = NOW()
	`

	_, err := h.db.Pool.Exec(context.Background(), query,
		userID,
		userData.ID,
		primaryEmail,
		username,
		displayName,
		userData.ImageURL,
		githubUsernamePtr,
		githubUsername != "",
	)

	if err != nil {
		log.Printf("Error inserting/updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	log.Printf("Created/updated user: clerk_id=%s, email=%s", userData.ID, primaryEmail)
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// handleUserUpdated handles user.updated events
func (h *ClerkWebhookHandler) handleUserUpdated(c *gin.Context, data json.RawMessage) {
	var userData ClerkUserData
	if err := json.Unmarshal(data, &userData); err != nil {
		log.Printf("Error parsing user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Extract primary email
	primaryEmail := ""
	for _, email := range userData.EmailAddresses {
		if email.ID == userData.PrimaryEmailID {
			primaryEmail = email.EmailAddress
			break
		}
	}

	// Extract GitHub username if connected
	githubUsername := ""
	for _, account := range userData.ExternalAccounts {
		if account.Provider == "oauth_github" {
			githubUsername = account.Username
			break
		}
	}

	// Build display name
	displayName := strings.TrimSpace(fmt.Sprintf("%s %s", userData.FirstName, userData.LastName))

	// Get username - use pointer to handle NULL properly
	var username *string
	if userData.Username != nil && *userData.Username != "" {
		username = userData.Username
	}

	// Get github username as pointer (NULL if empty)
	var githubUsernamePtr *string
	if githubUsername != "" {
		githubUsernamePtr = &githubUsername
	}

	// Update user in database - use NULLIF to avoid empty string unique constraint issues
	query := `
		UPDATE users 
		SET email = $2, 
			username = COALESCE(NULLIF($3, ''), username), 
			display_name = COALESCE(NULLIF($4, ''), display_name), 
			avatar_url = COALESCE(NULLIF($5, ''), avatar_url), 
			github_username = COALESCE(NULLIF($6, ''), github_username), 
			github_connected = $7,
			updated_at = NOW()
		WHERE clerk_user_id = $1
	`

	result, err := h.db.Pool.Exec(context.Background(), query,
		userData.ID,
		primaryEmail,
		username,
		displayName,
		userData.ImageURL,
		githubUsernamePtr,
		githubUsername != "",
	)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	rowsAffected := result.RowsAffected()
	log.Printf("Updated user: clerk_id=%s, rows_affected=%d", userData.ID, rowsAffected)
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// handleUserDeleted handles user.deleted events
func (h *ClerkWebhookHandler) handleUserDeleted(c *gin.Context, data json.RawMessage) {
	// For deleted events, data only contains the ID
	var deletedData struct {
		ID      string `json:"id"`
		Deleted bool   `json:"deleted"`
	}
	if err := json.Unmarshal(data, &deletedData); err != nil {
		log.Printf("Error parsing deleted user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Delete user from database (or soft delete)
	query := `DELETE FROM users WHERE clerk_user_id = $1`
	result, err := h.db.Pool.Exec(context.Background(), query, deletedData.ID)

	if err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected := result.RowsAffected()
	log.Printf("Deleted user: clerk_id=%s, rows_affected=%d", deletedData.ID, rowsAffected)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Helper function to parse Unix timestamp
func parseUnixTimestamp(s string) (time.Time, error) {
	var ts int64
	_, err := fmt.Sscanf(s, "%d", &ts)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}
