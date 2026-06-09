package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadVideoHandler handles uploading a video demo file to Supabase Storage
// POST /me/submissions/upload-video
func (h *Handlers) UploadVideoHandler(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 1. Parse multipart form file
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}
	defer file.Close()

	// 2. Validate file size (e.g., max 100MB)
	const maxFileSize = 100 * 1024 * 1024 // 100 MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file is too large (max 100MB)"})
		return
	}

	// 3. Validate content type (must be video)
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "video/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only video file uploads are supported"})
		return
	}

	// 4. Read file bytes
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read upload file"})
		return
	}

	// 5. Build Supabase upload path
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "https://mrcnfawfpqcrgewmjrhp.supabase.co"
	}
	supabaseURL = strings.TrimSuffix(supabaseURL, "/")

	// Prefer service role key for server-side uploads (bypasses RLS)
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
	}

	if supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase key is not configured. Set SUPABASE_SERVICE_ROLE_KEY in local.yaml"})
		return
	}

	bucketName := os.Getenv("SUPABASE_BUCKET")
	if bucketName == "" {
		bucketName = "submissions" // Default bucket
	}

	// Create a unique filename: <userId>-<timestamp>-<uuid><ext>
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s-%d-%s%s", userID, time.Now().Unix(), uuid.New().String()[:8], ext)
	uploadPath := fmt.Sprintf("video-demos/%s", uniqueFilename)

	// 6. Execute POST request to Supabase Storage REST API
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, uploadPath)
	req, err := http.NewRequest(http.MethodPost, uploadURL, &buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Supabase request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true") // Allow overwriting if same path

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase storage connection failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Supabase upload failed (HTTP %d): %s", resp.StatusCode, string(respBytes)),
		})
		return
	}

	// 7. Construct and return public access URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, uploadPath)
	c.JSON(http.StatusOK, gin.H{
		"video_url": publicURL,
		"filename":  uniqueFilename,
	})
}
