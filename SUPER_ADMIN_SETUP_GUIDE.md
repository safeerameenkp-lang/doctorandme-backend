# Super Admin Setup Guide

## Overview
This guide will walk you through setting up and using the Super Admin User Roles and User Management system in the Dr&Me backend.

## Features Implemented

### User Management
- ✅ List all users with advanced filtering and pagination
- ✅ Create new users with role assignments
- ✅ View detailed user information
- ✅ Update user profile information
- ✅ Delete users (soft delete)
- ✅ Block/Unblock users with reason tracking
- ✅ Activate/Deactivate users
- ✅ Change user passwords (admin override)
- ✅ Assign/Remove roles to/from users
- ✅ View user activity logs

### Role Management
- ✅ List all roles with filtering
- ✅ Create custom roles with permissions
- ✅ View role details
- ✅ Update role information
- ✅ Delete custom roles (soft delete)
- ✅ Activate/Deactivate roles
- ✅ Update role permissions
- ✅ View users assigned to a role
- ✅ Get permission templates

### Security Features
- ✅ Audit trail for all admin actions
- ✅ Super Admin middleware for authorization
- ✅ Token revocation on security actions
- ✅ System role protection
- ✅ Self-action prevention
- ✅ Comprehensive error handling

## Prerequisites

Before you begin, ensure you have:
- Docker and Docker Compose installed
- PostgreSQL database (included in docker-compose.yml)
- Go 1.21 or higher (for local development)
- PowerShell (Windows) or Bash (Linux/Mac)

## Installation Steps

### Step 1: Apply Database Migrations

First, you need to apply the new migration that adds user management features:

```bash
# On Linux/Mac
cd migrations
psql -U postgres -d drandme_db -f 005_user_management_features.sql

# Or using Docker
docker exec -i drandme-postgres psql -U postgres -d drandme_db < migrations/005_user_management_features.sql
```

**On Windows PowerShell:**
```powershell
# Using Docker
Get-Content migrations/005_user_management_features.sql | docker exec -i drandme-postgres psql -U postgres -d drandme_db
```

### Step 2: Create a Super Admin User

You need to create your first super admin user manually in the database:

```sql
-- Connect to your database
psql -U postgres -d drandme_db

-- Create super admin user
INSERT INTO users (first_name, last_name, username, email, password_hash, is_active)
VALUES (
    'Super',
    'Admin',
    'superadmin',
    'superadmin@drandme.com',
    '$2a$10$YourBcryptHashedPasswordHere', -- Use bcrypt to hash your password
    true
);

-- Get the super_admin role ID
SELECT id FROM roles WHERE name = 'super_admin';

-- Assign super_admin role (replace USER_ID and ROLE_ID with actual values)
INSERT INTO user_roles (user_id, role_id, is_active)
VALUES ('USER_ID', 'ROLE_ID', true);
```

**To generate a bcrypt hash for your password:**

```bash
# Using Go
echo 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { hash, _ := bcrypt.GenerateFromPassword([]byte("YourPassword"), 10); fmt.Println(string(hash)) }' > hash.go && go run hash.go && rm hash.go

# Using Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'YourPassword', bcrypt.gensalt()).decode())"

# Using Node.js
node -e "const bcrypt = require('bcrypt'); bcrypt.hash('YourPassword', 10, (err, hash) => console.log(hash));"
```

### Step 3: Rebuild and Start Services

```bash
# Stop existing services
docker-compose down

# Rebuild with new code
docker-compose build auth-service

# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f auth-service
```

### Step 4: Verify Installation

Run the test script to verify everything is working:

**On Windows:**
```powershell
.\scripts\test-super-admin-apis.ps1
```

**On Linux/Mac:**
```bash
chmod +x scripts/test-super-admin-apis.sh
./scripts/test-super-admin-apis.sh
```

## Configuration

### Environment Variables

Make sure these environment variables are set in your `.env` or `docker-compose.yml`:

```env
# JWT Configuration
JWT_ACCESS_SECRET=your-access-secret-key-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-key-min-32-chars

# Database Configuration
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secure-password
POSTGRES_DB=drandme_db

# Service Configuration
AUTH_SERVICE_PORT=8000
```

## API Endpoints

All Super Admin endpoints are prefixed with `/api/v1/auth/admin` and require:
1. Valid JWT access token
2. Super Admin role

