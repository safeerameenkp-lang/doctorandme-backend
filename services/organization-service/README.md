# Organization Service

Organization and clinic management microservice for the DrAndMe clinic management system.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development without Docker)

### Using Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourorg/drandme-organization-service.git
   cd drandme-organization-service
   ```

2. **Create `.env` file** (optional)
   ```env
   DB_HOST=postgres
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres123
   DB_NAME=drandme
   JWT_ACCESS_SECRET=your-access-secret-key-here
   JWT_REFRESH_SECRET=your-refresh-secret-key-here
   PORT=8081
   ```

3. **Start the service**
   ```bash
   docker-compose up --build
   ```

4. **Service will be available at**: `http://localhost:8081`

### Using Go (Local Development)

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set environment variables** (see Environment Variables section)

3. **Run the service**
   ```bash
   go run main.go
   ```

## 📡 API Endpoints

### Base Path: `/api/organizations`

#### Health Check:
- `GET /health` - Health check endpoint

#### Organizations:
- `POST /organizations` - Create organization
- `GET /organizations` - List organizations
- `GET /organizations/:id` - Get organization
- `PUT /organizations/:id` - Update organization
- `DELETE /organizations/:id` - Delete organization

#### Clinics:
- `POST /clinics` - Create clinic
- `GET /clinics` - List clinics
- `GET /clinics/:id` - Get clinic
- `PUT /clinics/:id` - Update clinic
- `DELETE /clinics/:id` - Delete clinic

#### Doctors:
- `POST /doctors` - Create doctor
- `GET /doctors` - List doctors
- `GET /doctors/:id` - Get doctor
- `PUT /doctors/:id` - Update doctor

#### And many more endpoints...

All endpoints require authentication via JWT token.

## 🔧 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `postgres123` |
| `DB_NAME` | Database name | `drandme` |
| `JWT_ACCESS_SECRET` | Secret for JWT validation | **Required** |
| `JWT_REFRESH_SECRET` | Secret for refresh tokens | **Required** |
| `PORT` | Service port | `8081` |

## 🏗️ Architecture

This service is **100% independent**:
- ✅ Own middleware package for JWT validation
- ✅ No dependencies on other services
- ✅ Can be deployed separately
- ✅ Can be developed by different teams

## 🔒 Security

- JWT token validation
- Role-based access control (RBAC)
- Database connection pooling
- Input validation

## 📦 Building Docker Image

```bash
docker build -t drandme-organization-service:latest .
```

## 🧪 Testing

```bash
# Health check
curl http://localhost:8081/api/organizations/health

# Get organizations (requires auth token)
curl http://localhost:8081/api/organizations/organizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📝 Development

### Project Structure
```
.
├── config/          # Database configuration
├── controllers/     # Request handlers
├── middleware/      # JWT validation, RBAC, CORS
├── models/         # Data models
├── routes/         # API routes
├── utils/          # Utility functions
├── main.go         # Application entry point
└── Dockerfile      # Docker build file
```

## 🔗 Integration with Kong API Gateway

When using Kong API Gateway, this service is accessible at:
- `http://kong:8000/api/organizations` (internal)
- `http://localhost:8000/api/organizations` (through Kong)

## 📚 Related Services

- **Auth Service**: Handles authentication and user management
- **Appointment Service**: Handles appointments and scheduling

All services communicate through Kong API Gateway, not directly.

## 📄 License

[Your License Here]

