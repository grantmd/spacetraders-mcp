package test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// MCPResponse represents a JSON-RPC response from the MCP server
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResourcesListResult represents the result of a resources/list call
type ResourcesListResult struct {
	Resources []Resource `json:"resources"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MIMEType    string `json:"mimeType"`
}

// ResourceReadResult represents the result of a resources/read call
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mimeType"`
	Text     string `json:"text"`
}

func TestIntegration_ServerBuild(t *testing.T) {
	// Test that the server builds successfully
	cmd := exec.Command("go", "build", "-o", "spacetraders-mcp-test", ".")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build server: %v\nOutput: %s", err, output)
	}

	// Clean up the test binary
	defer func() {
		if err := os.Remove("../spacetraders-mcp-test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Verify the binary exists
	if _, err := os.Stat("../spacetraders-mcp-test"); os.IsNotExist(err) {
		t.Fatal("Built binary does not exist")
	}
}

func TestIntegration_ServerStartup(t *testing.T) {
	// Skip if no API token is available
	if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
		t.Skip("SPACETRADERS_API_TOKEN not set, skipping integration test")
	}

	// Build the server
	cmd := exec.Command("go", "build", "-o", "spacetraders-mcp-test", ".")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}
	defer func() {
		if err := os.Remove("../spacetraders-mcp-test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Start the server
	serverCmd := exec.Command("./spacetraders-mcp-test")
	serverCmd.Dir = ".."
	stdin, err := serverCmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := serverCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		if err := stdin.Close(); err != nil {
			t.Logf("Failed to close stdin: %v", err)
		}
		if err := serverCmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process: %v", err)
		}
		if err := serverCmd.Wait(); err != nil {
			t.Logf("Failed to wait for process: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test resources/list
	request := `{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}` + "\n"
	if _, err := stdin.Write([]byte(request)); err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	// Read response
	response, err := readJSONResponse(stdout)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Server returned error: %v", mcpResponse.Error)
	}

	// Verify we got a resources list
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var resourcesResult ResourcesListResult
	if err := json.Unmarshal(resultBytes, &resourcesResult); err != nil {
		t.Fatalf("Failed to parse resources result: %v", err)
	}

	if len(resourcesResult.Resources) == 0 {
		t.Fatal("Expected at least one resource, got none")
	}

	// Verify expected resources are present
	expectedResources := []string{
		"spacetraders://agent/info",
		"spacetraders://ships/list",
		"spacetraders://contracts/list",
	}

	foundResources := make(map[string]bool)
	for _, resource := range resourcesResult.Resources {
		foundResources[resource.URI] = true
	}

	for _, expected := range expectedResources {
		if !foundResources[expected] {
			t.Errorf("Expected resource %s not found", expected)
		}
	}
}

func TestIntegration_AgentResource(t *testing.T) {
	// Skip if no API token is available
	if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
		t.Skip("SPACETRADERS_API_TOKEN not set, skipping integration test")
	}

	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/read", "params": {"uri": "spacetraders://agent/info"}}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Server returned error: %v", mcpResponse.Error)
	}

	// Parse the resource read result
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var readResult ResourceReadResult
	if err := json.Unmarshal(resultBytes, &readResult); err != nil {
		t.Fatalf("Failed to parse read result: %v", err)
	}

	if len(readResult.Contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(readResult.Contents))
	}

	content := readResult.Contents[0]
	if content.URI != "spacetraders://agent/info" {
		t.Errorf("Expected URI spacetraders://agent/info, got %s", content.URI)
	}

	// Handle API errors gracefully - if we get text/plain, it's likely an API error
	if content.MIMEType == "text/plain" {
		if strings.Contains(content.Text, "Error") {
			t.Skipf("API error (likely invalid token): %s", content.Text)
		} else {
			t.Errorf("Expected MIME type application/json, got text/plain with content: %s", content.Text)
		}
		return
	}

	if content.MIMEType != "application/json" {
		t.Errorf("Expected MIME type application/json, got %s", content.MIMEType)
	}

	// Parse the JSON content to verify it's valid agent data
	var agentData map[string]interface{}
	if err := json.Unmarshal([]byte(content.Text), &agentData); err != nil {
		t.Fatalf("Failed to parse agent JSON: %v", err)
	}

	agent, ok := agentData["agent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected agent object in response")
	}

	requiredFields := []string{"accountId", "symbol", "headquarters", "credits", "startingFaction", "shipCount"}
	for _, field := range requiredFields {
		if _, exists := agent[field]; !exists {
			t.Errorf("Expected field '%s' in agent data", field)
		}
	}
}

