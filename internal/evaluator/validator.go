package evaluator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateForbiddenPackages checks if the user's package.json contains any forbidden packages
func ValidateForbiddenPackages(repoDir string, forbiddenPackages []string) error {
	if len(forbiddenPackages) == 0 {
		return nil // No restrictions
	}

	pkgPath := filepath.Join(repoDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		// No package.json found — might be a non-Node project, skip check
		return nil
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("invalid package.json: %w", err)
	}

	// Check both "dependencies" and "devDependencies"
	depSections := []string{"dependencies", "devDependencies"}

	for _, section := range depSections {
		deps, ok := pkg[section].(map[string]interface{})
		if !ok {
			continue
		}

		for _, forbidden := range forbiddenPackages {
			forbidden = strings.TrimSpace(forbidden)
			if _, found := deps[forbidden]; found {
				return fmt.Errorf("forbidden package '%s' found in %s", forbidden, section)
			}
		}
	}

	return nil
}
