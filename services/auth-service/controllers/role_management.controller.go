package controllers

import (
	"auth-service/config"
	"auth-service/middleware"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RoleResponse represents the response structure for role data
type RoleResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	Permissions  map[string]interface{} `json:"permissions"`
	IsSystemRole bool                   `json:"is_system_role"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    *string                `json:"updated_at"`
	UsersCount   int                    `json:"users_count"`
}

// ListRolesInput represents the query parameters for listing roles
type ListRolesInput struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search       string `form:"search"`
	IsActive     *bool  `form:"is_active"`
	IsSystemRole *bool  `form:"is_system_role"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
}

// ListRoles retrieves all roles with filtering and pagination
func ListRoles(c *gin.Context) {
	adminID := c.GetString("user_id")

	var input ListRolesInput
	if err := c.ShouldBindQuery(&input); err != nil {
		middleware.SendValidationError(c, "Invalid query parameters", err.Error())
		return
	}

	// Set defaults
	if input.Page == 0 {
		input.Page = 1
	}
	if input.PageSize == 0 {
		input.PageSize = 20
	}
	if input.SortBy == "" {
		input.SortBy = "name"
	}
	if input.SortOrder == "" {
		input.SortOrder = "ASC"
	}

	// Validate sort order
	if input.SortOrder != "ASC" && input.SortOrder != "DESC" {
		middleware.SendValidationError(c, "Invalid sort order", "Sort order must be ASC or DESC")
		return
	}

	// Validate sort by field
	validSortFields := map[string]bool{
		"name":       true,
		"created_at": true,
		"updated_at": true,
	}
	if !validSortFields[input.SortBy] {
		middleware.SendValidationError(c, "Invalid sort field", "Invalid sort_by field")
		return
	}

	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Search filter
	if input.Search != "" {
		searchPattern := "%" + input.Search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(r.name ILIKE $%d OR r.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// IsActive filter
	if input.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	// IsSystemRole filter
	if input.IsSystemRole != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.is_system_role = $%d", argIndex))
		args = append(args, *input.IsSystemRole)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total roles
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles r %s", whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count roles")
		return
	}

	// Calculate offset
	offset := (input.Page - 1) * input.PageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT r.id, r.name, r.description, r.permissions, r.is_system_role, 
		       r.is_active, r.created_at, r.updated_at,
		       (SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = r.id AND ur.is_active = true) as users_count
		FROM roles r
		%s
		ORDER BY r.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, input.SortBy, input.SortOrder, argIndex, argIndex+1)

	args = append(args, input.PageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch roles")
		return
	}
	defer rows.Close()

	roles := make([]RoleResponse, 0, input.PageSize) // Optimizing slice allocation
	for rows.Next() {
		var role RoleResponse
		var permissionsJSON []byte
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &permissionsJSON,
			&role.IsSystemRole, &role.IsActive, &createdAt, &updatedAt,
			&role.UsersCount,
		)
		if err != nil {
			continue
		}

		// Parse permissions JSON
		var permissions map[string]interface{}
		if err := json.Unmarshal(permissionsJSON, &permissions); err != nil {
			permissions = make(map[string]interface{})
		}
		role.Permissions = permissions

		if createdAt.Valid {
			role.CreatedAt = createdAt.String
		}
		if updatedAt.Valid {
			role.UpdatedAt = &updatedAt.String
		}

		roles = append(roles, role)
	}

	// Check loop condition failures
	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Failed to process roles fully")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "LIST_ROLES", "Listed roles", c)

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"pagination": gin.H{
			"page":        input.Page,
			"page_size":   input.PageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + input.PageSize - 1) / input.PageSize,
		},
	})
}

// GetRole retrieves a single role by ID
func GetRole(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var role RoleResponse
	var permissionsJSON []byte
	var createdAt, updatedAt sql.NullString

	err := config.DB.QueryRowContext(ctx, `
		SELECT r.id, r.name, r.description, r.permissions, r.is_system_role, 
		       r.is_active, r.created_at, r.updated_at,
		       (SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = r.id AND ur.is_active = true) as users_count
		FROM roles r
		WHERE r.id = $1
	`, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &permissionsJSON,
		&role.IsSystemRole, &role.IsActive, &createdAt, &updatedAt,
		&role.UsersCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch role")
		return
	}

	// Parse permissions JSON
	var permissions map[string]interface{}
	if err := json.Unmarshal(permissionsJSON, &permissions); err != nil {
		permissions = make(map[string]interface{})
	}
	role.Permissions = permissions

	if createdAt.Valid {
		role.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		role.UpdatedAt = &updatedAt.String
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "VIEW_ROLE", fmt.Sprintf("Viewed role %s", roleID), c)

	c.JSON(http.StatusOK, role)
}

