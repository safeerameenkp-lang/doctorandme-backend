# RBAC Security Audit Report

## Executive Summary

**Audit Date:** October 7, 2025  
**System:** Dr&Me Healthcare RBAC System  
**Controllers Audited:** 
- `user_management.controller.go`
- `role_management.controller.go`
- `scoped_resources.controller.go`

**Overall Security Rating:** ⚠️ **GOOD with Critical Gaps** (7/10)

---

## Critical Security Issues Found

### 🔴 CRITICAL ISSUE #1: Missing Scope Validation in User Operations

**Location:** `user_management.controller.go` - CreateUser, UpdateUser, DeleteUser, AssignRole

**Problem:**
The functions don't validate if Org/Clinic Admins are operating within their scope when creating or modifying users.

**Current Code (CreateUser):**
```go
func CreateUser(c *gin.Context) {
    adminID := c.GetString("user_id")
    // ... input validation ...
    
    // ❌ NO CHECK if Org Admin is creating user in THEIR organization
    // ❌ NO CHECK if Clinic Admin is creating user in THEIR clinic
    
    err = config.DB.QueryRow(`
        INSERT INTO users (first_name, last_name, email, username, phone, password_hash, 
                           date_of_birth, gender, is_active, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
        RETURNING id
    `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, 
       string(passHash), input.DateOfBirth, input.Gender, isActive, adminID).Scan(&userID)
}
```

**Risk:** Org/Clinic Admins could potentially create users system-wide if middleware fails.

**Recommended Fix:**
```go
func CreateUser(c *gin.Context) {
    adminID := c.GetString("user_id")
    isSuperAdmin := c.GetBool("is_super_admin")
    isOrgAdmin := c.GetBool("is_organization_admin")
    isClinicAdmin := c.GetBool("is_clinic_admin")
    
    var input CreateUserInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }
    
    // ✅ VALIDATE SCOPE
    if !isSuperAdmin {
        if isOrgAdmin {
            // Validate organization_id is in their scope
            if len(input.RoleIDs) > 0 {
                for _, roleID := range input.RoleIDs {
                    // Check if role assignment includes organization_id
                    // and that organization_id is in admin's org list
                }
            }
        } else if isClinicAdmin {
            // Validate clinic_id is in their scope
            if len(input.RoleIDs) > 0 {
                for _, roleID := range input.RoleIDs {
                    // Check if role assignment includes clinic_id
                    // and that clinic_id is in admin's clinic list
                }
            }
        } else {
            // Regular users can't create other users
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            return
        }
    }
    
    // ... rest of function
}
```

---

### 🔴 CRITICAL ISSUE #2: Missing Privilege Escalation Prevention in AssignRole

**Location:** `user_management.controller.go` - AssignRole()

**Problem:**
No validation prevents Org/Clinic Admins from assigning admin-level roles.

**Current Code:**
```go
func AssignRole(c *gin.Context) {
    userID := c.Param("id")
    adminID := c.GetString("user_id")

    var input AssignRoleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Check if user exists
    var exists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
    // ...
    
    // Check if role exists
    err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1 AND is_active = true)`, 
        input.RoleID).Scan(&exists)
    // ...
    
    // ❌ NO CHECK: Is the role being assigned an admin role?
    // ❌ NO CHECK: Does the admin have permission to assign this role?
    // ❌ NO CHECK: Is the org/clinic context valid for this admin?
    
    _, err = config.DB.Exec(`
        INSERT INTO user_roles (user_id, role_id, organization_id, clinic_id, service_id)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (user_id, role_id, organization_id, clinic_id, service_id) 
        DO UPDATE SET is_active = true
    `, userID, input.RoleID, input.OrganizationID, input.ClinicID, input.ServiceID)
}
```

**Risk:** 
- Org Admin could assign `super_admin` role
- Clinic Admin could assign `organization_admin` role
- Privilege escalation attack vector

**Recommended Fix:**
```go
func AssignRole(c *gin.Context) {
    userID := c.Param("id")
    adminID := c.GetString("user_id")
    isSuperAdmin := c.GetBool("is_super_admin")
    isOrgAdmin := c.GetBool("is_organization_admin")
    isClinicAdmin := c.GetBool("is_clinic_admin")

    var input AssignRoleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }
    
    // ✅ GET ROLE DETAILS INCLUDING NAME
    var roleName string
    var roleExists bool
    err := config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1 AND is_active = true),
               name
        FROM roles 
        WHERE id = $1
    `, input.RoleID).Scan(&roleExists, &roleName)
    
    if err != nil || !roleExists {
        security.SendNotFoundError(c, "Role")
        return
    }
    
    // ✅ PREVENT PRIVILEGE ESCALATION
    adminRoles := []string{"super_admin", "organization_admin", "clinic_admin"}
    isAdminRole := false
    for _, ar := range adminRoles {
        if roleName == ar {
            isAdminRole = true
            break
        }
    }
    
    if isAdminRole && !isSuperAdmin {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Cannot assign admin roles",
            "message": "Only Super Admin can assign admin-level roles (super_admin, organization_admin, clinic_admin)",
        })
        return
    }
    
    // ✅ VALIDATE SCOPE
    if !isSuperAdmin {
        if isOrgAdmin {
            // Validate organization_id matches their org
            if input.OrganizationID == nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "organization_id required for org admin role assignment",
                })
                return
            }
            
            orgIDs, _ := c.Get("organization_ids")
            if orgIDList, ok := orgIDs.([]string); ok {
                found := false
                for _, oid := range orgIDList {
                    if oid == *input.OrganizationID {
                        found = true
                        break
                    }
                }
                if !found {
                    c.JSON(http.StatusForbidden, gin.H{
                        "error": "Cannot assign role in organization outside your scope",
                    })
                    return
                }
            }
        } else if isClinicAdmin {
            // Validate clinic_id matches their clinic
            if input.ClinicID == nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "clinic_id required for clinic admin role assignment",
                })
                return
            }
            
            clinicIDs, _ := c.Get("clinic_ids")
            if clinicIDList, ok := clinicIDs.([]string); ok {
                found := false
                for _, cid := range clinicIDList {
                    if cid == *input.ClinicID {
                        found = true
                        break
                    }
                }
                if !found {
                    c.JSON(http.StatusForbidden, gin.H{
                        "error": "Cannot assign role in clinic outside your scope",
                    })
                    return
                }
            }
        }
    }
    
    // Check if user exists and is in scope
    if !isSuperAdmin {
        // Verify target user is in admin's scope
        userInScope := false
        if isOrgAdmin {
            orgIDs, _ := c.Get("organization_ids")
            // Check if user has any role in these organizations
            // ... validation logic
        } else if isClinicAdmin {
            clinicIDs, _ := c.Get("clinic_ids")
            // Check if user has any role in these clinics
            // ... validation logic
        }
        
        if !userInScope {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Cannot assign role to user outside your scope",
            })
            return
        }
    }
    
    // ... rest of function
}
```

