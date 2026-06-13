package batches

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	UpsertBatch(ctx context.Context, tx *sql.Tx, dto UpdateBatchDTO) (uuid.UUID, error)
	ListBatches(ctx context.Context, pharmacyID uuid.UUID, medicineID *uuid.UUID, limit, offset int, search, supplierID, filter string) ([]Batch, int, error)
	ListSellableBatches(ctx context.Context, pharmacyID uuid.UUID, search string, limit int) ([]Batch, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (BatchStats, error)
	CreateBatchLog(ctx context.Context, tx *sql.Tx, log BatchAuditLog) error
	GetBatchAuditLogs(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]BatchAuditLog, error)
	Update(ctx context.Context, tx *sql.Tx, pharmacyID, batchID uuid.UUID, req EditBatchRequest) error
	GetBatch(ctx context.Context, pharmacyID, batchID uuid.UUID) (*Batch, error)
	UpdateBatchQuantity(ctx context.Context, tx *sql.Tx, pharmacyID, batchID uuid.UUID, quantityChange int) error
	AddReturnStock(ctx context.Context, tx *sql.Tx, dto UpdateBatchDTO, batchID uuid.UUID) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) UpsertBatch(ctx context.Context, tx *sql.Tx, dto UpdateBatchDTO) (uuid.UUID, error) {
	// Query to Insert new batch or Update existing one (Increment Quantity)
	query := `
		INSERT INTO inventory.batches (
			pharmacy_id, medicine_id, batch_no, mfg_date, expiry_date, rack_no,
			quantity_available, cost_price, mrp, unit_price,
			cgst_rate, sgst_rate, total_tax_percentage,
			retail_disc_perc, staff_disc_perc, special_disc_perc, max_disc_perc,
			supplier_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17,
			$18, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (pharmacy_id, medicine_id, batch_no) 
		DO UPDATE SET 
			mfg_date = EXCLUDED.mfg_date,
			expiry_date = EXCLUDED.expiry_date,
			quantity_available = inventory.batches.quantity_available + EXCLUDED.quantity_available,
			rack_no = COALESCE(NULLIF(EXCLUDED.rack_no, ''), inventory.batches.rack_no),
			cost_price = EXCLUDED.cost_price,
			mrp = EXCLUDED.mrp,
			unit_price = EXCLUDED.unit_price,
			cgst_rate = EXCLUDED.cgst_rate,
			sgst_rate = EXCLUDED.sgst_rate,
			total_tax_percentage = EXCLUDED.total_tax_percentage,
			retail_disc_perc = EXCLUDED.retail_disc_perc,
			staff_disc_perc = EXCLUDED.staff_disc_perc,
			special_disc_perc = EXCLUDED.special_disc_perc,
			max_disc_perc = EXCLUDED.max_disc_perc,
			supplier_id = EXCLUDED.supplier_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, quantity_available
	`

	var batchID uuid.UUID
	var balanceAfter int
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query,
			dto.PharmacyID, dto.MedicineID, dto.BatchNo, dto.MfgDate, dto.ExpiryDate, dto.RackNo,
			dto.QuantityToAdd, dto.CostPrice, dto.MRP, dto.UnitPrice,
			dto.CGSTRate, dto.SGSTRate, dto.TotalTaxPercentage,
			dto.RetailDiscPerc, dto.StaffDiscPerc, dto.SpecialDiscPerc, dto.MaxDiscPerc,
			dto.SupplierID,
		).Scan(&batchID, &balanceAfter)
	} else {
		err = r.db.QueryRowContext(ctx, query,
			dto.PharmacyID, dto.MedicineID, dto.BatchNo, dto.MfgDate, dto.ExpiryDate, dto.RackNo,
			dto.QuantityToAdd, dto.CostPrice, dto.MRP, dto.UnitPrice,
			dto.CGSTRate, dto.SGSTRate, dto.TotalTaxPercentage,
			dto.RetailDiscPerc, dto.StaffDiscPerc, dto.SpecialDiscPerc, dto.MaxDiscPerc,
			dto.SupplierID,
		).Scan(&batchID, &balanceAfter)
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert batch: %w", err)
	}

	// 2. Record in Ledger (Audit Trail)
	ledgerQuery := `
		INSERT INTO inventory.stock_ledger (
			pharmacy_id, medicine_id, batch_id, transaction_type, 
			quantity_change, balance_after, reference_type, 
			reference_id, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if tx != nil {
		_, err = tx.ExecContext(ctx, ledgerQuery,
			dto.PharmacyID, dto.MedicineID, batchID, dto.TransactionType,
			dto.QuantityToAdd, balanceAfter, dto.ReferenceType,
			dto.ReferenceID, dto.PerformedBy, dto.Notes,
		)
	} else {
		_, err = r.db.ExecContext(ctx, ledgerQuery,
			dto.PharmacyID, dto.MedicineID, batchID, dto.TransactionType,
			dto.QuantityToAdd, balanceAfter, dto.ReferenceType,
			dto.ReferenceID, dto.PerformedBy, dto.Notes,
		)
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to record stock ledger: %w", err)
	}

	return batchID, nil
}

func (r *postgresRepository) ListBatches(ctx context.Context, pharmacyID uuid.UUID, medicineID *uuid.UUID, limit, offset int, search, supplierID, filter string) ([]Batch, int, error) {
	// 1. Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM inventory.batches b
		LEFT JOIN inventory.medicines m ON b.medicine_id = m.id
		WHERE b.pharmacy_id = $1
	`
	countArgs := []interface{}{pharmacyID}
	placeholderID := 2

	if medicineID != nil {
		countQuery += fmt.Sprintf(" AND b.medicine_id = $%d", placeholderID)
		countArgs = append(countArgs, *medicineID)
		placeholderID++
	}

	if search != "" {
		countQuery += fmt.Sprintf(" AND (b.batch_no ILIKE $%d OR m.name ILIKE $%d)", placeholderID, placeholderID)
		countArgs = append(countArgs, "%"+search+"%")
		placeholderID++
	}

	if supplierID != "" {
		countQuery += fmt.Sprintf(" AND b.supplier_id = $%d", placeholderID)
		countArgs = append(countArgs, supplierID)
		placeholderID++
	}

	if filter == "expiring" {
		countQuery += " AND b.expiry_date >= CURRENT_DATE AND b.expiry_date < (CURRENT_DATE + INTERVAL '30 days') AND b.quantity_available > 0"
	} else if filter == "high_risk" {
		countQuery += " AND b.expiry_date >= CURRENT_DATE AND b.expiry_date < (CURRENT_DATE + INTERVAL '7 days') AND b.quantity_available > 0"
	} else if filter == "expired" {
		countQuery += " AND b.expiry_date < CURRENT_DATE AND b.quantity_available > 0"
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count batches: %w", err)
	}

	// 2. Get paginated data
	query := `
		SELECT 
			b.id, b.pharmacy_id, b.medicine_id, m.name as medicine_name, m.brand_name as medicine_brand, b.batch_no, b.mfg_date, b.expiry_date, b.rack_no,
			b.quantity_available, b.cost_price, b.mrp, b.unit_price,
			b.cgst_rate, b.sgst_rate, b.total_tax_percentage,
			b.retail_disc_perc, b.staff_disc_perc, b.special_disc_perc, b.max_disc_perc,
			b.created_at, b.updated_at, b.supplier_id, s.name as supplier_name
		FROM inventory.batches b
		LEFT JOIN inventory.medicines m ON b.medicine_id = m.id
		LEFT JOIN supplier_schema.suppliers s ON b.supplier_id = s.id
		WHERE b.pharmacy_id = $1
	`

	args := []interface{}{pharmacyID}
	placeholderID = 2

	if medicineID != nil {
		query += fmt.Sprintf(" AND b.medicine_id = $%d", placeholderID)
		args = append(args, *medicineID)
		placeholderID++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (b.batch_no ILIKE $%d OR m.name ILIKE $%d)", placeholderID, placeholderID)
		args = append(args, "%"+search+"%")
		placeholderID++
	}

	if supplierID != "" {
		query += fmt.Sprintf(" AND b.supplier_id = $%d", placeholderID)
		args = append(args, supplierID)
		placeholderID++
	}

	if filter == "expiring" {
		query += " AND b.expiry_date >= CURRENT_DATE AND b.expiry_date < (CURRENT_DATE + INTERVAL '30 days') AND b.quantity_available > 0"
	} else if filter == "high_risk" {
		query += " AND b.expiry_date >= CURRENT_DATE AND b.expiry_date < (CURRENT_DATE + INTERVAL '7 days') AND b.quantity_available > 0"
	} else if filter == "expired" {
		query += " AND b.expiry_date < CURRENT_DATE AND b.quantity_available > 0"
	}

	// Show batches expiring soonest first
	query += fmt.Sprintf(" ORDER BY b.expiry_date ASC LIMIT $%d OFFSET $%d", placeholderID, placeholderID+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query batches: %w", err)
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		var b Batch
		var medicineName sql.NullString
		var medicineBrand sql.NullString
		var supplierName sql.NullString
		err := rows.Scan(
			&b.ID, &b.PharmacyID, &b.MedicineID, &medicineName, &medicineBrand, &b.BatchNo, &b.MfgDate, &b.ExpiryDate, &b.RackNo,
			&b.QuantityAvailable, &b.CostPrice, &b.MRP, &b.UnitPrice,
			&b.CGSTRate, &b.SGSTRate, &b.TotalTaxPercentage,
			&b.RetailDiscPerc, &b.StaffDiscPerc, &b.SpecialDiscPerc, &b.MaxDiscPerc,
			&b.CreatedAt, &b.UpdatedAt, &b.SupplierID, &supplierName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan batch: %w", err)
		}

		if medicineName.Valid {
			b.MedicineName = medicineName.String
		}
		if medicineBrand.Valid {
			b.MedicineBrand = medicineBrand.String
		}
		if supplierName.Valid {
			b.SupplierName = supplierName.String
		}

		batches = append(batches, b)
	}

	return batches, total, nil
}

