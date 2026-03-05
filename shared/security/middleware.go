package security

import (
    "database/sql"
    "errors"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
)

// Database interface for dependency injection
type Database interface {
    QueryRow(query string, args ...interface{}) *sql.Row
    Query(query string, args ...interface{}) (*sql.Rows, error)
}

// JWT utilities
func SignAccessToken(userID string) (string, error) {
    secret := os.Getenv("JWT_ACCESS_SECRET")
    if secret == "" {
        return "", errors.New("JWT_ACCESS_SECRET not set")
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  userID,
        "exp":  time.Now().Add(15 * time.Minute).Unix(),
        "iat":  time.Now().Unix(),
        "type": "access",
    })
    return token.SignedString([]byte(secret))
}

func SignRefreshToken(userID string) (string, error) {
    secret := os.Getenv("JWT_REFRESH_SECRET")
    if secret == "" {
        return "", errors.New("JWT_REFRESH_SECRET not set")
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  userID,
        "exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
        "iat":  time.Now().Unix(),
        "type": "refresh",
    })
    return token.SignedString([]byte(secret))
}

func VerifyRefreshToken(tokenStr string) (*jwt.Token, error) {
    secret := os.Getenv("JWT_REFRESH_SECRET")
    if secret == "" {
        return nil, errors.New("JWT_REFRESH_SECRET not set")
    }

    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Check if token is valid and not expired
    if !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    // Check token type
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    
    tokenType, ok := claims["type"].(string)
    if !ok || tokenType != "refresh" {
        return nil, errors.New("invalid token type")
    }
    
    return token, nil
}

// AuthMiddleware creates a Gin middleware for JWT authentication
func AuthMiddleware(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := c.GetHeader("Authorization")
        if tokenStr == "" {
            SendError(c, http.StatusUnauthorized, CodeMissingToken, "Authentication required", 
                "Please provide a valid authorization token in the request header", nil)
            c.Abort()
            return
        }

        // Remove "Bearer " prefix if present
        if strings.HasPrefix(tokenStr, "Bearer ") {
            tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
        }

        secret := os.Getenv("JWT_ACCESS_SECRET")
        if secret == "" {
            SendError(c, http.StatusInternalServerError, CodeAuthVerificationError, "JWT configuration error", 
                "Server configuration error. Please try again later", nil)
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("unexpected signing method")
            }
            return []byte(secret), nil
        })

        if err != nil || !token.Valid {
            SendError(c, http.StatusUnauthorized, CodeInvalidToken, "Invalid or expired token", 
                "The provided token is invalid, expired, or malformed. Please login again to get a new token", nil)
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            SendError(c, http.StatusUnauthorized, CodeInvalidTokenFormat, "Invalid token format", 
                "The token format is invalid. Please login again to get a new token", nil)
            c.Abort()
            return
        }

        userID, ok := claims["sub"].(string)
        if !ok {
            SendError(c, http.StatusUnauthorized, CodeInvalidUserInfo, "Invalid user information", 
                "The token does not contain valid user information. Please login again", nil)
            c.Abort()
            return
        }

        var exists bool
        err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_active=true)`, userID).Scan(&exists)
        if err != nil {
            SendError(c, http.StatusInternalServerError, CodeAuthVerificationError, "Authentication verification failed", 
                "Unable to verify user status. Please try again later", nil)
            c.Abort()
            return
        }
        if !exists {
            SendError(c, http.StatusUnauthorized, CodeUserNotFoundOrInactive, "User account not found or inactive", 
                "Your account is not found or has been deactivated. Please contact support", nil)
            c.Abort()
            return
        }

        c.Set("user_id", userID)
        c.Next()
    }
}

// RequireRole creates a Gin middleware for role-based access control
func RequireRole(db Database, expectedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")

        if userID == "" {
            SendError(c, http.StatusUnauthorized, CodeUserNotAuthenticated, "User not authenticated", 
                "User authentication is required to access this resource", nil)
            c.Abort()
            return
        }

        rows, err := db.Query(`
            SELECT r.name
            FROM roles r
            JOIN user_roles ur ON ur.role_id = r.id
            WHERE ur.user_id = $1 AND ur.is_active = true
        `, userID)

        if err != nil {
            SendError(c, http.StatusInternalServerError, CodePermissionCheckError, "Failed to check user permissions", 
                "Unable to verify user permissions. Please try again later", nil)
            c.Abort()
            return
        }

        defer rows.Close()
        roles := []string{}
        for rows.Next() {
            var role string
            if err := rows.Scan(&role); err != nil {
                continue
            }
            roles = append(roles, role)
        }

        // Check if user has any of the required roles or is super_admin
        for _, userRole := range roles {
            if userRole == "super_admin" {
                c.Next()
                return
            }
            for _, expectedRole := range expectedRoles {
                if userRole == expectedRole {
                    c.Next()
                    return
                }
            }
        }

        // Create a more descriptive error message
        var roleList string
        if len(expectedRoles) == 1 {
            roleList = expectedRoles[0]
        } else if len(expectedRoles) == 2 {
            roleList = expectedRoles[0] + " or " + expectedRoles[1]
        } else {
            roleList = strings.Join(expectedRoles[:len(expectedRoles)-1], ", ") + ", or " + expectedRoles[len(expectedRoles)-1]
        }

        SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "Insufficient permissions", 
            "Access denied. This resource requires " + roleList + " role. Your current roles: " + strings.Join(roles, ", "), 
            gin.H{
                "required_roles": expectedRoles,
                "user_roles": roles,
            })
        c.Abort()
    }
}


// RequireSuperAdmin creates a Gin middleware that ensures the user has super_admin role
func RequireSuperAdmin(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")

        if userID == "" {
            SendError(c, http.StatusUnauthorized, CodeUserNotAuthenticated, "User not authenticated", 
                "User authentication is required to access this resource", nil)
            c.Abort()
            return
        }

        // Check if user has super_admin role
        var isSuperAdmin bool
        err := db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 AND r.name = 'super_admin' AND ur.is_active = true
            )
        `, userID).Scan(&isSuperAdmin)

        if err != nil {
            SendError(c, http.StatusInternalServerError, CodePermissionCheckError, "Failed to check user permissions", 
                "Unable to verify user permissions. Please try again later", nil)
            c.Abort()
            return
        }

        if !isSuperAdmin {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "Insufficient permissions", 
                "Access denied. This resource requires super_admin role", 
                gin.H{
                    "required_roles": []string{"super_admin"},
                })
            c.Abort()
            return
        }

        // ✅ FIX: Set context variables for downstream controllers
        c.Set("is_super_admin", true)
        c.Set("is_organization_admin", false)
        c.Set("is_clinic_admin", false)
        c.Next()
    }
}

