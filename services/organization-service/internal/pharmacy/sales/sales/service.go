package sales

import (
	"context"
	"fmt"
	"time"

	"organization-service/internal/pharmacy/sales/clients"

	"github.com/google/uuid"
)

type Service interface {
	CreateDraft(ctx context.Context, pharmacyID uuid.UUID, rxID string) (*Sale, error)
	CreateWalkInDraft(ctx context.Context, pharmacyID uuid.UUID, patient Patient) (*Sale, error)
	AddItemToDraft(ctx context.Context, pharmacyID, saleID uuid.UUID, req AddItemRequest) ([]SaleItem, error)
	UpdateItem(ctx context.Context, pharmacyID, saleID, itemID uuid.UUID, req UpdateItemRequest) error
	RemoveItem(ctx context.Context, pharmacyID, saleID, itemID uuid.UUID) error
	GetSaleWithDetails(ctx context.Context, pharmacyID, saleID uuid.UUID) (*Sale, error)
	FinalizeSale(ctx context.Context, pharmacyID, saleID uuid.UUID, req FinalizeSaleRequest) (*Sale, error)
	DispatchSale(ctx context.Context, pharmacyID, saleID uuid.UUID) (*Sale, error)
	SearchPatientsByPhone(ctx context.Context, pharmacyID uuid.UUID, phone string) ([]Patient, error)
	ProcessReturn(ctx context.Context, pharmacyID uuid.UUID, handledBy string, req CreateReturnRequest) (*SaleReturn, error)
	ListReturns(ctx context.Context, pharmacyID uuid.UUID) ([]SaleReturn, error)
	GetReturnDetails(ctx context.Context, pharmacyID, returnID uuid.UUID) (*SaleReturn, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID, targetDate, startDate, endDate time.Time, granularity string) (*SalesStats, error)
	ListPatients(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, search string) ([]Patient, int, error)
	GetPatientStats(ctx context.Context, pharmacyID uuid.UUID) (*PatientStats, error)
	GetPatientByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Patient, error)
	GetPatientSales(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error)
	GetPatientReturns(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error)
	ListSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, startDate, endDate time.Time, paymentMode, search string) ([]Sale, int, error)
	GetRecurringRefillsReport(ctx context.Context, pharmacyID uuid.UUID) ([]RecurringRefillReportItem, error)
}

type salesService struct {
	repo      Repository
	inventory clients.InventoryClient
	rxClient  clients.PrescriptionClient
}

func NewService(repo Repository, inv clients.InventoryClient, rx clients.PrescriptionClient) Service {
	return &salesService{
		repo:      repo,
		inventory: inv,
		rxClient:  rx,
	}
}

func (s *salesService) CreateDraft(ctx context.Context, pharmacyID uuid.UUID, rxID string) (*Sale, error) {
	// 1. Validate and Fetch Prescription
	rx, err := s.rxClient.GetPrescription(ctx, pharmacyID, rxID)
	if err != nil {
		return nil, fmt.Errorf("invalid prescription: %v", err)
	}

	// 2. Check if any sale already exists for this prescription
	existing, err := s.repo.GetLatestSaleByPrescriptionID(ctx, pharmacyID, rxID)
	if err == nil && existing != nil {
		// If it's DRAFT or PENDING, return the existing one
		if existing.Status == StatusDraft || existing.Status == StatusPending {
			// Attach prescription and items
			existing.Prescription = rx
			items, _ := s.repo.GetItemsBySaleID(ctx, existing.ID)
			for i := range items {
				availability, _ := s.inventory.GetAvailability(ctx, pharmacyID, items[i].ProductID)
				for _, b := range availability {
					if b.BatchID == items[i].BatchID {
						items[i].AvailableStock = b.Quantity
						break
					}
				}
			}
			existing.Items = items
			return existing, nil
		}

		// If it's already COMPLETED or DISPATCHED, prevent new draft
		if existing.Status == StatusCompleted || existing.Status == StatusDispatched {
			return nil, fmt.Errorf("this prescription has already been billed (Invoice: %s)", existing.InvoiceNumber)
		}
	}

	// 3. Handle Patient (Upsert from Prescription Data)
	p := Patient{
		ID:         uuid.New(),
		PharmacyID: pharmacyID,
		Name:       rx.PatientName,
		Phone:      rx.PatientPhone,
	}
	if err := s.repo.UpsertPatient(ctx, &p); err != nil {
		return nil, fmt.Errorf("failed to save patient profile: %v", err)
	}

	// 4. Create fresh draft if none exists
	sale := &Sale{
		ID:             uuid.New(),
		PharmacyID:     pharmacyID,
		SaleType:       TypeInternalRx,
		PrescriptionID: rxID,
		PatientID:      &p.ID,
		CustomerName:   p.Name,
		CustomerPhone:  p.Phone,
		CustomerAge:    p.Age,
		CustomerGender: p.Gender,
		Status:         StatusPending,
		TotalAmount:    0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Prescription:   rx,
	}

	if err := s.repo.CreateSale(ctx, sale); err != nil {
		return nil, err
	}

	// Stamp the latest_sale_id on the prescription for instant UI lookup
	_ = s.rxClient.UpdateLatestSaleID(ctx, pharmacyID, rxID, sale.ID)

	return sale, nil
}

