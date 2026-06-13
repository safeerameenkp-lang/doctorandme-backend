package prescriptions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p *Prescription) error
	GetByID(ctx context.Context, pharmacyID uuid.UUID, id string) (*Prescription, error)
	List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Prescription, int, SalesHistoryStats, error)
	UpdateStatus(ctx context.Context, pharmacyID uuid.UUID, id string, status string) error
	UpdateLatestSaleID(ctx context.Context, pharmacyID uuid.UUID, id string, saleID uuid.UUID) error
	UpdateBillingInfo(ctx context.Context, pharmacyID uuid.UUID, id string, amount float64, method string, handledBy string, invoiceNo string) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, p *Prescription) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO sales_schema.prescriptions (id, pharmacy_id, token_no, patient_name, patient_phone, doctor_name, date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query, p.ID, p.PharmacyID, p.TokenNo, p.PatientName, p.PatientPhone, p.DoctorName, p.Date, p.Status)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO sales_schema.prescription_items (
			id, prescription_id, product_id, medicine_name, medicine_brand, quantity, instructions,
			duration_days, dosage_per_day, morning, noon, night
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	for _, item := range p.Items {
		_, err = tx.ExecContext(
			ctx, itemQuery,
			item.ID, p.ID, item.ProductID, item.MedicineName, item.MedicineBrand, item.Quantity, item.Instructions,
			item.DurationDays, item.DosagePerDay, item.Morning, item.Noon, item.Night,
		)
		if err != nil {
			return err
		}
	}

	p.TotalMedicines = len(p.Items)
	return tx.Commit()
}

