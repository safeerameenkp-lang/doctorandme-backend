# Simple Microservices Architecture

## 🏗️ How It Works

```
User/Client
    │
    ▼
┌─────────────────┐
│  Kong Gateway   │  ← Single entry point (Port 8000)
│  (kong.yml)     │
└─────────────────┘
    │
    ├─── /api/auth ────────────► Auth Service (8080)
    │
    ├─── /api/organizations ───► Organization Service (8081)
    │
    └─── /api/v1 ───────────────► Appointment Service (8082)
                                    │
                                    ▼
                            ┌──────────────┐
                            │  PostgreSQL  │
                            │   Database   │
                            └──────────────┘
```

## 📦 Services

| Service | Port | API Path | What It Does |
|---------|------|----------|--------------|
| Auth | 8080 | `/api/auth` | Login, JWT tokens, users, roles |
| Organization | 8081 | `/api/organizations` | Orgs, clinics, doctors, patients |
| Appointment | 8082 | `/api/v1` | Appointments, check-ins, vitals |

## 🔐 Authentication

1. **Login**: `POST http://localhost:8000/api/auth/login`
2. **Get Token**: Response contains `accessToken`
3. **Use Token**: Send in `Authorization: Bearer <token>` header
4. **Works Everywhere**: Same token works for all services! ✅

## ✅ Can You Push to Separate Git Repos?

### YES! ✅

Each service is **100% independent**:
- ✅ No imports from other services
- ✅ Own Dockerfile
- ✅ Own migrations
- ✅ Own docker-compose.yml
- ✅ Own README.md

### Push Each Service:

```bash
# 1. Auth Service
cd services/auth-service
git init
git remote add origin https://github.com/yourorg/drandme-auth-service.git
git push

# 2. Organization Service  
cd services/organization-service
git init
git remote add origin https://github.com/yourorg/drandme-organization-service.git
git push

# 3. Appointment Service
cd services/appointment-service
git init
git remote add origin https://github.com/yourorg/drandme-appointment-service.git
git push
```

## 🌐 Where to Put Kong File?

### ✅ Create Separate Repository: `drandme-kong-gateway`

**Put kong.yml here!**

```
drandme-kong-gateway/
├── kong.yml           ← Kong routing config
├── docker-compose.yml ← Orchestrates all services
└── .env               ← Environment variables
```

**Why Separate?**
- ✅ Centralized gateway management
- ✅ Easy to update routing
- ✅ Services stay independent
- ✅ Best practice for microservices

## 🚀 Quick Setup

### 1. Push Services (3 separate repos)
- `drandme-auth-service`
- `drandme-organization-service`
- `drandme-appointment-service`

### 2. Create Kong Repo
- `drandme-kong-gateway`
- Contains: `kong.yml`, `docker-compose.yml`

### 3. Deploy
```bash
# In Kong repo
docker-compose up -d
```

## ✅ Summary

- ✅ **Each service** → Separate Git repo
- ✅ **Kong file** → Separate `drandme-kong-gateway` repo
- ✅ **Same JWT token** works for all services
- ✅ **All services** share same database

**Everything is ready!** 🎉