func (s *salesService) CreateWalkInDraft(ctx context.Context, pharmacyID uuid.UUID, p Patient) (*Sale, error) {
	// 1. Handle Patient (Upsert)
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.PharmacyID = pharmacyID
	if err := s.repo.UpsertPatient(ctx, &p); err != nil {
		return nil, fmt.Errorf("failed to save patient: %v", err)
	}

	// 2. Create Sale Record
	sale := &Sale{
		ID:             uuid.New(),
		PharmacyID:     pharmacyID,
		SaleType:       TypeWalkIn,
		PatientID:      &p.ID,
		CustomerName:   p.Name,
		CustomerPhone:  p.Phone,
		CustomerAge:    p.Age,
		CustomerGender: p.Gender,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateSale(ctx, sale); err != nil {
		return nil, err
	}

	return sale, nil
}

func (s *salesService) AddItemToDraft(ctx context.Context, pharmacyID, saleID uuid.UUID, req AddItemRequest) ([]SaleItem, error) {
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return nil, err
	}

	if sale.Status != StatusDraft && sale.Status != StatusPending {
		return nil, fmt.Errorf("cannot add items to a sale that is %s", sale.Status)
	}

	// 1. Fetch FEFO availability
	batches, err := s.inventory.GetAvailability(ctx, pharmacyID, req.ProductID)
	if err != nil || len(batches) == 0 {
		return nil, fmt.Errorf("no stock available for product %s", req.ProductID)
	}

	totalNeeded := req.Quantity
	var createdItems []SaleItem

	// 2. Loop through strictly ordered batches and lock dynamically
	for _, batch := range batches {
		// Strictly enforce Sellable conditions: Ignore empty or expired batches
		if batch.Quantity <= 0 || batch.ExpiryDate.Before(time.Now()) {
			continue
		}

		if totalNeeded <= 0 {
			break
		}

		qtyToTake := batch.Quantity
		if totalNeeded < qtyToTake {
			qtyToTake = totalNeeded
		}

		// Reserve stock in Inventory Service
		resID, err := s.inventory.ReserveStock(ctx, pharmacyID, req.ProductID, batch.BatchID, qtyToTake)
		if err != nil {
			// Rollback any successfully locked reservations so far!
			for _, item := range createdItems {
				_ = s.inventory.ReleaseStock(ctx, pharmacyID, item.ReservationID)
				_ = s.repo.DeleteItem(ctx, item.ID)
			}
			return nil, fmt.Errorf("failed to reserve stock: %v", err)
		}

		// Calculate Subtotal with Tax and Discount
		mrp := batch.MRP
		discPerc := batch.RetailDiscPerc
		taxPerc := batch.TotalTaxPercentage

		// Business Math
		discountAmount := (mrp * discPerc) / 100
		taxableAmount := mrp - discountAmount
		taxAmount := (taxableAmount * taxPerc) / 100
		finalUnitPrice := taxableAmount + taxAmount
		subtotal := finalUnitPrice * float64(qtyToTake)

		item := &SaleItem{
			ID:                 uuid.New(),
			SaleID:             saleID,
			ProductID:          req.ProductID,
			MedicineName:       batch.MedicineName,
			MedicineBrand:      batch.MedicineBrand,
			BatchID:            batch.BatchID,
			BatchNo:            batch.BatchNo,
			Quantity:           qtyToTake,
			AvailableStock:     batch.Quantity - qtyToTake,
			ExpiryDate:         batch.ExpiryDate,
			MRP:                mrp,
			Price:              batch.UnitPrice, // Storing original batch unit price as reference
			DiscountPercentage: discPerc,
			TaxPercentage:      taxPerc,
			Subtotal:           subtotal,
			RetailDiscPerc:     batch.RetailDiscPerc,
			StaffDiscPerc:      batch.StaffDiscPerc,
			SpecialDiscPerc:    batch.SpecialDiscPerc,
			MaxDiscPerc:        batch.MaxDiscPerc,
			ReservationID:      resID,
			RackNo:             batch.RackNo,
			CreatedAt:          time.Now(),
		}

		if err := s.repo.AddItem(ctx, item); err != nil {
			_ = s.inventory.ReleaseStock(ctx, pharmacyID, resID)
			for _, i := range createdItems {
				_ = s.inventory.ReleaseStock(ctx, pharmacyID, i.ReservationID)
				_ = s.repo.DeleteItem(ctx, i.ID)
			}
			return nil, err
		}

		createdItems = append(createdItems, *item)
		totalNeeded -= qtyToTake
	}

	if totalNeeded > 0 {
		// Not enough combined stock
		for _, item := range createdItems {
			_ = s.inventory.ReleaseStock(ctx, pharmacyID, item.ReservationID)
			_ = s.repo.DeleteItem(ctx, item.ID)
		}
		return nil, fmt.Errorf("insufficient stock in batch. Needed %d more", totalNeeded)
	}

	// Recalculate Sale Total
	s.updateSaleTotal(ctx, pharmacyID, saleID)

	return createdItems, nil
}

