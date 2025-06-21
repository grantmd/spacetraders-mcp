package logging

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestNewLogger(t *testing.T) {
	// Create a test MCP server
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false),
	)

	logger := NewLogger(s)

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	if logger.mcpServer != s {
		t.Error("MCP server not set correctly")
	}

	if logger.errorLogger == nil {
		t.Error("Error logger not initialized")
	}

	if logger.infoLogger == nil {
		t.Error("Info logger not initialized")
	}

	if logger.debugLogger == nil {
		t.Error("Debug logger not initialized")
	}
}

func TestNewLogger_NilServer(t *testing.T) {
	logger := NewLogger(nil)

	if logger == nil {
		t.Fatal("Expected non-nil logger even with nil server")
	}

	if logger.mcpServer != nil {
		t.Error("Expected nil MCP server")
	}
}

func TestLogger_Info(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil, // No MCP server for this test
	}

	testMessage := "Test info message"
	logger.Info("%s", testMessage)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected [INFO] prefix in output")
	}
	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected message '%s' in output", testMessage)
	}
}

func TestLogger_Error(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		errorLogger: log.New(&buf, "[ERROR] ", 0),
		mcpServer:   nil, // No MCP server for this test
	}

	testMessage := "Test error message"
	logger.Error("%s", testMessage)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected [ERROR] prefix in output")
	}
	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected message '%s' in output", testMessage)
	}
}

func TestLogger_Debug(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		debugLogger: log.New(&buf, "[DEBUG] ", 0),
		mcpServer:   nil, // No MCP server for this test
	}

	testMessage := "Test debug message"
	logger.Debug("%s", testMessage)

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Expected [DEBUG] prefix in output")
	}
	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected message '%s' in output", testMessage)
	}
}

func TestLogger_InfoWithFormatting(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	logger.Info("Test message with value: %d", 42)

	output := buf.String()
	if !strings.Contains(output, "Test message with value: 42") {
		t.Error("Expected formatted message in output")
	}
}

func TestLogger_WithContext(t *testing.T) {
	logger := NewLogger(nil)
	ctx := context.Background()
	component := "test-component"

	contextLogger := logger.WithContext(ctx, component)

	if contextLogger == nil {
		t.Fatal("Expected non-nil context logger")
	}

	if contextLogger.logger != logger {
		t.Error("Context logger should reference original logger")
	}

	if contextLogger.context != ctx {
		t.Error("Context not set correctly")
	}

	if contextLogger.component != component {
		t.Errorf("Expected component '%s', got '%s'", component, contextLogger.component)
	}
}

func TestContextLogger_Info(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	ctx := context.Background()
	component := "test-component"
	contextLogger := logger.WithContext(ctx, component)

	testMessage := "Test context message"
	contextLogger.Info(testMessage)

	output := buf.String()
	expectedMessage := "[test-component] " + testMessage
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected message '%s' in output, got '%s'", expectedMessage, output)
	}
}

func TestContextLogger_Error(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		errorLogger: log.New(&buf, "[ERROR] ", 0),
		mcpServer:   nil,
	}

	ctx := context.Background()
	component := "test-component"
	contextLogger := logger.WithContext(ctx, component)

	testMessage := "Test error message"
	contextLogger.Error(testMessage)

	output := buf.String()
	expectedMessage := "[test-component] " + testMessage
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected message '%s' in output", expectedMessage)
	}
}

func TestContextLogger_Debug(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		debugLogger: log.New(&buf, "[DEBUG] ", 0),
		mcpServer:   nil,
	}

	ctx := context.Background()
	component := "test-component"
	contextLogger := logger.WithContext(ctx, component)

	testMessage := "Test debug message"
	contextLogger.Debug(testMessage)

	output := buf.String()
	expectedMessage := "[test-component] " + testMessage
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected message '%s' in output", expectedMessage)
	}
}

func TestContextLogger_APICall(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	ctx := context.Background()
	component := "api-client"
	contextLogger := logger.WithContext(ctx, component)

	endpoint := "/my/agent"
	statusCode := 200
	duration := "123ms"

	contextLogger.APICall(endpoint, statusCode, duration)

	output := buf.String()
	if !strings.Contains(output, endpoint) {
		t.Errorf("Expected endpoint '%s' in output", endpoint)
	}
	if !strings.Contains(output, "200") {
		t.Error("Expected status code 200 in output")
	}
	if !strings.Contains(output, duration) {
		t.Errorf("Expected duration '%s' in output", duration)
	}
	if !strings.Contains(output, "[api-client]") {
		t.Error("Expected component name in output")
	}
}

func TestContextLogger_ResourceRead(t *testing.T) {
	// Test successful resource read
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	ctx := context.Background()
	component := "agent-resource"
	contextLogger := logger.WithContext(ctx, component)

	uri := "spacetraders://agent/info"
	contextLogger.ResourceRead(uri, true)

	output := buf.String()
	if !strings.Contains(output, "Resource read successful") {
		t.Error("Expected success message for successful resource read")
	}
	if !strings.Contains(output, uri) {
		t.Errorf("Expected URI '%s' in output", uri)
	}

	// Test failed resource read
	buf.Reset()
	logger.errorLogger = log.New(&buf, "[ERROR] ", 0)
	contextLogger.ResourceRead(uri, false)

	output = buf.String()
	if !strings.Contains(output, "Resource read failed") {
		t.Error("Expected failure message for failed resource read")
	}
}

func TestContextLogger_ToolCall(t *testing.T) {
	// Test successful tool call
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	ctx := context.Background()
	component := "navigation-tool"
	contextLogger := logger.WithContext(ctx, component)

	toolName := "navigate_ship"
	contextLogger.ToolCall(toolName, true)

	output := buf.String()
	if !strings.Contains(output, "Tool call successful") {
		t.Error("Expected success message for successful tool call")
	}
	if !strings.Contains(output, toolName) {
		t.Errorf("Expected tool name '%s' in output", toolName)
	}

	// Test failed tool call
	buf.Reset()
	logger.errorLogger = log.New(&buf, "[ERROR] ", 0)
	contextLogger.ToolCall(toolName, false)

	output = buf.String()
	if !strings.Contains(output, "Tool call failed") {
		t.Error("Expected failure message for failed tool call")
	}
}

func TestContextLogger_ConfigLoad(t *testing.T) {
	// Test successful config load
	var buf bytes.Buffer
	logger := &Logger{
		infoLogger: log.New(&buf, "[INFO] ", 0),
		mcpServer:  nil,
	}

	ctx := context.Background()
	component := "config"
	contextLogger := logger.WithContext(ctx, component)

	source := ".env"
	contextLogger.ConfigLoad(source, true)

	output := buf.String()
	if !strings.Contains(output, "Configuration loaded from") {
		t.Error("Expected success message for successful config load")
	}
	if !strings.Contains(output, source) {
		t.Errorf("Expected source '%s' in output", source)
	}

	// Test failed config load
	buf.Reset()
	logger.errorLogger = log.New(&buf, "[ERROR] ", 0)
	contextLogger.ConfigLoad(source, false)

	output = buf.String()
	if !strings.Contains(output, "Configuration load failed") {
		t.Error("Expected failure message for failed config load")
	}
}

func TestLogger_sendMCPLog(t *testing.T) {
	// Test that sendMCPLog doesn't panic with nil server
	logger := NewLogger(nil)

	// This should not panic
	logger.sendMCPLog("info", "test-logger", "test message")

	// Test with actual server (basic smoke test)
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithResourceCapabilities(false, false),
	)

	logger = NewLogger(s)

	// This should not panic
	logger.sendMCPLog("info", "test-logger", "test message")
}
