package sales

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateSale(ctx context.Context, s *Sale) error
	GetSaleByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Sale, error)
	UpdateSale(ctx context.Context, s *Sale) error
	GetLatestSaleByPrescriptionID(ctx context.Context, pharmacyID uuid.UUID, rxID string) (*Sale, error)

	AddItem(ctx context.Context, item *SaleItem) error
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*SaleItem, error)
	GetItemsBySaleID(ctx context.Context, saleID uuid.UUID) ([]SaleItem, error)
	UpdateItem(ctx context.Context, item *SaleItem) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error

	AddPayment(ctx context.Context, p *Payment) error
	GetPaymentsBySaleID(ctx context.Context, saleID uuid.UUID) ([]Payment, error)

	UpsertPatient(ctx context.Context, p *Patient) error
	GetPatient(ctx context.Context, pharmacyID uuid.UUID, phone, name string) (*Patient, error)
	UpdatePatientWallet(ctx context.Context, pharmacyID, patientID uuid.UUID, action string, amount float64) error
	ClearPatientBalances(ctx context.Context, pharmacyID, patientID uuid.UUID) error
	GetPatientByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Patient, error)
	SearchPatientsByPhone(ctx context.Context, pharmacyID uuid.UUID, phone string) ([]Patient, error)
	ListPatients(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, search string) ([]Patient, int, error)
	GetPatientSales(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error)
	GetPatientReturns(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error)

	GetStats(ctx context.Context, pharmacyID uuid.UUID, targetDate, startDate, endDate time.Time, granularity string) (*SalesStats, error)

	CreateReturn(ctx context.Context, ret *SaleReturn, items []SaleReturnItem) error
	GetReturnByID(ctx context.Context, pharmacyID, id uuid.UUID) (*SaleReturn, error)
	ListReturns(ctx context.Context, pharmacyID uuid.UUID) ([]SaleReturn, error)
	GetReturnsBySaleID(ctx context.Context, pharmacyID, saleID uuid.UUID) ([]SaleReturn, error)
	UpdateReturnedQuantity(ctx context.Context, tx *sql.Tx, itemID uuid.UUID, quantity int) error
	GetPatientStats(ctx context.Context, pharmacyID uuid.UUID) (*PatientStats, error)
	ListSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, startDate, endDate time.Time, paymentMode, search string) ([]Sale, int, error)
	GetRecurringRefillsReport(ctx context.Context, pharmacyID uuid.UUID) ([]RecurringRefillReportItem, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateSale(ctx context.Context, s *Sale) error {
	query := `
		INSERT INTO sales_schema.sales (
			id, pharmacy_id, sale_type, prescription_id, patient_id, customer_name, customer_phone, 
			customer_age, customer_gender, status, gross_amount, total_amount, total_discount, 
			total_tax, is_recurring, days_supply, next_refill_date, applied_credit, applied_due, generated_credit, generated_due, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.PharmacyID, s.SaleType, s.PrescriptionID, s.PatientID, s.CustomerName, s.CustomerPhone,
		s.CustomerAge, s.CustomerGender, s.Status, s.GrossAmount, s.TotalAmount, s.TotalDiscount,
		s.TotalTax, s.IsRecurring, s.DaysSupply, s.NextRefillDate, s.AppliedCredit, s.AppliedDue, s.GeneratedCredit, s.GeneratedDue, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *postgresRepository) GetSaleByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Sale, error) {
	query := `
		SELECT id, pharmacy_id, sale_type, prescription_id, patient_id, 
		       COALESCE(customer_name, '') as customer_name, 
		       COALESCE(customer_phone, '') as customer_phone, 
		       COALESCE(customer_age, 0) as customer_age, 
		       COALESCE(customer_gender, '') as customer_gender, 
		       COALESCE(customer_address, '') as customer_address, 
		       status, gross_amount, total_amount, 
		       total_discount, total_tax, COALESCE(invoice_number, '') as invoice_number, 
		       is_recurring, days_supply, next_refill_date, 
		       applied_credit, applied_due, generated_credit, generated_due, created_at, updated_at
		FROM sales_schema.sales
		WHERE id = $1 AND pharmacy_id = $2
	`
	s := &Sale{}
	err := r.db.QueryRowContext(ctx, query, id, pharmacyID).Scan(
		&s.ID, &s.PharmacyID, &s.SaleType, &s.PrescriptionID, &s.PatientID,
		&s.CustomerName, &s.CustomerPhone, &s.CustomerAge, &s.CustomerGender, &s.CustomerAddress,
		&s.Status, &s.GrossAmount, &s.TotalAmount, &s.TotalDiscount, &s.TotalTax,
		&s.InvoiceNumber, &s.IsRecurring, &s.DaysSupply, &s.NextRefillDate,
		&s.AppliedCredit, &s.AppliedDue, &s.GeneratedCredit, &s.GeneratedDue, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sale not found")
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *postgresRepository) GetLatestSaleByPrescriptionID(ctx context.Context, pharmacyID uuid.UUID, rxID string) (*Sale, error) {
	query := `
		SELECT id, pharmacy_id, sale_type, prescription_id, patient_id, 
		       COALESCE(customer_name, '') as customer_name, 
		       COALESCE(customer_phone, '') as customer_phone, 
		       COALESCE(customer_age, 0) as customer_age, 
		       COALESCE(customer_gender, '') as customer_gender, 
		       COALESCE(customer_address, '') as customer_address, 
		       status, gross_amount, total_amount, 
		       total_discount, total_tax, COALESCE(invoice_number, '') as invoice_number, 
		       is_recurring, days_supply, next_refill_date, 
		       applied_credit, applied_due, generated_credit, generated_due, created_at, updated_at
		FROM sales_schema.sales
		WHERE prescription_id = $1 AND pharmacy_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	s := &Sale{}
	err := r.db.QueryRowContext(ctx, query, rxID, pharmacyID).Scan(
		&s.ID, &s.PharmacyID, &s.SaleType, &s.PrescriptionID, &s.PatientID,
		&s.CustomerName, &s.CustomerPhone, &s.CustomerAge, &s.CustomerGender, &s.CustomerAddress,
		&s.Status, &s.GrossAmount, &s.TotalAmount, &s.TotalDiscount, &s.TotalTax,
		&s.InvoiceNumber, &s.IsRecurring, &s.DaysSupply, &s.NextRefillDate,
		&s.AppliedCredit, &s.AppliedDue, &s.GeneratedCredit, &s.GeneratedDue, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *postgresRepository) UpdateSale(ctx context.Context, s *Sale) error {
	query := `
		UPDATE sales_schema.sales
		SET status = $1, gross_amount = $2, total_amount = $3, total_discount = $4, total_tax = $5, 
		    invoice_number = $6, is_recurring = $7, days_supply = $8, next_refill_date = $9, applied_credit = $10, applied_due = $11, generated_credit = $12, generated_due = $13, updated_at = $14
		WHERE id = $15 AND pharmacy_id = $16
	`
	var invNum sql.NullString
	if s.InvoiceNumber != "" {
		invNum = sql.NullString{String: s.InvoiceNumber, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		s.Status, s.GrossAmount, s.TotalAmount, s.TotalDiscount, s.TotalTax,
		invNum, s.IsRecurring, s.DaysSupply, s.NextRefillDate, s.AppliedCredit, s.AppliedDue, s.GeneratedCredit, s.GeneratedDue, time.Now(), s.ID, s.PharmacyID,
	)
	return err
}

func (r *postgresRepository) AddItem(ctx context.Context, i *SaleItem) error {
	query := `
		INSERT INTO sales_schema.sale_items (
			id, sale_id, product_id, medicine_name, medicine_brand, batch_id, batch_no, 
			quantity, expiry_date, mrp, price, discount_percentage, 
			tax_percentage, subtotal, reservation_id, rack_no, created_at,
			retail_disc_perc, staff_disc_perc, special_disc_perc, max_disc_perc
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`
	_, err := r.db.ExecContext(ctx, query,
		i.ID, i.SaleID, i.ProductID, i.MedicineName, i.MedicineBrand, i.BatchID, i.BatchNo,
		i.Quantity, i.ExpiryDate, i.MRP, i.Price, i.DiscountPercentage,
		i.TaxPercentage, i.Subtotal, i.ReservationID, i.RackNo, i.CreatedAt,
		i.RetailDiscPerc, i.StaffDiscPerc, i.SpecialDiscPerc, i.MaxDiscPerc,
	)
	return err
}

func (r *postgresRepository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, medicine_name, medicine_brand, batch_id, batch_no, 
		       quantity, expiry_date, mrp, price, 
		       discount_percentage, tax_percentage, subtotal, 
		       reservation_id, rack_no, created_at,
		       retail_disc_perc, staff_disc_perc, special_disc_perc, max_disc_perc, returned_quantity
		FROM sales_schema.sale_items
		WHERE id = $1
	`
	i := &SaleItem{}
	var batchNo sql.NullString
	var expiryDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&i.ID, &i.SaleID, &i.ProductID, &i.MedicineName, &i.MedicineBrand, &i.BatchID, &batchNo,
		&i.Quantity, &expiryDate, &i.MRP, &i.Price,
		&i.DiscountPercentage, &i.TaxPercentage, &i.Subtotal, &i.ReservationID, &i.RackNo,
		&i.CreatedAt, &i.RetailDiscPerc, &i.StaffDiscPerc, &i.SpecialDiscPerc, &i.MaxDiscPerc,
		&i.ReturnedQuantity,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	if batchNo.Valid {
		i.BatchNo = batchNo.String
	}
	if expiryDate.Valid {
		i.ExpiryDate = expiryDate.Time
	}
	return i, err
}

func (r *postgresRepository) GetItemsBySaleID(ctx context.Context, saleID uuid.UUID) ([]SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, medicine_name, medicine_brand, batch_id, batch_no, 
		       quantity, expiry_date, mrp, price, 
		       discount_percentage, tax_percentage, subtotal, 
		       reservation_id, rack_no, created_at,
		       retail_disc_perc, staff_disc_perc, special_disc_perc, max_disc_perc, returned_quantity
		FROM sales_schema.sale_items
		WHERE sale_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SaleItem
	for rows.Next() {
		var i SaleItem
		var batchNo sql.NullString
		var expiryDate sql.NullTime

		if err := rows.Scan(
			&i.ID, &i.SaleID, &i.ProductID, &i.MedicineName, &i.MedicineBrand, &i.BatchID, &batchNo,
			&i.Quantity, &expiryDate, &i.MRP, &i.Price,
			&i.DiscountPercentage, &i.TaxPercentage, &i.Subtotal, &i.ReservationID, &i.RackNo,
			&i.CreatedAt, &i.RetailDiscPerc, &i.StaffDiscPerc, &i.SpecialDiscPerc, &i.MaxDiscPerc,
			&i.ReturnedQuantity,
		); err != nil {
			return nil, err
		}

		if batchNo.Valid {
			i.BatchNo = batchNo.String
		}
		if expiryDate.Valid {
			i.ExpiryDate = expiryDate.Time
		}

		items = append(items, i)
	}
	return items, nil
}

func (r *postgresRepository) UpdateItem(ctx context.Context, i *SaleItem) error {
	query := `
		UPDATE sales_schema.sale_items
		SET quantity = $1, price = $2, discount_percentage = $3, subtotal = $4,
		    retail_disc_perc = $5, staff_disc_perc = $6, special_disc_perc = $7, max_disc_perc = $8
		WHERE id = $9
	`
	_, err := r.db.ExecContext(ctx, query, i.Quantity, i.Price, i.DiscountPercentage, i.Subtotal, i.RetailDiscPerc, i.StaffDiscPerc, i.SpecialDiscPerc, i.MaxDiscPerc, i.ID)
	return err
}

func (r *postgresRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	query := `DELETE FROM sales_schema.sale_items WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, itemID)
	return err
}

func (r *postgresRepository) AddPayment(ctx context.Context, p *Payment) error {
	query := `
		INSERT INTO sales_schema.payments (id, sale_id, return_id, transaction_type, mode, amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.SaleID, p.ReturnID, p.TransactionType, p.Mode, p.Amount, p.CreatedAt)
	return err
}

func (r *postgresRepository) GetPaymentsBySaleID(ctx context.Context, saleID uuid.UUID) ([]Payment, error) {
	query := `
		SELECT id, sale_id, return_id, transaction_type, mode, amount, created_at
		FROM sales_schema.payments
		WHERE sale_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.SaleID, &p.ReturnID, &p.TransactionType, &p.Mode, &p.Amount, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *postgresRepository) UpsertPatient(ctx context.Context, p *Patient) error {
	query := `
		INSERT INTO sales_schema.patients (id, pharmacy_id, name, phone, gender, age, address, is_recurring, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, 0), NULLIF($7, ''), $8, $9)
		ON CONFLICT (pharmacy_id, phone, name) DO UPDATE 
		SET 
			gender = COALESCE(NULLIF(EXCLUDED.gender, ''), sales_schema.patients.gender),
			age = CASE WHEN EXCLUDED.age != 0 THEN EXCLUDED.age ELSE sales_schema.patients.age END,
			address = COALESCE(NULLIF(EXCLUDED.address, ''), sales_schema.patients.address),
			is_recurring = EXCLUDED.is_recurring OR sales_schema.patients.is_recurring,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query, p.ID, p.PharmacyID, p.Name, p.Phone, p.Gender, p.Age, p.Address, p.IsRecurring, time.Now()).Scan(&p.ID)
}

func (r *postgresRepository) GetPatient(ctx context.Context, pharmacyID uuid.UUID, phone, name string) (*Patient, error) {
	// Use ILIKE for case insensitive exact matching
	query := `SELECT id, pharmacy_id, name, phone, gender, age, address, is_recurring, due_amount, credit_amount, created_at, updated_at FROM sales_schema.patients WHERE pharmacy_id = $1 AND phone = $2 AND name ILIKE $3`
	p := &Patient{}
	var addr sql.NullString
	var gender sql.NullString
	var age sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, pharmacyID, phone, name).Scan(
		&p.ID, &p.PharmacyID, &p.Name, &p.Phone, &gender, &age, &addr, &p.IsRecurring, &p.DueAmount, &p.CreditAmount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if addr.Valid {
		p.Address = addr.String
	}
	if gender.Valid {
		p.Gender = gender.String
	}
	if age.Valid {
		p.Age = int(age.Int64)
	}
	return p, nil
}

func (r *postgresRepository) UpdatePatientWallet(ctx context.Context, pharmacyID, patientID uuid.UUID, action string, amount float64) error {
	var query string
	if action == "Credit Amount" || action == "CREDIT" {
		query = `UPDATE sales_schema.patients SET credit_amount = COALESCE(credit_amount, 0) + $1, updated_at = CURRENT_TIMESTAMP WHERE pharmacy_id = $2 AND id = $3`
	} else if action == "Due Amount" || action == "DUE" {
		query = `UPDATE sales_schema.patients SET due_amount = COALESCE(due_amount, 0) + $1, updated_at = CURRENT_TIMESTAMP WHERE pharmacy_id = $2 AND id = $3`
	} else {
		return nil
	}

	_, err := r.db.ExecContext(ctx, query, amount, pharmacyID, patientID)
	return err
}

func (r *postgresRepository) ClearPatientBalances(ctx context.Context, pharmacyID, patientID uuid.UUID) error {
	query := `UPDATE sales_schema.patients SET credit_amount = 0, due_amount = 0, updated_at = CURRENT_TIMESTAMP WHERE pharmacy_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, pharmacyID, patientID)
	return err
}

func (r *postgresRepository) SearchPatientsByPhone(ctx context.Context, pharmacyID uuid.UUID, phone string) ([]Patient, error) {
	query := `
		SELECT id, pharmacy_id, name, phone, gender, age, address, is_recurring, due_amount, credit_amount, created_at, updated_at 
		FROM sales_schema.patients 
		WHERE pharmacy_id = $1 AND phone ILIKE $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, phone+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []Patient
	for rows.Next() {
		var p Patient
		var addr sql.NullString
		var gender sql.NullString
		var age sql.NullInt64

		if err := rows.Scan(
			&p.ID, &p.PharmacyID, &p.Name, &p.Phone, &gender, &age, &addr, &p.IsRecurring, &p.DueAmount, &p.CreditAmount, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if addr.Valid {
			p.Address = addr.String
		}
		if gender.Valid {
			p.Gender = gender.String
		}
		if age.Valid {
			p.Age = int(age.Int64)
		}

		patients = append(patients, p)
	}
	return patients, nil
}

func (r *postgresRepository) ListPatients(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, search string) ([]Patient, int, error) {
	// 1. Get Count
	countQuery := `SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1`
	countArgs := []interface{}{pharmacyID}
	if search != "" {
		countQuery += ` AND (name ILIKE $2 OR phone ILIKE $2)`
		countArgs = append(countArgs, "%"+search+"%")
	}
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2. Get List
	query := `
		SELECT id, pharmacy_id, name, phone, gender, age, address, is_recurring, due_amount, credit_amount, created_at, updated_at 
		FROM sales_schema.patients 
		WHERE pharmacy_id = $1
	`
	args := []interface{}{pharmacyID}

	if search != "" {
		query += ` AND (name ILIKE $2 OR phone ILIKE $2)`
		args = append(args, "%"+search+"%")
		query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []Patient
	for rows.Next() {
		var p Patient
		var addr sql.NullString
		var gender sql.NullString
		var age sql.NullInt64

		if err := rows.Scan(
			&p.ID, &p.PharmacyID, &p.Name, &p.Phone, &gender, &age, &addr, &p.IsRecurring, &p.DueAmount, &p.CreditAmount, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if addr.Valid {
			p.Address = addr.String
		}
		if gender.Valid {
			p.Gender = gender.String
		}
		if age.Valid {
			p.Age = int(age.Int64)
		}

		patients = append(patients, p)
	}
	return patients, total, nil
}

func (r *postgresRepository) GetPatientByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Patient, error) {
	query := `SELECT id, pharmacy_id, name, phone, gender, age, address, is_recurring, due_amount, credit_amount, created_at, updated_at FROM sales_schema.patients WHERE pharmacy_id = $1 AND id = $2`
	p := &Patient{}
	var addr sql.NullString
	var gender sql.NullString
	var age sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, pharmacyID, id).Scan(
		&p.ID, &p.PharmacyID, &p.Name, &p.Phone, &gender, &age, &addr, &p.IsRecurring, &p.DueAmount, &p.CreditAmount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("patient not found")
	}
	if err != nil {
		return nil, err
	}

	if addr.Valid {
		p.Address = addr.String
	}
	if gender.Valid {
		p.Gender = gender.String
	}
	if age.Valid {
		p.Age = int(age.Int64)
	}
	return p, nil
}

func (r *postgresRepository) GetPatientSales(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM sales_schema.sales WHERE pharmacy_id = $1 AND patient_id = $2`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, pharmacyID, patientID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated purchases
	query := `
		SELECT 
			s.id, 
			COALESCE(s.invoice_number, ''), 
			s.total_amount, 
			s.created_at, 
			'Pharmacist'
		FROM sales_schema.sales s
		WHERE s.pharmacy_id = $1 AND s.patient_id = $2
		ORDER BY s.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var purchases []PatientPurchase
	for rows.Next() {
		var p PatientPurchase
		if err := rows.Scan(&p.ID, &p.InvoiceNumber, &p.TotalAmount, &p.BilledDate, &p.DoneBy); err != nil {
			return nil, 0, err
		}
		purchases = append(purchases, p)
	}
	return purchases, total, nil
}

func (r *postgresRepository) GetPatientReturns(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM sales_schema.sales_returns sr
		JOIN sales_schema.sales s ON sr.sale_id = s.id
		WHERE sr.pharmacy_id = $1 AND s.patient_id = $2
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, pharmacyID, patientID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated purchases (returns)
	query := `
		SELECT 
			sr.id, 
			sr.return_number, 
			sr.total_refund, 
			sr.created_at, 
			COALESCE(sr.handled_by, 'Pharmacist')
		FROM sales_schema.sales_returns sr
		JOIN sales_schema.sales s ON sr.sale_id = s.id
		WHERE sr.pharmacy_id = $1 AND s.patient_id = $2
		ORDER BY sr.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var purchases []PatientPurchase
	for rows.Next() {
		var p PatientPurchase
		if err := rows.Scan(&p.ID, &p.InvoiceNumber, &p.TotalAmount, &p.BilledDate, &p.DoneBy); err != nil {
			return nil, 0, err
		}
		purchases = append(purchases, p)
	}
	return purchases, total, nil
}

func (r *postgresRepository) GetPatientStats(ctx context.Context, pharmacyID uuid.UUID) (*PatientStats, error) {
	stats := &PatientStats{}

	// Total Patients
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1`,
		pharmacyID).Scan(&stats.TotalPatients)
	if err != nil {
		return nil, err
	}

	// New Patients Today
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1 AND DATE(created_at) = CURRENT_DATE`,
		pharmacyID).Scan(&stats.NewPatientsToday)
	if err != nil {
		return nil, err
	}

	// Recurring Patients
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1 AND is_recurring = true`,
		pharmacyID).Scan(&stats.RecurringPatients)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *postgresRepository) CreateReturn(ctx context.Context, ret *SaleReturn, items []SaleReturnItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO sales_schema.sales_returns (
			id, pharmacy_id, sale_id, return_number, status, total_refund, reason, handled_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, query,
		ret.ID, ret.PharmacyID, ret.SaleID, ret.ReturnNumber, ret.Status, ret.TotalRefund, ret.Reason, ret.HandledBy, ret.CreatedAt, ret.UpdatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO sales_schema.sales_return_items (
			id, return_id, sale_item_id, product_id, medicine_name, batch_id, batch_no, quantity, refund_amount, condition, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.ReturnID, item.SaleItemID, item.ProductID, item.MedicineName, item.BatchID, item.BatchNo, item.Quantity, item.RefundAmount, item.Condition, item.CreatedAt,
		)
		if err != nil {
			return err
		}

		// Update returned_quantity on the original sale item
		updateQuery := `
			UPDATE sales_schema.sale_items 
			SET returned_quantity = returned_quantity + $1 
			WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, updateQuery, item.Quantity, item.SaleItemID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepository) GetReturnByID(ctx context.Context, pharmacyID, id uuid.UUID) (*SaleReturn, error) {
	query := `
		SELECT sr.id, sr.pharmacy_id, sr.sale_id, s.invoice_number, sr.return_number, sr.status, sr.total_refund, sr.reason, sr.handled_by, COALESCE(p.mode, 'CASH'), sr.created_at, sr.updated_at
		FROM sales_schema.sales_returns sr
		JOIN sales_schema.sales s ON sr.sale_id = s.id
		LEFT JOIN sales_schema.payments p ON sr.id = p.return_id AND p.transaction_type = 'REFUND'
		WHERE sr.id = $1 AND sr.pharmacy_id = $2
	`
	ret := &SaleReturn{}
	err := r.db.QueryRowContext(ctx, query, id, pharmacyID).Scan(
		&ret.ID, &ret.PharmacyID, &ret.SaleID, &ret.InvoiceNumber, &ret.ReturnNumber, &ret.Status, &ret.TotalRefund, &ret.Reason, &ret.HandledBy, &ret.RefundMode, &ret.CreatedAt, &ret.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("return not found")
	}
	if err != nil {
		return nil, err
	}

	itemQuery := `
		SELECT id, return_id, sale_item_id, product_id, medicine_name, batch_id, batch_no, quantity, refund_amount, condition, created_at
		FROM sales_schema.sales_return_items
		WHERE return_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, ret.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item SaleReturnItem
		if err := rows.Scan(
			&item.ID, &item.ReturnID, &item.SaleItemID, &item.ProductID, &item.MedicineName, &item.BatchID, &item.BatchNo, &item.Quantity, &item.RefundAmount, &item.Condition, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		ret.Items = append(ret.Items, item)
	}

	return ret, nil
}

func (r *postgresRepository) ListReturns(ctx context.Context, pharmacyID uuid.UUID) ([]SaleReturn, error) {
	query := `
		SELECT sr.id, sr.pharmacy_id, sr.sale_id, s.invoice_number, sr.return_number, sr.status, sr.total_refund, sr.reason, sr.handled_by, COALESCE(p.mode, 'CASH'), sr.created_at, sr.updated_at
		FROM sales_schema.sales_returns sr
		JOIN sales_schema.sales s ON sr.sale_id = s.id
		LEFT JOIN sales_schema.payments p ON sr.id = p.return_id AND p.transaction_type = 'REFUND'
		WHERE sr.pharmacy_id = $1
		ORDER BY sr.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var returns []SaleReturn
	for rows.Next() {
		var ret SaleReturn
		if err := rows.Scan(
			&ret.ID, &ret.PharmacyID, &ret.SaleID, &ret.InvoiceNumber, &ret.ReturnNumber, &ret.Status, &ret.TotalRefund, &ret.Reason, &ret.HandledBy, &ret.RefundMode, &ret.CreatedAt, &ret.UpdatedAt,
		); err != nil {
			return nil, err
		}
		returns = append(returns, ret)
	}
	return returns, nil
}

