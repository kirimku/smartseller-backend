package entity

import (
	"database/sql/driver"
	"fmt"
	"mime"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AttachmentType represents the type of claim attachment
type AttachmentType string

const (
	AttachmentTypeReceipt  AttachmentType = "receipt"
	AttachmentTypePhoto    AttachmentType = "photo"
	AttachmentTypeInvoice  AttachmentType = "invoice"
	AttachmentTypeDocument AttachmentType = "document"
	AttachmentTypeVideo    AttachmentType = "video"
	AttachmentTypeOther    AttachmentType = "other"
)

// Valid validates the attachment type
func (at AttachmentType) Valid() bool {
	switch at {
	case AttachmentTypeReceipt, AttachmentTypePhoto, AttachmentTypeInvoice,
		AttachmentTypeDocument, AttachmentTypeVideo, AttachmentTypeOther:
		return true
	default:
		return false
	}
}

// String returns the string representation of AttachmentType
func (at AttachmentType) String() string {
	return string(at)
}

// Value implements the driver.Valuer interface for database storage
func (at AttachmentType) Value() (driver.Value, error) {
	return string(at), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (at *AttachmentType) Scan(value interface{}) error {
	if value == nil {
		*at = AttachmentTypeOther
		return nil
	}
	if str, ok := value.(string); ok {
		*at = AttachmentType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into AttachmentType", value)
}

// VirusScanStatus represents the virus scan status of an attachment
type VirusScanStatus string

const (
	VirusScanStatusPending  VirusScanStatus = "pending"
	VirusScanStatusClean    VirusScanStatus = "clean"
	VirusScanStatusInfected VirusScanStatus = "infected"
	VirusScanStatusFailed   VirusScanStatus = "failed"
)

// Valid validates the virus scan status
func (vss VirusScanStatus) Valid() bool {
	switch vss {
	case VirusScanStatusPending, VirusScanStatusClean, VirusScanStatusInfected, VirusScanStatusFailed:
		return true
	default:
		return false
	}
}

// ClaimAttachment represents a file attachment for a warranty claim
type ClaimAttachment struct {
	// Primary identification
	ID      uuid.UUID `json:"id" db:"id"`
	ClaimID uuid.UUID `json:"claim_id" db:"claim_id"`

	// Uploader information
	UploadedBy uuid.UUID `json:"uploaded_by" db:"uploaded_by"`

	// File information
	Filename         string `json:"filename" db:"filename"`
	OriginalFilename string `json:"original_filename" db:"original_filename"`
	FilePath         string `json:"file_path" db:"file_path"`
	FileURL          string `json:"file_url" db:"file_url"`
	FileSize         int64  `json:"file_size" db:"file_size"`
	MimeType         string `json:"mime_type" db:"mime_type"`

	// Categorization
	AttachmentType AttachmentType `json:"attachment_type" db:"attachment_type"`
	Description    *string        `json:"description,omitempty" db:"description"`

	// Processing status
	IsProcessed     bool    `json:"is_processed" db:"is_processed"`
	ProcessingNotes *string `json:"processing_notes,omitempty" db:"processing_notes"`

	// Security and validation
	Checksum        *string         `json:"checksum,omitempty" db:"checksum"`
	VirusScanStatus VirusScanStatus `json:"virus_scan_status" db:"virus_scan_status"`
	VirusScanDate   *time.Time      `json:"virus_scan_date,omitempty" db:"virus_scan_date"`

	// Upload timestamp
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`

	// Computed fields (not stored in database)
	FileExtension string `json:"file_extension" db:"-"`
	FileSizeHuman string `json:"file_size_human" db:"-"`
	IsImage       bool   `json:"is_image" db:"-"`
	IsDocument    bool   `json:"is_document" db:"-"`
	IsVideo       bool   `json:"is_video" db:"-"`
	IsSafe        bool   `json:"is_safe" db:"-"`
	ThumbnailURL  string `json:"thumbnail_url,omitempty" db:"-"`
}

// Allowed file extensions by category
var AllowedExtensions = map[AttachmentType][]string{
	AttachmentTypePhoto:    {".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"},
	AttachmentTypeDocument: {".pdf", ".doc", ".docx", ".txt", ".rtf"},
	AttachmentTypeVideo:    {".mp4", ".avi", ".mov", ".wmv", ".webm"},
	AttachmentTypeReceipt:  {".jpg", ".jpeg", ".png", ".pdf"},
	AttachmentTypeInvoice:  {".pdf", ".jpg", ".jpeg", ".png"},
	AttachmentTypeOther:    {".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx", ".txt"},
}

// Allowed MIME types
var AllowedMimeTypes = []string{
	"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/bmp",
	"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"text/plain", "application/rtf",
	"video/mp4", "video/avi", "video/quicktime", "video/x-ms-wmv", "video/webm",
}

// Maximum file sizes (in bytes)
const (
	MaxFileSizeImage    = 5 * 1024 * 1024  // 5MB for images
	MaxFileSizeDocument = 10 * 1024 * 1024 // 10MB for documents
	MaxFileSizeVideo    = 50 * 1024 * 1024 // 50MB for videos
)

// NewClaimAttachment creates a new claim attachment
func NewClaimAttachment(claimID, uploadedBy uuid.UUID, filename, originalFilename string, fileSize int64) *ClaimAttachment {
	return &ClaimAttachment{
		ID:               uuid.New(),
		ClaimID:          claimID,
		UploadedBy:       uploadedBy,
		Filename:         filename,
		OriginalFilename: originalFilename,
		FileSize:         fileSize,
		AttachmentType:   AttachmentTypeOther, // Default, should be set based on file
		IsProcessed:      false,
		VirusScanStatus:  VirusScanStatusPending,
		UploadedAt:       time.Now(),
	}
}

// Validate validates the claim attachment
func (ca *ClaimAttachment) Validate() error {
	// Required fields
	if ca.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if ca.UploadedBy == uuid.Nil {
		return fmt.Errorf("uploaded_by is required")
	}
	if ca.Filename == "" {
		return fmt.Errorf("filename is required")
	}
	if ca.OriginalFilename == "" {
		return fmt.Errorf("original_filename is required")
	}
	if ca.FileSize <= 0 {
		return fmt.Errorf("file_size must be positive")
	}

	// Validate attachment type
	if !ca.AttachmentType.Valid() {
		return fmt.Errorf("invalid attachment type: %s", ca.AttachmentType)
	}

	// Validate virus scan status
	if !ca.VirusScanStatus.Valid() {
		return fmt.Errorf("invalid virus scan status: %s", ca.VirusScanStatus)
	}

	// Validate file extension
	if err := ca.ValidateFileExtension(); err != nil {
		return fmt.Errorf("file extension validation failed: %w", err)
	}

	// Validate MIME type
	if err := ca.ValidateMimeType(); err != nil {
		return fmt.Errorf("mime type validation failed: %w", err)
	}

	// Validate file size
	if err := ca.ValidateFileSize(); err != nil {
		return fmt.Errorf("file size validation failed: %w", err)
	}

	return nil
}

// ValidateFileExtension validates the file extension
func (ca *ClaimAttachment) ValidateFileExtension() error {
	ext := strings.ToLower(filepath.Ext(ca.OriginalFilename))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	allowedExts, exists := AllowedExtensions[ca.AttachmentType]
	if !exists {
		return fmt.Errorf("no allowed extensions defined for attachment type: %s", ca.AttachmentType)
	}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("file extension %s not allowed for attachment type %s. Allowed: %v",
		ext, ca.AttachmentType, allowedExts)
}

// ValidateMimeType validates the MIME type
func (ca *ClaimAttachment) ValidateMimeType() error {
	if ca.MimeType == "" {
		return fmt.Errorf("mime_type is required")
	}

	for _, allowedType := range AllowedMimeTypes {
		if ca.MimeType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("mime type %s not allowed. Allowed: %v", ca.MimeType, AllowedMimeTypes)
}

// ValidateFileSize validates the file size based on attachment type
func (ca *ClaimAttachment) ValidateFileSize() error {
	var maxSize int64

	switch ca.AttachmentType {
	case AttachmentTypePhoto:
		maxSize = MaxFileSizeImage
	case AttachmentTypeDocument, AttachmentTypeReceipt, AttachmentTypeInvoice:
		maxSize = MaxFileSizeDocument
	case AttachmentTypeVideo:
		maxSize = MaxFileSizeVideo
	default:
		maxSize = MaxFileSizeDocument // Default to document size limit
	}

	if ca.FileSize > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes for attachment type %s",
			ca.FileSize, maxSize, ca.AttachmentType)
	}

	return nil
}

// DetectAttachmentType automatically detects attachment type from file extension and MIME type
func (ca *ClaimAttachment) DetectAttachmentType() {
	ext := strings.ToLower(filepath.Ext(ca.OriginalFilename))
	mimeType := strings.ToLower(ca.MimeType)

	// Check for images
	if strings.HasPrefix(mimeType, "image/") {
		ca.AttachmentType = AttachmentTypePhoto
		return
	}

	// Check for videos
	if strings.HasPrefix(mimeType, "video/") {
		ca.AttachmentType = AttachmentTypeVideo
		return
	}

	// Check for PDFs (commonly receipts/invoices)
	if mimeType == "application/pdf" {
		filename := strings.ToLower(ca.OriginalFilename)
		if strings.Contains(filename, "receipt") {
			ca.AttachmentType = AttachmentTypeReceipt
		} else if strings.Contains(filename, "invoice") {
			ca.AttachmentType = AttachmentTypeInvoice
		} else {
			ca.AttachmentType = AttachmentTypeDocument
		}
		return
	}

	// Check for documents
	documentExts := []string{".doc", ".docx", ".txt", ".rtf"}
	for _, docExt := range documentExts {
		if ext == docExt {
			ca.AttachmentType = AttachmentTypeDocument
			return
		}
	}

	// Default to other
	ca.AttachmentType = AttachmentTypeOther
}

// SetMimeTypeFromExtension sets MIME type based on file extension
func (ca *ClaimAttachment) SetMimeTypeFromExtension() {
	ca.MimeType = mime.TypeByExtension(filepath.Ext(ca.OriginalFilename))
	if ca.MimeType == "" {
		ca.MimeType = "application/octet-stream" // Default binary type
	}
}

// MarkAsProcessed marks the attachment as processed
func (ca *ClaimAttachment) MarkAsProcessed(notes string) {
	ca.IsProcessed = true
	if notes != "" {
		ca.ProcessingNotes = &notes
	}
}

// SetVirusScanResult sets the virus scan result
func (ca *ClaimAttachment) SetVirusScanResult(status VirusScanStatus) error {
	if !status.Valid() {
		return fmt.Errorf("invalid virus scan status: %s", status)
	}

	ca.VirusScanStatus = status
	now := time.Now()
	ca.VirusScanDate = &now

	return nil
}

// SetChecksum sets the file checksum
func (ca *ClaimAttachment) SetChecksum(checksum string) {
	ca.Checksum = &checksum
}

// ComputeFields calculates computed fields
func (ca *ClaimAttachment) ComputeFields() {
	// Set file extension
	ca.FileExtension = strings.ToLower(filepath.Ext(ca.OriginalFilename))

	// Set human-readable file size
	ca.FileSizeHuman = formatFileSize(ca.FileSize)

	// Determine file type flags
	ca.IsImage = strings.HasPrefix(ca.MimeType, "image/")
	ca.IsDocument = strings.HasPrefix(ca.MimeType, "application/") || strings.HasPrefix(ca.MimeType, "text/")
	ca.IsVideo = strings.HasPrefix(ca.MimeType, "video/")

	// Determine if file is safe
	ca.IsSafe = ca.VirusScanStatus == VirusScanStatusClean

	// Set thumbnail URL for images
	if ca.IsImage && ca.IsSafe {
		ca.ThumbnailURL = ca.FileURL + "?thumbnail=true"
	}
}

// formatFileSize formats file size into human-readable string
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}
}

// IsValidFilename checks if filename is valid (no dangerous characters)
func (ca *ClaimAttachment) IsValidFilename() bool {
	// Check for dangerous characters
	dangerousChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	if dangerousChars.MatchString(ca.OriginalFilename) {
		return false
	}

	// Check filename length
	if len(ca.OriginalFilename) > 255 {
		return false
	}

	// Check for reserved names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	baseFilename := strings.ToUpper(strings.TrimSuffix(ca.OriginalFilename, filepath.Ext(ca.OriginalFilename)))

	for _, reserved := range reservedNames {
		if baseFilename == reserved {
			return false
		}
	}

	return true
}

// GetDisplayName returns a display-friendly name
func (ca *ClaimAttachment) GetDisplayName() string {
	if ca.Description != nil && *ca.Description != "" {
		return *ca.Description
	}
	return ca.OriginalFilename
}

// CanDownload checks if the attachment can be downloaded
func (ca *ClaimAttachment) CanDownload() bool {
	return ca.IsSafe && ca.IsProcessed
}

// GetSecurityInfo returns security information about the file
func (ca *ClaimAttachment) GetSecurityInfo() map[string]interface{} {
	return map[string]interface{}{
		"virus_scan_status": ca.VirusScanStatus,
		"virus_scan_date":   ca.VirusScanDate,
		"is_safe":           ca.IsSafe,
		"checksum":          ca.Checksum,
		"is_processed":      ca.IsProcessed,
	}
}

// String returns a string representation of the claim attachment
func (ca *ClaimAttachment) String() string {
	return fmt.Sprintf("ClaimAttachment{ID: %s, Claim: %s, File: %s, Type: %s, Size: %s}",
		ca.ID.String(), ca.ClaimID.String(), ca.OriginalFilename, ca.AttachmentType, ca.FileSizeHuman)
}
