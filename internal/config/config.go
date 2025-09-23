package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v2"
)

// Config holds all configuration for the application
type Config struct {
	GoogleOAuthConfig *oauth2.Config
	SessionStore      *sessions.CookieStore
	Database          struct {
		URL          string
		MaxOpenConns int
		MaxIdleConns int
		MaxLifetime  time.Duration
	}
	Port           string
	BaseURL        string
	FrontendURL    string
	AllowedOrigins []string
	Environment    string
	Version        string
	RateLimit      struct {
		RequestsPerSecond float64
		ExpirationTTL     time.Duration
	}
	SessionConfig struct {
		Key      string
		MaxAge   int
		Secure   bool
		Domain   string
		SameSite http.SameSite
	}
	Payment struct {
		Gateway           string
		XenditApiKey      string
		XenditBaseURL     string
		XenditCallbackURL string
		// DurianPay configuration
		DurianPay DurianPayConfig
		// Legacy DurianPay fields for backward compatibility
		DurianPayApiKey       string
		DurianPayBaseURL      string
		DurianPayCallbackURL  string
		DurianPayClientKey    string
		DurianPayClientSecret string
		DurianPayMerchantID   string
		DurianPayPrivateKey   string
		InvoiceExpiryHours    int
		DefaultEWalletChannel string
	}
	// Location data structures
	LocationData struct {
		POSTCODES        map[string]map[string]map[string][]string
		DISTRICTS        map[string]map[string][]string
		DISTRICT_NAMES   map[string]map[string]string
		POSTCODE_MAPPING map[string][]map[string]string
		CITIES           map[string][]string
		PROVINCES        map[string]string
	}
	// JNT API Configuration
	JNTConfig struct {
		APIURL         string `env:"JNT_API_URL" envDefault:"https://api.jnt.co.id"`
		APIKey         string `env:"JNT_API_KEY" envDefault:""`
		Username       string `env:"JNT_USERNAME" envDefault:"KIRIMKU"`
		APIKeyName     string `env:"JNT_API_KEY_NAME" envDefault:""`
		OrderAPIURL    string `env:"JNT_ORDER_API_URL" envDefault:""`
		TariffAPIURL   string `env:"JNT_TARIFF_API_URL" envDefault:""`
		TariffAPIKey   string `env:"JNT_TARIFF_API_KEY" envDefault:""`
		TrackingAPIURL string `env:"JNT_TRACKING_API_URL" envDefault:""`
		CustName       string `env:"JNT_CUSTOMER_NAME" envDefault:"KIRIMKU"`
	}
	// JNE Config
	JNEAPIURL   string
	JNEUsername string
	JNEAPIKey   string

	// JNE mapping data
	JNEMapping                map[string]map[string]map[string]map[string]interface{} // Province -> City -> District -> Codes
	JNEJawaRegionMapping      map[string][]string                                     // jawa/non_jawa classification
	JNEWhitelistOriginMapping map[string][]string                                     // city code -> allowed cities

	// JNT mapping data - new structure
	JNTMapping map[string]map[string]map[string]map[string]interface{} // Province -> City -> District -> Codes

	// SiCepat mapping data
	SiCepatMappingOrigin      map[string]map[string]map[string]map[string]interface{} // Province -> City -> District -> Codes
	SiCepatMappingDestination map[string]map[string]map[string]map[string]interface{} // Province -> City -> District -> Codes

	// SAPX configuration
	SAPXConfig struct {
		APIURL             string        `env:"SAPX_API_URL"`
		APITrackerURL      string        `env:"SAPX_API_TRACKER_URL"`
		APIKeyPickup       string        `env:"SAPX_API_KEY_PICKUP"`
		APIKeyDropoff      string        `env:"SAPX_API_KEY_DROPOFF"`
		CustomerCodeNonCOD string        `env:"SAPX_CUSTOMER_CODE_NON_COD"`
		CustomerCodeCOD    string        `env:"SAPX_CUSTOMER_CODE_COD"`
		Timeout            time.Duration `env:"SAPX_TIMEOUT" envDefault:"30s"`
		MaxRetries         int           `env:"SAPX_MAX_RETRIES" envDefault:"3"`
	}

	// SAPX mapping data
	SAPXMapping map[string]map[string]map[string]map[string]interface{} // Province -> City -> District -> Codes

	// Mailgun configuration
	MailgunConfig struct {
		Domain    string `env:"MAILGUN_DOMAIN" required:"true"`
		APIKey    string `env:"MAILGUN_API_KEY" required:"true"`
		FromName  string `env:"SMTP_FROM_NAME"`
		FromEmail string `env:"SMTP_FROM_EMAIL"`
	}

	// Application-specific configuration
	App struct {
		WeightDiscrepancyThreshold float64 // Threshold for weight discrepancy in kg
		FeeDiscrepancyThreshold    float64 // Threshold for fee discrepancy in currency units
		AutoSettlementThreshold    float64 // Maximum amount for auto-settlement in currency units
		MaxDebtLimit               float64 // Maximum debt limit per user in currency units
	}

	// Logging configuration
	LogLevel string
	LogFile  string

	// Monitoring and Observability configuration
	Monitoring struct {
		// Loki configuration
		LokiURL       string `env:"LOKI_URL"`
		LokiUsername  string `env:"LOKI_USERNAME"`
		LokiPassword  string `env:"LOKI_PASSWORD"`
		LokiEnabled   bool   `env:"LOKI_ENABLED" envDefault:"false"`
		LokiBatchSize int    `env:"LOKI_BATCH_SIZE" envDefault:"100"`
		LokiBatchWait string `env:"LOKI_BATCH_WAIT" envDefault:"1s"`
		LokiLabels    string `env:"LOKI_LABELS" envDefault:"service=kirimku-backend"`

		// Prometheus metrics configuration
		MetricsEnabled   bool   `env:"METRICS_ENABLED" envDefault:"true"`
		MetricsNamespace string `env:"METRICS_NAMESPACE" envDefault:"kirimku"`
		MetricsSubsystem string `env:"METRICS_SUBSYSTEM" envDefault:"backend"`
	}

	// Telegram Alerting configuration
	Telegram struct {
		Enabled    bool          `env:"TELEGRAM_ENABLED" envDefault:"false"`
		BotToken   string        `env:"TELEGRAM_BOT_TOKEN"`
		ChatIDs    []string      `env:"TELEGRAM_CHAT_IDS"`                       // Comma-separated list of chat IDs
		AlertLevel string        `env:"TELEGRAM_ALERT_LEVEL" envDefault:"error"` // error, warn, info
		Timeout    time.Duration `env:"TELEGRAM_TIMEOUT" envDefault:"10s"`
	}

	// Telegram Refund Alerting configuration (separate bot for refund alerts)
	TelegramRefund struct {
		Enabled    bool          `env:"TELEGRAM_REFUND_ENABLED" envDefault:"false"`
		BotToken   string        `env:"TELEGRAM_REFUND_BOT_TOKEN"`
		ChatIDs    []string      `env:"TELEGRAM_REFUND_CHAT_IDS"`                      // Comma-separated list of chat IDs
		AlertLevel string        `env:"TELEGRAM_REFUND_ALERT_LEVEL" envDefault:"info"` // error, warn, info
		Timeout    time.Duration `env:"TELEGRAM_REFUND_TIMEOUT" envDefault:"10s"`
	}
	// SiCepat API Configuration
	SiCepatConfig struct {
		APIKey         string `env:"SICEPAT_API_KEY"`
		TrackingAPIKey string `env:"SICEPAT_API_TRACKING_KEY"`
		BaseURL        string `env:"SICEPAT_API_URL"`
		PickupURL      string `env:"SICEPAT_PICKUP_URL"`
		// Resi Number Range Configuration
		ResiRangeStart string `env:"SICEPAT_RESI_RANGE_START"`
		ResiRangeEnd   string `env:"SICEPAT_RESI_RANGE_END"`
	}
}

