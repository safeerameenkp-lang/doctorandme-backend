package services

import (
	"context"
	"database/sql"
	"errors"
	"organization-service/models"
	"time"
)

type ClinicPharmacyLinkService struct {
	DB *sql.DB
}

func NewClinicPharmacyLinkService(db *sql.DB) *ClinicPharmacyLinkService {
	return &ClinicPharmacyLinkService{DB: db}
}

// CreateLink directly links a clinic to a pharmacy and activates it immediately
func (s *ClinicPharmacyLinkService) CreateLink(ctx context.Context, input models.CreateClinicPharmacyLinkInput, userID string) (string, error) {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1. Verify clinic exists and is active
	var clinicActive bool
	err := s.DB.QueryRowContext(opCtx, `SELECT is_active FROM clinics WHERE id = $1`, input.ClinicID).Scan(&clinicActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("clinic not found")
		}
		return "", err
	}
	if !clinicActive {
		return "", errors.New("clinic is inactive")
	}

	// 2. Verify pharmacy exists and is active
	var pharmacyActive bool
	err = s.DB.QueryRowContext(opCtx, `SELECT is_active FROM pharmacies WHERE id = $1`, input.PharmacyID).Scan(&pharmacyActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("pharmacy not found")
		}
		return "", err
	}
	if !pharmacyActive {
		return "", errors.New("pharmacy is inactive")
	}

	// 3. Verify user has permission to link (admin on either clinic, pharmacy, organization side, or super_admin)
	var isAuthorized bool
	err = s.DB.QueryRowContext(opCtx, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND ur.is_active = true
			AND (
				r.name = 'super_admin'
				OR (ur.clinic_id = $2 AND r.name = 'clinic_admin')
				OR (ur.pharmacy_id = $3 AND r.name = 'pharmacy_admin')
				OR (ur.organization_id = (SELECT organization_id FROM clinics WHERE id = $2) AND r.name = 'organization_admin')
				OR (ur.organization_id = (SELECT organization_id FROM pharmacies WHERE id = $3) AND r.name = 'organization_admin')
			)
		)
	`, userID, input.ClinicID, input.PharmacyID).Scan(&isAuthorized)

	if err != nil {
		return "", err
	}
	if !isAuthorized {
		return "", errors.New("user not authorized to link this clinic and pharmacy")
	}

	// 4. Check if link already exists
	var existingLinkID string
	var existingIsActive bool
	err = s.DB.QueryRowContext(opCtx, `
		SELECT id, is_active FROM clinic_pharmacy_links 
		WHERE clinic_id = $1 AND pharmacy_id = $2
	`, input.ClinicID, input.PharmacyID).Scan(&existingLinkID, &existingIsActive)

	if err == nil {
		return "", errors.New("clinic is already linked to this pharmacy")
	} else if err != sql.ErrNoRows {
		return "", err
	}

	// 5. Insert active link
	var linkID string
	err = s.DB.QueryRowContext(opCtx, `
		INSERT INTO clinic_pharmacy_links (clinic_id, pharmacy_id, is_active)
		VALUES ($1, $2, TRUE)
		RETURNING id
	`, input.ClinicID, input.PharmacyID).Scan(&linkID)

	if err != nil {
		return "", err
	}

	return linkID, nil
}

// DeleteLink terminates a relationship link
func (s *ClinicPharmacyLinkService) DeleteLink(ctx context.Context, linkID string, userID string) error {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1. Fetch link details
	var clinicID, pharmacyID string
	err := s.DB.QueryRowContext(opCtx, `
		SELECT clinic_id, pharmacy_id FROM clinic_pharmacy_links WHERE id = $1
	`, linkID).Scan(&clinicID, &pharmacyID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("link not found")
		}
		return err
	}

	// 2. Verify authorization: either clinic admin, pharmacy admin, organization admin of either side, or super_admin
	var isAuthorized bool
	err = s.DB.QueryRowContext(opCtx, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND ur.is_active = true
			AND (
				r.name = 'super_admin'
				OR (ur.clinic_id = $2 AND r.name = 'clinic_admin')
				OR (ur.pharmacy_id = $3 AND r.name = 'pharmacy_admin')
				OR (ur.organization_id = (SELECT organization_id FROM clinics WHERE id = $2) AND r.name = 'organization_admin')
				OR (ur.organization_id = (SELECT organization_id FROM pharmacies WHERE id = $3) AND r.name = 'organization_admin')
			)
		)
	`, userID, clinicID, pharmacyID).Scan(&isAuthorized)

	if err != nil {
		return err
	}
	if !isAuthorized {
		return errors.New("user not authorized to delete this link")
	}

	// 3. Delete the link
	_, err = s.DB.ExecContext(opCtx, `DELETE FROM clinic_pharmacy_links WHERE id = $1`, linkID)
	return err
}
