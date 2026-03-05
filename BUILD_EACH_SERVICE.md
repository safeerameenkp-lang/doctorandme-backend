# Build Each Microservice Individually

## 🚀 Build Commands for Each Service

### Option 1: Using Docker Compose (Recommended)

```bash
# Build only auth-service
docker-compose build auth-service

# Build only organization-service
docker-compose build organization-service

# Build only appointment-service
docker-compose build appointment-service
```

### Option 2: Using Docker Directly

```bash
# Build auth-service
docker build -f services/auth-service/Dockerfile -t drandme-auth-service:latest .

# Build organization-service
docker build -f services/organization-service/Dockerfile -t drandme-organization-service:latest .

# Build appointment-service
docker build -f services/appointment-service/Dockerfile -t drandme-appointment-service:latest .
```

---

## 📋 Step-by-Step: Build Each Service

### 1. Build Auth Service

```bash
# Build
docker-compose build auth-service

# Or with Docker directly
docker build -f services/auth-service/Dockerfile -t drandme-auth-service:latest .

# Check if built
docker images | grep auth-service
```

### 2. Build Organization Service

```bash
# Build
docker-compose build organization-service

# Or with Docker directly
docker build -f services/organization-service/Dockerfile -t drandme-organization-service:latest .

# Check if built
docker images | grep organization-service
```

### 3. Build Appointment Service

```bash
# Build
docker-compose build appointment-service

# Or with Docker directly
docker build -f services/appointment-service/Dockerfile -t drandme-appointment-service:latest .

# Check if built
docker images | grep appointment-service
```

---

## 🔍 Verify Builds

```bash
# List all built images
docker images | grep drandme

# Check specific service
docker images drandme-auth-service
docker images drandme-organization-service
docker images drandme-appointment-service
```

---

## 🧪 Test Individual Service Build

### Test Auth Service

```bash
# Build
docker-compose build auth-service

# Run standalone (requires DB)
docker run --rm \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres123 \
  -e DB_NAME=drandme \
  -e JWT_ACCESS_SECRET=test-secret \
  -e JWT_REFRESH_SECRET=test-refresh \
  -e PORT=8080 \
  drandme-auth-service:latest
```

### Test Organization Service

```bash
# Build
docker-compose build organization-service

# Run standalone (requires DB)
docker run --rm \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres123 \
  -e DB_NAME=drandme \
  -e JWT_ACCESS_SECRET=test-secret \
  -e JWT_REFRESH_SECRET=test-refresh \
  -e PORT=8081 \
  drandme-organization-service:latest
```

### Test Appointment Service

```bash
# Build
docker-compose build appointment-service

# Run standalone (requires DB)
docker run --rm \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres123 \
  -e DB_NAME=drandme \
  -e JWT_ACCESS_SECRET=test-secret \
  -e JWT_REFRESH_SECRET=test-refresh \
  -e PORT=8082 \
  drandme-appointment-service:latest
```

---

## 🔄 Rebuild Specific Service

```bash
# Rebuild without cache
docker-compose build --no-cache auth-service
docker-compose build --no-cache organization-service
docker-compose build --no-cache appointment-service

# Or with Docker directly
docker build --no-cache -f services/auth-service/Dockerfile -t drandme-auth-service:latest .
```

---

## 📝 Build All Services One by One

```bash
# Build all services individually
docker-compose build auth-service
docker-compose build organization-service
docker-compose build appointment-service

# Then start all
docker-compose up -d
```

---

## ✅ Quick Reference

| Service | Build Command | Image Name |
|---------|--------------|------------|
| Auth | `docker-compose build auth-service` | `drandme-auth-service:latest` |
| Organization | `docker-compose build organization-service` | `drandme-organization-service:latest` |
| Appointment | `docker-compose build appointment-service` | `drandme-appointment-service:latest` |

---

## 🎯 Ready to Build!

Build each service individually as needed! 🚀

