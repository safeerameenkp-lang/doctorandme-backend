package stockouts

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateStockOut(ctx context.Context, tx *sql.Tx, stockOut *StockOut, items []StockOutItem) error
	GetStockOutByID(ctx context.Context, pharmacyID, id uuid.UUID) (*StockOut, []StockOutItem, error)
	ListStockOuts(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]StockOut, int, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (StockOutStats, error)
	CreateAuditLog(ctx context.Context, tx *sql.Tx, log StockOutAuditLog) error
	GetBatchForStockOut(ctx context.Context, pharmacyID, batchID uuid.UUID) (uuid.UUID, string, string, float64, time.Time, string, uuid.UUID, error)
	GetAuditLogs(ctx context.Context, pharmacyID, stockOutID uuid.UUID) ([]StockOutAuditLog, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *postgresRepository) CreateStockOut(ctx context.Context, tx *sql.Tx, stockOut *StockOut, items []StockOutItem) error {
	// 1. Insert Master Record
	query := `
		INSERT INTO inventory.stock_outs (
			id, pharmacy_id, status, type, reason, destination_type, destination_name, destination_id,
			total_loss_value, created_by_id, created_by_name, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := tx.ExecContext(ctx, query,
		stockOut.ID, stockOut.PharmacyID, stockOut.Status, stockOut.Type, stockOut.Reason,
		stockOut.DestinationType, stockOut.DestinationName, stockOut.DestinationID, stockOut.TotalLossValue,
		stockOut.CreatedByID, stockOut.CreatedByName, stockOut.CreatedAt, stockOut.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock_out master: %w", err)
	}

	// 2. Insert Items & Update Batches
	itemQuery := `
		INSERT INTO inventory.stock_out_items (
			id, stock_out_id, pharmacy_id, medicine_id, medicine_name, 
			batch_id, batch_no, expiry_date, unit_type, quantity, 
			unit_cost_price, total_loss, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	updateBatchQuery := `
		UPDATE inventory.batches 
		SET quantity_available = quantity_available - $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND pharmacy_id = $3 AND quantity_available >= $1
		RETURNING quantity_available, medicine_id
	`

	ledgerQuery := `
		INSERT INTO inventory.stock_ledger (
			pharmacy_id, medicine_id, batch_id, transaction_type, 
			quantity_change, balance_after, reference_type, 
			reference_id, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, item := range items {
		// Insert Item
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.StockOutID, item.PharmacyID, item.MedicineID, item.MedicineName,
			item.BatchID, item.BatchNo, item.ExpiryDate, item.UnitType, item.Quantity,
			item.UnitCostPrice, item.TotalLoss, item.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stock_out item: %w", err)
		}

		// Update Batch Quantity
		var balanceAfter int
		var medicineID uuid.UUID
		err = tx.QueryRowContext(ctx, updateBatchQuery, item.Quantity, item.BatchID, item.PharmacyID).Scan(&balanceAfter, &medicineID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("insufficient stock or batch not found for batch_id %s", item.BatchID)
			}
			return fmt.Errorf("failed to update batch quantity: %w", err)
		}

		// Record in Ledger
		_, err = tx.ExecContext(ctx, ledgerQuery,
			item.PharmacyID, medicineID, item.BatchID, "STOCK_OUT",
			-item.Quantity, balanceAfter, "STOCK_OUT",
			stockOut.ID, stockOut.CreatedByID, fmt.Sprintf("Stock out type: %s", stockOut.Type),
		)
		if err != nil {
			return fmt.Errorf("failed to record stock ledger: %w", err)
		}
	}

	return nil
}

func (r *postgresRepository) GetStockOutByID(ctx context.Context, pharmacyID, id uuid.UUID) (*StockOut, []StockOutItem, error) {
	masterQuery := `
		SELECT id, pharmacy_id, status, type, reason, destination_type, destination_name, destination_id, total_loss_value, created_by_id, created_by_name, created_at, updated_at
		FROM inventory.stock_outs
		WHERE id = $1 AND pharmacy_id = $2
	`
	var so StockOut
	err := r.db.QueryRowContext(ctx, masterQuery, id, pharmacyID).Scan(
		&so.ID, &so.PharmacyID, &so.Status, &so.Type, &so.Reason, &so.DestinationType, &so.DestinationName, &so.DestinationID,
		&so.TotalLossValue, &so.CreatedByID, &so.CreatedByName, &so.CreatedAt, &so.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	itemsQuery := `
		SELECT id, stock_out_id, pharmacy_id, medicine_id, medicine_name, batch_id, batch_no, expiry_date, unit_type, quantity, unit_cost_price, total_loss, created_at
		FROM inventory.stock_out_items
		WHERE stock_out_id = $1 AND pharmacy_id = $2
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id, pharmacyID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []StockOutItem
	for rows.Next() {
		var it StockOutItem
		err := rows.Scan(
			&it.ID, &it.StockOutID, &it.PharmacyID, &it.MedicineID, &it.MedicineName,
			&it.BatchID, &it.BatchNo, &it.ExpiryDate, &it.UnitType, &it.Quantity,
			&it.UnitCostPrice, &it.TotalLoss, &it.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, it)
	}

	return &so, items, nil
}

func (r *postgresRepository) ListStockOuts(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]StockOut, int, error) {
	countQuery := `SELECT COUNT(*) FROM inventory.stock_outs WHERE pharmacy_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, pharmacyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, pharmacy_id, status, type, reason, destination_type, destination_name, total_loss_value, created_by_id, created_by_name, created_at, updated_at
		FROM inventory.stock_outs
		WHERE pharmacy_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []StockOut
	for rows.Next() {
		var so StockOut
		err := rows.Scan(
			&so.ID, &so.PharmacyID, &so.Status, &so.Type, &so.Reason, &so.DestinationType, &so.DestinationName,
			&so.TotalLossValue, &so.CreatedByID, &so.CreatedByName, &so.CreatedAt, &so.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, so)
	}

	return results, total, nil
}

func (r *postgresRepository) GetStats(ctx context.Context, pharmacyID uuid.UUID) (StockOutStats, error) {
	var stats StockOutStats
	query := `
		SELECT 
			COUNT(id) as total_entries,
			COALESCE(SUM(total_loss_value) FILTER (WHERE type IN ('DAMAGED', 'EXPIRED', 'ADJUSTMENT')), 0) as total_loss_value,
			COUNT(id) FILTER (WHERE type = 'DAMAGED') as damaged_count,
			COUNT(id) FILTER (WHERE type = 'TRANSFER') as transfer_count
		FROM inventory.stock_outs
		WHERE pharmacy_id = $1 AND status = 'COMPLETED'
	`
	err := r.db.QueryRowContext(ctx, query, pharmacyID).Scan(
		&stats.TotalEntries, &stats.TotalLossValue, &stats.DamagedCount, &stats.TransferCount,
	)
	return stats, err
}

func (r *postgresRepository) CreateAuditLog(ctx context.Context, tx *sql.Tx, log StockOutAuditLog) error {
	query := `
		INSERT INTO inventory.stock_out_audit_logs (
			id, pharmacy_id, stock_out_id, action_type, changed_by, changed_by_name, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	id := log.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, id, log.PharmacyID, log.StockOutID, log.ActionType, log.ChangedBy, log.ChangedByName, time.Now())
	} else {
		_, err = r.db.ExecContext(ctx, query, id, log.PharmacyID, log.StockOutID, log.ActionType, log.ChangedBy, log.ChangedByName, time.Now())
	}
	return err
}

