# Kong Deployment Guide - Separate Repositories

## 🎯 Where to Put Kong File?

### ✅ Recommended: Separate Kong Repository

Create a dedicated repository: `drandme-kong-gateway`

---

## 📁 Kong Repository Structure

```
drandme-kong-gateway/
├── README.md
├── docker-compose.yml          # Orchestrates all services
├── kong.yml                    # Kong routing configuration
├── .env                        # Environment variables
├── .gitignore
└── scripts/
    └── deploy.sh              # Deployment script
```

---

## 📄 Files in Kong Repository

### 1. kong.yml

This is your Kong routing configuration (copy from monorepo):

```yaml
_format_version: "3.0"
_transform: true

services:
  - name: auth-service
    url: http://auth-service:8080
    routes:
      - name: auth-routes
        paths:
          - /api/auth
        strip_path: false
        preserve_host: true
    plugins:
      - name: cors
        config:
          origins: ["*"]
          methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
          headers: ["Content-Type", "Authorization"]
          credentials: true

  - name: organization-service
    url: http://organization-service:8081
    routes:
      - name: organization-routes
        paths:
          - /api/organizations
        strip_path: false
        preserve_host: true
    plugins:
      - name: cors
        config:
          origins: ["*"]
          methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
          headers: ["Content-Type", "Authorization"]
          credentials: true

  - name: appointment-service
    url: http://appointment-service:8082
    routes:
      - name: appointment-routes
        paths:
          - /api/v1
        strip_path: false
        preserve_host: true
    plugins:
      - name: cors
        config:
          origins: ["*"]
          methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
          headers: ["Content-Type", "Authorization"]
          credentials: true

plugins:
  - name: request-id
  - name: correlation-id
```

### 2. docker-compose.yml (Kong Repository)

This orchestrates all services:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: drandme
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres123}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - drandme_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Kong API Gateway
  kong:
    image: kong:3.4
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /kong/kong.yml
      KONG_PROXY_LISTEN: 0.0.0.0:8000
      KONG_ADMIN_LISTEN: 0.0.0.0:8001
    ports:
      - "8000:8000"   # API Gateway
      - "8001:8001"   # Admin API
    volumes:
      - ./kong.yml:/kong/kong.yml:ro
    depends_on:
      - auth-service
      - organization-service
      - appointment-service
    networks:
      - drandme_network
    restart: unless-stopped

  # Auth Service (from Docker image or build)
  auth-service:
    # Option A: Use pre-built image
    image: your-registry/drandme-auth-service:latest
    
    # Option B: Build from Git
    # build:
    #   context: https://github.com/yourorg/drandme-auth-service.git
    #   dockerfile: Dockerfile
    
    expose:
      - "8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-postgres123}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8080
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - drandme_network

  # Organization Service
  organization-service:
    image: your-registry/drandme-organization-service:latest
    expose:
      - "8081"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-postgres123}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8081
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - drandme_network

  # Appointment Service
  appointment-service:
    image: your-registry/drandme-appointment-service:latest
    expose:
      - "8082"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-postgres123}
      DB_NAME: drandme
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
      PORT: 8082
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - drandme_network

volumes:
  postgres_data:

networks:
  drandme_network:
    driver: bridge
```

### 3. .env (Kong Repository)

```env
# Database
DB_PASSWORD=your-strong-password-here

# JWT Secrets (MUST be same across all services)
JWT_ACCESS_SECRET=your-strong-access-secret-here
JWT_REFRESH_SECRET=your-strong-refresh-secret-here

# Docker Registry (if using)
REGISTRY_URL=your-registry.com
```

---

## 🚀 Deployment Steps

### Step 1: Push Services to Separate Repos

```bash
# Auth Service
cd services/auth-service
git init
git remote add origin https://github.com/yourorg/drandme-auth-service.git
git add .
git commit -m "Initial commit"
git push -u origin main

# Organization Service
cd services/organization-service
git init
git remote add origin https://github.com/yourorg/drandme-organization-service.git
git add .
git commit -m "Initial commit"
git push -u origin main

# Appointment Service
cd services/appointment-service
git init
git remote add origin https://github.com/yourorg/drandme-appointment-service.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

### Step 2: Create Kong Repository

```bash
mkdir drandme-kong-gateway
cd drandme-kong-gateway
git init

# Copy files from monorepo
cp ../kong.yml .
# Create docker-compose.yml (see above)
# Create .env file

git add .
git commit -m "Initial Kong configuration"
git remote add origin https://github.com/yourorg/drandme-kong-gateway.git
git push -u origin main
```

### Step 3: Build and Push Docker Images

```bash
# In each service repository
docker build -t your-registry/drandme-auth-service:latest .
docker push your-registry/drandme-auth-service:latest

docker build -t your-registry/drandme-organization-service:latest .
docker push your-registry/drandme-organization-service:latest

docker build -t your-registry/drandme-appointment-service:latest .
docker push your-registry/drandme-appointment-service:latest
```

### Step 4: Deploy with Kong

```bash
# In Kong repository
docker-compose up -d
```

---

## 🔄 Development Workflow

### Option A: Local Development (Each Service Separate)

Each service can be developed independently:

```bash
# In auth-service repo
docker-compose up  # Runs only auth-service + postgres
```

### Option B: Full Stack Development (Kong Repo)

Use Kong repo to run everything:

```bash
# In kong-gateway repo
docker-compose up  # Runs all services + Kong
```

---

## 📋 Summary

**✅ YES - You can push each service to separate Git repos!**

**Kong File Location**: 
- **Put in**: `drandme-kong-gateway` repository
- **Contains**: kong.yml, docker-compose.yml, .env

**Service Repositories**:
- Each service in its own repo
- Contains: code, Dockerfile, migrations, local docker-compose.yml

**Kong Repository**:
- Orchestrates all services
- Contains: kong.yml, docker-compose.yml, deployment configs

---

## 🎯 Quick Answer

**Q: Where do I put kong.yml when services are in separate repos?**

**A: Create a separate `drandme-kong-gateway` repository and put kong.yml there!**

This is the cleanest approach and follows microservices best practices. ✅

