# Simple Microservices Architecture Guide

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Kong API Gateway                      │
│                    (Port 8000)                           │
│              Single Entry Point                         │
└─────────────────────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌──────────┐  ┌──────────────┐  ┌──────────────┐
│  Auth    │  │ Organization│  │ Appointment │
│ Service  │  │   Service    │  │   Service   │
│  :8080   │  │    :8081     │  │    :8082    │
└──────────┘  └──────────────┘  └──────────────┘
     │               │               │
     └───────────────┼───────────────┘
                     │
                     ▼
            ┌─────────────────┐
            │   PostgreSQL     │
            │   Database      │
            │   (Shared)      │
            └─────────────────┘
```

## 📦 Services

### 1. Auth Service
- **Port**: 8080 (internal)
- **API Path**: `/api/auth`
- **Purpose**: User authentication, JWT tokens, roles
- **Repository**: `drandme-auth-service`

### 2. Organization Service
- **Port**: 8081 (internal)
- **API Path**: `/api/organizations`
- **Purpose**: Organizations, clinics, doctors, patients
- **Repository**: `drandme-organization-service`

### 3. Appointment Service
- **Port**: 8082 (internal)
- **API Path**: `/api/v1`
- **Purpose**: Appointments, check-ins, vitals, follow-ups
- **Repository**: `drandme-appointment-service`

## 🔐 Authentication Flow

1. **User logs in** → `POST http://localhost:8000/api/auth/login`
2. **Auth service** generates JWT token
3. **User uses token** for all other services:
   - `GET http://localhost:8000/api/organizations/clinics` (with token)
   - `GET http://localhost:8000/api/v1/appointments` (with token)

**✅ Same token works for ALL services!**

## 🗄️ Database

- **All services** share the **same PostgreSQL database**
- **Database name**: `drandme`
- **Each service** has its own migration files

---

## 📁 Repository Structure

### When Services Are in Separate Repos:

```
drandme-auth-service/
├── .gitignore
── Dockerfile
├ docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── middleware/
├── models/
├── routes/
└── migrations/
    ├── 001_initial_auth_schema.sql
    └── 002_user_management_features.sql

drandme-organization-service/
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── middleware/
├── models/
├── routes/
└── migrations/
    ├── 001_initial_organization_schema.sql
    └── ... (18 files)

drandme-appointment-service/
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── main.go
├── config/
├── controllers/
├── middleware/
├── models/
├── routes/
└── migrations/
    ├── 001_initial_appointment_schema.sql
    └── ... (12 files)
```

---

## 🌐 Kong Configuration - Where to Put It?

### Option 1: Separate Kong Repository (Recommended) ✅

Create a dedicated repository for Kong and infrastructure:

```
drandme-kong-gateway/
├── README.md
├── docker-compose.yml
├── kong.yml
└── .env
```

**Why?**
- ✅ Centralized API gateway management
- ✅ Easy to update routing without touching services
- ✅ Can be managed by DevOps team
- ✅ Version control for API routing

**docker-compose.yml** (in Kong repo):
```yaml
version: '3.8'

services:
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
      - kong_network

  # Services (from other repos)
  auth-service:
    image: your-registry/drandme-auth-service:latest
    # or build from git
    networks:
      - kong_network

  organization-service:
    image: your-registry/drandme-organization-service:latest
    networks:
      - kong_network

  appointment-service:
    image: your-registry/drandme-appointment-service:latest
    networks:
      - kong_network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: drandme
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres123
    networks:
      - kong_network

networks:
  kong_network:
    driver: bridge
```

### Option 2: Include in Each Service (Not Recommended) ❌

**Problem**: Kong config would be duplicated in each repo

### Option 3: Include in One Service (Auth Service) ⚠️

**Problem**: Other services depend on auth-service repo for Kong config

---

## ✅ Recommended Setup: Separate Kong Repository

### Step 1: Create Kong Repository

```bash
mkdir drandme-kong-gateway
cd drandme-kong-gateway
git init
```

### Step 2: Copy Kong Files

```bash
# Copy from monorepo
cp kong.yml .
cp docker-compose.yml .  # (modified version)
```

### Step 3: Update kong.yml for Production

Update service URLs to use Docker service names or production URLs:

```yaml
services:
  - name: auth-service
    url: http://auth-service:8080  # Docker network
    # OR for production:
    # url: https://auth-service.yourdomain.com
```

### Step 4: Deploy Services

Each service repository builds and pushes Docker images:

```bash
# In each service repo
docker build -t your-registry/drandme-auth-service:latest .
docker push your-registry/drandme-auth-service:latest
```

### Step 5: Deploy Kong

In Kong repository, pull and run all services:

```bash
docker-compose up -d
```

---

## 🚀 Deployment Flow

### Development (Monorepo)
```bash
# All in one place
docker-compose up
```

### Production (Separate Repos)
```bash
# 1. Build and push each service
cd drandme-auth-service
docker build -t auth-service:latest .
docker push auth-service:latest

# 2. Deploy Kong + Services
cd drandme-kong-gateway
docker-compose up -d
```

---

## ✅ Can You Push Each Service to Separate Git Repos?

### YES! ✅ Each service is 100% independent:

1. **✅ No cross-service imports**
2. **✅ Each has its own go.mod**
3. **✅ Each has its own Dockerfile**
4. **✅ Each has its own migrations**
5. **✅ Each has its own docker-compose.yml** (for local dev)

### What Each Service Needs:

- ✅ All source code
- ✅ Dockerfile
- ✅ docker-compose.yml (for local development)
- ✅ go.mod and go.sum
- ✅ migrations/ directory
- ✅ README.md
- ✅ .gitignore

### What Goes in Kong Repository:

- ✅ kong.yml (routing configuration)
- ✅ docker-compose.yml (orchestrates all services)
- ✅ .env (environment variables)
- ✅ README.md (deployment instructions)

---

## 📋 Quick Checklist

### Before Pushing to Separate Repos:

- [x] ✅ No cross-service imports (verified)
- [x] ✅ Each service has own migrations (done)
- [x] ✅ Each service has own Dockerfile (done)
- [x] ✅ Each service has own docker-compose.yml (done)
- [x] ✅ JWT secrets configured (done)
- [x] ✅ Kong routing configured (done)

### After Pushing to Separate Repos:

1. ✅ Create Kong repository
2. ✅ Copy kong.yml to Kong repo
3. ✅ Update docker-compose.yml in Kong repo
4. ✅ Deploy services as Docker images
5. ✅ Deploy Kong to orchestrate services

---

## 🎯 Summary

**✅ YES - You can push each service to separate Git repos!**

**Kong File Location**: 
- **Recommended**: Separate `drandme-kong-gateway` repository
- **Contains**: kong.yml, docker-compose.yml, deployment configs

**Each Service Repo Contains**:
- Service code
- Migrations
- Dockerfile
- Local docker-compose.yml (for development)

**Kong Repo Contains**:
- kong.yml (routing)
- docker-compose.yml (orchestration)
- Deployment scripts

---

## 📝 Example: Production Setup

```
GitHub:
- drandme-auth-service (service code)
- drandme-organization-service (service code)
- drandme-appointment-service (service code)
- drandme-kong-gateway (Kong + orchestration)
- drandme-database-migrations (optional - shared migrations)
```

**Deployment**:
1. Build Docker images from service repos
2. Deploy using Kong repo's docker-compose.yml
3. Kong routes traffic to services

---

## ✅ Ready to Push!

Your services are ready to be pushed to separate repositories! 🚀