// DurianPayConfig holds DurianPay-specific configuration
type DurianPayConfig struct {
	// Basic Auth Configuration (Legacy API)
	APIKey      string `env:"DURIANPAY_API_KEY"`
	BaseURL     string `env:"DURIANPAY_BASE_URL"`
	CallbackURL string `env:"DURIANPAY_CALLBACK_URL"`

	// SNAP API Configuration (Advanced Features)
	ClientKey    string `env:"DURIANPAY_CLIENT_KEY"`
	ClientSecret string `env:"DURIANPAY_CLIENT_SECRET"`
	MerchantID   string `env:"DURIANPAY_MERCHANT_ID"`
	PrivateKey   string `env:"DURIANPAY_PRIVATE_KEY"`
	PublicKey    string `env:"DURIANPAY_PUBLIC_KEY"`

	// Environment Configuration
	Environment string `env:"DURIANPAY_ENVIRONMENT" envDefault:"sandbox"` // sandbox/production

	// Feature Flags
	EnableSNAP bool `env:"DURIANPAY_ENABLE_SNAP" envDefault:"false"`

	// Timeouts and Limits
	Timeout    time.Duration `env:"DURIANPAY_TIMEOUT" envDefault:"30s"`
	MaxRetries int           `env:"DURIANPAY_MAX_RETRIES" envDefault:"3"`

	// Webhook Configuration
	WebhookURL    string `env:"DURIANPAY_WEBHOOK_URL"`
	WebhookSecret string `env:"DURIANPAY_WEBHOOK_SECRET"`

	// Legacy fields for backward compatibility
	DurianPayApiKey       string `env:"DURIANPAY_API_KEY"`       // Deprecated: use APIKey
	DurianPayBaseURL      string `env:"DURIANPAY_BASE_URL"`      // Deprecated: use BaseURL
	DurianPayCallbackURL  string `env:"DURIANPAY_CALLBACK_URL"`  // Deprecated: use CallbackURL
	DurianPayClientKey    string `env:"DURIANPAY_CLIENT_KEY"`    // Deprecated: use ClientKey
	DurianPayClientSecret string `env:"DURIANPAY_CLIENT_SECRET"` // Deprecated: use ClientSecret
	DurianPayMerchantID   string `env:"DURIANPAY_MERCHANT_ID"`   // Deprecated: use MerchantID
	DurianPayPrivateKey   string `env:"DURIANPAY_PRIVATE_KEY"`   // Deprecated: use PrivateKey
}

