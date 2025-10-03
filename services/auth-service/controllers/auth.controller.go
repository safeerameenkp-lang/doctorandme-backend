package controllers

import (
    "auth-service/config"
    "auth-service/models"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "regexp"
    "strconv"
    "time"
    
    "shared-security"
)

// HealthCheck endpoint
func HealthCheck(c *gin.Context) {
    // Test database connection
    err := config.DB.Ping()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error": "Database connection failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "service": "auth-service",
        "timestamp": time.Now().Unix(),
    })
}

type RegisterInput struct {
    FirstName string `json:"first_name" binding:"required,min=2,max=50"`
    LastName  string `json:"last_name" binding:"required,min=2,max=50"`
    Email     string `json:"email" binding:"omitempty,email"`
    Username  string `json:"username" binding:"required,min=3,max=30"`
    Phone     string `json:"phone" binding:"omitempty,len=10"`
    Password  string `json:"password" binding:"required,min=8"`
}

func Register(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Validate email format if provided
    if input.Email != "" {
        emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
        if !emailRegex.MatchString(input.Email) {
            security.SendValidationError(c, "Invalid email format", "Please provide a valid email address")
            return
        }
    }

    // Validate phone format if provided
    if input.Phone != "" {
        phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
        if !phoneRegex.MatchString(input.Phone) {
            security.SendValidationError(c, "Invalid phone format", "Please provide a valid phone number")
            return
        }
    }

    // Check if username already exists
    var existingUser models.User
    err := config.DB.QueryRow(`SELECT username FROM users WHERE username = $1`, input.Username).Scan(&existingUser.Username)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        return
    }

    // Check if email already exists (if provided)
    if input.Email != "" {
        var existingUser models.User
        err = config.DB.QueryRow(`SELECT email FROM users WHERE email = $1`, input.Email).Scan(&existingUser.Email)
        if err == nil {
            c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
            return
        }
    }

    passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    var userID string
    err = config.DB.QueryRow(`
        INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
        VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
    `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, string(passHash)).Scan(&userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    // Assign default patient role
    var roleID string
    err = config.DB.QueryRow(`SELECT id FROM roles WHERE name='patient' LIMIT 1`).Scan(&roleID)
    if err == nil && roleID != "" {
        _, err = config.DB.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2)`, userID, roleID)
        if err != nil {
            // Log error but don't fail registration
            c.Header("X-Warning", "User created but role assignment failed")
        }
    }

    accessToken, err := security.SignAccessToken(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
        return
    }

    refreshToken, err := security.SignRefreshToken(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
        return
    }

    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    _, err = config.DB.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, userID, refreshToken, expiresAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
        return
    }

    // Create user response using model
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
    Login    string `json:"login" binding:"required"`    // email or phone
    Password string `json:"password" binding:"required"`
}
func Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Fetch user
    var user models.User
    err := config.DB.QueryRow(`
        SELECT id, password_hash, first_name, last_name, email, username, phone
        FROM users
        WHERE (email = $1 OR phone = $1) AND is_active = true
    `, input.Login).Scan(&user.ID, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Phone)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Update last login
    _, err = config.DB.Exec(`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1`, user.ID)
    if err != nil {
        // Log error but don't fail login
        c.Header("X-Warning", "Failed to update last login timestamp")
    }

    // Generate tokens
    accessToken, err := security.SignAccessToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
        return
    }

    refreshToken, err := security.SignRefreshToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
        return
    }

    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    _, err = config.DB.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, user.ID, refreshToken, expiresAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
        return
    }

    // Fetch user roles with organization/clinic context
    rows, err := config.DB.Query(`
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

    var roles []map[string]interface{}
    for rows.Next() {
        var roleID, roleName string
        var permissionsJSON []byte
        var orgID, clinicID, serviceID *string
        
        err = rows.Scan(&roleID, &roleName, &permissionsJSON, &orgID, &clinicID, &serviceID)
        if err != nil {
            continue // Skip invalid role data
        }

        var permissions map[string]interface{}
        err = json.Unmarshal(permissionsJSON, &permissions)
        if err != nil {
            permissions = make(map[string]interface{}) // Default to empty permissions
        }

        role := map[string]interface{}{
            "id":          roleID,
            "name":        roleName,
            "permissions": permissions,
        }
        
        // Add context information if available
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

    c.JSON(http.StatusOK, gin.H{
        "id":           user.ID,
        "firstName":    user.FirstName,
        "lastName":     user.LastName,
        "email":        user.Email,
        "username":     user.Username,
        "phone":        user.Phone,
        "roles":        roles,
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
        "tokenType":    "Bearer",
        "expiresIn":    3600, // 1 hour for access token
    })
}

type RefreshInput struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}
func Refresh(c *gin.Context) {
    var input RefreshInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    token, err := security.VerifyRefreshToken(input.RefreshToken)
    if err != nil || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
        return
    }

    claims := token.Claims.(jwt.MapClaims)
    userID := claims["sub"].(string)

    var refreshToken models.RefreshToken
    err = config.DB.QueryRow(`
        SELECT id FROM refresh_tokens 
        WHERE user_id = $1 AND token = $2 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL
    `, userID, input.RefreshToken).Scan(&refreshToken.ID)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
        return
    }

    // Revoke old token
    _, err = config.DB.Exec(`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`, refreshToken.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke old token"})
        return
    }

    // Generate new tokens
    newAccessToken, err := security.SignAccessToken(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
        return
    }

    newRefreshToken, err := security.SignRefreshToken(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
        return
    }

    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    _, err = config.DB.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, userID, newRefreshToken, expiresAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
        return
    }

    // Fetch user details
    var user models.User
    err = config.DB.QueryRow(`
        SELECT id, first_name, last_name, email, username, phone
        FROM users
        WHERE id = $1 AND is_active = true
    `, userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Phone)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Fetch roles for refreshed user with organization/clinic context
    rows, err := config.DB.Query(`
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

    var roles []map[string]interface{}
    for rows.Next() {
        var roleID, roleName string
        var permissionsJSON []byte
        var orgID, clinicID, serviceID *string
        
        err = rows.Scan(&roleID, &roleName, &permissionsJSON, &orgID, &clinicID, &serviceID)
        if err != nil {
            continue // Skip invalid role data
        }

        var permissions map[string]interface{}
        err = json.Unmarshal(permissionsJSON, &permissions)
        if err != nil {
            permissions = make(map[string]interface{}) // Default to empty permissions
        }

        role := map[string]interface{}{
            "id":          roleID,
            "name":        roleName,
            "permissions": permissions,
        }
        
        // Add context information if available
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

    c.JSON(http.StatusOK, gin.H{
        "id":           user.ID,
        "firstName":    user.FirstName,
        "lastName":     user.LastName,
        "email":        user.Email,
        "username":     user.Username,
        "phone":        user.Phone,
        "roles":        roles,
        "accessToken":  newAccessToken,
        "refreshToken": newRefreshToken,
        "tokenType":    "Bearer",
        "expiresIn":    3600, // 1 hour for access token
    })
}


type LogoutInput struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

func Logout(c *gin.Context) {
    var input LogoutInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
        return
    }

    // Revoke refresh token
    result, err := config.DB.Exec(`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token = $1`, input.RefreshToken)
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
    userID := c.GetString("user_id")
    
    var user models.User
    err := config.DB.QueryRow(`
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
        err := config.DB.QueryRow(`SELECT id FROM users WHERE email = $1 AND id != $2`, *input.Email, userID).Scan(&existingUserID)
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
    
    result, err := config.DB.Exec(query, args...)
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
    userID := c.GetString("user_id")
    var input ChangePasswordInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
        return
    }
    
    // Get current password hash
    var currentPasswordHash string
    err := config.DB.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentPasswordHash)
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
    _, err = config.DB.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, string(newPasswordHash), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}