# Super Admin API Documentation

## Overview
This document describes the comprehensive User Roles and User Management APIs for the Super Admin. These APIs allow the Super Admin (SaaS owner) to manage all users, roles, and permissions in the multi-tenant platform.

## Authentication
All Super Admin APIs require authentication with a valid JWT access token and the `super_admin` role.

### Headers
```
Authorization: Bearer <access_token>
```

## Base URL
```
http://localhost:8000/api/v1/auth/admin
```

---

## User Management APIs

### 1. List All Users
Retrieve a paginated list of all users with filtering options.

**Endpoint:** `GET /admin/users`

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in first name, last name, username, or email
- `role` (string, optional): Filter by role name
- `is_active` (boolean, optional): Filter by active status
- `is_blocked` (boolean, optional): Filter by blocked status
- `sort_by` (string, optional): Field to sort by (created_at, updated_at, first_name, last_name, email, username, last_login)
- `sort_order` (string, optional): Sort order (ASC or DESC, default: DESC)

**Response:**
```json
{
  "users": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "username": "username",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "+1234567890",
      "date_of_birth": "1990-01-01T00:00:00Z",
      "gender": "male",
      "is_active": true,
      "is_blocked": false,
      "blocked_at": null,
      "blocked_reason": null,
      "last_login": "2024-01-01T10:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z",
      "roles": [
        {
          "id": "role-uuid",
          "name": "doctor",
          "description": "Doctor role",
          "permissions": {},
          "is_active": true
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 150,
    "total_pages": 8
  }
}
```

---

### 2. Get Single User
Retrieve detailed information about a specific user.

**Endpoint:** `GET /admin/users/:id`

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "username",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "gender": "male",
  "is_active": true,
  "is_blocked": false,
  "blocked_at": null,
  "blocked_reason": null,
  "last_login": "2024-01-01T10:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z",
  "roles": []
}
```

---

### 3. Create New User
Create a new user account.

**Endpoint:** `POST /admin/users`

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "phone": "+1234567890",
  "password": "SecurePassword123",
  "date_of_birth": "1990-01-01",
  "gender": "male",
  "is_active": true,
  "role_ids": ["role-uuid-1", "role-uuid-2"]
}
```

**Validation Rules:**
- `first_name`: Required, 2-50 characters
- `last_name`: Required, 2-50 characters
- `username`: Required, 3-30 characters, unique
- `email`: Optional, must be valid email format, unique if provided
- `phone`: Optional, must be valid phone format
- `password`: Required, minimum 8 characters
- `date_of_birth`: Optional, ISO date format
- `gender`: Optional
- `is_active`: Optional, default: true
- `role_ids`: Optional, array of role UUIDs

**Response:**
```json
{
  "message": "User created successfully",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "is_active": true,
    "is_blocked": false,
    "roles": []
  }
}
```

---

### 4. Update User
Update user information.

