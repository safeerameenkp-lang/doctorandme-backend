# Security Fixes Implementation Guide

## Critical Fixes Required

This document provides the exact code changes needed to fix the critical security issues identified in the audit.

---

## Fix #1: Add Scope Validation Helper Functions

**File:** `services/auth-service/controllers/user_management.controller.go`

Add these helper functions before the existing functions:

```go
// validateUserInScope checks if a user is within the admin's scope
func validateUserInScope(userID, adminID string, isSuperAdmin, isOrgAdmin, isClinicAdmin bool, c *gin.Context) bool {
	if isSuperAdmin {
		return true
	}
	
	if isOrgAdmin {
		orgIDs, _ := c.Get("organization_ids")
		if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
			placeholders := []string{}
			args := []interface{}{userID}
			for i, orgID := range orgIDList {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
				args = append(args, orgID)
			}
			
			query := fmt.Sprintf(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles 
					WHERE user_id = $1 
					AND organization_id IN (%s) 
					AND is_active = true
				)
			`, strings.Join(placeholders, ","))
			
			var exists bool
			config.DB.QueryRow(query, args...).Scan(&exists)
			return exists
		}
	} else if isClinicAdmin {
		clinicIDs, _ := c.Get("clinic_ids")
		if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
			placeholders := []string{}
			args := []interface{}{userID}
			for i, clinicID := range clinicIDList {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
				args = append(args, clinicID)
			}
			
			query := fmt.Sprintf(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles 
					WHERE user_id = $1 
					AND clinic_id IN (%s) 
					AND is_active = true
				)
			`, strings.Join(placeholders, ","))
			
			var exists bool
			config.DB.QueryRow(query, args...).Scan(&exists)
			return exists
		}
	}
	
	return false
}

// validateRoleAssignmentScope checks if admin can assign role with given context
func validateRoleAssignmentScope(roleName string, input AssignRoleInput, isSuperAdmin, isOrgAdmin, isClinicAdmin bool, c *gin.Context) error {
	// Check privilege escalation
	adminRoles := []string{"super_admin", "organization_admin", "clinic_admin"}
	isAdminRole := false
	for _, ar := range adminRoles {
		if roleName == ar {
			isAdminRole = true
			break
		}
	}
	
	if isAdminRole && !isSuperAdmin {
		return fmt.Errorf("only Super Admin can assign admin-level roles")
	}
	
	// Validate scope
	if !isSuperAdmin {
		if isOrgAdmin {
			if input.OrganizationID == nil {
				return fmt.Errorf("organization_id required for org admin role assignment")
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
					return fmt.Errorf("cannot assign role in organization outside your scope")
				}
			} else {
				return fmt.Errorf("no organization access")
			}
		} else if isClinicAdmin {
			if input.ClinicID == nil {
				return fmt.Errorf("clinic_id required for clinic admin role assignment")
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
					return fmt.Errorf("cannot assign role in clinic outside your scope")
				}
			} else {
				return fmt.Errorf("no clinic access")
			}
		}
	}
	
	return nil
}
```

---

## Fix #2: Update UpdateUser Function

**File:** `services/auth-service/controllers/user_management.controller.go`

Replace the UpdateUser function with this secured version:

```go
// UpdateUser updates a user's information
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User not found or outside your scope",
		})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		security.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate email format if provided
	if input.Email != nil && *input.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*input.Email) {
			security.SendValidationError(c, "Invalid email format", "Please provide a valid email address")
			return
		}

		// Check if email already exists for another user
		var exists bool
		err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`, 
			*input.Email, userID).Scan(&exists)
		if err != nil {
			security.SendDatabaseError(c, "Failed to check email")
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
	}

	// Validate phone format if provided
	if input.Phone != nil && *input.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(*input.Phone) {
			security.SendValidationError(c, "Invalid phone format", "Please provide a valid phone number")
			return
		}
	}

	// Build dynamic update query
	query := "UPDATE users SET updated_by = $1"
	args := []interface{}{adminID}
	argIndex := 2

	if input.FirstName != nil {
		query += fmt.Sprintf(", first_name = $%d", argIndex)
		args = append(args, *input.FirstName)
		argIndex++
	}
	if input.LastName != nil {
		query += fmt.Sprintf(", last_name = $%d", argIndex)
		args = append(args, *input.LastName)
		argIndex++
	}
	if input.Email != nil {
		query += fmt.Sprintf(", email = $%d", argIndex)
		args = append(args, *input.Email)
		argIndex++
	}
	if input.Phone != nil {
		query += fmt.Sprintf(", phone = $%d", argIndex)
		args = append(args, *input.Phone)
		argIndex++
	}
	if input.DateOfBirth != nil {
		query += fmt.Sprintf(", date_of_birth = $%d", argIndex)
		args = append(args, *input.DateOfBirth)
		argIndex++
	}
	if input.Gender != nil {
		query += fmt.Sprintf(", gender = $%d", argIndex)
		args = append(args, *input.Gender)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, userID)

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		security.SendDatabaseError(c, "Failed to update user")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		security.SendNotFoundError(c, "User")
		return
	}

	// Log activity
	logUserActivity(adminID, "UPDATE_USER", fmt.Sprintf("Updated user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
```

---

## Fix #3: Update DeleteUser Function

Add scope validation:

```go
// DeleteUser deletes a user (soft delete by marking as inactive)
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	// Prevent super admin from deleting themselves
	if userID == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		security.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		security.SendNotFoundError(c, "User")
		return
	}

	// Soft delete: mark as inactive and blocked
	_, err = config.DB.Exec(`
		UPDATE users 
		SET is_active = false, is_blocked = true, blocked_at = CURRENT_TIMESTAMP, 
		    blocked_by = $1, blocked_reason = 'Account deleted by administrator', updated_by = $1
		WHERE id = $2
	`, adminID, userID)

	if err != nil {
		security.SendDatabaseError(c, "Failed to delete user")
		return
	}

	// Revoke all active refresh tokens
	_, err = config.DB.Exec(`
		UPDATE refresh_tokens 
		SET revoked_at = CURRENT_TIMESTAMP 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)

	// Log activity
	logUserActivity(adminID, "DELETE_USER", fmt.Sprintf("Deleted user %s", userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
```

---

## Fix #4: Update AssignRole Function

Replace with this secured version:

```go
// AssignRole assigns a role to a user
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

	// ✅ SECURITY: Validate user is in admin's scope
	if !validateUserInScope(userID, adminID, isSuperAdmin, isOrgAdmin, isClinicAdmin, c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User not found or outside your scope",
		})
		return
	}

	// Check if user exists
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		security.SendDatabaseError(c, "Failed to check user")
		return
	}
	if !exists {
		security.SendNotFoundError(c, "User")
		return
	}

	// ✅ SECURITY: Get role details including name
	var roleName string
	err = config.DB.QueryRow(`
		SELECT name FROM roles 
		WHERE id = $1 AND is_active = true
	`, input.RoleID).Scan(&roleName)
	
	if err != nil {
		if err == sql.ErrNoRows {
			security.SendNotFoundError(c, "Role")
			return
		}
		security.SendDatabaseError(c, "Failed to check role")
		return
	}

	// ✅ SECURITY: Validate role assignment scope and prevent privilege escalation
	if err := validateRoleAssignmentScope(roleName, input, isSuperAdmin, isOrgAdmin, isClinicAdmin, c); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Assign role
	_, err = config.DB.Exec(`
		INSERT INTO user_roles (user_id, role_id, organization_id, clinic_id, service_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, role_id, organization_id, clinic_id, service_id) 
		DO UPDATE SET is_active = true
	`, userID, input.RoleID, input.OrganizationID, input.ClinicID, input.ServiceID)

	if err != nil {
		security.SendDatabaseError(c, "Failed to assign role")
		return
	}

	// Log activity
	logUserActivity(adminID, "ASSIGN_ROLE", 
		fmt.Sprintf("Assigned role %s to user %s", roleName, userID), c)

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}
```

---

## Fix #5: Update Login to Check is_blocked

**File:** `services/auth-service/controllers/auth.controller.go`

Replace the Login function with this:

```go
func Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Fetch user - ✅ SECURITY: Check both is_active and is_blocked
    var user models.User
    var isBlocked bool
    err := config.DB.QueryRow(`
        SELECT id, password_hash, first_name, last_name, email, username, phone, is_blocked
        FROM users
        WHERE (email = $1 OR phone = $1 OR username = $1) 
        AND is_active = true
        AND is_blocked = false
    `, input.Login).Scan(&user.ID, &user.PasswordHash, &user.FirstName, &user.LastName, 
        &user.Email, &user.Username, &user.Phone, &isBlocked)
    
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or account blocked"})
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Update last login
    _, err = config.DB.Exec(`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1`, user.ID)
    if err != nil {
        c.Header("X-Warning", "Failed to update last login timestamp")
    }

    // Generate tokens
    accessToken, err := security.SignAccessToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
        return
    }

    refreshToken, err := security.SignRefreshToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
        return
    }

    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    _, err = config.DB.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`, 
        user.ID, refreshToken, expiresAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
        return
    }

    // Fetch user roles with organization/clinic context
    rows, err := config.DB.Query(`
        SELECT r.id, r.name, r.permissions, ur.organization_id, ur.clinic_id, ur.service_id
        FROM roles r
        JOIN user_roles ur ON ur.role_id = r.id
        WHERE ur.user_id = $1 AND ur.is_active = true
    `, user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles"})
        return
    }
    defer rows.Close()

    var roles []map[string]interface{}
    for rows.Next() {
        var roleID, roleName string
        var permissionsJSON []byte
        var orgID, clinicID, serviceID *string
        
        err = rows.Scan(&roleID, &roleName, &permissionsJSON, &orgID, &clinicID, &serviceID)
        if err != nil {
            continue
        }

        var permissions map[string]interface{}
        err = json.Unmarshal(permissionsJSON, &permissions)
        if err != nil {
            permissions = make(map[string]interface{})
        }

        role := map[string]interface{}{
            "id":          roleID,
            "name":        roleName,
            "permissions": permissions,
        }
        
        if orgID != nil {
            role["organization_id"] = *orgID
        }
        if clinicID != nil {
            role["clinic_id"] = *clinicID
        }
        if serviceID != nil {
            role["service_id"] = *serviceID
        }

        roles = append(roles, role)
    }

    c.JSON(http.StatusOK, gin.H{
        "id":           user.ID,
        "firstName":    user.FirstName,
        "lastName":     user.LastName,
        "email":        user.Email,
        "username":     user.Username,
        "phone":        user.Phone,
        "roles":        roles,
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
        "tokenType":    "Bearer",
        "expiresIn":    3600,
    })
}
```

---

## Testing the Fixes

### Test 1: Org Admin Cannot Update User Outside Scope

```bash
# Login as Org Admin for Organization A
TOKEN_ORG_A=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"orgadmin_a","password":"pass"}' | jq -r '.accessToken')

