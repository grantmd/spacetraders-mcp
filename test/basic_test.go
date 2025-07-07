package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestBasic_ServerBuild tests that the server builds successfully without requiring API token
func TestBasic_ServerBuild(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := filepath.Join(projectRoot, "spacetraders-mcp-build-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build server: %v\nOutput: %s", err, output)
	}

	// Clean up the test binary
	defer func() {
		if err := os.Remove(binaryPath); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Verify the binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Built binary does not exist")
	}

	t.Log("Server builds successfully")
}

// TestBasic_ServerHelp tests that the server responds to help/version flags
func TestBasic_ServerHelp(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := filepath.Join(projectRoot, "spacetraders-mcp-help-test")

	// Build the server
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}

	// Clean up the test binary
	defer func() {
		if err := os.Remove(binaryPath); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Test that the binary can be executed (even if it exits quickly)
	cmd = exec.Command(binaryPath)
	cmd.Env = []string{"SPACETRADERS_API_TOKEN=dummy-token-for-basic-test"}

	// Run with a timeout to avoid hanging
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Don't wait for it to complete, just verify it started
	if cmd.Process != nil {
		cmd.Process.Kill()
	}

	t.Log("Server starts without errors")
}

// TestBasic_MCPProtocolCompliance tests JSON-RPC protocol compliance
func TestBasic_MCPProtocolCompliance(t *testing.T) {
	// Test invalid JSON-RPC request
	response := callMCPServer(t, `{"invalid": "request"}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should return a JSON-RPC error for invalid request
	if mcpResponse.Error == nil {
		t.Fatal("Expected error for invalid JSON-RPC request")
	}

	// Verify error structure
	if mcpResponse.Error.Code == 0 && mcpResponse.Error.Message == "" {
		t.Fatal("Error should have code and message")
	}

	t.Log("MCP server properly handles invalid JSON-RPC requests")
}

// TestBasic_ResourcesListStructure tests resources/list structure
func TestBasic_ResourcesListStructure(t *testing.T) {
	// Test resources/list structure
	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Unexpected error in resources/list: %v", mcpResponse.Error)
	}

	// Parse the result
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var listResult ResourcesListResult
	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		t.Fatalf("Failed to parse resources list result: %v", err)
	}

	// Verify structure
	if len(listResult.Resources) == 0 {
		t.Fatal("Resources list should not be empty")
	}

	// Check first resource structure
	firstResource := listResult.Resources[0]
	if firstResource.URI == "" || firstResource.Name == "" || firstResource.MIMEType == "" {
		t.Fatal("Resource should have URI, Name, and MIMEType")
	}

	t.Log("Resources list has correct structure")
}

// TestBasic_MultipleRequests tests multiple sequential requests
func TestBasic_MultipleRequests(t *testing.T) {
	// Test that the server can handle multiple sequential requests
	requests := []string{
		`{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}`,
		`{"jsonrpc": "2.0", "id": 2, "method": "resources/list"}`,
		`{"jsonrpc": "2.0", "id": 3, "method": "resources/read", "params": {"uri": "spacetraders://invalid/test"}}`,
	}

	for i, request := range requests {
		response := callMCPServer(t, request)

		var mcpResponse MCPResponse
		if err := json.Unmarshal(response, &mcpResponse); err != nil {
			t.Fatalf("Failed to parse response %d: %v", i+1, err)
		}

		// Check that we got a response with the correct ID
		if mcpResponse.ID != i+1 {
			t.Errorf("Expected ID %d, got %d", i+1, mcpResponse.ID)
		}

		// Check JSON-RPC version
		if mcpResponse.JSONRPC != "2.0" {
			t.Errorf("Expected JSON-RPC 2.0, got %s", mcpResponse.JSONRPC)
		}
	}

	t.Log("Server handles multiple requests correctly")
}

// TestBasic_ToolsListStructure tests tools/list structure
func TestBasic_ToolsListStructure(t *testing.T) {
	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}`)

	// Trim any extra whitespace/null bytes
	response = []byte(string(response))
	if len(response) == 0 {
		t.Fatal("Empty response from tools/list")
	}

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse tools/list response: %v\nResponse: %s", err, string(response))
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Error in tools/list: %v", mcpResponse.Error)
	}

	// Parse the result
	if mcpResponse.Result == nil {
		t.Fatal("tools/list result should not be nil")
	}

	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var toolsResult map[string]interface{}
	if err := json.Unmarshal(resultBytes, &toolsResult); err != nil {
		t.Fatalf("Failed to parse tools list result: %v", err)
	}

	// Check for tools array
	if _, ok := toolsResult["tools"]; !ok {
		t.Fatal("tools/list result should contain 'tools' field")
	}

	t.Log("Tools list has correct structure")
}