func (r *postgresRepository) GetReturnsBySaleID(ctx context.Context, pharmacyID, saleID uuid.UUID) ([]SaleReturn, error) {
	query := `
		SELECT sr.id, sr.pharmacy_id, sr.sale_id, s.invoice_number, sr.return_number, sr.status, sr.total_refund, sr.reason, sr.handled_by, COALESCE(p.mode, 'CASH'), sr.created_at, sr.updated_at
		FROM sales_schema.sales_returns sr
		JOIN sales_schema.sales s ON sr.sale_id = s.id
		LEFT JOIN sales_schema.payments p ON sr.id = p.return_id AND p.transaction_type = 'REFUND'
		WHERE sr.pharmacy_id = $1 AND sr.sale_id = $2
		ORDER BY sr.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var returns []SaleReturn
	for rows.Next() {
		var ret SaleReturn
		if err := rows.Scan(
			&ret.ID, &ret.PharmacyID, &ret.SaleID, &ret.InvoiceNumber, &ret.ReturnNumber, &ret.Status, &ret.TotalRefund, &ret.Reason, &ret.HandledBy, &ret.RefundMode, &ret.CreatedAt, &ret.UpdatedAt,
		); err != nil {
			return nil, err
		}
		returns = append(returns, ret)
	}
	return returns, nil
}

func (r *postgresRepository) UpdateReturnedQuantity(ctx context.Context, tx *sql.Tx, itemID uuid.UUID, quantity int) error {
	query := `UPDATE sales_schema.sale_items SET returned_quantity = returned_quantity + $1 WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, quantity, itemID)
	} else {
		_, err = r.db.ExecContext(ctx, query, quantity, itemID)
	}
	return err
}