func (s *salesService) UpdateItem(ctx context.Context, pharmacyID, saleID, itemID uuid.UUID, req UpdateItemRequest) error {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return err
	}

	if sale.Status != StatusDraft && sale.Status != StatusPending {
		return fmt.Errorf("cannot update items in a sale that is %s", sale.Status)
	}

	// 1. Update reservation in Inventory Service if quantity changed
	if item.Quantity != req.Quantity {
		err = s.inventory.UpdateReservation(ctx, pharmacyID, item.ReservationID, req.Quantity)
		if err != nil {
			return fmt.Errorf("failed to update reservation: %v", err)
		}
		item.Quantity = req.Quantity
	}

	// 2. Update Discount if provided
	if req.DiscountPercentage != nil {
		item.DiscountPercentage = *req.DiscountPercentage
	}

	// 3. Recalculate Subtotal
	discountAmount := (item.MRP * item.DiscountPercentage) / 100
	taxableAmount := item.MRP - discountAmount
	taxAmount := (taxableAmount * item.TaxPercentage) / 100
	finalUnitPrice := taxableAmount + taxAmount
	item.Subtotal = finalUnitPrice * float64(item.Quantity)

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return err
	}

	s.updateSaleTotal(ctx, pharmacyID, saleID)
	return nil
}

func (s *salesService) RemoveItem(ctx context.Context, pharmacyID, saleID, itemID uuid.UUID) error {
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return err
	}

	if sale.Status != StatusDraft && sale.Status != StatusPending {
		return fmt.Errorf("cannot remove items from a sale that is %s", sale.Status)
	}

	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	// Release reservation
	err = s.inventory.ReleaseStock(ctx, pharmacyID, item.ReservationID)
	if err != nil {
		return fmt.Errorf("failed to release stock: %v", err)
	}

	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		return err
	}

	s.updateSaleTotal(ctx, pharmacyID, saleID)
	return nil
}

