package handlers

import (
	"net/http"
	"strconv"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/gin-gonic/gin"
)

const defaultShowcaseLimit = 12

// GetOpenShowcasesHandler returns paginated open showcase submissions.
// Query params: search (string), page (int, 1-based, default 1), limit (int, default 12)
// GET /showcase  (public)
func (h *Handlers) GetOpenShowcasesHandler(c *gin.Context) {
	// Best-effort: pass current user ID to check if they have liked each item
	currentUserID, _ := middleware.GetUserID(c)

	search := c.Query("search")

	page := 1
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}

	limit := defaultShowcaseLimit
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	subs, total, err := h.DB.GetOpenShowcases(currentUserID, search, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch open showcases"})
		return
	}

	if subs == nil {
		subs = []db.SubmissionDetail{}
	}

	hasMore := (offset + len(subs)) < total

	c.JSON(http.StatusOK, gin.H{
		"data":     subs,
		"count":    len(subs),
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": hasMore,
	})
}

// GetUserShowcaseHandler returns all submissions (challenge + open) for the authenticated user
// GET /me/showcase
func (h *Handlers) GetUserShowcaseHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	subs, err := h.DB.GetUserSubmissions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your showcase"})
		return
	}

	if subs == nil {
		subs = []db.SubmissionDetail{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  subs,
		"count": len(subs),
	})
}

