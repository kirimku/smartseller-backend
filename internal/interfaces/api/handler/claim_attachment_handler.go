package handler

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// ClaimAttachmentHandler handles claim attachment-related HTTP requests
type ClaimAttachmentHandler struct {
	logger *slog.Logger
}

// NewClaimAttachmentHandler creates a new claim attachment handler
func NewClaimAttachmentHandler() *ClaimAttachmentHandler {
	return &ClaimAttachmentHandler{
		logger: slog.Default(),
	}
}

// ListAttachments handles GET /api/v1/admin/warranty/claims/:claim_id/attachments
func (h *ClaimAttachmentHandler) ListAttachments(c *gin.Context) {
	claimID := c.Param("id")
	
	// Validate claim ID
	if claimID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID is required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	attachments := []*dto.ClaimAttachmentResponse{
		{
			ID:                 "att_001",
			ClaimID:            claimID,
			FileName:           "receipt.pdf",
			FilePath:           "/uploads/claims/2024/01/receipt.pdf",
			FileURL:            "https://cdn.example.com/claims/receipt.pdf",
			FileSize:           1024000,
			FileType:           "pdf",
			MimeType:           "application/pdf",
			AttachmentType:     "receipt",
			UploadedBy:         "user_123",
			SecurityScanStatus: "passed",
			CreatedAt:          time.Now().Add(-24 * time.Hour),
			UpdatedAt:          time.Now().Add(-12 * time.Hour),
		},
		{
			ID:                 "att_002",
			ClaimID:            claimID,
			FileName:           "damage_photo.jpg",
			FilePath:           "/uploads/claims/2024/01/damage_photo.jpg",
			FileURL:            "https://cdn.example.com/claims/damage_photo.jpg",
			FileSize:           2048000,
			FileType:           "jpg",
			MimeType:           "image/jpeg",
			AttachmentType:     "photo",
			UploadedBy:         "user_123",
			SecurityScanStatus: "pending",
			CreatedAt:          time.Now().Add(-12 * time.Hour),
			UpdatedAt:          time.Now().Add(-12 * time.Hour),
		},
	}

	response := dto.ClaimAttachmentListResponse{
		Attachments: attachments,
		Total:       len(attachments),
	}

	h.logger.Info("Listed claim attachments", "claim_id", claimID, "count", len(attachments))
	c.JSON(http.StatusOK, response)
}

// UploadAttachment handles POST /api/v1/admin/warranty/claims/:claim_id/attachments/upload
func (h *ClaimAttachmentHandler) UploadAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	
	// Validate claim ID
	if claimID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID is required",
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_form",
			"message": "Failed to parse multipart form",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "file_required",
			"message": "File is required",
		})
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateFile(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	// Security scan (mock implementation)
	if err := h.scanFileForViruses(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "security_scan_failed",
			"message": "File failed security scan",
		})
		return
	}

	// Mock file upload - replace with actual file storage service
	attachmentID := fmt.Sprintf("att_%d", time.Now().Unix())
	
	response := dto.ClaimAttachmentResponse{
		ID:                 attachmentID,
		ClaimID:            claimID,
		FileName:           header.Filename,
		FilePath:           fmt.Sprintf("/uploads/claims/%s/%s", claimID, header.Filename),
		FileURL:            fmt.Sprintf("https://cdn.example.com/claims/%s", header.Filename),
		FileSize:           header.Size,
		FileType:           strings.TrimPrefix(filepath.Ext(header.Filename), "."),
		MimeType:           header.Header.Get("Content-Type"),
		AttachmentType:     "document", // Default type
		UploadedBy:         "current_user", // Replace with actual user ID
		SecurityScanStatus: "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	h.logger.Info("Uploaded attachment", "claim_id", claimID, "file_name", header.Filename, "attachment_id", attachmentID)
	c.JSON(http.StatusCreated, response)
}

// DownloadAttachment handles GET /api/v1/admin/warranty/claims/:claim_id/attachments/:attachment_id/download
func (h *ClaimAttachmentHandler) DownloadAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate parameters
	if claimID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and attachment ID are required",
		})
		return
	}

	// Mock implementation - replace with actual file retrieval
	// In real implementation, you would:
	// 1. Verify attachment belongs to claim
	// 2. Check user permissions
	// 3. Retrieve file from storage
	// 4. Stream file to response

	h.logger.Info("Downloaded attachment", "claim_id", claimID, "attachment_id", attachmentID)
	
	// Mock file response
	c.Header("Content-Disposition", "attachment; filename=sample_file.pdf")
	c.Header("Content-Type", "application/pdf")
	c.String(http.StatusOK, "Mock file content")
}

// DeleteAttachment handles DELETE /api/v1/admin/warranty/claims/:claim_id/attachments/:attachment_id
func (h *ClaimAttachmentHandler) DeleteAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate parameters
	if claimID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and attachment ID are required",
		})
		return
	}

	// Mock implementation - replace with actual deletion logic
	// In real implementation, you would:
	// 1. Verify attachment belongs to claim
	// 2. Check user permissions
	// 3. Delete file from storage
	// 4. Remove database record

	h.logger.Info("Deleted attachment", "claim_id", claimID, "attachment_id", attachmentID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Attachment deleted successfully",
	})
}

// ApproveAttachment handles POST /api/v1/admin/warranty/claims/:claim_id/attachments/:attachment_id/approve
func (h *ClaimAttachmentHandler) ApproveAttachment(c *gin.Context) {
	claimID := c.Param("claim_id")
	attachmentID := c.Param("attachment_id")
	
	// Validate parameters
	if claimID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and attachment ID are required",
		})
		return
	}

	var req dto.AttachmentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
		return
	}

	// Mock implementation - replace with actual approval logic
	now := time.Now()
	approved := req.Action == "approve"
	
	response := dto.ClaimAttachmentResponse{
		ID:                 attachmentID,
		ClaimID:            claimID,
		FileName:           "sample_file.pdf",
		FilePath:           "/uploads/claims/sample/sample_file.pdf",
		FileURL:            "https://cdn.example.com/claims/sample_file.pdf",
		FileSize:           1024000,
		FileType:           "pdf",
		MimeType:           "application/pdf",
		AttachmentType:     "document",
		UploadedBy:         "user_123",
		SecurityScanStatus: "passed",
		CreatedAt:          now.Add(-24 * time.Hour),
		UpdatedAt:          now,
	}

	h.logger.Info("Processed attachment approval", "claim_id", claimID, "attachment_id", attachmentID, "action", req.Action, "approved", approved)
	c.JSON(http.StatusOK, response)
}

// validateFile validates uploaded file
func (h *ClaimAttachmentHandler) validateFile(header *multipart.FileHeader) error {
	// Check file size (10MB limit)
	maxSize := int64(10 << 20) // 10 MB
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds limit of %d bytes", maxSize)
	}

	// Check file extension
	allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".doc", ".docx", ".txt"}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return nil
		}
	}
	
	return fmt.Errorf("file type %s is not allowed", ext)
}

// scanFileForViruses performs virus scanning (mock implementation)
func (h *ClaimAttachmentHandler) scanFileForViruses(file multipart.File) error {
	// Mock implementation - always pass
	// In real implementation, integrate with antivirus service
	
	// Read first few bytes to simulate scanning
	buffer := make([]byte, 1024)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file for scanning")
	}
	
	// Reset file pointer
	file.Seek(0, 0)
	
	return nil
}