func (s *salesService) updateSaleTotal(ctx context.Context, pharmacyID, saleID uuid.UUID) {
	items, err := s.repo.GetItemsBySaleID(ctx, saleID)
	if err != nil {
		return
	}

	var total, totalDiscount, totalTax, grossAmount float64
	for _, item := range items {
		// Business Math (Must match AddItemToDraft)
		discAmount := (item.MRP * item.DiscountPercentage) / 100
		taxableAmount := item.MRP - discAmount
		taxAmount := (taxableAmount * item.TaxPercentage) / 100

		grossAmount += item.MRP * float64(item.Quantity)
		totalDiscount += discAmount * float64(item.Quantity)
		totalTax += taxAmount * float64(item.Quantity)
		total += item.Subtotal
	}

	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err == nil {
		sale.GrossAmount = grossAmount
		sale.TotalAmount = total
		sale.TotalDiscount = totalDiscount
		sale.TotalTax = totalTax
		_ = s.repo.UpdateSale(ctx, sale)
	}
}

func (s *salesService) GetSaleWithDetails(ctx context.Context, pharmacyID, saleID uuid.UUID) (*Sale, error) {
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return nil, err
	}

	// 1. Enrich with Prescription Info
	if sale.PrescriptionID != "" {
		rx, err := s.rxClient.GetPrescription(ctx, pharmacyID, sale.PrescriptionID)
		if err == nil {
			sale.Prescription = rx
		}
	}

	// 2. Enrich with Items and Live stock
	items, err := s.repo.GetItemsBySaleID(ctx, saleID)
	if err == nil {
		for i := range items {
			availability, err := s.inventory.GetAvailability(ctx, pharmacyID, items[i].ProductID)
			if err == nil {
				for _, b := range availability {
					if b.BatchID == items[i].BatchID {
						items[i].AvailableStock = b.Quantity
						if items[i].RackNo == "N/A" || items[i].RackNo == "" {
							items[i].RackNo = b.RackNo
						}
						break
					}
				}
			}
		}
		sale.Items = items
	}

	// 3. Populate Payment Mode, Collected Amount & Handled By (Audit Info)
	payments, err := s.repo.GetPaymentsBySaleID(ctx, saleID)
	if err == nil && len(payments) > 0 {
		sale.PaymentMode = string(payments[0].Mode)
		sale.CollectedAmount = payments[0].Amount
	}
	// For now we hardcode Pharmacist until we have a full Auth integration for this field
	sale.HandledBy = "Pharmacist"

	// 4. Fetch Patient Wallet Balances
	if sale.PatientID != nil {
		patient, err := s.repo.GetPatientByID(ctx, pharmacyID, *sale.PatientID)
		if err == nil && patient != nil {
			sale.PatientDueAmount = patient.DueAmount
			sale.PatientCreditAmount = patient.CreditAmount
		}
	}

	return sale, nil
}

