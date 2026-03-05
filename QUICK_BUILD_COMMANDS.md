# Quick Build Commands - Microservices

## 🚀 One-Command Build & Run

### Windows (PowerShell)
```powershell
# Run the build script
.\BUILD_AND_RUN.ps1

# Or manually
docker-compose up -d --build
```

### Linux/Mac
```bash
# Run the build script
chmod +x BUILD_AND_RUN.sh
./BUILD_AND_RUN.sh

# Or manually
docker-compose up -d --build
```

---

## 📋 Step-by-Step Build

### 1. Build All Services
```bash
docker-compose build
```

### 2. Start All Services
```bash
docker-compose up -d
```

### 3. View Logs
```bash
docker-compose logs -f
```

### 4. Check Status
```bash
docker-compose ps
```

---

## 🔍 Verify Services

```bash
# Kong Gateway
curl http://localhost:8000/api/auth/health

# All Services
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health
```

---

## 🛑 Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

## ✅ That's It!

Run `docker-compose up -d --build` to build and start everything! 🚀