func (r *postgresRepository) GetBatchForStockOut(ctx context.Context, pharmacyID, batchID uuid.UUID) (uuid.UUID, string, string, float64, time.Time, string, uuid.UUID, error) {
	query := `
		SELECT b.medicine_id, m.name, b.batch_no, b.cost_price, b.expiry_date, m.unit_type, b.supplier_id
		FROM inventory.batches b
		JOIN inventory.medicines m ON b.medicine_id = m.id
		WHERE b.id = $1 AND b.pharmacy_id = $2
	`
	var medicineID uuid.UUID
	var medicineName string
	var batchNo string
	var costPrice float64
	var expiryDate time.Time
	var unitType string
	var supplierID uuid.UUID

	err := r.db.QueryRowContext(ctx, query, batchID, pharmacyID).Scan(&medicineID, &medicineName, &batchNo, &costPrice, &expiryDate, &unitType, &supplierID)
	if err != nil {
		return uuid.Nil, "", "", 0, time.Time{}, "", uuid.Nil, err
	}
	return medicineID, medicineName, batchNo, costPrice, expiryDate, unitType, supplierID, nil
}

func (r *postgresRepository) GetAuditLogs(ctx context.Context, pharmacyID, stockOutID uuid.UUID) ([]StockOutAuditLog, error) {
	query := `
		SELECT id, pharmacy_id, stock_out_id, action_type, changed_by, changed_by_name, changed_at
		FROM inventory.stock_out_audit_logs
		WHERE pharmacy_id = $1 AND stock_out_id = $2
		ORDER BY changed_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, stockOutID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock out audit logs: %w", err)
	}
	defer rows.Close()

	var logs []StockOutAuditLog
	for rows.Next() {
		var l StockOutAuditLog
		err := rows.Scan(&l.ID, &l.PharmacyID, &l.StockOutID, &l.ActionType, &l.ChangedBy, &l.ChangedByName, &l.ChangedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock out audit log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}
