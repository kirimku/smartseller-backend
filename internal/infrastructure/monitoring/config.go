package monitoring

import (
	"time"
)

// Config holds all monitoring configuration
type Config struct {
	Performance PerformanceConfig `yaml:"performance"`
	Alerts      AlertConfig       `yaml:"alerts"`
	Dashboard   DashboardConfig   `yaml:"dashboard"`
}

// PerformanceConfig configures query performance monitoring
type PerformanceConfig struct {
	Enabled            bool          `yaml:"enabled" env:"MONITORING_ENABLED" default:"true"`
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold" env:"SLOW_QUERY_THRESHOLD" default:"500ms"`
	MaxSlowQueries     int           `yaml:"max_slow_queries" env:"MAX_SLOW_QUERIES" default:"1000"`
	MetricsInterval    time.Duration `yaml:"metrics_interval" env:"METRICS_INTERVAL" default:"60s"`
	EnableStackTrace   bool          `yaml:"enable_stack_trace" env:"ENABLE_STACK_TRACE" default:"false"`
	LogSlowQueries     bool          `yaml:"log_slow_queries" env:"LOG_SLOW_QUERIES" default:"true"`
	QueryNormalization bool          `yaml:"query_normalization" env:"QUERY_NORMALIZATION" default:"true"`
}

// AlertConfig configures alerting thresholds
type AlertConfig struct {
	Enabled             bool          `yaml:"enabled" env:"ALERTS_ENABLED" default:"true"`
	SlowQueryAlertCount int           `yaml:"slow_query_alert_count" env:"SLOW_QUERY_ALERT_COUNT" default:"10"`
	ErrorRateThreshold  float64       `yaml:"error_rate_threshold" env:"ERROR_RATE_THRESHOLD" default:"5.0"`
	LatencyThreshold    time.Duration `yaml:"latency_threshold" env:"LATENCY_THRESHOLD" default:"1s"`
	AlertCooldown       time.Duration `yaml:"alert_cooldown" env:"ALERT_COOLDOWN" default:"5m"`
	TelegramEnabled     bool          `yaml:"telegram_enabled" env:"TELEGRAM_ALERTS_ENABLED" default:"false"`
	TelegramBotToken    string        `yaml:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID      string        `yaml:"telegram_chat_id" env:"TELEGRAM_CHAT_ID"`
	EmailEnabled        bool          `yaml:"email_enabled" env:"EMAIL_ALERTS_ENABLED" default:"false"`
	EmailRecipients     []string      `yaml:"email_recipients" env:"EMAIL_RECIPIENTS"`
}

// DashboardConfig configures the monitoring dashboard
type DashboardConfig struct {
	Enabled         bool   `yaml:"enabled" env:"DASHBOARD_ENABLED" default:"true"`
	Port            int    `yaml:"port" env:"DASHBOARD_PORT" default:"8081"`
	Path            string `yaml:"path" env:"DASHBOARD_PATH" default:"/monitoring"`
	RefreshInterval int    `yaml:"refresh_interval" env:"DASHBOARD_REFRESH" default:"30"`
	MaxQueryHistory int    `yaml:"max_query_history" env:"MAX_QUERY_HISTORY" default:"100"`
	AuthEnabled     bool   `yaml:"auth_enabled" env:"DASHBOARD_AUTH_ENABLED" default:"true"`
	Username        string `yaml:"username" env:"DASHBOARD_USERNAME" default:"admin"`
	Password        string `yaml:"password" env:"DASHBOARD_PASSWORD"`
}

// DefaultConfig returns default monitoring configuration
func DefaultConfig() Config {
	return Config{
		Performance: PerformanceConfig{
			Enabled:            true,
			SlowQueryThreshold: 500 * time.Millisecond,
			MaxSlowQueries:     1000,
			MetricsInterval:    60 * time.Second,
			EnableStackTrace:   false,
			LogSlowQueries:     true,
			QueryNormalization: true,
		},
		Alerts: AlertConfig{
			Enabled:             true,
			SlowQueryAlertCount: 10,
			ErrorRateThreshold:  5.0,
			LatencyThreshold:    1 * time.Second,
			AlertCooldown:       5 * time.Minute,
			TelegramEnabled:     false,
			EmailEnabled:        false,
		},
		Dashboard: DashboardConfig{
			Enabled:         true,
			Port:            8081,
			Path:            "/monitoring",
			RefreshInterval: 30,
			MaxQueryHistory: 100,
			AuthEnabled:     true,
			Username:        "admin",
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Performance.SlowQueryThreshold <= 0 {
		c.Performance.SlowQueryThreshold = 500 * time.Millisecond
	}

	if c.Performance.MaxSlowQueries <= 0 {
		c.Performance.MaxSlowQueries = 1000
	}

	if c.Performance.MetricsInterval <= 0 {
		c.Performance.MetricsInterval = 60 * time.Second
	}

	if c.Alerts.ErrorRateThreshold < 0 || c.Alerts.ErrorRateThreshold > 100 {
		c.Alerts.ErrorRateThreshold = 5.0
	}

	if c.Alerts.AlertCooldown <= 0 {
		c.Alerts.AlertCooldown = 5 * time.Minute
	}

	if c.Dashboard.Port <= 0 || c.Dashboard.Port > 65535 {
		c.Dashboard.Port = 8081
	}

	if c.Dashboard.RefreshInterval <= 0 {
		c.Dashboard.RefreshInterval = 30
	}

	if c.Dashboard.MaxQueryHistory <= 0 {
		c.Dashboard.MaxQueryHistory = 100
	}

	return nil
}
