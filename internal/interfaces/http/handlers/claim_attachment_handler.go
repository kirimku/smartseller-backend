package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/logger"
)

// ClaimAttachmentHandler handles claim attachment related HTTP requests
type ClaimAttachmentHandler struct {
	logger zerolog.Logger
}

// NewClaimAttachmentHandler creates a new claim attachment handler
func NewClaimAttachmentHandler() *ClaimAttachmentHandler {
	return &ClaimAttachmentHandler{
		logger: logger.Logger,
	}
}

// ListClaimAttachments handles GET /api/v1/admin/warranty/claims/{claim_id}/attachments
func (h *ClaimAttachmentHandler) ListClaimAttachments(c *gin.Context) {
	claimID := c.Param("claim_id")
	
	// Validate claim ID format
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	h.logger.Info().Str("claim_id", claimID).Msg("Listing claim attachments")

	// Mock response - replace with actual service call
	attachments := []*dto.ClaimAttachmentResponse{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440000",
			ClaimID:            claimID,
			FileName:           "receipt.pdf",
			FilePath:           "/uploads/claims/2024/01/receipt.pdf",
			FileURL:            "https://cdn.example.com/claims/receipt.pdf",
			FileSize:           1048576,
			FileType:           "pdf",
			MimeType:           "application/pdf",
			AttachmentType:     "receipt",
			Description:        stringPtr("Purchase receipt"),
			UploadedBy:         "550e8400-e29b-41d4-a716-446655440002",
			SecurityScanStatus: "passed",
			SecurityScanResult: stringPtr("No threats detected"),
			CreatedAt:          time.Now().Add(-24 * time.Hour),
			UpdatedAt:          time.Now().Add(-24 * time.Hour),
		},
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			ClaimID:            claimID,
			FileName:           "device_photo.jpg",
			FilePath:           "/uploads/claims/2024/01/device_photo.jpg",
			FileURL:            "https://cdn.example.com/claims/device_photo.jpg",
			FileSize:           2097152,
			FileType:           "jpg",
			MimeType:           "image/jpeg",
			AttachmentType:     "photo",
			Description:        stringPtr("Photo of damaged device"),
			UploadedBy:         "550e8400-e29b-41d4-a716-446655440002",
			SecurityScanStatus: "passed",
			SecurityScanResult: stringPtr("No threats detected"),
			CreatedAt:          time.Now().Add(-12 * time.Hour),
			UpdatedAt:          time.Now().Add(-12 * time.Hour),
		},
	}

	response := &dto.ClaimAttachmentListResponse{
		Attachments: attachments,
		Total:       len(attachments),
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Int("attachment_count", len(attachments)).
		Msg("Successfully listed claim attachments")

	c.JSON(http.StatusOK, response)
}

// UploadClaimAttachment handles POST /api/v1/admin/warranty/claims/{claim_id}/attachments
func (h *ClaimAttachmentHandler) UploadClaimAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	
	// Validate claim ID format
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": "Failed to parse multipart form",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "File required",
			"message": "Please provide a file to upload",
		})
		return
	}
	defer file.Close()

	// Get attachment metadata
	attachmentType := c.Request.FormValue("attachment_type")
	description := c.Request.FormValue("description")

	// Validate attachment type
	validTypes := []string{"receipt", "photo", "video", "document", "other"}
	if !contains(validTypes, attachmentType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid attachment type",
			"message": "Attachment type must be one of: receipt, photo, video, document, other",
		})
		return
	}

	// Validate file
	if err := h.validateFile(header); err != nil {
		h.logger.Error().Err(err).Msg("File validation failed")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "File validation failed",
			"message": err.Error(),
		})
		return
	}

	// Perform security scan
	scanResult, err := h.performSecurityScan(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("Security scan failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Security scan failed",
			"message": "Failed to scan file for security threats",
		})
		return
	}

	if scanResult.Status != "passed" {
		h.logger.Warn().Str("scan_result", scanResult.Result).Msg("File failed security scan")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Security scan failed",
			"message": "File contains potential security threats",
		})
		return
	}

	// Generate unique filename and save file
	fileID := uuid.New().String()
	fileExt := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s%s", fileID, fileExt)
	filePath := filepath.Join("/uploads/claims", time.Now().Format("2006/01"), fileName)

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create upload directory")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Upload failed",
			"message": "Failed to create upload directory",
		})
		return
	}

	// Save file (mock implementation)
	h.logger.Info().
		Str("claim_id", claimID).
		Str("file_name", header.Filename).
		Int64("file_size", header.Size).
		Str("attachment_type", attachmentType).
		Str("file_path", filePath).
		Msg("File upload simulated successfully")

	// Mock response - replace with actual service call
	attachment := &dto.ClaimAttachmentResponse{
		ID:                 uuid.New().String(),
		ClaimID:            claimID,
		FileName:           header.Filename,
		FilePath:           filePath,
		FileURL:            fmt.Sprintf("https://cdn.example.com/claims/%s", fileName),
		FileSize:           header.Size,
		FileType:           strings.TrimPrefix(fileExt, "."),
		MimeType:           header.Header.Get("Content-Type"),
		AttachmentType:     attachmentType,
		Description:        stringPtrIfNotEmpty(description),
		UploadedBy:         "550e8400-e29b-41d4-a716-446655440002", // Mock user ID
		SecurityScanStatus: scanResult.Status,
		SecurityScanResult: &scanResult.Result,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachment.ID).
		Str("file_name", attachment.FileName).
		Msg("Successfully uploaded claim attachment")

	c.JSON(http.StatusCreated, attachment)
}

