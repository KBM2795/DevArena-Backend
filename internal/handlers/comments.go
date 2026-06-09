package handlers

import (
	"net/http"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/gin-gonic/gin"
)

// GetCommentsHandler returns comments for a challenge or submission
// Query params:
//   - challenge_id  (optional) – returns all challenge-level comments
//   - submission_id (optional) – returns project-specific comments (takes precedence)
func (h *Handlers) GetCommentsHandler(c *gin.Context) {
	submissionID := c.Query("submission_id")
	challengeID := c.Query("challenge_id")

	var subPtr, chalPtr *string
	if submissionID != "" {
		subPtr = &submissionID
	}
	if challengeID != "" {
		chalPtr = &challengeID
	}

	comments, err := h.DB.GetComments(chalPtr, subPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	if comments == nil {
		comments = []db.CommentDetail{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  comments,
		"count": len(comments),
	})
}

// PostCommentHandler adds a new comment to a challenge or project discussion
// Body (JSON):
//
//	{
//	  "message":       "Great project!",
//	  "challenge_id":  "ch-xxx",   // optional
//	  "submission_id": "sub-xxx"   // optional
//	}
func (h *Handlers) PostCommentHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Message      string  `json:"message" binding:"required"`
		ChallengeID  *string `json:"challenge_id"`
		SubmissionID *string `json:"submission_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	if len(req.Message) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message too long (max 2000 chars)"})
		return
	}

	comment, err := h.DB.SaveComment(userID, req.ChallengeID, req.SubmissionID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    comment,
		"message": "Comment posted successfully",
	})
}