# Try to update user from Organization B
curl -X PUT http://localhost:8000/api/v1/auth/org-admin/users/ORG_B_USER_ID \
  -H "Authorization: Bearer $TOKEN_ORG_A" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Hacked"}'

# Expected: 403 Forbidden - "User not found or outside your scope"
```

### Test 2: Clinic Admin Cannot Assign Admin Roles

```bash
# Login as Clinic Admin
TOKEN_CLINIC=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"clinicadmin","password":"pass"}' | jq -r '.accessToken')

# Try to assign super_admin role
curl -X POST http://localhost:8000/api/v1/auth/clinic-admin/users/USER_ID/roles \
  -H "Authorization: Bearer $TOKEN_CLINIC" \
  -H "Content-Type: application/json" \
  -d '{"role_id":"SUPER_ADMIN_ROLE_ID","clinic_id":"THEIR_CLINIC_ID"}'

# Expected: 403 Forbidden - "only Super Admin can assign admin-level roles"
```

### Test 3: Blocked User Cannot Login

```bash
# Block user as Super Admin
curl -X POST http://localhost:8000/api/v1/auth/admin/users/USER_ID/block \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason":"Test blocking"}'

# Try to login as blocked user
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"blockeduser","password":"correctpassword"}'

# Expected: 401 Unauthorized - "Invalid credentials or account blocked"
```

---

## Deployment Checklist

- [ ] Backup database before applying fixes
- [ ] Apply fixes to user_management.controller.go
- [ ] Apply fixes to auth.controller.go
- [ ] Rebuild services
- [ ] Run test suite
- [ ] Test scope validation with real users
- [ ] Test privilege escalation prevention
- [ ] Test blocked user login prevention
- [ ] Monitor logs for any issues
- [ ] Document changes in CHANGELOG.md

---

## Rollback Plan

If issues occur after deployment:

1. **Immediate:** Revert to previous Docker image
   ```bash
   docker-compose down
   docker-compose pull auth-service:previous-tag
   docker-compose up -d
   ```

2. **Database:** No schema changes, so no rollback needed

3. **Monitoring:** Check these logs for errors:
   - Failed scope validations (might be false positives)
   - User lockouts (blocked users trying to login)
   - Role assignment failures

---

**Implementation Priority:** CRITICAL - Deploy within 48 hours  
**Estimated Implementation Time:** 2-3 hours  
**Testing Time:** 2-3 hours  
**Total:** 1 business day