---

### 🔴 CRITICAL ISSUE #3: Missing Scope Validation in UpdateUser and DeleteUser

**Problem:** Same as Issue #1 - these operations don't validate if the target user is within the admin's scope.

**Recommended Fix:**
Add scope validation at the beginning of each function:

```go
func UpdateUser(c *gin.Context) {
    userID := c.Param("id")
    adminID := c.GetString("user_id")
    isSuperAdmin := c.GetBool("is_super_admin")
    isOrgAdmin := c.GetBool("is_organization_admin")
    isClinicAdmin := c.GetBool("is_clinic_admin")
    
    // ✅ VALIDATE USER IS IN ADMIN'S SCOPE
    if !isSuperAdmin {
        userInScope := false
        
        if isOrgAdmin {
            orgIDs, _ := c.Get("organization_ids")
            if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
                // Check if user has roles in any of these organizations
                var count int
                placeholders := []string{}
                args := []interface{}{userID}
                for i, orgID := range orgIDList {
                    placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
                    args = append(args, orgID)
                }
                
                query := fmt.Sprintf(`
                    SELECT COUNT(*) FROM user_roles 
                    WHERE user_id = $1 
                    AND organization_id IN (%s) 
                    AND is_active = true
                `, strings.Join(placeholders, ","))
                
                config.DB.QueryRow(query, args...).Scan(&count)
                userInScope = count > 0
            }
        } else if isClinicAdmin {
            clinicIDs, _ := c.Get("clinic_ids")
            if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
                // Check if user has roles in any of these clinics
                var count int
                placeholders := []string{}
                args := []interface{}{userID}
                for i, clinicID := range clinicIDList {
                    placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
                    args = append(args, clinicID)
                }
                
                query := fmt.Sprintf(`
                    SELECT COUNT(*) FROM user_roles 
                    WHERE user_id = $1 
                    AND clinic_id IN (%s) 
                    AND is_active = true
                `, strings.Join(placeholders, ","))
                
                config.DB.QueryRow(query, args...).Scan(&count)
                userInScope = count > 0
            }
        }
        
        if !userInScope {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "User not found or outside your scope",
            })
            return
        }
    }
    
    // ... rest of function
}
```

