# Role Hierarchy and Scoping Documentation

## Overview

The Dr&Me platform implements a **hierarchical role-based access control (RBAC)** system with three levels of administrative access. Each level has specific scoping rules that restrict what data and operations they can access.

## Role Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                      SUPER ADMIN                             │
│                    (SaaS Owner/Platform Admin)               │
│                                                              │
│  • Platform-wide access to ALL data                         │
│  • Can manage users, orgs, clinics across entire platform   │
│  • Can create/modify/delete custom roles                    │
│  • Full system configuration access                         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ├────────────────────────────────┐
                            │                                │
                            ▼                                ▼
           ┌────────────────────────────┐   ┌────────────────────────────┐
           │   ORGANIZATION ADMIN        │   │   ORGANIZATION ADMIN       │
           │   (Organization A)          │   │   (Organization B)         │
           │                             │   │                            │
           │  • Scoped to Org A only     │   │  • Scoped to Org B only    │
           │  • Manages Org A users      │   │  • Manages Org B users     │
           │  • Manages Org A clinics    │   │  • Manages Org B clinics   │
           │  • Cannot access Org B      │   │  • Cannot access Org A     │
           └────────────────────────────┘   └────────────────────────────┘
                     │                                 │
           ┌─────────┴──────────┐             ┌──────┴────────────┐
           ▼                    ▼             ▼                   ▼
    ┌──────────────┐    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │CLINIC ADMIN  │    │CLINIC ADMIN  │  │CLINIC ADMIN  │  │CLINIC ADMIN  │
    │(Clinic 1)    │    │(Clinic 2)    │  │(Clinic 3)    │  │(Clinic 4)    │
    │              │    │              │  │              │  │              │
    │• Clinic 1    │    │• Clinic 2    │  │• Clinic 3    │  │• Clinic 4    │
    │  users only  │    │  users only  │  │  users only  │  │  users only  │
    └──────────────┘    └──────────────┘  └──────────────┘  └──────────────┘
