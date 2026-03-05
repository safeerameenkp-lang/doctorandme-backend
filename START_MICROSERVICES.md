# 🚀 Start Microservices - Quick Guide

## One Command to Build & Run Everything

```bash
docker-compose up -d --build
```

That's it! This will:
1. ✅ Build all 3 microservices (auth, organization, appointment)
2. ✅ Start PostgreSQL database
3. ✅ Run migrations in correct order
4. ✅ Start Kong API Gateway
5. ✅ Start all services

---

## Verify Everything is Running

```bash
# Check all containers
docker-compose ps

# Check logs
docker-compose logs -f

# Test services
curl http://localhost:8000/api/auth/health
curl http://localhost:8000/api/organizations/health
curl http://localhost:8000/api/v1/health
```

---

## Access Points

- **Kong Gateway**: http://localhost:8000
- **Kong Admin**: http://localhost:8001
- **PgAdmin**: http://localhost:5050
- **PostgreSQL**: localhost:5432

---

## Stop Services

```bash
docker-compose down
```

---

## 🎯 That's All!

Just run `docker-compose up -d --build` and you're done! 🚀

