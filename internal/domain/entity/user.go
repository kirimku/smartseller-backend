package entity

import (
	"database/sql"
	"time"
)

// UserType represents the type of user
type UserType string

const (
	// UserTypeIndividual represents an individual seller
	UserTypeIndividual UserType = "individual"

	// UserTypeBusiness represents a business seller
	UserTypeBusiness UserType = "business"

	// UserTypeEnterprise represents an enterprise seller
	UserTypeEnterprise UserType = "enterprise"
)

// UserTier represents the tier/level of a user for privilege purposes
type UserTier string

const (
	// UserTierBasic is the basic tier (default)
	UserTierBasic UserTier = "basic"

	// UserTierPremium is the premium tier with additional features
	UserTierPremium UserTier = "premium"

	// UserTierPro is the professional tier with advanced features
	UserTierPro UserTier = "pro"

	// UserTierEnterprise is the enterprise tier with full features
	UserTierEnterprise UserTier = "enterprise"
)

// UserRole represents the role of a user for access control
type UserRole string

const (
	// UserRoleOwner has full system access including all administrative operations
	UserRoleOwner UserRole = "owner"

	// UserRoleAdmin has administrative access to the system
	UserRoleAdmin UserRole = "admin"

	// UserRoleManager has access to manage system operations but with limited admin abilities
	UserRoleManager UserRole = "manager"

	// UserRoleSupport has access to customer support functions
	UserRoleSupport UserRole = "support"

	// UserRoleUser is a regular user with no administrative privileges
	UserRoleUser UserRole = "user"
)

// Permission represents a specific action that can be performed
type Permission string

const (
	// PermissionCreateUser allows creating new users
	PermissionCreateUser Permission = "create_user"

	// PermissionReadUser allows viewing user details
	PermissionReadUser Permission = "read_user"

	// PermissionUpdateUser allows updating user information
	PermissionUpdateUser Permission = "update_user"

	// PermissionDeleteUser allows deleting users
	PermissionDeleteUser Permission = "delete_user"

	// PermissionManageRoles allows changing user roles
	PermissionManageRoles Permission = "manage_roles"

	// PermissionViewReports allows viewing business reports and analytics
	PermissionViewReports Permission = "view_reports"

	// PermissionManageProducts allows managing product inventory
	PermissionManageProducts Permission = "manage_products"

	// PermissionManageOrders allows managing customer orders
	PermissionManageOrders Permission = "manage_orders"

	// PermissionManageSystem allows configuring system settings
	PermissionManageSystem Permission = "manage_system"
)

// User represents a user in the system
type User struct {
	ID                   string         `db:"id" json:"id"`
	GoogleID             string         `db:"google_id" json:"google_id"`
	Name                 string         `db:"name" json:"name"`
	Email                string         `db:"email" json:"email"`
	Phone                string         `db:"phone" json:"phone"`
	Picture              string         `db:"picture" json:"picture"`
	PasswordHash         string         `db:"password_hash" json:"-"`
	PasswordSalt         string         `db:"password_salt" json:"-"`
	PasswordResetToken   sql.NullString `db:"password_reset_token" json:"-"`
	PasswordResetExpires sql.NullTime   `db:"password_reset_expires" json:"-"`
	UserType             UserType       `db:"user_type" json:"user_type"`
	UserTier             UserTier       `db:"user_tier" json:"user_tier"`
	SalesCount           int            `db:"sales_count" json:"sales_count"`
	IsAdmin              bool           `db:"is_admin" json:"is_admin"`
	UserRole             UserRole       `db:"user_role" json:"user_role"`
	RefreshToken         string         `db:"refresh_token" json:"-"`
	AccessToken          string         `db:"access_token" json:"-"`
	TokenExpiry          sql.NullTime   `db:"token_expiry" json:"-"`
	AcceptTerms          bool           `db:"accept_terms" json:"accept_terms"`
	AcceptPromos         bool           `db:"accept_promos" json:"accept_promos"`
	CreatedAt            time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt            sql.NullTime   `db:"deleted_at" json:"-"`
}

// GetUsersRequest represents the request parameters for user listing
type GetUsersRequest struct {
	Page     int
	Limit    int
	Search   string // Search in email and phone
	UserType *UserType
	UserTier *UserTier
	UserRole *UserRole
}