**Endpoint:** `PUT /admin/users/:id`

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "newmail@example.com",
  "phone": "+9876543210",
  "date_of_birth": "1992-05-15",
  "gender": "female"
}
```

**Note:** All fields are optional. Only provided fields will be updated.

**Response:**
```json
{
  "message": "User updated successfully"
}
```

---

### 5. Delete User (Soft Delete)
Mark a user as inactive and blocked.

**Endpoint:** `DELETE /admin/users/:id`

**Response:**
```json
{
  "message": "User deleted successfully"
}
```

**Note:** This is a soft delete. The user is marked as inactive and blocked with reason "Account deleted by administrator". All active refresh tokens are revoked.

---

### 6. Block User
Block a user account with a reason.

**Endpoint:** `POST /admin/users/:id/block`

**Request Body:**
```json
{
  "reason": "Suspicious activity detected on the account"
}
```

**Validation Rules:**
- `reason`: Required, 5-500 characters

**Response:**
```json
{
  "message": "User blocked successfully"
}
```

**Note:** All active refresh tokens are revoked when a user is blocked.

---

### 7. Unblock User
Unblock a previously blocked user account.

**Endpoint:** `POST /admin/users/:id/unblock`

**Response:**
```json
{
  "message": "User unblocked successfully"
}
```

---

### 8. Activate User
Activate a user account.

**Endpoint:** `POST /admin/users/:id/activate`

**Response:**
```json
{
  "message": "User activated successfully"
}
```

---

### 9. Deactivate User
Deactivate a user account.

**Endpoint:** `POST /admin/users/:id/deactivate`

**Response:**
```json
{
  "message": "User deactivated successfully"
}
```

**Note:** All active refresh tokens are revoked when a user is deactivated.

---

### 10. Admin Change Password
Change a user's password (admin override).

**Endpoint:** `POST /admin/users/:id/change-password`

**Request Body:**
```json
{
  "new_password": "NewSecurePassword123"
}
```

**Validation Rules:**
- `new_password`: Required, minimum 8 characters

**Response:**
```json
{
  "message": "Password changed successfully. User must login again."
}
```

**Note:** All active refresh tokens are revoked for security after password change.

---

### 11. Assign Role to User
Assign a role to a user with optional context (organization, clinic, service).

**Endpoint:** `POST /admin/users/:id/roles`

**Request Body:**
```json
{
  "role_id": "role-uuid",
  "organization_id": "org-uuid",
  "clinic_id": "clinic-uuid",
  "service_id": "service-uuid"
}
```

**Validation Rules:**
- `role_id`: Required, must be a valid and active role UUID
- `organization_id`: Optional
- `clinic_id`: Optional
- `service_id`: Optional

**Response:**
```json
{
  "message": "Role assigned successfully"
}
```

---

### 12. Remove Role from User
Remove a role assignment from a user.

**Endpoint:** `DELETE /admin/users/:id/roles/:role_id`

**Response:**
```json
{
  "message": "Role removed successfully"
}
```

---

### 13. Get User Activity Logs
Retrieve activity logs for a specific user.

**Endpoint:** `GET /admin/users/:id/activity-logs`

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 50, max: 100)

**Response:**
```json
{
  "logs": [
    {
      "id": "log-uuid",
      "performed_by": "admin-uuid",
      "performed_by_name": "Admin User",
      "action_type": "BLOCK_USER",
      "action_description": "Blocked user uuid: Suspicious activity",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "metadata": {
        "endpoint": "/admin/users/uuid/block",
        "method": "POST"
      },
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_count": 250,
    "total_pages": 5
  }
}
```

---

## Role Management APIs

### 14. List All Roles
Retrieve a paginated list of all roles with filtering options.

**Endpoint:** `GET /admin/roles`

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in role name or description
- `is_active` (boolean, optional): Filter by active status
- `is_system_role` (boolean, optional): Filter by system role status
- `sort_by` (string, optional): Field to sort by (name, created_at, updated_at)
- `sort_order` (string, optional): Sort order (ASC or DESC, default: ASC)

**Response:**
```json
{
  "roles": [
    {
      "id": "role-uuid",
      "name": "doctor",
      "description": "Doctor role with patient management permissions",
      "permissions": {
        "patients": ["read", "update"],
        "appointments": ["read", "create", "update"],
        "prescriptions": ["read", "create", "update"]
      },
      "is_system_role": true,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z",
      "users_count": 45
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 9,
    "total_pages": 1
  }
}
```

---

### 15. Get Single Role
Retrieve detailed information about a specific role.

**Endpoint:** `GET /admin/roles/:id`

**Response:**
```json
{
  "id": "role-uuid",
  "name": "doctor",
  "description": "Doctor role with patient management permissions",
  "permissions": {
    "patients": ["read", "update"],
    "appointments": ["read", "create", "update"],
    "prescriptions": ["read", "create", "update"]
  },
  "is_system_role": true,
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z",
  "users_count": 45
}
```

---

### 16. Create New Role
Create a new custom role.

**Endpoint:** `POST /admin/roles`

**Request Body:**
```json
{
  "name": "Custom Manager",
  "description": "Custom manager role with specific permissions",
  "permissions": {
    "users": ["read", "update"],
    "reports": ["read", "create"],
    "dashboard": ["read"]
  }
}
```

**Validation Rules:**
- `name`: Required, 3-50 characters, will be converted to lowercase with underscores
- `description`: Optional
- `permissions`: Required, JSON object with permission definitions

**Response:**
```json
{
  "message": "Role created successfully",
  "role_id": "new-role-uuid",
  "name": "custom_manager"
}
```

---

### 17. Update Role
Update role information (custom roles only).

**Endpoint:** `PUT /admin/roles/:id`

**Request Body:**
```json
{
  "name": "Updated Role Name",
  "description": "Updated description",
  "permissions": {
    "users": ["read", "update", "delete"],
    "reports": ["read", "create", "update"]
  }
}
```

**Note:** 
- All fields are optional. Only provided fields will be updated.
- System roles cannot be modified.

**Response:**
```json
{
  "message": "Role updated successfully"
}
```

---

### 18. Delete Role (Soft Delete)
Mark a role as inactive (custom roles only).

**Endpoint:** `DELETE /admin/roles/:id`

**Response:**
```json
{
  "message": "Role deleted successfully"
}
```

**Note:** 
- System roles cannot be deleted.
- Roles with assigned users cannot be deleted.

---

### 19. Activate Role
Activate a previously deactivated role.

**Endpoint:** `POST /admin/roles/:id/activate`

**Response:**
```json
{
  "message": "Role activated successfully"
}
```

---

### 20. Deactivate Role
Deactivate a role (custom roles only).

**Endpoint:** `POST /admin/roles/:id/deactivate`

**Response:**
```json
{
  "message": "Role deactivated successfully"
}
```

**Note:** System roles cannot be deactivated.

---

### 21. Update Role Permissions
Update only the permissions of a role (custom roles only).

**Endpoint:** `PUT /admin/roles/:id/permissions`

**Request Body:**
```json
{
  "permissions": {
    "users": ["read", "create", "update", "delete"],
    "roles": ["read"],
    "organizations": ["read", "update"],
    "clinics": ["read", "create", "update", "delete"]
  }
}
```

**Validation Rules:**
- `permissions`: Required, JSON object with permission definitions

**Response:**
```json
{
  "message": "Role permissions updated successfully"
}
```

**Note:** System roles' permissions cannot be modified.

---

### 22. Get Role Users
Retrieve all users assigned to a specific role.

**Endpoint:** `GET /admin/roles/:id/users`

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)

**Response:**
```json
{
  "users": [
    {
      "id": "user-uuid",
      "email": "user@example.com",
      "username": "username",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "+1234567890",
      "is_active": true,
      "is_blocked": false,
      "last_login": "2024-01-01T10:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "assigned_at": "2024-01-02T00:00:00Z",
      "organization_id": "org-uuid",
      "clinic_id": "clinic-uuid",
      "service_id": null
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 45,
    "total_pages": 3
  }
}
```

---

### 23. Get Permission Templates
Retrieve common permission templates for creating new roles.

**Endpoint:** `GET /admin/permission-templates`

**Response:**
```json
{
  "templates": [
    {
      "name": "Admin Role",
      "description": "Full access to manage users, roles, and system settings",
      "permissions": {
        "users": ["create", "read", "update", "delete"],
        "roles": ["create", "read", "update", "delete"],
        "organizations": ["create", "read", "update", "delete"],
        "clinics": ["create", "read", "update", "delete"],
        "services": ["create", "read", "update", "delete"]
      }
    },
    {
      "name": "Manager Role",
      "description": "Can view and manage resources within assigned scope",
      "permissions": {
        "users": ["create", "read", "update"],
        "roles": ["read"],
        "clinics": ["read", "update"],
        "staff": ["create", "read", "update"]
      }
    },
    {
      "name": "Staff Role",
      "description": "Can view and manage daily operations",
      "permissions": {
        "patients": ["read", "create", "update"],
        "appointments": ["read", "create", "update"],
        "billing": ["read", "create"]
      }
    },
    {
      "name": "Viewer Role",
      "description": "Read-only access to resources",
      "permissions": {
        "users": ["read"],
        "roles": ["read"],
        "patients": ["read"],
        "appointments": ["read"],
        "reports": ["read"]
      }
    }
  ]
}
```

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Validation failed",
  "message": "Invalid input data",
  "code": "VALIDATION_ERROR",
  "details": "field validation error details"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid or expired token",
  "message": "The provided token is invalid, expired, or malformed. Please login again to get a new token",
  "code": "INVALID_TOKEN"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions",
  "message": "Access denied. This resource requires super_admin role",
  "code": "INSUFFICIENT_PERMISSIONS",
  "details": {
    "required_roles": ["super_admin"]
  }
}
```

### 404 Not Found
```json
{
  "error": "Resource not found",
  "message": "The requested User was not found",
  "code": "RESOURCE_NOT_FOUND"
}
```

### 409 Conflict
```json
{
  "error": "Username already exists"
}
```

### 500 Internal Server Error
```json
{
  "error": "Database error",
  "message": "Failed to fetch users",
  "code": "DATABASE_ERROR"
}
```

---

## Security Features

### 1. Audit Trail
All administrative actions are logged in the `user_activity_logs` table with:
- User who performed the action
- Action type and description
- IP address and user agent
- Timestamp
- Additional metadata

### 2. Soft Delete
Users and roles are never hard deleted. They are marked as inactive to maintain data integrity and audit trails.

### 3. Token Revocation
When security-sensitive actions are performed (block, deactivate, password change), all active refresh tokens are automatically revoked to force re-authentication.

### 4. System Role Protection
System roles (predefined roles like super_admin, doctor, patient, etc.) cannot be:
- Modified
- Deleted
- Deactivated

### 5. Self-Action Prevention
Super admins cannot:
- Delete their own account
- Block their own account
- Deactivate their own account

### 6. Password Security
- Passwords are hashed using bcrypt with default cost
- Minimum password length: 8 characters
- Passwords are never returned in API responses

---

## Permission Structure

Permissions are stored as JSON objects with resources as keys and arrays of actions as values:

```json
{
  "resource_name": ["action1", "action2", "action3"]
}
```

### Common Resources:
- `users`: User management
- `roles`: Role management
- `organizations`: Organization management
- `clinics`: Clinic management
- `services`: External service management
- `patients`: Patient management
- `appointments`: Appointment management
- `prescriptions`: Prescription management
- `lab_orders`: Lab order management
- `lab_results`: Lab result management
- `billing`: Billing management
- `payments`: Payment management
- `reports`: Report management
- `dashboard`: Dashboard access

### Common Actions:
- `create`: Create new resource
- `read`: View resource
- `update`: Modify resource
- `delete`: Remove resource

---

## Best Practices

1. **Always use HTTPS** in production environments
2. **Implement rate limiting** on authentication endpoints
3. **Monitor activity logs** for suspicious patterns
4. **Regularly rotate JWT secrets**
5. **Use strong password policies**
6. **Implement IP whitelisting** for super admin access if possible
7. **Enable two-factor authentication** for super admin accounts (future enhancement)
8. **Regular security audits** of user permissions
9. **Backup user data** before performing bulk operations
10. **Test permission changes** in a staging environment first

---

## Multi-Tenancy Support

The system supports multi-tenancy through the `user_roles` table context fields:
- `organization_id`: Links role to specific organization
- `clinic_id`: Links role to specific clinic
- `service_id`: Links role to specific external service

This allows users to have different roles in different contexts within the same platform.

---

## Future Enhancements

1. Two-factor authentication (2FA) for super admin
2. IP whitelisting for admin access
3. Bulk user operations (import/export)
4. Advanced audit log filtering and search
5. Role hierarchy and inheritance
6. Custom permission builder UI
7. Scheduled user account reviews
8. Automated compliance reporting
9. Password complexity requirements configuration
10. Session management and device tracking

