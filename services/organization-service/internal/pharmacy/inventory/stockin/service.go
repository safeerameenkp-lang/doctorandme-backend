package stockin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"database/sql"
	"organization-service/internal/pharmacy/inventory/batches"
	"organization-service/internal/pharmacy/inventory/medicines"
)

// Service defines the business logic for stock-in operations
type Service interface {
	AddStockIn(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req CreatePurchaseRequest) (*Purchase, error)
	GetStockInDetails(ctx context.Context, pharmacyID, purchaseID uuid.UUID) (*Purchase, []PurchaseItem, error)
	ListStockIn(ctx context.Context, pharmacyID uuid.UUID, page, pageSize int) ([]Purchase, int, error)
	GetStockInStats(ctx context.Context, pharmacyID uuid.UUID) (*StockInStats, error)
	GetStockInHistory(ctx context.Context, purchaseID, pharmacyID uuid.UUID) ([]*StockInAuditLog, error)
}

type service struct {
	repo       Repository
	medRepo    medicines.Repository
	batchesSvc batches.Service
}

// NewService creates a new instance of the stock-in service
func NewService(repo Repository, medRepo medicines.Repository, batchesSvc batches.Service) Service {
	return &service{
		repo:       repo,
		medRepo:    medRepo,
		batchesSvc: batchesSvc,
	}
}

