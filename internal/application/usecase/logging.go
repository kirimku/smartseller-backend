package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity level of log messages
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogContext contains contextual information for logging
type LogContext struct {
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	RequestID  *string                `json:"request_id,omitempty"`
	Operation  string                 `json:"operation"`
	Entity     string                 `json:"entity"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Error      error                  `json:"error,omitempty"`
	StackTrace *string                `json:"stack_trace,omitempty"`
}

// UseCaseLogger provides structured logging for use case operations
type UseCaseLogger struct {
	logger  *slog.Logger
	useCase string
	enabled map[LogLevel]bool
}

// NewUseCaseLogger creates a new use case logger
func NewUseCaseLogger(logger *slog.Logger, useCase string) *UseCaseLogger {
	return &UseCaseLogger{
		logger:  logger.With(slog.String("use_case", useCase)),
		useCase: useCase,
		enabled: map[LogLevel]bool{
			LogLevelDebug: true,
			LogLevelInfo:  true,
			LogLevelWarn:  true,
			LogLevelError: true,
		},
	}
}

// SetLogLevel enables or disables a specific log level
func (l *UseCaseLogger) SetLogLevel(level LogLevel, enabled bool) {
	l.enabled[level] = enabled
}

// Debug logs a debug message
func (l *UseCaseLogger) Debug(ctx context.Context, message string, logCtx LogContext) {
	if !l.enabled[LogLevelDebug] {
		return
	}

	attrs := l.buildLogArgs(logCtx)
	l.logger.DebugContext(ctx, message, attrs...)
}

// Info logs an info message
func (l *UseCaseLogger) Info(ctx context.Context, message string, logCtx LogContext) {
	if !l.enabled[LogLevelInfo] {
		return
	}

	attrs := l.buildLogArgs(logCtx)
	l.logger.InfoContext(ctx, message, attrs...)
}

// Warn logs a warning message
func (l *UseCaseLogger) Warn(ctx context.Context, message string, logCtx LogContext) {
	if !l.enabled[LogLevelWarn] {
		return
	}

	attrs := l.buildLogArgs(logCtx)
	l.logger.WarnContext(ctx, message, attrs...)
}

// Error logs an error message
func (l *UseCaseLogger) Error(ctx context.Context, message string, logCtx LogContext) {
	if !l.enabled[LogLevelError] {
		return
	}

	// Include stack trace for errors
	if logCtx.Error != nil && logCtx.StackTrace == nil {
		stackTrace := getStackTrace(2) // Skip current and calling frame
		logCtx.StackTrace = &stackTrace
	}

	attrs := l.buildLogArgs(logCtx)
	l.logger.ErrorContext(ctx, message, attrs...)
}

// LogOperation logs the start and completion of an operation
func (l *UseCaseLogger) LogOperation(ctx context.Context, operation string, entity string, entityID *uuid.UUID, fn func() error) error {
	startTime := time.Now()

	logCtx := LogContext{
		Operation: operation,
		Entity:    entity,
		EntityID:  entityID,
	}

	l.Info(ctx, fmt.Sprintf("Starting %s", operation), logCtx)

	err := fn()
	duration := time.Since(startTime)
	logCtx.Duration = &duration

	if err != nil {
		logCtx.Error = err
		l.Error(ctx, fmt.Sprintf("Failed %s", operation), logCtx)
	} else {
		l.Info(ctx, fmt.Sprintf("Completed %s", operation), logCtx)
	}

	return err
}

// LogBusinessEvent logs important business events
func (l *UseCaseLogger) LogBusinessEvent(ctx context.Context, event string, entity string, entityID uuid.UUID, metadata map[string]interface{}) {
	logCtx := LogContext{
		Operation: event,
		Entity:    entity,
		EntityID:  &entityID,
		Metadata:  metadata,
	}

	l.Info(ctx, fmt.Sprintf("Business event: %s", event), logCtx)
}

// LogValidationError logs validation errors with detailed context
func (l *UseCaseLogger) LogValidationError(ctx context.Context, field string, value interface{}, rule string, entity string) {
	logCtx := LogContext{
		Operation: "validation",
		Entity:    entity,
		Metadata: map[string]interface{}{
			"field": field,
			"value": value,
			"rule":  rule,
		},
	}

	l.Warn(ctx, "Validation failed", logCtx)
}

// LogPerformanceMetrics logs performance metrics for operations
func (l *UseCaseLogger) LogPerformanceMetrics(ctx context.Context, operation string, duration time.Duration, metrics map[string]interface{}) {
	logCtx := LogContext{
		Operation: operation,
		Duration:  &duration,
		Metadata:  metrics,
	}

	// Log as warning if operation is slow
	threshold := 1 * time.Second
	if duration > threshold {
		logCtx.Metadata["slow_operation"] = true
		l.Warn(ctx, fmt.Sprintf("Slow operation detected: %s", operation), logCtx)
	} else {
		l.Debug(ctx, fmt.Sprintf("Performance metrics: %s", operation), logCtx)
	}
}

// LogSecurityEvent logs security-related events
func (l *UseCaseLogger) LogSecurityEvent(ctx context.Context, event string, userID *uuid.UUID, details map[string]interface{}) {
	logCtx := LogContext{
		Operation: event,
		UserID:    userID,
		Entity:    "security",
		Metadata:  details,
	}

	l.Warn(ctx, fmt.Sprintf("Security event: %s", event), logCtx)
}

// buildLogArgs builds slog arguments from LogContext
func (l *UseCaseLogger) buildLogArgs(logCtx LogContext) []any {
	var args []any

	if logCtx.UserID != nil {
		args = append(args, slog.String("user_id", logCtx.UserID.String()))
	}

	if logCtx.RequestID != nil {
		args = append(args, slog.String("request_id", *logCtx.RequestID))
	}

	if logCtx.Operation != "" {
		args = append(args, slog.String("operation", logCtx.Operation))
	}

	if logCtx.Entity != "" {
		args = append(args, slog.String("entity", logCtx.Entity))
	}

	if logCtx.EntityID != nil {
		args = append(args, slog.String("entity_id", logCtx.EntityID.String()))
	}

	if logCtx.Duration != nil {
		args = append(args, slog.Duration("duration", *logCtx.Duration))
	}

	if logCtx.Error != nil {
		args = append(args, slog.String("error", logCtx.Error.Error()))

		// Add error code if it's a UseCaseError
		if useCaseErr, ok := AsUseCaseError(logCtx.Error); ok {
			args = append(args, slog.String("error_code", string(useCaseErr.Code)))
			args = append(args, slog.Int("http_status", useCaseErr.HTTPStatus))
		}
	}

	if logCtx.StackTrace != nil {
		args = append(args, slog.String("stack_trace", *logCtx.StackTrace))
	}

	// Add metadata as individual arguments
	if logCtx.Metadata != nil {
		for key, value := range logCtx.Metadata {
			args = append(args, slog.Any(key, value))
		}
	}

	return args
}

// buildAttributes builds slog attributes from LogContext (kept for compatibility)
func (l *UseCaseLogger) buildAttributes(logCtx LogContext) []slog.Attr {
	var attrs []slog.Attr

	if logCtx.UserID != nil {
		attrs = append(attrs, slog.String("user_id", logCtx.UserID.String()))
	}

	if logCtx.RequestID != nil {
		attrs = append(attrs, slog.String("request_id", *logCtx.RequestID))
	}

	if logCtx.Operation != "" {
		attrs = append(attrs, slog.String("operation", logCtx.Operation))
	}

	if logCtx.Entity != "" {
		attrs = append(attrs, slog.String("entity", logCtx.Entity))
	}

	if logCtx.EntityID != nil {
		attrs = append(attrs, slog.String("entity_id", logCtx.EntityID.String()))
	}

	if logCtx.Duration != nil {
		attrs = append(attrs, slog.Duration("duration", *logCtx.Duration))
	}

	if logCtx.Error != nil {
		attrs = append(attrs, slog.String("error", logCtx.Error.Error()))

		// Add error code if it's a UseCaseError
		if useCaseErr, ok := AsUseCaseError(logCtx.Error); ok {
			attrs = append(attrs, slog.String("error_code", string(useCaseErr.Code)))
			attrs = append(attrs, slog.Int("http_status", useCaseErr.HTTPStatus))
		}
	}

	if logCtx.StackTrace != nil {
		attrs = append(attrs, slog.String("stack_trace", *logCtx.StackTrace))
	}

	// Add metadata as nested attributes
	if logCtx.Metadata != nil {
		for key, value := range logCtx.Metadata {
			attrs = append(attrs, slog.Any(key, value))
		}
	}

	return attrs
}

// getStackTrace returns a formatted stack trace
func getStackTrace(skip int) string {
	var trace string
	pc := make([]uintptr, 10)
	n := runtime.Callers(skip+2, pc)

	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}

	return trace
}

// WithContext creates a new logger with additional context
func (l *UseCaseLogger) WithContext(key string, value interface{}) *UseCaseLogger {
	newLogger := &UseCaseLogger{
		logger:  l.logger.With(slog.Any(key, value)),
		useCase: l.useCase,
		enabled: l.enabled,
	}
	return newLogger
}

// WithUserID creates a new logger with user context
func (l *UseCaseLogger) WithUserID(userID uuid.UUID) *UseCaseLogger {
	return l.WithContext("user_id", userID.String())
}

// WithRequestID creates a new logger with request context
func (l *UseCaseLogger) WithRequestID(requestID string) *UseCaseLogger {
	return l.WithContext("request_id", requestID)
}

// ProductLogger creates a logger specifically for product operations
func ProductLogger(logger *slog.Logger) *UseCaseLogger {
	return NewUseCaseLogger(logger, "product")
}

// CategoryLogger creates a logger specifically for category operations
func CategoryLogger(logger *slog.Logger) *UseCaseLogger {
	return NewUseCaseLogger(logger, "category")
}

// VariantLogger creates a logger specifically for variant operations
func VariantLogger(logger *slog.Logger) *UseCaseLogger {
	return NewUseCaseLogger(logger, "variant")
}

// ImageLogger creates a logger specifically for image operations
func ImageLogger(logger *slog.Logger) *UseCaseLogger {
	return NewUseCaseLogger(logger, "image")
}

// OrchestrationLogger creates a logger specifically for orchestration operations
func OrchestrationLogger(logger *slog.Logger) *UseCaseLogger {
	return NewUseCaseLogger(logger, "orchestration")
}

// LoggingMiddleware provides a middleware for automatic operation logging
type LoggingMiddleware struct {
	logger *UseCaseLogger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *UseCaseLogger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// WrapOperation wraps an operation with automatic logging
func (m *LoggingMiddleware) WrapOperation(operation string, entity string) func(ctx context.Context, entityID *uuid.UUID, fn func() error) error {
	return func(ctx context.Context, entityID *uuid.UUID, fn func() error) error {
		return m.logger.LogOperation(ctx, operation, entity, entityID, fn)
	}
}

// AuditLogger provides audit logging capabilities
type AuditLogger struct {
	logger *slog.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *slog.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger.With(slog.String("component", "audit")),
	}
}

// LogDataChange logs data changes for audit purposes
func (a *AuditLogger) LogDataChange(ctx context.Context, userID uuid.UUID, entity string, entityID uuid.UUID, operation string, oldData, newData map[string]interface{}) {
	a.logger.InfoContext(ctx, "Data change audit",
		slog.String("user_id", userID.String()),
		slog.String("entity", entity),
		slog.String("entity_id", entityID.String()),
		slog.String("operation", operation),
		slog.Any("old_data", oldData),
		slog.Any("new_data", newData),
		slog.Time("timestamp", time.Now()),
	)
}

// LogAccessAttempt logs access attempts for audit purposes
func (a *AuditLogger) LogAccessAttempt(ctx context.Context, userID *uuid.UUID, resource string, action string, allowed bool, reason string) {
	level := slog.LevelInfo
	if !allowed {
		level = slog.LevelWarn
	}

	var userIDStr string
	if userID != nil {
		userIDStr = userID.String()
	}

	a.logger.Log(ctx, level, "Access attempt audit",
		slog.String("user_id", userIDStr),
		slog.String("resource", resource),
		slog.String("action", action),
		slog.Bool("allowed", allowed),
		slog.String("reason", reason),
		slog.Time("timestamp", time.Now()),
	)
}

// MetricsLogger provides metrics logging for monitoring
type MetricsLogger struct {
	logger *slog.Logger
}

// NewMetricsLogger creates a new metrics logger
func NewMetricsLogger(logger *slog.Logger) *MetricsLogger {
	return &MetricsLogger{
		logger: logger.With(slog.String("component", "metrics")),
	}
}

// LogOperationMetrics logs operation metrics
func (m *MetricsLogger) LogOperationMetrics(ctx context.Context, operation string, duration time.Duration, success bool, metadata map[string]interface{}) {
	args := []any{
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.Bool("success", success),
		slog.Time("timestamp", time.Now()),
	}

	for key, value := range metadata {
		args = append(args, slog.Any(key, value))
	}

	m.logger.InfoContext(ctx, "Operation metrics", args...)
}

// LogResourceUsage logs resource usage metrics
func (m *MetricsLogger) LogResourceUsage(ctx context.Context, resource string, usage map[string]interface{}) {
	args := []any{
		slog.String("resource", resource),
		slog.Time("timestamp", time.Now()),
	}

	for key, value := range usage {
		args = append(args, slog.Any(key, value))
	}

	m.logger.InfoContext(ctx, "Resource usage", args...)
}
