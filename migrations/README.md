# Database Migrations

This directory contains all database migrations for the DrAndMe system.

## ⚠️ For Independent Repositories

When services are separated into independent repositories, these migrations should be moved to a **dedicated migrations repository**:

- Repository: `drandme-database-migrations`
- Purpose: Centralized database schema management
- Used by: All services (they share the same database)

See `MIGRATIONS_FOR_INDEPENDENT_REPOS.md` for details.

## 📋 Migration Files

Migrations are run in numerical order:

1. `001_initial_schema.sql` - Base database schema
2. `002_add_mo_id_to_patients.sql` - Patient MO ID field
3. `003_admin_features.sql` - Admin features
4. `004_add_appointment_fields.sql` - Appointment fields
5. `005_user_management_features.sql` - User management
6. ... and more

## 🚀 Usage

### With Docker Compose (Monorepo)

Migrations are automatically applied when PostgreSQL starts:

```bash
docker-compose up postgres
```

### Manual Execution

```bash
# Run specific migration
psql -h localhost -U postgres -d drandme -f migrations/001_initial_schema.sql

# Or use the init script
./scripts/init-database.sh
```

## 📝 Adding New Migrations

1. Create new file: `030_your_migration_name.sql`
2. Use sequential numbering
3. Test locally first
4. Document changes in migration file

## ⚠️ Important Notes

- **Never modify existing migrations** - create new ones instead
- **Always test migrations** on a copy of production data
- **Backup database** before running migrations in production
- **Run migrations in order** - they depend on each other

