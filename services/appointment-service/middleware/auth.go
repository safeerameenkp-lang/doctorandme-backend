package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Database interface for dependency injection
type Database interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// Common error codes
const (
	CodeMissingToken            = "MISSING_TOKEN"
	CodeInvalidToken            = "INVALID_TOKEN"
	CodeInvalidTokenFormat      = "INVALID_TOKEN_FORMAT"
	CodeInvalidUserInfo         = "INVALID_USER_INFO"
	CodeUserNotFoundOrInactive  = "USER_NOT_FOUND_OR_INACTIVE"
	CodeAuthVerificationError   = "AUTH_VERIFICATION_ERROR"
	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	CodePermissionCheckError    = "PERMISSION_CHECK_ERROR"
	CodeValidationError         = "VALIDATION_ERROR"
	CodeResourceNotFound        = "RESOURCE_NOT_FOUND"
	CodeDatabaseError           = "DATABASE_ERROR"
)

// SendError sends a standardized error response
func SendError(c *gin.Context, statusCode int, errorCode, errorMessage, detailedMessage string, details interface{}) {
	response := ErrorResponse{
		Error:   errorMessage,
		Message: detailedMessage,
		Code:    errorCode,
	}

	if details != nil {
		response.Details = details
	}

	c.JSON(statusCode, response)
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
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

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
			SendError(c, http.StatusUnauthorized, CodeMissingToken, "User not authenticated",
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
			"Access denied. This resource requires "+roleList+" role. Your current roles: "+strings.Join(roles, ", "),
			gin.H{
				"required_roles": expectedRoles,
				"user_roles":     roles,
			})
		c.Abort()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowOrigin := ""

		if origin != "" {
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
				allowOrigin = origin
			} else {
				allowOrigin = origin
			}
		} else {
			allowOrigin = "*"
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

// SendValidationError sends a validation error response
func SendValidationError(c *gin.Context, message string, details interface{}) {
	SendError(c, http.StatusBadRequest, CodeValidationError, "Validation failed", message, details)
}

// SendNotFoundError sends a not found error response
func SendNotFoundError(c *gin.Context, resource string) {
	SendError(c, http.StatusNotFound, CodeResourceNotFound, "Resource not found",
		"The requested "+resource+" was not found", nil)
}

// SendDatabaseError sends a database error response
func SendDatabaseError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, CodeDatabaseError, "Database error",
		message, nil)
}
