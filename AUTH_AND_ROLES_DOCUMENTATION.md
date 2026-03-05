# Authentication, Roles, and Clinic-Based Access Control

This document outlines the implementation details of the Authentication Service, covering login, registration, role-based access control (RBAC), and clinic-based scoping methods used across the Doctor & Me platform.

## 1. Authentication Methods

The Auth Service (`auth-service`) manages all authentication. It provides RESTful APIs to register new users and authenticate existing ones using JWT (JSON Web Tokens).

### 1.1 Signup / Registration Method
**Endpoint:** `POST /register`
**Purpose:** Creates a new user in the system and automatically provisions them with access and refresh tokens.

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",     // Optional
  "username": "johndoe123",              // Required, Min 3 chars
  "phone": "1234567890",                 // Optional
  "password": "SecurePassword123"        // Required, Min 8 chars
}
```

**Implementation Details:**
- Automatically validates email and phone number formats using Regex.
- Ensures the `username` and `email` are unique.
- Hashes the password using `bcrypt` (default cost).
- Generates an `accessToken` (15 minutes expiry) and a `refreshToken` (7 days expiry).
- Note: New users are currently assigned a default `super_admin` role automatically (intended to be updated in production depending on the sign-up flow).

### 1.2 Sign In / Login Method
**Endpoint:** `POST /login`
**Purpose:** Authenticates a user and returns JWT access & refresh tokens along with role and clinic context.

**Request Body:**
```json
{
  "login": "johndoe123",   // Can be Username, Email, or Phone number
  "password": "SecurePassword123"
}
```

**Implementation Details:**
- Checks if the user exists via `email`, `phone`, or `username` AND ensures `is_active = true` AND `is_blocked = false`.
- Verifies the requested password against the hashed password.
- Updates the `last_login` timestamp.
- Returns the user's details, tokens, and an array of `roles`.
- **Role Context Loading:** Crucially, during login, the system loads not just the role name/ID but also the `organization_id`, `clinic_id`, and `service_id` tied to that user's role assignment.

---

## 2. Role-Based Access Control (RBAC)

The system uses a robust middleware-based approach to control route access depending on user roles defined in the database.

### 2.1 The `AuthMiddleware`
Before any role check occurs, the request must pass through `AuthMiddleware(db)`. 
- Expects an `Authorization: Bearer <token>` header.
- Verifies the JWT signature with `JWT_ACCESS_SECRET`.
- Validates the user exists in the DB and is active.
- Sets `user_id` inside the Gin context.

### 2.2 Role Middlewares
The `middleware.go` defines several role-check functions applied to Gin router groups:

- **`RequireRole(db, expectedRoles...)`**: 
  Checks if the user has one of the passed expected roles in the `user_roles` table. A `super_admin` will automatically pass this check.
  
- **`RequireSuperAdmin(db)`**:
  Strictly requires the `super_admin` role. Sets context flags:
  ```go
  c.Set("is_super_admin", true)
  c.Set("is_organization_admin", false)
  c.Set("is_clinic_admin", false)
  ```

- **`RequireAnyAdmin(db)`**:
  Allows access if the user is a `super_admin`, `organization_admin`, or `clinic_admin`.

---

## 3. Scoped Access (Clinic & Organization Based)

Beyond mere roles, the platform relies on **contextual scoping** (Resource-based Access Control). A user might be a `clinic_admin`, but they should only be able to view or edit resources related to **their** specific clinic.

### 3.1 Organization-Based Scoping
**Middleware:** `RequireOrganizationAdmin(db)`
- Verifies the user has the `organization_admin` role (or is `super_admin`).
- Retrieves all `organization_id`s the user has access to from the `user_roles` table.
- Sets the `organization_ids` list in the Gin Context. Controllers can retrieve this array and append it to their SQL `WHERE` clauses to strictly filter data by the user's organization(s).

### 3.2 Clinic-Based Scoping
**Middleware:** `RequireClinicAdmin(db)`
- Verifies the user has the `clinic_admin` role (super_admins bypass this).
- Calls `GetUserClinicContext(db, userID)` which executes:
  ```sql
  SELECT DISTINCT ur.clinic_id
  FROM user_roles ur
  WHERE ur.user_id = $1 AND ur.clinic_id IS NOT NULL AND ur.is_active = true
  ```
- Collects all authorized `clinic_id`s for the user.
- Sets them in the Gin context: `c.Set("clinic_ids", clinicIDs)`.

### 3.3 How Controllers Use Scoping
In route definitions (see `routes/auth.routes.go`), routes are grouped by their scope:
1. `/admin/*` -> Protected by `RequireSuperAdmin`
2. `/org-admin/*` -> Protected by `RequireOrganizationAdmin`
3. `/clinic-admin/*` -> Protected by `RequireClinicAdmin`

Inside controllers like `ScopedListUsers` or `ListClinics`, the handler checks the Gin context flags (`is_super_admin`, `is_clinic_admin`, etc.). 
- If `is_super_admin` is true, it queries all records.
- If `is_clinic_admin` is true, it extracts `clinic_ids` from the context and injects a `WHERE clinic_id IN (...)` clause into the database query, ensuring users only see data scoped to their specific clinic.
