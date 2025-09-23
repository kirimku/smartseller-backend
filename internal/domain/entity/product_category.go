package entity

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProductCategory represents a hierarchical product category
type ProductCategory struct {
	// Primary identification
	ID uuid.UUID `json:"id" db:"id"`

	// Category information
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Slug        string  `json:"slug" db:"slug"`

	// Hierarchy
	ParentID *uuid.UUID `json:"parent_id" db:"parent_id"`

	// Display and ordering
	SortOrder int  `json:"sort_order" db:"sort_order"`
	IsActive  bool `json:"is_active" db:"is_active"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored in database)
	Level    int                `json:"level" db:"-"`
	Path     []*ProductCategory `json:"path,omitempty" db:"-"`
	Children []*ProductCategory `json:"children,omitempty" db:"-"`
	Parent   *ProductCategory   `json:"parent,omitempty" db:"-"`
}

// NewProductCategory creates a new product category
func NewProductCategory(name, slug string, parentID *uuid.UUID) *ProductCategory {
	return &ProductCategory{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		ParentID:  parentID,
		SortOrder: 0,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ValidateName validates the category name
func (pc *ProductCategory) ValidateName() error {
	if pc.Name == "" {
		return fmt.Errorf("category name is required")
	}

	if len(pc.Name) > 255 {
		return fmt.Errorf("category name cannot exceed 255 characters")
	}

	// Name should contain at least one alphabetic character
	hasLetter := regexp.MustCompile(`[a-zA-Z]`)
	if !hasLetter.MatchString(pc.Name) {
		return fmt.Errorf("category name must contain at least one letter")
	}

	return nil
}

// ValidateSlug validates the category slug
func (pc *ProductCategory) ValidateSlug() error {
	if pc.Slug == "" {
		return fmt.Errorf("category slug is required")
	}

	if len(pc.Slug) > 255 {
		return fmt.Errorf("category slug cannot exceed 255 characters")
	}

	// Slug should be URL-friendly: lowercase letters, numbers, and hyphens only
	validSlug := regexp.MustCompile(`^[a-z0-9\-]+$`)
	if !validSlug.MatchString(pc.Slug) {
		return fmt.Errorf("slug can only contain lowercase letters, numbers, and hyphens")
	}

	// Slug cannot start or end with hyphen
	if strings.HasPrefix(pc.Slug, "-") || strings.HasSuffix(pc.Slug, "-") {
		return fmt.Errorf("slug cannot start or end with a hyphen")
	}

	return nil
}

// Validate performs comprehensive validation of the category
func (pc *ProductCategory) Validate() error {
	// Validate name
	if err := pc.ValidateName(); err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}

	// Validate slug
	if err := pc.ValidateSlug(); err != nil {
		return fmt.Errorf("slug validation failed: %w", err)
	}

	// Sort order should not be negative
	if pc.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	// Cannot be parent of itself
	if pc.ParentID != nil && *pc.ParentID == pc.ID {
		return fmt.Errorf("category cannot be its own parent")
	}

	return nil
}

// GenerateSlug generates a URL-friendly slug from the category name
func (pc *ProductCategory) GenerateSlug() {
	if pc.Name == "" {
		return
	}

	// Convert to lowercase
	slug := strings.ToLower(pc.Name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 255 characters
	if len(slug) > 255 {
		slug = slug[:255]
	}

	pc.Slug = slug
}

// IsRoot returns true if this is a root category (no parent)
func (pc *ProductCategory) IsRoot() bool {
	return pc.ParentID == nil
}

// IsChild returns true if this category has a parent
func (pc *ProductCategory) IsChild() bool {
	return pc.ParentID != nil
}

// HasChildren returns true if this category has child categories
func (pc *ProductCategory) HasChildren() bool {
	return len(pc.Children) > 0
}

// GetAncestors returns the path from root to this category (excluding this category)
func (pc *ProductCategory) GetAncestors() []*ProductCategory {
	if len(pc.Path) <= 1 {
		return []*ProductCategory{}
	}
	// Return all but the last element (which is this category)
	return pc.Path[:len(pc.Path)-1]
}

// GetBreadcrumb returns the full path including this category
func (pc *ProductCategory) GetBreadcrumb() []*ProductCategory {
	return pc.Path
}

// IsDescendantOf checks if this category is a descendant of the given category
func (pc *ProductCategory) IsDescendantOf(ancestor *ProductCategory) bool {
	if ancestor == nil {
		return false
	}

	// Check if the ancestor is in this category's path
	for _, pathCategory := range pc.Path {
		if pathCategory.ID == ancestor.ID {
			return true
		}
	}

	return false
}

// IsAncestorOf checks if this category is an ancestor of the given category
func (pc *ProductCategory) IsAncestorOf(descendant *ProductCategory) bool {
	if descendant == nil {
		return false
	}

	return descendant.IsDescendantOf(pc)
}

// GetLevel returns the hierarchy level (0 for root categories)
func (pc *ProductCategory) GetLevel() int {
	return len(pc.Path) - 1 // Path includes this category, so level is path length - 1
}

// CanMoveTo checks if this category can be moved to the specified parent
func (pc *ProductCategory) CanMoveTo(newParentID *uuid.UUID) error {
	// Cannot move to itself
	if newParentID != nil && *newParentID == pc.ID {
		return fmt.Errorf("category cannot be moved to itself")
	}

	// If moving to root (newParentID is nil), it's always allowed
	if newParentID == nil {
		return nil
	}

	// Cannot move to one of its descendants (would create a cycle)
	// This check should be done at the service level with access to the full hierarchy

	return nil
}

// MoveTo moves this category to a new parent
func (pc *ProductCategory) MoveTo(newParentID *uuid.UUID) error {
	if err := pc.CanMoveTo(newParentID); err != nil {
		return err
	}

	pc.ParentID = newParentID
	pc.UpdatedAt = time.Now()
	return nil
}

// Activate activates the category
func (pc *ProductCategory) Activate() {
	pc.IsActive = true
	pc.UpdatedAt = time.Now()
}

// Deactivate deactivates the category
func (pc *ProductCategory) Deactivate() {
	pc.IsActive = false
	pc.UpdatedAt = time.Now()
}

// UpdateSortOrder updates the sort order
func (pc *ProductCategory) UpdateSortOrder(sortOrder int) error {
	if sortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	pc.SortOrder = sortOrder
	pc.UpdatedAt = time.Now()
	return nil
}

// Update updates the category information
func (pc *ProductCategory) Update(name, description *string) error {
	if name != nil {
		pc.Name = *name
		pc.GenerateSlug() // Auto-regenerate slug when name changes
	}

	if description != nil {
		pc.Description = description
	}

	pc.UpdatedAt = time.Now()

	// Validate after update
	return pc.Validate()
}

// AddChild adds a child category (for building hierarchy)
func (pc *ProductCategory) AddChild(child *ProductCategory) {
	if child == nil {
		return
	}

	// Check if child already exists
	for _, existingChild := range pc.Children {
		if existingChild.ID == child.ID {
			return // Already exists
		}
	}

	pc.Children = append(pc.Children, child)
	child.Parent = pc
}

// RemoveChild removes a child category
func (pc *ProductCategory) RemoveChild(childID uuid.UUID) {
	for i, child := range pc.Children {
		if child.ID == childID {
			// Remove from slice
			pc.Children = append(pc.Children[:i], pc.Children[i+1:]...)
			child.Parent = nil
			break
		}
	}
}

// GetChildBySlug finds a child category by slug
func (pc *ProductCategory) GetChildBySlug(slug string) *ProductCategory {
	for _, child := range pc.Children {
		if child.Slug == slug {
			return child
		}
	}
	return nil
}

// GetFullPath returns the full path as a string (e.g., "Electronics > Computers > Laptops")
func (pc *ProductCategory) GetFullPath() string {
	if len(pc.Path) == 0 {
		return pc.Name
	}

	names := make([]string, len(pc.Path))
	for i, category := range pc.Path {
		names[i] = category.Name
	}

	return strings.Join(names, " > ")
}

// GetSlugPath returns the full slug path (e.g., "electronics/computers/laptops")
func (pc *ProductCategory) GetSlugPath() string {
	if len(pc.Path) == 0 {
		return pc.Slug
	}

	slugs := make([]string, len(pc.Path))
	for i, category := range pc.Path {
		slugs[i] = category.Slug
	}

	return strings.Join(slugs, "/")
}

// ComputeFields calculates computed fields for the category
func (pc *ProductCategory) ComputeFields() {
	pc.Level = pc.GetLevel()
}

// String returns a string representation of the category
func (pc *ProductCategory) String() string {
	return fmt.Sprintf("ProductCategory{ID: %s, Name: %s, Slug: %s, Level: %d}",
		pc.ID.String(), pc.Name, pc.Slug, pc.GetLevel())
}
