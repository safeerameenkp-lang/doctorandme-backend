package controllers

import (
	"auth-service/config"
	"auth-service/models"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"auth-service/middleware"
)

// HealthCheck endpoint
func HealthCheck(c *gin.Context) {
	// Test database connection
	err := config.DB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "auth-service",
		"timestamp": time.Now().Unix(),
	})
}

type RegisterInput struct {
	FirstName string `json:"first_name" binding:"max=50"`
	LastName  string `json:"last_name" binding:"max=50"`
	Email     string `json:"email" binding:"omitempty,email"`
	Username  string `json:"username" binding:"required,min=1,max=30"`
	Phone     string `json:"phone" binding:"omitempty,len=10"`
	Password  string `json:"password" binding:"required,min=8"`
}

func Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input RegisterInput
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

	if input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(input.Email) {
			middleware.SendValidationError(c, "Invalid email format", "Please provide a valid email address")
			return
		}
	}

	if input.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(input.Phone) {
			middleware.SendValidationError(c, "Invalid phone format", "Please provide a valid phone number")
			return
		}
	}

	// Consolidated existence check (Restored for security)
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

	passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	var userID string
	err = tx.QueryRowContext(ctx, `
        INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
        VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
    `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, string(passHash)).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	var roleID string
	_ = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name='super_admin' LIMIT 1`).Scan(&roleID)
	if roleID != "" {
		_, _ = tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2)`, userID, roleID)
	}

	accessToken, err := middleware.SignAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := middleware.SignRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, userID, refreshToken, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	user := models.User{
		ID:        userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     &input.Email,
		Username:  input.Username,
		Phone:     &input.Phone,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

type LoginInput struct {
	Login    string `json:"login" binding:"required"` // email, phone, or username
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// ✅ Fetch user and check both is_active and is_blocked
	var user models.User
	var isBlocked bool
	err := config.DB.QueryRowContext(ctx, `
        SELECT id, password_hash, first_name, last_name, email, username, phone, is_blocked
        FROM users
        WHERE (email = $1 OR phone = $1 OR username = $1) 
        AND is_active = true
        AND is_blocked = false
    `, input.Login).Scan(&user.ID, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Phone, &isBlocked)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or account blocked"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
			"debug": "Password mismatch",
		})
		return
	}

	// Update last login asynchronously to avoid blocking the login response
	go func(uid string) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer bgCancel()
		_, _ = config.DB.ExecContext(bgCtx, `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1`, uid)
	}(user.ID)

	// Generate tokens
	accessToken, err := middleware.SignAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := middleware.SignRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = config.DB.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, user.ID, refreshToken, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	// Fetch user roles with organization/clinic context
	rows, err := config.DB.QueryContext(ctx, `
        SELECT r.id, r.name, r.permissions, ur.organization_id, ur.clinic_id, ur.service_id
        FROM roles r
        JOIN user_roles ur ON ur.role_id = r.id
        WHERE ur.user_id = $1 AND ur.is_active = true
    `, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles"})
		return
	}
	defer rows.Close()

	var topOrgID, topClinicID, topServiceID interface{}
	var roles []map[string]interface{}
	for rows.Next() {
		var roleID, roleName string
		var permissionsJSON []byte
		var orgID, clinicID, serviceID *string

		err = rows.Scan(&roleID, &roleName, &permissionsJSON, &orgID, &clinicID, &serviceID)
		if err != nil {
			continue
		}

		// Capture first available IDs for top-level response
		if topOrgID == nil && orgID != nil {
			topOrgID = *orgID
		}
		if topClinicID == nil && clinicID != nil {
			topClinicID = *clinicID
		}
		if topServiceID == nil && serviceID != nil {
			topServiceID = *serviceID
		}

		var permissions map[string]interface{}
		if err = json.Unmarshal(permissionsJSON, &permissions); err != nil {
			permissions = make(map[string]interface{})
		}

		role := map[string]interface{}{
			"id":          roleID,
			"name":        roleName,
			"permissions": permissions,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"firstName":      user.FirstName,
		"lastName":       user.LastName,
		"email":          user.Email,
		"username":       user.Username,
		"phone":          user.Phone,
		"organizationId": topOrgID,
		"clinicId":       topClinicID,
		"serviceId":      topServiceID,
		"roles":          roles,
		"accessToken":    accessToken,
		"refreshToken":   refreshToken,
		"tokenType":      "Bearer",
		"expiresIn":      3600,
	})
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	token, err := middleware.VerifyRefreshToken(input.RefreshToken)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	var refreshTokenID string
	err = tx.QueryRowContext(ctx, `
        SELECT id FROM refresh_tokens 
        WHERE user_id = $1 AND token = $2 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL
        FOR UPDATE SKIP LOCKED
    `, userID, input.RefreshToken).Scan(&refreshTokenID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Revoke old token
	_, err = tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`, refreshTokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke old token"})
		return
	}

	// Generate new tokens
	newAccessToken, err := middleware.SignAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	newRefreshToken, err := middleware.SignRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, userID, newRefreshToken, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	// Fetch user details
	var user models.User
	err = tx.QueryRowContext(ctx, `
        SELECT id, first_name, last_name, email, username, phone
        FROM users
        WHERE id = $1 AND is_active = true
    `, userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	// Fetch roles for refreshed user
	rows, err := config.DB.QueryContext(ctx, `
        SELECT r.id, r.name, r.permissions, ur.organization_id, ur.clinic_id, ur.service_id
        FROM roles r
        JOIN user_roles ur ON ur.role_id = r.id
        WHERE ur.user_id = $1 AND ur.is_active = true
    `, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles"})
		return
	}
	defer rows.Close()

	var topOrgID, topClinicID, topServiceID interface{}
	var roles []map[string]interface{}
	for rows.Next() {
		var roleID, roleName string
		var permissionsJSON []byte
		var orgID, clinicID, serviceID *string

		err = rows.Scan(&roleID, &roleName, &permissionsJSON, &orgID, &clinicID, &serviceID)
		if err != nil {
			continue
		}

		// Capture first available IDs for top-level response
		if topOrgID == nil && orgID != nil {
			topOrgID = *orgID
		}
		if topClinicID == nil && clinicID != nil {
			topClinicID = *clinicID
		}
		if topServiceID == nil && serviceID != nil {
			topServiceID = *serviceID
		}

		var permissions map[string]interface{}
		if err = json.Unmarshal(permissionsJSON, &permissions); err != nil {
			permissions = make(map[string]interface{})
		}

		role := map[string]interface{}{
			"id":          roleID,
			"name":        roleName,
			"permissions": permissions,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"firstName":      user.FirstName,
		"lastName":       user.LastName,
		"email":          user.Email,
		"username":       user.Username,
		"phone":          user.Phone,
		"organizationId": topOrgID,
		"clinicId":       topClinicID,
		"serviceId":      topServiceID,
		"roles":          roles,
		"accessToken":    newAccessToken,
		"refreshToken":   newRefreshToken,
		"tokenType":      "Bearer",
		"expiresIn":      3600,
	})
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input LogoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Revoke refresh token
	result, err := config.DB.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token = $1`, input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check logout status"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Profile management endpoints
func GetProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetString("user_id")

	var user models.User
	err := config.DB.QueryRowContext(ctx, `
        SELECT id, email, username, first_name, last_name, phone, date_of_birth, gender, is_active, last_login, created_at
        FROM users WHERE id = $1
    `, userID).Scan(&user.ID, &user.Email, &user.Username, &user.FirstName, &user.LastName, &user.Phone, &user.DateOfBirth, &user.Gender, &user.IsActive, &user.LastLogin, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type UpdateProfileInput struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
}

func UpdateProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Validate email format if provided
	if input.Email != nil && *input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*input.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Check if email already exists for another user
		var existingUserID string
		err := config.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1 AND id != $2`, *input.Email, userID).Scan(&existingUserID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
	}

	// Validate phone format if provided
	if input.Phone != nil && *input.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(*input.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone format"})
			return
		}
	}

	// Build dynamic update query
	query := "UPDATE users SET "
	args := []interface{}{}
	argIndex := 1

	if input.FirstName != nil {
		query += "first_name = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.FirstName)
		argIndex++
	}
	if input.LastName != nil {
		query += "last_name = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.LastName)
		argIndex++
	}
	if input.Email != nil {
		query += "email = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Email)
		argIndex++
	}
	if input.Phone != nil {
		query += "phone = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Phone)
		argIndex++
	}
	if input.DateOfBirth != nil {
		query += "date_of_birth = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.DateOfBirth)
		argIndex++
	}
	if input.Gender != nil {
		query += "gender = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Gender)
		argIndex++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Remove trailing comma and add WHERE clause
	query = query[:len(query)-2] + " WHERE id = $" + strconv.Itoa(argIndex)
	args = append(args, userID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func ChangePassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get current password hash
	var currentPasswordHash string
	err := config.DB.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentPasswordHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(input.CurrentPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password
	_, err = config.DB.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(newPasswordHash), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// HashPasswordUtility - Utility endpoint to hash passwords for database seeding
// This should be removed or secured in production
type HashPasswordInput struct {
	Password string `json:"password" binding:"required"`
}

func HashPasswordUtility(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input HashPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Password hashing is deliberately expensive, do it in a goroutine with select to respect context
	errCh := make(chan error, 1)
	hashCh := make(chan []byte, 1)

	go func() {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			errCh <- err
			return
		}
		hashCh <- hash
	}()

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timed out"})
		return
	case err := <-errCh:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password", "details": err.Error()})
		return
	case hashedPassword := <-hashCh:
		c.JSON(http.StatusOK, gin.H{
			"password": input.Password,
			"hash":     string(hashedPassword),
			"note":     "Use this hash value in your database password_hash column",
		})
	}
}
