package evaluator

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// PipelineConfig holds all configuration needed for the evaluation pipeline
type PipelineConfig struct {
	GitHubToken string       // PAT for accessing private test repo
	Docker      DockerConfig // Docker execution settings
}

// PipelineResult is the final output of the evaluation pipeline
type PipelineResult struct {
	TestResult *models.TestResult // Test pass/fail counts + details
	Score      int                // Calculated score
	MaxScore   int                // Maximum possible score
}

// RunPipeline runs the full evaluation pipeline for a submission
// Steps: clone → validate → fetch tests → docker run → parse → score
// Returns: Result, RepoDir (for AI review), CleanupFunc, Error
func RunPipeline(repoURL, branch, challengeID string, testRepoURL string, forbiddenPackages []string, maxScore int, cfg PipelineConfig) (*PipelineResult, string, func(), error) {

	// ─── Step 1: Create temp directory ───────────────────
	tempDir, err := os.MkdirTemp("", "devarena-eval-*")
	cleanup := func() {
		log.Printf("[Pipeline] Cleaning up temp dir: %s", tempDir)
		os.RemoveAll(tempDir)
	}

	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	repoDir := filepath.Join(tempDir, "repo")

	// ─── Step 2: Clone user's repo ───────────────────────
	log.Printf("[Pipeline] Cloning %s (branch: %s)...", repoURL, branch)
	err = gitClone(repoURL, branch, repoDir)
	if err != nil {
		cleanup()
		return nil, "", nil, fmt.Errorf("git clone failed: %w", err)
	}
	log.Printf("[Pipeline] Clone successful")

	// DEBUG: Log repo structure
	logRepoStructure(repoDir)

	// ─── Step 2.5: Detect Project Root ───────────────────
	// Some users might push a folder containing the code instead of the code itself
	// We search for package.json to find the real root
	if newRoot, found := findProjectRoot(repoDir); found {
		log.Printf("[Pipeline] Detected nested project root: %s", newRoot)
		repoDir = newRoot
	}

	// ─── Step 3: Validate template ───────────────────────
	log.Printf("[Pipeline] Validating forbidden packages...")
	if err := ValidateForbiddenPackages(repoDir, forbiddenPackages); err != nil {
		cleanup()
		return nil, "", nil, fmt.Errorf("template validation failed: %w", err)
	}
	log.Printf("[Pipeline] Validation passed")

	// ─── Step 4: Fetch + inject private tests ────────────
	if testRepoURL != "" {
		log.Printf("[Pipeline] Fetching private tests from GitHub...")
		testContent, testFilename, err := FetchTestFile(testRepoURL, cfg.GitHubToken)
		if err != nil {
			cleanup()
			return nil, "", nil, fmt.Errorf("failed to fetch private tests: %w", err)
		}

		// Determine where to inject the test file
		testDir := filepath.Join(repoDir, "src", "tests")
		if _, err := os.Stat(filepath.Join(repoDir, "tests")); err == nil {
			testDir = filepath.Join(repoDir, "tests")
		}

		if err := os.MkdirAll(testDir, 0755); err != nil {
			cleanup()
			return nil, "", nil, fmt.Errorf("failed to create test directory: %w", err)
		}

		// Remove existing test files
		files, _ := filepath.Glob(filepath.Join(testDir, "*"))
		for _, f := range files {
			if filepath.Ext(f) == ".tsx" || filepath.Ext(f) == ".ts" || filepath.Ext(f) == ".js" {
				if filepath.Base(f) != "setup.ts" {
					os.Remove(f)
				}
			}
		}

		testPath := filepath.Join(testDir, testFilename)
		if err := os.WriteFile(testPath, testContent, 0644); err != nil {
			cleanup()
			return nil, "", nil, fmt.Errorf("failed to write test file: %w", err)
		}
		log.Printf("[Pipeline] Injected test file: %s", testPath)
	}

	// ─── Step 5: Run tests in Docker ─────────────────────
	log.Printf("[Pipeline] Running tests in Docker container...")
	stdout, stderr, err := RunTestsInDocker(repoDir, cfg.Docker)
	if err != nil {
		log.Printf("[Pipeline] Docker stderr: %s", stderr)
		cleanup() // Failures here should clean up
		return nil, "", nil, fmt.Errorf("docker test execution failed: %w", err)
	}
	log.Printf("[Pipeline] Docker execution complete")

	// ─── Step 6: Parse test output ───────────────────────
	log.Printf("[Pipeline] Parsing test results...")
	testResult, err := ParseVitestOutput(stdout)
	if err != nil {
		log.Printf("[Pipeline] Failed to parse vitest output: %v", err)
		// We still return result for raw output visibility, but maybe fail pipeline?
		// Actually, if parse fails, we can't score.
		cleanup()
		return nil, "", nil, fmt.Errorf("failed to parse test output: %w", err)
	}

	// ─── Step 7: Calculate score ─────────────────────────
	score := CalculateScore(testResult, maxScore)
	log.Printf("[Pipeline] Results: %d/%d tests passed → Score: %d/%d",
		testResult.Passed, testResult.Total, score, maxScore)

	return &PipelineResult{
		TestResult: testResult,
		Score:      score,
		MaxScore:   maxScore,
	}, repoDir, cleanup, nil
}

// gitClone clones a git repository with depth 1
func gitClone(repoURL, branch, destDir string) error {
	// Check if git is available
	gitCmd := "git"
	if runtime.GOOS == "windows" {
		// Try to find git on Windows
		if _, err := exec.LookPath("git"); err != nil {
			gitCmd = "git.exe"
		}
	}

	args := []string{"clone", "--depth", "1", "--branch", branch, repoURL, destDir}
	cmd := exec.Command(gitCmd, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func logRepoStructure(root string) {
	log.Println("[Debug] Repo structure:")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		// Skip .git and node_modules to keep logs clean
		if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		log.Printf("  - %s", rel)
		return nil
	})
	if err != nil {
		log.Printf("[Debug] Failed to walk repo: %v", err)
	}
}

// findProjectRoot looks for package.json in root or immediate subdirs
func findProjectRoot(rootDir string) (string, bool) {
	// 1. Check root
	if _, err := os.Stat(filepath.Join(rootDir, "package.json")); err == nil {
		return rootDir, false // Root is valid
	}

	// 2. Check depth-1 subdirectories
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			possibleRoot := filepath.Join(rootDir, entry.Name())
			if _, err := os.Stat(filepath.Join(possibleRoot, "package.json")); err == nil {
				return possibleRoot, true // Found nested root
			}
		}
	}

	return "", false // Not found
}
