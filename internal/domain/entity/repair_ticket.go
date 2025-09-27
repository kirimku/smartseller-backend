package entity

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RepairStatus represents the status of a repair ticket
type RepairStatus string

const (
	RepairStatusAssigned     RepairStatus = "assigned"      // Assigned to technician
	RepairStatusInProgress   RepairStatus = "in_progress"   // Repair in progress
	RepairStatusWaitingParts RepairStatus = "waiting_parts" // Waiting for parts
	RepairStatusCompleted    RepairStatus = "completed"     // Repair completed
	RepairStatusFailed       RepairStatus = "failed"        // Repair failed
	RepairStatusCancelled    RepairStatus = "cancelled"     // Repair cancelled
)

// Valid validates the repair status
func (rs RepairStatus) Valid() bool {
	switch rs {
	case RepairStatusAssigned, RepairStatusInProgress, RepairStatusWaitingParts,
		RepairStatusCompleted, RepairStatusFailed, RepairStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of RepairStatus
func (rs RepairStatus) String() string {
	return string(rs)
}

// Value implements the driver.Valuer interface for database storage
func (rs RepairStatus) Value() (driver.Value, error) {
	return string(rs), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (rs *RepairStatus) Scan(value interface{}) error {
	if value == nil {
		*rs = RepairStatusAssigned
		return nil
	}
	if str, ok := value.(string); ok {
		*rs = RepairStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into RepairStatus", value)
}

// PartUsage represents a part used in repair
type PartUsage struct {
	PartNumber  string          `json:"part_number"`
	PartName    string          `json:"part_name"`
	Quantity    int             `json:"quantity"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
	TotalCost   decimal.Decimal `json:"total_cost"`
	Description *string         `json:"description,omitempty"`
	Supplier    *string         `json:"supplier,omitempty"`
}

// TestResult represents test results after repair
type TestResult struct {
	TestName    string     `json:"test_name"`
	Result      string     `json:"result"` // passed, failed, warning
	Description *string    `json:"description,omitempty"`
	TestedAt    time.Time  `json:"tested_at"`
	TestedBy    *uuid.UUID `json:"tested_by,omitempty"`
}

// RepairTicket represents a detailed repair workflow ticket
type RepairTicket struct {
	// Primary identification
	ID      uuid.UUID `json:"id" db:"id"`
	ClaimID uuid.UUID `json:"claim_id" db:"claim_id"`

	// Technician assignment
	TechnicianID uuid.UUID `json:"technician_id" db:"technician_id"`

	// Scheduling
	AssignedAt           time.Time  `json:"assigned_at" db:"assigned_at"`
	StartDate            *time.Time `json:"start_date,omitempty" db:"start_date"`
	TargetCompletionDate *time.Time `json:"target_completion_date,omitempty" db:"target_completion_date"`
	ActualCompletionDate *time.Time `json:"actual_completion_date,omitempty" db:"actual_completion_date"`

	// Repair process tracking
	Status RepairStatus `json:"status" db:"status"`

	// Technical details
	Diagnosis   string   `json:"diagnosis" db:"diagnosis"`
	RepairSteps []string `json:"repair_steps,omitempty" db:"repair_steps"`

	// Parts and labor
	PartsUsed  []PartUsage      `json:"parts_used,omitempty" db:"parts_used"`
	LaborHours decimal.Decimal  `json:"labor_hours" db:"labor_hours"`
	HourlyRate *decimal.Decimal `json:"hourly_rate,omitempty" db:"hourly_rate"`
	PartsCost  decimal.Decimal  `json:"parts_cost" db:"parts_cost"`
	LaborCost  decimal.Decimal  `json:"labor_cost" db:"labor_cost"`
	TotalCost  decimal.Decimal  `json:"total_cost" db:"total_cost"`

	// Quality assurance
	QualityCheckPassed *bool        `json:"quality_check_passed,omitempty" db:"quality_check_passed"`
	QualityNotes       *string      `json:"quality_notes,omitempty" db:"quality_notes"`
	TestResults        []TestResult `json:"test_results,omitempty" db:"test_results"`

	// Documentation
	BeforePhotos  []string `json:"before_photos,omitempty" db:"before_photos"`
	AfterPhotos   []string `json:"after_photos,omitempty" db:"after_photos"`
	ProcessPhotos []string `json:"process_photos,omitempty" db:"process_photos"`

	// Technical notes
	TechnicianNotes *string `json:"technician_notes,omitempty" db:"technician_notes"`
	SupervisorNotes *string `json:"supervisor_notes,omitempty" db:"supervisor_notes"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored in database)
	Duration        string   `json:"duration" db:"-"`
	EfficiencyScore *float64 `json:"efficiency_score,omitempty" db:"-"`
	IsOverdue       bool     `json:"is_overdue" db:"-"`
	CompletionRate  float64  `json:"completion_rate" db:"-"`
}

// NewRepairTicket creates a new repair ticket
func NewRepairTicket(claimID, technicianID uuid.UUID, targetCompletion *time.Time) *RepairTicket {
	now := time.Now()
	return &RepairTicket{
		ID:                   uuid.New(),
		ClaimID:              claimID,
		TechnicianID:         technicianID,
		AssignedAt:           now,
		TargetCompletionDate: targetCompletion,
		Status:               RepairStatusAssigned,
		LaborHours:           decimal.Zero,
		PartsCost:            decimal.Zero,
		LaborCost:            decimal.Zero,
		TotalCost:            decimal.Zero,
		PartsUsed:            []PartUsage{},
		TestResults:          []TestResult{},
		RepairSteps:          []string{},
		BeforePhotos:         []string{},
		AfterPhotos:          []string{},
		ProcessPhotos:        []string{},
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// Validate validates the repair ticket
func (rt *RepairTicket) Validate() error {
	// Required fields
	if rt.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if rt.TechnicianID == uuid.Nil {
		return fmt.Errorf("technician_id is required")
	}

	// Validate status
	if !rt.Status.Valid() {
		return fmt.Errorf("invalid repair status: %s", rt.Status)
	}

	// Validate diagnosis for non-assigned tickets
	if rt.Status != RepairStatusAssigned && rt.Diagnosis == "" {
		return fmt.Errorf("diagnosis is required for repair status: %s", rt.Status)
	}

	// Validate monetary values
	if rt.LaborHours.IsNegative() {
		return fmt.Errorf("labor_hours cannot be negative")
	}
	if rt.PartsCost.IsNegative() {
		return fmt.Errorf("parts_cost cannot be negative")
	}
	if rt.LaborCost.IsNegative() {
		return fmt.Errorf("labor_cost cannot be negative")
	}

	// Validate dates
	if rt.StartDate != nil && rt.StartDate.Before(rt.AssignedAt) {
		return fmt.Errorf("start_date cannot be before assigned_at")
	}
	if rt.ActualCompletionDate != nil && rt.StartDate != nil && rt.ActualCompletionDate.Before(*rt.StartDate) {
		return fmt.Errorf("actual_completion_date cannot be before start_date")
	}

	return nil
}

// Start starts the repair work
func (rt *RepairTicket) Start(diagnosis string) error {
	if rt.Status != RepairStatusAssigned {
		return fmt.Errorf("can only start assigned repairs, current status: %s", rt.Status)
	}

	now := time.Now()
	rt.StartDate = &now
	rt.Diagnosis = diagnosis
	rt.Status = RepairStatusInProgress
	rt.UpdatedAt = now

	return nil
}

// AddRepairStep adds a repair step
func (rt *RepairTicket) AddRepairStep(step string) {
	rt.RepairSteps = append(rt.RepairSteps, step)
	rt.UpdatedAt = time.Now()
}

// AddPartUsage adds a part used in repair
func (rt *RepairTicket) AddPartUsage(partNumber, partName string, quantity int, unitCost decimal.Decimal) {
	totalCost := unitCost.Mul(decimal.NewFromInt(int64(quantity)))

	part := PartUsage{
		PartNumber: partNumber,
		PartName:   partName,
		Quantity:   quantity,
		UnitCost:   unitCost,
		TotalCost:  totalCost,
	}

	rt.PartsUsed = append(rt.PartsUsed, part)
	rt.CalculatePartsCost()
	rt.UpdatedAt = time.Now()
}

// CalculatePartsCost calculates total parts cost
func (rt *RepairTicket) CalculatePartsCost() {
	total := decimal.Zero
	for _, part := range rt.PartsUsed {
		total = total.Add(part.TotalCost)
	}
	rt.PartsCost = total
	rt.CalculateTotalCost()
}

// CalculateLaborCost calculates labor cost
func (rt *RepairTicket) CalculateLaborCost() {
	if rt.HourlyRate != nil {
		rt.LaborCost = rt.LaborHours.Mul(*rt.HourlyRate)
	} else {
		rt.LaborCost = decimal.Zero
	}
	rt.CalculateTotalCost()
}

// CalculateTotalCost calculates total repair cost
func (rt *RepairTicket) CalculateTotalCost() {
	rt.TotalCost = rt.PartsCost.Add(rt.LaborCost)
}

// AddTestResult adds a test result
func (rt *RepairTicket) AddTestResult(testName, result string, description *string, testedBy *uuid.UUID) {
	test := TestResult{
		TestName:    testName,
		Result:      result,
		Description: description,
		TestedAt:    time.Now(),
		TestedBy:    testedBy,
	}

	rt.TestResults = append(rt.TestResults, test)
	rt.UpdatedAt = time.Now()
}

// AddPhoto adds a photo to the appropriate category
func (rt *RepairTicket) AddPhoto(photoURL, category string) error {
	switch category {
	case "before":
		rt.BeforePhotos = append(rt.BeforePhotos, photoURL)
	case "after":
		rt.AfterPhotos = append(rt.AfterPhotos, photoURL)
	case "process":
		rt.ProcessPhotos = append(rt.ProcessPhotos, photoURL)
	default:
		return fmt.Errorf("invalid photo category: %s", category)
	}

	rt.UpdatedAt = time.Now()
	return nil
}

// MarkWaitingForParts marks the repair as waiting for parts
func (rt *RepairTicket) MarkWaitingForParts(notes string) error {
	if rt.Status != RepairStatusInProgress {
		return fmt.Errorf("can only mark waiting for parts while in progress, current status: %s", rt.Status)
	}

	rt.Status = RepairStatusWaitingParts
	if notes != "" {
		rt.TechnicianNotes = &notes
	}
	rt.UpdatedAt = time.Now()

	return nil
}

// ResumeRepair resumes repair after parts arrival
func (rt *RepairTicket) ResumeRepair() error {
	if rt.Status != RepairStatusWaitingParts {
		return fmt.Errorf("can only resume repair from waiting_parts status, current status: %s", rt.Status)
	}

	rt.Status = RepairStatusInProgress
	rt.UpdatedAt = time.Now()

	return nil
}

// Complete completes the repair
func (rt *RepairTicket) Complete(qualityCheckPassed bool, qualityNotes *string) error {
	if rt.Status != RepairStatusInProgress {
		return fmt.Errorf("can only complete in-progress repairs, current status: %s", rt.Status)
	}

	now := time.Now()
	rt.ActualCompletionDate = &now
	rt.QualityCheckPassed = &qualityCheckPassed
	if qualityNotes != nil {
		rt.QualityNotes = qualityNotes
	}
	rt.Status = RepairStatusCompleted
	rt.UpdatedAt = now

	// Recalculate costs
	rt.CalculateLaborCost()

	return nil
}

// MarkFailed marks the repair as failed
func (rt *RepairTicket) MarkFailed(reason string) error {
	if rt.Status == RepairStatusCompleted || rt.Status == RepairStatusFailed || rt.Status == RepairStatusCancelled {
		return fmt.Errorf("cannot mark completed/failed/cancelled repair as failed, current status: %s", rt.Status)
	}

	rt.Status = RepairStatusFailed
	rt.TechnicianNotes = &reason
	rt.UpdatedAt = time.Now()

	return nil
}

// Cancel cancels the repair
func (rt *RepairTicket) Cancel(reason string) error {
	if rt.Status == RepairStatusCompleted {
		return fmt.Errorf("cannot cancel completed repair")
	}

	rt.Status = RepairStatusCancelled
	rt.SupervisorNotes = &reason
	rt.UpdatedAt = time.Now()

	return nil
}

// ComputeFields calculates computed fields
func (rt *RepairTicket) ComputeFields() {
	now := time.Now()

	// Calculate duration
	if rt.ActualCompletionDate != nil && rt.StartDate != nil {
		duration := rt.ActualCompletionDate.Sub(*rt.StartDate)
		rt.Duration = formatDuration(duration)
	} else if rt.StartDate != nil {
		duration := now.Sub(*rt.StartDate)
		rt.Duration = formatDuration(duration)
	}

	// Check if overdue
	if rt.TargetCompletionDate != nil && rt.Status != RepairStatusCompleted && rt.Status != RepairStatusCancelled {
		rt.IsOverdue = now.After(*rt.TargetCompletionDate)
	}

	// Calculate completion rate (based on repair steps)
	if len(rt.RepairSteps) > 0 {
		rt.CompletionRate = float64(len(rt.RepairSteps)) / 10.0 * 100 // Assuming 10 steps for full repair
		if rt.CompletionRate > 100 {
			rt.CompletionRate = 100
		}
	}

	// Calculate efficiency score (if completed)
	if rt.Status == RepairStatusCompleted && rt.TargetCompletionDate != nil && rt.ActualCompletionDate != nil {
		targetDuration := rt.TargetCompletionDate.Sub(rt.AssignedAt).Hours()
		actualDuration := rt.ActualCompletionDate.Sub(rt.AssignedAt).Hours()

		if actualDuration > 0 {
			efficiency := (targetDuration / actualDuration) * 100
			if efficiency > 200 {
				efficiency = 200 // Cap at 200%
			}
			rt.EfficiencyScore = &efficiency
		}
	}
}

// formatDuration formats duration into human-readable string
func formatDuration(d time.Duration) string {
	if d.Hours() < 1 {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	} else if d.Hours() < 24 {
		return fmt.Sprintf("%.1f hours", d.Hours())
	} else {
		return fmt.Sprintf("%.1f days", d.Hours()/24)
	}
}

// IsCompleted checks if the repair is completed
func (rt *RepairTicket) IsCompleted() bool {
	return rt.Status == RepairStatusCompleted
}

// CanTransitionTo checks if the repair can transition to the specified status
func (rt *RepairTicket) CanTransitionTo(newStatus RepairStatus) bool {
	switch rt.Status {
	case RepairStatusAssigned:
		return newStatus == RepairStatusInProgress || newStatus == RepairStatusCancelled
	case RepairStatusInProgress:
		return newStatus == RepairStatusWaitingParts || newStatus == RepairStatusCompleted ||
			newStatus == RepairStatusFailed || newStatus == RepairStatusCancelled
	case RepairStatusWaitingParts:
		return newStatus == RepairStatusInProgress || newStatus == RepairStatusCancelled
	case RepairStatusCompleted, RepairStatusFailed, RepairStatusCancelled:
		return false // Terminal statuses
	default:
		return false
	}
}

// String returns a string representation of the repair ticket
func (rt *RepairTicket) String() string {
	return fmt.Sprintf("RepairTicket{ID: %s, Claim: %s, Status: %s, Technician: %s}",
		rt.ID.String(), rt.ClaimID.String(), rt.Status, rt.TechnicianID.String())
}
