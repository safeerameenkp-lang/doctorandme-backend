package patient

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// PatientRepository interface enforces bounds
type PatientRepository interface {
	WithTransaction(ctx context.Context, fn func(txRepo PatientRepository) error) error

	CheckPhoneExists(ctx context.Context, phone string, excludePatientID string) (bool, error)
	CheckMoIDExists(ctx context.Context, moID string, excludePatientID string) (bool, error)
	CheckClinicExists(ctx context.Context, clinicID string) (bool, error)
	CheckPatientClinicAssignment(ctx context.Context, patientID, clinicID string) (bool, error)

	CreateUser(ctx context.Context, input CreatePatientInput) (string, error)
	AssignPatientRole(ctx context.Context, userID string) error
	CreatePatientRecord(ctx context.Context, userID string, input CreatePatientInput) (string, error)
	AssignPatientToClinic(ctx context.Context, patientID, clinicID string, isPrimary bool) error

	GetClinicName(ctx context.Context, clinicID string) (string, error)
	GetPatientUserID(ctx context.Context, patientID string) (string, error)

	ListPatients(ctx context.Context, clinicID, search string, onlyActive bool) ([]PatientResponse, error)
	GetPatientByID(ctx context.Context, patientID string) (*PatientResponse, error)

	UpdatePatientDynamic(ctx context.Context, patientID string, query string, args []interface{}) error
	UpdateUserDynamic(ctx context.Context, userID string, query string, args []interface{}) error

	GetPatientAppointmentCount(ctx context.Context, patientID string) (int, error)
	SoftDeletePatientAndUser(ctx context.Context, patientID string, userID string) error
}

type DBExecer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type patientRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewPatientRepository(db *sql.DB) PatientRepository {
	return &patientRepository{db: db}
}

// execer allows transparent usage of either *sql.DB or *sql.Tx
func (r *patientRepository) execer() DBExecer {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// WithTransaction executes the given function within a database transaction controlled by the repository
func (r *patientRepository) WithTransaction(ctx context.Context, fn func(txRepo PatientRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	repoTx := &patientRepository{db: r.db, tx: tx}

	if err := fn(repoTx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *patientRepository) CheckPhoneExists(ctx context.Context, phone string, excludePatientID string) (bool, error) {
	var exists bool
	var err error
	if excludePatientID == "" {
		err = r.execer().QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM users 
				WHERE phone = $1 AND is_active = true
			)`, phone).Scan(&exists)
	} else {
		err = r.execer().QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM users u
				JOIN patients p ON p.user_id = u.id
				WHERE u.phone = $1 AND p.id != $2 AND u.is_active = true
			)`, phone, excludePatientID).Scan(&exists)
	}
	return exists, err
}

func (r *patientRepository) CheckMoIDExists(ctx context.Context, moID string, excludePatientID string) (bool, error) {
	var exists bool
	var err error
	if excludePatientID == "" {
		err = r.execer().QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM patients 
				WHERE mo_id = $1 AND is_active = true
			)`, moID).Scan(&exists)
	} else {
		err = r.execer().QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM patients 
				WHERE mo_id = $1 AND id != $2 AND is_active = true
			)`, moID, excludePatientID).Scan(&exists)
	}
	return exists, err
}

func (r *patientRepository) CheckClinicExists(ctx context.Context, clinicID string) (bool, error) {
	var exists bool
	err := r.execer().QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clinics 
			WHERE id = $1 AND is_active = true
		)`, clinicID).Scan(&exists)
	return exists, err
}

func (r *patientRepository) CheckPatientClinicAssignment(ctx context.Context, patientID, clinicID string) (bool, error) {
	var exists bool
	err := r.execer().QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM patient_clinics 
			WHERE patient_id = $1 AND clinic_id = $2
		)`, patientID, clinicID).Scan(&exists)
	return exists, err
}

func (r *patientRepository) CreateUser(ctx context.Context, input CreatePatientInput) (string, error) {
	var userID string
	err := r.execer().QueryRowContext(ctx, `
		INSERT INTO users (first_name, last_name, phone, email, date_of_birth, gender, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, CURRENT_TIMESTAMP)
		RETURNING id
	`, input.FirstName, input.LastName, input.Phone, input.Email, input.DateOfBirth, input.Gender).Scan(&userID)
	return userID, err
}

func (r *patientRepository) AssignPatientRole(ctx context.Context, userID string) error {
	_, err := r.execer().ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'patient' AND is_active = true
	`, userID)
	return err
}

func (r *patientRepository) CreatePatientRecord(ctx context.Context, userID string, input CreatePatientInput) (string, error) {
	var patientID string
	err := r.execer().QueryRowContext(ctx, `
		INSERT INTO patients (user_id, mo_id, medical_history, allergies, blood_group, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, true, CURRENT_TIMESTAMP)
		RETURNING id
	`, userID, input.MOID, input.MedicalHistory, input.Allergies, input.BloodGroup).Scan(&patientID)
	return patientID, err
}

func (r *patientRepository) AssignPatientToClinic(ctx context.Context, patientID, clinicID string, isPrimary bool) error {
	_, err := r.execer().ExecContext(ctx, `
		INSERT INTO patient_clinics (patient_id, clinic_id, is_primary, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, patientID, clinicID, isPrimary)
	return err
}

