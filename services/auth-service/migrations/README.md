# Auth Service Migrations

This directory contains database migrations specific to the auth-service.

## Migration Files

- `001_initial_auth_schema.sql` - Creates core authentication tables (users, roles, user_roles, refresh_tokens)
- `002_user_management_features.sql` - Adds user blocking, auditing, and password reset functionality

## Running Migrations

### Using Docker Compose

Migrations are automatically run when you start the service with docker-compose. The migration service runs before the main service starts.

### Manual Execution

You can run migrations manually using the provided script:

```bash
cd services/auth-service/migrations
./run-migrations.sh
```

Or using psql directly:

```bash
export PGPASSWORD=postgres123
for f in *.sql; do
  psql -h localhost -p 5432 -U postgres -d drandme -f "$f"
done
```

## Migration Order

Migrations are executed in alphabetical order. Ensure your migration files are named with a numeric prefix (e.g., `001_`, `002_`) to maintain the correct order.

## Dependencies

- Auth service migrations should run first as other services depend on the `users` and `roles` tables.

