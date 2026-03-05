# Database Migrations for Independent Repositories

## 📋 The Challenge

Since all services share the **same PostgreSQL database**, we need to decide how to handle database migrations when services are in separate repositories.

## 🎯 Recommended Approach: Separate Migrations Repository

### Option 1: Dedicated Migrations Repository (Recommended)

Create a separate repository for database migrations:

```
drandme-database-migrations/
├── README.md
├── docker-compose.yml
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_mo_id_to_patients.sql
│   └── ...
└── scripts/
    └── run-migrations.sh
```

**Benefits:**
- ✅ Single source of truth for database schema
- ✅ Centralized migration management
- ✅ Easy to track database changes
- ✅ Can be run independently or as part of CI/CD

**Usage:**
```bash
# Clone migrations repo
git clone https://github.com/yourorg/drandme-database-migrations.git
cd drandme-database-migrations

# Run migrations
docker-compose up
```

### Option 2: Include in Auth Service

Since `auth-service` contains the initial schema and user management, you can include all migrations there:

```
drandme-auth-service/
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_mo_id_to_patients.sql
│   └── ...
└── ...
```

**Benefits:**
- ✅ Simple - migrations with the service that needs them most
- ✅ No separate repo needed

**Drawbacks:**
- ⚠️ Other services need to know where migrations are
- ⚠️ Less clear separation of concerns

### Option 3: Include in Each Service (Not Recommended)

Include all migrations in each service repository.

**Drawbacks:**
- ❌ Duplication - same files in multiple repos
- ❌ Hard to keep in sync
- ❌ Confusion about which migrations to run

## 🚀 Implementation: Option 1 (Recommended)

### Step 1: Create Migrations Repository

```bash
# Create new repository
mkdir drandme-database-migrations
cd drandme-database-migrations
git init

# Copy migrations
cp -r ../drandme-backend/migrations .
cp ../drandme-backend/scripts/init-database.sh scripts/
```

### Step 2: Create docker-compose.yml

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
      - ./migrations:/docker-entrypoint-initdb.d
      - ./scripts/init-database.sh:/docker-entrypoint-initdb.d/99-init-database.sh
    networks:
      - migrations_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:

networks:
  migrations_network:
    driver: bridge
```

### Step 3: Create README.md

```markdown
# Database Migrations

This repository contains all database migrations for the DrAndMe system.

## Usage

```bash
# Start PostgreSQL with migrations
docker-compose up

# Or run migrations manually
./scripts/run-migrations.sh
```

## Migration Order

Migrations are run in numerical order:
- 001_initial_schema.sql - Base schema
- 002_add_mo_id_to_patients.sql - Patient MO ID
- 003_admin_features.sql - Admin features
- ...
```

## 📝 Updating Service docker-compose.yml

When services are independent, they should **not** include migrations. Instead:

### Option A: Point to External Database

```yaml
# services/auth-service/docker-compose.yml
services:
  auth-service:
    environment:
      DB_HOST: postgres  # External database
      DB_PORT: 5432
      # ... other config
```

### Option B: Include PostgreSQL (for local dev only)

```yaml
# services/auth-service/docker-compose.yml
services:
  postgres:
    image: postgres:15-alpine
    # Note: No migrations - database should already exist
    environment:
      POSTGRES_DB: drandme
      # ...
```

## 🔄 Migration Workflow

### Development

1. **Run migrations repository first:**
   ```bash
   cd drandme-database-migrations
   docker-compose up
   ```

2. **Then start services:**
   ```bash
   cd drandme-auth-service
   docker-compose up
   ```

### Production

1. **Run migrations as part of deployment pipeline**
2. **Services connect to existing database**
3. **Migrations run before services start**

## ✅ Checklist for Independent Repos

- [ ] Create `drandme-database-migrations` repository
- [ ] Copy all migration files
- [ ] Create docker-compose.yml for migrations
- [ ] Update service docker-compose.yml to remove migrations
- [ ] Document migration workflow
- [ ] Set up CI/CD for migrations

## 📚 Related

- See `INDEPENDENT_REPOS_GUIDE.md` for service separation
- See `SEPARATE_REPOS_CHECKLIST.md` for complete checklist

