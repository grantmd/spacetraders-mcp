package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorNC     = "\033[0m" // No Color
)

func main() {
	fmt.Printf("%sSpaceTraders MCP Server Test Runner%s\n", colorYellow, colorNC)
	fmt.Println("======================================")

	// Check if we should run integration tests
	runIntegration := false
	if len(os.Args) > 1 && os.Args[1] == "--integration" {
		runIntegration = true
	}

	// Run unit tests first
	fmt.Printf("\n%sRunning unit tests...%s\n", colorYellow, colorNC)
	if err := runUnitTests(); err != nil {
		fmt.Printf("%sUnit tests failed: %v%s\n", colorRed, err, colorNC)
		os.Exit(1)
	}
	fmt.Printf("%sUnit tests passed!%s\n", colorGreen, colorNC)

	// Run integration tests
	fmt.Printf("\n%sRunning integration tests...%s\n", colorYellow, colorNC)
	if err := runIntegrationTests(runIntegration); err != nil {
		fmt.Printf("%sIntegration tests failed: %v%s\n", colorRed, err, colorNC)
		os.Exit(1)
	}
	fmt.Printf("%sIntegration tests passed!%s\n", colorGreen, colorNC)

	// Test build
	fmt.Printf("\n%sTesting build...%s\n", colorYellow, colorNC)
	if err := testBuild(); err != nil {
		fmt.Printf("%sBuild test failed: %v%s\n", colorRed, err, colorNC)
		os.Exit(1)
	}
	fmt.Printf("%sBuild test passed!%s\n", colorGreen, colorNC)

	fmt.Printf("\n%sAll tests completed successfully!%s\n", colorGreen, colorNC)

	// Print usage information
	fmt.Printf("\n%sTips:%s\n", colorYellow, colorNC)
	fmt.Println("- Run with --integration to test with real API calls (requires SPACETRADERS_API_TOKEN)")
	fmt.Println("- Make sure your .env file contains a valid SPACETRADERS_API_TOKEN for integration tests")
	fmt.Println("- Unit tests and protocol compliance tests run without requiring an API token")
	fmt.Println("- Integration tests will be skipped if no API token is provided")
}

func runUnitTests() error {
	// Run unit tests in the pkg directory
	cmd := exec.Command("go", "test", "./pkg/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runIntegrationTests(withRealAPI bool) error {
	args := []string{"test", "-v", "./test/..."}

	if !withRealAPI {
		// Run without real API calls
		args = append(args, "-timeout", "30s")
	} else {
		// Run with longer timeout for real API calls
		args = append(args, "-timeout", "60s")

		// Check if API token is available
		if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
			fmt.Printf("%sWarning: SPACETRADERS_API_TOKEN not set. Some tests will be skipped.%s\n", colorYellow, colorNC)
		}
	}

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()

	// Print the output
	fmt.Print(string(output))

	// Check for specific test results
	outputStr := string(output)
	if strings.Contains(outputStr, "FAIL") {
		return fmt.Errorf("some tests failed")
	}

	return err
}

func testBuild() error {
	// Test that the main binary builds successfully
	cmd := exec.Command("go", "build", "-o", "spacetraders-mcp-test", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %v", err)
	}

	// Clean up the test binary
	if err := os.Remove("spacetraders-mcp-test"); err != nil {
		fmt.Printf("%sWarning: Failed to clean up test binary: %v%s\n", colorYellow, err, colorNC)
	}

	return nil
}
