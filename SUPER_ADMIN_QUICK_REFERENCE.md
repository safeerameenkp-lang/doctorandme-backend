# Super Admin Quick Reference

## Quick Start

### 1. Login
```bash
POST /api/v1/auth/login
{
  "login": "superadmin",
  "password": "your-password"
}
```

### 2. Use Token
```bash
Authorization: Bearer YOUR_ACCESS_TOKEN
```

## Common Operations

### User Management

#### List Users
```bash
GET /api/v1/auth/admin/users?page=1&page_size=20&search=john&is_active=true
```

#### Create User
```bash
POST /api/v1/auth/admin/users
{
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "Password123",
  "role_ids": ["role-uuid"]
}
```

#### Block User
```bash
POST /api/v1/auth/admin/users/{id}/block
{
  "reason": "Suspicious activity"
}
```

#### Change Password
```bash
POST /api/v1/auth/admin/users/{id}/change-password
{
  "new_password": "NewPassword123"
}
```

### Role Management

#### List Roles
```bash
GET /api/v1/auth/admin/roles?page=1&page_size=20
```

#### Create Role
```bash
POST /api/v1/auth/admin/roles
{
  "name": "Custom Manager",
  "description": "Custom role",
  "permissions": {
    "users": ["read", "update"],
    "reports": ["read", "create"]
  }
}
```

#### Assign Role to User
```bash
POST /api/v1/auth/admin/users/{user_id}/roles
{
  "role_id": "role-uuid",
  "organization_id": "org-uuid",  // optional
  "clinic_id": "clinic-uuid"       // optional
}
```

#### Update Permissions
```bash
PUT /api/v1/auth/admin/roles/{id}/permissions
{
  "permissions": {
    "users": ["read", "create", "update", "delete"],
    "roles": ["read"]
  }
}
```

## Query Parameters

### List Users
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `search`: Search in name, username, email
- `role`: Filter by role name
- `is_active`: Filter by active status (true/false)
- `is_blocked`: Filter by blocked status (true/false)
- `sort_by`: created_at, updated_at, first_name, last_name, email, username, last_login
- `sort_order`: ASC or DESC

### List Roles
- `page`: Page number
- `page_size`: Items per page
- `search`: Search in name or description
- `is_active`: Filter by active status
- `is_system_role`: Filter by system role status
- `sort_by`: name, created_at, updated_at
- `sort_order`: ASC or DESC

## Permission Structure

```json
{
  "resource": ["action1", "action2"]
}
```

### Common Resources
- `users`: User management
- `roles`: Role management
- `organizations`: Organization management
- `clinics`: Clinic management
- `patients`: Patient management
- `appointments`: Appointment management
- `prescriptions`: Prescription management
- `billing`: Billing operations
- `reports`: Report generation

### Common Actions
- `create`: Create new items
- `read`: View items
- `update`: Modify items
- `delete`: Remove items

## Status Codes

- `200`: Success
- `201`: Created
- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Not Found
- `409`: Conflict
- `500`: Internal Server Error

## Error Response Format

```json
{
  "error": "Error title",
  "message": "Detailed error message",
  "code": "ERROR_CODE",
  "details": { ... }
}
```

## System Roles (Protected)

These roles cannot be modified or deleted:
- `super_admin`
- `organization_admin`
- `clinic_admin`
- `doctor`
- `receptionist`
- `pharmacist`
- `lab_technician`
- `billing_staff`
- `patient`

## Best Practices

1. **Always use HTTPS** in production
2. **Set appropriate token expiration** times
3. **Monitor activity logs** regularly
4. **Use strong passwords** (min 8 chars)
5. **Implement rate limiting** on auth endpoints
6. **Back up data** before bulk operations
7. **Test in staging** before production changes
8. **Document custom roles** and their purposes
9. **Regular security audits** of permissions
10. **Archive old activity logs** periodically

## Testing

### Windows
```powershell
.\scripts\test-super-admin-apis.ps1
```

### Linux/Mac
```bash
./scripts/test-super-admin-apis.sh
```

## Database Queries

### Find User ID
```sql
SELECT id, username, email FROM users WHERE username = 'johndoe';
```

### Check User Roles
```sql
SELECT u.username, r.name as role_name, ur.is_active
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.username = 'johndoe';
```

### View Activity Logs
```sql
SELECT * FROM user_activity_logs 
WHERE user_id = 'USER_UUID' 
ORDER BY created_at DESC 
LIMIT 10;
```

### Count Users by Role
```sql
SELECT r.name, COUNT(ur.user_id) as user_count
FROM roles r
LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.is_active = true
GROUP BY r.name;
```

## Troubleshooting

### Token Expired
- Refresh token or login again
- Default expiry: 15 minutes for access token

### Permission Denied
- Verify super_admin role is assigned
- Check role is active
- Ensure token is valid

### User Creation Fails
- Check username/email uniqueness
- Verify password meets requirements (min 8 chars)
- Ensure valid email format

### Cannot Delete Role
- Check if users are assigned to the role
- System roles cannot be deleted
- Remove users from role first

## Quick Reference Commands

### PowerShell Testing
```powershell
# Login
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/login" `
  -Method POST -ContentType "application/json" `
  -Body '{"login":"superadmin","password":"YourPassword"}'
$token = $response.accessToken

# List users
Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/admin/users" `
  -Method GET -Headers @{Authorization="Bearer $token"}

# Create user
Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/admin/users" `
  -Method POST -Headers @{Authorization="Bearer $token"} `
  -ContentType "application/json" `
  -Body '{"first_name":"Test","last_name":"User","username":"testuser","password":"Password123"}'
```

### Bash/Curl Testing
```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"superadmin","password":"YourPassword"}' \
  | jq -r '.accessToken')

# List users
curl -X GET http://localhost:8000/api/v1/auth/admin/users \
  -H "Authorization: Bearer $TOKEN"

# Create user
curl -X POST http://localhost:8000/api/v1/auth/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"User","username":"testuser","password":"Password123"}'
```

## Support

- **Full API Docs:** `SUPER_ADMIN_API_DOCUMENTATION.md`
- **Setup Guide:** `SUPER_ADMIN_SETUP_GUIDE.md`
- **Main README:** `README.md`

---

**Version:** 1.0.0 | **Last Updated:** October 7, 2025

