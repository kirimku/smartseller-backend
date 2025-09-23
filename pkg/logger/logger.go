package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kirimku/smartseller-backend/pkg/loki"
)

var (
	// Default logger instance
	stdLogger = log.New(os.Stdout, "", log.LstdFlags)

	// Log level constants
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"

	// Current log level (default to INFO)
	currentLevel = LevelInfo

	// Log file handle
	logFile *os.File

	// Loki client
	lokiClient *loki.LokiClient

	// Enable Loki logging
	lokiEnabled = false
)

// Init initializes the logger with custom configuration
func Init(level string) {
	if level != "" {
		currentLevel = strings.ToUpper(level)
	}
	lokiClient = loki.GetGlobalLokiClient()
	lokiEnabled = true
}

// InitWithFile initializes the logger with custom configuration and file output
func InitWithFile(level string, filename string) error {
	if level != "" {
		currentLevel = strings.ToUpper(level)
	}

	if filename != "" {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}

		// Close previous log file if it exists
		if logFile != nil {
			logFile.Close()
		}

		logFile = file
		multiWriter := io.MultiWriter(os.Stdout, file)
		stdLogger = log.New(multiWriter, "", log.LstdFlags)
	}

	lokiClient = loki.GetGlobalLokiClient()
	lokiEnabled = true

	return nil
}

// Close closes the log file if it's open
func Close() {
	if logFile != nil {
		logFile.Close()
	}
	if lokiClient != nil {
		lokiClient.Close()
	}
}

// shouldLog determines if a message with the given level should be logged
func shouldLog(level string) bool {
	switch currentLevel {
	case LevelDebug:
		return true
	case LevelInfo:
		return level != LevelDebug
	case LevelWarn:
		return level != LevelDebug && level != LevelInfo
	case LevelError:
		return level == LevelError
	default:
		return true
	}
}

// logMessage formats and logs a message with the given level and key-value pairs
func logMessage(level string, msg string, keyvals ...interface{}) {
	// Get file and line information
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "unknown"
	if ok {
		// Extract just the file name, not the full path
		parts := strings.Split(file, "/")
		fileInfo = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
	}

	// Format key-value pairs
	var kvStr string
	fields := make(map[string]interface{})

	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = "MISSING_VALUE"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}

		// Format the value appropriately
		var vStr string
		switch val := v.(type) {
		case error:
			vStr = val.Error()
		case fmt.Stringer:
			vStr = val.String()
		default:
			vStr = fmt.Sprintf("%v", val)
		}

		kvStr += fmt.Sprintf(" %s=%s", k, vStr)

		// Store for Loki
		if kStr, ok := k.(string); ok {
			fields[kStr] = v
		}
	}

	// Add file info to fields
	fields["file"] = fileInfo

	// Log the message
	stdLogger.Printf("[%s] %s [%s]%s", level, msg, fileInfo, kvStr)

	// Log to Loki if enabled
	if lokiEnabled && lokiClient != nil {
		lokiClient.Log(strings.ToLower(level), msg, fields)
	}
}

// Debug logs a debug message with key-value pairs
func Debug(msg string, keyvals ...interface{}) {
	if shouldLog(LevelDebug) {
		logMessage(LevelDebug, msg, keyvals...)
	}
}

// Info logs an info message with key-value pairs
func Info(msg string, keyvals ...interface{}) {
	if shouldLog(LevelInfo) {
		logMessage(LevelInfo, msg, keyvals...)
	}
}

// Warn logs a warning message with key-value pairs
func Warn(msg string, keyvals ...interface{}) {
	if shouldLog(LevelWarn) {
		logMessage(LevelWarn, msg, keyvals...)
	}
}

// Error logs an error message with key-value pairs
func Error(msg string, keyvals ...interface{}) {
	if shouldLog(LevelError) {
		logMessage(LevelError, msg, keyvals...)
	}
}

// Context-aware logging functions

// DebugWithContext logs a debug message with context
func DebugWithContext(ctx context.Context, msg string, keyvals ...interface{}) {
	if shouldLog(LevelDebug) {
		logMessageWithContext(ctx, LevelDebug, msg, keyvals...)
	}
}

