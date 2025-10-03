# Authentication and Authorization Error Handling

This document describes the comprehensive error handling system implemented across all microservices for authentication and authorization.

## Overview

All APIs in the system now properly handle authentication and authorization errors with consistent, descriptive error messages. When a token is not passed or a user's role is insufficient to access an API, clear error messages are returned.

## Error Response Format

All error responses follow a standardized format:

```json
{
  "error": "Short error description",
  "message": "Detailed error message for the user",
  "code": "ERROR_CODE",
  "details": {} // Optional additional details
}
```

## Authentication Error Codes

### Missing Token
- **Code**: `MISSING_TOKEN`
- **Status**: `401 Unauthorized`
- **Message**: "Please provide a valid authorization token in the request header"
- **Occurs when**: No Authorization header is provided

### Invalid Token
- **Code**: `INVALID_TOKEN`
- **Status**: `401 Unauthorized`
- **Message**: "The provided token is invalid, expired, or malformed. Please login again to get a new token"
- **Occurs when**: Token is expired, malformed, or invalid

### Invalid Token Format
- **Code**: `INVALID_TOKEN_FORMAT`
- **Status**: `401 Unauthorized`
- **Message**: "The token format is invalid. Please login again to get a new token"
- **Occurs when**: Token claims cannot be parsed

### Invalid User Information
- **Code**: `INVALID_USER_INFO`
- **Status**: `401 Unauthorized`
- **Message**: "The token does not contain valid user information. Please login again"
- **Occurs when**: Token doesn't contain valid user ID

### User Not Found or Inactive
- **Code**: `USER_NOT_FOUND_OR_INACTIVE`
- **Status**: `401 Unauthorized`
- **Message**: "Your account is not found or has been deactivated. Please contact support"
- **Occurs when**: User doesn't exist or is inactive

### Authentication Verification Error
- **Code**: `AUTH_VERIFICATION_ERROR`
- **Status**: `500 Internal Server Error`
- **Message**: "Unable to verify user status. Please try again later"
- **Occurs when**: Database error during user verification

## Authorization Error Codes

### User Not Authenticated
- **Code**: `USER_NOT_AUTHENTICATED`
- **Status**: `401 Unauthorized`
- **Message**: "User authentication is required to access this resource"
- **Occurs when**: User ID is not available in context

### Insufficient Permissions
- **Code**: `INSUFFICIENT_PERMISSIONS`
- **Status**: `403 Forbidden`
- **Message**: "Access denied. This resource requires [role] role. Your current roles: [user_roles]"
- **Occurs when**: User doesn't have required role
- **Additional fields**:
  - `required_roles`: Array of required roles
  - `user_roles`: Array of user's current roles

### Permission Check Error
- **Code**: `PERMISSION_CHECK_ERROR`
- **Status**: `500 Internal Server Error`
- **Message**: "Unable to verify user permissions. Please try again later"
- **Occurs when**: Database error during permission check

## Implementation Details

### Shared Security Middleware

The error handling is implemented in the shared security module (`shared/security/`):

1. **AuthMiddleware**: Handles token validation
2. **RequireRole**: Handles role-based access control
3. **Error utilities**: Standardized error response functions

### Service Integration

All services use the shared security middleware:

- **Auth Service**: Uses `AuthMiddleware` for protected endpoints
- **Organization Service**: Uses both `AuthMiddleware` and `RequireRole`
- **Appointment Service**: Uses both `AuthMiddleware` and `RequireRole`

### Error Response Functions

The following utility functions are available for consistent error handling:

```go
// Send standardized error response
security.SendError(c, statusCode, errorCode, errorMessage, detailedMessage, details)

// Send validation error
security.SendValidationError(c, message, details)

// Send not found error
security.SendNotFoundError(c, resource)

// Send conflict error
security.SendConflictError(c, message, details)

// Send internal server error
security.SendInternalError(c, message)

// Send database error
security.SendDatabaseError(c, message)
```

## Testing

A comprehensive test script is provided (`scripts/test-auth-errors.ps1`) that verifies:

1. **Missing Token Scenarios**: Tests all protected endpoints without tokens
2. **Invalid Token Scenarios**: Tests with malformed or invalid tokens
3. **Insufficient Permissions**: Tests role-based access control
4. **Health Checks**: Verifies public endpoints work without authentication
5. **Public Endpoints**: Verifies auth endpoints are accessible

### Running Tests

```powershell
# Run the authentication error tests
.\scripts\test-auth-errors.ps1

# With custom service URLs
.\scripts\test-auth-errors.ps1 -AuthServiceUrl "http://localhost:8081" -OrgServiceUrl "http://localhost:8082" -AppointmentServiceUrl "http://localhost:8083"
```

## Endpoint Protection

### Public Endpoints (No Authentication Required)
- `GET /api/health` - Health checks for all services
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `POST /api/refresh` - Token refresh
- `POST /api/logout` - User logout

### Protected Endpoints (Authentication Required)
All other endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

### Role-Based Endpoints
Many endpoints require specific roles:

- **super_admin**: Full system access
- **organization_admin**: Organization management
- **clinic_admin**: Clinic management
- **doctor**: Doctor-specific operations
- **receptionist**: Reception operations
- **patient**: Patient-specific operations

## Error Handling Best Practices

1. **Always check authentication first**: Use `AuthMiddleware` before any business logic
2. **Use role-based access control**: Apply `RequireRole` for sensitive operations
3. **Provide clear error messages**: Include both error codes and user-friendly messages
4. **Include relevant details**: Add context like required roles and user roles
5. **Handle edge cases**: Account for database errors and malformed requests
6. **Test thoroughly**: Use the provided test script to verify error handling

## Security Considerations

1. **Token Validation**: All tokens are validated for format, expiration, and user status
2. **Role Verification**: User roles are checked against the database for each request
3. **Error Information**: Error messages don't leak sensitive information
4. **Consistent Responses**: All services return errors in the same format
5. **Proper Status Codes**: HTTP status codes accurately reflect the error type

## Troubleshooting

### Common Issues

1. **"MISSING_TOKEN" errors**: Ensure Authorization header is included
2. **"INVALID_TOKEN" errors**: Check token expiration and format
3. **"INSUFFICIENT_PERMISSIONS" errors**: Verify user has required role
4. **Database errors**: Check database connectivity and user table

### Debugging

1. Check the error code in the response
2. Review the detailed message for context
3. Verify user roles in the database
4. Check token validity and expiration
5. Review service logs for additional details

This comprehensive error handling system ensures that all APIs provide clear, consistent feedback when authentication or authorization fails, improving the developer experience and system security.
