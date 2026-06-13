package medicines

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrDuplicateMedicine = errors.New("duplicate medicine: a medicine with the same details already exists")
	ErrMedicineNotFound  = errors.New("medicine not found")
)

type Repository interface {
	Create(ctx context.Context, medicines []*Medicine) error
	GetByID(ctx context.Context, id, pharmacyID uuid.UUID) (*Medicine, error)
	List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Medicine, int, error)
	Search(ctx context.Context, pharmacyID uuid.UUID, search, brandName, category, barcode, supplierID string, isActive *bool, hasStock bool, limit, offset int) ([]*Medicine, int, error)
	Update(ctx context.Context, med *Medicine) error
	ValidateSuppliers(ctx context.Context, pharmacyID uuid.UUID, supplierIDs []uuid.UUID) (bool, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (*MedicineStats, error)
	CreateLog(ctx context.Context, tx *sql.Tx, log *MedicineAuditLog) error
	GetAuditLogs(ctx context.Context, medicineID, pharmacyID uuid.UUID) ([]*MedicineAuditLog, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

const medicineColumns = `
	id, pharmacy_id, created_by, created_by_name, updated_by, updated_by_name, name, dosage_form, category, manufacturer, 
	supplier_id, hsn_code, unit_type, brand_name, mfg_license, 
	schedule_type, is_rx_required, barcode, storage_condition, 
	cgst_rate, sgst_rate, is_active, created_at, updated_at
`

func (r *postgresRepository) Create(ctx context.Context, medicines []*Medicine) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt := `
		INSERT INTO inventory.medicines (
			id, pharmacy_id, created_by, created_by_name, updated_by, updated_by_name, name, dosage_form, category, manufacturer, 
			supplier_id, hsn_code, unit_type, brand_name, mfg_license, 
			schedule_type, is_rx_required, barcode, storage_condition, 
			cgst_rate, sgst_rate, is_active, created_at, updated_at
		) VALUES `

	var vals []interface{}
	for i, med := range medicines {
		stmt += fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			i*24+1, i*24+2, i*24+3, i*24+4, i*24+5, i*24+6, i*24+7, i*24+8, i*24+9, i*24+10, i*24+11, i*24+12, i*24+13, i*24+14, i*24+15, i*24+16, i*24+17, i*24+18, i*24+19, i*24+20, i*24+21, i*24+22, i*24+23, i*24+24)
		vals = append(vals,
			med.ID, med.PharmacyID, med.CreatedBy, med.CreatedByName, med.UpdatedBy, med.UpdatedByName, med.Name, med.DosageForm, med.Category, med.Manufacturer,
			med.SupplierID, med.HSNCode, med.UnitType, med.BrandName, med.MfgLicense,
			med.ScheduleType, med.IsRxRequired, med.Barcode, med.StorageCondition,
			med.CGSTRate, med.SGSTRate, med.IsActive, time.Now().UTC(), time.Now().UTC(),
		)
	}
	stmt = stmt[:len(stmt)-1] // remove trailing comma

	_, err = tx.ExecContext(ctx, stmt, vals...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateMedicine
		}
		return err
	}

	// 3. Create audit logs
	for _, med := range medicines {
		log := &MedicineAuditLog{
			ID:         uuid.New(),
			PharmacyID: med.PharmacyID,
			MedicineID: med.ID,
			ActionType: "CREATE",
			ChangedBy:  med.CreatedBy,
			ChangedByName: med.CreatedByName,
		}
		if err := r.CreateLog(ctx, tx, log); err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
	}

	return tx.Commit()
}

func (r *postgresRepository) GetByID(ctx context.Context, id, pharmacyID uuid.UUID) (*Medicine, error) {
	med := &Medicine{}
	query := fmt.Sprintf(`SELECT %s FROM inventory.medicines WHERE id = $1 AND pharmacy_id = $2`, medicineColumns)
	
	err := r.scanMedicine(r.db.QueryRowContext(ctx, query, id, pharmacyID), med)
	if err == sql.ErrNoRows {
		return nil, ErrMedicineNotFound
	}
	return med, err
}

