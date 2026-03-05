# 🔧 Login Password Fix Guide

## The Problem

You're getting this error when trying to login:
```json
{
  "debug": "Password mismatch",
  "error": "Invalid credentials"
}
```

This happens because **passwords in your database are not properly bcrypt-hashed**. When you manually insert users into the database, you need to hash the passwords first.

---

## Quick Fix Solutions

### Option 1: Use the Password Hash Utility (Recommended)

I've added a utility endpoint to generate password hashes for you.

#### Step 1: Run the PowerShell Script

```powershell
cd scripts
.\fix-password.ps1
```

The script will:
1. Ask for your password
2. Generate a bcrypt hash
3. Provide SQL commands to update your database

#### Step 2: Update the Database

Copy the SQL command from the script output and run it in your database:

```sql
UPDATE users SET password_hash = '$2a$10$...' WHERE username = 'your_username';
```

### Option 2: Use the API Directly

You can also call the API directly:

```powershell
# Using PowerShell
$body = @{ password = "your_password" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/auth/hash-password" -Method POST -Body $body -ContentType "application/json"
```

Or using curl:
```bash
curl -X POST http://localhost:8080/auth/hash-password \
  -H "Content-Type: application/json" \
  -d '{"password":"your_password"}'
```

Response:
```json
{
  "password": "your_password",
  "hash": "$2a$10$K8J9xF...",
  "note": "Use this hash value in your database password_hash column"
}
```

### Option 3: Use the Register API (Best for New Users)

Instead of manually inserting users, use the register endpoint which automatically hashes passwords:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe",
    "email": "john@example.com",
    "phone": "1234567890",
    "password": "your_password"
  }'
```

---

## Verify the Fix

After updating the password hash in your database, try logging in again:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "your_username",
    "password": "your_password"
  }'
```

You should now receive a successful response with access and refresh tokens!

---

## Common Issues

### Issue 1: Service Not Running
**Error**: Cannot connect to http://localhost:8080

**Solution**: Start your services:
```bash
docker-compose up -d
```

### Issue 2: User Not Found
**Error**: "Invalid credentials or account blocked"

**Solution**: Check your user exists and is active:
```sql
SELECT id, username, email, is_active, is_blocked FROM users WHERE username = 'your_username';
```

Make sure:
- `is_active` = true
- `is_blocked` = false

### Issue 3: Still Getting Password Mismatch

**Solution**: Verify the password_hash field was updated:
```sql
SELECT username, password_hash FROM users WHERE username = 'your_username';
```

The hash should start with `$2a$10$` or `$2b$10$` (bcrypt format).

---

## Manual Database Update Example

If you prefer to do it manually via SQL:

```sql
-- 1. Generate the hash using the API first
-- 2. Then run this SQL (replace values with yours):

UPDATE users 
SET password_hash = '$2a$10$YourGeneratedHashHere...' 
WHERE username = 'your_username';

-- Verify it worked:
SELECT username, 
       LEFT(password_hash, 20) as hash_preview,
       is_active,
       is_blocked
FROM users 
WHERE username = 'your_username';
```

---

## Security Note

⚠️ **IMPORTANT**: The `/hash-password` utility endpoint should be removed or secured in production. It's only meant for development and testing.

To remove it later, delete this line from `services/auth-service/routes/auth.routes.go`:
```go
rg.POST("/hash-password", controllers.HashPasswordUtility)
```

---

## Testing Your Login

Once fixed, you can test login with any of these methods:

### PowerShell:
```powershell
$loginBody = @{
    login = "your_username"
    password = "your_password"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$response
```

### Using the existing test script:
```powershell
.\test-login-refresh.ps1
```

---

## Need More Help?

If you're still having issues:

1. Check the auth-service logs:
   ```bash
   docker-compose logs auth-service
   ```

2. Verify database connection:
   ```bash
   curl http://localhost:8080/auth/health
   ```

3. Check that your user exists:
   ```sql
   SELECT * FROM users WHERE username = 'your_username';
   ```