func (r *postgresRepository) GetStats(ctx context.Context, pharmacyID uuid.UUID, targetDate, startDate, endDate time.Time, granularity string) (*SalesStats, error) {
	stats := &SalesStats{}

	// If startDate or endDate is zero, default to targetDate (single day)
	var queryStart, queryEnd time.Time
	if startDate.IsZero() || endDate.IsZero() {
		queryStart = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		queryEnd = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())
	} else {
		queryStart = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		queryEnd = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	}

	// Sales & Volume (Filtered by queryStart and queryEnd)
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*) 
		FROM sales_schema.sales 
		WHERE pharmacy_id = $1 AND status IN ('COMPLETED', 'DISPATCHED') AND created_at >= $2 AND created_at <= $3`,
		pharmacyID, queryStart, queryEnd).Scan(&stats.DailySales, &stats.SalesVolume)
	if err != nil {
		return nil, err
	}

	// Total Patients
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1`,
		pharmacyID).Scan(&stats.TotalPatients)
	if err != nil {
		return nil, err
	}

	// New Patients (Filtered by queryStart and queryEnd)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients 
		WHERE pharmacy_id = $1 AND created_at >= $2 AND created_at <= $3`,
		pharmacyID, queryStart, queryEnd).Scan(&stats.NewPatients)
	if err != nil {
		return nil, err
	}

	// Recurring Patients (Total chronic patients)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sales_schema.patients WHERE pharmacy_id = $1 AND is_recurring = true`,
		pharmacyID).Scan(&stats.RecurringPatients)
	if err != nil {
		return nil, err
	}

	// Recurring Sales (Filtered by queryStart and queryEnd)
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0) FROM sales_schema.sales 
		WHERE pharmacy_id = $1 AND is_recurring = true AND status IN ('COMPLETED', 'DISPATCHED') AND created_at >= $2 AND created_at <= $3`,
		pharmacyID, queryStart, queryEnd).Scan(&stats.RecurringSales)
	if err != nil {
		return nil, err
	}

	// Dynamic Sales Trend based on granularity
	trendQuery := fmt.Sprintf(`
		SELECT 
			d.date,
			COALESCE(SUM(s.total_amount), 0) as amount
		FROM (
			SELECT generate_series(DATE_TRUNC('%s', $1::timestamp), DATE_TRUNC('%s', $2::timestamp), '1 %s'::interval) as date
		) d
		LEFT JOIN sales_schema.sales s ON DATE_TRUNC('%s', s.created_at) = d.date 
			AND s.pharmacy_id = $3 
			AND s.status IN ('COMPLETED', 'DISPATCHED')
		GROUP BY d.date
		ORDER BY d.date ASC
	`, granularity, granularity, granularity, granularity)

	rows, err := r.db.QueryContext(ctx, trendQuery, startDate, endDate, pharmacyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d time.Time
		var amt float64
		if err := rows.Scan(&d, &amt); err != nil {
			return nil, err
		}
		stats.TrendAmounts = append(stats.TrendAmounts, amt)
		format := "02 Jan"
		if granularity == "month" {
			format = "Jan 2006"
		}
		stats.TrendDates = append(stats.TrendDates, d.Format(format))
	}

	return stats, nil
}

func (r *postgresRepository) ListSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, startDate, endDate time.Time, paymentMode, search string) ([]Sale, int, error) {
	// First query count
	countQuery := `
		SELECT COUNT(s.id)
		FROM sales_schema.sales s
		WHERE s.pharmacy_id = $1 AND s.status IN ('COMPLETED', 'DISPATCHED')
	`
	countArgs := []interface{}{pharmacyID}
	argIndex := 2

	if !startDate.IsZero() {
		countQuery += fmt.Sprintf(" AND s.created_at >= $%d", argIndex)
		countArgs = append(countArgs, startDate)
		argIndex++
	}
	if !endDate.IsZero() {
		countQuery += fmt.Sprintf(" AND s.created_at <= $%d", argIndex)
		countArgs = append(countArgs, endDate)
		argIndex++
	}
	if paymentMode != "" {
		countQuery += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM sales_schema.payments WHERE sale_id = s.id AND transaction_type = 'PAYMENT' AND mode = $%d)", argIndex)
		countArgs = append(countArgs, paymentMode)
		argIndex++
	}
	if search != "" {
		countQuery += fmt.Sprintf(" AND (s.invoice_number ILIKE $%d OR s.customer_name ILIKE $%d OR s.customer_phone ILIKE $%d)", argIndex, argIndex, argIndex)
		countArgs = append(countArgs, "%"+search+"%")
		argIndex++
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Then query data
	dataQuery := `
		SELECT 
			s.id, s.pharmacy_id, s.sale_type, s.prescription_id, s.patient_id, 
			COALESCE(s.customer_name, '') as customer_name, 
			COALESCE(s.customer_phone, '') as customer_phone, 
			COALESCE(s.customer_age, 0) as customer_age, 
			COALESCE(s.customer_gender, '') as customer_gender, 
			COALESCE(s.customer_address, '') as customer_address, 
			s.status, s.gross_amount, s.total_amount, s.total_discount, s.total_tax, 
			COALESCE(s.invoice_number, '') as invoice_number, s.is_recurring, s.days_supply, s.next_refill_date, 
			s.applied_credit, s.applied_due, s.generated_credit, s.generated_due, s.created_at, s.updated_at,
			COALESCE((SELECT mode FROM sales_schema.payments WHERE sale_id = s.id AND transaction_type = 'PAYMENT' LIMIT 1), '') as payment_mode
		FROM sales_schema.sales s
		WHERE s.pharmacy_id = $1 AND s.status IN ('COMPLETED', 'DISPATCHED')
	`
	dataArgs := []interface{}{pharmacyID}
	argIndex = 2

	if !startDate.IsZero() {
		dataQuery += fmt.Sprintf(" AND s.created_at >= $%d", argIndex)
		dataArgs = append(dataArgs, startDate)
		argIndex++
	}
	if !endDate.IsZero() {
		dataQuery += fmt.Sprintf(" AND s.created_at <= $%d", argIndex)
		dataArgs = append(dataArgs, endDate)
		argIndex++
	}
	if paymentMode != "" {
		dataQuery += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM sales_schema.payments WHERE sale_id = s.id AND transaction_type = 'PAYMENT' AND mode = $%d)", argIndex)
		dataArgs = append(dataArgs, paymentMode)
		argIndex++
	}
	if search != "" {
		dataQuery += fmt.Sprintf(" AND (s.invoice_number ILIKE $%d OR s.customer_name ILIKE $%d OR s.customer_phone ILIKE $%d)", argIndex, argIndex, argIndex)
		dataArgs = append(dataArgs, "%"+search+"%")
		argIndex++
	}

	dataQuery += fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sales []Sale
	for rows.Next() {
		var s Sale
		err := rows.Scan(
			&s.ID, &s.PharmacyID, &s.SaleType, &s.PrescriptionID, &s.PatientID,
			&s.CustomerName, &s.CustomerPhone, &s.CustomerAge, &s.CustomerGender, &s.CustomerAddress,
			&s.Status, &s.GrossAmount, &s.TotalAmount, &s.TotalDiscount, &s.TotalTax,
			&s.InvoiceNumber, &s.IsRecurring, &s.DaysSupply, &s.NextRefillDate,
			&s.AppliedCredit, &s.AppliedDue, &s.GeneratedCredit, &s.GeneratedDue, &s.CreatedAt, &s.UpdatedAt,
			&s.PaymentMode,
		)
		if err != nil {
			return nil, 0, err
		}

		sales = append(sales, s)
	}

	return sales, total, nil
}

func (r *postgresRepository) GetRecurringRefillsReport(ctx context.Context, pharmacyID uuid.UUID) ([]RecurringRefillReportItem, error) {
	query := `
		SELECT DISTINCT ON (patient_id)
			id,
			COALESCE(customer_name, '') as customer_name,
			COALESCE(invoice_number, '') as invoice_number,
			created_at as last_refill_date,
			days_supply,
			next_refill_date
		FROM sales_schema.sales
		WHERE pharmacy_id = $1 AND is_recurring = true AND status IN ('COMPLETED', 'DISPATCHED')
		ORDER BY patient_id, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, pharmacyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []RecurringRefillReportItem
	for rows.Next() {
		var item RecurringRefillReportItem
		var nextRefill sql.NullTime

		err := rows.Scan(
			&item.SaleID,
			&item.PatientName,
			&item.InvoiceNumber,
			&item.LastRefillDate,
			&item.DaysSupply,
			&nextRefill,
		)
		if err != nil {
			return nil, err
		}

		if nextRefill.Valid {
			item.NextRefillDate = &nextRefill.Time
		}

		report = append(report, item)
	}

	return report, nil
}
