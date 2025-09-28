package dto

import (
	"time"
)

// MobileWarrantyScanRequest represents a mobile warranty scan request
type MobileWarrantyScanRequest struct {
	ScanType    string `json:"scan_type" binding:"required" example:"qr_code"` // qr_code, barcode, manual
	ScannedData string `json:"scanned_data" binding:"required" example:"WB-2024-001234567"`
	DeviceInfo  MobileDeviceInfo `json:"device_info,omitempty"`
}

// MobileWarrantyScanResponse represents the response from warranty scanning
type MobileWarrantyScanResponse struct {
	Success     bool                     `json:"success" example:"true"`
	ScanType    string                   `json:"scan_type" example:"qr_code"`
	ScannedData string                   `json:"scanned_data" example:"WB-2024-001234567"`
	Warranty    *MobileWarrantyInfo      `json:"warranty,omitempty"`
	Actions     []MobileWarrantyAction   `json:"actions"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// MobileWarrantyInfo represents warranty information optimized for mobile
type MobileWarrantyInfo struct {
	ID              string    `json:"id" example:"warranty-12345"`
	ProductName     string    `json:"product_name" example:"SmartPhone Pro Max"`
	ProductModel    string    `json:"product_model" example:"SPM-2024"`
	SerialNumber    string    `json:"serial_number" example:"SN123456789"`
	PurchaseDate    time.Time `json:"purchase_date" example:"2023-01-15T00:00:00Z"`
	ExpiryDate      time.Time `json:"expiry_date" example:"2025-01-15T00:00:00Z"`
	Status          string    `json:"status" example:"active"`
	CoverageType    string    `json:"coverage_type" example:"comprehensive"`
	IsValid         bool      `json:"is_valid" example:"true"`
	DaysRemaining   int       `json:"days_remaining" example:"365"`
	Retailer        string    `json:"retailer" example:"TechStore Plus"`
	WarrantyTerms   string    `json:"warranty_terms" example:"2-year comprehensive warranty"`
}

// MobileWarrantyAction represents available actions for a warranty
type MobileWarrantyAction struct {
	Type        string `json:"type" example:"view_details"`
	Label       string `json:"label" example:"View Full Details"`
	Description string `json:"description" example:"See complete warranty information"`
	Enabled     bool   `json:"enabled" example:"true"`
}

// MobileDeviceInfo represents mobile device information
type MobileDeviceInfo struct {
	DeviceID     string `json:"device_id" example:"device-12345"`
	Platform     string `json:"platform" example:"ios"` // ios, android
	AppVersion   string `json:"app_version" example:"1.2.3"`
	OSVersion    string `json:"os_version" example:"17.0"`
	DeviceModel  string `json:"device_model" example:"iPhone 15 Pro"`
	Timezone     string `json:"timezone" example:"America/New_York"`
	Language     string `json:"language" example:"en"`
}

// MobileCameraPermissionsResponse represents camera permission status
type MobileCameraPermissionsResponse struct {
	HasPermission    bool                              `json:"has_permission" example:"true"`
	PermissionLevel  string                            `json:"permission_level" example:"full"`
	DeviceType       string                            `json:"device_type" example:"ios"`
	AppVersion       string                            `json:"app_version" example:"1.2.3"`
	RequiredFeatures []string                          `json:"required_features"`
	Recommendations  []MobilePermissionRecommendation  `json:"recommendations"`
	Metadata         map[string]interface{}            `json:"metadata,omitempty"`
}

// MobilePermissionRecommendation represents permission recommendations
type MobilePermissionRecommendation struct {
	Feature     string `json:"feature" example:"camera"`
	Status      string `json:"status" example:"granted"`
	Description string `json:"description" example:"Camera access is required for QR code scanning"`
	Action      string `json:"action" example:"none"`
}

// MobilePhotoUploadResponse represents photo upload response
type MobilePhotoUploadResponse struct {
	Success     bool                   `json:"success" example:"true"`
	PhotoID     string                 `json:"photo_id" example:"photo-12345"`
	ClaimID     string                 `json:"claim_id" example:"claim-001"`
	FileName    string                 `json:"file_name" example:"damage_photo.jpg"`
	FileSize    int                    `json:"file_size" example:"1024000"`
	UploadedAt  time.Time              `json:"uploaded_at" example:"2024-01-15T10:30:00Z"`
	Description string                 `json:"description" example:"Photo of damaged screen"`
	Category    string                 `json:"category" example:"damage_evidence"`
	Processing  MobilePhotoProcessing  `json:"processing"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MobilePhotoProcessing represents photo processing status
type MobilePhotoProcessing struct {
	Status           string  `json:"status" example:"queued"`
	EstimatedTime    string  `json:"estimated_time" example:"2-5 minutes"`
	CompressionRatio float64 `json:"compression_ratio" example:"0.7"`
	ThumbnailReady   bool    `json:"thumbnail_ready" example:"false"`
}

// MobileOfflineSyncResponse represents offline sync data
type MobileOfflineSyncResponse struct {
	SyncTimestamp       time.Time              `json:"sync_timestamp" example:"2024-01-15T10:30:00Z"`
	LastSync            time.Time              `json:"last_sync" example:"2024-01-15T06:30:00Z"`
	DeviceID            string                 `json:"device_id" example:"device-12345"`
	HasUpdates          bool                   `json:"has_updates" example:"true"`
	DataVersion         string                 `json:"data_version" example:"1.0.0"`
	SyncData            MobileSyncData         `json:"sync_data"`
	NextSyncRecommended time.Time              `json:"next_sync_recommended" example:"2024-01-15T14:30:00Z"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// MobileSyncData represents data for offline synchronization
type MobileSyncData struct {
	Warranties []MobileWarrantySync `json:"warranties"`
	Claims     []MobileClaimSync    `json:"claims"`
	Settings   MobileSettingsSync   `json:"settings"`
}

// MobileWarrantySync represents warranty sync data
type MobileWarrantySync struct {
	ID           string                 `json:"id" example:"warranty-12345"`
	Action       string                 `json:"action" example:"update"` // create, update, delete
	LastModified time.Time              `json:"last_modified" example:"2024-01-15T09:30:00Z"`
	Data         map[string]interface{} `json:"data"`
}

// MobileClaimSync represents claim sync data
type MobileClaimSync struct {
	ID           string                 `json:"id" example:"claim-001"`
	Action       string                 `json:"action" example:"update"` // create, update, delete
	LastModified time.Time              `json:"last_modified" example:"2024-01-15T08:30:00Z"`
	Data         map[string]interface{} `json:"data"`
}

// MobileSettingsSync represents settings sync data
type MobileSettingsSync struct {
	NotificationPreferences map[string]bool        `json:"notification_preferences"`
	AppSettings             map[string]interface{} `json:"app_settings"`
}

// MobilePushRegistrationRequest represents push notification registration
type MobilePushRegistrationRequest struct {
	DeviceID    string           `json:"device_id" binding:"required" example:"device-12345"`
	PushToken   string           `json:"push_token" binding:"required" example:"fcm_token_12345"`
	Platform    string           `json:"platform" binding:"required" example:"ios"`
	AppVersion  string           `json:"app_version" binding:"required" example:"1.2.3"`
	DeviceInfo  MobileDeviceInfo `json:"device_info,omitempty"`
	Preferences MobilePushPreferences `json:"preferences,omitempty"`
}

// MobilePushRegistrationResponse represents push registration response
type MobilePushRegistrationResponse struct {
	Success          bool                   `json:"success" example:"true"`
	DeviceID         string                 `json:"device_id" example:"device-12345"`
	PushToken        string                 `json:"push_token" example:"fcm_token_12345"`
	RegisteredAt     time.Time              `json:"registered_at" example:"2024-01-15T10:30:00Z"`
	Platform         string                 `json:"platform" example:"ios"`
	AppVersion       string                 `json:"app_version" example:"1.2.3"`
	NotificationID   string                 `json:"notification_id" example:"notif-12345"`
	Preferences      MobilePushPreferences  `json:"preferences"`
	TestNotification MobileTestNotification `json:"test_notification"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// MobilePushPreferences represents push notification preferences
type MobilePushPreferences struct {
	ClaimUpdates      bool              `json:"claim_updates" example:"true"`
	WarrantyReminders bool              `json:"warranty_reminders" example:"true"`
	PromotionalOffers bool              `json:"promotional_offers" example:"false"`
	SystemAlerts      bool              `json:"system_alerts" example:"true"`
	QuietHours        MobileQuietHours  `json:"quiet_hours"`
}

// MobileQuietHours represents quiet hours for notifications
type MobileQuietHours struct {
	Enabled   bool   `json:"enabled" example:"true"`
	StartTime string `json:"start_time" example:"22:00"`
	EndTime   string `json:"end_time" example:"08:00"`
	Timezone  string `json:"timezone" example:"UTC"`
}

// MobileTestNotification represents test notification info
type MobileTestNotification struct {
	Sent      bool      `json:"sent" example:"true"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Message   string    `json:"message" example:"Welcome! Push notifications are now enabled."`
}

// Mobile-specific error responses
type MobileErrorResponse struct {
	Error     string                 `json:"error" example:"invalid_scan_data"`
	Message   string                 `json:"message" example:"The scanned data is not valid"`
	Code      string                 `json:"code" example:"MOB_400"`
	Details   string                 `json:"details,omitempty" example:"QR code format not recognized"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	RequestID string                 `json:"request_id,omitempty" example:"req_123456789"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Mobile    struct {
		RetryAfter int  `json:"retry_after,omitempty"` // seconds
		Offline    bool `json:"offline_capable"`
	} `json:"mobile"`
}

// MobileOptimizedResponse represents a mobile-optimized API response
type MobileOptimizedResponse struct {
	Success bool                `json:"success"`
	Data    interface{}         `json:"data"`
	Meta    MobileResponseMeta  `json:"meta"`
}

// MobileResponseMeta contains metadata for mobile-optimized responses
type MobileResponseMeta struct {
	MobileOptimized bool      `json:"mobile_optimized"`
	Timestamp       time.Time `json:"timestamp"`
	APIVersion      string    `json:"api_version"`
	Compression     string    `json:"compression,omitempty"`
	CacheControl    string    `json:"cache_control,omitempty"`
}

// Mobile optimization metadata
type MobileOptimizationInfo struct {
	Compressed     bool   `json:"compressed" example:"true"`
	CacheDuration  string `json:"cache_duration" example:"5m"`
	OfflineReady   bool   `json:"offline_ready" example:"true"`
	DataSize       string `json:"data_size" example:"2.3KB"`
	ResponseTime   string `json:"response_time" example:"150ms"`
	ServerVersion  string `json:"server_version" example:"1.0.0"`
}