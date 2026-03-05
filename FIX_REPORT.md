# Production Issue Investigation & Fix Report

## Issue Description
Appointment and timeslot data was disappearing after Docker restarts or system reboots. This was identified as a critical production issue.

## Root Cause Discovery
The investigation revealed a destructive database migration script within the `organization-service`.
- File: `services/organization-service/migrations/009_truncate_doctor_time_slots.sql`
- Content: `TRUNCATE TABLE doctor_time_slots RESTART IDENTITY CASCADE;`
- Mechanism: The migration runner script in `docker-compose.yml` executes **ALL** SQL files in the migrations directory on every startup. This specific file was causing the `doctor_time_slots` table—and by extension, related session data and appointment links—to be wiped every time the containers restarted.

## Fix Implementation
- **Action**: Permanently deleted the dangerous migration file `009_truncate_doctor_time_slots.sql`.
- **Result**: The table will no longer be truncated on startup. Data will now persist correctly across restarts.

## Environment Verification
1.  **Database Persistence**: Checked `docker-compose.yml`. The `postgres` service correctly uses a named volume `postgres_data` mapped to `/var/lib/postgresql/data`. This ensures data is stored on the host filesystem and survives container removal.
2.  **Shared Database**: Confirmed that `appointment-service` and `organization-service` both connect to the same persistent Postgres instance (`drandme` database). Code analysis of `appointment-service` confirms it queries `doctor_time_slots` directly, which explains why the truncation affected booking functionality.
3.  **Microservice Communication**: Verified that services are correctly addressing each other (or the shared DB) via Docker network aliases (`postgres`, etc.) and not `localhost`.

## Recommendations for Production Architecture
To prevent similar issues in the future and harden the architecture:

1.  **Adhere to "Migration State"**:
    - **Current**: shell script runs `*.sql` blindly.
    - **Recommended**: Use a real migration tool (like `golang-migrate`, `Flyway`, or `Liquibase`) that tracks which migrations have been applied in a `schema_migrations` table. This prevents re-running idempotent scripts and allows for safe rollbacks.

2.  **Database Isolation**:
    - **Current**: Services share a single `drandme` database and access each other's tables (e.g., `appointment-service` reads `doctor_time_slots`).
    - **Recommended**: Each microservice should own its data. `appointment-service` should call `organization-service` API to check availability, rather than querying the table directly. This decouples the services and allows them to evolve independently.

3.  **Secrets Management**:
    - **Current**: Secrets are hardcoded in `docker-compose.yml`.
    - **Recommended**: Move `DB_PASSWORD`, `JWT_SECRET`, etc., to a `.env` file (not committed to git) and reference them in `docker-compose.yml` like `${DB_PASSWORD}`.

## Next Steps for User
- **Rebuild Containers**: Run `docker-compose down` and `docker-compose up -d --build` to ensure the old migration file is removed from the container's filesystem (since it was volume-mounted, local deletion should be reflected immediately, but a restart is good practice).
- **Re-create Data**: Since previous restarts truncated the data, you will need to re-create the doctor time slots and sessions one last time. They will now persist permanently.
