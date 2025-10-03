package security

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
    Error   string      `json:"error"`
    Message string      `json:"message"`
    Code    string      `json:"code"`
    Details interface{} `json:"details,omitempty"`
}

// Common error codes
const (
    // Authentication errors
    CodeMissingToken           = "MISSING_TOKEN"
    CodeInvalidToken          = "INVALID_TOKEN"
    CodeInvalidTokenFormat    = "INVALID_TOKEN_FORMAT"
    CodeInvalidUserInfo       = "INVALID_USER_INFO"
    CodeUserNotFoundOrInactive = "USER_NOT_FOUND_OR_INACTIVE"
    CodeAuthVerificationError  = "AUTH_VERIFICATION_ERROR"
    CodeUserNotAuthenticated  = "USER_NOT_AUTHENTICATED"

    // Authorization errors
    CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
    CodePermissionCheckError     = "PERMISSION_CHECK_ERROR"

    // Validation errors
    CodeValidationError = "VALIDATION_ERROR"

    // Resource errors
    CodeResourceNotFound = "RESOURCE_NOT_FOUND"

    // Server errors
    CodeDatabaseError = "DATABASE_ERROR"
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