// LocationData contains location information
type LocationData struct {
	PROVINCES map[string]string              // City -> Province
	DISTRICTS map[string]map[string][]string // Province -> City -> []District
}

var AppConfig Config

// LoadConfig initializes the application configuration
func LoadConfig() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Load and log Mailgun configuration
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	smtpFromName := os.Getenv("SMTP_FROM_NAME")
	smtpFromEmail := os.Getenv("SMTP_FROM_EMAIL")

	// Store in AppConfig
	AppConfig.MailgunConfig.Domain = mailgunDomain
	AppConfig.MailgunConfig.APIKey = mailgunAPIKey
	AppConfig.MailgunConfig.FromName = smtpFromName
	AppConfig.MailgunConfig.FromEmail = smtpFromEmail

	// Mask API key for secure logging
	maskedAPIKey := "not set"
	if mailgunAPIKey != "" {
		if len(mailgunAPIKey) > 10 {
			maskedAPIKey = mailgunAPIKey[:6] + "..." + mailgunAPIKey[len(mailgunAPIKey)-4:]
		} else {
			maskedAPIKey = "[set but too short]"
		}
	}

	log.Printf("EMAIL CONFIG - Mailgun Domain: %s", mailgunDomain)
	log.Printf("EMAIL CONFIG - Mailgun API Key: %s", maskedAPIKey)
	log.Printf("EMAIL CONFIG - From Name: %s", smtpFromName)
	log.Printf("EMAIL CONFIG - From Email: %s", smtpFromEmail)

	// Set environment
	AppConfig.Environment = getEnvWithDefault("APP_ENV", "development")
	AppConfig.Version = getEnvWithDefault("APP_VERSION", "1.0.0")

	// Configure logging
	AppConfig.LogLevel = getEnvWithDefault("LOG_LEVEL", "info")
	AppConfig.LogFile = getEnvWithDefault("LOG_FILE", "")

	// Configure Google OAuth
	AppConfig.GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  getEnvWithDefault("GOOGLE_REDIRECT_URL", "http://localhost:5173/auth/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Configure session
	AppConfig.SessionConfig = struct {
		Key      string
		MaxAge   int
		Secure   bool
		Domain   string
		SameSite http.SameSite
	}{
		Key:      os.Getenv("SESSION_KEY"),
		MaxAge:   getEnvAsInt("SESSION_MAX_AGE", 86400),
		Secure:   AppConfig.Environment == "production" || getEnvAsBool("SESSION_SECURE", false),
		Domain:   getEnvWithDefault("SESSION_DOMAIN", ""),
		SameSite: getSameSiteMode(getEnvWithDefault("SESSION_SAME_SITE", "lax")),
	}

	// Initialize session store with consistent configuration
	AppConfig.SessionStore = sessions.NewCookieStore([]byte(AppConfig.SessionConfig.Key))
	AppConfig.SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   AppConfig.SessionConfig.MaxAge,
		HttpOnly: true,
		Secure:   AppConfig.SessionConfig.Secure,
		SameSite: AppConfig.SessionConfig.SameSite,
		Domain:   AppConfig.SessionConfig.Domain,
	}

	// Configure database
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbName == "" {
		dbName = "kirimku"
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	// Use standard connection string format with explicit parameters
	AppConfig.Database.URL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	// Configure database pool
	AppConfig.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	AppConfig.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	maxLifetime := getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)
	AppConfig.Database.MaxLifetime = time.Duration(maxLifetime) * time.Second

	// Configure server port
	AppConfig.Port = getPort()

	// Configure URLs
	AppConfig.BaseURL = os.Getenv("BASE_URL")
	if AppConfig.BaseURL == "" {
		AppConfig.BaseURL = "http://localhost:" + AppConfig.Port
	}

	AppConfig.FrontendURL = os.Getenv("FRONTEND_URL")
	if AppConfig.FrontendURL == "" {
		AppConfig.FrontendURL = "http://localhost:3000"
	}

	// Configure allowed origins for CORS
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		AppConfig.AllowedOrigins = []string{AppConfig.FrontendURL}
	} else {
		AppConfig.AllowedOrigins = strings.Split(allowedOrigins, ",")
	}

	// Configure rate limiter
	requestsPerSecond := os.Getenv("RATE_LIMIT_PER_SECOND")
	if requestsPerSecond == "" {
		AppConfig.RateLimit.RequestsPerSecond = 1.0
	} else {
		rps, err := strconv.ParseFloat(requestsPerSecond, 64)
		if err != nil {
			return fmt.Errorf("invalid RATE_LIMIT_PER_SECOND value: %v", err)
		}
		AppConfig.RateLimit.RequestsPerSecond = rps
	}

	ttl := os.Getenv("RATE_LIMIT_TTL")
	if ttl == "" {
		AppConfig.RateLimit.ExpirationTTL = time.Hour
	} else {
		duration, err := time.ParseDuration(ttl)
		if err != nil {
			return fmt.Errorf("invalid RATE_LIMIT_TTL value: %v", err)
		}
		AppConfig.RateLimit.ExpirationTTL = duration
	}

	// Load location data with the correct path
	configDir := getEnvWithDefault("CONFIG_DIR", "./internal/config")
	if err := loadLocationData(configDir); err != nil {
		log.Printf("Warning: Failed to load location data: %v", err)
	}

	// Load JNE configuration
	AppConfig.JNEAPIURL = getEnvWithDefault("JNE_API_URL", "https://api.jne.co.id")
	AppConfig.JNEUsername = getEnvWithDefault("JNE_USERNAME", "")
	AppConfig.JNEAPIKey = getEnvWithDefault("JNE_API_KEY", "")

	// Load JNE mapping
	mappingPath := filepath.Join("internal", "config", "mapping_area", "jne_kirimku_mapping.yaml")
	if err := loadJNEMapping(mappingPath, &AppConfig); err != nil {
		return err
	}

	// Load JNE Jawa/Non-Jawa region mapping
	regionMappingPath := filepath.Join("internal", "config", "jne_code_mapping_jawa_and_non_jawa.yaml")
	if err := loadJNEJawaRegionMapping(regionMappingPath, &AppConfig); err != nil {
		return err
	}

	// Load JNE whitelist origin mapping
	whitelistMappingPath := filepath.Join("internal", "config", "jne_whitelist_origin_mapping.yaml")
	if err := loadJNEWhitelistOriginMapping(whitelistMappingPath, &AppConfig); err != nil {
		return err
	}

	// Load JNT mapping
	mappingPath = filepath.Join("internal", "config", "jnt_code_mapping_new.yaml")
	if err := loadJNTMapping(mappingPath, &AppConfig); err != nil {
		return err
	}

	// Configure JNT settings
	AppConfig.JNTConfig.APIURL = getEnvWithDefault("JNT_API_URL", "https://api.jnt.co.id")
	AppConfig.JNTConfig.APIKey = getEnvWithDefault("JNT_API_KEY", "")
	AppConfig.JNTConfig.Username = getEnvWithDefault("JNT_USERNAME", "KIRIMKU")
	AppConfig.JNTConfig.APIKeyName = getEnvWithDefault("JNT_API_KEY_NAME", "")
	AppConfig.JNTConfig.OrderAPIURL = getEnvWithDefault("JNT_ORDER_API_URL", "")
	AppConfig.JNTConfig.TariffAPIURL = getEnvWithDefault("JNT_TARIFF_API_URL", "")
	AppConfig.JNTConfig.TariffAPIKey = getEnvWithDefault("JNT_TARIFF_API_KEY", "")
	AppConfig.JNTConfig.TrackingAPIURL = getEnvWithDefault("JNT_TRACKING_API_URL", "")
	AppConfig.JNTConfig.CustName = getEnvWithDefault("JNT_CUSTOMER_NAME", "KIRIMKU")

	// Load SiCepat mapping
	mappingPath = filepath.Join("internal", "config", "mapping_area", "sicepat_kirimku_mapping_origin.yaml")
	if err := loadSiCepatMapping(mappingPath, &AppConfig); err != nil {
		return err
	}

	// Load SiCepat destination mapping
	mappingPath = filepath.Join("internal", "config", "mapping_area", "sicepat_kirimku_mapping_destination.yaml")
	if err := loadSiCepatDestinationMapping(mappingPath, &AppConfig); err != nil {
		return err
	}

	// Configure SAPX settings
	AppConfig.SAPXConfig.APIURL = getEnvWithDefault("SAPX_API_URL", "https://api.coresyssap.com")
	AppConfig.SAPXConfig.APITrackerURL = getEnvWithDefault("SAPX_API_TRACKER_URL", "https://track.coresyssap.com")
	AppConfig.SAPXConfig.CustomerCodeNonCOD = getEnvWithDefault("SAPX_CUSTOMER_CODE_NON_COD", "")
	AppConfig.SAPXConfig.CustomerCodeCOD = getEnvWithDefault("SAPX_CUSTOMER_CODE_COD", "")
	AppConfig.SAPXConfig.APIKeyPickup = getEnvWithDefault("SAPX_API_KEY_PICKUP", "")
	AppConfig.SAPXConfig.APIKeyDropoff = getEnvWithDefault("SAPX_API_KEY_DROPOFF", "")
	AppConfig.SAPXConfig.Timeout = getEnvAsDuration("SAPX_TIMEOUT", 30*time.Second)
	AppConfig.SAPXConfig.MaxRetries = getEnvAsInt("SAPX_MAX_RETRIES", 3)

	// TODO: Update SAPX coverage mapping until all of area are valid
	// Load SAPX mapping
	mappingPath = filepath.Join("internal", "config", "sapx_coverage_mapping.yaml") // TODO: Update filename
	if err := loadSAPXMapping(mappingPath, &AppConfig); err != nil {
		return err
	}

	// Configure payment settings
	AppConfig.Payment.Gateway = getEnvWithDefault("PAYMENT_GATEWAY", "xendit")
	AppConfig.Payment.XenditApiKey = getEnvWithDefault("XENDIT_API_KEY", "")
	AppConfig.Payment.XenditBaseURL = getEnvWithDefault("XENDIT_BASE_URL", "https://api.xendit.co")
	AppConfig.Payment.XenditCallbackURL = getEnvWithDefault("XENDIT_CALLBACK_URL", fmt.Sprintf("%s/webhooks/payments/xendit", AppConfig.BaseURL))

	// Configure DurianPay settings
	AppConfig.Payment.DurianPay.APIKey = getEnvWithDefault("DURIANPAY_API_KEY", "")
	AppConfig.Payment.DurianPay.BaseURL = getEnvWithDefault("DURIANPAY_BASE_URL", "https://api-sandbox.durianpay.id")
	AppConfig.Payment.DurianPay.CallbackURL = getEnvWithDefault("DURIANPAY_CALLBACK_URL", fmt.Sprintf("%s/webhooks/payments/durianpay", AppConfig.BaseURL))
	AppConfig.Payment.DurianPay.ClientKey = getEnvWithDefault("DURIANPAY_CLIENT_KEY", "")
	AppConfig.Payment.DurianPay.ClientSecret = getEnvWithDefault("DURIANPAY_CLIENT_SECRET", "")
	AppConfig.Payment.DurianPay.MerchantID = getEnvWithDefault("DURIANPAY_MERCHANT_ID", "")
	AppConfig.Payment.DurianPay.PrivateKey = getEnvWithDefault("DURIANPAY_PRIVATE_KEY", "")
	AppConfig.Payment.DurianPay.PublicKey = getEnvWithDefault("DURIANPAY_PUBLIC_KEY", "")
	AppConfig.Payment.DurianPay.Environment = getEnvWithDefault("DURIANPAY_ENVIRONMENT", "sandbox")
	AppConfig.Payment.DurianPay.EnableSNAP = getEnvAsBool("DURIANPAY_ENABLE_SNAP", false)
	AppConfig.Payment.DurianPay.Timeout = getEnvAsDuration("DURIANPAY_TIMEOUT", 30*time.Second)
	AppConfig.Payment.DurianPay.MaxRetries = getEnvAsInt("DURIANPAY_MAX_RETRIES", 3)
	AppConfig.Payment.DurianPay.WebhookURL = getEnvWithDefault("DURIANPAY_WEBHOOK_URL", fmt.Sprintf("%s/webhooks/payments/durianpay", AppConfig.BaseURL))
	AppConfig.Payment.DurianPay.WebhookSecret = getEnvWithDefault("DURIANPAY_WEBHOOK_SECRET", "")

	// Set legacy fields for backward compatibility
	AppConfig.Payment.DurianPay.DurianPayApiKey = AppConfig.Payment.DurianPay.APIKey
	AppConfig.Payment.DurianPay.DurianPayBaseURL = AppConfig.Payment.DurianPay.BaseURL
	AppConfig.Payment.DurianPay.DurianPayCallbackURL = AppConfig.Payment.DurianPay.CallbackURL
	AppConfig.Payment.DurianPay.DurianPayClientKey = AppConfig.Payment.DurianPay.ClientKey
	AppConfig.Payment.DurianPay.DurianPayClientSecret = AppConfig.Payment.DurianPay.ClientSecret
	AppConfig.Payment.DurianPay.DurianPayMerchantID = AppConfig.Payment.DurianPay.MerchantID
	AppConfig.Payment.DurianPay.DurianPayPrivateKey = AppConfig.Payment.DurianPay.PrivateKey

	// Legacy fields on Payment struct
	AppConfig.Payment.DurianPayApiKey = AppConfig.Payment.DurianPay.APIKey
	AppConfig.Payment.DurianPayBaseURL = AppConfig.Payment.DurianPay.BaseURL
	AppConfig.Payment.DurianPayCallbackURL = AppConfig.Payment.DurianPay.CallbackURL
	AppConfig.Payment.DurianPayClientKey = AppConfig.Payment.DurianPay.ClientKey
	AppConfig.Payment.DurianPayClientSecret = AppConfig.Payment.DurianPay.ClientSecret
	AppConfig.Payment.DurianPayMerchantID = AppConfig.Payment.DurianPay.MerchantID
	AppConfig.Payment.DurianPayPrivateKey = AppConfig.Payment.DurianPay.PrivateKey

	AppConfig.Payment.InvoiceExpiryHours = getEnvAsInt("INVOICE_EXPIRY_HOURS", 24)
	AppConfig.Payment.DefaultEWalletChannel = getEnvWithDefault("DEFAULT_EWALLET_CHANNEL", "SHOPEEPAY")

	// Configure app-specific settings
	AppConfig.App.WeightDiscrepancyThreshold = getEnvAsFloat("WEIGHT_DISCREPANCY_THRESHOLD", 0.1) // Default 0.1 kg
	AppConfig.App.FeeDiscrepancyThreshold = getEnvAsFloat("FEE_DISCREPANCY_THRESHOLD", 1000.0)    // Default 1000 currency units
	AppConfig.App.AutoSettlementThreshold = getEnvAsFloat("AUTO_SETTLEMENT_THRESHOLD", 10000.0)   // Default 10000 currency units
	AppConfig.App.MaxDebtLimit = getEnvAsFloat("MAX_DEBT_LIMIT", 100000.0)                        // Default 100000 currency units

	// Configure monitoring and observability settings
	AppConfig.Monitoring.LokiURL = getEnvWithDefault("LOKI_URL", "")
	AppConfig.Monitoring.LokiUsername = getEnvWithDefault("LOKI_USERNAME", "")
	AppConfig.Monitoring.LokiPassword = getEnvWithDefault("LOKI_PASSWORD", "")
	AppConfig.Monitoring.LokiEnabled = getEnvAsBool("LOKI_ENABLED", false)
	AppConfig.Monitoring.LokiBatchSize = getEnvAsInt("LOKI_BATCH_SIZE", 100)
	AppConfig.Monitoring.LokiBatchWait = getEnvWithDefault("LOKI_BATCH_WAIT", "1s")
	AppConfig.Monitoring.LokiLabels = getEnvWithDefault("LOKI_LABELS", "service=kirimku-backend")
	AppConfig.Monitoring.MetricsEnabled = getEnvAsBool("METRICS_ENABLED", true)
	AppConfig.Monitoring.MetricsNamespace = getEnvWithDefault("METRICS_NAMESPACE", "kirimku")
	AppConfig.Monitoring.MetricsSubsystem = getEnvWithDefault("METRICS_SUBSYSTEM", "backend")

	// Configure Telegram alerting
	AppConfig.Telegram.Enabled = getEnvAsBool("TELEGRAM_ENABLED", false)
	AppConfig.Telegram.BotToken = getEnvWithDefault("TELEGRAM_BOT_TOKEN", "")
	AppConfig.Telegram.ChatIDs = getEnvAsSlice("TELEGRAM_CHAT_IDS", ",")
	AppConfig.Telegram.AlertLevel = getEnvWithDefault("TELEGRAM_ALERT_LEVEL", "error")
	AppConfig.Telegram.Timeout = getEnvAsDuration("TELEGRAM_TIMEOUT", 10*time.Second)

	// Configure Telegram refund alerting (separate bot)
	AppConfig.TelegramRefund.Enabled = getEnvAsBool("TELEGRAM_REFUND_ENABLED", false)
	AppConfig.TelegramRefund.BotToken = getEnvWithDefault("TELEGRAM_REFUND_BOT_TOKEN", "")
	AppConfig.TelegramRefund.ChatIDs = getEnvAsSlice("TELEGRAM_REFUND_CHAT_IDS", ",")
	AppConfig.TelegramRefund.AlertLevel = getEnvWithDefault("TELEGRAM_REFUND_ALERT_LEVEL", "info")
	AppConfig.TelegramRefund.Timeout = getEnvAsDuration("TELEGRAM_REFUND_TIMEOUT", 10*time.Second)

	// Validate DurianPay configuration
	if err := AppConfig.ValidateDurianPayConfig(); err != nil {
		return fmt.Errorf("invalid DurianPay configuration: %v", err)
	}

	// Load SiCepat configuration
	AppConfig.SiCepatConfig.APIKey = getEnvWithDefault("SICEPAT_API_KEY", "")
	AppConfig.SiCepatConfig.TrackingAPIKey = getEnvWithDefault("SICEPAT_API_TRACKING_KEY", "")
	AppConfig.SiCepatConfig.BaseURL = getEnvWithDefault("SICEPAT_API_URL", "https://api.sicepat.com")
	AppConfig.SiCepatConfig.PickupURL = getEnvWithDefault("SICEPAT_PICKUP_URL", "https://pickup.sicepat.com")
	AppConfig.SiCepatConfig.ResiRangeStart = getEnvWithDefault("SICEPAT_RESI_RANGE_START", "100000000000")
	AppConfig.SiCepatConfig.ResiRangeEnd = getEnvWithDefault("SICEPAT_RESI_RANGE_END", "100000009999")

	return nil
}

