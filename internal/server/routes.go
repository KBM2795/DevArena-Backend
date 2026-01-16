package server

import (
	"fmt"
	"net/http"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/handlers"
	"github.com/KBM2795/DevArena-Backend/internal/webhooks"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all application routes
func (s *Server) RegisterRoutes() {
	// Health check - outside of API versioning
	s.router.GET("/health", handlers.HealthHandler)

	// Webhook routes (no auth, but signature verified)
	s.registerWebhookRoutes()

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public routes (no auth required)
		s.registerPublicRoutes(v1)

		// Protected routes (auth required)
		protected := v1.Group("/")
		jwtMiddleware, _ := middleware.NewJWTMiddleware(s.config.Clerk.PEMPublicKey, s.config.Clerk.AuthorizedParties)
		protected.Use(jwtMiddleware.Authenticate())
		s.registerProtectedRoutes(protected)
	}
}

// registerWebhookRoutes registers webhook endpoints
func (s *Server) registerWebhookRoutes() {
	webhookHandler := webhooks.NewClerkWebhookHandler(s.db, s.config.Clerk.WebhookSigningSecret)

	// Clerk webhooks - POST /api/webhooks
	s.router.POST("/api/webhooks", webhookHandler.HandleWebhook)
}

// registerPublicRoutes registers routes that don't require authentication
func (s *Server) registerPublicRoutes(rg *gin.RouterGroup) {

	rg.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to DevArena API v1",
			"version": "1.0.0",
		})
	})

	// Auth routes
	// auth := rg.Group("/auth")
	// {
	// 	auth.POST("/register", s.handleRegister)
	// 	auth.POST("/login", s.handleLogin)
	// }

	// Challenge routes (public read)
	// challenges := rg.Group("/challenges")
	// {
	// 	challenges.GET("/", s.handleGetChallenges)
	// 	challenges.GET("/:id", s.handleGetChallenge)
	// }
}

// registerProtectedRoutes registers routes that require authentication
func (s *Server) registerProtectedRoutes(rg *gin.RouterGroup) {
	rg.GET("/protected", func(c *gin.Context) {
		userID, exists := middleware.GetUserID(c)
		fmt.Printf("Protected Route: Key exists: %v, UserID: %s\n", exists, userID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to DevArena API v1 protected route",
			"version": "1.0.0",
			"user_id": userID,
		})
	})
	// User routes
	// users := rg.Group("/users")
	// {
	// 	users.GET("/me", s.handleGetCurrentUser)
	// 	users.PUT("/me", s.handleUpdateCurrentUser)
	// }

	// // Challenge submission routes
	// submissions := rg.Group("/submissions")
	// {
	// 	submissions.POST("/", s.handleCreateSubmission)
	// 	submissions.GET("/", s.handleGetMySubmissions)
	// }
}
