package supplier

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrSupplierNotFound   = errors.New("supplier not found")
	ErrSupplierNameExists = errors.New("a supplier with this name already exists for your pharmacy")
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *Supplier) error
	FindByID(ctx context.Context, id, pharmacyID uuid.UUID) (*Supplier, error)
	FindAll(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Supplier, int, error)
	Search(ctx context.Context, pharmacyID uuid.UUID, search string, limit, offset int) ([]*Supplier, int, error)
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id, pharmacyID uuid.UUID) error // Hard delete if needed, but we use activation/deactivation
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (*SupplierStats, error)
	CreateAuditLog(ctx context.Context, log *SupplierAuditLog) error
	GetHistory(ctx context.Context, supplierID, pharmacyID uuid.UUID) ([]*SupplierAuditLog, error)
}

type postgresSupplierRepo struct {
	db *sql.DB
}

func NewPostgresSupplierRepository(db *sql.DB) SupplierRepository {
	return &postgresSupplierRepo{db: db}
}

func (r *postgresSupplierRepo) Create(ctx context.Context, s *Supplier) error {
	query := `
		INSERT INTO supplier_schema.suppliers (
			id, pharmacy_id, name, supplier_type, contact_person, contact_number, 
			website, email, address, state, pincode, 
			gst_number, pan_number, license_number, 
			bank_name, account_name, account_number, ifsc_code,
			credit_period_days, credit_limit, is_active, created_at, updated_at,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.PharmacyID, s.Name, s.SupplierType, s.ContactPerson, s.ContactNumber,
		s.Website, s.Email, s.Address, s.State, s.Pincode,
		s.GSTNumber, s.PANNumber, s.LicenseNumber,
		s.BankDetails.BankName, s.BankDetails.AccountName, s.BankDetails.AccountNumber, s.BankDetails.IFSCCode,
		s.CreditTerms.CreditPeriodDays, s.CreditTerms.CreditLimit, s.IsActive, s.CreatedAt, s.UpdatedAt,
		s.CreatedBy, s.UpdatedBy,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && (pqErr.Constraint == "uq_suppliers_pharmacy_name" || pqErr.Constraint == "suppliers_name_key") {
				return ErrSupplierNameExists
			}
		}
		return err
	}
	return nil
}

func (r *postgresSupplierRepo) FindByID(ctx context.Context, id, pharmacyID uuid.UUID) (*Supplier, error) {
	query := `
		SELECT 
			id, pharmacy_id, name, supplier_type, contact_person, contact_number, 
			website, email, address, state, pincode, 
			gst_number, pan_number, license_number, 
			bank_name, account_name, account_number, ifsc_code,
			credit_period_days, credit_limit, is_active, created_at, updated_at,
			created_by, updated_by
		FROM supplier_schema.suppliers
		WHERE id = $1 AND pharmacy_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, id, pharmacyID)

	var s Supplier
	err := row.Scan(
		&s.ID, &s.PharmacyID, &s.Name, &s.SupplierType, &s.ContactPerson, &s.ContactNumber,
		&s.Website, &s.Email, &s.Address, &s.State, &s.Pincode,
		&s.GSTNumber, &s.PANNumber, &s.LicenseNumber,
		&s.BankDetails.BankName, &s.BankDetails.AccountName, &s.BankDetails.AccountNumber, &s.BankDetails.IFSCCode,
		&s.CreditTerms.CreditPeriodDays, &s.CreditTerms.CreditLimit, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		&s.CreatedBy, &s.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSupplierNotFound
	}
	return &s, err
}

