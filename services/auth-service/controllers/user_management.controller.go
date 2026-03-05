package controllers

import (
	"auth-service/config"
	"auth-service/middleware"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ListUsersInput represents the query parameters for listing users
type ListUsersInput struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search"`
	Role      string `form:"role"`
	IsActive  *bool  `form:"is_active"`
	IsBlocked *bool  `form:"is_blocked"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// UserResponse represents the response structure for user data
type UserResponse struct {
	ID            string                   `json:"id"`
	Email         *string                  `json:"email"`
	Username      string                   `json:"username"`
	FirstName     string                   `json:"first_name"`
	LastName      string                   `json:"last_name"`
	Phone         *string                  `json:"phone"`
	DateOfBirth   *time.Time               `json:"date_of_birth"`
	Gender        *string                  `json:"gender"`
	IsActive      bool                     `json:"is_active"`
	IsBlocked     bool                     `json:"is_blocked"`
	BlockedAt     *time.Time               `json:"blocked_at,omitempty"`
	BlockedReason *string                  `json:"blocked_reason,omitempty"`
	LastLogin     *time.Time               `json:"last_login"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     *time.Time               `json:"updated_at"`
	Roles         []map[string]interface{} `json:"roles"`
}

// ListUsers retrieves all users with filtering and pagination
func ListUsers(c *gin.Context) {
	adminID := c.GetString("user_id")

	var input ListUsersInput
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
		input.SortBy = "created_at"
	}
	if input.SortOrder == "" {
		input.SortOrder = "DESC"
	}

	// Validate sort order
	if input.SortOrder != "ASC" && input.SortOrder != "DESC" {
		middleware.SendValidationError(c, "Invalid sort order", "Sort order must be ASC or DESC")
		return
	}

	// Validate sort by field
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"username":   true,
		"last_login": true,
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
			fmt.Sprintf("(u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.username ILIKE $%d OR u.email ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// IsActive filter
	if input.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	// IsBlocked filter
	if input.IsBlocked != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_blocked = $%d", argIndex))
		args = append(args, *input.IsBlocked)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total users
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u %s", whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count users")
		return
	}

	// Calculate offset
	offset := (input.Page - 1) * input.PageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.username, u.first_name, u.last_name, u.phone, 
		       u.date_of_birth, u.gender, u.is_active, u.is_blocked, u.blocked_at, 
		       u.blocked_reason, u.last_login, u.created_at, u.updated_at
		FROM users u
		%s
		ORDER BY u.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, input.SortBy, input.SortOrder, argIndex, argIndex+1)

	args = append(args, input.PageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch users")
		return
	}
	defer rows.Close()

	users := make([]UserResponse, 0, input.PageSize) // Prevent memory reallocation loops
	var userIDs []string

	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.FirstName, &user.LastName,
			&user.Phone, &user.DateOfBirth, &user.Gender, &user.IsActive, &user.IsBlocked,
			&user.BlockedAt, &user.BlockedReason, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}

		user.Roles = make([]map[string]interface{}, 0)
		users = append(users, user)
		userIDs = append(userIDs, user.ID)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Error processing users")
		return
	}

	// SUPER OPTIMIZATION: Fix N+1 Query Problem for Roles
	if len(userIDs) > 0 {
		placeholders := make([]string, len(userIDs))
		roleArgs := make([]interface{}, len(userIDs))
		for i, id := range userIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			roleArgs[i] = id
		}

		rolesQuery := fmt.Sprintf(`
			SELECT ur.user_id, r.id, r.name, r.description, r.permissions, 
			       ur.organization_id, ur.clinic_id, ur.service_id, ur.is_active
			FROM roles r
			JOIN user_roles ur ON ur.role_id = r.id
			WHERE ur.user_id IN (%s)
		`, strings.Join(placeholders, ","))

		roleRows, err := config.DB.QueryContext(ctx, rolesQuery, roleArgs...)
		if err == nil {
			defer roleRows.Close()

			// Map roles fast by UserID
			roleMap := make(map[string][]map[string]interface{})

			for roleRows.Next() {
				var uid, roleID, roleName string
				var description *string
				var permissionsJSON []byte
				var orgID, clinicID, serviceID *string
				var isActive bool

				if err := roleRows.Scan(&uid, &roleID, &roleName, &description, &permissionsJSON,
					&orgID, &clinicID, &serviceID, &isActive); err != nil {
					continue
				}

				var permissions map[string]interface{}
				_ = json.Unmarshal(permissionsJSON, &permissions)

				role := map[string]interface{}{
					"id":          roleID,
					"name":        roleName,
					"description": description,
					"permissions": permissions,
					"is_active":   isActive,
				}

				if orgID != nil {
					role["organization_id"] = *orgID
				}
				if clinicID != nil {
					role["clinic_id"] = *clinicID
				}
				if serviceID != nil {
					role["service_id"] = *serviceID
				}

				roleMap[uid] = append(roleMap[uid], role)
			}

			// Apply mapped roles dynamically to slice
			for i := range users {
				if rMap, exists := roleMap[users[i].ID]; exists {
					users[i].Roles = rMap
				}
			}
		}
	}

	// Filter purely in memory if specific role filter was supplied
	if input.Role != "" {
		filteredUsers := make([]UserResponse, 0, len(users))
		for _, user := range users {
			hasRole := false
			for _, role := range user.Roles {
				if name, ok := role["name"].(string); ok && name == input.Role {
					hasRole = true
					break
				}
			}
			if hasRole {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
		totalCount = len(users) // Note: this count modifies paginated total
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "LIST_USERS", "Listed users", c)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        input.Page,
			"page_size":   input.PageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + input.PageSize - 1) / input.PageSize,
		},
	})
}

