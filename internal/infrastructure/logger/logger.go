package logger

import (
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// InitLogger initializes the global logger with proper configuration
func InitLogger() {
	// Configure log level from environment
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	level := zerolog.InfoLevel // default level
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure log format and output
	logFormat := strings.ToLower(os.Getenv("LOG_FORMAT"))

	// Setup file output in JSON format for ELK
	logFile := os.Getenv("LOG_FILE")
	var output io.Writer

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal().Err(err).Str("path", logFile).Msg("Failed to open log file")
		}

		// Configure console output with file
		if logFormat == "pretty" {
			output = zerolog.MultiLevelWriter(
				zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
					NoColor:    false,
				},
				file, // Also write JSON to file
			)
		} else {
			output = zerolog.MultiLevelWriter(file)
		}
	} else {
		// No file logging, only console
		if logFormat == "pretty" {
			output = zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
				NoColor:    false,
			}
		} else {
			output = os.Stdout
		}
	}

	// Initialize logger with ELK-friendly fields
	hostname, _ := os.Hostname()
	Logger = zerolog.New(output).With().
		Timestamp().
		Str("host", hostname).
		Str("environment", os.Getenv("APP_ENV")).
		Str("service", "kirimku-backend").
		Str("version", os.Getenv("APP_VERSION")).
		Caller().
		Logger()

	// Set standard time format for ELK
	zerolog.TimeFieldFormat = time.RFC3339Nano

	Logger.Info().
		Str("level", level.String()).
		Str("format", logFormat).
		Str("file", func() string {
			if logFile == "" {
				return "stdout"
			}
			return logFile
		}()).
		Msg("Logger initialized with ELK compatibility")
}

// RequestLogger adds common request fields to the logger in ELK-friendly format
func RequestLogger(r *http.Request) *zerolog.Event {
	return Logger.Info().
		Str("type", "request").
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_ip", r.RemoteAddr).
		Str("user_agent", r.UserAgent()).
		Str("request_id", r.Header.Get("X-Request-ID")).
		Str("referer", r.Referer()).
		Str("protocol", r.Proto)
}

// ErrorLogger adds error fields in ELK-friendly format
func ErrorLogger() *zerolog.Event {
	pc, file, line, _ := runtime.Caller(1)
	return Logger.Error().
		Str("type", "error").
		Str("file", file).
		Int("line", line).
		Str("function", runtime.FuncForPC(pc).Name())
}

// DebugLogger returns a debug level logger
func DebugLogger() *zerolog.Event {
	return Logger.Debug().
		Str("type", "debug").
		Str("component", "debug")
}

// WarnLogger returns a warning level logger
func WarnLogger() *zerolog.Event {
	return Logger.Warn().
		Str("type", "warn").
		Str("component", "warn")
}

// DBLogger adds database specific fields in ELK-friendly format
func DBLogger() *zerolog.Event {
	return Logger.Info().
		Str("type", "database").
		Str("component", "database")
}

// MetricsLogger adds metrics fields in ELK-friendly format
func MetricsLogger() *zerolog.Event {
	return Logger.Info().
		Str("type", "metrics").
		Str("component", "monitoring")
}

// AuditLogger adds audit fields in ELK-friendly format
func AuditLogger() *zerolog.Event {
	return Logger.Info().
		Str("type", "audit").
		Str("component", "audit")
}

// SecurityLogger adds security-related fields in ELK-friendly format
func SecurityLogger() *zerolog.Event {
	return Logger.Info().
		Str("type", "security").
		Str("component", "security")
}

// AuthLogger adds authentication-specific fields to the logger
func AuthLogger() *zerolog.Event {
	return Logger.Info().
		Str("type", "auth").
		Str("component", "auth")
}

func LogSessionError(r *http.Request, err error) {
	log.Printf("Session error: %v", err)
	log.Printf("Request details:")
	log.Printf("  Method: %s", r.Method)
	log.Printf("  Path: %s", r.URL.Path)
	log.Printf("  Host: %s", r.Host)
	log.Printf("  Origin: %s", r.Header.Get("Origin"))
	log.Printf("  User-Agent: %s", r.Header.Get("User-Agent"))
	log.Printf("  Cookie Headers: %v", r.Header["Cookie"])
}