// InfoWithContext logs an info message with context
func InfoWithContext(ctx context.Context, msg string, keyvals ...interface{}) {
	if shouldLog(LevelInfo) {
		logMessageWithContext(ctx, LevelInfo, msg, keyvals...)
	}
}

// WarnWithContext logs a warning message with context
func WarnWithContext(ctx context.Context, msg string, keyvals ...interface{}) {
	if shouldLog(LevelWarn) {
		logMessageWithContext(ctx, LevelWarn, msg, keyvals...)
	}
}

// ErrorWithContext logs an error message with context
func ErrorWithContext(ctx context.Context, msg string, keyvals ...interface{}) {
	if shouldLog(LevelError) {
		logMessageWithContext(ctx, LevelError, msg, keyvals...)
	}
}

// logMessageWithContext formats and logs a message with context
func logMessageWithContext(ctx context.Context, level string, msg string, keyvals ...interface{}) {
	// Get file and line information
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "unknown"
	if ok {
		// Extract just the file name, not the full path
		parts := strings.Split(file, "/")
		fileInfo = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
	}

	// Format key-value pairs
	var kvStr string
	fields := make(map[string]interface{})

	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = "MISSING_VALUE"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}

		// Format the value appropriately
		var vStr string
		switch val := v.(type) {
		case error:
			vStr = val.Error()
		case fmt.Stringer:
			vStr = val.String()
		default:
			vStr = fmt.Sprintf("%v", val)
		}

		kvStr += fmt.Sprintf(" %s=%s", k, vStr)

		// Store for Loki
		if kStr, ok := k.(string); ok {
			fields[kStr] = v
		}
	}

	// Add file info to fields
	fields["file"] = fileInfo

	// Log the message
	stdLogger.Printf("[%s] %s [%s]%s", level, msg, fileInfo, kvStr)

	// Log to Loki if enabled
	if lokiEnabled && lokiClient != nil {
		lokiClient.LogWithContext(ctx, strings.ToLower(level), msg, fields)
	}
}

// Business event logging functions

// LogTransaction logs transaction-related events
func LogTransaction(ctx context.Context, transactionID, event, status string, amount float64, userID string) {
	InfoWithContext(ctx, fmt.Sprintf("Transaction event: %s", event),
		"transaction_id", transactionID,
		"event", event,
		"status", status,
		"amount", amount,
		"user_id", userID,
		"timestamp", time.Now().Unix(),
	)
}

// LogWalletOperation logs wallet-related operations
func LogWalletOperation(ctx context.Context, walletID, operation, status string, amount float64, userID string) {
	InfoWithContext(ctx, fmt.Sprintf("Wallet operation: %s", operation),
		"wallet_id", walletID,
		"operation", operation,
		"status", status,
		"amount", amount,
		"user_id", userID,
		"timestamp", time.Now().Unix(),
	)
}

// LogAPICall logs API call information
func LogAPICall(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration, userID string) {
	InfoWithContext(ctx, "API call",
		"method", method,
		"endpoint", endpoint,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"user_id", userID,
		"timestamp", time.Now().Unix(),
	)
}

// LogExternalAPICall logs external API call information
func LogExternalAPICall(ctx context.Context, service, method, endpoint string, statusCode int, duration time.Duration) {
	InfoWithContext(ctx, fmt.Sprintf("External API call: %s", service),
		"service", service,
		"method", method,
		"endpoint", endpoint,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"timestamp", time.Now().Unix(),
	)
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(ctx context.Context, event, userID, ipAddress string, details ...interface{}) {
	keyvals := []interface{}{
		"security_event", event,
		"user_id", userID,
		"ip_address", ipAddress,
		"timestamp", time.Now().Unix(),
	}
	keyvals = append(keyvals, details...)

	WarnWithContext(ctx, fmt.Sprintf("Security event: %s", event), keyvals...)
}

// LogBusinessEvent logs business-related events (payments, orders, etc.)
func LogBusinessEvent(ctx context.Context, event string, entityID, entityType string, details ...interface{}) {
	keyvals := []interface{}{
		"business_event", event,
		"entity_id", entityID,
		"entity_type", entityType,
		"timestamp", time.Now().Unix(),
	}
	keyvals = append(keyvals, details...)

	InfoWithContext(ctx, fmt.Sprintf("Business event: %s", event), keyvals...)
}
