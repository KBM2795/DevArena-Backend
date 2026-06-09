package server

import (
	"fmt"
	"net/http"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/handlers"
	mw "github.com/KBM2795/DevArena-Backend/internal/middleware"
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
	// Create handlers with database dependency
	h := handlers.NewHandlers(s.db)

	rg.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to DevArena API v1",
			"version": "1.0.0",
		})
	})

	// Challenge routes (public - no auth required)
	rg.GET("/challenges", h.GetChallengesHandler)
	rg.GET("/challenges/:id", h.GetChallengeByIDHandler)

	// Submissions for a challenge - anyone can browse
	rg.GET("/challenges/:id/submissions", h.ChallengeSubmissionsHandler)

	// Open showcase - anyone can browse
	rg.GET("/showcase", h.GetOpenShowcasesHandler)

	// Single submission detail (public, like count included)
	rg.GET("/submissions/:id", h.SubmissionStatusHandler)

	// Like status for a submission (no auth needed to read)
	rg.GET("/submissions/:id/like", h.GetLikeStatusHandler)

	// Comments - public read
	rg.GET("/comments", h.GetCommentsHandler)

	// Tags route (for filter dropdowns)
	rg.GET("/tags", h.GetTagsHandler)

	// Leaderboard route (public)
	rg.GET("/leaderboard", h.LeaderboardHandler)
}

// registerProtectedRoutes registers routes that require authentication
func (s *Server) registerProtectedRoutes(rg *gin.RouterGroup) {
	// Create handlers with database dependency
	h := handlers.NewHandlers(s.db)

	rg.GET("/protected", func(c *gin.Context) {
		userID, exists := middleware.GetUserID(c)
		fmt.Printf("Protected Route: Key exists: %v, UserID: %s\n", exists, userID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to DevArena API v1 protected route",
			"version": "1.0.0",
			"user_id": userID,
		})
	})

	// Onboarding routes
	rg.POST("/onboarding", h.OnboardingHandler)

	// Profile routes
	rg.GET("/profile", h.ProfileHandler)

	// Starter pack route (user-specific challenges)
	rg.GET("/me/starter-pack", h.GetStarterPackHandler)

	// Activity route for contribution graph
	rg.GET("/me/activity", h.ActivityHandler)

	// Stats routes
	rg.GET("/me/stats", h.StatsHandler)
	rg.GET("/me/tech-focus", h.TechFocusHandler)

	rg.GET("/me/submissions", h.RecentSubmissionsHandler)
	rg.GET("/me/showcase", h.GetUserShowcaseHandler)
	rg.GET("/me/challenges/:id/submissions", h.ChallengeSubmissionsHandler)

	rg.POST("/me/submissions", mw.StrictRateLimiter(5, 10), h.SubmitHandler)
	rg.POST("/me/submissions/upload-video", h.UploadVideoHandler)

	// Single submission detail (auth version – needed for like status of logged-in user)
	rg.GET("/me/submissions/:id", h.SubmissionStatusHandler)

	// Like / unlike a submission (toggle)
	rg.POST("/submissions/:id/like", h.ToggleLikeHandler)

	// Post a comment (challenge or project level)
	rg.POST("/comments", h.PostCommentHandler)

	// Notification routes
	rg.GET("/me/notifications", h.GetNotifications)
	rg.POST("/me/notifications/:id/read", h.MarkNotificationRead)
	rg.POST("/me/notifications/read-all", h.MarkAllNotificationsRead)
	rg.DELETE("/me/notifications/:id", h.DeleteNotification)
}