func (r *postgresRepository) List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Medicine, int, error) {
	// 1. Get total count first (more efficient than window function for large datasets)
	countQuery := `SELECT COUNT(*) FROM inventory.medicines WHERE pharmacy_id = $1`
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, pharmacyID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []*Medicine{}, 0, nil
	}

	// 2. Get data
	query := fmt.Sprintf(`SELECT %s FROM inventory.medicines WHERE pharmacy_id = $1`, medicineColumns)
	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var medicines []*Medicine
	for rows.Next() {
		med := &Medicine{}
		if err := r.scanMedicine(rows, med); err != nil {
			return nil, 0, err
		}
		medicines = append(medicines, med)
	}
	return medicines, totalCount, nil
}

// Search with optional filters
func (r *postgresRepository) Search(ctx context.Context, pharmacyID uuid.UUID, search, brandName, category, barcode, supplierID string, isActive *bool, hasStock bool, limit, offset int) ([]*Medicine, int, error) {
	whereClause := " WHERE m.pharmacy_id = $1"
	args := []interface{}{pharmacyID}
	argCount := 1

	if search != "" {
		argCount++
		whereClause += fmt.Sprintf(` AND name ILIKE $%d`, argCount)
		args = append(args, "%"+search+"%")
	}

	if brandName != "" {
		argCount++
		whereClause += fmt.Sprintf(` AND brand_name ILIKE $%d`, argCount)
		args = append(args, "%"+brandName+"%")
	}

	if category != "" {
		argCount++
		whereClause += fmt.Sprintf(` AND category = $%d`, argCount)
		args = append(args, category)
	}

	if barcode != "" {
		argCount++
		whereClause += fmt.Sprintf(` AND barcode = $%d`, argCount)
		args = append(args, barcode)
	}

	if supplierID != "" {
		argCount++
		whereClause += fmt.Sprintf(` AND (m.supplier_id = $%d OR EXISTS (SELECT 1 FROM inventory.batches b WHERE b.medicine_id = m.id AND b.supplier_id = $%d AND b.quantity_available > 0))`, argCount, argCount)
		args = append(args, supplierID)
	}

	if isActive != nil {
		argCount++
		whereClause += fmt.Sprintf(` AND m.is_active = $%d`, argCount)
		args = append(args, *isActive)
	}

	if hasStock {
		whereClause += ` AND EXISTS (SELECT 1 FROM inventory.batches b WHERE b.medicine_id = m.id AND b.quantity_available > 0)`
	}

	// 1. Get total count for search
	var totalCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inventory.medicines m"+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []*Medicine{}, 0, nil
	}

	// 2. Get data for search
	argCount++
	query := fmt.Sprintf(`SELECT %s FROM inventory.medicines m`, medicineColumns) + whereClause
	query += fmt.Sprintf(` ORDER BY m.created_at DESC LIMIT $%d OFFSET $%d`, argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var medicines []*Medicine
	for rows.Next() {
		med := &Medicine{}
		if err := r.scanMedicine(rows, med); err != nil {
			return nil, 0, err
		}
		medicines = append(medicines, med)
	}
	return medicines, totalCount, nil
}

