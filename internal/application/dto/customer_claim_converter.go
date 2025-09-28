package dto

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
)

// ToCustomerClaimSubmissionResponse converts claim entity to submission response
func ToCustomerClaimSubmissionResponse(claimID, warrantyID string) CustomerClaimSubmissionResponse {
	now := time.Now()
	estimatedResolution := now.AddDate(0, 0, 7) // 7 days from now
	
	return CustomerClaimSubmissionResponse{
		ClaimID:             claimID,
		ClaimNumber:         fmt.Sprintf("CLM-%s", claimID[:8]),
		Status:              "submitted",
		SubmittedAt:         now,
		EstimatedResolution: estimatedResolution,
		Priority:            "medium",
		AssignedAgent:       "Support Team",
		NextSteps: []string{
			"Your claim has been submitted successfully",
			"Our technical team will review your claim within 24 hours",
			"You will receive an email confirmation shortly",
			"Track your claim status using the provided tracking number",
		},
		ContactInfo: CustomerClaimContactInfo{
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			Phone:            "+1234567890",
			Address:          "123 Main Street",
			City:             "New York",
			PostalCode:       "10001",
			PreferredContact: "email",
		},
		TrackingInfo: CustomerClaimTrackingInfo{
			TrackingNumber: fmt.Sprintf("TRK-%s", claimID[:10]),
			StatusURL:      fmt.Sprintf("https://support.example.com/claims/%s", claimID),
			SupportEmail:   "claims@example.com",
			SupportPhone:   "+1-800-SUPPORT",
		},
	}
}

// ToCustomerClaimListResponse converts claim entities to list response
func ToCustomerClaimListResponse(page, limit, totalCount int) CustomerClaimListResponse {
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrevious := page > 1
	
	// Mock claims data
	claims := []CustomerClaimInfo{
		{
			ClaimID:      "claim-001",
			ClaimNumber:  "CLM-001",
			Status:       "in_progress",
			IssueType:    "defect",
			Severity:     "medium",
			SubmittedAt:  time.Now().AddDate(0, 0, -5),
			UpdatedAt:    time.Now().AddDate(0, 0, -1),
			ProductName:  "Smartphone Pro Max",
			ProductSKU:   "SKU-PHONE-001",
			WarrantyID:   "warranty-001",
			Priority:     "medium",
			DaysOpen:     5,
			LastActivity: "Technical review completed",
		},
		{
			ClaimID:      "claim-002",
			ClaimNumber:  "CLM-002",
			Status:       "resolved",
			IssueType:    "malfunction",
			Severity:     "high",
			SubmittedAt:  time.Now().AddDate(0, 0, -15),
			UpdatedAt:    time.Now().AddDate(0, 0, -3),
			ProductName:  "Wireless Headphones",
			ProductSKU:   "SKU-AUDIO-002",
			WarrantyID:   "warranty-002",
			Priority:     "high",
			DaysOpen:     12,
			LastActivity: "Replacement shipped",
		},
	}
	
	return CustomerClaimListResponse{
		Claims:      claims,
		TotalCount:  totalCount,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
		Summary: CustomerClaimsSummary{
			TotalClaims:    totalCount,
			OpenClaims:     8,
			ResolvedClaims: 15,
			PendingClaims:  2,
		},
		RequestTime: time.Now(),
	}
}

