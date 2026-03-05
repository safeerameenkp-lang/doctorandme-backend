# Auth Service

Authentication and authorization microservice for the DrAndMe clinic management system.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development without Docker)

### Using Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourorg/drandme-auth-service.git
   cd drandme-auth-service
   ```

2. **Create `.env` file** (optional, defaults provided in docker-compose.yml)
   ```env
   DB_HOST=postgres
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres123
   DB_NAME=drandme
   JWT_ACCESS_SECRET=your-access-secret-key-here
   JWT_REFRESH_SECRET=your-refresh-secret-key-here
   PORT=8080
   ```

3. **Start the service**
   ```bash
   docker-compose up --build
   ```

4. **Service will be available at**: `http://localhost:8080`

### Using Go (Local Development)

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set environment variables**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres123
   export DB_NAME=drandme
   export JWT_ACCESS_SECRET=your-access-secret-key-here
   export JWT_REFRESH_SECRET=your-refresh-secret-key-here
   export PORT=8080
   ```

3. **Run the service**
   ```bash
   go run main.go
   ```

## 📡 API Endpoints

### Base Path: `/api/auth`

#### Public Endpoints:
- `GET /health` - Health check
- `POST /register` - Register a new user
- `POST /login` - Login with credentials
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout and revoke token

#### Protected Endpoints (Require Authentication):
- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `POST /change-password` - Change password

#### Admin Endpoints (Require Admin Role):
- `GET /admin/users` - List all users
- `POST /admin/users` - Create user
- `GET /admin/roles` - List roles
- `POST /admin/roles` - Create role
- And more...

See full API documentation in the codebase.

## 🔧 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `postgres123` |
| `DB_NAME` | Database name | `drandme` |
| `JWT_ACCESS_SECRET` | Secret for access tokens | **Required** |
| `JWT_REFRESH_SECRET` | Secret for refresh tokens | **Required** |
| `PORT` | Service port | `8080` |

## 🏗️ Architecture

This service is **100% independent**:
- ✅ Own middleware package
- ✅ No dependencies on other services
- ✅ Can be deployed separately
- ✅ Can be developed by different teams

## 🔒 Security

- JWT-based authentication
- Role-based access control (RBAC)
- Password hashing with bcrypt
- Token refresh mechanism
- Multi-device login support

## 📦 Building Docker Image

```bash
docker build -t drandme-auth-service:latest .
```

## 🧪 Testing

```bash
# Health check
curl http://localhost:8080/api/auth/health

# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'
```

## 📝 Development

### Project Structure
```
.
├── config/          # Database configuration
├── controllers/     # Request handlers
├── middleware/      # Auth, RBAC, CORS middleware
├── models/         # Data models
├── routes/         # API routes
├── main.go         # Application entry point
└── Dockerfile      # Docker build file
```

## 🔗 Integration with Kong API Gateway

When using Kong API Gateway, this service is accessible at:
- `http://kong:8000/api/auth` (internal)
- `http://localhost:8000/api/auth` (through Kong)

## 📚 Related Services

- **Organization Service**: Manages organizations and clinics
- **Appointment Service**: Handles appointments and scheduling

All services communicate through Kong API Gateway, not directly.

## 📄 License

[Your License Here]

