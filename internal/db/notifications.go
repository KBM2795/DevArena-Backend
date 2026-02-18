package db

import (
	"context"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// CreateNotification adds a new notification for a user
func (db *Database) CreateNotification(userID, title, message, notifType, link string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, message, type, link)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, title, message, notifType, link)
	return err
}

// GetUnreadNotifications returns all unread notifications for a user, ordered by newest first
func (db *Database) GetUnreadNotifications(clerkUserID string) ([]models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, message, type, link, is_read, created_at
		FROM notifications
		WHERE user_id = $1 AND is_read = FALSE
		ORDER BY created_at DESC
	`, internalUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.Link, &n.IsRead, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// GetAllNotifications returns recent notifications (read and unread)
func (db *Database) GetAllNotifications(clerkUserID string, limit int) ([]models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, title, message, type, link, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, internalUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.Link, &n.IsRead, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// MarkNotificationRead marks a specific notification as read
func (db *Database) MarkNotificationRead(id, clerkUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify ownership via checking user_id mapping
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE notifications SET is_read = TRUE, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, id, internalUserID)
	return err
}

// MarkAllNotificationsRead marks all notifications as read for a user
func (db *Database) MarkAllNotificationsRead(clerkUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE notifications SET is_read = TRUE, updated_at = NOW()
		WHERE user_id = $1 AND is_read = FALSE
	`, internalUserID)
	return err
}

// DeleteNotification removes a notification
func (db *Database) DeleteNotification(id, clerkUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		DELETE FROM notifications
		WHERE id = $1 AND user_id = $2
	`, id, internalUserID)
	return err
}