func (r *postgresRepository) ListSellableBatches(ctx context.Context, pharmacyID uuid.UUID, search string, limit int) ([]Batch, error) {
	query := `
		SELECT 
			b.id, b.pharmacy_id, b.medicine_id, m.name as medicine_name, m.brand_name as medicine_brand, b.batch_no, b.mfg_date, b.expiry_date, b.rack_no,
			b.quantity_available, b.cost_price, b.mrp, b.unit_price,
			b.cgst_rate, b.sgst_rate, b.total_tax_percentage,
			b.retail_disc_perc, b.staff_disc_perc, b.special_disc_perc, b.max_disc_perc,
			b.created_at, b.updated_at, b.supplier_id, s.name as supplier_name
		FROM inventory.batches b
		LEFT JOIN inventory.medicines m ON b.medicine_id = m.id
		LEFT JOIN supplier_schema.suppliers s ON b.supplier_id = s.id
		WHERE b.pharmacy_id = $1 
		  AND b.quantity_available > 0 
		  AND b.expiry_date > CURRENT_DATE
	`
	args := []interface{}{pharmacyID}
	placeholderID := 2

	if search != "" {
		query += fmt.Sprintf(" AND (m.name ILIKE $%d OR b.batch_no ILIKE $%d)", placeholderID, placeholderID)
		args = append(args, "%"+search+"%")
		placeholderID++
	}

	// Always prioritize batches expiring soonest for sales (FEFO)
	query += fmt.Sprintf(" ORDER BY b.expiry_date ASC LIMIT $%d", placeholderID)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sellable batches: %w", err)
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		var b Batch
		var medicineName sql.NullString
		var medicineBrand sql.NullString
		var supplierName sql.NullString
		err := rows.Scan(
			&b.ID, &b.PharmacyID, &b.MedicineID, &medicineName, &medicineBrand, &b.BatchNo, &b.MfgDate, &b.ExpiryDate, &b.RackNo,
			&b.QuantityAvailable, &b.CostPrice, &b.MRP, &b.UnitPrice,
			&b.CGSTRate, &b.SGSTRate, &b.TotalTaxPercentage,
			&b.RetailDiscPerc, &b.StaffDiscPerc, &b.SpecialDiscPerc, &b.MaxDiscPerc,
			&b.CreatedAt, &b.UpdatedAt, &b.SupplierID, &supplierName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch: %w", err)
		}

		if medicineName.Valid {
			b.MedicineName = medicineName.String
		}
		if medicineBrand.Valid {
			b.MedicineBrand = medicineBrand.String
		}
		if supplierName.Valid {
			b.SupplierName = supplierName.String
		}

		batches = append(batches, b)
	}

	return batches, nil
}