---

## ⚠️ HIGH Priority Issues

### Issue #4: Blocked User Can Still Login

**Location:** `auth.controller.go` - Login()

**Problem:** Login only checks `is_active` but not `is_blocked`

**Current Code:**
```go
err := config.DB.QueryRow(`
    SELECT id, password_hash, first_name, last_name, email, username, phone
    FROM users
    WHERE (email = $1 OR phone = $1 OR username = $1) AND is_active = true
`, input.Login).Scan(...)
```

**Risk:** Blocked users can still login

**Fix:**
```go
err := config.DB.QueryRow(`
    SELECT id, password_hash, first_name, last_name, email, username, phone, is_blocked
    FROM users
    WHERE (email = $1 OR phone = $1 OR username = $1) 
    AND is_active = true 
    AND is_blocked = false  // ✅ ADD THIS
`, input.Login).Scan(...)
```

---

### Issue #5: Missing Rate Limiting

**Location:** All controllers

**Problem:** No rate limiting on authentication endpoints

**Recommendation:**
```go
// Add rate limiting middleware
import "github.com/gin-contrib/limiter"

func SetupRateLimiting(r *gin.Engine) {
    // Login endpoint: 5 attempts per minute
    loginLimiter := limiter.New(limiter.Config{
        Rate:  time.Minute,
        Limit: 5,
    })
    
    r.POST("/login", loginLimiter.Limit(), controllers.Login)
    
    // Admin endpoints: 100 requests per minute
    adminLimiter := limiter.New(limiter.Config{
        Rate:  time.Minute,
        Limit: 100,
    })
    
    admin := r.Group("/admin")
    admin.Use(adminLimiter.Limit())
}
```

---

### Issue #6: SQL Injection Risk in Dynamic Query Building

**Location:** `user_management.controller.go` - ScopedListUsers

**Problem:** Dynamic SQL with string concatenation

**Current Code:**
```go
query := fmt.Sprintf(`
    SELECT ...
    FROM users u
    %s
    ORDER BY u.%s %s
    LIMIT $%d OFFSET $%d
`, whereClause, input.SortBy, input.SortOrder, argIndex, argIndex+1)
```

**Risk:** If `input.SortBy` or `input.SortOrder` aren't properly validated, SQL injection possible

**Current Mitigation:** ✅ Validation exists:
```go
validSortFields := map[string]bool{
    "created_at": true, "updated_at": true, "first_name": true,
    "last_name": true, "email": true, "username": true, "last_login": true,
}
if !validSortFields[input.SortBy] {
    security.SendValidationError(c, "Invalid sort field", "Invalid sort_by field")
    return
}
```

**Status:** ✅ PROTECTED (Good validation)

---

## 🟡 Medium Priority Issues

### Issue #7: Missing Input Sanitization

**Problem:** No HTML/XSS sanitization on user inputs

**Recommendation:**
```go
import "html"

func sanitizeInput(input string) string {
    return html.EscapeString(strings.TrimSpace(input))
}

// Usage in CreateUser:
input.FirstName = sanitizeInput(input.FirstName)
input.LastName = sanitizeInput(input.LastName)
```

