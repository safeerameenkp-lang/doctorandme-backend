# Migration Structure - Service-Specific Migrations

## Overview

Each microservice now has its own migration directory containing only the migrations relevant to that service. This allows each service to be completely independent and deployable in separate repositories.

## Migration Organization

### Auth Service (`services/auth-service/migrations/`)
**Purpose**: Core authentication and authorization tables

**Migrations**:
- `001_initial_auth_schema.sql` - Creates users, roles, user_roles, refresh_tokens tables
- `002_user_management_features.sql` - Adds user blocking, auditing, password reset

**Dependencies**: None (runs first)

### Organization Service (`services/organization-service/migrations/`)
**Purpose**: Organization, clinic, doctor, and patient management

**Migrations**:
- `001_initial_organization_schema.sql` - Creates organizations, clinics, doctors, patients tables
- `002_add_mo_id_to_patients.sql` - Adds MO ID field
- `003_doctor_leave_management.sql` - Doctor leave management
- `004_clinic_specific_doctor_fees.sql` - Clinic-specific fees
- `005_doctor_time_slots.sql` - Time slot management
- `006_add_specific_date_to_time_slots.sql` - Date-specific slots
- `007_departments.sql` - Department management
- `008_doctor_tokens.sql` - Token management
- `009_truncate_doctor_time_slots.sql` - Maintenance
- `010_make_day_of_week_nullable.sql` - Schema update
- `011_session_based_slots.sql` - Session-based slots
- `012_add_clinic_id_to_session_tables.sql` - Schema update
- `013_make_time_columns_nullable.sql` - Schema update
- `014_relax_slot_type_constraint.sql` - Schema update
- `015_clinic_specific_patients.sql` - Clinic-specific patients
- `016_add_missing_patient_fields.sql` - Patient fields
- `017_create_doctor_tokens_table.sql` - Token table
- `018_rename_slot_types.sql` - Schema update

**Dependencies**: Requires auth-service migrations (users, roles tables)

### Appointment Service (`services/appointment-service/migrations/`)
**Purpose**: Appointment booking, check-ins, vitals, and follow-ups

**Migrations**:
- `001_initial_appointment_schema.sql` - Creates appointments, patient_checkins, patient_vitals tables
- `002_add_appointment_fields.sql` - Additional appointment fields
- `003_add_slot_id_to_appointments.sql` - Links to time slots
- `004_add_clinic_patient_to_appointments.sql` - Clinic patient support
- `005_add_token_number_to_appointments.sql` - Token numbers
- `006_add_individual_slot_id_to_appointments.sql` - Individual slot links
- `007_add_slot_capacity_tracking.sql` - Capacity tracking
- `008_fix_duplicate_free_followups.sql` - Follow-up fixes
- `009_create_follow_ups_table.sql` - Follow-ups table
- `010_add_followup_status_to_clinic_patients.sql` - Status tracking
- `011_fix_followup_constraint.sql` - Constraint fixes
- `012_add_logic_status_to_followups.sql` - Logic status

**Dependencies**: Requires both auth-service and organization-service migrations

## Migration Execution Order

When deploying all services together, migrations must run in this order:

1. **Auth Service** migrations (no dependencies)
2. **Organization Service** migrations (depends on auth)
3. **Appointment Service** migrations (depends on auth + organization)

## Docker Compose Integration

The root `docker-compose.yml` includes migration services that run automatically:

- `auth-migrations` - Runs auth-service migrations
- `organization-migrations` - Runs organization-service migrations
- `appointment-migrations` - Runs appointment-service migrations

Each service depends on its migration service completing successfully before starting.

## Running Migrations

### Automatic (Docker Compose)
```bash
docker-compose up
```
Migrations run automatically before services start.

### Manual (Individual Service)
```bash
cd services/auth-service/migrations
./run-migrations.sh
```

### Manual (psql)
```bash
export PGPASSWORD=postgres123
for f in services/auth-service/migrations/*.sql; do
  psql -h localhost -p 5432 -U postgres -d drandme -f "$f"
done
```

## Independent Repository Deployment

When each service is in its own repository:

1. **Auth Service Repository**: Contains only `services/auth-service/migrations/`
2. **Organization Service Repository**: Contains only `services/organization-service/migrations/`
3. **Appointment Service Repository**: Contains only `services/appointment-service/migrations/`

Each service's `docker-compose.yml` (in `examples/`) includes its own migration service.

## Migration Naming Convention

Migrations are named with numeric prefixes to ensure execution order:
- `001_*.sql`
- `002_*.sql`
- `003_*.sql`
- etc.

Migrations are executed in alphabetical order, so the numeric prefix ensures correct sequencing.

## Notes

- All services share the same database (`drandme`) but have separate migration sets
- Foreign key references to tables in other services are handled gracefully (tables may not exist yet)
- Migration scripts use `CREATE TABLE IF NOT EXISTS` and `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` for idempotency
- Each service's migrations are self-contained and can be run independently (assuming dependencies exist)