// RequireAdminRole creates a Gin middleware that ensures the user has admin-level roles
func RequireAdminRole(db Database) gin.HandlerFunc {
    adminRoles := []string{"super_admin", "organization_admin", "clinic_admin"}
    return RequireRole(db, adminRoles...)
}

// GetUserOrganizationContext retrieves all organizations the user has access to
func GetUserOrganizationContext(db Database, userID string) ([]string, error) {
    rows, err := db.Query(`
        SELECT DISTINCT ur.organization_id
        FROM user_roles ur
        WHERE ur.user_id = $1 
        AND ur.organization_id IS NOT NULL
        AND ur.is_active = true
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orgIDs []string
    for rows.Next() {
        var orgID string
        if err := rows.Scan(&orgID); err != nil {
            continue
        }
        orgIDs = append(orgIDs, orgID)
    }
    return orgIDs, nil
}

// GetUserClinicContext retrieves all clinics the user has access to
func GetUserClinicContext(db Database, userID string) ([]string, error) {
    rows, err := db.Query(`
        SELECT DISTINCT ur.clinic_id
        FROM user_roles ur
        WHERE ur.user_id = $1 
        AND ur.clinic_id IS NOT NULL
        AND ur.is_active = true
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var clinicIDs []string
    for rows.Next() {
        var clinicID string
        if err := rows.Scan(&clinicID); err != nil {
            continue
        }
        clinicIDs = append(clinicIDs, clinicID)
    }
    return clinicIDs, nil
}

// RequireOrganizationAdmin middleware ensures user is admin of the specified organization
func RequireOrganizationAdmin(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            SendError(c, http.StatusUnauthorized, CodeUserNotAuthenticated, "User not authenticated", 
                "User authentication is required to access this resource", nil)
            c.Abort()
            return
        }

        // Check if user is super_admin (they have access to everything)
        var isSuperAdmin bool
        err := db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 AND r.name = 'super_admin' AND ur.is_active = true
            )
        `, userID).Scan(&isSuperAdmin)

        if err != nil {
            SendError(c, http.StatusInternalServerError, CodePermissionCheckError, "Failed to check permissions", 
                "Unable to verify user permissions. Please try again later", nil)
            c.Abort()
            return
        }

        if isSuperAdmin {
            c.Set("is_super_admin", true)
            c.Next()
            return
        }

        // Check if user is organization_admin and get their organizations
        var isOrgAdmin bool
        err = db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 AND r.name = 'organization_admin' AND ur.is_active = true
            )
        `, userID).Scan(&isOrgAdmin)

        if err != nil || !isOrgAdmin {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "Insufficient permissions", 
                "Access denied. This resource requires organization_admin or super_admin role", 
                gin.H{"required_roles": []string{"organization_admin", "super_admin"}})
            c.Abort()
            return
        }

        // Get user's organization context
        orgIDs, err := GetUserOrganizationContext(db, userID)
        if err != nil || len(orgIDs) == 0 {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "No organization access", 
                "You are not assigned to any organization", nil)
            c.Abort()
            return
        }

        c.Set("is_super_admin", false)
        c.Set("is_organization_admin", true)
        c.Set("organization_ids", orgIDs)
        c.Next()
    }
}