---

### Issue #8: Weak Password Policy

**Location:** All password change functions

**Current:** Minimum 8 characters only

**Recommendation:**
```go
func validatePasswordStrength(password string) error {
    if len(password) < 12 {
        return errors.New("Password must be at least 12 characters")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-Z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
    
    if !(hasUpper && hasLower && hasNumber && hasSpecial) {
        return errors.New("Password must contain uppercase, lowercase, number, and special character")
    }
    
    return nil
}
```

---

### Issue #9: Missing CSRF Protection

**Problem:** No CSRF tokens for state-changing operations

**Recommendation:**
```go
import "github.com/gin-contrib/csrf"

func SetupCSRF(r *gin.Engine) {
    r.Use(csrf.Middleware(csrf.Options{
        Secret: os.Getenv("CSRF_SECRET"),
        ErrorFunc: func(c *gin.Context) {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "CSRF token validation failed",
            })
            c.Abort()
        },
    }))
}
```

---

### Issue #10: Missing Request ID Tracking

**Problem:** Hard to trace requests across logs

**Recommendation:**
```go
import "github.com/google/uuid"

func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// Update logUserActivity to include request_id
func logUserActivity(performedBy, actionType, description string, c *gin.Context) {
    requestID := c.GetString("request_id")
    // ... include requestID in metadata
}
```

---

## 🟢 Missing Endpoints Identified

### Category: User Management

1. ✅ **Bulk User Import** (High Priority for SaaS)
   ```
   POST /admin/users/bulk-import
   - Import users from CSV
   - Validate and create in batch
   - Return success/failure report
   ```

2. ✅ **Bulk User Export** (High Priority)
   ```
   GET /admin/users/export?format=csv
   - Export users to CSV/Excel
   - Filter by scope
   - Include roles and status
   ```

3. ✅ **User Session Management** (High Priority)
   ```
   GET /admin/users/:id/sessions
   POST /admin/users/:id/sessions/revoke-all
   DELETE /admin/users/:id/sessions/:session_id
   ```

4. ✅ **Password Reset (Admin-initiated)** (Medium Priority)
   ```
   POST /admin/users/:id/password-reset
   - Generate reset token
   - Send via email
   - Expires in 24 hours
   ```

5. ✅ **User Impersonation** (Medium Priority for Support)
   ```
   POST /admin/users/:id/impersonate
   - Super Admin only
   - Audit logged
   - Limited duration
   ```

### Category: Role Management

6. ✅ **Bulk Role Assignment** (High Priority)
   ```
   POST /admin/roles/bulk-assign
   {
     "user_ids": ["id1", "id2"],
     "role_id": "role-id",
     "organization_id": "org-id"
   }
   ```

7. ✅ **Role Usage Analytics** (Medium Priority)
   ```
   GET /admin/roles/:id/analytics
   - Users with role over time
   - Permission usage stats
   - Audit trail summary
   ```

### Category: Organization/Clinic Management

8. ✅ **Organization Statistics** (High Priority)
   ```
   GET /admin/organizations/:id/statistics
   - User count
   - Clinic count
   - Active/inactive breakdown
   - Growth metrics
   ```

9. ✅ **Clinic Transfer** (Medium Priority)
   ```
   POST /admin/clinics/:id/transfer
   {
     "new_organization_id": "org-id",
     "transfer_users": true,
     "transfer_patients": true
   }
   ```

### Category: Audit & Compliance

10. ✅ **Advanced Audit Logs** (High Priority)
    ```
    GET /admin/audit-logs
    - Filter by date range
    - Filter by action type
    - Filter by user/admin
    - Export capability
    ```

11. ✅ **Compliance Reports** (High Priority for Healthcare)
    ```
    GET /admin/reports/compliance
    - HIPAA access logs
    - User access patterns
    - Data modification history
    - Failed login attempts
    ```

12. ✅ **Data Access Report** (HIPAA Required)
    ```
    GET /admin/reports/data-access
    - Who accessed patient data
    - When and from where
    - What actions were performed
    ```

### Category: Security

