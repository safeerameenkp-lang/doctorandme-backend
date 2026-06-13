package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) RecordMovement(ctx context.Context, dto RecordMovementDTO) (int, error) {
	// Start an atomic transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Lock the batch row and get current balance (FOR UPDATE ensures thread safety)
	var currentBalance int
	err = tx.QueryRowContext(ctx, 
		"SELECT quantity_available FROM inventory.batches WHERE id = $1 AND pharmacy_id = $2 FOR UPDATE",
		dto.BatchID, dto.PharmacyID).Scan(&currentBalance)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("batch not found: %w", err)
		}
		return 0, fmt.Errorf("failed to lock batch for update: %w", err)
	}

	// 2. Calculate New Balance
	newBalance := currentBalance + dto.QuantityChange
	if newBalance < 0 {
		return 0, fmt.Errorf("insufficient stock: current %d, required %d", currentBalance, -dto.QuantityChange)
	}

	// 3. Update the Batch stock level
	_, err = tx.ExecContext(ctx,
		"UPDATE inventory.batches SET quantity_available = $1, updated_at = NOW() WHERE id = $2",
		newBalance, dto.BatchID)
	if err != nil {
		return 0, fmt.Errorf("failed to update batch stock: %w", err)
	}

	// 4. Create the Ledger Audit Entry
	_, err = tx.ExecContext(ctx, `
		INSERT INTO inventory.stock_ledger (
			pharmacy_id, medicine_id, batch_id, transaction_type, 
			quantity_change, balance_after, reference_type, 
			reference_id, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		dto.PharmacyID, dto.MedicineID, dto.BatchID, dto.TransactionType,
		dto.QuantityChange, newBalance, dto.ReferenceType,
		dto.ReferenceID, dto.PerformedBy, dto.Notes)
	
	if err != nil {
		return 0, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	// 5. Commit the transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit ledger transaction: %w", err)
	}

	return newBalance, nil
}

func (r *repository) GetByBatch(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]StockLedger, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pharmacy_id, medicine_id, batch_id, transaction_type, 
		       quantity_change, balance_after, reference_type, 
		       reference_id, performed_by, notes, created_at
		FROM inventory.stock_ledger
		WHERE pharmacy_id = $1 AND batch_id = $2
		ORDER BY created_at DESC`, pharmacyID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []StockLedger
	for rows.Next() {
		var l StockLedger
		err := rows.Scan(&l.ID, &l.PharmacyID, &l.MedicineID, &l.BatchID, &l.TransactionType,
			&l.QuantityChange, &l.BalanceAfter, &l.ReferenceType,
			&l.ReferenceID, &l.PerformedBy, &l.Notes, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *repository) GetByMedicine(ctx context.Context, pharmacyID, medicineID uuid.UUID) ([]StockLedger, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pharmacy_id, medicine_id, batch_id, transaction_type, 
		       quantity_change, balance_after, reference_type, 
		       reference_id, performed_by, notes, created_at
		FROM inventory.stock_ledger
		WHERE pharmacy_id = $1 AND medicine_id = $2
		ORDER BY created_at DESC`, pharmacyID, medicineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []StockLedger
	for rows.Next() {
		var l StockLedger
		err := rows.Scan(&l.ID, &l.PharmacyID, &l.MedicineID, &l.BatchID, &l.TransactionType,
			&l.QuantityChange, &l.BalanceAfter, &l.ReferenceType,
			&l.ReferenceID, &l.PerformedBy, &l.Notes, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
