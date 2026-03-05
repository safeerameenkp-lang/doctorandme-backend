# Appointment Service

Appointment booking and management microservice for the DrAndMe clinic management system.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development without Docker)

### Using Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourorg/drandme-appointment-service.git
   cd drandme-appointment-service
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
   PORT=8082
   ```

3. **Start the service**
   ```bash
   docker-compose up --build
   ```

4. **Service will be available at**: `http://localhost:8082`

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

### Base Path: `/api/v1`

#### Health Check:
- `GET /health` - Health check endpoint

#### Appointments:
- `POST /appointments/simple` - Create simple appointment
- `GET /appointments/simple-list` - List appointments
- `GET /appointments/simple/:id` - Get appointment details
- `POST /appointments/simple/:id/reschedule` - Reschedule appointment
- `POST /appointments` - Create appointment
- `GET /appointments` - List appointments
- `PUT /appointments/:id` - Update appointment
- `POST /appointments/:id/cancel` - Cancel appointment

#### Check-ins:
- `POST /checkins` - Create check-in
- `GET /checkins` - List check-ins
- `GET /checkins/:id` - Get check-in

#### Vitals:
- `POST /vitals` - Record vitals
- `GET /vitals` - List vitals
- `GET /vitals/appointment/:appointment_id` - Get vitals by appointment

#### Reports:
- `GET /reports/daily-collection` - Daily collection report
- `GET /reports/pending-payments` - Pending payments report

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
| `PORT` | Service port | `8082` |

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
docker build -t drandme-appointment-service:latest .
```

## 🧪 Testing

```bash
# Health check
curl http://localhost:8082/api/v1/health

# Get appointments (requires auth token)
curl http://localhost:8082/api/v1/appointments \
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
- `http://kong:8000/api/v1` (internal)
- `http://localhost:8000/api/v1` (through Kong)

## 📚 Related Services

- **Auth Service**: Handles authentication and user management
- **Organization Service**: Manages organizations and clinics

All services communicate through Kong API Gateway, not directly.

## 📄 License

[Your License Here]

