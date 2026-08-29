package handlers

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/auth/middleware"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfsign "github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getS3Client builds an S3 client from environment configuration.
func getS3Client() (*s3.Client, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-south-1"
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}

// deleteS3Object removes an object from S3. Given a full URL (S3 or CloudFront)
// it derives the S3 key; otherwise it treats the input as a raw key.
func deleteS3Object(raw string) error {
	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		bucketName = "devarena-videos"
	}

	key := raw
	// If it's a full URL (https://host/...), strip the scheme and host to get the key.
	if strings.Contains(raw, "://") {
		after := raw[strings.Index(raw, "://")+3:]
		// For S3 URLs the bucket is part of the host, for CloudFront URLs it is not.
		parts := strings.SplitN(after, "/", 2)
		if len(parts) == 2 {
			if strings.Contains(parts[0], ".s3.") || strings.Contains(parts[0], ".s3-") {
				bucketName = parts[0]
				key = parts[1]
			} else {
				key = parts[1]
			}
		}
	}

	client, err := getS3Client()
	if err != nil {
		return fmt.Errorf("failed to init AWS S3 client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}

// UploadVideoHandler handles uploading a video demo file to AWS S3
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

	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		bucketName = "devarena-videos"
	}

	// 5. Build S3 upload path
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s-%d-%s%s", userID, time.Now().Unix(), uuid.New().String()[:8], ext)
	key := fmt.Sprintf("video-demos/%s", uniqueFilename)

	// 6. Upload to S3
	client, err := getS3Client()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init AWS S3 client: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S3 upload failed: " + err.Error()})
		return
	}

	// 7. Build a signed CloudFront URL for playback (1-day expiry)
	cfDomain := os.Getenv("AWS_CF_DOMAIN")
	if cfDomain != "" && os.Getenv("AWS_CF_KEY_ID") != "" && os.Getenv("AWS_CF_PRIVATE_KEY_PATH") != "" {
		signedURL, err := signCloudFrontURL(cfDomain, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign video URL: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"video_url": signedURL,
			"filename":  uniqueFilename,
		})
		return
	}

	// Fall back to a public S3 URL when CloudFront is not configured
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-south-1"
	}
	c.JSON(http.StatusOK, gin.H{
		"video_url": fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, key),
		"filename":  uniqueFilename,
	})
}

// signCloudFrontURL creates a signed CloudFront URL valid for 24 hours.
// It uses the AWS CloudFront canned-policy signing algorithm (RSA-SHA1 over
// the canned policy JSON) with the correct URL-safe base64 escaping.
func signCloudFrontURL(baseDomain, key string) (string, error) {
	cfDomain := strings.TrimSuffix(baseDomain, "/")
	cfKeyID := os.Getenv("AWS_CF_KEY_ID")
	privateKeyPath := os.Getenv("AWS_CF_PRIVATE_KEY_PATH")

	if cfKeyID == "" || privateKeyPath == "" {
		return "", fmt.Errorf("AWS_CF_KEY_ID and AWS_CF_PRIVATE_KEY_PATH are required")
	}

	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CloudFront private key: %w", err)
	}

	// Parse the RSA private key (PEM block, PKCS8 or PKCS1).
	// Be tolerant of copy-paste corruption: strip a UTF-8 BOM, normalize line
	// endings, and extract only the region between the first BEGIN and last END
	// fence so stray whitespace/headers around the block don't break decoding.
	block, err := decodePEMTolerant(privateKeyPEM)
	if err != nil {
		return "", err
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an RSA key")
	}

	// Sign the CloudFront URL (canned policy with 24h expiry)
	resource := fmt.Sprintf("%s/%s", cfDomain, key)
	signer := cfsign.NewURLSigner(cfKeyID, rsaKey)
	return signer.Sign(resource, time.Now().Add(24*time.Hour))
}

// decodePEMTolerant extracts and decodes a PEM block, tolerating common
// copy-paste corruption (BOM, CRLF, stray whitespace/headers around the block).
func decodePEMTolerant(raw []byte) (*pem.Block, error) {
	s := string(raw)
	// Strip UTF-8 BOM
	s = strings.TrimPrefix(s, "\ufeff")
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Extract everything between the first BEGIN and last END fence.
	begin := strings.Index(s, "-----BEGIN")
	end := strings.LastIndex(s, "-----END")
	var blockData []byte
	if begin < 0 || end < 0 {
		blockData = []byte(s)
	} else {
		if end+len("-----END")+1 < len(s) {
			blockData = []byte(s[begin : end+len("-----END")+1])
		} else {
			blockData = []byte(s[begin:])
		}
	}

	block, _ := pem.Decode(blockData)
	if block == nil {
		// Last resort: try decoding the whole (trimmed) input.
		block, _ = pem.Decode(bytes.TrimSpace(raw))
	}
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from private key")
	}
	return block, nil
}