func (r *postgresRepository) GetStats(ctx context.Context, pharmacyID uuid.UUID) (BatchStats, error) {
	var stats BatchStats
	query := `
		SELECT 
			COALESCE(SUM(quantity_available), 0) as total_stocks,
			COALESCE(SUM(quantity_available * cost_price), 0) as total_stock_value,
			COUNT(id) FILTER (WHERE quantity_available <= 0) as out_of_stock,
			COUNT(id) FILTER (WHERE expiry_date < CURRENT_DATE AND quantity_available > 0) as expired_stock,
			COUNT(id) FILTER (WHERE expiry_date >= CURRENT_DATE AND expiry_date < (CURRENT_DATE + INTERVAL '30 days') AND quantity_available > 0) as expiring_soon,
			COALESCE(SUM(quantity_available * cost_price) FILTER (WHERE expiry_date >= CURRENT_DATE AND expiry_date < (CURRENT_DATE + INTERVAL '30 days') AND quantity_available > 0), 0) as expiring_soon_value,
			COUNT(id) FILTER (WHERE expiry_date >= CURRENT_DATE AND expiry_date < (CURRENT_DATE + INTERVAL '7 days') AND quantity_available > 0) as high_risk_count
		FROM inventory.batches
		WHERE pharmacy_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, pharmacyID).Scan(
		&stats.TotalStocks,
		&stats.TotalStockValue,
		&stats.OutOfStock,
		&stats.ExpiredStock,
		&stats.ExpiringSoon,
		&stats.ExpiringSoonValue,
		&stats.HighRiskCount,
	)
	if err != nil {
		return stats, fmt.Errorf("failed to get batch stats: %w", err)
	}

	return stats, nil
}

func (r *postgresRepository) CreateBatchLog(ctx context.Context, tx *sql.Tx, log BatchAuditLog) error {
	query := `
		INSERT INTO inventory.batch_audit_logs (
			id, pharmacy_id, batch_id, action_type, changed_by, changed_by_name, notes, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	id := log.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query,
			id, log.PharmacyID, log.BatchID, log.ActionType,
			log.ChangedBy, log.ChangedByName, log.Notes, time.Now(),
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			id, log.PharmacyID, log.BatchID, log.ActionType,
			log.ChangedBy, log.ChangedByName, log.Notes, time.Now(),
		)
	}
	return err
}