func (r *postgresSupplierRepo) FindAll(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Supplier, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM supplier_schema.suppliers WHERE pharmacy_id = $1`
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, pharmacyID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, pharmacy_id, name, supplier_type, contact_person, contact_number, 
			website, email, address, state, pincode, 
			gst_number, pan_number, license_number, 
			bank_name, account_name, account_number, ifsc_code,
			credit_period_days, credit_limit, is_active, created_at, updated_at,
			created_by, updated_by
		FROM supplier_schema.suppliers
		WHERE pharmacy_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var suppliers []*Supplier
	for rows.Next() {
		var s Supplier
		err := rows.Scan(
			&s.ID, &s.PharmacyID, &s.Name, &s.SupplierType, &s.ContactPerson, &s.ContactNumber,
			&s.Website, &s.Email, &s.Address, &s.State, &s.Pincode,
			&s.GSTNumber, &s.PANNumber, &s.LicenseNumber,
			&s.BankDetails.BankName, &s.BankDetails.AccountName, &s.BankDetails.AccountNumber, &s.BankDetails.IFSCCode,
			&s.CreditTerms.CreditPeriodDays, &s.CreditTerms.CreditLimit, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
			&s.CreatedBy, &s.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		suppliers = append(suppliers, &s)
	}
	return suppliers, totalCount, nil
}

func (r *postgresSupplierRepo) Search(ctx context.Context, pharmacyID uuid.UUID, search string, limit, offset int) ([]*Supplier, int, error) {
	whereClause := " WHERE pharmacy_id = $1 AND name ILIKE $2"
	args := []interface{}{pharmacyID, "%" + search + "%"}

	var totalCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM supplier_schema.suppliers"+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, pharmacy_id, name, supplier_type, contact_person, contact_number, 
			website, email, address, state, pincode, 
			gst_number, pan_number, license_number, 
			bank_name, account_name, account_number, ifsc_code,
			credit_period_days, credit_limit, is_active, created_at, updated_at,
			created_by, updated_by
		FROM supplier_schema.suppliers
	` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var suppliers []*Supplier
	for rows.Next() {
		var s Supplier
		err := rows.Scan(
			&s.ID, &s.PharmacyID, &s.Name, &s.SupplierType, &s.ContactPerson, &s.ContactNumber,
			&s.Website, &s.Email, &s.Address, &s.State, &s.Pincode,
			&s.GSTNumber, &s.PANNumber, &s.LicenseNumber,
			&s.BankDetails.BankName, &s.BankDetails.AccountName, &s.BankDetails.AccountNumber, &s.BankDetails.IFSCCode,
			&s.CreditTerms.CreditPeriodDays, &s.CreditTerms.CreditLimit, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
			&s.CreatedBy, &s.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		suppliers = append(suppliers, &s)
	}
	return suppliers, totalCount, nil
}

func (r *postgresSupplierRepo) Update(ctx context.Context, s *Supplier) error {
	query := `
		UPDATE supplier_schema.suppliers
		SET 
			name = $3, supplier_type = $4, contact_person = $5, contact_number = $6,
			website = $7, email = $8, address = $9, state = $10, pincode = $11,
			gst_number = $12, pan_number = $13, license_number = $14,
			bank_name = $15, account_name = $16, account_number = $17, ifsc_code = $18,
			credit_period_days = $19, credit_limit = $20, is_active = $21, updated_at = $22, updated_by = $23
		WHERE id = $1 AND pharmacy_id = $2
	`
	res, err := r.db.ExecContext(ctx, query,
		s.ID, s.PharmacyID, s.Name, s.SupplierType, s.ContactPerson, s.ContactNumber,
		s.Website, s.Email, s.Address, s.State, s.Pincode,
		s.GSTNumber, s.PANNumber, s.LicenseNumber,
		s.BankDetails.BankName, s.BankDetails.AccountName, s.BankDetails.AccountNumber, s.BankDetails.IFSCCode,
		s.CreditTerms.CreditPeriodDays, s.CreditTerms.CreditLimit, s.IsActive, s.CreatedAt, s.UpdatedAt,
		s.CreatedBy, s.UpdatedBy,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && (pqErr.Constraint == "uq_suppliers_pharmacy_name" || pqErr.Constraint == "suppliers_name_key") {
				return ErrSupplierNameExists
			}
		}
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrSupplierNotFound
	}
	return nil
}

func (r *postgresSupplierRepo) Delete(ctx context.Context, id, pharmacyID uuid.UUID) error {
	query := `DELETE FROM supplier_schema.suppliers WHERE id = $1 AND pharmacy_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, pharmacyID)
	return err
}

func (r *postgresSupplierRepo) GetStats(ctx context.Context, pharmacyID uuid.UUID) (*SupplierStats, error) {
	query := `
		SELECT 
            COUNT(id) as total,
            COUNT(id) FILTER (WHERE is_active = true) as active,
            COUNT(id) FILTER (WHERE credit_period_days > 0 AND is_active = true) as credit,
            COUNT(id) FILTER (WHERE gst_number IS NULL OR gst_number = '') as gst_pending
        FROM supplier_schema.suppliers
        WHERE pharmacy_id = $1
	`
	stats := &SupplierStats{}
	err := r.db.QueryRowContext(ctx, query, pharmacyID).Scan(
		&stats.TotalSuppliers,
		&stats.ActiveSuppliers,
		&stats.CreditSuppliers,
		&stats.GSTPending,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *postgresSupplierRepo) CreateAuditLog(ctx context.Context, l *SupplierAuditLog) error {
	query := `
		INSERT INTO supplier_schema.supplier_audit_logs (
			id, pharmacy_id, supplier_id, action_type, changed_by, changed_by_name, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		l.ID, l.PharmacyID, l.SupplierID, l.ActionType, l.ChangedBy, l.ChangedByName, l.ChangedAt,
	)
	return err
}

func (r *postgresSupplierRepo) GetHistory(ctx context.Context, supplierID, pharmacyID uuid.UUID) ([]*SupplierAuditLog, error) {
	query := `
		SELECT 
			id, pharmacy_id, supplier_id, action_type, changed_by, changed_by_name, changed_at
		FROM supplier_schema.supplier_audit_logs
		WHERE supplier_id = $1 AND pharmacy_id = $2
		ORDER BY changed_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, supplierID, pharmacyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*SupplierAuditLog
	for rows.Next() {
		var l SupplierAuditLog
		err := rows.Scan(
			&l.ID, &l.PharmacyID, &l.SupplierID, &l.ActionType, &l.ChangedBy, &l.ChangedByName, &l.ChangedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
