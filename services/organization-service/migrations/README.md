# Organization Service Migrations

This directory contains database migrations specific to the organization-service.

## Migration Files

- `001_initial_organization_schema.sql` - Creates core organization tables (organizations, clinics, doctors, patients, etc.)
- `002_add_mo_id_to_patients.sql` - Adds MO ID field to patients
- `003_doctor_leave_management.sql` - Creates doctor leave management table
- `004_clinic_specific_doctor_fees.sql` - Adds clinic-specific fee management
- `005_doctor_time_slots.sql` - Creates doctor time slots table
- `006_add_specific_date_to_time_slots.sql` - Adds date-specific slot support
- `007_departments.sql` - Creates departments table
- `008_doctor_tokens.sql` - Creates doctor token management
- `009_truncate_doctor_time_slots.sql` - Truncates time slots (maintenance)
- `010_make_day_of_week_nullable.sql` - Makes day_of_week nullable
- `011_session_based_slots.sql` - Creates session-based slot system
- `012_add_clinic_id_to_session_tables.sql` - Adds clinic_id to sessions
- `013_make_time_columns_nullable.sql` - Makes time columns nullable
- `014_relax_slot_type_constraint.sql` - Relaxes slot type constraints
- `015_clinic_specific_patients.sql` - Creates clinic-specific patients table
- `016_add_missing_patient_fields.sql` - Adds missing patient fields
- `017_create_doctor_tokens_table.sql` - Creates doctor tokens table
- `018_rename_slot_types.sql` - Renames slot types

## Running Migrations

### Using Docker Compose

Migrations are automatically run when you start the service with docker-compose. The migration service runs before the main service starts.

### Manual Execution

You can run migrations manually using the provided script:

```bash
cd services/organization-service/migrations
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

- Organization service migrations depend on auth-service migrations (users, roles tables must exist first).

