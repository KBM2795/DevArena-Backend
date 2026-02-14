package evaluator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// GitHubContentResponse is the API response for GET /repos/{owner}/{repo}/contents/{path}
type GitHubContentResponse struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Type     string `json:"type"`
}

// FetchTestFile downloads a private test file from GitHub using the Contents API
// testRepoURL format: "https://github.com/{owner}/{repo}/blob/{branch}/{path}"
// Returns the decoded file content and filename
func FetchTestFile(testRepoURL, githubToken string) (content []byte, filename string, err error) {
	// Parse the GitHub URL to extract owner, repo, branch, and path
	apiURL, fname, err := convertToAPIURL(testRepoURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid test_repo_url: %w", err)
	}

	// Make the API request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var ghResp GitHubContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, "", fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	if ghResp.Type != "file" {
		return nil, "", fmt.Errorf("expected file but got %s", ghResp.Type)
	}

	// Decode base64 content (GitHub returns content with newlines in base64)
	// Remove newlines from base64 string before decoding
	cleanContent := regexp.MustCompile(`\s`).ReplaceAllString(ghResp.Content, "")
	decoded, err := base64.StdEncoding.DecodeString(cleanContent)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return decoded, fname, nil
}

// convertToAPIURL converts a GitHub blob URL to a Contents API URL
// Input:  "https://github.com/autonerveai27/private_tests_DevArena/blob/main/challenge-1-css-flexbox/private-tests/layout.grader.test.tsx"
// Output: "https://api.github.com/repos/autonerveai27/private_tests_DevArena/contents/challenge-1-css-flexbox/private-tests/layout.grader.test.tsx?ref=main"
func convertToAPIURL(githubURL string) (apiURL string, filename string, err error) {
	// Pattern: https://github.com/{owner}/{repo}/blob/{branch}/{path...}
	re := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$`)
	matches := re.FindStringSubmatch(githubURL)
	if matches == nil {
		return "", "", fmt.Errorf("URL does not match GitHub blob pattern: %s", githubURL)
	}

	owner := matches[1]
	repo := matches[2]
	branch := matches[3]
	path := matches[4]

	// Extract filename from path
	parts := regexp.MustCompile(`[/\\]`).Split(path, -1)
	fname := parts[len(parts)-1]

	apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	return apiURL, fname, nil
}