13. ✅ **Two-Factor Authentication Management** (High Priority)
    ```
    POST /admin/users/:id/2fa/enable
    POST /admin/users/:id/2fa/disable
    POST /admin/users/:id/2fa/reset
    ```

14. ✅ **Security Events** (High Priority)
    ```
    GET /admin/security/events
    - Failed login attempts
    - Password changes
    - Role escalations
    - Unusual access patterns
    ```

15. ✅ **IP Whitelist Management** (Medium Priority)
    ```
    GET /admin/security/ip-whitelist
    POST /admin/security/ip-whitelist
    DELETE /admin/security/ip-whitelist/:id
    ```

---

## Recommendations Summary

### Immediate Actions (Critical)

1. **Add Scope Validation** to CreateUser, UpdateUser, DeleteUser
2. **Add Privilege Escalation Prevention** to AssignRole
3. **Add is_blocked Check** to Login
4. **Implement Rate Limiting** on all endpoints

### Short-term (Within 1 Week)

5. Add missing endpoints for user session management
6. Implement audit log export functionality
7. Add bulk user operations (import/export)
8. Strengthen password policy
9. Add input sanitization

### Medium-term (Within 1 Month)

10. Implement 2FA
11. Add CSRF protection
12. Implement request ID tracking
13. Add compliance reporting endpoints
14. Implement user impersonation for support

### Long-term (Within 3 Months)

15. Add advanced analytics
16. Implement IP whitelisting
17. Add security event monitoring
18. Implement automated compliance reports

---

## Security Score Breakdown

| Category | Score | Status |
|----------|-------|--------|
| Authentication | 7/10 | ⚠️ Good (missing 2FA, rate limiting) |
| Authorization | 6/10 | ⚠️ Needs Work (scope validation gaps) |
| Input Validation | 8/10 | ✅ Good (SQL injection protected) |
| Audit Trail | 9/10 | ✅ Excellent (comprehensive logging) |
| Scope Enforcement | 5/10 | 🔴 Needs Work (critical gaps) |
| Password Security | 6/10 | ⚠️ Needs Work (weak policy) |
| API Security | 7/10 | ⚠️ Good (missing CSRF, rate limiting) |

**Overall:** 7/10 - Good foundation but critical gaps need immediate attention

---

## Code Quality Assessment

### Strengths ✅

1. **Well-structured** controllers with clear separation of concerns
2. **Comprehensive** error handling
3. **Good** use of middleware for authentication
4. **Excellent** audit logging
5. **Clean** code with proper documentation
6. **Proper** use of parameterized queries (SQL injection protected)

### Areas for Improvement ⚠️

1. **Missing** scope validation in CRUD operations
2. **Inconsistent** admin role checking
3. **No** unit tests visible
4. **No** integration tests visible
5. **Limited** error context in some functions
6. **No** request ID tracking

---

## Testing Recommendations

### Unit Tests Needed

```go
// user_management_test.go
func TestCreateUser_OrgAdminCannotCreateOutsideScope(t *testing.T) {}
func TestAssignRole_PreventPrivilegeEscalation(t *testing.T) {}
func TestUpdateUser_ScopeValidation(t *testing.T) {}
```

### Integration Tests Needed

```go
func TestE2E_OrgAdminWorkflow(t *testing.T) {}
func TestE2E_CrossTenantIsolation(t *testing.T) {}
func TestE2E_PrivilegeEscalationPrevention(t *testing.T) {}
```

### Security Tests Needed

```go
func TestSecurity_SQLInjectionAttempts(t *testing.T) {}
func TestSecurity_XSSAttemptsBlocked(t *testing.T) {}
func TestSecurity_RateLimitingEnforced(t *testing.T) {}
```

---

## Conclusion

The RBAC system has a **solid foundation** with excellent audit logging and good code structure. However, there are **critical security gaps** in scope validation and privilege escalation prevention that must be addressed immediately before production deployment.

**Priority:** Address Critical Issues #1, #2, #3, and #4 within 48 hours.

---

**Audit Conducted By:** AI Security Analyst  
**Date:** October 7, 2025  
**Next Audit:** After critical fixes implemented

