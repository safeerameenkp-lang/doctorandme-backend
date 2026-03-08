package services

import (
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"organization-service/models"
	"organization-service/utils"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ClinicService struct {
	DB *sql.DB
}

func NewClinicService(db *sql.DB) *ClinicService {
	return &ClinicService{DB: db}
}

// CreateClinic handles the creation logic including image upload
func (s *ClinicService) CreateClinic(ctx context.Context, input models.CreateClinicInput, logoFile *multipart.FileHeader) (string, *string, error) {
	// Verify organization exists
	var orgExists bool

	// Create strict operations context mapped bounds
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`, input.OrganizationID).Scan(&orgExists)
	if err != nil || !orgExists {
		return "", nil, errors.New("organization not found")
	}

	// Verify user exists and context isn't halted
	var userExists bool
	err = s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, input.UserID).Scan(&userExists)
	if err != nil || !userExists {
		return "", nil, errors.New("user not found")
	}

	// Validate and format ClinicType
	input.ClinicType = strings.TrimSpace(input.ClinicType)
	if input.ClinicType == "" {
		return "", nil, errors.New("clinic_type cannot be empty")
	}
	input.ClinicType = strings.Title(strings.ToLower(input.ClinicType))

	// Handle Logo Upload (Do this outside tx to avoid locks during IO)
	var logoPath *string
	if logoFile != nil {
		if err := utils.ValidateImage(logoFile); err != nil {
			return "", nil, err
		}

		type result struct {
			path string
			err  error
		}
		resChan := make(chan result, 1)

		go func() {
			file, err := logoFile.Open()
			if err != nil {
				resChan <- result{"", err}
				return
			}
			defer file.Close()

			savedPath, err := utils.SaveOptimizedImage(file, logoFile.Filename, "clinics")
			resChan <- result{savedPath, err}
		}()

		res := <-resChan
		if res.err != nil {
			return "", nil, res.err
		}
		logoPath = &res.path
	}

	// Start transaction for atomic creation and role assignment
	tx, err := s.DB.Begin()
	if err != nil {
		return "", nil, errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// Auto-generate clinic code if missing (Inside transaction for locking)
	var clinicCode string
	if input.ClinicCode != nil && *input.ClinicCode != "" {
		clinicCode = *input.ClinicCode
	} else {
		var err error
		clinicCode, err = utils.GenerateClinicCode(opCtx, tx, s.DB, input.Name)
		if err != nil {
			return "", nil, errors.New("failed to generate clinic code")
		}
	}

	var clinicID string
	err = tx.QueryRowContext(opCtx, `
        INSERT INTO clinics (organization_id, user_id, clinic_code, name, clinic_type, email, phone, address, license_number, logo)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
    `, input.OrganizationID, input.UserID, clinicCode, input.Name, input.ClinicType, input.Email, input.Phone, input.Address, input.LicenseNumber, logoPath).Scan(&clinicID)

	if err != nil {
		return "", nil, errors.New("failed to insert clinic: " + err.Error())
	}

	// Assign clinic_admin role to the assigned user
	var roleID string
	err = tx.QueryRowContext(opCtx, `SELECT id FROM roles WHERE name='clinic_admin' LIMIT 1`).Scan(&roleID)
	if err == nil {
		// Only assign if role found (don't fail critical path if role table is weird, but usually it should be there)
		_, _ = tx.ExecContext(opCtx, `
            INSERT INTO user_roles (user_id, role_id, clinic_id)
            VALUES ($1, $2, $3)
            ON CONFLICT (user_id, role_id, clinic_id) DO NOTHING
        `, input.UserID, roleID, clinicID)
	}

	if err = tx.Commit(); err != nil {
		return "", nil, errors.New("failed to commit clinic creation")
	}

	return clinicID, logoPath, nil
}

// CreateClinicWithAdmin handles creating a clinic along with a new admin user
func (s *ClinicService) CreateClinicWithAdmin(ctx context.Context, input models.CreateClinicWithAdminInput, logoFile *multipart.FileHeader) (string, string, *string, error) {
	// Setup operations bound
	opCtx, cancel := context.WithTimeout(ctx, 15*time.Second) // Longer timeout due to hashing and file io
	defer cancel()

	// Validate admin email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.AdminEmail) {
		return "", "", nil, errors.New("invalid admin email format")
	}

	// Validate admin phone format if provided
	if input.AdminPhone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(input.AdminPhone) {
			return "", "", nil, errors.New("invalid admin phone format")
		}
	}

	// Verify organization exists
	var orgExists bool
	err := s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`, input.OrganizationID).Scan(&orgExists)
	if err != nil || !orgExists {
		return "", "", nil, errors.New("organization not found")
	}

	// Validate and format ClinicType
	input.ClinicType = strings.TrimSpace(input.ClinicType)
	if input.ClinicType == "" {
		return "", "", nil, errors.New("clinic_type cannot be empty")
	}
	input.ClinicType = strings.Title(strings.ToLower(input.ClinicType))

	// Handle Logo Upload (do this before transaction to avoid holding lock during file I/O)
	var logoPath *string
	if logoFile != nil {
		if err := utils.ValidateImage(logoFile); err != nil {
			return "", "", nil, err
		}

		type result struct {
			path string
			err  error
		}
		resChan := make(chan result, 1)

		go func() {
			file, err := logoFile.Open()
			if err != nil {
				resChan <- result{"", err}
				return
			}
			defer file.Close()

			savedPath, err := utils.SaveOptimizedImage(file, logoFile.Filename, "clinics")
			resChan <- result{savedPath, err}
		}()

		// Wait for result
		res := <-resChan
		if res.err != nil {
			return "", "", nil, res.err
		}
		logoPath = &res.path
	}

	// Start transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return "", "", nil, errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// Auto-generate clinic code if missing (Inside transaction for locking)
	var clinicCode string
	if input.ClinicCode != nil && *input.ClinicCode != "" {
		clinicCode = *input.ClinicCode
	} else {
		var err error
		clinicCode, err = utils.GenerateClinicCode(opCtx, tx, s.DB, input.Name)
		if err != nil {
			return "", "", nil, errors.New("failed to generate clinic code")
		}
	}

	// 1. Check if admin user already exists (by email) to avoid conflicts
	var adminID string
	err = tx.QueryRowContext(opCtx, `SELECT id FROM users WHERE email = $1`, input.AdminEmail).Scan(&adminID)

	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist, create new one
			passHash, err := bcrypt.GenerateFromPassword([]byte(input.AdminPassword), bcrypt.DefaultCost)
			if err != nil {
				return "", "", nil, errors.New("failed to hash admin password")
			}

			err = tx.QueryRowContext(opCtx, `
                INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
                VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
            `, input.AdminFirstName, input.AdminLastName, input.AdminEmail, input.AdminUsername, input.AdminPhone, string(passHash)).Scan(&adminID)
			if err != nil {
				return "", "", nil, errors.New("failed to create admin user: " + err.Error())
			}
		} else {
			return "", "", nil, errors.New("database error checking user: " + err.Error())
		}
	} else {
		// Existing user found, we will simply link the new clinic to them
		// (We don't update password here for security reasons, unless explicitly required)
	}

	// 2. Create clinic with admin as user_id (the owner)
	var clinicID string
	err = tx.QueryRowContext(opCtx, `
        INSERT INTO clinics (organization_id, user_id, clinic_code, name, clinic_type, email, phone, address, license_number, logo)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
    `, input.OrganizationID, adminID, clinicCode, input.Name, input.ClinicType, input.Email, input.Phone, input.Address, input.LicenseNumber, logoPath).Scan(&clinicID)
	if err != nil {
		return "", "", nil, errors.New("failed to create clinic: " + err.Error())
	}

	// Assign clinic_admin role
	var roleID string
	err = tx.QueryRowContext(opCtx, `SELECT id FROM roles WHERE name='clinic_admin' LIMIT 1`).Scan(&roleID)
	if err != nil {
		return "", "", nil, errors.New("failed to find clinic_admin role")
	}

	_, err = tx.ExecContext(opCtx, `INSERT INTO user_roles (user_id, role_id, clinic_id) VALUES ($1,$2,$3)`, adminID, roleID, clinicID)
	if err != nil {
		return "", "", nil, errors.New("failed to assign clinic admin role")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return "", "", nil, errors.New("failed to commit transaction")
	}

	return clinicID, adminID, logoPath, nil
}