func (s *salesService) FinalizeSale(ctx context.Context, pharmacyID, saleID uuid.UUID, req FinalizeSaleRequest) (*Sale, error) {
	// 1. Fetch Sale and Items
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return nil, err
	}

	if sale.Status != StatusPending {
		return nil, fmt.Errorf("only pending sales can be finalized, current status: %s", sale.Status)
	}

	items, err := s.repo.GetItemsBySaleID(ctx, saleID)
	if err != nil || len(items) == 0 {
		return nil, fmt.Errorf("cannot finalize a sale with no items")
	}

	// 2. Confirm Stock in Inventory Service
	for _, item := range items {
		if item.ReservationID != "" {
			err := s.inventory.ConfirmStock(ctx, pharmacyID, item.ReservationID)
			if err != nil {
				return nil, fmt.Errorf("failed to confirm stock for %s: %v", item.MedicineName, err)
			}
		}
	}

	// 3. Generate Invoice Number
	invoiceNo := fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	sale.InvoiceNumber = invoiceNo
	sale.Status = StatusCompleted
	sale.IsRecurring = req.IsRecurring
	sale.DaysSupply = req.DaysSupply
	if sale.IsRecurring && sale.DaysSupply > 0 {
		refillDate := time.Now().AddDate(0, 0, sale.DaysSupply)
		sale.NextRefillDate = &refillDate
	}
	sale.UpdatedAt = time.Now()

	// 3b. Adjust TotalAmount based on existing Patient balances
	if sale.PatientID != nil {
		patient, err := s.repo.GetPatientByID(ctx, pharmacyID, *sale.PatientID)
		if err == nil {
			// Save the snapshotted applied values into the sale
			sale.AppliedCredit = patient.CreditAmount
			sale.AppliedDue = patient.DueAmount

			// Apply existing due and credit to the current bill
			sale.TotalAmount = sale.TotalAmount + patient.DueAmount - patient.CreditAmount
			if sale.TotalAmount < 0 {
				sale.TotalAmount = 0
			}
		}
	}

	var paymentAmount float64
	var generatedCredit float64
	var generatedDue float64

	if sale.PatientID != nil && req.WalletAction != "" && req.WalletAmount > 0 {
		collectedAmount := req.WalletAmount
		paymentAmount = collectedAmount

		if req.WalletAction == "CREDIT" {
			generatedCredit = collectedAmount - sale.TotalAmount
			if generatedCredit < 0 {
				generatedCredit = 0
			}
			sale.GeneratedCredit = generatedCredit
			sale.GeneratedDue = 0.00
		} else if req.WalletAction == "DUE" {
			generatedDue = sale.TotalAmount - collectedAmount
			if generatedDue < 0 {
				generatedDue = 0
			}
			sale.GeneratedDue = generatedDue
			sale.GeneratedCredit = 0.00
		}
	} else {
		paymentAmount = sale.TotalAmount
		sale.GeneratedCredit = 0.00
		sale.GeneratedDue = 0.00
	}

	// 4. Record Payment
	payment := &Payment{
		ID:              uuid.New(),
		SaleID:          saleID,
		TransactionType: TxTypePayment,
		Mode:            req.PaymentMode,
		Amount:          paymentAmount,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.AddPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to record payment: %v", err)
	}

	// 5. Update Sale in DB
	if err := s.repo.UpdateSale(ctx, sale); err != nil {
		return nil, fmt.Errorf("failed to update sale status: %v", err)
	}

	// 5b. Update Patient Recurring Status
	if req.IsRecurring && sale.PatientID != nil {
		p := &Patient{
			ID:          *sale.PatientID,
			PharmacyID:  sale.PharmacyID,
			Name:        sale.CustomerName,
			Phone:       sale.CustomerPhone,
			IsRecurring: true,
		}
		_ = s.repo.UpsertPatient(ctx, p)
	}

	// 5c. Update Patient Wallet
	if sale.PatientID != nil {
		// First, clear the existing balances because they were applied to the current bill's TotalAmount
		_ = s.repo.ClearPatientBalances(ctx, pharmacyID, *sale.PatientID)

		// Then, update patient wallet in DB with calculated amounts
		if req.WalletAction == "CREDIT" && generatedCredit > 0 {
			err := s.repo.UpdatePatientWallet(ctx, pharmacyID, *sale.PatientID, "CREDIT", generatedCredit)
			if err != nil {
				return nil, fmt.Errorf("failed to add new patient wallet balance: %v", err)
			}
		} else if req.WalletAction == "DUE" && generatedDue > 0 {
			err := s.repo.UpdatePatientWallet(ctx, pharmacyID, *sale.PatientID, "DUE", generatedDue)
			if err != nil {
				return nil, fmt.Errorf("failed to add new patient wallet balance: %v", err)
			}
		}
	}

	// 6. Update Prescription Status
	if sale.PrescriptionID != "" {
		_ = s.rxClient.UpdateStatus(ctx, pharmacyID, sale.PrescriptionID, "COMPLETED")
		_ = s.rxClient.UpdateBillingInfo(ctx, pharmacyID, sale.PrescriptionID, sale.TotalAmount, string(req.PaymentMode), "Pharmacist", invoiceNo)
	}

	return s.GetSaleWithDetails(ctx, pharmacyID, saleID)
}

func (s *salesService) DispatchSale(ctx context.Context, pharmacyID, saleID uuid.UUID) (*Sale, error) {
	// 1. Fetch Sale
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, saleID)
	if err != nil {
		return nil, err
	}

	// 2. Validate Status
	if sale.Status != StatusCompleted {
		return nil, fmt.Errorf("only completed sales can be dispatched, current status: %s", sale.Status)
	}

	// 3. Update Status
	sale.Status = StatusDispatched
	sale.UpdatedAt = time.Now()

	// 4. Update Sale in DB
	if err := s.repo.UpdateSale(ctx, sale); err != nil {
		return nil, fmt.Errorf("failed to update sale status: %v", err)
	}

	// 5. Update Prescription Status
	if sale.PrescriptionID != "" {
		_ = s.rxClient.UpdateStatus(ctx, pharmacyID, sale.PrescriptionID, "DISPATCHED")
	}

	return s.GetSaleWithDetails(ctx, pharmacyID, saleID)
}

