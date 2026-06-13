package stockin

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository defines the interface for stock-in database operations
type Repository interface {
	// Updated signature to accept a callback for atomic batch updates
	CreatePurchase(ctx context.Context, purchase *Purchase, items []PurchaseItem, onTx func(tx *sql.Tx) error) error
	GetPurchaseByID(ctx context.Context, pharmacyID, purchaseID uuid.UUID) (*Purchase, []PurchaseItem, error)
	ListPurchases(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Purchase, int, error)
	CheckPurchaseExists(ctx context.Context, pharmacyID, supplierID uuid.UUID, invoiceNo string) (bool, error)
	GetStockInStats(ctx context.Context, pharmacyID uuid.UUID) (*StockInStats, error)
	CreateLog(ctx context.Context, tx *sql.Tx, log *StockInAuditLog) error
	GetAuditLogs(ctx context.Context, stockInID, pharmacyID uuid.UUID) ([]*StockInAuditLog, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new instance of the stock-in repository
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreatePurchase(ctx context.Context, purchase *Purchase, items []PurchaseItem, onTx func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on failure
	defer tx.Rollback()

	// 1. Insert Purchase Header
	// Note: due_amount is a generated column, so we do not insert it.
	queryHeader := `
		INSERT INTO inventory.purchases (
			id, pharmacy_id, supplier_id, invoice_no, purchase_date, 
			received_by, grand_total, paid_amount, payment_status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, queryHeader,
		purchase.ID, purchase.PharmacyID, purchase.SupplierID, purchase.InvoiceNo, 
		purchase.PurchaseDate, purchase.ReceivedBy, purchase.GrandTotal, 
		purchase.PaidAmount, purchase.PaymentStatus, purchase.Notes,
	)
	if err != nil {
		return fmt.Errorf("failed to insert purchase header: %w", err)
	}

	// 2. Insert Purchase Items
	queryItem := `
		INSERT INTO inventory.purchase_items (
			id, purchase_id, pharmacy_id, medicine_id, batch_no, mfg_date, expiry_date, rack_no,
			unit_mode, units_per_mode, received_qty, bonus_qty, total_qty_units, base_unit,
			purchase_price_per_mode, mrp_per_mode, 
			cgst_rate, sgst_rate, total_tax_percentage,
			retail_discount_percentage, staff_discount_percentage, special_discount_percentage, max_discount_percentage,
			cost_price_per_mode, cost_price_per_unit, item_total_amount
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	for _, item := range items {
		_, err = tx.ExecContext(ctx, queryItem,
			item.ID, item.PurchaseID, item.PharmacyID, item.MedicineID, item.BatchNo, item.MfgDate, item.ExpiryDate, item.RackNo,
			item.UnitMode, item.UnitsPerMode, item.ReceivedQty, item.BonusQty, item.TotalQtyUnits, item.BaseUnit,
			item.PurchasePricePerMode, item.MRPPerMode,
			item.CGSTRate, item.SGSTRate, item.TotalTaxPercentage,
			item.RetailDiscountPercentage, item.StaffDiscountPercentage, item.SpecialDiscountPercentage, item.MaxDiscountPercentage,
			item.CostPricePerMode, item.CostPricePerUnit, item.ItemTotalAmount,
		)
		if err != nil {
			return fmt.Errorf("failed to insert purchase item (%s): %w", item.MedicineID, err)
		}
	}

	// 3. Execute Callback (Batch Updates)
	if onTx != nil {
		if err := onTx(tx); err != nil {
			// Rollback handled by defer if commit not reached, but good to be explicit/safe
			return fmt.Errorf("failed to execute batch updates: %w", err)
		}
	}

	// 4. Commit Transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetPurchaseByID(ctx context.Context, pharmacyID, purchaseID uuid.UUID) (*Purchase, []PurchaseItem, error) {
	// 1. Get Header
	queryHeader := `
		SELECT p.id, p.pharmacy_id, p.supplier_id, s.name as supplier_name, p.invoice_no, p.purchase_date, p.received_by, 
		       p.grand_total, p.paid_amount, p.due_amount, p.payment_status, 
			   p.notes, p.created_at, p.updated_at
		FROM inventory.purchases p
		LEFT JOIN supplier_schema.suppliers s ON p.supplier_id = s.id
		WHERE p.id = $1 AND p.pharmacy_id = $2
	`
	var p Purchase
	err := r.db.QueryRowContext(ctx, queryHeader, purchaseID, pharmacyID).Scan(
		&p.ID, &p.PharmacyID, &p.SupplierID, &p.SupplierName, &p.InvoiceNo, &p.PurchaseDate, &p.ReceivedBy, 
		&p.GrandTotal, &p.PaidAmount, &p.DueAmount, &p.PaymentStatus,
		&p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("purchase not found")
		}
		return nil, nil, fmt.Errorf("failed to fetch purchase: %w", err)
	}

	// 2. Get Items
	queryItems := `
		SELECT 
			pi.id, pi.purchase_id, pi.pharmacy_id, pi.medicine_id, m.name as medicine_name, m.brand_name as medicine_brand,
			pi.batch_no, pi.mfg_date, pi.expiry_date, pi.rack_no,
			pi.unit_mode, pi.units_per_mode, pi.received_qty, pi.bonus_qty, pi.total_qty_units, pi.base_unit,
			pi.purchase_price_per_mode, pi.mrp_per_mode, 
			pi.cgst_rate, pi.sgst_rate, pi.total_tax_percentage,
			pi.retail_discount_percentage, pi.staff_discount_percentage, pi.special_discount_percentage, pi.max_discount_percentage,
			pi.cost_price_per_mode, pi.cost_price_per_unit, pi.item_total_amount, pi.created_at
		FROM inventory.purchase_items pi
		LEFT JOIN inventory.medicines m ON pi.medicine_id = m.id
		WHERE pi.purchase_id = $1 AND pi.pharmacy_id = $2
	`
	rows, err := r.db.QueryContext(ctx, queryItems, purchaseID, pharmacyID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch purchase items: %w", err)
	}
	defer rows.Close()

	var items []PurchaseItem
	for rows.Next() {
		var i PurchaseItem
		err := rows.Scan(
			&i.ID, &i.PurchaseID, &i.PharmacyID, &i.MedicineID, &i.MedicineName, &i.MedicineBrand,
			&i.BatchNo, &i.MfgDate, &i.ExpiryDate, &i.RackNo,
			&i.UnitMode, &i.UnitsPerMode, &i.ReceivedQty, &i.BonusQty, &i.TotalQtyUnits, &i.BaseUnit,
			&i.PurchasePricePerMode, &i.MRPPerMode,
			&i.CGSTRate, &i.SGSTRate, &i.TotalTaxPercentage,
			&i.RetailDiscountPercentage, &i.StaffDiscountPercentage, &i.SpecialDiscountPercentage, &i.MaxDiscountPercentage,
			&i.CostPricePerMode, &i.CostPricePerUnit, &i.ItemTotalAmount, &i.CreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan purchase item: %w", err)
		}
		items = append(items, i)
	}

	return &p, items, nil
}

func (r *postgresRepository) ListPurchases(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Purchase, int, error) {
	// First get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM inventory.purchases WHERE pharmacy_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, pharmacyID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count purchases: %w", err)
	}

	query := `
		SELECT p.id, p.pharmacy_id, p.supplier_id, COALESCE(s.name, 'Unknown') as supplier_name, p.invoice_no, p.purchase_date, p.received_by, 
		       p.grand_total, p.paid_amount, p.due_amount, p.payment_status, p.created_at
		FROM inventory.purchases p
		LEFT JOIN supplier_schema.suppliers s ON p.supplier_id = s.id
		WHERE p.pharmacy_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchases: %w", err)
	}
	defer rows.Close()

	var purchases []Purchase
	for rows.Next() {
		var p Purchase
		err := rows.Scan(
			&p.ID, &p.PharmacyID, &p.SupplierID, &p.SupplierName, &p.InvoiceNo, &p.PurchaseDate, &p.ReceivedBy, 
			&p.GrandTotal, &p.PaidAmount, &p.DueAmount, &p.PaymentStatus, &p.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan purchase: %w", err)
		}
		purchases = append(purchases, p)
	}

	return purchases, total, nil
}

func (r *postgresRepository) CheckPurchaseExists(ctx context.Context, pharmacyID, supplierID uuid.UUID, invoiceNo string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM inventory.purchases 
			WHERE pharmacy_id = $1 AND supplier_id = $2 AND invoice_no = $3
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, pharmacyID, supplierID, invoiceNo).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check purchase existence: %w", err)
	}
	return exists, nil
}
func (r *postgresRepository) GetStockInStats(ctx context.Context, pharmacyID uuid.UUID) (*StockInStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(grand_total), 0) as total_amount,
			COALESCE(SUM(paid_amount), 0) as paid_amount,
			COALESCE(SUM(due_amount), 0) as due_amount
		FROM inventory.purchases
		WHERE pharmacy_id = $1
	`
	var stats StockInStats
	err := r.db.QueryRowContext(ctx, query, pharmacyID).Scan(&stats.TotalAmount, &stats.PaidAmount, &stats.DueAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stock-in stats: %w", err)
	}
	return &stats, nil
}

func (r *postgresRepository) CreateLog(ctx context.Context, tx *sql.Tx, log *StockInAuditLog) error {
	query := `
		INSERT INTO inventory.stock_in_audit_logs (
			id, pharmacy_id, stock_in_id, action_type, changed_by, changed_by_name, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.ExecContext(ctx, query,
		log.ID, log.PharmacyID, log.StockInID, log.ActionType, log.ChangedBy, log.ChangedByName, log.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock in audit log: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetAuditLogs(ctx context.Context, stockInID, pharmacyID uuid.UUID) ([]*StockInAuditLog, error) {
	query := `
		SELECT 
			id, pharmacy_id, stock_in_id, action_type, changed_by, changed_by_name, changed_at
		FROM inventory.stock_in_audit_logs
		WHERE stock_in_id = $1 AND pharmacy_id = $2
		ORDER BY changed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, stockInID, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*StockInAuditLog
	for rows.Next() {
		log := &StockInAuditLog{}
		err := rows.Scan(
			&log.ID, &log.PharmacyID, &log.StockInID,
			&log.ActionType, &log.ChangedBy, &log.ChangedByName, &log.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("audit log row iteration error: %w", err)
	}

	return logs, nil
}
