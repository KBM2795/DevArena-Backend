package models

// TestResult holds the parsed output from Vitest JSON reporter
type TestResult struct {
	Passed  int          `json:"passed"`
	Failed  int          `json:"failed"`
	Total   int          `json:"total"`
	Details []TestDetail `json:"details"`
}

// TestDetail holds info about a single test case
type TestDetail struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass" or "fail"
	Error  string `json:"error,omitempty"`
}

// VitestOutput is the top-level JSON structure from vitest --reporter=json
type VitestOutput struct {
	NumTotalTests  int               `json:"numTotalTests"`
	NumPassedTests int               `json:"numPassedTests"`
	NumFailedTests int               `json:"numFailedTests"`
	TestResults    []VitestTestSuite `json:"testResults"`
}

// VitestTestSuite represents a single test file's results
type VitestTestSuite struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"` // "passed" or "failed"
	AssertionResults []VitestAssertion `json:"assertionResults"`
}

// VitestAssertion represents a single test assertion (it/test block)
type VitestAssertion struct {
	FullName        string   `json:"fullName"`
	Title           string   `json:"title"`
	Status          string   `json:"status"` // "passed" or "failed"
	FailureMessages []string `json:"failureMessages"`
}
