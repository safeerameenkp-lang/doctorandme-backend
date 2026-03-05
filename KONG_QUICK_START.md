# Kong API Gateway - Quick Start Guide

## 🚀 Quick Start

### 1. Start All Services with Kong

```bash
docker-compose up --build
```

This will start:
- PostgreSQL database
- Kong API Gateway (port 8000)
- Auth Service (internal)
- Organization Service (internal)
- Appointment Service (internal)

### 2. Access Services Through Kong

All API calls should go through Kong at `http://localhost:8000`:

```bash
# Health checks
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health

# Auth endpoints
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'

# Organization endpoints (requires auth token)
curl http://localhost:8000/api/organizations/organizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Appointment endpoints (requires auth token)
curl http://localhost:8000/api/v1/appointments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. Kong Admin API

Access Kong Admin API for management:

```bash
# Get all services
curl http://localhost:8001/services

# Get all routes
curl http://localhost:8001/routes

# Get service health
curl http://localhost:8001/status
```

## 🔧 Configuration

### Kong Configuration File

The Kong configuration is in `kong.yml`. It defines:
- Services (auth, organization, appointment)
- Routes (API paths)
- Plugins (CORS, request ID, etc.)

### Update Kong Configuration

1. Edit `kong.yml`
2. Restart Kong:
   ```bash
   docker-compose restart kong
   ```

Or reload Kong configuration:
```bash
docker-compose exec kong kong reload
```

## 📊 Service Endpoints

### Auth Service
- Base path: `/api/auth`
- Internal URL: `http://auth-service:8080`
- Public URL: `http://localhost:8000/api/auth`

### Organization Service
- Base path: `/api/organizations`
- Internal URL: `http://organization-service:8081`
- Public URL: `http://localhost:8000/api/organizations`

### Appointment Service
- Base path: `/api/v1`
- Internal URL: `http://appointment-service:8082`
- Public URL: `http://localhost:8000/api/v1`

## 🛠️ Troubleshooting

### Check Kong Status

```bash
docker-compose exec kong kong health
```

### View Kong Logs

```bash
docker-compose logs -f kong
```

### Check Service Connectivity

```bash
# Test if services are reachable from Kong
docker-compose exec kong wget -O- http://auth-service:8080/api/auth/health
docker-compose exec kong wget -O- http://organization-service:8081/api/organizations/health
docker-compose exec kong wget -O- http://appointment-service:8082/api/v1/health
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart auth-service
docker-compose restart kong
```

## 🔒 Security Notes

1. **Services are not directly exposed**: Only Kong is exposed on port 8000
2. **Internal communication**: Services communicate via Docker network
3. **CORS**: Configured in Kong for all services
4. **Authentication**: Still handled by services (JWT validation)

## 📝 Next Steps

- See `MICROSERVICES_MIGRATION_GUIDE.md` for separating services into different repositories
- See `SHARED_SECURITY_MODULE_SETUP.md` for setting up the shared security module
- Configure Kong plugins for rate limiting, authentication, etc.