func (r *patientRepository) GetClinicName(ctx context.Context, clinicID string) (string, error) {
	var name string
	err := r.execer().QueryRowContext(ctx, `SELECT name FROM clinics WHERE id = $1`, clinicID).Scan(&name)
	return name, err
}

func (r *patientRepository) GetPatientUserID(ctx context.Context, patientID string) (string, error) {
	var userID string
	err := r.execer().QueryRowContext(ctx, `SELECT user_id FROM patients WHERE id = $1`, patientID).Scan(&userID)
	return userID, err
}

func (r *patientRepository) ListPatients(ctx context.Context, clinicID, search string, onlyActive bool) ([]PatientResponse, error) {
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("pc.clinic_id = $%d", argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	if onlyActive {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if search != "" {
		searchCondition := fmt.Sprintf(`(
			LOWER(u.first_name) LIKE LOWER($%d) OR 
			LOWER(u.last_name) LIKE LOWER($%d) OR 
			LOWER(u.phone) LIKE LOWER($%d) OR 
			LOWER(p.mo_id) LIKE LOWER($%d)
		)`, argIndex, argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.user_id, p.mo_id, u.first_name, u.last_name, u.phone, u.email,
		       u.date_of_birth, u.gender, p.medical_history, p.allergies, p.blood_group,
		       p.is_active, p.created_at, p.updated_at, c.name as clinic_name
		FROM patients p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN patient_clinics pc ON pc.patient_id = p.id
		LEFT JOIN clinics c ON c.id = pc.clinic_id
		%s
		ORDER BY p.created_at DESC
	`, whereClause)

	rows, err := r.execer().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []PatientResponse
	for rows.Next() {
		var patient PatientResponse
		var createdAt, updatedAt string
		var moID, email, dateOfBirth, gender, medicalHistory, allergies, bloodGroup, clinicName *string

		err := rows.Scan(
			&patient.ID, &patient.UserID, &moID, &patient.FirstName, &patient.LastName,
			&patient.Phone, &email, &dateOfBirth, &gender, &medicalHistory,
			&allergies, &bloodGroup, &patient.IsActive, &createdAt, &updatedAt, &clinicName,
		)
		if err != nil {
			continue
		}

		if moID != nil {
			patient.MOID = *moID
		}
		if email != nil {
			patient.Email = *email
		}
		if dateOfBirth != nil {
			patient.DateOfBirth = *dateOfBirth
		}
		if gender != nil {
			patient.Gender = *gender
		}
		if medicalHistory != nil {
			patient.MedicalHistory = *medicalHistory
		}
		if allergies != nil {
			patient.Allergies = *allergies
		}
		if bloodGroup != nil {
			patient.BloodGroup = *bloodGroup
		}
		if clinicName != nil {
			patient.ClinicName = *clinicName
		}
		patient.CreatedAt = createdAt
		patient.UpdatedAt = updatedAt

		patients = append(patients, patient)
	}

	return patients, nil
}

func (r *patientRepository) GetPatientByID(ctx context.Context, patientID string) (*PatientResponse, error) {
	var patient PatientResponse
	var createdAt, updatedAt string
	var moID, email, dateOfBirth, gender, medicalHistory, allergies, bloodGroup *string

	err := r.execer().QueryRowContext(ctx, `
		SELECT p.id, p.user_id, p.mo_id, u.first_name, u.last_name, u.phone, u.email,
		       u.date_of_birth, u.gender, p.medical_history, p.allergies, p.blood_group,
		       p.is_active, p.created_at, p.updated_at
		FROM patients p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = $1
	`, patientID).Scan(
		&patient.ID, &patient.UserID, &moID, &patient.FirstName, &patient.LastName,
		&patient.Phone, &email, &dateOfBirth, &gender, &medicalHistory,
		&allergies, &bloodGroup, &patient.IsActive, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if moID != nil {
		patient.MOID = *moID
	}
	if email != nil {
		patient.Email = *email
	}
	if dateOfBirth != nil {
		patient.DateOfBirth = *dateOfBirth
	}
	if gender != nil {
		patient.Gender = *gender
	}
	if medicalHistory != nil {
		patient.MedicalHistory = *medicalHistory
	}
	if allergies != nil {
		patient.Allergies = *allergies
	}
	if bloodGroup != nil {
		patient.BloodGroup = *bloodGroup
	}
	patient.CreatedAt = createdAt
	patient.UpdatedAt = updatedAt

	return &patient, nil
}

func (r *patientRepository) UpdatePatientDynamic(ctx context.Context, patientID string, query string, args []interface{}) error {
	_, err := r.execer().ExecContext(ctx, query, args...)
	return err
}

func (r *patientRepository) UpdateUserDynamic(ctx context.Context, userID string, query string, args []interface{}) error {
	_, err := r.execer().ExecContext(ctx, query, args...)
	return err
}

func (r *patientRepository) GetPatientAppointmentCount(ctx context.Context, patientID string) (int, error) {
	var count int
	err := r.execer().QueryRowContext(ctx, `SELECT COUNT(*) FROM appointments WHERE patient_id = $1`, patientID).Scan(&count)
	return count, err
}

func (r *patientRepository) SoftDeletePatientAndUser(ctx context.Context, patientID string, userID string) error {
	_, err := r.execer().ExecContext(ctx, `UPDATE patients SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, patientID)
	if err != nil {
		return err
	}
	_, err = r.execer().ExecContext(ctx, `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	return err
}