func (r *postgresRepository) Update(ctx context.Context, med *Medicine) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE inventory.medicines SET
			name = $1, dosage_form = $2, category = $3, manufacturer = $4,
			supplier_id = $5, hsn_code = $6, unit_type = $7, brand_name = $8,
			mfg_license = $9, schedule_type = $10, is_rx_required = $11, barcode = $12,
			storage_condition = $13, cgst_rate = $14, sgst_rate = $15, is_active = $16, updated_by = $17, updated_by_name = $18, updated_at = CURRENT_TIMESTAMP
		WHERE id = $19 AND pharmacy_id = $20`

	_, err = tx.ExecContext(ctx, query,
		med.Name, med.DosageForm, med.Category, med.Manufacturer,
		med.SupplierID, med.HSNCode, med.UnitType, med.BrandName,
		med.MfgLicense, med.ScheduleType, med.IsRxRequired, med.Barcode,
		med.StorageCondition, med.CGSTRate, med.SGSTRate, med.IsActive, med.UpdatedBy, med.UpdatedByName,
		med.ID, med.PharmacyID,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateMedicine
		}
		return err
	}

	// Create audit log
	log := &MedicineAuditLog{
		ID:         uuid.New(),
		PharmacyID: med.PharmacyID,
		MedicineID: med.ID,
		ActionType: "UPDATE",
		ChangedBy:  med.UpdatedBy,
		ChangedByName: med.UpdatedByName,
	}
	if err := r.CreateLog(ctx, tx, log); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return tx.Commit()
}

func (r *postgresRepository) ValidateSuppliers(ctx context.Context, pharmacyID uuid.UUID, supplierIDs []uuid.UUID) (bool, error) {
	if len(supplierIDs) == 0 {
		return true, nil
	}

	query := `
		SELECT COUNT(DISTINCT id) FROM supplier_schema.suppliers 
		WHERE id = ANY($1) AND pharmacy_id = $2 AND is_active = true`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, pq.Array(supplierIDs), pharmacyID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate suppliers bulk: %w", err)
	}
	
	uniqueExpected := make(map[uuid.UUID]bool)
	for _, id := range supplierIDs {
		uniqueExpected[id] = true
	}

	return count == len(uniqueExpected), nil
}

func (r *postgresRepository) GetStats(ctx context.Context, pharmacyID uuid.UUID) (*MedicineStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(1) FILTER (WHERE is_active = true) as active,
			COUNT(1) FILTER (WHERE is_rx_required = true) as restricted,
			COUNT(1) FILTER (WHERE cgst_rate = 0 AND sgst_rate = 0) as zero_gst
		FROM inventory.medicines 
		WHERE pharmacy_id = $1`

	stats := &MedicineStats{}
	err := r.db.QueryRowContext(ctx, query, pharmacyID).Scan(
		&stats.TotalMedicines,
		&stats.ActiveMedicines,
		&stats.RestrictedMedicines,
		&stats.ZeroGstMedicines,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *postgresRepository) CreateLog(ctx context.Context, tx *sql.Tx, log *MedicineAuditLog) error {
	query := `
		INSERT INTO inventory.medicine_audit_logs (
			id, pharmacy_id, medicine_id, action_type, changed_by, changed_by_name, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, log.ID, log.PharmacyID, log.MedicineID, log.ActionType, log.ChangedBy, log.ChangedByName, time.Now().UTC())
	} else {
		_, err = r.db.ExecContext(ctx, query, log.ID, log.PharmacyID, log.MedicineID, log.ActionType, log.ChangedBy, log.ChangedByName, time.Now().UTC())
	}
	return err
}

func (r *postgresRepository) GetAuditLogs(ctx context.Context, medicineID, pharmacyID uuid.UUID) ([]*MedicineAuditLog, error) {
	query := `
		SELECT id, pharmacy_id, medicine_id, action_type, changed_by, changed_by_name, changed_at 
		FROM inventory.medicine_audit_logs 
		WHERE medicine_id = $1 AND pharmacy_id = $2 
		ORDER BY changed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, medicineID, pharmacyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*MedicineAuditLog
	for rows.Next() {
		log := &MedicineAuditLog{}
		err := rows.Scan(&log.ID, &log.PharmacyID, &log.MedicineID, &log.ActionType, &log.ChangedBy, &log.ChangedByName, &log.ChangedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// scanMedicine is a helper method to scan a row into a Medicine struct
func (r *postgresRepository) scanMedicine(scanner interface {
	Scan(dest ...interface{}) error
}, med *Medicine) error {
	return scanner.Scan(
		&med.ID, &med.PharmacyID, &med.CreatedBy, &med.CreatedByName, &med.UpdatedBy, &med.UpdatedByName, &med.Name, &med.DosageForm, &med.Category,
		&med.Manufacturer, &med.SupplierID, &med.HSNCode, &med.UnitType,
		&med.BrandName, &med.MfgLicense, &med.ScheduleType, &med.IsRxRequired,
		&med.Barcode, &med.StorageCondition, &med.CGSTRate, &med.SGSTRate,
		&med.IsActive, &med.CreatedAt, &med.UpdatedAt,
	)
}