func (s *service) AddStockIn(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req CreatePurchaseRequest) (*Purchase, error) {
	// 1. Validate Supplier (Rule 1: Must belong to pharmacy and be active)
	isSupplierValid, err := s.medRepo.ValidateSuppliers(ctx, pharmacyID, []uuid.UUID{req.SupplierID})
	if err != nil {
		return nil, fmt.Errorf("error validating supplier: %w", err)
	}
	if !isSupplierValid {
		return nil, fmt.Errorf("invalid or inactive supplier for this pharmacy")
	}

	// 2. Check for Duplicate Invoice (Rule: One invoice per supplier)
	exists, err := s.repo.CheckPurchaseExists(ctx, pharmacyID, req.SupplierID, req.InvoiceNo)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("duplicate bill: invoice '%s' already exists for this supplier", req.InvoiceNo)
	}

	purchaseID := uuid.New()
	
	// Prepare items and perform calculations
	var items []PurchaseItem
	var calculatedGrandTotal float64

	for _, reqItem := range req.Items {
		// 2. Fetch/Validate Medicine (Rule 3 & 4: Must belong to pharmacy and be active)
		med, err := s.medRepo.GetByID(ctx, reqItem.MedicineID, pharmacyID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch/validate medicine (%s): %w", reqItem.MedicineID, err)
		}
		if !med.IsActive {
			return nil, fmt.Errorf("medicine %s is inactive", med.Name)
		}

		// 3. Automated Unit/Mode Logic (Rule 6, 7, 9)
		unitMode := med.UnitType // Rule 6: Fetch from medicine table
		unitsPerMode := reqItem.UnitsPerMode
		
		// Rule 7: Only Strip mode needs entry, others default to 1
		if strings.ToLower(unitMode) != "strip" {
			unitsPerMode = 1
		}

		// Rule 9: Base Unit selection
		baseUnit := "piece"
		switch strings.ToLower(unitMode) {
		case "strip":
			baseUnit = "tablets"
		case "bottle":
			baseUnit = "bottles"
		}

		// 4. Quantity Calculations (Rules for inventory tracking)
		totalQtyUnits := (reqItem.ReceivedQty + reqItem.BonusQty) * unitsPerMode

		// 5. Derived Cost Logic (New Requirement)
		// "cost unit type and cost rate need to calculate based item total amount and recieved quantity"
		// "tax and discount dont have any calculation role here"
		costPricePerMode := reqItem.ItemTotalAmount / float64(reqItem.ReceivedQty)
		costPricePerUnit := costPricePerMode / float64(unitsPerMode)

		// Accumulate for Grand Total Integrity Check
		calculatedGrandTotal += reqItem.ItemTotalAmount

		item := PurchaseItem{
			ID:                        uuid.New(),
			PurchaseID:                purchaseID,
			PharmacyID:                pharmacyID,
			MedicineID:                reqItem.MedicineID,
			BatchNo:                   reqItem.BatchNo,
			MfgDate:                   reqItem.MfgDate,
			ExpiryDate:                reqItem.ExpiryDate,
			RackNo:                    reqItem.RackNo,
			UnitMode:                  unitMode,
			UnitsPerMode:              unitsPerMode,
			ReceivedQty:               reqItem.ReceivedQty,
			BonusQty:                  reqItem.BonusQty,
			TotalQtyUnits:             totalQtyUnits,
			BaseUnit:                  baseUnit,
			PurchasePricePerMode:      reqItem.PurchasePricePerMode,
			MRPPerMode:                reqItem.MRPPerMode,
			
			// Rates are stored for billing, but not used in procurement cost math
			CGSTRate:                  med.CGSTRate,
			SGSTRate:                  med.SGSTRate,
			TotalTaxPercentage:        med.CGSTRate + med.SGSTRate,
			
			RetailDiscountPercentage:  reqItem.RetailDiscountPercentage,
			StaffDiscountPercentage:   reqItem.StaffDiscountPercentage,
			SpecialDiscountPercentage: reqItem.SpecialDiscountPercentage,
			MaxDiscountPercentage:     reqItem.MaxDiscountPercentage,
			
			CostPricePerMode:          costPricePerMode,
			CostPricePerUnit:          costPricePerUnit,
			ItemTotalAmount:           reqItem.ItemTotalAmount,
			CreatedAt:                 time.Now(),
		}
		items = append(items, item)
	}

	// Rule 5: Cross-check Grand Total for financial integrity
	// "validate the total amount of invoice is correct by adding item tamout of each items"
	diff := req.GrandTotal - calculatedGrandTotal
	if diff < -0.01 || diff > 0.01 {
		return nil, fmt.Errorf("financial integrity check failed: expected total %0.2f (sum of items), but request grand total is %0.2f", calculatedGrandTotal, req.GrandTotal)
	}

	// Create Header
	// 6. Payment Status Logic (Auto-calculated)
	paymentStatus := "unpaid"
	if req.PaidAmount >= req.GrandTotal {
		paymentStatus = "paid"
	} else if req.PaidAmount > 0 {
		paymentStatus = "partial"
	}
	
	// Due Amount is calculated by DB (Generated Column), but we can set it here for the return object
	dueAmount := req.GrandTotal - req.PaidAmount

	// Create Header
	purchase := &Purchase{
		ID:            purchaseID,
		PharmacyID:    pharmacyID,
		SupplierID:    req.SupplierID,
		InvoiceNo:     req.InvoiceNo,
		PurchaseDate:  req.PurchaseDate,
		ReceivedBy:    req.ReceivedBy,
		GrandTotal:    req.GrandTotal,
		PaidAmount:    req.PaidAmount,
		DueAmount:     dueAmount,
		PaymentStatus: paymentStatus,
		Notes:         req.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database with atomic batch updates
	err = s.repo.CreatePurchase(ctx, purchase, items, func(tx *sql.Tx) error {
		for _, item := range items {
			// Map PurchaseItem to UpdateBatchDTO
			dto := batches.UpdateBatchDTO{
				PharmacyID:      item.PharmacyID,
				MedicineID:      item.MedicineID,
				BatchNo:         item.BatchNo,
				MfgDate:         item.MfgDate,
				ExpiryDate:      item.ExpiryDate,
				RackNo:          item.RackNo,
				QuantityToAdd:   item.TotalQtyUnits, // Adding total units to stock
				
				CostPrice:       item.CostPricePerUnit,
				MRP:             item.MRPPerMode / float64(item.UnitsPerMode), // Calculate MRP per Unit
				UnitPrice:       item.MRPPerMode / float64(item.UnitsPerMode), // Using MRP as Unit Price basis for now
				
				CGSTRate:        item.CGSTRate,
				SGSTRate:        item.SGSTRate,
				TotalTaxPercentage: item.TotalTaxPercentage,
				
				RetailDiscPerc:  item.RetailDiscountPercentage,
				StaffDiscPerc:   item.StaffDiscountPercentage,
				SpecialDiscPerc: item.SpecialDiscountPercentage,
				MaxDiscPerc:     item.MaxDiscountPercentage,
				
				SupplierID:      req.SupplierID,

				// Ledger Metadata
				TransactionType: "PURCHASE",
				ReferenceType:   "PURCHASE_INVOICE",
				ReferenceID:     &purchaseID,
				Notes:           fmt.Sprintf("Stock added via Invoice %s", req.InvoiceNo),
			}
			
			batchID, err := s.batchesSvc.Repo().UpsertBatch(ctx, tx, dto)
			if err != nil {
				return err
			}

			// Create Batch Audit Log
			if err := s.batchesSvc.Repo().CreateBatchLog(ctx, tx, batches.BatchAuditLog{
				ID:            uuid.New(),
				PharmacyID:    pharmacyID,
				BatchID:       batchID,
				ActionType:    "STOCK_IN", // Action type for stock in
				ChangedBy:     userID,
				ChangedByName: userName,
				Notes:         fmt.Sprintf("Stock added via Invoice %s", req.InvoiceNo),
			}); err != nil {
				return err
			}
		}

		// Create Audit Log
		log := &StockInAuditLog{
			ID:            uuid.New(),
			PharmacyID:    pharmacyID,
			StockInID:     purchaseID,
			ActionType:    "CREATE",
			ChangedBy:     userID,
			ChangedByName: userName,
			ChangedAt:     time.Now(),
		}
		if err := s.repo.CreateLog(ctx, tx, log); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return purchase, nil
}

func (s *service) GetStockInDetails(ctx context.Context, pharmacyID, purchaseID uuid.UUID) (*Purchase, []PurchaseItem, error) {
	return s.repo.GetPurchaseByID(ctx, pharmacyID, purchaseID)
}

func (s *service) ListStockIn(ctx context.Context, pharmacyID uuid.UUID, page, pageSize int) ([]Purchase, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.repo.ListPurchases(ctx, pharmacyID, pageSize, offset)
}
func (s *service) GetStockInStats(ctx context.Context, pharmacyID uuid.UUID) (*StockInStats, error) {
	return s.repo.GetStockInStats(ctx, pharmacyID)
}

func (s *service) GetStockInHistory(ctx context.Context, purchaseID, pharmacyID uuid.UUID) ([]*StockInAuditLog, error) {
	return s.repo.GetAuditLogs(ctx, purchaseID, pharmacyID)
}
