package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DockerConfig holds configuration for the sandboxed test runner
type DockerConfig struct {
	Image       string        // Docker image (default: node:20-alpine)
	MemoryLimit string        // Memory limit (default: 512m)
	CPUs        string        // CPU limit (default: 1)
	PidsLimit   string        // PID limit (default: 100)
	Timeout     time.Duration // Timeout for container execution (default: 120s)
}

// DefaultDockerConfig returns safe default configuration
func DefaultDockerConfig() DockerConfig {
	return DockerConfig{
		Image:       "node:20-alpine",
		MemoryLimit: "1024m",
		CPUs:        "1",
		PidsLimit:   "100",
		Timeout:     300 * time.Second,
	}
}

// RunTestsInDocker runs npm install + vitest inside a sandboxed Docker container
// Returns the raw stdout (Vitest JSON output) and any error
func RunTestsInDocker(repoDir string, cfg DockerConfig) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Convert Windows path to Docker-compatible path if needed
	mountPath := repoDir
	if runtime.GOOS == "windows" {
		mountPath = strings.ReplaceAll(repoDir, `\`, `/`)
	}

	// ─── Step 1: Install Dependencies (With Network) ───────────────────
	log.Printf("[Docker] Step 1: Installing dependencies (with network)...")
	installArgs := []string{
		"run",
		"--rm",
		"--memory=" + cfg.MemoryLimit,
		"--cpus=" + cfg.CPUs,
		"--dns=8.8.8.8", // Use Google DNS to fix EAI_AGAIN errors
		"-v", mountPath + ":/app",
		"-w", "/app",
		cfg.Image,
		"sh", "-c",
		"npm install --verbose --ignore-scripts",
	}

	installCmd := exec.CommandContext(ctx, "docker", installArgs...)
	// Stream install logs to stdout/stderr for user visibility
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return "", "", fmt.Errorf("npm install failed: %w", err)
	}

	// ─── Step 2: Run Tests (No Network) ──────────────────────────────
	log.Printf("[Docker] Step 2: Running tests (no network)...")
	testArgs := []string{
		"run",
		"--rm",
		"--network=none", // Isolate user code
		"--memory=" + cfg.MemoryLimit,
		"--cpus=" + cfg.CPUs,
		"--pids-limit=" + cfg.PidsLimit,
		"-v", mountPath + ":/app",
		"-w", "/app",
		cfg.Image,
		"sh", "-c",
		// Use --no-file-parallelism to avoid OOM/crashes in Docker
		// Use --pool=forks as it can be more stable than threads for JSDOM
		"./node_modules/.bin/vitest run --reporter=json --no-file-parallelism --pool=forks",
	}

	cmd := exec.CommandContext(ctx, "docker", testArgs...)

	var stdoutBuf, stderrBuf bytes.Buffer
	// Stream output to both buffer (for parsing) and stdout/stderr
	cmd.Stdout = io.MultiWriter(&stdoutBuf, os.Stdout)
	cmd.Stderr = io.MultiWriter(&stderrBuf, os.Stderr)

	err = cmd.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	// Check timeout
	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, fmt.Errorf("test execution timed out after %s", cfg.Timeout)
	}

	if err != nil {
		if len(stdout) > 0 {
			return stdout, stderr, nil
		}
		return stdout, stderr, fmt.Errorf("docker execution failed: %w\nstderr: %s", err, stderr)
	}

	return stdout, stderr, nil
}
