# Microservices Migration Guide - Kong API Gateway

This guide will help you migrate from a monorepo to a fully separated microservices architecture using Kong API Gateway, with each service in its own repository.

## 🎯 Goals

1. **100% Microservices**: Each service runs independently
2. **Kong API Gateway**: All services accessed through Kong (no direct service-to-service communication)
3. **Separate Repositories**: Each service in its own Git repository
4. **No Direct Imports**: Services never import or link to each other

## 📋 Current Architecture

### Services
- **Auth Service** (Port 8080) - `/api/auth`
- **Organization Service** (Port 8081) - `/api/organizations`
- **Appointment Service** (Port 8082) - `/api/v1`

### Shared Module
- `shared/security` - JWT, Auth middleware, RBAC

## 🚀 Step 1: Setup Kong API Gateway

### Current Setup (Monorepo)

The `docker-compose.yml` has been updated to include Kong. Services are now accessed through Kong at port `8000`:

```bash
# Start all services with Kong
docker-compose up --build
```

### Access Points

- **Kong Gateway**: `http://localhost:8000` (Main entry point)
- **Kong Admin API**: `http://localhost:8001` (Management)
- **Services** (internal only, not exposed):
  - Auth Service: `http://auth-service:8080`
  - Organization Service: `http://organization-service:8081`
  - Appointment Service: `http://appointment-service:8082`

### API Routes Through Kong

All API calls should go through Kong:

```bash
# Auth Service
POST http://localhost:8000/api/auth/register
POST http://localhost:8000/api/auth/login

# Organization Service
GET http://localhost:8000/api/organizations/organizations
POST http://localhost:8000/api/organizations/clinics

# Appointment Service
GET http://localhost:8000/api/v1/appointments
POST http://localhost:8000/api/v1/appointments/simple
```

## 🔄 Step 2: Separate Services into Different Repositories

### Repository Structure

Create separate repositories for each service:

```
drandme-auth-service/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── models/
└── routes/

drandme-organization-service/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── models/
└── routes/

drandme-appointment-service/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── models/
└── routes/

drandme-shared-security/
├── go.mod
├── middleware.go
├── errors.go
└── README.md
```

### Step 2.1: Create Shared Security Module Repository

1. **Create new repository**: `drandme-shared-security`
2. **Copy shared module**:
   ```bash
   cp -r shared/security/* drandme-shared-security/
   ```
3. **Update go.mod** to be a proper Go module:
   ```go
   module github.com/yourorg/drandme-shared-security
   
   go 1.21
   ```
4. **Push to Git** and tag a version:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

### Step 2.2: Migrate Auth Service

1. **Create new repository**: `drandme-auth-service`
2. **Copy service files**:
   ```bash
   cp -r services/auth-service/* drandme-auth-service/
   ```
3. **Update go.mod** to use shared module from Git:
   ```go
   module github.com/yourorg/drandme-auth-service
   
   go 1.21
   
   require (
       github.com/gin-gonic/gin v1.9.1
       github.com/golang-jwt/jwt/v5 v5.0.0
       github.com/lib/pq v1.10.9
       golang.org/x/crypto v0.14.0
       github.com/yourorg/drandme-shared-security v1.0.0
   )
   ```
4. **Remove local replace directive** (if exists):
   ```go
   // Remove this line:
   // replace shared-security => ./shared-security
   ```
5. **Update imports** in all files:
   ```go
   // Change from:
   import "shared-security"
   
   // To:
   import "github.com/yourorg/drandme-shared-security"
   ```
6. **Create standalone docker-compose.yml** (see example below)
7. **Update Dockerfile** if needed (should work as-is)

### Step 2.3: Migrate Organization Service

Follow the same steps as Auth Service:
1. Create repository: `drandme-organization-service`
2. Copy service files
3. Update go.mod to use shared module from Git
4. Update imports
5. Create standalone docker-compose.yml

### Step 2.4: Migrate Appointment Service

Follow the same steps as Auth Service:
1. Create repository: `drandme-appointment-service`
2. Copy service files
3. Update go.mod to use shared module from Git
4. Update imports
5. Create standalone docker-compose.yml

## 🐳 Step 3: Individual Service Docker Compose

Each service repository should have its own `docker-compose.yml` for local development:

### Example: Auth Service docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: drandme
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - auth_network

  auth-service:
    build: .
    expose:
      - "8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres123
      DB_NAME: drandme
      JWT_ACCESS_SECRET: your-access-secret-key-here
      JWT_REFRESH_SECRET: your-refresh-secret-key-here
      PORT: 8080
    depends_on:
      - postgres
    networks:
      - auth_network