func (s *salesService) SearchPatientsByPhone(ctx context.Context, pharmacyID uuid.UUID, phone string) ([]Patient, error) {
	if len(phone) < 3 {
		return []Patient{}, nil
	}
	return s.repo.SearchPatientsByPhone(ctx, pharmacyID, phone)
}

func (s *salesService) ProcessReturn(ctx context.Context, pharmacyID uuid.UUID, handledBy string, req CreateReturnRequest) (*SaleReturn, error) {
	// 1. Validate Sale
	sale, err := s.repo.GetSaleByID(ctx, pharmacyID, req.SaleID)
	if err != nil {
		return nil, fmt.Errorf("sale not found: %w", err)
	}

	if sale.Status != StatusCompleted && sale.Status != StatusDispatched {
		return nil, fmt.Errorf("only completed or dispatched sales can be returned")
	}

	// 1.5 Safety Check: Return Window (30 Days)
	if time.Since(sale.CreatedAt) > 30*24*time.Hour {
		return nil, fmt.Errorf("returns are only allowed within 30 days of the original purchase date (Sold on: %s)", sale.CreatedAt.Format("02 Jan 2006"))
	}

	// 2. Fetch original items to validate quantities
	originalItems, err := s.repo.GetItemsBySaleID(ctx, req.SaleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sale items: %w", err)
	}

	itemMap := make(map[uuid.UUID]SaleItem)
	for _, item := range originalItems {
		itemMap[item.ID] = item
	}

	// 3. Prepare Return Items and Inventory Update
	var returnItems []SaleReturnItem
	var invReturnReq []clients.ReturnItemRequest
	var totalRefund float64
	returnID := uuid.New()

	for _, reqItem := range req.Items {
		origItem, ok := itemMap[reqItem.SaleItemID]
		if !ok {
			return nil, fmt.Errorf("item %s not found in original sale", reqItem.SaleItemID)
		}

		if reqItem.Quantity+origItem.ReturnedQuantity > origItem.Quantity {
			return nil, fmt.Errorf("cannot return more than sold. Original: %d, Already Returned: %d, Attempting: %d for item %s",
				origItem.Quantity, origItem.ReturnedQuantity, reqItem.Quantity, origItem.MedicineName)
		}

		// Calculate refund for this item (proportionate to subtotal)
		itemRefund := (origItem.Subtotal / float64(origItem.Quantity)) * float64(reqItem.Quantity)
		totalRefund += itemRefund

		// Safety Check: Force DAMAGED if medicine is expired
		effectiveCondition := reqItem.Condition
		if !origItem.ExpiryDate.IsZero() && origItem.ExpiryDate.Before(time.Now()) {
			effectiveCondition = "DAMAGED"
		}

		retItem := SaleReturnItem{
			ID:           uuid.New(),
			ReturnID:     returnID,
			SaleItemID:   reqItem.SaleItemID,
			ProductID:    origItem.ProductID,
			MedicineName: origItem.MedicineName,
			BatchID:      origItem.BatchID,
			BatchNo:      origItem.BatchNo,
			Quantity:     reqItem.Quantity,
			RefundAmount: itemRefund,
			Condition:    effectiveCondition,
			CreatedAt:    time.Now(),
		}
		returnItems = append(returnItems, retItem)

		// Only add to inventory if condition is SELLABLE and NOT EXPIRED
		if effectiveCondition == "SELLABLE" {
			invReturnReq = append(invReturnReq, clients.ReturnItemRequest{
				BatchID:  origItem.BatchID,
				Quantity: reqItem.Quantity,
				Reason:   fmt.Sprintf("Return from Sale %s", sale.InvoiceNumber),
			})
		}
	}

	// 4. Update Inventory if needed
	if len(invReturnReq) > 0 {
		if err := s.inventory.ReturnItems(ctx, pharmacyID, invReturnReq); err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}
	}

	// 5. Save Return to DB
	ret := &SaleReturn{
		ID:            returnID,
		PharmacyID:    pharmacyID,
		SaleID:        req.SaleID,
		InvoiceNumber: sale.InvoiceNumber,
		ReturnNumber:  fmt.Sprintf("RET-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8]),
		Status:        ReturnStatusCompleted,
		TotalRefund:   totalRefund,
		Reason:        req.Reason,
		HandledBy:     handledBy,
		RefundMode:    req.RefundMode,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Items:         returnItems,
	}

	if err := s.repo.CreateReturn(ctx, ret, returnItems); err != nil {
		return nil, fmt.Errorf("failed to save return record: %w", err)
	}

	// 5.5 If RefundMode is CREDIT, update patient wallet
	if req.RefundMode == "CREDIT" {
		if sale.PatientID == nil {
			return nil, fmt.Errorf("cannot refund to store credit for walk-in customer with no profile")
		}
		err := s.repo.UpdatePatientWallet(ctx, pharmacyID, *sale.PatientID, "CREDIT", totalRefund)
		if err != nil {
			return nil, fmt.Errorf("failed to update patient wallet: %w", err)
		}
	}

	// 6. Record Refund in Payments table
	refundPayment := &Payment{
		ID:              uuid.New(),
		SaleID:          req.SaleID,
		ReturnID:        &returnID,
		TransactionType: TxTypeRefund,
		Mode:            req.RefundMode,
		Amount:          totalRefund,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.AddPayment(ctx, refundPayment); err != nil {
		return nil, fmt.Errorf("failed to record refund payment: %w", err)
	}

	return ret, nil
}

func (s *salesService) GetStats(ctx context.Context, pharmacyID uuid.UUID, targetDate, startDate, endDate time.Time, granularity string) (*SalesStats, error) {
	return s.repo.GetStats(ctx, pharmacyID, targetDate, startDate, endDate, granularity)
}

func (s *salesService) ListReturns(ctx context.Context, pharmacyID uuid.UUID) ([]SaleReturn, error) {
	return s.repo.ListReturns(ctx, pharmacyID)
}

func (s *salesService) GetReturnDetails(ctx context.Context, pharmacyID, returnID uuid.UUID) (*SaleReturn, error) {
	return s.repo.GetReturnByID(ctx, pharmacyID, returnID)
}

func (s *salesService) ListPatients(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, search string) ([]Patient, int, error) {
	return s.repo.ListPatients(ctx, pharmacyID, limit, offset, search)
}

func (s *salesService) GetPatientStats(ctx context.Context, pharmacyID uuid.UUID) (*PatientStats, error) {
	return s.repo.GetPatientStats(ctx, pharmacyID)
}

func (s *salesService) GetPatientByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Patient, error) {
	return s.repo.GetPatientByID(ctx, pharmacyID, id)
}

func (s *salesService) GetPatientSales(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error) {
	return s.repo.GetPatientSales(ctx, pharmacyID, patientID, limit, offset)
}

func (s *salesService) GetPatientReturns(ctx context.Context, pharmacyID, patientID uuid.UUID, limit, offset int) ([]PatientPurchase, int, error) {
	return s.repo.GetPatientReturns(ctx, pharmacyID, patientID, limit, offset)
}

func (s *salesService) ListSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int, startDate, endDate time.Time, paymentMode, search string) ([]Sale, int, error) {
	if !startDate.IsZero() && !endDate.IsZero() {
		days := endDate.Sub(startDate).Hours() / 24
		if days > 31 { // Allowing 31 days to cover full calendar month (up to 31 days)
			return nil, 0, fmt.Errorf("date range cannot exceed 30 days")
		}
	}
	return s.repo.ListSales(ctx, pharmacyID, limit, offset, startDate, endDate, paymentMode, search)
}

func (s *salesService) GetRecurringRefillsReport(ctx context.Context, pharmacyID uuid.UUID) ([]RecurringRefillReportItem, error) {
	return s.repo.GetRecurringRefillsReport(ctx, pharmacyID)
}