// CreateRoleInput represents the input for creating a new role
type CreateRoleInput struct {
	Name        string                 `json:"name" binding:"required,min=3,max=50"`
	Description *string                `json:"description"`
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
}

// CreateRole creates a new role
func CreateRole(c *gin.Context) {
	adminID := c.GetString("user_id")

	var input CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Validate role name (lowercase with underscores)
	roleName := strings.ToLower(strings.ReplaceAll(input.Name, " ", "_"))

	// Check if role already exists
	var exists bool
	err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE name = $1)`, roleName).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Role already exists"})
		return
	}

	// Serialize permissions
	permissionsJSON, err := json.Marshal(input.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize permissions"})
		return
	}

	// Insert role
	var roleID string
	err = config.DB.QueryRowContext(ctx, `
		INSERT INTO roles (name, description, permissions, is_system_role, created_by)
		VALUES ($1, $2, $3, false, $4)
		RETURNING id
	`, roleName, input.Description, permissionsJSON, adminID).Scan(&roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create role")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "CREATE_ROLE", fmt.Sprintf("Created role %s (%s)", roleName, roleID), c)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"role_id": roleID,
		"name":    roleName,
	})
}

// UpdateRoleInput represents the input for updating a role
type UpdateRoleInput struct {
	Name        *string                 `json:"name"`
	Description *string                 `json:"description"`
	Permissions *map[string]interface{} `json:"permissions"`
}

// UpdateRole updates a role's information
func UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if role exists and is not a system role
	var isSystemRole bool
	err := config.DB.QueryRowContext(ctx, `SELECT is_system_role FROM roles WHERE id = $1`, roleID).Scan(&isSystemRole)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}

	if isSystemRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system roles"})
		return
	}

	// Build dynamic update query
	query := "UPDATE roles SET updated_by = $1"
	args := []interface{}{adminID}
	argIndex := 2

	if input.Name != nil {
		roleName := strings.ToLower(strings.ReplaceAll(*input.Name, " ", "_"))
		query += fmt.Sprintf(", name = $%d", argIndex)
		args = append(args, roleName)
		argIndex++
	}
	if input.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, *input.Description)
		argIndex++
	}
	if input.Permissions != nil {
		permissionsJSON, err := json.Marshal(*input.Permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize permissions"})
			return
		}
		query += fmt.Sprintf(", permissions = $%d", argIndex)
		args = append(args, permissionsJSON)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, roleID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update role")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.SendNotFoundError(c, "Role")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "UPDATE_ROLE", fmt.Sprintf("Updated role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// DeleteRole deletes a role (soft delete by marking as inactive)
func DeleteRole(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if role exists and is not a system role
	var isSystemRole bool
	var usersCount int
	err := config.DB.QueryRowContext(ctx, `
		SELECT r.is_system_role, 
		       (SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = r.id AND ur.is_active = true)
		FROM roles r
		WHERE r.id = $1
	`, roleID).Scan(&isSystemRole, &usersCount)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}

	if isSystemRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system roles"})
		return
	}

	if usersCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Cannot delete role. %d users are assigned to this role", usersCount),
		})
		return
	}

	// Soft delete: mark as inactive
	_, err = config.DB.ExecContext(ctx, `
		UPDATE roles 
		SET is_active = false, updated_by = $1
		WHERE id = $2
	`, adminID, roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete role")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "DELETE_ROLE", fmt.Sprintf("Deleted role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// ActivateRole activates a role
func ActivateRole(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if role exists
	var exists bool
	err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "Role")
		return
	}

	// Activate role
	_, err = config.DB.ExecContext(ctx, `
		UPDATE roles 
		SET is_active = true, updated_by = $1
		WHERE id = $2
	`, adminID, roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to activate role")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "ACTIVATE_ROLE", fmt.Sprintf("Activated role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role activated successfully"})
}

// DeactivateRole deactivates a role
func DeactivateRole(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	// Check if role exists and is not a system role
	var isSystemRole bool
	err := config.DB.QueryRow(`SELECT is_system_role FROM roles WHERE id = $1`, roleID).Scan(&isSystemRole)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}

	if isSystemRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot deactivate system roles"})
		return
	}

	// Deactivate role
	_, err = config.DB.Exec(`
		UPDATE roles 
		SET is_active = false, updated_by = $1
		WHERE id = $2
	`, adminID, roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to deactivate role")
		return
	}

	// Log activity
	logUserActivity(adminID, "DEACTIVATE_ROLE", fmt.Sprintf("Deactivated role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role deactivated successfully"})
}

// UpdateRolePermissionsInput represents the input for updating role permissions
type UpdateRolePermissionsInput struct {
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
}

// UpdateRolePermissions updates only the permissions of a role
func UpdateRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	var input UpdateRolePermissionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Check if role exists and is not a system role
	var isSystemRole bool
	err := config.DB.QueryRow(`SELECT is_system_role FROM roles WHERE id = $1`, roleID).Scan(&isSystemRole)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}

	if isSystemRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify permissions of system roles"})
		return
	}

	// Serialize permissions
	permissionsJSON, err := json.Marshal(input.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize permissions"})
		return
	}

	// Update permissions
	_, err = config.DB.Exec(`
		UPDATE roles 
		SET permissions = $1, updated_by = $2
		WHERE id = $3
	`, permissionsJSON, adminID, roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update role permissions")
		return
	}

	// Log activity
	logUserActivity(adminID, "UPDATE_ROLE_PERMISSIONS", fmt.Sprintf("Updated permissions for role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role permissions updated successfully"})
}

// GetRoleUsers retrieves all users assigned to a specific role
func GetRoleUsers(c *gin.Context) {
	roleID := c.Param("id")
	adminID := c.GetString("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	// Single query for Count + Existence check + Data fetch to prevent three round trips
	var exists bool
	err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "Role")
		return
	}

	var totalCount int
	err = config.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM user_roles ur 
		WHERE ur.role_id = $1 AND ur.is_active = true
	`, roleID).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count users")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch users
	rows, err := config.DB.QueryContext(ctx, `
		SELECT u.id, u.email, u.username, u.first_name, u.last_name, u.phone, 
		       u.is_active, u.is_blocked, u.last_login, u.created_at,
		       ur.organization_id, ur.clinic_id, ur.service_id, ur.assigned_at
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		WHERE ur.role_id = $1 AND ur.is_active = true
		ORDER BY ur.assigned_at DESC
		LIMIT $2 OFFSET $3
	`, roleID, pageSize, offset)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch users")
		return
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0, pageSize)
	for rows.Next() {
		var userID, username, firstName, lastName string
		var email, phone *string
		var isActive, isBlocked bool
		var lastLogin, createdAt, assignedAt sql.NullTime
		var orgID, clinicID, serviceID *string

		err := rows.Scan(&userID, &email, &username, &firstName, &lastName, &phone,
			&isActive, &isBlocked, &lastLogin, &createdAt,
			&orgID, &clinicID, &serviceID, &assignedAt)
		if err != nil {
			continue
		}

		user := map[string]interface{}{
			"id":         userID,
			"email":      email,
			"username":   username,
			"first_name": firstName,
			"last_name":  lastName,
			"phone":      phone,
			"is_active":  isActive,
			"is_blocked": isBlocked,
		}

		if lastLogin.Valid {
			user["last_login"] = lastLogin.Time
		}
		if createdAt.Valid {
			user["created_at"] = createdAt.Time
		}
		if assignedAt.Valid {
			user["assigned_at"] = assignedAt.Time
		}
		if orgID != nil {
			user["organization_id"] = *orgID
		}
		if clinicID != nil {
			user["clinic_id"] = *clinicID
		}
		if serviceID != nil {
			user["service_id"] = *serviceID
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Failed to complete processing role users")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "VIEW_ROLE_USERS", fmt.Sprintf("Viewed users for role %s", roleID), c)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
	})
}

