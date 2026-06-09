package handlers

import (
	"net/http"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/gin-gonic/gin"
)

// ToggleLikeHandler toggles a like on a submission for the authenticated user
// POST /submissions/:id/like
func (h *Handlers) ToggleLikeHandler(c *gin.Context) {
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

	liked, likeCount, err := h.DB.ToggleLike(userID, submissionID)
	if err != nil {
		if err.Error() == "submission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
		return
	}

	action := "liked"
	if !liked {
		action = "unliked"
	}

	c.JSON(http.StatusOK, gin.H{
		"liked":      liked,
		"like_count": likeCount,
		"message":    "Submission " + action + " successfully",
	})
}

// GetLikeStatusHandler returns the like count and whether the current user has liked a submission
// GET /submissions/:id/like
func (h *Handlers) GetLikeStatusHandler(c *gin.Context) {
	submissionID := c.Param("id")
	if submissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Submission ID required"})
		return
	}

	// userID is optional here – works for guests too
	currentUserID, _ := middleware.GetUserID(c)

	likeCount, hasLiked, err := h.DB.GetLikeStatus(currentUserID, submissionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get like status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"like_count": likeCount,
		"liked":      hasLiked,
	})
}
