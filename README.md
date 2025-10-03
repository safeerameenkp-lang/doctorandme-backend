# SaaS Clinic Management System

A microservices-based clinic management system built with Go (Gin framework) and PostgreSQL.

## Architecture

This system uses a microservices architecture where each service runs in its own Docker container but connects to the same PostgreSQL database. The services share common security middleware through a shared module for consistency and maintainability.

### Services

1. **Auth Service** (Port 8080)
   - User registration and authentication
   - JWT token management
   - Role-based access control (RBAC)

2. **Organization Service** (Port 8081)
   - Organization management
   - Clinic management
   - External service management (labs, pharmacies)
   - Clinic-service linking

### Shared Security Module

The `shared/security/` module provides:
- JWT token generation and validation
- Authentication middleware
- Role-based access control (RBAC) middleware
- CORS middleware

This ensures consistent security logic across all services and makes maintenance easier.

## Database Schema

The system uses PostgreSQL with UUID primary keys and includes the following main entities:

- **users**: Core user information
- **roles**: System and custom roles
- **organizations**: Business entities
- **clinics**: Child entities of organizations
- **external_services**: Labs and pharmacies
- **user_roles**: Maps users to roles within organizations/clinics/services
- **clinic_service_links**: Links clinics to external services
- **refresh_tokens**: Multi-device login support

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd drandme-backend
   ```

2. **Start the services**
   ```bash
   docker-compose up --build
   ```

3. **The services will be available at:**
   - Auth Service: http://localhost:8080
   - Organization Service: http://localhost:8081
   - PostgreSQL: localhost:5432

## API Endpoints

### Auth Service (`/api/auth`)

**Purpose**: Handles authentication, authorization, JWT management, and user profile management only.

**Public Endpoints:**
- `GET /health` - Health check endpoint
- `POST /register` - Register a new user
- `POST /login` - Login with email/phone and password
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout and revoke refresh token

**Protected Endpoints (All Authenticated Users):**
- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `POST /change-password` - Change user password

### Organization Service (`/api/organizations`)

**Purpose**: Manages organizations, clinics, and creates admin accounts. Handles all business entity management.

**Health Check:**
- `GET /health` - Health check endpoint

**Organizations:**
- `POST /organizations` - Create organization (super_admin only)
- `POST /organizations/with-admin` - Create organization with admin account (super_admin only)
- `GET /organizations` - List all organizations
- `GET /organizations/:id` - Get organization by ID
- `PUT /organizations/:id` - Update organization (super_admin, organization_admin)
- `DELETE /organizations/:id` - Delete organization (super_admin only)

**Clinics:**
- `POST /clinics` - Create clinic (super_admin, organization_admin)
- `POST /clinics/with-admin` - Create clinic with admin account (super_admin, organization_admin)
- `GET /clinics` - List clinics (filter by organization_id)
- `GET /clinics/:id` - Get clinic by ID
- `PUT /clinics/:id` - Update clinic (super_admin, organization_admin, clinic_admin)
- `DELETE /clinics/:id` - Delete clinic (super_admin, organization_admin)

**Staff Management (Clinic Admin only):**
- `POST /staff` - Create staff member
- `GET /staff/clinic/:clinic_id` - Get clinic staff
- `PUT /staff/clinic/:clinic_id/:user_id/role` - Update staff role
- `DELETE /staff/clinic/:clinic_id/:user_id` - Deactivate staff member

**Doctor Management:**
- `POST /doctors` - Create regular doctor profile (clinic_admin only)
- `POST /doctors/main` - Create main doctor profile (super_admin only)
- `GET /doctors` - List doctors (filter by clinic_id)
- `GET /doctors/:id` - Get doctor by ID
- `PUT /doctors/:id` - Update doctor profile (clinic_admin, doctor)
- `DELETE /doctors/:id` - Delete doctor profile (clinic_admin only)

**Clinic Doctor Links (for main doctors):**
- `POST /clinic-doctor-links` - Link main doctor to clinic (clinic_admin only)
- `GET /clinic-doctor-links` - List links (filter by clinic_id, doctor_id)
- `GET /clinic-doctor-links/:id` - Get link by ID
- `DELETE /clinic-doctor-links/:id` - Unlink doctor from clinic (clinic_admin only)

**Doctor Schedule Management:**
- `POST /doctor-schedules` - Create doctor schedule (clinic_admin, doctor)
- `GET /doctor-schedules` - List schedules (filter by doctor_id, day_of_week)
- `GET /doctor-schedules/:id` - Get schedule by ID
- `PUT /doctor-schedules/:id` - Update schedule (clinic_admin, doctor)
- `DELETE /doctor-schedules/:id` - Delete schedule (clinic_admin, doctor)

**External Services:**
- `POST /services` - Create external service (super_admin only)
- `GET /services` - List services (filter by service_type)
- `GET /services/:id` - Get service by ID
- `PUT /services/:id` - Update service (super_admin only)
- `DELETE /services/:id` - Delete service (super_admin only)

**Clinic Service Links:**
- `POST /links` - Create clinic-service link (super_admin only)
- `GET /links` - List links (filter by clinic_id or service_id)
- `GET /links/:id` - Get link by ID
- `DELETE /links/:id` - Delete link (super_admin only)

## Authentication

All organization service endpoints require authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Default Roles and Hierarchy

The system implements a hierarchical role-based access control (RBAC) system:

### Role Hierarchy:
1. **super_admin** - Full system access
   - Can create organizations and organization admins
   - Can manage all organizations, clinics, and services
   - Can create and manage all user roles

2. **organization_admin** - Organization-level access
   - Can manage their organization and create clinics
   - Can create clinic admins for their organization's clinics
   - Can update organization information

3. **clinic_admin** - Clinic-level access
   - Can manage their specific clinic
   - Can create and manage staff members (doctors, receptionists, etc.)
   - Can assign roles to staff within their clinic

4. **Staff Roles** (assigned by clinic admin):
   - **doctor** - Patient care, appointments, prescriptions
   - **receptionist** - Patient management, appointments, billing
   - **pharmacist** - Prescription management, medication inventory
   - **lab_technician** - Lab orders, results, reports
   - **billing_staff** - Billing, payments, invoices

5. **patient** - Basic profile and appointment access

### Workflow:
1. **Super Admin** creates an organization with admin account using Organization Service
2. **Organization Admin** creates clinics with admin accounts using Organization Service
3. **Clinic Admin** creates and manages **Staff Members** using Organization Service
4. **Auth Service** handles all authentication and authorization
5. Each role has specific permissions within their scope

## Environment Variables

### Auth Service
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: Database name
- `JWT_ACCESS_SECRET`: Secret for access tokens
- `JWT_REFRESH_SECRET`: Secret for refresh tokens
- `PORT`: Service port (default: 8080)

### Organization Service
- Same as Auth Service
- `PORT`: Service port (default: 8081)

## Development

### Running Locally

1. **Start PostgreSQL**
   ```bash
   docker-compose up postgres -d
   ```

2. **Run migrations**
   ```bash
   ./scripts/migrate.sh
   ```

3. **Start services individually**
   ```bash
   # Auth Service
   cd services/auth-service
   go run main.go

   # Organization Service
   cd services/organization-service
   go run main.go
   ```

### Adding New Services

1. Create a new directory under `services/`
2. Follow the same structure as existing services
3. Add the service to `docker-compose.yml`
4. Update this README with the new service information

## Database Migrations

Database migrations are located in the `migrations/` directory and are automatically applied when starting PostgreSQL via Docker Compose.

To run migrations manually:
```bash
./scripts/migrate.sh
```

## Security Features

- JWT-based authentication
- Role-based access control (RBAC)
- Password hashing with bcrypt
- Refresh token rotation
- Multi-device login support

## Production Considerations

- Change default JWT secrets
- Use environment-specific database credentials
- Enable SSL/TLS for database connections
- Implement proper logging and monitoring
- Add rate limiting and request validation
- Use secrets management for sensitive data

## License

[Add your license information here]
