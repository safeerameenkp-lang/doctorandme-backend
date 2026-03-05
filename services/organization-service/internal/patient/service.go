package patient

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotFound represents resource not found
var ErrNotFound = errors.New("not found")

// PatientService interface bounds the business logic limits
type PatientService interface {
	CreatePatient(ctx context.Context, input CreatePatientInput) (*PatientResponse, error)
	CreatePatientWithClinic(ctx context.Context, input CreatePatientInput) (*PatientResponse, string, error)
	ListPatients(ctx context.Context, clinicID, search, onlyActive string) ([]PatientResponse, error)
	GetPatient(ctx context.Context, patientID string) (*PatientResponse, error)
	UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error
	DeletePatient(ctx context.Context, patientID string) error
	AssignPatientToClinic(ctx context.Context, patientID string, input AssignClinicInput) error
}

type patientService struct {
	repo PatientRepository
}

// NewPatientService injects repository dependency
func NewPatientService(repo PatientRepository) PatientService {
	return &patientService{repo: repo}
}

func (s *patientService) CreatePatient(ctx context.Context, input CreatePatientInput) (*PatientResponse, error) {
	// Check if phone exists
	phoneExists, err := s.repo.CheckPhoneExists(ctx, input.Phone, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check phone number: %w", err)
	}
	if phoneExists {
		return nil, errors.New("phone_exists")
	}

	// Check if Mo ID exists
	if input.MOID != nil && *input.MOID != "" {
		moIDExists, err := s.repo.CheckMoIDExists(ctx, *input.MOID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to check MO ID: %w", err)
		}
		if moIDExists {
			return nil, errors.New("mo_id_exists")
		}
	}

	var patientID, userID string

	// Transaction boundary strictly managed inside the Service
	err = s.repo.WithTransaction(ctx, func(txRepo PatientRepository) error {
		userID, err = txRepo.CreateUser(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to create user account: %w", err)
		}

		if err = txRepo.AssignPatientRole(ctx, userID); err != nil {
			return fmt.Errorf("failed to assign patient role: %w", err)
		}

		patientID, err = txRepo.CreatePatientRecord(ctx, userID, input)
		if err != nil {
			return fmt.Errorf("failed to create patient record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// We return the minimal response directly replicating original behavior
	resp := &PatientResponse{
		ID:        patientID,
		UserID:    userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		IsActive:  true,
	}

	if input.MOID != nil {
		resp.MOID = *input.MOID
	}
	if input.Email != nil {
		resp.Email = *input.Email
	}
	if input.DateOfBirth != nil {
		resp.DateOfBirth = *input.DateOfBirth
	}
	if input.Gender != nil {
		resp.Gender = *input.Gender
	}
	if input.MedicalHistory != nil {
		resp.MedicalHistory = *input.MedicalHistory
	}
	if input.Allergies != nil {
		resp.Allergies = *input.Allergies
	}
	if input.BloodGroup != nil {
		resp.BloodGroup = *input.BloodGroup
	}

	return resp, nil
}

func (s *patientService) CreatePatientWithClinic(ctx context.Context, input CreatePatientInput) (*PatientResponse, string, error) {
	if input.ClinicID == nil || *input.ClinicID == "" {
		return nil, "", errors.New("missing_clinic_id")
	}

	// Verify clinic exists
	clinicExists, err := s.repo.CheckClinicExists(ctx, *input.ClinicID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check clinic: %w", err)
	}
	if !clinicExists {
		return nil, "", errors.New("clinic_not_found")
	}

	// Check if phone exists
	phoneExists, err := s.repo.CheckPhoneExists(ctx, input.Phone, "")
	if err != nil {
		return nil, "", fmt.Errorf("failed to check phone number: %w", err)
	}
	if phoneExists {
		return nil, "", errors.New("phone_exists")
	}

	// Check if Mo ID exists
	if input.MOID != nil && *input.MOID != "" {
		moIDExists, err := s.repo.CheckMoIDExists(ctx, *input.MOID, "")
		if err != nil {
			return nil, "", fmt.Errorf("failed to check MO ID: %w", err)
		}
		if moIDExists {
			return nil, "", errors.New("mo_id_exists")
		}
	}

	var patientID, userID string

	// Transaction boundary strictly managed inside the Service
	err = s.repo.WithTransaction(ctx, func(txRepo PatientRepository) error {
		userID, err = txRepo.CreateUser(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to create user account: %w", err)
		}

		if err = txRepo.AssignPatientRole(ctx, userID); err != nil {
			return fmt.Errorf("failed to assign patient role: %w", err)
		}

		patientID, err = txRepo.CreatePatientRecord(ctx, userID, input)
		if err != nil {
			return fmt.Errorf("failed to create patient record: %w", err)
		}

		if err = txRepo.AssignPatientToClinic(ctx, patientID, *input.ClinicID, true); err != nil {
			return fmt.Errorf("failed to assign patient to clinic: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	clinicName, err := s.repo.GetClinicName(ctx, *input.ClinicID)
	if err != nil {
		clinicName = "Unknown Clinic"
	}

	resp := &PatientResponse{
		ID:        patientID,
		UserID:    userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		IsActive:  true,
	}

	if input.MOID != nil {
		resp.MOID = *input.MOID
	}
	if input.Email != nil {
		resp.Email = *input.Email
	}
	if input.DateOfBirth != nil {
		resp.DateOfBirth = *input.DateOfBirth
	}
	if input.Gender != nil {
		resp.Gender = *input.Gender
	}
	if input.MedicalHistory != nil {
		resp.MedicalHistory = *input.MedicalHistory
	}
	if input.Allergies != nil {
		resp.Allergies = *input.Allergies
	}
	if input.BloodGroup != nil {
		resp.BloodGroup = *input.BloodGroup
	}

	return resp, clinicName, nil
}

func (s *patientService) ListPatients(ctx context.Context, clinicID, search, onlyActive string) ([]PatientResponse, error) {
	activeFlag := true
	if onlyActive == "false" {
		activeFlag = false
	}

	return s.repo.ListPatients(ctx, clinicID, search, activeFlag)
}

func (s *patientService) GetPatient(ctx context.Context, patientID string) (*PatientResponse, error) {
	patient, err := s.repo.GetPatientByID(ctx, patientID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch patient: %w", err)
	}
	return patient, nil
}

func (s *patientService) UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error {
	patientQuery := `UPDATE patients SET updated_at = CURRENT_TIMESTAMP`
	patientArgs := []interface{}{}
	argIndex := 1

	if input.MOID != nil {
		moIDExists, err := s.repo.CheckMoIDExists(ctx, *input.MOID, patientID)
		if err != nil {
			return fmt.Errorf("failed to check MO ID: %w", err)
		}
		if moIDExists {
			return errors.New("mo_id_exists")
		}
		patientQuery += fmt.Sprintf(`, mo_id = $%d`, argIndex)
		patientArgs = append(patientArgs, *input.MOID)
		argIndex++
	}

	if input.MedicalHistory != nil {
		patientQuery += fmt.Sprintf(`, medical_history = $%d`, argIndex)
		patientArgs = append(patientArgs, *input.MedicalHistory)
		argIndex++
	}

	if input.Allergies != nil {
		patientQuery += fmt.Sprintf(`, allergies = $%d`, argIndex)
		patientArgs = append(patientArgs, *input.Allergies)
		argIndex++
	}

	if input.BloodGroup != nil {
		patientQuery += fmt.Sprintf(`, blood_group = $%d`, argIndex)
		patientArgs = append(patientArgs, *input.BloodGroup)
		argIndex++
	}

	if input.IsActive != nil {
		patientQuery += fmt.Sprintf(`, is_active = $%d`, argIndex)
		patientArgs = append(patientArgs, *input.IsActive)
		argIndex++
	}

	patientQuery += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	patientArgs = append(patientArgs, patientID)

	// User updates
	userQuery := `UPDATE users SET updated_at = CURRENT_TIMESTAMP`
	userArgs := []interface{}{}
	userArgIndex := 1

	if input.FirstName != nil {
		userQuery += fmt.Sprintf(`, first_name = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.FirstName)
		userArgIndex++
	}

	if input.LastName != nil {
		userQuery += fmt.Sprintf(`, last_name = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.LastName)
		userArgIndex++
	}

	if input.Phone != nil {
		phoneExists, err := s.repo.CheckPhoneExists(ctx, *input.Phone, patientID)
		if err != nil {
			return fmt.Errorf("failed to check phone number: %w", err)
		}
		if phoneExists {
			return errors.New("phone_exists")
		}
		userQuery += fmt.Sprintf(`, phone = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.Phone)
		userArgIndex++
	}

	if input.Email != nil {
		userQuery += fmt.Sprintf(`, email = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.Email)
		userArgIndex++
	}

	if input.DateOfBirth != nil {
		userQuery += fmt.Sprintf(`, date_of_birth = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.DateOfBirth)
		userArgIndex++
	}

	if input.Gender != nil {
		userQuery += fmt.Sprintf(`, gender = $%d`, userArgIndex)
		userArgs = append(userArgs, *input.Gender)
		userArgIndex++
	}

	userID, err := s.repo.GetPatientUserID(ctx, patientID)
	if err != nil {
		return ErrNotFound
	}

	if len(userArgs) > 0 {
		userQuery += fmt.Sprintf(` WHERE id = $%d`, userArgIndex)
		userArgs = append(userArgs, userID)
	}

	// Dynamic Updating through a transaction
	return s.repo.WithTransaction(ctx, func(txRepo PatientRepository) error {
		if len(patientArgs) > 1 {
			if err := txRepo.UpdatePatientDynamic(ctx, patientID, patientQuery, patientArgs); err != nil {
				return fmt.Errorf("failed to update patient: %w", err)
			}
		}

		if len(userArgs) > 1 {
			if err := txRepo.UpdateUserDynamic(ctx, userID, userQuery, userArgs); err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}
		return nil
	})
}

func (s *patientService) DeletePatient(ctx context.Context, patientID string) error {
	count, err := s.repo.GetPatientAppointmentCount(ctx, patientID)
	if err != nil {
		return fmt.Errorf("failed to check patient appointments: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("patient has %d appointments", count)
	}

	userID, err := s.repo.GetPatientUserID(ctx, patientID)
	if err != nil {
		return ErrNotFound
	}

	return s.repo.WithTransaction(ctx, func(txRepo PatientRepository) error {
		if err := txRepo.SoftDeletePatientAndUser(ctx, patientID, userID); err != nil {
			return err
		}
		return nil
	})
}

func (s *patientService) AssignPatientToClinic(ctx context.Context, patientID string, input AssignClinicInput) error {
	// Verify patient
	// passing moID empty checks patient exists if not filtered
	_, err := s.repo.CheckMoIDExists(ctx, "", patientID)

	// actually let's use a explicit patient existence check or we have GetPatientByID
	_, err = s.repo.GetPatientByID(ctx, patientID)
	if err != nil {
		return ErrNotFound
	}

	// Verify clinic exists
	clinicExists, err := s.repo.CheckClinicExists(ctx, input.ClinicID)
	if err != nil {
		return fmt.Errorf("failed to check clinic: %w", err)
	}
	if !clinicExists {
		return errors.New("clinic_not_found")
	}

	alreadyAssigned, err := s.repo.CheckPatientClinicAssignment(ctx, patientID, input.ClinicID)
	if err != nil {
		return fmt.Errorf("failed to check clinic assignment: %w", err)
	}
	if alreadyAssigned {
		return errors.New("already_assigned")
	}

	if err := s.repo.AssignPatientToClinic(ctx, patientID, input.ClinicID, false); err != nil {
		return fmt.Errorf("failed to assign patient to clinic: %w", err)
	}

	return nil
}
