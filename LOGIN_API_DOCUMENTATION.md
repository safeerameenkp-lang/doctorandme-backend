# Login API Documentation

## 🔐 Login API Endpoint

### Endpoint
```
POST http://localhost:8000/api/auth/login
```

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "login": "user@example.com",  // Can be email, phone, or username
  "password": "your-password"
}
```

### Request Body Fields
- `login` (string, required): Email, phone number, or username
- `password` (string, required): User password (minimum 8 characters)

### Success Response (200 OK)
```json
{
  "id": "user-uuid",
  "firstName": "John",
  "lastName": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "phone": "1234567890",
  "roles": ["user", "doctor"],
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600
}
```

### Error Responses

#### 400 Bad Request - Validation Error
```json
{
  "error": "Validation failed",
  "message": "Invalid input data",
  "code": "VALIDATION_ERROR",
  "details": "Key: 'LoginInput.Login' Error:Field validation for 'Login' failed on the 'required' tag"
}
```

#### 401 Unauthorized - Invalid Credentials
```json
{
  "error": "Invalid credentials or account blocked"
}
```

#### 401 Unauthorized - Password Mismatch
```json
{
  "error": "Invalid credentials",
  "debug": "Password mismatch"
}
```

---

## 📋 Complete Auth API Endpoints

### Base URL
- **Direct Service**: `http://localhost:8080/api/auth`
- **Through Kong Gateway**: `http://localhost:8000/api/auth`

### Public Endpoints (No Authentication Required)

#### 1. Health Check
```
GET /api/auth/health
```

#### 2. Register
```
POST /api/auth/register
```
**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "phone": "1234567890",
  "password": "password123"
}
```

#### 3. Login ⭐
```
POST /api/auth/login
```
**Request Body:**
```json
{
  "login": "user@example.com",  // email, phone, or username
  "password": "password123"
}
```

#### 4. Refresh Token
```
POST /api/auth/refresh
```
**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 5. Logout
```
POST /api/auth/logout
```
**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### Protected Endpoints (Require Authentication)

**All protected endpoints require:**
```
Authorization: Bearer <access_token>
```

#### 6. Get Profile
```
GET /api/auth/profile
```

#### 7. Update Profile
```
PUT /api/auth/profile
```

#### 8. Change Password
```
POST /api/auth/change-password
```

---

## 🧪 Testing the Login API

### Using cURL
```bash
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "user@example.com",
    "password": "password123"
  }'
```

### Using PowerShell
```powershell
$body = @{
    login = "user@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

### Using Postman
1. Method: `POST`
2. URL: `http://localhost:8000/api/auth/login`
3. Headers: `Content-Type: application/json`
4. Body (raw JSON):
```json
{
  "login": "user@example.com",
  "password": "password123"
}
```

---

## 🔑 Using the Access Token

After successful login, use the `accessToken` in subsequent requests:

```
Authorization: Bearer <access_token>
```

### Example: Get Profile
```bash
curl -X GET http://localhost:8000/api/auth/profile \
  -H "Authorization: Bearer <access_token>"
```

---

## 📝 Notes

1. **Login Field**: The `login` field accepts:
   - Email address
   - Phone number
   - Username

2. **Token Expiry**: 
   - Access Token: 1 hour (3600 seconds)
   - Refresh Token: 7 days

3. **Security**:
   - Account must be active (`is_active = true`)
   - Account must not be blocked (`is_blocked = false`)
   - Password is hashed using bcrypt

4. **Last Login**: The API automatically updates the `last_login` timestamp on successful login.

---

## ✅ Quick Test

Test if the login API is working:

```bash
# Health check
curl http://localhost:8000/api/auth/health

# Login (replace with actual credentials)
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"test@example.com","password":"test123"}'
```

