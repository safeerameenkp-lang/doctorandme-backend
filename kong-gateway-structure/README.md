# Kong API Gateway - Microservices Orchestration

This repository contains the Kong API Gateway configuration and orchestration for all DrAndMe microservices.

## 📁 Repository Structure

```
drandme-kong-gateway/
├── README.md                    # This file
├── docker-compose.yml           # Orchestrates all services
├── kong.yml                     # Kong routing configuration
├── .env.example                 # Environment variables template
├── .env                         # Your actual environment variables (gitignored)
├── .gitignore                   # Git ignore rules
│
├── migrations/                  # Database migrations (optional)
│   ├── auth-service/
│   │   ├── 001_initial_auth_schema.sql
│   │   └── 002_user_management_features.sql
│   ├── organization-service/
│   │   ├── 001_initial_organization_schema.sql
│   │   └── ... (18 files)
│   └── appointment-service/
│       ├── 001_initial_appointment_schema.sql
│       └── ... (12 files)
│
└── scripts/                     # Deployment scripts
    ├── deploy.sh                # Main deployment script
    ├── setup.sh                 # Initial setup script
    ├── health-check.sh          # Health check all services
    └── update-kong.sh           # Update Kong configuration
```

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/yourorg/drandme-kong-gateway.git
cd drandme-kong-gateway

# Copy environment template
cp .env.example .env

# Edit .env with your values
nano .env
```

### 2. Configure Environment Variables

Edit `.env` file:
```env
DB_PASSWORD=your-strong-password
JWT_ACCESS_SECRET=your-strong-secret-min-32-chars
JWT_REFRESH_SECRET=your-strong-secret-min-32-chars
```

### 3. Deploy Services

**Option A: Using Pre-built Docker Images**
```bash
# Update .env with your registry URLs
AUTH_SERVICE_IMAGE=your-registry/drandme-auth-service:latest
ORG_SERVICE_IMAGE=your-registry/drandme-organization-service:latest
APPT_SERVICE_IMAGE=your-registry/drandme-appointment-service:latest

docker-compose up -d
```

**Option B: Build from Source**
```bash
# Clone service repositories first
git clone https://github.com/yourorg/drandme-auth-service.git
git clone https://github.com/yourorg/drandme-organization-service.git
git clone https://github.com/yourorg/drandme-appointment-service.git

# Update docker-compose.yml to build from local paths
docker-compose up -d --build
```

## 🌐 API Endpoints

All APIs are accessed through Kong Gateway:

- **Kong Gateway**: `http://localhost:8000`
- **Kong Admin API**: `http://localhost:8001`

### Service Endpoints:

- **Auth Service**: `http://localhost:8000/api/auth`
- **Organization Service**: `http://localhost:8000/api/organizations`
- **Appointment Service**: `http://localhost:8000/api/v1`

## 📋 Services

This repository orchestrates:

1. **PostgreSQL Database** - Shared database for all services
2. **Kong API Gateway** - Routes all API requests
3. **Auth Service** - Authentication and authorization
4. **Organization Service** - Organizations, clinics, doctors
5. **Appointment Service** - Appointments and bookings

## 🔐 Authentication

1. Login: `POST http://localhost:8000/api/auth/login`
2. Get JWT token from response
3. Use token in `Authorization: Bearer <token>` header
4. Same token works for all services! ✅

## 📦 Dependencies

- Docker & Docker Compose
- Kong 3.4+
- PostgreSQL 15+
- Service Docker images (or build from source)

## 🛠️ Management

### Start Services
```bash
docker-compose up -d
```

### Stop Services
```bash
docker-compose down
```

### View Logs
```bash
docker-compose logs -f kong
docker-compose logs -f auth-service
```

### Health Check
```bash
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health
```

## 📝 Configuration

### Kong Configuration
- Edit `kong.yml` to modify routing
- Restart Kong: `docker-compose restart kong`

### Service Configuration
- Edit `docker-compose.yml` for service settings
- Edit `.env` for environment variables

## 🔄 Updates

### Update Kong Config
1. Edit `kong.yml`
2. Restart Kong: `docker-compose restart kong`

### Update Service Images
1. Update `.env` with new image tags
2. Restart service: `docker-compose up -d --no-deps <service-name>`

## 📚 Related Repositories

- `drandme-auth-service` - Authentication service
- `drandme-organization-service` - Organization service
- `drandme-appointment-service` - Appointment service

## ✅ Status

All services are configured and ready for deployment! 🚀