func (r *postgresRepository) GetByID(ctx context.Context, pharmacyID uuid.UUID, id string) (*Prescription, error) {
	query := `
		SELECT id, pharmacy_id, token_no, patient_name, patient_phone, doctor_name, date, status, bill_amount, payment_method, handled_by_name, latest_sale_id, invoice_number
		FROM sales_schema.prescriptions
		WHERE id = $1 AND pharmacy_id = $2
	`
	p := &Prescription{}
	var phone sql.NullString
	var token sql.NullString
	var patientName sql.NullString
	var doctorName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, pharmacyID).Scan(
		&p.ID, &p.PharmacyID, &token, &patientName, &phone, &doctorName, &p.Date, &p.Status, &p.BillAmount, &p.PaymentMethod, &p.HandledByName, &p.LatestSaleID, &p.InvoiceNumber,
	)
	p.PatientPhone = phone.String
	p.TokenNo = token.String
	p.PatientName = patientName.String
	p.DoctorName = doctorName.String

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prescription not found")
	}
	if err != nil {
		return nil, err
	}

	itemQuery := `
		SELECT id, prescription_id, product_id, medicine_name, medicine_brand, quantity, instructions,
		       duration_days, dosage_per_day, morning, noon, night
		FROM sales_schema.prescription_items
		WHERE prescription_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item PrescriptionItem
		var instr sql.NullString
		if err := rows.Scan(
			&item.ID, &item.PrescriptionID, &item.ProductID, &item.MedicineName, &item.MedicineBrand, &item.Quantity, &instr,
			&item.DurationDays, &item.DosagePerDay, &item.Morning, &item.Noon, &item.Night,
		); err != nil {
			return nil, err
		}
		item.Instructions = instr.String
		p.Items = append(p.Items, item)
	}

	p.TotalMedicines = len(p.Items)
	return p, nil
}

func (r *postgresRepository) List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Prescription, int, SalesHistoryStats, error) {
	// 1. Get Today's Stats
	var stats SalesHistoryStats
	statsQuery := `
		SELECT
			(
				SELECT COUNT(id) FROM sales_schema.prescriptions 
				WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date
			) + 
			(
				SELECT COUNT(id) FROM sales_schema.sales 
				WHERE pharmacy_id = $1 AND sale_type = 'WALK_IN' AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
			) +
			(
				SELECT COUNT(id) FROM sales_schema.sales_returns 
				WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
			) as total_sales,

			COALESCE(
				(
					SELECT SUM(COALESCE(bill_amount, 0)) FROM sales_schema.prescriptions 
					WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date
				), 0
			) + 
			COALESCE(
				(
					SELECT SUM(COALESCE(total_amount, 0)) FROM sales_schema.sales 
					WHERE pharmacy_id = $1 AND sale_type = 'WALK_IN' AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
				), 0
			) -
			COALESCE(
				(
					SELECT SUM(COALESCE(total_refund, 0)) FROM sales_schema.sales_returns 
					WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
				), 0
			) as total_amount,

			(
				SELECT COUNT(id) FROM sales_schema.prescriptions 
				WHERE pharmacy_id = $1 AND status = 'PENDING' AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date
			) +
			(
				SELECT COUNT(id) FROM sales_schema.sales 
				WHERE pharmacy_id = $1 AND sale_type = 'WALK_IN' AND status = 'PENDING' AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
			) as pending_sales,

			(
				SELECT COUNT(id) FROM sales_schema.prescriptions 
				WHERE pharmacy_id = $1 AND status IN ('DISPENSED', 'DISPATCHED') AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date
			) + 
			(
				SELECT COUNT(id) FROM sales_schema.sales 
				WHERE pharmacy_id = $1 AND sale_type = 'WALK_IN' AND status IN ('COMPLETED', 'DISPATCHED') AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
			) +
			(
				SELECT COUNT(id) FROM sales_schema.sales_returns 
				WHERE pharmacy_id = $1 AND status = 'RETURNED' AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date
			) as completed_sales
	`
	if err := r.db.QueryRowContext(ctx, statsQuery, pharmacyID).Scan(&stats.TotalSales, &stats.TotalAmount, &stats.PendingSales, &stats.CompletedSales); err != nil {
		return nil, 0, stats, err
	}

	// 2. Get Total Count (Unified)
	var total int
	countQuery := `
		SELECT 
			(SELECT COUNT(id) FROM sales_schema.prescriptions WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date) + 
			(SELECT COUNT(id) FROM sales_schema.sales WHERE pharmacy_id = $1 AND sale_type = 'WALK_IN' AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date) +
			(SELECT COUNT(id) FROM sales_schema.sales_returns WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', created_at)::date = timezone('Asia/Kolkata', now())::date)
	`
	if err := r.db.QueryRowContext(ctx, countQuery, pharmacyID).Scan(&total); err != nil {
		return nil, 0, stats, err
	}

	// 3. Get Unified List
	query := `
		SELECT 
			id, pharmacy_id, token_no, patient_name, patient_phone, doctor_name, date, status, 
			bill_amount, payment_method, handled_by_name, latest_sale_id, invoice_number
		FROM (
			SELECT 
				id::text, pharmacy_id, token_no, patient_name, patient_phone, doctor_name, date, status, 
				bill_amount, payment_method, handled_by_name, latest_sale_id, invoice_number
			FROM sales_schema.prescriptions
			WHERE pharmacy_id = $1 AND timezone('Asia/Kolkata', date)::date = timezone('Asia/Kolkata', now())::date

			UNION ALL

			SELECT 
				s.id::text, s.pharmacy_id, 'WALK-IN' as token_no, COALESCE(s.customer_name, 'Walk-in Customer') as patient_name, 
				COALESCE(s.customer_phone, '-') as patient_phone, 'Self/Walk-In' as doctor_name, s.created_at as date, 
				s.status, s.total_amount as bill_amount, 
				(SELECT mode FROM sales_schema.payments WHERE sale_id = s.id LIMIT 1) as payment_method,
				NULL as handled_by_name, NULL as latest_sale_id, s.invoice_number
			FROM sales_schema.sales s
			WHERE s.pharmacy_id = $1 AND s.sale_type = 'WALK_IN' AND timezone('Asia/Kolkata', s.created_at)::date = timezone('Asia/Kolkata', now())::date

			UNION ALL

			SELECT 
				sr.id::text, sr.pharmacy_id, 'RETURN' as token_no, 
				COALESCE(s.customer_name, 'Walk-in Customer') as patient_name, 
				COALESCE(s.customer_phone, '-') as patient_phone, 
				CONCAT('Orig Ref: ', COALESCE(s.invoice_number, 'N/A')) as doctor_name, 
				sr.created_at as date, 'RETURNED' as status, 
				-sr.total_refund as bill_amount, 
				COALESCE((SELECT mode FROM sales_schema.payments WHERE return_id = sr.id LIMIT 1), 'REFUND') as payment_method,
				sr.handled_by as handled_by_name, 
				sr.sale_id as latest_sale_id, 
				sr.return_number as invoice_number
			FROM sales_schema.sales_returns sr
			JOIN sales_schema.sales s ON sr.sale_id = s.id
			WHERE sr.pharmacy_id = $1 AND timezone('Asia/Kolkata', sr.created_at)::date = timezone('Asia/Kolkata', now())::date
		) AS unified
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		return nil, 0, stats, err
	}
	defer rows.Close()

	var prescriptions []Prescription
	for rows.Next() {
		var p Prescription
		var phone sql.NullString
		var token sql.NullString
		var patientName sql.NullString
		var doctorName sql.NullString
		if err := rows.Scan(
			&p.ID, &p.PharmacyID, &token, &patientName, &phone, &doctorName, &p.Date, &p.Status,
			&p.BillAmount, &p.PaymentMethod, &p.HandledByName, &p.LatestSaleID, &p.InvoiceNumber,
		); err != nil {
			return nil, 0, stats, err
		}
		p.PatientPhone = phone.String
		p.TokenNo = token.String
		p.PatientName = patientName.String
		p.DoctorName = doctorName.String
		prescriptions = append(prescriptions, p)
	}

	return prescriptions, total, stats, nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, pharmacyID uuid.UUID, id string, status string) error {
	query := `UPDATE sales_schema.prescriptions SET status = $1 WHERE id = $2 AND pharmacy_id = $3`
	_, err := r.db.ExecContext(ctx, query, status, id, pharmacyID)
	return err
}

func (r *postgresRepository) UpdateLatestSaleID(ctx context.Context, pharmacyID uuid.UUID, id string, saleID uuid.UUID) error {
	query := `UPDATE sales_schema.prescriptions SET latest_sale_id = $1 WHERE id = $2 AND pharmacy_id = $3`
	_, err := r.db.ExecContext(ctx, query, saleID, id, pharmacyID)
	return err
}

func (r *postgresRepository) UpdateBillingInfo(ctx context.Context, pharmacyID uuid.UUID, id string, amount float64, method string, handledBy string, invoiceNo string) error {
	query := `
		UPDATE sales_schema.prescriptions 
		SET bill_amount = $1, payment_method = $2, handled_by_name = $3, invoice_number = $4
		WHERE id = $5 AND pharmacy_id = $6
	`
	_, err := r.db.ExecContext(ctx, query, amount, method, handledBy, invoiceNo, id, pharmacyID)
	return err
}