// GetPermissionTemplates returns common permission templates for different types of roles
func GetPermissionTemplates(c *gin.Context) {
	templates := map[string]interface{}{
		"templates": []map[string]interface{}{
			{
				"name":        "Admin Role",
				"description": "Full access to manage users, roles, and system settings",
				"permissions": map[string]interface{}{
					"users":         []string{"create", "read", "update", "delete"},
					"roles":         []string{"create", "read", "update", "delete"},
					"organizations": []string{"create", "read", "update", "delete"},
					"clinics":       []string{"create", "read", "update", "delete"},
					"services":      []string{"create", "read", "update", "delete"},
				},
			},
			{
				"name":        "Manager Role",
				"description": "Can view and manage resources within assigned scope",
				"permissions": map[string]interface{}{
					"users":   []string{"create", "read", "update"},
					"roles":   []string{"read"},
					"clinics": []string{"read", "update"},
					"staff":   []string{"create", "read", "update"},
				},
			},
			{
				"name":        "Staff Role",
				"description": "Can view and manage daily operations",
				"permissions": map[string]interface{}{
					"patients":     []string{"read", "create", "update"},
					"appointments": []string{"read", "create", "update"},
					"billing":      []string{"read", "create"},
				},
			},
			{
				"name":        "Viewer Role",
				"description": "Read-only access to resources",
				"permissions": map[string]interface{}{
					"users":        []string{"read"},
					"roles":        []string{"read"},
					"patients":     []string{"read"},
					"appointments": []string{"read"},
					"reports":      []string{"read"},
				},
			},
		},
	}

	c.JSON(http.StatusOK, templates)
}
