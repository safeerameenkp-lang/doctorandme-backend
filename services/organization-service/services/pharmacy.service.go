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

type PharmacyService struct {
	DB *sql.DB
}

func NewPharmacyService(db *sql.DB) *PharmacyService {
	return &PharmacyService{DB: db}
}

// CreatePharmacy handles the creation logic including image upload
func (s *PharmacyService) CreatePharmacy(ctx context.Context, input models.CreatePharmacyInput, logoFile *multipart.FileHeader) (string, *string, error) {
	// Verify organization exists
	var orgExists bool

	// Create strict operations context mapped bounds
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`, input.OrganizationID).Scan(&orgExists)
	if err != nil || !orgExists {
		return "", nil, errors.New("organization not found")
	}

	// Verify clinic exists if provided
	if input.ClinicID != nil && *input.ClinicID != "" {
		var clinicExists bool
		err = s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1)`, *input.ClinicID).Scan(&clinicExists)
		if err != nil || !clinicExists {
			return "", nil, errors.New("clinic not found")
		}
	}

	// Verify user exists and context isn't halted
	var userExists bool
	err = s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, input.UserID).Scan(&userExists)
	if err != nil || !userExists {
		return "", nil, errors.New("user not found")
	}

	// Validate that the user is not already an organization admin
	var isOrgAdmin bool
	err = s.DB.QueryRowContext(opCtx, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur 
			JOIN roles r ON ur.role_id = r.id 
			WHERE ur.user_id = $1 AND r.name = 'organization_admin'
		)`, input.UserID).Scan(&isOrgAdmin)
	if err == nil && isOrgAdmin {
		return "", nil, errors.New("this user is already an organization admin and cannot be assigned as a pharmacy admin")
	}

	// Validate and format PharmacyType
	input.PharmacyType = strings.TrimSpace(input.PharmacyType)
	if input.PharmacyType == "" {
		return "", nil, errors.New("pharmacy_type cannot be empty")
	}
	input.PharmacyType = strings.Title(strings.ToLower(input.PharmacyType))

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

			savedPath, err := utils.SaveOptimizedImage(file, logoFile.Filename, "pharmacies")
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

	// Auto-generate pharmacy code if missing (Inside transaction for locking)
	var pharmacyCode string
	if input.PharmacyCode != nil && *input.PharmacyCode != "" {
		pharmacyCode = *input.PharmacyCode
	} else {
		var err error
		pharmacyCode, err = utils.GeneratePharmacyCode(opCtx, tx, s.DB, input.Name)
		if err != nil {
			return "", nil, errors.New("failed to generate pharmacy code")
		}
	}

	var pharmacyID string
	err = tx.QueryRowContext(opCtx, `
        INSERT INTO pharmacies (organization_id, clinic_id, user_id, pharmacy_code, name, pharmacy_type, email, phone, address, license_number, logo)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id
    `, input.OrganizationID, input.ClinicID, input.UserID, pharmacyCode, input.Name, input.PharmacyType, input.Email, input.Phone, input.Address, input.LicenseNumber, logoPath).Scan(&pharmacyID)

	if err != nil {
		return "", nil, errors.New("failed to insert pharmacy: " + err.Error())
	}

	// Assign pharmacy_admin role to the assigned user
	var roleID string
	err = tx.QueryRowContext(opCtx, `SELECT id FROM roles WHERE name='pharmacy_admin' LIMIT 1`).Scan(&roleID)
	if err == nil {
		_, _ = tx.ExecContext(opCtx, `
            INSERT INTO user_roles (user_id, role_id, pharmacy_id)
            VALUES ($1, $2, $3)
            ON CONFLICT (user_id, role_id, pharmacy_id) DO NOTHING
        `, input.UserID, roleID, pharmacyID)
	}

	if err = tx.Commit(); err != nil {
		return "", nil, errors.New("failed to commit pharmacy creation")
	}

	return pharmacyID, logoPath, nil
}

// CreatePharmacyWithAdmin handles creating a pharmacy along with a new admin user
func (s *PharmacyService) CreatePharmacyWithAdmin(ctx context.Context, input models.CreatePharmacyWithAdminInput, logoFile *multipart.FileHeader) (string, string, *string, error) {
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

	// Verify clinic exists if provided
	if input.ClinicID != nil && *input.ClinicID != "" {
		var clinicExists bool
		err = s.DB.QueryRowContext(opCtx, `SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1)`, *input.ClinicID).Scan(&clinicExists)
		if err != nil || !clinicExists {
			return "", "", nil, errors.New("clinic not found")
		}
	}

	// Validate and format PharmacyType
	input.PharmacyType = strings.TrimSpace(input.PharmacyType)
	if input.PharmacyType == "" {
		return "", "", nil, errors.New("pharmacy_type cannot be empty")
	}
	input.PharmacyType = strings.Title(strings.ToLower(input.PharmacyType))

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

			savedPath, err := utils.SaveOptimizedImage(file, logoFile.Filename, "pharmacies")
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

	// Auto-generate pharmacy code if missing (Inside transaction for locking)
	var pharmacyCode string
	if input.PharmacyCode != nil && *input.PharmacyCode != "" {
		pharmacyCode = *input.PharmacyCode
	} else {
		var err error
		pharmacyCode, err = utils.GeneratePharmacyCode(opCtx, tx, s.DB, input.Name)
		if err != nil {
			return "", "", nil, errors.New("failed to generate pharmacy code")
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
		// Existing user found, validate that they are not already an organization admin
		var isOrgAdmin bool
		err = tx.QueryRowContext(opCtx, `
			SELECT EXISTS(
				SELECT 1 FROM user_roles ur 
				JOIN roles r ON ur.role_id = r.id 
				WHERE ur.user_id = $1 AND r.name = 'organization_admin'
			)`, adminID).Scan(&isOrgAdmin)
		if err == nil && isOrgAdmin {
			return "", "", nil, errors.New("this user is already an organization admin and cannot be a pharmacy admin")
		}
	}

	// 2. Create pharmacy with admin as user_id (the owner)
	var pharmacyID string
	err = tx.QueryRowContext(opCtx, `
        INSERT INTO pharmacies (organization_id, clinic_id, user_id, pharmacy_code, name, pharmacy_type, email, phone, address, license_number, logo)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id
    `, input.OrganizationID, input.ClinicID, adminID, pharmacyCode, input.Name, input.PharmacyType, input.Email, input.Phone, input.Address, input.LicenseNumber, logoPath).Scan(&pharmacyID)
	if err != nil {
		return "", "", nil, errors.New("failed to create pharmacy: " + err.Error())
	}

	// Assign pharmacy_admin role
	var roleID string
	err = tx.QueryRowContext(opCtx, `SELECT id FROM roles WHERE name='pharmacy_admin' LIMIT 1`).Scan(&roleID)
	if err != nil {
		return "", "", nil, errors.New("failed to find pharmacy_admin role")
	}

	_, err = tx.ExecContext(opCtx, `INSERT INTO user_roles (user_id, role_id, pharmacy_id) VALUES ($1,$2,$3)`, adminID, roleID, pharmacyID)
	if err != nil {
		return "", "", nil, errors.New("failed to assign pharmacy admin role")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return "", "", nil, errors.New("failed to commit transaction")
	}

	return pharmacyID, adminID, logoPath, nil
}

// DeletePharmacy handles deleting a pharmacy and its associated admin user
func (s *PharmacyService) DeletePharmacy(ctx context.Context, pharmacyID string) error {
	opCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 1. Get user_id and logo path associated with the pharmacy before deleting it
	var userID sql.NullString
	var logoPath sql.NullString
	err := s.DB.QueryRowContext(opCtx, `SELECT user_id, logo FROM pharmacies WHERE id = $1`, pharmacyID).Scan(&userID, &logoPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pharmacy not found")
		}
		return errors.New("failed to find pharmacy details: " + err.Error())
	}

	// 2. Check if this user is linked to ANY other pharmacy, clinic or organization
	// This prevents deleting a user who might have roles elsewhere
	deleteUser := false
	if userID.Valid && userID.String != "" {
		var otherLinks int
		// Check user_roles for any other pharmacies or clinics
		err = s.DB.QueryRowContext(opCtx, `SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND (pharmacy_id != $2 OR clinic_id IS NOT NULL)`, userID.String, pharmacyID).Scan(&otherLinks)
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

	// 4. Delete the pharmacy
	result, err := tx.ExecContext(opCtx, `DELETE FROM pharmacies WHERE id = $1`, pharmacyID)
	if err != nil {
		return errors.New("failed to delete pharmacy: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("pharmacy not found")
	}

	// 5. Delete the user if they have no other associations
	if deleteUser {
		_, err = tx.ExecContext(opCtx, `DELETE FROM users WHERE id = $1`, userID.String)
		if err != nil {
			return errors.New("failed to delete associated user: " + err.Error())
		}
	}

	// 6. Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.New("failed to commit pharmacy deletion")
	}

	// 7. Post-commit cleanup: Delete the logo file from disk
	if logoPath.Valid && logoPath.String != "" {
		_ = utils.DeleteImage(logoPath.String)
	}

	return nil
}