### User Management Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/users` | List all users |
| GET | `/admin/users/:id` | Get single user |
| POST | `/admin/users` | Create new user |
| PUT | `/admin/users/:id` | Update user |
| DELETE | `/admin/users/:id` | Delete user |
| POST | `/admin/users/:id/block` | Block user |
| POST | `/admin/users/:id/unblock` | Unblock user |
| POST | `/admin/users/:id/activate` | Activate user |
| POST | `/admin/users/:id/deactivate` | Deactivate user |
| POST | `/admin/users/:id/change-password` | Change user password |
| POST | `/admin/users/:id/roles` | Assign role to user |
| DELETE | `/admin/users/:id/roles/:role_id` | Remove role from user |
| GET | `/admin/users/:id/activity-logs` | Get user activity logs |

### Role Management Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/roles` | List all roles |
| GET | `/admin/roles/:id` | Get single role |
| POST | `/admin/roles` | Create new role |
| PUT | `/admin/roles/:id` | Update role |
| DELETE | `/admin/roles/:id` | Delete role |
| POST | `/admin/roles/:id/activate` | Activate role |
| POST | `/admin/roles/:id/deactivate` | Deactivate role |
| PUT | `/admin/roles/:id/permissions` | Update role permissions |
| GET | `/admin/roles/:id/users` | Get users with role |
| GET | `/admin/permission-templates` | Get permission templates |

## Usage Examples

### Example 1: Login as Super Admin

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "superadmin",
    "password": "YourPassword"
  }'
```

**Response:**
```json
{
  "id": "user-uuid",
  "firstName": "Super",
  "lastName": "Admin",
  "username": "superadmin",
  "email": "superadmin@drandme.com",
  "roles": [
    {
      "id": "role-uuid",
      "name": "super_admin",
      "permissions": { ... }
    }
  ],
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600
}
```

### Example 2: List All Users

```bash
curl -X GET "http://localhost:8000/api/v1/auth/admin/users?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Example 3: Create a New User

```bash
curl -X POST http://localhost:8000/api/v1/auth/admin/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "password": "SecurePassword123",
    "is_active": true,
    "role_ids": ["role-uuid"]
  }'
```

### Example 4: Block a User

```bash
curl -X POST http://localhost:8000/api/v1/auth/admin/users/USER_ID/block \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Suspicious activity detected"
  }'
```

### Example 5: Create a Custom Role

```bash
curl -X POST http://localhost:8000/api/v1/auth/admin/roles \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Custom Manager",
    "description": "Custom manager with specific permissions",
    "permissions": {
      "users": ["read", "update"],
      "reports": ["read", "create"],
      "dashboard": ["read"]
    }
  }'
```

### Example 6: Assign Role to User

```bash
curl -X POST http://localhost:8000/api/v1/auth/admin/users/USER_ID/roles \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "ROLE_ID",
    "organization_id": "ORG_ID",
    "clinic_id": "CLINIC_ID"
  }'
```

## Testing

### Automated Testing

Run the comprehensive test suite:

**Windows:**
```powershell
# Set your credentials if different
$env:SUPER_ADMIN_USERNAME = "superadmin"
$env:SUPER_ADMIN_PASSWORD = "YourPassword"

# Run tests
.\scripts\test-super-admin-apis.ps1
```

**Linux/Mac:**
```bash
# Set your credentials if different
export SUPER_ADMIN_USERNAME="superadmin"
export SUPER_ADMIN_PASSWORD="YourPassword"

# Run tests
./scripts/test-super-admin-apis.sh
```

### Manual Testing with Postman

1. Import the API endpoints into Postman
2. Create an environment with variables:
   - `base_url`: `http://localhost:8000/api/v1/auth`
   - `access_token`: (will be set after login)
3. Login and copy the access token
4. Use the token in the Authorization header for admin endpoints

## Database Schema

The new migration adds the following tables and fields:

### Updated Tables

**users table - new fields:**
- `is_blocked` (BOOLEAN): Whether user is blocked
- `blocked_at` (TIMESTAMP): When user was blocked
- `blocked_by` (UUID): Who blocked the user
- `blocked_reason` (TEXT): Reason for blocking
- `updated_at` (TIMESTAMP): Last update timestamp
- `updated_by` (UUID): Who last updated the user
- `created_by` (UUID): Who created the user

**roles table - new fields:**
- `description` (TEXT): Role description
- `is_system_role` (BOOLEAN): Whether it's a system role
- `is_active` (BOOLEAN): Whether role is active
- `updated_at` (TIMESTAMP): Last update timestamp
- `updated_by` (UUID): Who last updated the role
- `created_by` (UUID): Who created the role

### New Tables

**user_activity_logs:**
- Tracks all admin actions on users
- Includes IP address, user agent, and metadata
- Used for audit trails and compliance

