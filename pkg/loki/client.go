package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kirimku/smartseller-backend/internal/config"
)

// LokiClient represents a client for sending logs to Grafana Loki
type LokiClient struct {
	endpoint  string
	username  string
	password  string
	client    *http.Client
	buffer    []LogEntry
	mutex     sync.Mutex
	batchSize int
	flushTime time.Duration
	stop      chan bool
}

// LogEntry represents a log entry for Loki
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Labels    map[string]string      `json:"labels"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// LokiPushRequest represents the request structure for Loki's push API
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a log stream in Loki
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// NewLokiClient creates a new Loki client
func NewLokiClient() *LokiClient {
	endpoint := os.Getenv("LOKI_URL")
	username := os.Getenv("LOKI_USERNAME")
	password := os.Getenv("LOKI_PASSWORD")
	enabled := os.Getenv("LOKI_ENABLED") == "true"

	if endpoint == "" || !enabled {
		// Return a no-op client if not configured or disabled
		return &LokiClient{
			client:    &http.Client{Timeout: 10 * time.Second},
			buffer:    make([]LogEntry, 0),
			batchSize: 100,
			flushTime: 10 * time.Second,
			stop:      make(chan bool),
		}
	}

	client := &LokiClient{
		endpoint:  endpoint,
		username:  username,
		password:  password,
		client:    &http.Client{Timeout: 10 * time.Second},
		buffer:    make([]LogEntry, 0),
		batchSize: 100,
		flushTime: 10 * time.Second,
		stop:      make(chan bool),
	}

	// Start background flusher
	go client.backgroundFlusher()

	return client
}

// Log sends a log entry to Loki
func (l *LokiClient) Log(level, message string, fields map[string]interface{}) {
	if l.endpoint == "" {
		return // No-op if not configured
	}

	labels := map[string]string{
		"job":         "kirimku-backend",
		"service":     "kirimku-backend",
		"environment": config.AppConfig.Environment,
		"level":       level,
		"host":        getHostname(),
	}

	// Add custom labels from fields
	if fields != nil {
		if userID, ok := fields["user_id"].(string); ok {
			labels["user_id"] = userID
		}
		if endpoint, ok := fields["endpoint"].(string); ok {
			labels["endpoint"] = endpoint
		}
		if method, ok := fields["method"].(string); ok {
			labels["method"] = method
		}
		if traceID, ok := fields["trace_id"].(string); ok {
			labels["trace_id"] = traceID
		}
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Labels:    labels,
		Fields:    fields,
	}

	l.mutex.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.batchSize
	l.mutex.Unlock()

	if shouldFlush {
		go l.flush()
	}
}

// LogWithContext logs with context information
func (l *LokiClient) LogWithContext(ctx context.Context, level, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	// Extract trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}

	// Extract user ID from context if available
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	l.Log(level, message, fields)
}

// Info logs an info message
func (l *LokiClient) Info(message string, fields map[string]interface{}) {
	l.Log("info", message, fields)
}

// Error logs an error message
func (l *LokiClient) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.Log("error", message, fields)
}

// Warn logs a warning message
func (l *LokiClient) Warn(message string, fields map[string]interface{}) {
	l.Log("warn", message, fields)
}

// Debug logs a debug message
func (l *LokiClient) Debug(message string, fields map[string]interface{}) {
	l.Log("debug", message, fields)
}

// flush sends buffered logs to Loki
func (l *LokiClient) flush() {
	if l.endpoint == "" {
		return
	}

	l.mutex.Lock()
	if len(l.buffer) == 0 {
		l.mutex.Unlock()
		return
	}

	entries := make([]LogEntry, len(l.buffer))
	copy(entries, l.buffer)
	l.buffer = l.buffer[:0] // Clear buffer
	l.mutex.Unlock()

	// Group entries by labels
	streams := make(map[string][]LogEntry)
	for _, entry := range entries {
		key := labelsToKey(entry.Labels)
		streams[key] = append(streams[key], entry)
	}

	// Convert to Loki format
	lokiStreams := make([]LokiStream, 0, len(streams))
	for key, entries := range streams {
		labels := keyToLabels(key, entries[0].Labels)
		values := make([][]string, len(entries))

		for i, entry := range entries {
			// Create log line with structured data
			logLine := map[string]interface{}{
				"level":   entry.Level,
				"message": entry.Message,
			}
			if entry.Fields != nil {
				logLine["fields"] = entry.Fields
			}

			logLineJSON, _ := json.Marshal(logLine)
			values[i] = []string{
				fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
				string(logLineJSON),
			}
		}

		lokiStreams = append(lokiStreams, LokiStream{
			Stream: labels,
			Values: values,
		})
	}

	request := LokiPushRequest{Streams: lokiStreams}
	l.sendToLoki(request)
}

// sendToLoki sends the request to Loki
func (l *LokiClient) sendToLoki(request LokiPushRequest) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling log data: %v\n", err)
		return
	}

	// Use the endpoint as-is, don't append the path since LOKI_URL already includes it
	req, err := http.NewRequest("POST", l.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if l.username != "" && l.password != "" {
		req.SetBasicAuth(l.username, l.password)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		fmt.Printf("Error sending logs to Loki: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Read response body for more details
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		fmt.Printf("Loki returned error status: %d, response: %s\n", resp.StatusCode, string(body[:n]))
		return
	}

	fmt.Printf("Successfully sent %d log entries to Loki (status: %d)\n", len(request.Streams), resp.StatusCode)
}

// backgroundFlusher flushes logs periodically
func (l *LokiClient) backgroundFlusher() {
	ticker := time.NewTicker(l.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.flush()
		case <-l.stop:
			l.flush() // Final flush
			return
		}
	}
}

// Close gracefully shuts down the Loki client
func (l *LokiClient) Close() {
	close(l.stop)
	time.Sleep(100 * time.Millisecond) // Allow final flush
}

// Helper functions
func labelsToKey(labels map[string]string) string {
	key, _ := json.Marshal(labels)
	return string(key)
}

func keyToLabels(key string, fallback map[string]string) map[string]string {
	var labels map[string]string
	if err := json.Unmarshal([]byte(key), &labels); err != nil {
		return fallback
	}
	return labels
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// Global Loki client instance
var globalLokiClient *LokiClient
var once sync.Once

// GetGlobalLokiClient returns the global Loki client instance
func GetGlobalLokiClient() *LokiClient {
	once.Do(func() {
		globalLokiClient = NewLokiClient()
	})
	return globalLokiClient
}