// RequireClinicAdmin middleware ensures user is admin of the specified clinic
func RequireClinicAdmin(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            SendError(c, http.StatusUnauthorized, CodeUserNotAuthenticated, "User not authenticated", 
                "User authentication is required to access this resource", nil)
            c.Abort()
            return
        }

        // Check if user is super_admin (they have access to everything)
        var isSuperAdmin bool
        err := db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 AND r.name = 'super_admin' AND ur.is_active = true
            )
        `, userID).Scan(&isSuperAdmin)

        if err != nil {
            SendError(c, http.StatusInternalServerError, CodePermissionCheckError, "Failed to check permissions", 
                "Unable to verify user permissions. Please try again later", nil)
            c.Abort()
            return
        }

        if isSuperAdmin {
            c.Set("is_super_admin", true)
            c.Set("is_organization_admin", false)
            c.Set("is_clinic_admin", false)
            c.Next()
            return
        }

        // Check if user is clinic_admin and get their clinics
        var isClinicAdmin bool
        err = db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 AND r.name = 'clinic_admin' AND ur.is_active = true
            )
        `, userID).Scan(&isClinicAdmin)

        if err != nil || !isClinicAdmin {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "Insufficient permissions", 
                "Access denied. This resource requires clinic_admin or super_admin role", 
                gin.H{"required_roles": []string{"clinic_admin", "super_admin"}})
            c.Abort()
            return
        }

        // Get user's clinic context
        clinicIDs, err := GetUserClinicContext(db, userID)
        if err != nil || len(clinicIDs) == 0 {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "No clinic access", 
                "You are not assigned to any clinic", nil)
            c.Abort()
            return
        }

        c.Set("is_super_admin", false)
        c.Set("is_organization_admin", false)
        c.Set("is_clinic_admin", true)
        c.Set("clinic_ids", clinicIDs)
        c.Next()
    }
}

// RequireAnyAdmin middleware ensures user has any admin role
func RequireAnyAdmin(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            SendError(c, http.StatusUnauthorized, CodeUserNotAuthenticated, "User not authenticated", 
                "User authentication is required to access this resource", nil)
            c.Abort()
            return
        }

        // Check if user has any admin role
        var hasAdminRole bool
        err := db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM roles r
                JOIN user_roles ur ON ur.role_id = r.id
                WHERE ur.user_id = $1 
                AND r.name IN ('super_admin', 'organization_admin', 'clinic_admin')
                AND ur.is_active = true
            )
        `, userID).Scan(&hasAdminRole)

        if err != nil {
            SendError(c, http.StatusInternalServerError, CodePermissionCheckError, "Failed to check permissions", 
                "Unable to verify user permissions. Please try again later", nil)
            c.Abort()
            return
        }

        if !hasAdminRole {
            SendError(c, http.StatusForbidden, CodeInsufficientPermissions, "Insufficient permissions", 
                "Access denied. This resource requires an admin role", 
                gin.H{"required_roles": []string{"super_admin", "organization_admin", "clinic_admin"}})
            c.Abort()
            return
        }

        c.Next()
    }
}

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        log.Printf("CORS Request from Origin: %s", origin)

        allowOrigin := ""

        if origin != "" {
            // Allow any localhost or 127.0.0.1
            if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
                allowOrigin = origin
            } else {
                // Optionally allow other production origins
                allowOrigin = origin // or "*" if you want to allow all
            }
        } else {
            allowOrigin = "*" // no origin header? allow all
        }

        c.Header("Access-Control-Allow-Origin", allowOrigin)
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, X-File-Name")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}