package handlers

import (
	"net/http"
	"strconv"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/KBM2795/DevArena-Backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ToggleFollowHandler follows or unfollows a user
// POST /users/:id/follow
func (h *Handlers) ToggleFollowHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	isFollowing, followersCount, err := h.DB.ToggleFollow(userID, targetID)
	if err != nil {
		if err.Error() == "cannot follow yourself" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
			return
		}
		if err.Error() == "target user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle follow"})
		return
	}

	action := "followed"
	if !isFollowing {
		action = "unfollowed"
	}

	c.JSON(http.StatusOK, gin.H{
		"following":       isFollowing,
		"followers_count": followersCount,
		"message":         "User " + action + " successfully",
	})
}

// GetFollowStatusHandler returns follow status and counts for a user
// GET /users/:id/follow
func (h *Handlers) GetFollowStatusHandler(c *gin.Context) {
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	// Optional: requesting user for is_following check
	currentUserID, _ := middleware.GetUserID(c)

	isFollowing, followersCount, followingCount, err := h.DB.GetFollowStatus(currentUserID, targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"following":       isFollowing,
		"followers_count": followersCount,
		"following_count": followingCount,
	})
}

// GetFollowersHandler returns a paginated list of a user's followers
// GET /users/:id/followers
func (h *Handlers) GetFollowersHandler(c *gin.Context) {
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	currentUserID, _ := middleware.GetUserID(c)

	users, total, err := h.DB.GetFollowers(targetID, currentUserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followers"})
		return
	}

	if users == nil {
		users = []models.FollowUser{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
	})
}

// GetFollowingHandler returns a paginated list of users that a user follows
// GET /users/:id/following
func (h *Handlers) GetFollowingHandler(c *gin.Context) {
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	currentUserID, _ := middleware.GetUserID(c)

	users, total, err := h.DB.GetFollowing(targetID, currentUserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following"})
		return
	}

	if users == nil {
		users = []models.FollowUser{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
	})
}

// GetFollowingFeedHandler returns showcase submissions from followed users
// GET /me/following-feed
func (h *Handlers) GetFollowingFeedHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))

	submissions, total, err := h.DB.GetFollowingFeedSubmissions(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following feed"})
		return
	}

	if submissions == nil {
		submissions = []db.SubmissionDetail{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  submissions,
		"total": total,
		"page":  page,
	})
}

// SearchUsersHandler searches users by username or display name
// GET /users/search?q=query
func (h *Handlers) SearchUsersHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	currentUserID, _ := middleware.GetUserID(c)

	users, err := h.DB.SearchUsers(query, currentUserID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	if users == nil {
		users = []models.FollowUser{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// GetPublicUserProfileHandler returns a public user profile by username or id
// GET /users/:id/profile
func (h *Handlers) GetPublicUserProfileHandler(c *gin.Context) {
	username := c.Param("id")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or ID required"})
		return
	}

	// Optional auth for is_following
	currentUserID, _ := middleware.GetUserID(c)

	profile, err := h.DB.GetPublicUserProfile(username, currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
