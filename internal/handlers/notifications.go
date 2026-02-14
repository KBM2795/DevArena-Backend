package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetNotifications returns the user's notifications
// Query params: limit (default 20), unread_only (bool)
func (h *Handlers) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	unreadOnly := c.Query("unread_only") == "true"

	if unreadOnly {
		notifs, err := h.DB.GetUnreadNotifications(userID)
		if err != nil {
			log.Printf("Failed to to get unread notifications: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
			return
		}
		c.JSON(http.StatusOK, notifs)
	} else {
		// Default fetch (limit 20)
		notifs, err := h.DB.GetAllNotifications(userID, 50)
		if err != nil {
			log.Printf("Failed to get all notifications: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
			return
		}
		c.JSON(http.StatusOK, notifs)
	}
}

// MarkNotificationRead marks a single notification as read
func (h *Handlers) MarkNotificationRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.DB.MarkNotificationRead(id, userID); err != nil {
		log.Printf("Failed to mark notification read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllNotificationsRead marks all user notifications as read
func (h *Handlers) MarkAllNotificationsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.DB.MarkAllNotificationsRead(userID); err != nil {
		log.Printf("Failed to mark all notifications read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteNotification deletes a notification
func (h *Handlers) DeleteNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.DB.DeleteNotification(id, userID); err != nil {
		log.Printf("Failed to delete notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