// DeleteClinic handles deleting a clinic and its associated admin user
func (s *ClinicService) DeleteClinic(ctx context.Context, clinicID string) error {
	opCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 1. Get user_id and logo path associated with the clinic before deleting it
	var userID sql.NullString
	var logoPath sql.NullString
	err := s.DB.QueryRowContext(opCtx, `SELECT user_id, logo FROM clinics WHERE id = $1`, clinicID).Scan(&userID, &logoPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("clinic not found")
		}
		return errors.New("failed to find clinic details: " + err.Error())
	}

	// 2. Check if this user is linked to ANY other clinic or organization
	// This prevents deleting a user who might have roles elsewhere
	deleteUser := false
	if userID.Valid && userID.String != "" {
		var otherLinks int
		// Check user_roles for any other clinics
		err = s.DB.QueryRowContext(opCtx, `SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND clinic_id != $2`, userID.String, clinicID).Scan(&otherLinks)
		if err == nil && otherLinks == 0 {
			deleteUser = true
		}
	}

	// 3. Start transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// 4. Delete the clinic (CASCADE will handle staff, doctors, links in this DB)
	result, err := tx.ExecContext(opCtx, `DELETE FROM clinics WHERE id = $1`, clinicID)
	if err != nil {
		return errors.New("failed to delete clinic: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("clinic not found")
	}

	// 5. Delete the user if they have no other associations
	if deleteUser {
		// user_roles (CASCADE) and other associated auth data will be cleaned up automatically by DB FKs
		_, err = tx.ExecContext(opCtx, `DELETE FROM users WHERE id = $1`, userID.String)
		if err != nil {
			// We might not want to block clinic deletion if user deletion fails due to complex external relations
			// but for this task, we try to fulfill the request.
			return errors.New("failed to delete associated user: " + err.Error())
		}
	}

	// 6. Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.New("failed to commit clinic deletion")
	}

	// 7. Post-commit cleanup: Delete the logo file from disk
	if logoPath.Valid && logoPath.String != "" {
		_ = utils.DeleteImage(logoPath.String)
	}

	return nil
}