// GetUser retrieves a single user by ID
func GetUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user UserResponse
	err := config.DB.QueryRowContext(ctx, `
		SELECT u.id, u.email, u.username, u.first_name, u.last_name, u.phone, 
		       u.date_of_birth, u.gender, u.is_active, u.is_blocked, u.blocked_at, 
		       u.blocked_reason, u.last_login, u.created_at, u.updated_at
		FROM users u
		WHERE u.id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.FirstName, &user.LastName,
		&user.Phone, &user.DateOfBirth, &user.Gender, &user.IsActive, &user.IsBlocked,
		&user.BlockedAt, &user.BlockedReason, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "User")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch user")
		return
	}

	// Fetch roles securely bounded
	roles, err := getUserRoles(userID, "")
	if err == nil {
		user.Roles = roles
	} else {
		user.Roles = make([]map[string]interface{}, 0)
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "VIEW_USER", fmt.Sprintf("Viewed user %s", userID), c)

	c.JSON(http.StatusOK, user)
}

// CreateUserInput represents the input for creating a new user
type CreateUserInput struct {
	FirstName   string   `json:"first_name" binding:"max=50"`
	LastName    string   `json:"last_name" binding:"max=50"`
	Email       string   `json:"email" binding:"omitempty,email"`
	Username    string   `json:"username" binding:"required,min=3,max=30"`
	Phone       string   `json:"phone" binding:"omitempty"`
	Password    string   `json:"password" binding:"required,min=8"`
	DateOfBirth *string  `json:"date_of_birth"`
	Gender      *string  `json:"gender"`
	IsActive    *bool    `json:"is_active"`
	RoleIDs     []string `json:"role_ids" binding:"omitempty"`
}

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	adminID := c.GetString("user_id")

	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Default first_name and last_name to username if not provided
	if input.FirstName == "" {
		input.FirstName = input.Username
	}
	if input.LastName == "" {
		input.LastName = input.Username
	}

	// Validate email format if provided
	if input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(input.Email) {
			middleware.SendValidationError(c, "Invalid email format", "Please provide a valid email address")
			return
		}
	}

	// Validate phone format if provided
	if input.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(input.Phone) {
			middleware.SendValidationError(c, "Invalid phone format", "Please provide a valid phone number")
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if username/email already exists atomically using a single query
	var usernameExists, emailExists bool
	if input.Email != "" {
		_ = config.DB.QueryRowContext(ctx, `SELECT 
            EXISTS(SELECT 1 FROM users WHERE username = $1), 
            EXISTS(SELECT 1 FROM users WHERE email = $2)`,
			input.Username, input.Email).Scan(&usernameExists, &emailExists)

		if usernameExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		if emailExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
	} else {
		_ = config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, input.Username).Scan(&usernameExists)
		if usernameExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
	}

	// Hash password (CPU Bound)
	passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Set default is_active if not provided
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to start database transaction")
		return
	}
	defer tx.Rollback()

	// Insert user
	var userID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (first_name, last_name, email, username, phone, password_hash, 
		                   date_of_birth, gender, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id
	`, input.FirstName, input.LastName, input.Email, input.Username, input.Phone,
		string(passHash), input.DateOfBirth, input.Gender, isActive, adminID).Scan(&userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create user")
		return
	}

	// Assign roles if provided executing inside connection pool transaction
	if len(input.RoleIDs) > 0 {
		for _, roleID := range input.RoleIDs {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO user_roles (user_id, role_id) 
				VALUES ($1, $2)
				ON CONFLICT (user_id, role_id, organization_id, clinic_id, service_id) DO NOTHING
			`, userID, roleID)
			if err != nil {
				c.Header("X-Warning", "User created but some role assignments failed")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to complete user creation safely")
		return
	}

	// Assemble response struct without needing a second DB Query!
	var emailPtr, phonePtr *string
	if input.Email != "" {
		emailPtr = &input.Email
	}
	if input.Phone != "" {
		phonePtr = &input.Phone
	}

	user := UserResponse{
		ID:        userID,
		Username:  input.Username,
		Email:     emailPtr,
		Phone:     phonePtr,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		Roles:     make([]map[string]interface{}, 0),
	}

	// Fetch purely assigned role mappings without fully recalling user payload
	if roles, err := getUserRoles(userID, ""); err == nil {
		user.Roles = roles
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "CREATE_USER", fmt.Sprintf("Created user %s (%s)", input.Username, userID), c)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
}

// UpdateUser updates a user's information
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Validate email format if provided
	if input.Email != nil && *input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*input.Email) {
			middleware.SendValidationError(c, "Invalid email format", "Please provide a valid email address")
			return
		}

		// Check if email already exists for another user
		var exists bool
		err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`,
			*input.Email, userID).Scan(&exists)
		if err != nil {
			middleware.SendDatabaseError(c, "Failed to check email")
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
	}

	// Validate phone format if provided
	if input.Phone != nil && *input.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(*input.Phone) {
			middleware.SendValidationError(c, "Invalid phone format", "Please provide a valid phone number")
			return
		}
	}

	// Build dynamic update query safely mapping via struct
	query := "UPDATE users SET updated_by = $1"
	args := make([]interface{}, 0, 8)
	args = append(args, adminID)
	argIndex := 2

	if input.FirstName != nil {
		query += fmt.Sprintf(", first_name = $%d", argIndex)
		args = append(args, *input.FirstName)
		argIndex++
	}
	if input.LastName != nil {
		query += fmt.Sprintf(", last_name = $%d", argIndex)
		args = append(args, *input.LastName)
		argIndex++
	}
	if input.Email != nil {
		query += fmt.Sprintf(", email = $%d", argIndex)
		args = append(args, *input.Email)
		argIndex++
	}
	if input.Phone != nil {
		query += fmt.Sprintf(", phone = $%d", argIndex)
		args = append(args, *input.Phone)
		argIndex++
	}
	if input.DateOfBirth != nil {
		query += fmt.Sprintf(", date_of_birth = $%d", argIndex)
		args = append(args, *input.DateOfBirth)
		argIndex++
	}
	if input.Gender != nil {
		query += fmt.Sprintf(", gender = $%d", argIndex)
		args = append(args, *input.Gender)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, userID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update user")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Log activity asynchronously
	go logUserActivity(adminID, "UPDATE_USER", fmt.Sprintf("Updated user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user (soft delete by marking as inactive)
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// Prevent super admin from deleting themselves
	if userID == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if user exists
	var exists bool
	err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Proceed securely into Database Blocks
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Transaction failure")
		return
	}
	defer tx.Rollback()

	// Soft delete: mark as inactive and blocked
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET is_active = false, is_blocked = true, blocked_at = CURRENT_TIMESTAMP, 
		    blocked_by = $1, blocked_reason = 'Account deleted by administrator', updated_by = $1
		WHERE id = $2
	`, adminID, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete user")
		return
	}

	// Revoke all active refresh tokens safely via FOR UPDATE skips behind the scene
	_, _ = tx.ExecContext(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	// Persist
	_ = tx.Commit()

	// Log activity asynchronously
	go logUserActivity(adminID, "DELETE_USER", fmt.Sprintf("Deleted user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// BlockUserInput represents the input for blocking a user
type BlockUserInput struct {
	Reason string `json:"reason" binding:"required,min=5,max=500"`
}

// BlockUser blocks a user account
func BlockUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	var input BlockUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Prevent super admin from blocking themselves
	if userID == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block your own account"})
		return
	}

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if user exists
	var exists bool
	err := config.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Internal server fault")
		return
	}
	defer tx.Rollback()

	// Block user
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET is_blocked = true, blocked_at = CURRENT_TIMESTAMP, blocked_by = $1, 
		    blocked_reason = $2, updated_by = $1
		WHERE id = $3
	`, adminID, input.Reason, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to block user")
		return
	}

	// Revoke tokens
	_, _ = tx.ExecContext(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	_ = tx.Commit()

	// Log activity asynchronously
	go logUserActivity(adminID, "BLOCK_USER", fmt.Sprintf("Blocked user %s: %s", userID, input.Reason), c)

	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// UnblockUser unblocks a user account
func UnblockUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Unblock user
	_, err = config.DB.Exec(`
		UPDATE users 
		SET is_blocked = false, blocked_at = NULL, blocked_by = NULL, 
		    blocked_reason = NULL, updated_by = $1
		WHERE id = $2
	`, adminID, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to unblock user")
		return
	}

	// Log activity
	logUserActivity(adminID, "UNBLOCK_USER", fmt.Sprintf("Unblocked user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// ActivateUser activates a user account
func ActivateUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Activate user
	_, err = config.DB.Exec(`
		UPDATE users 
		SET is_active = true, updated_by = $1
		WHERE id = $2
	`, adminID, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to activate user")
		return
	}

	// Log activity
	logUserActivity(adminID, "ACTIVATE_USER", fmt.Sprintf("Activated user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

// DeactivateUser deactivates a user account
func DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// Prevent super admin from deactivating themselves
	if userID == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot deactivate your own account"})
		return
	}

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Deactivate user
	_, err = config.DB.Exec(`
		UPDATE users 
		SET is_active = false, updated_by = $1
		WHERE id = $2
	`, adminID, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to deactivate user")
		return
	}

	// Revoke all active refresh tokens
	_, err = config.DB.Exec(`
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	// Log activity
	logUserActivity(adminID, "DEACTIVATE_USER", fmt.Sprintf("Deactivated user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// AdminChangePasswordInput represents the input for admin password change
type AdminChangePasswordInput struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// AdminChangePassword allows admin to change any user's password
func AdminChangePassword(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	var input AdminChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// Hash new password
	passHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	_, err = config.DB.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_by = $2
		WHERE id = $3
	`, string(passHash), adminID, userID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update password")
		return
	}

	// Revoke all active refresh tokens for security
	_, err = config.DB.Exec(`
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	// Log activity
	logUserActivity(adminID, "ADMIN_CHANGE_PASSWORD", fmt.Sprintf("Changed password for user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully. User must login again."})
}

// AssignRoleInput represents the input for assigning a role
type AssignRoleInput struct {
	RoleID         string  `json:"role_id" binding:"required"`
	OrganizationID *string `json:"organization_id"`
	ClinicID       *string `json:"clinic_id"`
	ServiceID      *string `json:"service_id"`
}

// AssignRole assigns a role to a user
func AssignRole(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	var input AssignRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "User")
		return
	}

	// ✅ SECURITY: Get role details including name to check privilege escalation
	var roleName string
	err = config.DB.QueryRow(`
		SELECT name FROM roles 
		WHERE id = $1 AND is_active = true
	`, input.RoleID).Scan(&roleName)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Role")
			return
		}
		middleware.SendDatabaseError(c, "Failed to check role")
		return
	}

	// ✅ SECURITY: Validate role assignment scope and prevent privilege escalation
	if err := validateRoleAssignmentScope(roleName, input, isSuperAdmin, isOrgAdmin, isClinicAdmin, c); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Permission denied",
			"message": err.Error(),
		})
		return
	}

	// Assign role
	_, err = config.DB.Exec(`
		INSERT INTO user_roles (user_id, role_id, organization_id, clinic_id, service_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, role_id, organization_id, clinic_id, service_id) 
		DO UPDATE SET is_active = true
	`, userID, input.RoleID, input.OrganizationID, input.ClinicID, input.ServiceID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to assign role")
		return
	}

	// Log activity (roleName already fetched earlier for validation)
	logUserActivity(adminID, "ASSIGN_ROLE",
		fmt.Sprintf("Assigned role %s to user %s", roleName, userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RemoveRole removes a role from a user
func RemoveRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("role_id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "User not found or outside your scope",
		})
		return
	}

	// Check if assignment exists
	var exists bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)
	`, userID, roleID).Scan(&exists)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check role assignment")
		return
	}
	if !exists {
		middleware.SendNotFoundError(c, "Role assignment")
		return
	}

	// Remove role (soft delete by marking as inactive)
	_, err = config.DB.Exec(`
		UPDATE user_roles 
		SET is_active = false 
		WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to remove role")
		return
	}

	// Get role name for logging
	var roleName string
	config.DB.QueryRow(`SELECT name FROM roles WHERE id = $1`, roleID).Scan(&roleName)

	// Log activity
	logUserActivity(adminID, "REMOVE_ROLE",
		fmt.Sprintf("Removed role %s from user %s", roleName, userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetUserActivityLogs retrieves activity logs for a user
func GetUserActivityLogs(c *gin.Context) {
	userID := c.Param("id")

	page := 1
	pageSize := 50
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "50")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	offset := (page - 1) * pageSize

	// Count total logs
	var totalCount int
	err := config.DB.QueryRow(`
		SELECT COUNT(*) FROM user_activity_logs WHERE user_id = $1
	`, userID).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count activity logs")
		return
	}

	// Fetch logs
	rows, err := config.DB.Query(`
		SELECT ual.id, ual.performed_by, ual.action_type, ual.action_description, 
		       ual.ip_address, ual.user_agent, ual.metadata, ual.created_at,
		       COALESCE(u.first_name || ' ' || u.last_name, 'System') as performed_by_name
		FROM user_activity_logs ual
		LEFT JOIN users u ON u.id = ual.performed_by
		WHERE ual.user_id = $1
		ORDER BY ual.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch activity logs")
		return
	}
	defer rows.Close()

	logs := []map[string]interface{}{}
	for rows.Next() {
		var log struct {
			ID                string
			PerformedBy       *string
			ActionType        string
			ActionDescription string
			IPAddress         *string
			UserAgent         *string
			MetadataJSON      []byte
			CreatedAt         time.Time
			PerformedByName   string
		}

		err := rows.Scan(&log.ID, &log.PerformedBy, &log.ActionType, &log.ActionDescription,
			&log.IPAddress, &log.UserAgent, &log.MetadataJSON, &log.CreatedAt, &log.PerformedByName)
		if err != nil {
			continue
		}

		var metadata map[string]interface{}
		json.Unmarshal(log.MetadataJSON, &metadata)

		logEntry := map[string]interface{}{
			"id":                 log.ID,
			"performed_by":       log.PerformedBy,
			"performed_by_name":  log.PerformedByName,
			"action_type":        log.ActionType,
			"action_description": log.ActionDescription,
			"ip_address":         log.IPAddress,
			"user_agent":         log.UserAgent,
			"metadata":           metadata,
			"created_at":         log.CreatedAt,
		}
		logs = append(logs, logEntry)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
	})
}

// Helper function to get user roles
func getUserRoles(userID string, roleFilter string) ([]map[string]interface{}, error) {
	query := `
		SELECT r.id, r.name, r.description, r.permissions, ur.organization_id, 
		       ur.clinic_id, ur.service_id, ur.is_active
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`
	args := []interface{}{userID}

	if roleFilter != "" {
		query += " AND r.name = $2"
		args = append(args, roleFilter)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Prevents downstream deadlocks internally
	defer cancel()

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []map[string]interface{} // Initialize safely array instead of nil behavior dynamically
	for rows.Next() {
		var roleID, roleName string
		var description *string
		var permissionsJSON []byte
		var orgID, clinicID, serviceID *string
		var isActive bool

		err = rows.Scan(&roleID, &roleName, &description, &permissionsJSON,
			&orgID, &clinicID, &serviceID, &isActive)
		if err != nil {
			continue
		}

		var permissions map[string]interface{}
		_ = json.Unmarshal(permissionsJSON, &permissions)

		role := map[string]interface{}{
			"id":          roleID,
			"name":        roleName,
			"description": description,
			"permissions": permissions,
			"is_active":   isActive,
		}

		if orgID != nil {
			role["organization_id"] = *orgID
		}
		if clinicID != nil {
			role["clinic_id"] = *clinicID
		}
		if serviceID != nil {
			role["service_id"] = *serviceID
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// Helper function to log user activity
func logUserActivity(performedBy, actionType, description string, c *gin.Context) {
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	userID := c.Param("id")
	if userID == "" {
		userID = performedBy // For actions not related to a specific user
	}

	metadata := gin.H{
		"endpoint": c.Request.URL.Path,
		"method":   c.Request.Method,
	}
	metadataJSON, _ := json.Marshal(metadata)

	config.DB.Exec(`
		INSERT INTO user_activity_logs (user_id, performed_by, action_type, action_description, 
		                                 ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, performedBy, actionType, description, ipAddress, userAgent, metadataJSON)
}

// validateUserInScope checks if a user is within the admin's scope
func validateUserInScope(userID, adminID string, isSuperAdmin, isOrgAdmin, isClinicAdmin bool, c *gin.Context) bool {
	if isSuperAdmin {
		return true
	}

	if isOrgAdmin {
		orgIDs, _ := c.Get("organization_ids")
		if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
			placeholders := []string{}
			args := []interface{}{userID}
			for i, orgID := range orgIDList {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
				args = append(args, orgID)
			}

			query := fmt.Sprintf(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles 
					WHERE user_id = $1 
					AND organization_id IN (%s) 
					AND is_active = true
				)
			`, strings.Join(placeholders, ","))

			var exists bool
			config.DB.QueryRow(query, args...).Scan(&exists)
			return exists
		}
	} else if isClinicAdmin {
		clinicIDs, _ := c.Get("clinic_ids")
		if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
			placeholders := []string{}
			args := []interface{}{userID}
			for i, clinicID := range clinicIDList {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
				args = append(args, clinicID)
			}

			query := fmt.Sprintf(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles 
					WHERE user_id = $1 
					AND clinic_id IN (%s) 
					AND is_active = true
				)
			`, strings.Join(placeholders, ","))

			var exists bool
			config.DB.QueryRow(query, args...).Scan(&exists)
			return exists
		}
	}

	return false
}

// validateRoleAssignmentScope checks if admin can assign role with given context
func validateRoleAssignmentScope(roleName string, input AssignRoleInput, isSuperAdmin, isOrgAdmin, isClinicAdmin bool, c *gin.Context) error {
	// Check privilege escalation
	adminRoles := []string{"super_admin", "organization_admin", "clinic_admin"}
	isAdminRole := false
	for _, ar := range adminRoles {
		if roleName == ar {
			isAdminRole = true
			break
		}
	}

	if isAdminRole && !isSuperAdmin {
		return fmt.Errorf("only Super Admin can assign admin-level roles (super_admin, organization_admin, clinic_admin)")
	}

	// Validate scope
	if !isSuperAdmin {
		if isOrgAdmin {
			if input.OrganizationID == nil {
				return fmt.Errorf("organization_id required for organization admin role assignment")
			}

			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok {
				found := false
				for _, oid := range orgIDList {
					if oid == *input.OrganizationID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("cannot assign role in organization outside your scope")
				}
			} else {
				return fmt.Errorf("no organization access")
			}
		} else if isClinicAdmin {
			if input.ClinicID == nil {
				return fmt.Errorf("clinic_id required for clinic admin role assignment")
			}

			clinicIDs, _ := c.Get("clinic_ids")
			if clinicIDList, ok := clinicIDs.([]string); ok {
				found := false
				for _, cid := range clinicIDList {
					if cid == *input.ClinicID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("cannot assign role in clinic outside your scope")
				}
			} else {
				return fmt.Errorf("no clinic access")
			}
		}
	}

	return nil
}

// Helper function to check if user has access to organization
func hasOrganizationAccess(userID, organizationID string, isSuperAdmin bool) bool {
	if isSuperAdmin {
		return true
	}

	var hasAccess bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			WHERE ur.user_id = $1 
			AND ur.organization_id = $2
			AND ur.is_active = true
		)
	`, userID, organizationID).Scan(&hasAccess)

	return err == nil && hasAccess
}

// Helper function to check if user has access to clinic
func hasClinicAccess(userID, clinicID string, isSuperAdmin bool) bool {
	if isSuperAdmin {
		return true
	}

	var hasAccess bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			WHERE ur.user_id = $1 
			AND ur.clinic_id = $2
			AND ur.is_active = true
		)
	`, userID, clinicID).Scan(&hasAccess)

	return err == nil && hasAccess
}

// ScopedListUsers retrieves users based on admin's scope (org/clinic)
func ScopedListUsers(c *gin.Context) {
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	var input ListUsersInput
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
		input.SortBy = "created_at"
	}
	if input.SortOrder == "" {
		input.SortOrder = "DESC"
	}

	// Validate inputs
	if input.SortOrder != "ASC" && input.SortOrder != "DESC" {
		middleware.SendValidationError(c, "Invalid sort order", "Sort order must be ASC or DESC")
		return
	}

	validSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "first_name": true,
		"last_name": true, "email": true, "username": true, "last_login": true,
	}
	if !validSortFields[input.SortBy] {
		middleware.SendValidationError(c, "Invalid sort field", "Invalid sort_by field")
		return
	}

	// Build WHERE clause based on scope
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Scope filtering
	if !isSuperAdmin {
		if isOrgAdmin {
			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
				placeholders := []string{}
				for _, orgID := range orgIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, orgID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`u.id IN (
						SELECT DISTINCT ur.user_id FROM user_roles ur 
						WHERE ur.organization_id IN (%s) AND ur.is_active = true
					)`, strings.Join(placeholders, ",")))
			}
		} else if isClinicAdmin {
			clinicIDs, _ := c.Get("clinic_ids")
			if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
				placeholders := []string{}
				for _, clinicID := range clinicIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, clinicID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`u.id IN (
						SELECT DISTINCT ur.user_id FROM user_roles ur 
						WHERE ur.clinic_id IN (%s) AND ur.is_active = true
					)`, strings.Join(placeholders, ",")))
			}
		}
	}

	// Additional filters
	if input.Search != "" {
		searchPattern := "%" + input.Search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.username ILIKE $%d OR u.email ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if input.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	if input.IsBlocked != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_blocked = $%d", argIndex))
		args = append(args, *input.IsBlocked)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u %s", whereClause)
	var totalCount int
	err := config.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count users")
		return
	}

	// Calculate offset
	offset := (input.Page - 1) * input.PageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.username, u.first_name, u.last_name, u.phone, 
		       u.date_of_birth, u.gender, u.is_active, u.is_blocked, u.blocked_at, 
		       u.blocked_reason, u.last_login, u.created_at, u.updated_at
		FROM users u
		%s
		ORDER BY u.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, input.SortBy, input.SortOrder, argIndex, argIndex+1)

	args = append(args, input.PageSize, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch users")
		return
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.FirstName, &user.LastName,
			&user.Phone, &user.DateOfBirth, &user.Gender, &user.IsActive, &user.IsBlocked,
			&user.BlockedAt, &user.BlockedReason, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	// Fetch roles for users
	for i := range users {
		roles, err := getUserRoles(users[i].ID, input.Role)
		if err == nil {
			users[i].Roles = roles
		} else {
			users[i].Roles = []map[string]interface{}{}
		}
	}

	// Filter by role if specified
	if input.Role != "" {
		filteredUsers := []UserResponse{}
		for _, user := range users {
			for _, role := range user.Roles {
				if roleName, ok := role["name"].(string); ok && roleName == input.Role {
					filteredUsers = append(filteredUsers, user)
					break
				}
			}
		}
		users = filteredUsers
		totalCount = len(users)
	}

	// Log activity
	logUserActivity(adminID, "LIST_USERS_SCOPED", "Listed users in scope", c)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        input.Page,
			"page_size":   input.PageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + input.PageSize - 1) / input.PageSize,
		},
		"scope": gin.H{
			"is_super_admin":        isSuperAdmin,
			"is_organization_admin": isOrgAdmin,
			"is_clinic_admin":       isClinicAdmin,
		},
	})
}