**password_reset_tokens:**
- Stores password reset tokens
- Includes expiry and usage tracking
- Future enhancement for password reset functionality

## Security Considerations

### Production Deployment

1. **Use HTTPS Only**
   - Never transmit access tokens over HTTP in production
   - Configure SSL/TLS certificates

2. **Strong Passwords**
   - Enforce minimum password length (8+ characters)
   - Require password complexity
   - Implement password expiry policies

3. **Token Management**
   - Keep JWT secrets secure and rotate regularly
   - Set appropriate token expiration times
   - Implement token refresh mechanism

4. **Rate Limiting**
   - Implement rate limiting on authentication endpoints
   - Prevent brute force attacks
   - Use tools like nginx or API gateway

5. **IP Whitelisting**
   - Consider whitelisting admin IP addresses
   - Implement geo-blocking if applicable
   - Use VPN for remote admin access

6. **Audit Logs**
   - Regularly review activity logs
   - Set up alerts for suspicious activities
   - Archive logs for compliance

7. **Database Security**
   - Use strong database passwords
   - Limit database access to specific IPs
   - Enable database encryption at rest
   - Regular database backups

8. **Environment Variables**
   - Never commit secrets to version control
   - Use secrets management tools (HashiCorp Vault, AWS Secrets Manager)
   - Rotate secrets regularly

## Troubleshooting

### Issue: "Authentication required" error

**Solution:**
- Ensure you're sending the Authorization header: `Bearer YOUR_TOKEN`
- Check if your token hasn't expired (default: 15 minutes)
- Refresh your token if expired

### Issue: "Insufficient permissions" error

**Solution:**
- Verify the user has super_admin role assigned
- Check if the role is active: `SELECT * FROM user_roles WHERE user_id='YOUR_ID'`
- Ensure the role assignment has `is_active = true`

### Issue: Migration fails

**Solution:**
- Check if previous migrations are applied
- Ensure database connection is working
- Check for syntax errors in migration file
- Verify PostgreSQL version compatibility

### Issue: Cannot create super admin

**Solution:**
- Ensure super_admin role exists: `SELECT * FROM roles WHERE name='super_admin'`
- If not, run migration `001_initial_schema.sql` first
- Check password hash is correctly formatted (bcrypt)

### Issue: Test script fails

**Solution:**
- Verify services are running: `docker-compose ps`
- Check service logs: `docker-compose logs auth-service`
- Ensure correct credentials in test script
- Verify port 8000 is not blocked by firewall

## Monitoring and Maintenance

### Regular Tasks

1. **Daily:**
   - Review activity logs for anomalies
   - Check for failed login attempts
   - Monitor API response times

2. **Weekly:**
   - Review blocked users
   - Audit role assignments
   - Check for inactive users

3. **Monthly:**
   - Rotate JWT secrets
   - Review and update permissions
   - Archive old activity logs
   - Update documentation

4. **Quarterly:**
   - Security audit
   - Performance optimization
   - Backup verification
   - Disaster recovery test

## Migration from Existing System

If you have existing users in your system:

1. **Backup existing data:**
```bash
docker exec drandme-postgres pg_dump -U postgres drandme_db > backup_before_migration.sql
```

2. **Apply new migration:**
```bash
docker exec -i drandme-postgres psql -U postgres -d drandme_db < migrations/005_user_management_features.sql
```

3. **Verify data integrity:**
```sql
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM roles;
SELECT COUNT(*) FROM user_roles;
```

4. **Test functionality:**
- Run the test script
- Manually test critical endpoints
- Verify existing users can still login

## Support and Documentation

- **Full API Documentation:** See `SUPER_ADMIN_API_DOCUMENTATION.md`
- **Architecture Overview:** See main `README.md`
- **Security Best Practices:** See above section

## Future Enhancements

Planned features for future releases:

1. **Two-Factor Authentication (2FA)**
   - TOTP support
   - SMS verification
   - Email verification

2. **Advanced Audit Features**
   - Export audit logs
   - Advanced filtering
   - Compliance reports

3. **Bulk Operations**
   - Bulk user import/export
   - Bulk role assignments
   - CSV import support

4. **Role Hierarchy**
   - Parent-child role relationships
   - Permission inheritance
   - Role templates

5. **Session Management**
   - Active session tracking
   - Device management
   - Force logout all devices

6. **Password Policies**
   - Configurable complexity requirements
   - Password history
   - Expiry policies

## License

This project is part of the Dr&Me healthcare management system.

## Contributors

- Backend Development Team
- Security Team
- DevOps Team

---

**Last Updated:** October 7, 2025
**Version:** 1.0.0

