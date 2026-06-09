package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

// ProfileHandler returns the authenticated user's profile
func (h *Handlers) ProfileHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profile, err := h.DB.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetTagsHandler returns all available tags grouped by category
func (h *Handlers) GetTagsHandler(c *gin.Context) {
	tags, err := h.DB.GetTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

// GetChallengesHandler returns paginated challenges with filters
func (h *Handlers) GetChallengesHandler(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	difficulty := c.Query("difficulty")
	challengeType := c.Query("type")
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "newest")
	tagsStr := c.Query("tags")

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	params := db.ChallengeQueryParams{
		Page:       page,
		Limit:      limit,
		Difficulty: difficulty,
		Type:       challengeType,
		Search:     search,
		Sort:       sort,
		Tags:       tags,
	}

	result, err := h.DB.GetChallenges(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch challenges"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetChallengeByIDHandler returns a single challenge by ID
func (h *Handlers) GetChallengeByIDHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Challenge ID required"})
		return
	}

	challenge, err := h.DB.GetChallengeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"challenge": challenge,
	})
}

// GetStarterPackHandler returns the authenticated user's starter pack challenges
func (h *Handlers) GetStarterPackHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	challenges, err := h.DB.GetStarterPackChallenges(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch starter pack"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  challenges,
		"total": len(challenges),
	})
}

// LeaderboardHandler returns the top ranked users (public endpoint)
func (h *Handlers) LeaderboardHandler(c *gin.Context) {
	limit := 10 // Default to top 10

	// Allow custom limit via query param (max 50)
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 50 {
				parsedLimit = 50
			}
			limit = parsedLimit
		}
	}

	entries, err := h.DB.GetLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": len(entries),
	})
}

// ActivityHandler returns the user's daily activity for the contribution graph
func (h *Handlers) ActivityHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get year from query param, default to current year
	year := time.Now().Year()
	if yearStr := c.Query("year"); yearStr != "" {
		if parsedYear, err := strconv.Atoi(yearStr); err == nil {
			year = parsedYear
		}
	}

	activity, err := h.DB.GetUserActivity(userID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity"})
		return
	}

	// Return empty array if nil
	if activity == nil {
		activity = []db.ActivityEntry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activity,
		"year": year,
	})
}

// StatsHandler returns comprehensive user statistics
func (h *Handlers) StatsHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	stats, err := h.DB.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RecentSubmissionsHandler returns the user's recent submissions
func (h *Handlers) RecentSubmissionsHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 20 {
				parsedLimit = 20
			}
			limit = parsedLimit
		}
	}

	submissions, err := h.DB.GetRecentSubmissions(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	if submissions == nil {
		submissions = []db.RecentSubmission{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  submissions,
		"total": len(submissions),
	})
}

// TechFocusHandler returns the user's tech focus breakdown
func (h *Handlers) TechFocusHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	techFocus, err := h.DB.GetUserTechFocus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tech focus"})
		return
	}

	if techFocus == nil {
		techFocus = []db.TechFocus{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": techFocus,
	})
}

// ReviewHandler returns the user's submission details for a specific challenge (replaces AI review)
func (h *Handlers) ReviewHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	challengeID := c.Param("challengeId")
	if challengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Challenge ID required"})
		return
	}

	subs, err := h.DB.GetSubmissionsForChallenge(challengeID, userID)
	if err != nil || len(subs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No submission found for this challenge"})
		return
	}

	var ownSub *db.SubmissionDetail
	for _, s := range subs {
		if s.ClerkUserID == userID {
			ownSub = &s
			break
		}
	}

	if ownSub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No submission found for this challenge"})
		return
	}

	c.JSON(http.StatusOK, ownSub)
}

// SubmitHandler creates a new showcase submission or updates it if challenge-based and already exists
func (h *Handlers) SubmitHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		ChallengeID *string `json:"challenge_id"`
		Title       string  `json:"title" binding:"required"`
		RepoURL     string  `json:"repo_url" binding:"required"`
		VideoURL    string  `json:"video_url"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and repo_url are required"})
		return
	}

	// Validate repo URL (basic check)
	if !strings.HasPrefix(req.RepoURL, "https://github.com/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only GitHub repository URLs are supported"})
		return
	}

	// Create or Update submission in DB
	submissionID, updated, err := h.DB.CreateOrUpdateSubmission(userID, req.ChallengeID, req.Title, req.RepoURL, req.VideoURL, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit project"})
		return
	}

	// Recalculate score and streaks immediately
	_ = h.DB.UpdateUserStats(c.Request.Context(), userID)

	message := "Showcase project submitted successfully"
	if updated {
		message = "Showcase project updated successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"submission_id": submissionID,
		"updated":       updated,
		"message":       message,
	})
}

// SubmissionStatusHandler returns details for a single submission
func (h *Handlers) SubmissionStatusHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	submissionID := c.Param("id")
	if submissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Submission ID required"})
		return
	}

	sub, err := h.DB.GetSubmissionByID(submissionID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ChallengeSubmissionsHandler returns all user submissions for a specific challenge
func (h *Handlers) ChallengeSubmissionsHandler(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	challengeID := c.Param("id")
	if challengeID == "" {
		challengeID = c.Param("challengeId")
	}
	if challengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Challenge ID required"})
		return
	}

	subs, err := h.DB.GetSubmissionsForChallenge(challengeID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": subs,
		"count":       len(subs),
	})
}