// loadLocationData loads and processes the cities.yaml file
func loadLocationData(configDir string) error {
	// Initialize location data maps
	AppConfig.LocationData.POSTCODES = make(map[string]map[string]map[string][]string)
	AppConfig.LocationData.DISTRICTS = make(map[string]map[string][]string)
	AppConfig.LocationData.DISTRICT_NAMES = make(map[string]map[string]string)
	AppConfig.LocationData.POSTCODE_MAPPING = make(map[string][]map[string]string)
	AppConfig.LocationData.CITIES = make(map[string][]string)
	AppConfig.LocationData.PROVINCES = make(map[string]string)

	// Get absolute path of config directory
	absConfigDir, err := filepath.Abs(configDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Construct yaml path
	yamlPath := filepath.Join(absConfigDir, "cities.yaml")
	log.Printf("Attempting to load cities.yaml from: %s", yamlPath)

	// Check if file exists
	if _, err = os.Stat(yamlPath); os.IsNotExist(err) {
		// If not found in the default location, try the internal config path
		projectRoot := filepath.Dir(filepath.Dir(absConfigDir)) // Go up two levels
		internalPath := filepath.Join(projectRoot, "internal", "config", "cities.yaml")

		log.Printf("Trying internal config path: %s", internalPath)
		if _, err = os.Stat(internalPath); err == nil {
			yamlPath = internalPath
			log.Printf("Found cities.yaml at: %s", yamlPath)
		} else {
			return fmt.Errorf("cities.yaml not found in any location")
		}
	}

	// Read and parse the YAML file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("cannot read cities.yaml: %w", err)
	}

	// Debug: Print the size of loaded data
	log.Printf("Read %d bytes from cities.yaml", len(data))

	// Parse YAML
	if err := yaml.Unmarshal(data, &AppConfig.LocationData.POSTCODES); err != nil {
		return fmt.Errorf("cannot unmarshal cities.yaml: %w", err)
	}

	// Process the data to create derived structures
	for prov, hkot := range AppConfig.LocationData.POSTCODES {
		// Initialize province in DISTRICTS if not present
		if _, exists := AppConfig.LocationData.DISTRICTS[prov]; !exists {
			AppConfig.LocationData.DISTRICTS[prov] = make(map[string][]string)
		}

		// Initialize CITIES for this province
		AppConfig.LocationData.CITIES[prov] = make([]string, 0, len(hkot))

		for kota, hkec := range hkot {
			// Add city to CITIES
			AppConfig.LocationData.CITIES[prov] = append(AppConfig.LocationData.CITIES[prov], kota)

			// Add province to PROVINCES
			AppConfig.LocationData.PROVINCES[kota] = prov

			// Get districts (areas) for this city
			districts := make([]string, 0, len(hkec))
			for kec := range hkec {
				districts = append(districts, kec)

				// Update DISTRICT_NAMES
				if _, exists := AppConfig.LocationData.DISTRICT_NAMES[kec]; !exists {
					AppConfig.LocationData.DISTRICT_NAMES[kec] = make(map[string]string)
				}
				AppConfig.LocationData.DISTRICT_NAMES[kec][kota] = prov

				// Process postcodes
				for _, postcode := range hkec[kec] {
					if _, exists := AppConfig.LocationData.POSTCODE_MAPPING[postcode]; !exists {
						AppConfig.LocationData.POSTCODE_MAPPING[postcode] = make([]map[string]string, 0)
					}
					AppConfig.LocationData.POSTCODE_MAPPING[postcode] = append(AppConfig.LocationData.POSTCODE_MAPPING[postcode], map[string]string{
						"area":     kec,
						"city":     kota,
						"province": prov,
						"country":  "Indonesia",
					})
				}
			}

			// Add districts to DISTRICTS
			AppConfig.LocationData.DISTRICTS[prov][kota] = districts
		}
	}

	log.Printf("Successfully loaded location data: %d provinces, %d cities",
		len(AppConfig.LocationData.DISTRICTS), len(AppConfig.LocationData.PROVINCES))
	return nil
}

