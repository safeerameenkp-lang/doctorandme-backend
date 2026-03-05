# Quick Build Each Service

## 🚀 One-Line Commands

### Build Individual Services

```bash
# Auth Service
docker-compose build auth-service

# Organization Service
docker-compose build organization-service

# Appointment Service
docker-compose build appointment-service
```

### Build All Services One by One

```bash
docker-compose build auth-service
docker-compose build organization-service
docker-compose build appointment-service
```

### Using Scripts

**PowerShell:**
```powershell
.\build-each-service.ps1 auth
.\build-each-service.ps1 organization
.\build-each-service.ps1 appointment
.\build-each-service.ps1 all
```

**Bash:**
```bash
chmod +x build-each-service.sh
./build-each-service.sh auth
./build-each-service.sh organization
./build-each-service.sh appointment
./build-each-service.sh all
```

---

## ✅ Verify Builds

```bash
# Check built images
docker images | grep drandme

# Or
docker images drandme-auth-service
docker images drandme-organization-service
docker images drandme-appointment-service
```

---

## 🎯 That's It!

Build each service as needed! 🚀