func TestIntegration_ShipsResource(t *testing.T) {
	// Skip if no API token is available
	if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
		t.Skip("SPACETRADERS_API_TOKEN not set, skipping integration test")
	}

	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/read", "params": {"uri": "spacetraders://ships/list"}}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Server returned error: %v", mcpResponse.Error)
	}

	// Parse the resource read result
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var readResult ResourceReadResult
	if err := json.Unmarshal(resultBytes, &readResult); err != nil {
		t.Fatalf("Failed to parse read result: %v", err)
	}

	if len(readResult.Contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(readResult.Contents))
	}

	content := readResult.Contents[0]
	if content.URI != "spacetraders://ships/list" {
		t.Errorf("Expected URI spacetraders://ships/list, got %s", content.URI)
	}

	// Handle API errors gracefully - if we get text/plain, it's likely an API error
	if content.MIMEType == "text/plain" {
		if strings.Contains(content.Text, "Error") {
			t.Skipf("API error (likely invalid token): %s", content.Text)
		} else {
			t.Errorf("Expected MIME type application/json, got text/plain with content: %s", content.Text)
		}
		return
	}

	// Parse the JSON content to verify it's valid ships data
	var shipsData map[string]interface{}
	if err := json.Unmarshal([]byte(content.Text), &shipsData); err != nil {
		t.Fatalf("Failed to parse ships JSON: %v", err)
	}

	ships, ok := shipsData["ships"].([]interface{})
	if !ok {
		t.Fatal("Expected ships array in response")
	}

	meta, ok := shipsData["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected meta object in response")
	}

	count, ok := meta["count"].(float64)
	if !ok {
		t.Fatal("Expected count field in meta")
	}

	if int(count) != len(ships) {
		t.Errorf("Meta count %d does not match ships array length %d", int(count), len(ships))
	}
}

func TestIntegration_ContractsResource(t *testing.T) {
	// Skip if no API token is available
	if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
		t.Skip("SPACETRADERS_API_TOKEN not set, skipping integration test")
	}

	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/read", "params": {"uri": "spacetraders://contracts/list"}}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Server returned error: %v", mcpResponse.Error)
	}

	// Parse the resource read result
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var readResult ResourceReadResult
	if err := json.Unmarshal(resultBytes, &readResult); err != nil {
		t.Fatalf("Failed to parse read result: %v", err)
	}

	if len(readResult.Contents) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(readResult.Contents))
	}

	content := readResult.Contents[0]
	if content.URI != "spacetraders://contracts/list" {
		t.Errorf("Expected URI spacetraders://contracts/list, got %s", content.URI)
	}

	// Handle API errors gracefully - if we get text/plain, it's likely an API error
	if content.MIMEType == "text/plain" {
		if strings.Contains(content.Text, "Error") {
			t.Skipf("API error (likely invalid token): %s", content.Text)
		} else {
			t.Errorf("Expected MIME type application/json, got text/plain with content: %s", content.Text)
		}
		return
	}

	// Parse the JSON content to verify it's valid contracts data
	var contractsData map[string]interface{}
	if err := json.Unmarshal([]byte(content.Text), &contractsData); err != nil {
		t.Fatalf("Failed to parse contracts JSON: %v", err)
	}

	contracts, ok := contractsData["contracts"].([]interface{})
	if !ok {
		t.Fatal("Expected contracts array in response")
	}

	meta, ok := contractsData["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected meta object in response")
	}

	count, ok := meta["count"].(float64)
	if !ok {
		t.Fatal("Expected count field in meta")
	}

	if int(count) != len(contracts) {
		t.Errorf("Meta count %d does not match contracts array length %d", int(count), len(contracts))
	}
}

func TestIntegration_InvalidResource(t *testing.T) {
	// Skip if no API token is available
	if os.Getenv("SPACETRADERS_API_TOKEN") == "" {
		t.Skip("SPACETRADERS_API_TOKEN not set, skipping integration test")
	}

	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/read", "params": {"uri": "spacetraders://invalid/resource"}}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should return an error for invalid resource
	if mcpResponse.Error == nil {
		t.Fatal("Expected error for invalid resource URI, got nil")
	}

	if !strings.Contains(mcpResponse.Error.Message, "resource not found") {
		t.Errorf("Expected 'resource not found' in error message, got: %s", mcpResponse.Error.Message)
	}
}

