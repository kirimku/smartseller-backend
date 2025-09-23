package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// Role-based permissions mapping
var rolePermissions = map[entity.UserRole][]entity.Permission{
	entity.UserRoleOwner: {
		entity.PermissionCreateUser,
		entity.PermissionReadUser,
		entity.PermissionUpdateUser,
		entity.PermissionDeleteUser,
		entity.PermissionManageRoles,
		entity.PermissionViewTransactions,
		entity.PermissionManageSystem,
	},
	entity.UserRoleAdmin: {
		entity.PermissionCreateUser,
		entity.PermissionReadUser,
		entity.PermissionUpdateUser,
		entity.PermissionDeleteUser,
		entity.PermissionViewTransactions,
		entity.PermissionManageSystem,
	},
	entity.UserRoleManager: {
		entity.PermissionReadUser,
		entity.PermissionUpdateUser,
		entity.PermissionViewTransactions,
	},
	entity.UserRoleSupport: {
		entity.PermissionReadUser,
		entity.PermissionViewTransactions,
	},
	entity.UserRoleUser: {},
}

// HasPermission checks if a user role has a specific permission
func HasPermission(role entity.UserRole, permission entity.Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// IsUserAuthorized checks if the user in the current context has the specified permission
func IsUserAuthorized(c *gin.Context, permission entity.Permission) bool {
	// First, check for legacy admin flag (for backward compatibility)
	if IsAdminUser(c) {
		return true
	}

	// Then check for role-based permissions
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(entity.UserRole)
	if !ok {
		return false
	}

	return HasPermission(role, permission)
}

// CanDeleteUser checks if the user has permission to delete users
func CanDeleteUser(c *gin.Context) bool {
	return IsUserAuthorized(c, entity.PermissionDeleteUser)
}

// CanManageRoles checks if the user has permission to manage roles
func CanManageRoles(c *gin.Context) bool {
	return IsUserAuthorized(c, entity.PermissionManageRoles)
}