// DownloadClaimAttachment handles GET /api/v1/admin/warranty/claims/{claim_id}/attachments/{attachment_id}/download
func (h *ClaimAttachmentHandler) DownloadClaimAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(attachmentID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid attachment ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid attachment ID format",
			"message": "Attachment ID must be a valid UUID",
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Msg("Downloading claim attachment")

	// Mock file download - replace with actual service call
	fileName := "receipt.pdf"
	filePath := "/uploads/claims/2024/01/receipt.pdf"
	
	// In a real implementation, you would:
	// 1. Verify the attachment belongs to the claim
	// 2. Check user permissions
	// 3. Stream the file from storage
	
	// For now, return a mock response
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", "1048576")
	
	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Str("file_name", fileName).
		Str("file_path", filePath).
		Msg("Successfully initiated attachment download")

	// Mock file content
	c.String(http.StatusOK, "Mock file content for attachment download")
}

// DeleteClaimAttachment handles DELETE /api/v1/admin/warranty/claims/{claim_id}/attachments/{attachment_id}
func (h *ClaimAttachmentHandler) DeleteClaimAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(attachmentID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid attachment ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid attachment ID format",
			"message": "Attachment ID must be a valid UUID",
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Msg("Deleting claim attachment")

	// Mock deletion - replace with actual service call
	// In a real implementation, you would:
	// 1. Verify the attachment exists and belongs to the claim
	// 2. Check user permissions
	// 3. Delete the file from storage
	// 4. Remove the database record

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Msg("Successfully deleted claim attachment")

	c.JSON(http.StatusOK, gin.H{
		"message": "Attachment deleted successfully",
	})
}

// ApproveClaimAttachment handles POST /api/v1/admin/warranty/claims/{claim_id}/attachments/{attachment_id}/approve
func (h *ClaimAttachmentHandler) ApproveClaimAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(attachmentID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid attachment ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid attachment ID format",
			"message": "Attachment ID must be a valid UUID",
		})
		return
	}

	// Parse request body
	var req dto.AttachmentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse approval request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Str("action", req.Action).
		Msg("Processing attachment approval")

	// Mock approval process - replace with actual service call
	// In a real implementation, you would:
	// 1. Verify the attachment exists and belongs to the claim
	// 2. Check user permissions
	// 3. Update the attachment approval status
	// 4. Create timeline entry

	h.logger.Info().
		Str("claim_id", claimID).
		Str("attachment_id", attachmentID).
		Str("action", req.Action).
		Msg("Successfully processed attachment approval")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Attachment %s successfully", req.Action+"d"),
		"action":  req.Action,
	})
}

// Helper functions

func (h *ClaimAttachmentHandler) validateFile(header *multipart.FileHeader) error {
	// Check file size (10MB max)
	maxSize := int64(10 << 20) // 10 MB
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Check file extension
	allowedExts := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".doc", ".docx", ".txt", ".mp4", ".avi", ".mov"}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !contains(allowedExts, ext) {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	return nil
}

type SecurityScanResult struct {
	Status string
	Result string
}

func (h *ClaimAttachmentHandler) performSecurityScan(file multipart.File) (*SecurityScanResult, error) {
	// Mock security scan - replace with actual virus scanning service
	// In a real implementation, you would:
	// 1. Use a virus scanning service like ClamAV
	// 2. Check file signatures
	// 3. Scan for malicious content
	
	// Reset file pointer
	file.Seek(0, io.SeekStart)
	
	// Read first few bytes to check file signature
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file for scanning: %w", err)
	}
	
	// Reset file pointer again
	file.Seek(0, io.SeekStart)
	
	// Mock scan result
	return &SecurityScanResult{
		Status: "passed",
		Result: "No threats detected",
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func stringPtr(s string) *string {
	return &s
}

func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}