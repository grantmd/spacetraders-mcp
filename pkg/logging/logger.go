package logging

import (
	"context"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Logger provides structured logging for the SpaceTraders MCP server
type Logger struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	mcpServer   *server.MCPServer
}

// NewLogger creates a new logger instance
func NewLogger(mcpServer *server.MCPServer) *Logger {
	return &Logger{
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		infoLogger:  log.New(os.Stderr, "[INFO] ", log.LstdFlags),
		debugLogger: log.New(os.Stderr, "[DEBUG] ", log.LstdFlags),
		mcpServer:   mcpServer,
	}
}

// Info logs an informational message
func (l *Logger) Info(message string, args ...interface{}) {
	l.infoLogger.Printf(message, args...)

	// Also send to MCP client if available
	if l.mcpServer != nil {
		l.sendMCPLog(mcp.LoggingLevelInfo, "spacetraders-mcp", message)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	l.errorLogger.Printf(message, args...)

	// Also send to MCP client if available
	if l.mcpServer != nil {
		l.sendMCPLog(mcp.LoggingLevelError, "spacetraders-mcp", message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	l.debugLogger.Printf(message, args...)

	// Also send to MCP client if available
	if l.mcpServer != nil {
		l.sendMCPLog(mcp.LoggingLevelDebug, "spacetraders-mcp", message)
	}
}

// WithContext adds context information to log messages
func (l *Logger) WithContext(ctx context.Context, component string) *ContextLogger {
	return &ContextLogger{
		logger:    l,
		context:   ctx,
		component: component,
	}
}

// sendMCPLog sends a log message to the MCP client
func (l *Logger) sendMCPLog(level mcp.LoggingLevel, logger string, message string) {
	// Create logging notification
	notification := mcp.NewLoggingMessageNotification(level, logger, message)

	// Send notification to client (this would be handled by the server framework)
	// For now, we'll just ensure the structure is correct
	_ = notification
}

// ContextLogger provides logging with context information
type ContextLogger struct {
	logger    *Logger
	context   context.Context
	component string
}

// Info logs an informational message with context
func (cl *ContextLogger) Info(message string, args ...interface{}) {
	contextMessage := "[" + cl.component + "] " + message
	cl.logger.Info(contextMessage, args...)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(message string, args ...interface{}) {
	contextMessage := "[" + cl.component + "] " + message
	cl.logger.Error(contextMessage, args...)
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(message string, args ...interface{}) {
	contextMessage := "[" + cl.component + "] " + message
	cl.logger.Debug(contextMessage, args...)
}

// APICall logs an API call with timing information
func (cl *ContextLogger) APICall(endpoint string, statusCode int, duration string) {
	cl.Info("API call: %s -> %d (%s)", endpoint, statusCode, duration)
}

// ResourceRead logs a resource read operation
func (cl *ContextLogger) ResourceRead(uri string, success bool) {
	if success {
		cl.Info("Resource read successful: %s", uri)
	} else {
		cl.Error("Resource read failed: %s", uri)
	}
}

// ToolCall logs a tool call operation
func (cl *ContextLogger) ToolCall(toolName string, success bool) {
	if success {
		cl.Info("Tool call successful: %s", toolName)
	} else {
		cl.Error("Tool call failed: %s", toolName)
	}
}

// ConfigLoad logs configuration loading events
func (cl *ContextLogger) ConfigLoad(source string, success bool) {
	if success {
		cl.Info("Configuration loaded from: %s", source)
	} else {
		cl.Error("Configuration load failed from: %s", source)
	}
}
