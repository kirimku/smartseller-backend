package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TimelineEventType represents the type of timeline event
type TimelineEventType string

const (
	TimelineEventStatusChange       TimelineEventType = "status_change"
	TimelineEventNoteAdded          TimelineEventType = "note_added"
	TimelineEventAttachmentUploaded TimelineEventType = "attachment_uploaded"
	TimelineEventAssignmentChanged  TimelineEventType = "assignment_changed"
	TimelineEventRepairStarted      TimelineEventType = "repair_started"
	TimelineEventRepairCompleted    TimelineEventType = "repair_completed"
	TimelineEventShipmentCreated    TimelineEventType = "shipment_created"
	TimelineEventDeliveryUpdate     TimelineEventType = "delivery_update"
	TimelineEventCustomerContact    TimelineEventType = "customer_contact"
	TimelineEventSystemUpdate       TimelineEventType = "system_update"
)

// Valid validates the timeline event type
func (tet TimelineEventType) Valid() bool {
	switch tet {
	case TimelineEventStatusChange, TimelineEventNoteAdded, TimelineEventAttachmentUploaded,
		TimelineEventAssignmentChanged, TimelineEventRepairStarted, TimelineEventRepairCompleted,
		TimelineEventShipmentCreated, TimelineEventDeliveryUpdate, TimelineEventCustomerContact,
		TimelineEventSystemUpdate:
		return true
	default:
		return false
	}
}

// String returns the string representation of TimelineEventType
func (tet TimelineEventType) String() string {
	return string(tet)
}