func (r *postgresRepository) GetBatchAuditLogs(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]BatchAuditLog, error) {
	query := `
		SELECT id, pharmacy_id, batch_id, action_type, changed_by, changed_by_name, notes, changed_at
		FROM inventory.batch_audit_logs
		WHERE pharmacy_id = $1 AND batch_id = $2
		ORDER BY changed_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to list batch audit logs: %w", err)
	}
	defer rows.Close()

	var logs []BatchAuditLog
	for rows.Next() {
		var log BatchAuditLog
		var changedBy uuid.UUID
		var changedByName sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&log.ID, &log.PharmacyID, &log.BatchID, &log.ActionType,
			&changedBy, &changedByName, &notes, &log.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch audit log: %w", err)
		}
		log.ChangedBy = changedBy
		log.ChangedByName = changedByName.String
		log.Notes = notes.String
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *postgresRepository) Update(ctx context.Context, tx *sql.Tx, pharmacyID, batchID uuid.UUID, req EditBatchRequest) error {
	query := `
		UPDATE inventory.batches 
		SET 
			rack_no = $3, mrp = $4, unit_price = $5,
			cgst_rate = $6, sgst_rate = $7, total_tax_percentage = $8,
			retail_disc_perc = $9, staff_disc_perc = $10, special_disc_perc = $11, max_disc_perc = $12,
			updated_at = CURRENT_TIMESTAMP
		WHERE pharmacy_id = $1 AND id = $2
	`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query,
			pharmacyID, batchID,
			req.RackNo, req.MRP, req.UnitPrice,
			req.CGSTRate, req.SGSTRate, req.TotalTaxPercentage,
			req.RetailDiscPerc, req.StaffDiscPerc, req.SpecialDiscPerc, req.MaxDiscPerc,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			pharmacyID, batchID,
			req.RackNo, req.MRP, req.UnitPrice,
			req.CGSTRate, req.SGSTRate, req.TotalTaxPercentage,
			req.RetailDiscPerc, req.StaffDiscPerc, req.SpecialDiscPerc, req.MaxDiscPerc,
		)
	}
	return err
}

func (r *postgresRepository) GetBatch(ctx context.Context, pharmacyID, batchID uuid.UUID) (*Batch, error) {
	query := `
		SELECT 
			id, pharmacy_id, medicine_id, batch_no, mfg_date, expiry_date, rack_no,
			quantity_available, cost_price, mrp, unit_price,
			cgst_rate, sgst_rate, total_tax_percentage,
			retail_disc_perc, staff_disc_perc, special_disc_perc, max_disc_perc,
			supplier_id, created_at, updated_at
		FROM inventory.batches
		WHERE id = $1 AND pharmacy_id = $2
	`
	b := &Batch{}
	err := r.db.QueryRowContext(ctx, query, batchID, pharmacyID).Scan(
		&b.ID, &b.PharmacyID, &b.MedicineID, &b.BatchNo, &b.MfgDate, &b.ExpiryDate, &b.RackNo,
		&b.QuantityAvailable, &b.CostPrice, &b.MRP, &b.UnitPrice,
		&b.CGSTRate, &b.SGSTRate, &b.TotalTaxPercentage,
		&b.RetailDiscPerc, &b.StaffDiscPerc, &b.SpecialDiscPerc, &b.MaxDiscPerc,
		&b.SupplierID, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("batch not found")
	}
	return b, err
}

func (r *postgresRepository) UpdateBatchQuantity(ctx context.Context, tx *sql.Tx, pharmacyID, batchID uuid.UUID, quantityChange int) error {
	query := `
		UPDATE inventory.batches
		SET quantity_available = quantity_available + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND pharmacy_id = $3
	`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, quantityChange, batchID, pharmacyID)
	} else {
		_, err = r.db.ExecContext(ctx, query, quantityChange, batchID, pharmacyID)
	}
	return err
}

func (r *postgresRepository) AddReturnStock(ctx context.Context, tx *sql.Tx, dto UpdateBatchDTO, batchID uuid.UUID) error {
	// 1. Update ONLY quantity
	query := `
		UPDATE inventory.batches 
		SET quantity_available = quantity_available + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND pharmacy_id = $3
		RETURNING quantity_available
	`

	var balanceAfter int
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, dto.QuantityToAdd, batchID, dto.PharmacyID).Scan(&balanceAfter)
	} else {
		err = r.db.QueryRowContext(ctx, query, dto.QuantityToAdd, batchID, dto.PharmacyID).Scan(&balanceAfter)
	}

	if err != nil {
		return fmt.Errorf("failed to update batch quantity: %w", err)
	}

	// 2. Record in Ledger (Audit Trail)
	ledgerQuery := `
		INSERT INTO inventory.stock_ledger (
			pharmacy_id, medicine_id, batch_id, transaction_type, 
			quantity_change, balance_after, reference_type, 
			reference_id, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if tx != nil {
		_, err = tx.ExecContext(ctx, ledgerQuery,
			dto.PharmacyID, dto.MedicineID, batchID, dto.TransactionType,
			dto.QuantityToAdd, balanceAfter, dto.ReferenceType,
			dto.ReferenceID, dto.PerformedBy, dto.Notes,
		)
	} else {
		_, err = r.db.ExecContext(ctx, ledgerQuery,
			dto.PharmacyID, dto.MedicineID, batchID, dto.TransactionType,
			dto.QuantityToAdd, balanceAfter, dto.ReferenceType,
			dto.ReferenceID, dto.PerformedBy, dto.Notes,
		)
	}

	return err
}

func (r *postgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