// loadJNEMapping loads the JNE mapping from YAML file
func loadJNEMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string]map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.JNEMapping = mapping
	return nil
}

// loadJNTMapping loads the JNT mapping from YAML file
func loadJNTMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string]map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.JNTMapping = mapping
	return nil
}

// loadSiCepatMapping loads the SiCepat mapping from YAML file
func loadSiCepatMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string]map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.SiCepatMappingOrigin = mapping
	return nil
}

// loadSiCepatDestinationMapping loads the SiCepat destination mapping from YAML file
func loadSiCepatDestinationMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string]map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.SiCepatMappingDestination = mapping
	return nil
}

// loadJNEJawaRegionMapping loads the JNE Jawa region mapping from YAML file
func loadJNEJawaRegionMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string][]string
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.JNEJawaRegionMapping = mapping
	return nil
}

// loadJNEWhitelistOriginMapping loads the JNE whitelist origin mapping from YAML file
func loadJNEWhitelistOriginMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string][]string
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.JNEWhitelistOriginMapping = mapping
	return nil
}

// loadSAPXMapping loads the SAPX mapping from YAML file
func loadSAPXMapping(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal YAML data
	var mapping map[string]map[string]map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return err
	}

	cfg.SAPXMapping = mapping
	return nil
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvWithDefault(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}

func getEnvAsSlice(key, separator string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return strings.Split(value, separator)
	}
	return []string{}
}

func getSameSiteMode(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// ValidateDurianPayConfig validates DurianPay configuration
func (c *Config) ValidateDurianPayConfig() error {
	if c.Payment.DurianPay.APIKey == "" {
		return fmt.Errorf("DURIANPAY_API_KEY is required")
	}
	if c.Payment.DurianPay.BaseURL == "" {
		return fmt.Errorf("DURIANPAY_BASE_URL is required")
	}
	if c.Payment.DurianPay.Environment != "sandbox" && c.Payment.DurianPay.Environment != "production" {
		return fmt.Errorf("DURIANPAY_ENVIRONMENT must be either 'sandbox' or 'production'")
	}
	return nil
}