// Value implements the driver.Valuer interface for database storage
func (tet TimelineEventType) Value() (driver.Value, error) {
	return string(tet), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (tet *TimelineEventType) Scan(value interface{}) error {
	if value == nil {
		*tet = TimelineEventSystemUpdate
		return nil
	}
	if str, ok := value.(string); ok {
		*tet = TimelineEventType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into TimelineEventType", value)
}

// ActorType represents the type of actor who performed the action
type ActorType string

const (
	ActorTypeCustomer   ActorType = "customer"
	ActorTypeAdmin      ActorType = "admin"
	ActorTypeTechnician ActorType = "technician"
	ActorTypeSystem     ActorType = "system"
)

// Valid validates the actor type
func (at ActorType) Valid() bool {
	switch at {
	case ActorTypeCustomer, ActorTypeAdmin, ActorTypeTechnician, ActorTypeSystem:
		return true
	default:
		return false
	}
}

// String returns the string representation of ActorType
func (at ActorType) String() string {
	return string(at)
}

// Value implements the driver.Valuer interface for database storage
func (at ActorType) Value() (driver.Value, error) {
	return string(at), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (at *ActorType) Scan(value interface{}) error {
	if value == nil {
		*at = ActorTypeSystem
		return nil
	}
	if str, ok := value.(string); ok {
		*at = ActorType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ActorType", value)
}

// TimelineMetadata represents additional structured data for timeline events
type TimelineMetadata struct {
	AttachmentID     *uuid.UUID             `json:"attachment_id,omitempty"`
	OldValue         *string                `json:"old_value,omitempty"`
	NewValue         *string                `json:"new_value,omitempty"`
	FieldName        *string                `json:"field_name,omitempty"`
	TechnicianID     *uuid.UUID             `json:"technician_id,omitempty"`
	TrackingNumber   *string                `json:"tracking_number,omitempty"`
	ShippingProvider *string                `json:"shipping_provider,omitempty"`
	EstimatedDate    *time.Time             `json:"estimated_date,omitempty"`
	ActualDate       *time.Time             `json:"actual_date,omitempty"`
	ContactMethod    *string                `json:"contact_method,omitempty"`
	ContactReason    *string                `json:"contact_reason,omitempty"`
	Additional       map[string]interface{} `json:"additional,omitempty"`
}

// Value implements driver.Valuer interface for database storage
func (tm TimelineMetadata) Value() (driver.Value, error) {
	return json.Marshal(tm)
}

// Scan implements sql.Scanner interface for database retrieval
func (tm *TimelineMetadata) Scan(value interface{}) error {
	if value == nil {
		*tm = TimelineMetadata{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into TimelineMetadata", value)
	}

	return json.Unmarshal(b, tm)
}

// ClaimTimeline represents an audit trail entry for warranty claim changes
type ClaimTimeline struct {
	// Primary identification
	ID      uuid.UUID `json:"id" db:"id"`
	ClaimID uuid.UUID `json:"claim_id" db:"claim_id"`

	// Event details
	EventType  TimelineEventType `json:"event_type" db:"event_type"`
	FromStatus *string           `json:"from_status,omitempty" db:"from_status"`
	ToStatus   *string           `json:"to_status,omitempty" db:"to_status"`

	// Actor information
	ActorID   *uuid.UUID `json:"actor_id,omitempty" db:"actor_id"`
	ActorType ActorType  `json:"actor_type" db:"actor_type"`

	// Event description and metadata
	Description string           `json:"description" db:"description"`
	Metadata    TimelineMetadata `json:"metadata" db:"metadata"`

	// Visibility control
	IsCustomerVisible bool `json:"is_customer_visible" db:"is_customer_visible"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Computed fields (not stored in database)
	ActorName    string `json:"actor_name" db:"-"`
	RelativeTime string `json:"relative_time" db:"-"`
	Icon         string `json:"icon" db:"-"`
	Color        string `json:"color" db:"-"`
	IsImportant  bool   `json:"is_important" db:"-"`
}

// NewClaimTimeline creates a new claim timeline entry
func NewClaimTimeline(claimID uuid.UUID, eventType TimelineEventType, description string, actorType ActorType) *ClaimTimeline {
	return &ClaimTimeline{
		ID:                uuid.New(),
		ClaimID:           claimID,
		EventType:         eventType,
		Description:       description,
		ActorType:         actorType,
		Metadata:          TimelineMetadata{},
		IsCustomerVisible: true, // Default to visible
		CreatedAt:         time.Now(),
	}
}

// NewStatusChangeEvent creates a timeline entry for status changes
func NewStatusChangeEvent(claimID uuid.UUID, fromStatus, toStatus string, actorID *uuid.UUID, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("Claim status changed from %s to %s", fromStatus, toStatus)

	timeline := NewClaimTimeline(claimID, TimelineEventStatusChange, description, actorType)
	timeline.FromStatus = &fromStatus
	timeline.ToStatus = &toStatus
	timeline.ActorID = actorID
	timeline.Metadata.OldValue = &fromStatus
	timeline.Metadata.NewValue = &toStatus
	timeline.Metadata.FieldName = stringPtr("status")

	return timeline
}

// NewAssignmentEvent creates a timeline entry for technician assignments
func NewAssignmentEvent(claimID, technicianID uuid.UUID, technicianName string, actorID *uuid.UUID, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("Claim assigned to technician %s", technicianName)

	timeline := NewClaimTimeline(claimID, TimelineEventAssignmentChanged, description, actorType)
	timeline.ActorID = actorID
	timeline.Metadata.TechnicianID = &technicianID
	timeline.Metadata.NewValue = &technicianName

	return timeline
}

// NewNoteEvent creates a timeline entry for notes
func NewNoteEvent(claimID uuid.UUID, note string, actorID *uuid.UUID, actorType ActorType, isCustomerVisible bool) *ClaimTimeline {
	description := "Note added to claim"
	if actorType == ActorTypeCustomer {
		description = "Customer added a note"
	}

	timeline := NewClaimTimeline(claimID, TimelineEventNoteAdded, description, actorType)
	timeline.ActorID = actorID
	timeline.IsCustomerVisible = isCustomerVisible
	timeline.Metadata.Additional = map[string]interface{}{
		"note_content": note,
	}

	return timeline
}

// NewAttachmentEvent creates a timeline entry for file uploads
func NewAttachmentEvent(claimID, attachmentID uuid.UUID, filename string, actorID *uuid.UUID, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("File uploaded: %s", filename)

	timeline := NewClaimTimeline(claimID, TimelineEventAttachmentUploaded, description, actorType)
	timeline.ActorID = actorID
	timeline.Metadata.AttachmentID = &attachmentID
	timeline.Metadata.Additional = map[string]interface{}{
		"filename": filename,
	}

	return timeline
}

// NewShipmentEvent creates a timeline entry for shipment creation
func NewShipmentEvent(claimID uuid.UUID, provider, trackingNumber string, estimatedDate *time.Time, actorID *uuid.UUID, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("Item shipped via %s (Tracking: %s)", provider, trackingNumber)

	timeline := NewClaimTimeline(claimID, TimelineEventShipmentCreated, description, actorType)
	timeline.ActorID = actorID
	timeline.Metadata.TrackingNumber = &trackingNumber
	timeline.Metadata.ShippingProvider = &provider
	timeline.Metadata.EstimatedDate = estimatedDate

	return timeline
}

// NewDeliveryEvent creates a timeline entry for delivery updates
func NewDeliveryEvent(claimID uuid.UUID, status string, actualDate *time.Time, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("Delivery status updated: %s", status)

	timeline := NewClaimTimeline(claimID, TimelineEventDeliveryUpdate, description, actorType)
	timeline.Metadata.NewValue = &status
	timeline.Metadata.ActualDate = actualDate

	return timeline
}

// NewContactEvent creates a timeline entry for customer contact
func NewContactEvent(claimID uuid.UUID, method, reason string, actorID *uuid.UUID, actorType ActorType) *ClaimTimeline {
	description := fmt.Sprintf("Customer contacted via %s", method)

	timeline := NewClaimTimeline(claimID, TimelineEventCustomerContact, description, actorType)
	timeline.ActorID = actorID
	timeline.Metadata.ContactMethod = &method
	timeline.Metadata.ContactReason = &reason
	timeline.IsCustomerVisible = false // Contact events are typically internal

	return timeline
}

// NewSystemEvent creates a timeline entry for system updates
func NewSystemEvent(claimID uuid.UUID, description string, metadata map[string]interface{}) *ClaimTimeline {
	timeline := NewClaimTimeline(claimID, TimelineEventSystemUpdate, description, ActorTypeSystem)
	timeline.IsCustomerVisible = false // System events are typically internal
	timeline.Metadata.Additional = metadata

	return timeline
}

// Validate validates the timeline entry
func (ct *ClaimTimeline) Validate() error {
	// Required fields
	if ct.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if ct.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate event type
	if !ct.EventType.Valid() {
		return fmt.Errorf("invalid event type: %s", ct.EventType)
	}

	// Validate actor type
	if !ct.ActorType.Valid() {
		return fmt.Errorf("invalid actor type: %s", ct.ActorType)
	}

	// Status change events should have from/to status
	if ct.EventType == TimelineEventStatusChange {
		if ct.FromStatus == nil || ct.ToStatus == nil {
			return fmt.Errorf("status change events must have from_status and to_status")
		}
	}

	// Non-system events should have actor_id
	if ct.ActorType != ActorTypeSystem && ct.ActorID == nil {
		return fmt.Errorf("non-system events must have actor_id")
	}

	return nil
}

// SetActor sets the actor information
func (ct *ClaimTimeline) SetActor(actorID uuid.UUID, actorType ActorType) {
	ct.ActorID = &actorID
	ct.ActorType = actorType
}

// AddMetadata adds additional metadata
func (ct *ClaimTimeline) AddMetadata(key string, value interface{}) {
	if ct.Metadata.Additional == nil {
		ct.Metadata.Additional = make(map[string]interface{})
	}
	ct.Metadata.Additional[key] = value
}

// SetVisibility sets whether the event is visible to customers
func (ct *ClaimTimeline) SetVisibility(isVisible bool) {
	ct.IsCustomerVisible = isVisible
}

// ComputeFields calculates computed fields
func (ct *ClaimTimeline) ComputeFields() {
	// Calculate relative time
	ct.RelativeTime = formatRelativeTime(ct.CreatedAt)

	// Set icon and color based on event type
	ct.setIconAndColor()

	// Determine if event is important
	ct.IsImportant = ct.isImportantEvent()
}

// setIconAndColor sets the icon and color based on event type
func (ct *ClaimTimeline) setIconAndColor() {
	switch ct.EventType {
	case TimelineEventStatusChange:
		ct.Icon = "status"
		ct.Color = "blue"
		if ct.ToStatus != nil {
			switch *ct.ToStatus {
			case "completed":
				ct.Color = "green"
			case "rejected", "cancelled":
				ct.Color = "red"
			case "in_repair":
				ct.Color = "orange"
			}
		}
	case TimelineEventAssignmentChanged:
		ct.Icon = "user"
		ct.Color = "purple"
	case TimelineEventNoteAdded:
		ct.Icon = "note"
		ct.Color = "gray"
	case TimelineEventAttachmentUploaded:
		ct.Icon = "attachment"
		ct.Color = "blue"
	case TimelineEventRepairStarted:
		ct.Icon = "tool"
		ct.Color = "orange"
	case TimelineEventRepairCompleted:
		ct.Icon = "check"
		ct.Color = "green"
	case TimelineEventShipmentCreated:
		ct.Icon = "truck"
		ct.Color = "blue"
	case TimelineEventDeliveryUpdate:
		ct.Icon = "package"
		ct.Color = "green"
	case TimelineEventCustomerContact:
		ct.Icon = "phone"
		ct.Color = "indigo"
	case TimelineEventSystemUpdate:
		ct.Icon = "system"
		ct.Color = "gray"
	default:
		ct.Icon = "info"
		ct.Color = "gray"
	}
}

// isImportantEvent determines if the event is important for highlighting
func (ct *ClaimTimeline) isImportantEvent() bool {
	switch ct.EventType {
	case TimelineEventStatusChange:
		if ct.ToStatus != nil {
			switch *ct.ToStatus {
			case "validated", "completed", "rejected", "shipped":
				return true
			}
		}
	case TimelineEventRepairCompleted, TimelineEventShipmentCreated:
		return true
	}
	return false
}

// formatRelativeTime formats time relative to now
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("January 2, 2006")
	}
}

// GetDisplayDescription returns a customer-friendly description
func (ct *ClaimTimeline) GetDisplayDescription() string {
	switch ct.EventType {
	case TimelineEventStatusChange:
		if ct.ToStatus != nil {
			switch *ct.ToStatus {
			case "validated":
				return "Your warranty claim has been approved and is being processed"
			case "rejected":
				return "Your warranty claim has been rejected"
			case "in_repair":
				return "Your product is now being repaired"
			case "repaired":
				return "Repair has been completed successfully"
			case "shipped":
				return "Your item has been shipped back to you"
			case "delivered":
				return "Your item has been delivered"
			case "completed":
				return "Your warranty claim has been completed"
			}
		}
	case TimelineEventShipmentCreated:
		if ct.Metadata.TrackingNumber != nil {
			return fmt.Sprintf("Your item is on its way! Track it with: %s", *ct.Metadata.TrackingNumber)
		}
	}
	return ct.Description
}

// GetMetadataValue gets a specific metadata value
func (ct *ClaimTimeline) GetMetadataValue(key string) interface{} {
	if ct.Metadata.Additional != nil {
		return ct.Metadata.Additional[key]
	}
	return nil
}

// IsStatusChange checks if this is a status change event
func (ct *ClaimTimeline) IsStatusChange() bool {
	return ct.EventType == TimelineEventStatusChange
}

// IsCustomerAction checks if this was an action performed by the customer
func (ct *ClaimTimeline) IsCustomerAction() bool {
	return ct.ActorType == ActorTypeCustomer
}

// String returns a string representation of the claim timeline
func (ct *ClaimTimeline) String() string {
	return fmt.Sprintf("ClaimTimeline{ID: %s, Claim: %s, Event: %s, Actor: %s, Time: %s}",
		ct.ID.String(), ct.ClaimID.String(), ct.EventType, ct.ActorType, ct.CreatedAt.Format(time.RFC3339))
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