```

## Three Levels of Admin Access

### 1. Super Admin / SaaS Owner

**Access Level:** Platform-wide (Global)

**Capabilities:**
- ✅ List, view, create, update, delete ANY user across the entire platform
- ✅ Block/unblock ANY user
- ✅ Activate/deactivate ANY user
- ✅ Change password for ANY user
- ✅ Assign/remove roles for ANY user at any scope (organization, clinic, service)
- ✅ Create, modify, delete custom roles
- ✅ Manage role permissions
- ✅ Access ALL organizations, clinics, and services
- ✅ View activity logs for ANY user
- ✅ Full platform configuration

**API Base Path:** `/api/v1/auth/admin/*`

**Scope:** No restrictions - can access everything

**Example Use Cases:**
- Platform owner managing the entire SaaS system
- Technical support accessing any tenant's data
- System-wide user management and troubleshooting
- Creating new organizations and assigning organization admins

---

### 2. Organization Admin

**Access Level:** Organization-scoped

**Capabilities:**
- ✅ List, view, create, update users **within their organization only**
- ✅ Activate/deactivate users **within their organization**
- ✅ Assign/remove roles **within their organization context**
- ✅ Manage clinics **within their organization**
- ✅ Manage doctors, staff **within their organization's clinics**
- ✅ View (but not create/modify) roles
- ❌ CANNOT access users from other organizations
- ❌ CANNOT block/unblock users (Super Admin only)
- ❌ CANNOT delete users
- ❌ CANNOT create or modify roles
- ❌ CANNOT change user passwords (except their own)

**API Base Path:** `/api/v1/auth/org-admin/*`

**Scope:** Limited to their assigned organization(s)

**Scoping Mechanism:**
- User must have `organization_admin` role
- Role assignment must have `organization_id` set
- All queries automatically filter by the admin's organization_id(s)

**Example Use Cases:**
- Hospital system administrator managing all their clinics
- Corporate admin managing company-wide users
- Regional manager overseeing multiple clinic locations

---

### 3. Clinic Admin

**Access Level:** Clinic-scoped

**Capabilities:**
- ✅ List, view, create, update users **within their clinic only**
- ✅ Activate/deactivate users **within their clinic**
- ✅ Assign/remove roles **within their clinic context**
- ✅ Manage doctors **within their clinic**
- ✅ Manage staff **within their clinic**
- ✅ Manage patients **within their clinic**
- ✅ View (but not create/modify) roles
- ❌ CANNOT access users from other clinics
- ❌ CANNOT access organization-wide data
- ❌ CANNOT block/unblock users
- ❌ CANNOT delete users
- ❌ CANNOT create or modify roles
- ❌ CANNOT change user passwords (except their own)

**API Base Path:** `/api/v1/auth/clinic-admin/*`

**Scope:** Limited to their assigned clinic(s)

**Scoping Mechanism:**
- User must have `clinic_admin` role
- Role assignment must have `clinic_id` set
- All queries automatically filter by the admin's clinic_id(s)

**Example Use Cases:**
- Single clinic manager
- Clinic receptionist with admin privileges
- Department head managing departmental staff

---

## API Endpoint Comparison

### User Management Endpoints

| Operation | Super Admin | Org Admin | Clinic Admin |
|-----------|-------------|-----------|--------------|
| List Users | ✅ All users | ✅ Org scope | ✅ Clinic scope |
| Get User | ✅ Any user | ✅ Org users only | ✅ Clinic users only |
| Create User | ✅ Anywhere | ✅ In org | ✅ In clinic |
| Update User | ✅ Any user | ✅ Org users only | ✅ Clinic users only |
| Delete User | ✅ Any user | ❌ No | ❌ No |
| Block User | ✅ Any user | ❌ No | ❌ No |
| Unblock User | ✅ Any user | ❌ No | ❌ No |
| Activate User | ✅ Any user | ✅ Org users only | ✅ Clinic users only |
| Deactivate User | ✅ Any user | ✅ Org users only | ✅ Clinic users only |
| Change Password | ✅ Any user | ❌ No | ❌ No |
| Assign Role | ✅ Any scope | ✅ Org scope | ✅ Clinic scope |
| Remove Role | ✅ Any scope | ✅ Org scope | ✅ Clinic scope |
| View Activity Logs | ✅ Any user | ❌ No | ❌ No |

### Role Management Endpoints

| Operation | Super Admin | Org Admin | Clinic Admin |
|-----------|-------------|-----------|--------------|
| List Roles | ✅ All roles | ✅ View only | ✅ View only |
| Get Role | ✅ Any role | ✅ View only | ✅ View only |
| Create Role | ✅ Yes | ❌ No | ❌ No |
| Update Role | ✅ Custom roles | ❌ No | ❌ No |
| Delete Role | ✅ Custom roles | ❌ No | ❌ No |
| Modify Permissions | ✅ Custom roles | ❌ No | ❌ No |
| View Role Users | ✅ All | ❌ No | ❌ No |

---

## API Paths and Endpoints

### Super Admin Endpoints
**Base:** `/api/v1/auth/admin`

```
GET    /admin/users                          - List all users (platform-wide)
GET    /admin/users/:id                      - Get any user
POST   /admin/users                          - Create user anywhere
PUT    /admin/users/:id                      - Update any user
DELETE /admin/users/:id                      - Delete any user
POST   /admin/users/:id/block                - Block any user
POST   /admin/users/:id/unblock              - Unblock any user
POST   /admin/users/:id/activate             - Activate any user
POST   /admin/users/:id/deactivate           - Deactivate any user
POST   /admin/users/:id/change-password      - Change any user's password
POST   /admin/users/:id/roles                - Assign role (any scope)
DELETE /admin/users/:id/roles/:role_id       - Remove role (any scope)
GET    /admin/users/:id/activity-logs        - View user activity logs

GET    /admin/roles                          - List all roles
GET    /admin/roles/:id                      - Get role details
POST   /admin/roles                          - Create custom role
PUT    /admin/roles/:id                      - Update role
DELETE /admin/roles/:id                      - Delete role
POST   /admin/roles/:id/activate             - Activate role
POST   /admin/roles/:id/deactivate           - Deactivate role
PUT    /admin/roles/:id/permissions          - Update role permissions
GET    /admin/roles/:id/users                - Get users with role
GET    /admin/permission-templates           - Get permission templates
```

### Organization Admin Endpoints
**Base:** `/api/v1/auth/org-admin`

```
GET    /org-admin/users                      - List users (org scope)
GET    /org-admin/users/:id                  - Get user (if in org)
POST   /org-admin/users                      - Create user in org
PUT    /org-admin/users/:id                  - Update user (if in org)
POST   /org-admin/users/:id/activate         - Activate user (if in org)
POST   /org-admin/users/:id/deactivate       - Deactivate user (if in org)
POST   /org-admin/users/:id/roles            - Assign role (org scope)
DELETE /org-admin/users/:id/roles/:role_id   - Remove role (org scope)

GET    /org-admin/roles                      - View all roles
GET    /org-admin/roles/:id                  - View role details
```

### Clinic Admin Endpoints
**Base:** `/api/v1/auth/clinic-admin`

```
GET    /clinic-admin/users                   - List users (clinic scope)
GET    /clinic-admin/users/:id               - Get user (if in clinic)
POST   /clinic-admin/users                   - Create user in clinic
PUT    /clinic-admin/users/:id               - Update user (if in clinic)
POST   /clinic-admin/users/:id/activate      - Activate user (if in clinic)
POST   /clinic-admin/users/:id/deactivate    - Deactivate user (if in clinic)
POST   /clinic-admin/users/:id/roles         - Assign role (clinic scope)
DELETE /clinic-admin/users/:id/roles/:role_id- Remove role (clinic scope)

GET    /clinic-admin/roles                   - View all roles
GET    /clinic-admin/roles/:id               - View role details
```

---

## How Scoping Works

### 1. Authentication Flow

```
1. User logs in → receives JWT token
2. JWT contains user_id
3. Middleware validates JWT
4. Middleware checks user's roles:
   - Is super_admin? → Full access, no restrictions
   - Is organization_admin? → Get organization_id(s) from user_roles
   - Is clinic_admin? → Get clinic_id(s) from user_roles
5. Context (org/clinic IDs) stored in request context
6. All queries automatically filter by this context
```

### 2. Database Scoping

**For Organization Admin:**
```sql
-- Automatically added to all user queries
WHERE u.id IN (
    SELECT DISTINCT ur.user_id 
    FROM user_roles ur 
    WHERE ur.organization_id IN ('admin_org_id_1', 'admin_org_id_2')
    AND ur.is_active = true
)
```

**For Clinic Admin:**
```sql
-- Automatically added to all user queries
WHERE u.id IN (
    SELECT DISTINCT ur.user_id 
    FROM user_roles ur 
    WHERE ur.clinic_id IN ('admin_clinic_id_1', 'admin_clinic_id_2')
    AND ur.is_active = true
)
```

### 3. Multi-Scope Support

A single user can be:
- Organization Admin for **multiple organizations**
- Clinic Admin for **multiple clinics**
- Both Organization Admin AND Clinic Admin (in different contexts)

**Example:**
```json
{
  "user_id": "user-123",
  "roles": [
    {
      "role": "organization_admin",
      "organization_id": "org-abc",
      "scope": "Organization ABC"
    },
    {
      "role": "clinic_admin",
      "clinic_id": "clinic-xyz",
      "scope": "Downtown Clinic"
    }
  ]
}
```

---

## Role Assignment Rules

### Super Admin Assigning Roles

Can assign any role to any user with any scope:

```json
POST /api/v1/auth/admin/users/{user_id}/roles
{
  "role_id": "organization_admin_role_id",
  "organization_id": "org-123",
  "clinic_id": null,
  "service_id": null
}
```

### Organization Admin Assigning Roles

Can only assign roles within their organization:

```json
POST /api/v1/auth/org-admin/users/{user_id}/roles
{
  "role_id": "doctor_role_id",
  "organization_id": "their_org_id",  // Must be their org
  "clinic_id": "clinic_in_their_org",  // Optional
  "service_id": null
}
```

✅ Allowed: Assign doctor role to user in their clinic
❌ Forbidden: Assign role in different organization
❌ Forbidden: Assign super_admin or organization_admin roles

### Clinic Admin Assigning Roles

Can only assign roles within their clinic:

```json
POST /api/v1/auth/clinic-admin/users/{user_id}/roles
{
  "role_id": "receptionist_role_id",
  "organization_id": null,
  "clinic_id": "their_clinic_id",  // Must be their clinic
  "service_id": null
}
```

✅ Allowed: Assign receptionist role to user in their clinic
❌ Forbidden: Assign role in different clinic
❌ Forbidden: Assign admin roles

---

## Security Considerations

### 1. Automatic Scope Enforcement

- Scoping is **enforced at the middleware level**
- Cannot be bypassed by manipulating API parameters
- Database queries automatically include scope filters
- Even if IDs are guessed, access is denied if outside scope

### 2. Privilege Escalation Prevention

- Lower-level admins cannot assign higher-level roles
- Organization Admins cannot create other Organization Admins
- Clinic Admins cannot create Organization Admins
- Only Super Admin can assign admin-level roles

### 3. Cross-Tenant Isolation

- Organization A admin cannot see Organization B data
- Clinic 1 admin cannot see Clinic 2 data
- Enforced through database-level filtering
- Logged attempts trigger security alerts

### 4. Audit Trail

All admin actions are logged with:
- Who performed the action (admin user_id)
- What action was performed
- On which resource (user_id, role_id, etc.)
- When it occurred (timestamp)
- From where (IP address, user agent)
- Within what scope (organization_id, clinic_id)

---

## Examples and Use Cases

### Example 1: Hospital Chain (Multi-Organization)

**Scenario:** A hospital chain with 3 hospitals, each with multiple clinics

```
Platform (Super Admin: Platform Owner)
├── Hospital A (Org Admin: John)
│   ├── Clinic A1 (Clinic Admin: Alice)
│   ├── Clinic A2 (Clinic Admin: Bob)
│   └── Clinic A3 (Clinic Admin: Carol)
├── Hospital B (Org Admin: Jane)
│   ├── Clinic B1 (Clinic Admin: Dave)
│   └── Clinic B2 (Clinic Admin: Eve)
└── Hospital C (Org Admin: Jack)
    └── Clinic C1 (Clinic Admin: Frank)
```

**Access Matrix:**
- **Platform Owner:** Can manage ALL hospitals and clinics
- **John (Hospital A Org Admin):** Can only manage Hospital A and its clinics (A1, A2, A3)
- **Alice (Clinic A1 Admin):** Can only manage Clinic A1 users and staff
- **Jane (Hospital B Org Admin):** Can only manage Hospital B and its clinics (B1, B2)
- **Dave (Clinic B1 Admin):** Can only manage Clinic B1 users and staff

### Example 2: Creating a New Clinic Admin

**As Super Admin:**
```bash
# 1. Create the user
POST /api/v1/auth/admin/users
{
  "first_name": "Alice",
  "last_name": "Smith",
  "username": "alice.smith",
  "email": "alice@clinic-a1.com",
  "password": "SecurePass123"
}
# Returns: user_id = "user-alice-123"

# 2. Assign clinic_admin role with clinic scope
POST /api/v1/auth/admin/users/user-alice-123/roles
{
  "role_id": "clinic-admin-role-id",
  "clinic_id": "clinic-a1-uuid"
}
```

**As Organization Admin:**
```bash
# Can do the same, but clinic must be in their organization
POST /api/v1/auth/org-admin/users/user-alice-123/roles
{
  "role_id": "clinic-admin-role-id",
  "clinic_id": "clinic-a1-uuid"  // Must be in their org
}
```

### Example 3: Listing Users by Admin Level

**Super Admin sees ALL users:**
```bash
GET /api/v1/auth/admin/users
# Returns: All 1000+ users across all organizations and clinics
```

**Organization Admin sees only their org:**
```bash
GET /api/v1/auth/org-admin/users
# Returns: Only users in Hospital A (maybe 200 users)
```

**Clinic Admin sees only their clinic:**
```bash
GET /api/v1/auth/clinic-admin/users
# Returns: Only users in Clinic A1 (maybe 20 users)
```

---

## Testing Scoping

### Test Scenario 1: Cross-Organization Access

```bash
# As Org Admin for Hospital A, try to access Hospital B user
GET /api/v1/auth/org-admin/users/hospital-b-user-id

# Expected Result: 404 Not Found or 403 Forbidden
# Reason: User is outside their organization scope
```

### Test Scenario 2: Role Assignment Validation

```bash
# As Clinic Admin, try to assign organization_admin role
POST /api/v1/auth/clinic-admin/users/some-user-id/roles
{
  "role_id": "organization-admin-role-id"
}

# Expected Result: 403 Forbidden
# Reason: Cannot assign admin-level roles
```

### Test Scenario 3: Scope Enforcement

```bash
# As Org Admin for Org A, create user and assign to Org B
POST /api/v1/auth/org-admin/users
{
  "first_name": "Test",
  "username": "test",
  "password": "pass"
}
# Then assign to Org B
POST /api/v1/auth/org-admin/users/{new-user-id}/roles
{
  "role_id": "doctor-role-id",
  "organization_id": "org-b-uuid"  // Different org!
}

# Expected Result: 403 Forbidden
# Reason: Cannot assign roles outside their organization
```

---

## Migration from Single-Tier to Hierarchical

If you have existing admins, you need to:

1. **Identify admin level** for each user
2. **Assign proper role** (super_admin, organization_admin, clinic_admin)
3. **Set context** (organization_id or clinic_id in user_roles)

```sql
-- Example: Convert existing admin to organization admin
UPDATE user_roles 
SET role_id = (SELECT id FROM roles WHERE name = 'organization_admin'),
    organization_id = 'their-org-uuid'
WHERE user_id = 'admin-user-uuid'
AND role_id = (SELECT id FROM roles WHERE name = 'admin');
```

---

## Best Practices

1. **Principle of Least Privilege**
   - Start with the lowest necessary admin level
   - Only escalate when truly needed
   - Regularly audit and downgrade as appropriate

2. **Clear Responsibility Assignment**
   - One person = one primary admin scope
   - Document who is responsible for what
   - Use naming conventions (e.g., "Alice - Clinic A1 Admin")

3. **Regular Access Reviews**
   - Quarterly review of admin access
   - Remove access for departed staff immediately
   - Audit activity logs for unusual patterns

4. **Training and Documentation**
   - Train admins on their specific scope
   - Document what they can and cannot do
   - Provide scope-specific user guides

5. **Monitoring and Alerts**
   - Alert on failed authorization attempts
   - Monitor cross-scope access attempts
   - Review admin activity logs regularly

---

## Troubleshooting

### Problem: Organization Admin can't see any users

**Check:**
1. Is organization_id set in their user_roles?
2. Are there users assigned to that organization?
3. Is their role active (is_active = true)?

```sql
SELECT * FROM user_roles WHERE user_id = 'admin-user-id';
```

### Problem: Clinic Admin sees users from other clinics

**This should never happen!** It indicates a security issue.

**Check:**
1. Verify middleware is applied to the endpoint
2. Check database query includes clinic_id filter
3. Review activity logs for the endpoint

### Problem: Cannot assign role to user

**Possible causes:**
1. User not in your scope (org/clinic)
2. Trying to assign admin role without permission
3. Target org/clinic ID not in your scope

---

## Summary

| Feature | Super Admin | Organization Admin | Clinic Admin |
|---------|-------------|-------------------|--------------|
| **Scope** | Platform-wide | Organization-level | Clinic-level |
| **User Management** | All users | Org users only | Clinic users only |
| **Can Delete Users** | ✅ Yes | ❌ No | ❌ No |
| **Can Block Users** | ✅ Yes | ❌ No | ❌ No |
| **Can Create Roles** | ✅ Yes | ❌ No | ❌ No |
| **Can Assign Admin Roles** | ✅ Yes | ❌ No | ❌ No |
| **View Activity Logs** | ✅ All | ❌ No | ❌ No |
| **Base API Path** | `/admin/*` | `/org-admin/*` | `/clinic-admin/*` |

---

**Last Updated:** October 7, 2025  
**Version:** 1.0.0  
**Maintainer:** Dr&Me Platform Team