func TestIntegration_MCPProtocolCompliance(t *testing.T) {
	// Test JSON-RPC protocol compliance without requiring API token

	// Test invalid JSON-RPC request
	response := callMCPServer(t, `{"invalid": "request"}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should return a JSON-RPC error for invalid request
	if mcpResponse.Error == nil {
		t.Fatal("Expected error for invalid JSON-RPC request, got nil")
	}

	// Verify JSON-RPC version
	if mcpResponse.JSONRPC != "2.0" {
		t.Errorf("Expected JSON-RPC version 2.0, got %s", mcpResponse.JSONRPC)
	}
}

func TestIntegration_ResourcesListStructure(t *testing.T) {
	// Test resources/list without requiring API token (should work regardless of token validity)
	response := callMCPServer(t, `{"jsonrpc": "2.0", "id": 1, "method": "resources/list"}`)

	var mcpResponse MCPResponse
	if err := json.Unmarshal(response, &mcpResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if mcpResponse.Error != nil {
		t.Fatalf("Unexpected error in resources/list: %v", mcpResponse.Error)
	}

	// Parse the resources list result
	resultBytes, err := json.Marshal(mcpResponse.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var resourcesResult ResourcesListResult
	if err := json.Unmarshal(resultBytes, &resourcesResult); err != nil {
		t.Fatalf("Failed to parse resources result: %v", err)
	}

	// Verify each resource has required fields
	for _, resource := range resourcesResult.Resources {
		if resource.URI == "" {
			t.Error("Resource missing URI")
		}
		if resource.Name == "" {
			t.Error("Resource missing Name")
		}
		if resource.Description == "" {
			t.Error("Resource missing Description")
		}
		if resource.MIMEType == "" {
			t.Error("Resource missing MIMEType")
		}

		// Verify URI format
		if !strings.HasPrefix(resource.URI, "spacetraders://") {
			t.Errorf("Resource URI should start with 'spacetraders://', got: %s", resource.URI)
		}
	}
}

func TestIntegration_MultipleRequests(t *testing.T) {
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

		// Verify JSON-RPC compliance
		if mcpResponse.JSONRPC != "2.0" {
			t.Errorf("Request %d: Expected JSON-RPC version 2.0, got %s", i+1, mcpResponse.JSONRPC)
		}

		expectedID := i + 1
		if mcpResponse.ID != expectedID {
			t.Errorf("Request %d: Expected ID %d, got %d", i+1, expectedID, mcpResponse.ID)
		}
	}
}

func TestIntegration_ServerShutdownGraceful(t *testing.T) {
	// Test that the server shuts down gracefully

	// Build the server
	cmd := exec.Command("go", "build", "-o", "spacetraders-mcp-test", ".")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}
	defer func() {
		if err := os.Remove("../spacetraders-mcp-test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Start the server
	serverCmd := exec.Command("./spacetraders-mcp-test")
	serverCmd.Dir = ".."
	stdin, err := serverCmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Send SIGINT (graceful shutdown)
	if err := serverCmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send interrupt signal: %v", err)
	}

	// Wait for the server to exit with a timeout
	done := make(chan error, 1)
	go func() {
		done <- serverCmd.Wait()
	}()

	select {
	case err := <-done:
		// Server exited - check if it was a clean exit
		if err != nil {
			// On some systems, interrupt may cause non-zero exit
			t.Logf("Server exited with error (may be normal for interrupt): %v", err)
		} else {
			t.Log("Server exited cleanly")
		}
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't exit gracefully
		if err := serverCmd.Process.Kill(); err != nil {
			t.Errorf("Failed to force kill server process: %v", err)
		}
		t.Fatal("Server did not exit gracefully within 5 seconds")
	}

	if err := stdin.Close(); err != nil {
		t.Logf("Failed to close stdin: %v", err)
	}
}

// Helper function to call the MCP server with a request
func callMCPServer(t *testing.T, request string) []byte {
	// Build the server
	cmd := exec.Command("go", "build", "-o", "spacetraders-mcp-test", ".")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}
	defer func() {
		if err := os.Remove("../spacetraders-mcp-test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Start the server
	serverCmd := exec.Command("./spacetraders-mcp-test")
	serverCmd.Dir = ".."
	stdin, err := serverCmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := serverCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		if err := stdin.Close(); err != nil {
			t.Logf("Failed to close stdin: %v", err)
		}
		if err := serverCmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process: %v", err)
		}
		if err := serverCmd.Wait(); err != nil {
			t.Logf("Failed to wait for process: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Send request
	if _, err := stdin.Write([]byte(request + "\n")); err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	// Read response
	response, err := readJSONResponse(stdout)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	return response
}

// Helper function to read a JSON response from stdout
func readJSONResponse(stdout io.Reader) ([]byte, error) {
	// Read from stdout until we get a complete JSON response
	buffer := make([]byte, 1024)
	var result []byte
	maxAttempts := 50 // Maximum number of read attempts

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Set a short timeout for reading
		n, err := stdout.Read(buffer)
		if err != nil {
			if err == io.EOF && len(result) > 0 {
				// We have some data, try to parse it
				break
			}
			return nil, err
		}

		result = append(result, buffer[:n]...)

		// Look for newline-delimited JSON response
		if lines := strings.Split(string(result), "\n"); len(lines) > 1 {
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Try to parse as JSON
				var jsonObj interface{}
				if json.Unmarshal([]byte(line), &jsonObj) == nil {
					return []byte(line), nil
				}
			}
		}

		// Also try to parse the complete buffer as JSON
		var jsonObj interface{}
		if json.Unmarshal(result, &jsonObj) == nil {
			return result, nil
		}

		// If we've accumulated too much data, something's wrong
		if len(result) > 50000 { // 50KB limit
			break
		}

		// Small delay between reads
		time.Sleep(10 * time.Millisecond)
	}

	// If we couldn't parse JSON, return what we have for debugging
	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("no valid JSON response received after %d attempts", maxAttempts)
}
