# Docker Build & Run Guide - Microservices

## 🚀 Complete Guide to Build and Run All Microservices

### Step 1: Build All Services

```bash
# Build all services at once
docker-compose build

# Or build specific service
docker-compose build auth-service
docker-compose build organization-service
docker-compose build appointment-service
```

### Step 2: Start All Services

```bash
# Start all services (build if needed)
docker-compose up -d

# Or start with build
docker-compose up -d --build

# View logs
docker-compose logs -f
```

### Step 3: Check Service Status

```bash
# Check all services are running
docker-compose ps

# Check specific service logs
docker-compose logs auth-service
docker-compose logs organization-service
docker-compose logs appointment-service
docker-compose logs kong
```

### Step 4: Verify Services

```bash
# Check Kong Gateway
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health

# Check Kong Admin
curl http://localhost:8001/status
```

---

## 📋 Complete Build Process

### Option 1: Build Everything (Recommended)

```bash
# 1. Stop any running containers
docker-compose down

# 2. Remove old images (optional)
docker-compose down --rmi all

# 3. Build all services
docker-compose build --no-cache

# 4. Start all services
docker-compose up -d

# 5. View logs
docker-compose logs -f
```

### Option 2: Build and Run in One Command

```bash
# Build and start all services
docker-compose up -d --build

# Follow logs
docker-compose logs -f
```

### Option 3: Build Services Individually

```bash
# Build auth service
docker-compose build auth-service

# Build organization service
docker-compose build organization-service

# Build appointment service
docker-compose build appointment-service

# Start all
docker-compose up -d
```

---

## 🔍 Troubleshooting

### If Build Fails

```bash
# Check Dockerfile syntax
docker build -f services/auth-service/Dockerfile -t test-auth .

# Check for missing files
ls -la services/auth-service/

# Check Go module
cd services/auth-service
go mod tidy
```

### If Services Don't Start

```bash
# Check logs
docker-compose logs

# Check if ports are in use
netstat -an | grep 8000
netstat -an | grep 5432

# Restart specific service
docker-compose restart auth-service
```

### If Database Connection Fails

```bash
# Check postgres is running
docker-compose ps postgres

# Check postgres logs
docker-compose logs postgres

# Test connection
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme
```

---

## ✅ Verification Checklist

After building and starting:

- [ ] All containers are running: `docker-compose ps`
- [ ] Kong Gateway responds: `curl http://localhost:8000/api/auth/health`
- [ ] Auth service healthy: `curl http://localhost:8000/api/auth/health`
- [ ] Organization service healthy: `curl http://localhost:8000/api/organizations/health`
- [ ] Appointment service healthy: `curl http://localhost:8000/api/v1/health`
- [ ] Database connected: Check service logs for "Connected to Postgres"
- [ ] Migrations completed: Check migration service logs

---

## 🎯 Quick Start Commands

```bash
# Complete setup (one command)
docker-compose up -d --build

# Check status
docker-compose ps

# View all logs
docker-compose logs -f

# Stop everything
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

## 📝 Expected Output

After running `docker-compose up -d --build`, you should see:

```
✅ postgres - Running
✅ auth-migrations - Completed
✅ organization-migrations - Completed
✅ appointment-migrations - Completed
✅ kong - Running
✅ auth-service - Running
✅ organization-service - Running
✅ appointment-service - Running
✅ pgadmin - Running (optional)
```

---

## 🌐 Access Points

- **Kong Gateway**: http://localhost:8000
- **Kong Admin**: http://localhost:8001
- **PgAdmin**: http://localhost:5050
- **PostgreSQL**: localhost:5432

---

## ✅ Ready to Build!

All services are configured and ready for Docker build! 🚀

