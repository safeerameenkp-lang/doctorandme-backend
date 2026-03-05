# Appointment Service Migrations

This directory contains database migrations specific to the appointment-service.

## Migration Files

- `001_initial_appointment_schema.sql` - Creates core appointment tables (appointments, patient_checkins, patient_vitals)
- `002_add_appointment_fields.sql` - Adds additional appointment fields
- `003_add_slot_id_to_appointments.sql` - Links appointments to time slots
- `004_add_clinic_patient_to_appointments.sql` - Adds clinic patient support
- `005_add_token_number_to_appointments.sql` - Adds token number field
- `006_add_individual_slot_id_to_appointments.sql` - Links to individual slots
- `007_add_slot_capacity_tracking.sql` - Adds capacity tracking
- `008_fix_duplicate_free_followups.sql` - Fixes duplicate follow-ups
- `009_create_follow_ups_table.sql` - Creates follow-ups table
- `010_add_followup_status_to_clinic_patients.sql` - Adds follow-up status tracking
- `011_fix_followup_constraint.sql` - Fixes follow-up constraints
- `012_add_logic_status_to_followups.sql` - Adds logic status to follow-ups

## Running Migrations

### Using Docker Compose

Migrations are automatically run when you start the service with docker-compose. The migration service runs before the main service starts.

### Manual Execution

You can run migrations manually using the provided script:

```bash
cd services/appointment-service/migrations
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

- Appointment service migrations depend on:
  - Auth-service migrations (users table must exist)
  - Organization-service migrations (patients, clinics, doctors tables must exist)

