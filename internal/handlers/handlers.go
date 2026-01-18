package handlers

import (
	"net/http"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/gin-gonic/gin"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	DB *db.Database
}

// NewHandlers creates a new Handlers instance
func NewHandlers(database *db.Database) *Handlers {
	return &Handlers{DB: database}
}

// HealthHandler checks database connection health
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Server is running"})
}

// OnboardingHandler handles onboarding data
func (h *Handlers) OnboardingHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var onboardingData struct {
		Experience   string   `json:"experience"`
		Paths        []string `json:"paths"`
		Technologies []string `json:"technologies"`
	}


	if err := c.ShouldBindJSON(&onboardingData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Save onboarding data to database
	if err := h.DB.SaveOnboardingData(userID, onboardingData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save onboarding data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Onboarding data saved successfully"})
}