volumes:
  postgres_data:

networks:
  auth_network:
    driver: bridge
```

## 🌐 Step 4: Production Deployment with Kong

### Option A: Kong in Docker Compose (Development/Staging)

Create a separate `docker-compose.kong.yml` that orchestrates all services:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: drandme
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - microservices_network

  kong:
    image: kong:3.4
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /kong/kong.yml
      KONG_PROXY_LISTEN: 0.0.0.0:8000
      KONG_ADMIN_LISTEN: 0.0.0.0:8001
    ports:
      - "8000:8000"
      - "8001:8001"
    volumes:
      - ./kong.yml:/kong/kong.yml:ro
    networks:
      - microservices_network
    depends_on:
      - auth-service
      - organization-service
      - appointment-service

  auth-service:
    image: yourregistry/drandme-auth-service:latest
    expose:
      - "8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8080
    networks:
      - microservices_network
    depends_on:
      - postgres

  organization-service:
    image: yourregistry/drandme-organization-service:latest
    expose:
      - "8081"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8081
    networks:
      - microservices_network
    depends_on:
      - postgres

  appointment-service:
    image: yourregistry/drandme-appointment-service:latest
    expose:
      - "8082"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8082
    networks:
      - microservices_network
    depends_on:
      - postgres

volumes:
  postgres_data:

networks:
  microservices_network:
    driver: bridge
```

### Option B: Kubernetes (Production)

Deploy each service as a Kubernetes Deployment with Kong Ingress Controller.

### Option C: Cloud Services

- **AWS**: API Gateway + ECS/EKS
- **GCP**: Cloud Endpoints + Cloud Run/GKE
- **Azure**: API Management + AKS

## 🔒 Step 5: Service Communication Rules

### ✅ Allowed Communication Methods

1. **Through Kong API Gateway**: Services can call other services via Kong
   ```go
   // Example: Appointment service needs user info
   resp, err := http.Get("http://kong:8000/api/auth/profile")
   ```

2. **Database**: Services share the same database (eventual consistency)

3. **Message Queue** (Optional): For async communication
   - RabbitMQ
   - Apache Kafka
   - AWS SQS

### ❌ Forbidden Communication Methods

1. **Direct HTTP calls** between services (bypassing Kong)
   ```go
   // ❌ DON'T DO THIS
   resp, err := http.Get("http://auth-service:8080/api/auth/profile")
   ```

2. **Direct imports** between services
   ```go
   // ❌ DON'T DO THIS
   import "github.com/yourorg/drandme-auth-service/models"
   ```

3. **Shared code** (except shared-security module)

## 📝 Step 6: Update Environment Variables

### Shared Environment Variables

All services need:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`

### Service-Specific Variables

Add service-specific variables as needed:
- `KONG_URL` (if service needs to call other services)
- `REDIS_URL` (for caching)
- `RABBITMQ_URL` (for message queue)

## 🧪 Step 7: Testing

### Test Individual Services

```bash
# Test Auth Service
cd drandme-auth-service
docker-compose up
curl http://localhost:8080/api/auth/health
```

### Test Through Kong

```bash
# Start all services with Kong
docker-compose -f docker-compose.kong.yml up

# Test through Kong
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health
```

## 📚 Additional Resources

- [Kong Documentation](https://docs.konghq.com/)
- [Kong Declarative Config](https://docs.konghq.com/gateway/latest/production/deployment-topologies/db-less-and-declarative-config/)
- [Go Modules](https://go.dev/ref/mod)

## ✅ Migration Checklist

- [ ] Create `drandme-shared-security` repository
- [ ] Publish shared-security module to Git
- [ ] Create `drandme-auth-service` repository
- [ ] Update auth-service to use shared module from Git
- [ ] Create `drandme-organization-service` repository
- [ ] Update organization-service to use shared module from Git
- [ ] Create `drandme-appointment-service` repository
- [ ] Update appointment-service to use shared module from Git
- [ ] Test each service independently
- [ ] Setup Kong in production
- [ ] Update frontend to use Kong URL
- [ ] Remove direct service-to-service calls
- [ ] Document API endpoints
- [ ] Setup CI/CD for each service

## 🎉 Benefits

1. **Independent Deployment**: Deploy services separately
2. **Scalability**: Scale services independently
3. **Team Autonomy**: Different teams can own different services
4. **Technology Flexibility**: Use different tech stacks per service
5. **Fault Isolation**: One service failure doesn't bring down others
6. **API Gateway**: Centralized routing, authentication, rate limiting