// ToCustomerClaimDetailResponse converts claim entity to detailed response
func ToCustomerClaimDetailResponse(claimID string) CustomerClaimDetailResponse {
	now := time.Now()
	submittedAt := now.AddDate(0, 0, -5)
	estimatedResolution := submittedAt.AddDate(0, 0, 7)
	
	cost := decimal.NewFromFloat(150.00)
	
	return CustomerClaimDetailResponse{
		ClaimID:             claimID,
		ClaimNumber:         fmt.Sprintf("CLM-%s", claimID[:8]),
		Status:              "in_progress",
		IssueType:           "defect",
		Description:         "Device screen flickering intermittently, especially in low light conditions",
		Severity:            "medium",
		Priority:            "medium",
		SubmittedAt:         submittedAt,
		UpdatedAt:           now.AddDate(0, 0, -1),
		EstimatedResolution: estimatedResolution,
		ContactInfo: CustomerClaimContactInfo{
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			Phone:            "+1234567890",
			Address:          "123 Main Street",
			City:             "New York",
			PostalCode:       "10001",
			PreferredContact: "email",
		},
		ProductInfo: CustomerClaimProductInfo{
			SerialNumber:     "SN123456789",
			PurchaseDate:     now.AddDate(-1, 0, 0),
			PurchaseLocation: "TechStore Inc",
			UsageFrequency:   "daily",
			Environment:      "indoor",
		},
		WarrantyInfo: CustomerClaimWarrantyInfo{
			WarrantyID:     "warranty-001",
			ProductName:    "Smartphone Pro Max 256GB",
			ProductSKU:     "SKU-PHONE-001",
			SerialNumber:   "SN123456789",
			ActivatedAt:    now.AddDate(-1, 0, 0),
			ExpiresAt:      now.AddDate(1, 0, 0),
			WarrantyPeriod: 24,
			CoverageType:   "comprehensive",
			IsActive:       true,
		},
		Attachments: []CustomerClaimAttachment{
			{
				Type:        "photo",
				FileName:    "screen_issue.jpg",
				FileSize:    1024000,
				ContentType: "image/jpeg",
				URL:         "https://example.com/attachments/screen_issue.jpg",
				Description: "Photo showing screen flickering issue",
			},
			{
				Type:        "video",
				FileName:    "issue_demonstration.mp4",
				FileSize:    5120000,
				ContentType: "video/mp4",
				URL:         "https://example.com/attachments/issue_demo.mp4",
				Description: "Video demonstrating the flickering issue",
			},
		},
		Timeline: []CustomerClaimTimelineItem{
			{
				ID:          "timeline-001",
				Event:       "claim_submitted",
				Description: "Claim submitted by customer",
				Timestamp:   submittedAt,
				Actor:       "John Doe",
				ActorType:   "customer",
			},
			{
				ID:          "timeline-002",
				Event:       "claim_acknowledged",
				Description: "Claim acknowledged and assigned to technical team",
				Timestamp:   submittedAt.Add(time.Hour * 2),
				Actor:       "Support System",
				ActorType:   "system",
			},
			{
				ID:          "timeline-003",
				Event:       "technical_review",
				Description: "Technical review completed - issue confirmed",
				Timestamp:   now.AddDate(0, 0, -1),
				Actor:       "Tech Support",
				ActorType:   "agent",
			},
		},
		Resolution: &CustomerClaimResolution{
			Type:        "repair",
			Description: "Screen replacement required",
			ResolvedAt:  time.Time{}, // Not resolved yet
			ResolvedBy:  "",
			Cost:        &cost,
		},
		AssignedAgent: &CustomerClaimAgent{
			ID:     "agent-001",
			Name:   "Sarah Johnson",
			Email:  "sarah.johnson@example.com",
			Phone:  "+1-800-SUPPORT",
			Role:   "Senior Technical Support",
			Avatar: "https://example.com/avatars/sarah.jpg",
		},
		Communication: []CustomerClaimCommunication{
			{
				ID:        "comm-001",
				Type:      "email",
				Direction: "outbound",
				Subject:   "Claim Confirmation - CLM-001",
				Message:   "Your warranty claim has been received and is being processed.",
				Timestamp: submittedAt.Add(time.Hour),
				Sender:    "Support Team",
				SenderType: "agent",
				Read:      true,
			},
			{
				ID:        "comm-002",
				Type:      "email",
				Direction: "outbound",
				Subject:   "Technical Review Update",
				Message:   "Our technical team has reviewed your claim and confirmed the issue. We will proceed with the repair process.",
				Timestamp: now.AddDate(0, 0, -1),
				Sender:    "Sarah Johnson",
				SenderType: "agent",
				Read:      true,
			},
		},
	}
}

// ToCustomerClaimAttachmentUploadResponse converts to attachment upload response
func ToCustomerClaimAttachmentUploadResponse(claimID, attachmentID string) CustomerClaimAttachmentUploadResponse {
	return CustomerClaimAttachmentUploadResponse{
		AttachmentID: attachmentID,
		UploadURL:    fmt.Sprintf("https://uploads.example.com/claims/%s/attachments/%s", claimID, attachmentID),
		ExpiresAt:    time.Now().Add(time.Hour * 24),
		MaxFileSize:  10485760, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "video/mp4", "application/pdf"},
	}
}

// ToCustomerClaimFeedbackResponse converts to feedback submission response
func ToCustomerClaimFeedbackResponse(claimID, feedbackID string, rating int) CustomerClaimFeedbackResponse {
	averageRating := float64(rating)
	
	return CustomerClaimFeedbackResponse{
		FeedbackID:      feedbackID,
		ClaimID:         claimID,
		SubmittedAt:     time.Now(),
		AverageRating:   averageRating,
		ThankYouMessage: "Thank you for your feedback! Your input helps us improve our service quality.",
		FollowUpSurvey:  nil, // Optional follow-up survey
	}
}

// ToCustomerClaimUpdateResponse converts to claim update response
func ToCustomerClaimUpdateResponse(claimID string, updatedFields []string) CustomerClaimUpdateResponse {
	return CustomerClaimUpdateResponse{
		ClaimID:       claimID,
		UpdatedAt:     time.Now(),
		UpdatedFields: updatedFields,
		Message:       "Your claim information has been updated successfully.",
	}
}

// Helper function to calculate days between dates
func daysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

// Helper function to determine claim priority based on severity and issue type
func determinePriority(severity, issueType string) string {
	if severity == "critical" {
		return "high"
	}
	if severity == "high" && (issueType == "defect" || issueType == "malfunction") {
		return "high"
	}
	if severity == "medium" {
		return "medium"
	}
	return "low"
}

// Helper function to generate next steps based on claim status
func generateNextSteps(status string) []string {
	switch status {
	case "submitted":
		return []string{
			"Your claim has been submitted successfully",
			"Our technical team will review your claim within 24 hours",
			"You will receive an email confirmation shortly",
			"Track your claim status using the provided tracking number",
		}
	case "under_review":
		return []string{
			"Your claim is currently under technical review",
			"Our experts are analyzing the reported issue",
			"You may be contacted for additional information",
			"Expected review completion within 2-3 business days",
		}
	case "approved":
		return []string{
			"Your claim has been approved for processing",
			"Repair/replacement process will begin shortly",
			"You will receive shipping instructions if applicable",
			"Estimated completion time: 5-7 business days",
		}
	case "in_progress":
		return []string{
			"Your claim is currently being processed",
			"Repair/replacement is in progress",
			"You will be notified of any updates",
			"Track progress through your customer portal",
		}
	default:
		return []string{
			"Please check your claim status for updates",
			"Contact support if you have any questions",
		}
	}
}