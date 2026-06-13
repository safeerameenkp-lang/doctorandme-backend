package stockouts

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateStockOut(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req CreateStockOutRequest) (*StockOut, error)
	GetStockOutDetails(ctx context.Context, pharmacyID, id uuid.UUID) (*StockOut, []StockOutItem, error)
	ListStockOuts(ctx context.Context, pharmacyID uuid.UUID, page, pageSize int) ([]StockOut, int, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (StockOutStats, error)
	GetAuditLogs(ctx context.Context, pharmacyID, stockOutID uuid.UUID) ([]StockOutAuditLog, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateStockOut(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req CreateStockOutRequest) (*StockOut, error) {
	stockOutID := uuid.New()
	now := time.Now().UTC()

	var items []StockOutItem
	var totalLossValue float64

	// 1. Process and Validate Items
	for _, reqItem := range req.Items {
		// Fetch Batch Details (Price/Names/Expiry/Unit/Supplier)
		medID, medName, batchNo, costPrice, expiryDate, unitType, itemSupplierID, err := s.repo.GetBatchForStockOut(ctx, pharmacyID, reqItem.BatchID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate batch %s: %w", reqItem.BatchID, err)
		}

		// CROSS-CHECK: If returning to a supplier, ensure it's the CORRECT supplier
		if req.DestinationType == "SUPPLIER" && req.DestinationID != nil {
			if itemSupplierID != *req.DestinationID {
				return nil, fmt.Errorf("supplier mismatch for medicine '%s': this batch was bought from a different supplier", medName)
			}
		}

		itemTotalLoss := costPrice * float64(reqItem.Quantity)
		totalLossValue += itemTotalLoss

		item := StockOutItem{
			ID:            uuid.New(),
			StockOutID:    stockOutID,
			PharmacyID:    pharmacyID,
			MedicineID:    medID,
			MedicineName:  medName,
			BatchID:       reqItem.BatchID,
			BatchNo:       batchNo,
			ExpiryDate:    expiryDate,
			UnitType:      unitType,
			Quantity:      reqItem.Quantity,
			UnitCostPrice: costPrice,
			TotalLoss:     itemTotalLoss,
			CreatedAt:     now,
		}
		items = append(items, item)
	}

	// 2. Create Master Record
	stockOut := &StockOut{
		ID:              stockOutID,
		PharmacyID:      pharmacyID,
		Status:          StatusCompleted,
		Type:            req.Type,
		Reason:          req.Reason,
		DestinationType: req.DestinationType,
		DestinationName: req.DestinationName,
		DestinationID:   req.DestinationID,
		TotalLossValue:  totalLossValue,
		CreatedByID:     userID,
		CreatedByName:   userName,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 3. Persist in Transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.repo.CreateStockOut(ctx, tx, stockOut, items); err != nil {
		return nil, err
	}

	// 4. Create Audit Log
	auditLog := StockOutAuditLog{
		ID:            uuid.New(),
		PharmacyID:    pharmacyID,
		StockOutID:    stockOutID,
		ActionType:    "CREATE",
		ChangedBy:     userID,
		ChangedByName: userName,
		ChangedAt:     now,
	}
	if err := s.repo.CreateAuditLog(ctx, tx, auditLog); err != nil {
		return nil, fmt.Errorf("failed to create stock out audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return stockOut, nil
}

func (s *service) GetStockOutDetails(ctx context.Context, pharmacyID, id uuid.UUID) (*StockOut, []StockOutItem, error) {
	return s.repo.GetStockOutByID(ctx, pharmacyID, id)
}

func (s *service) ListStockOuts(ctx context.Context, pharmacyID uuid.UUID, page, pageSize int) ([]StockOut, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.repo.ListStockOuts(ctx, pharmacyID, pageSize, offset)
}

func (s *service) GetStats(ctx context.Context, pharmacyID uuid.UUID) (StockOutStats, error) {
	return s.repo.GetStats(ctx, pharmacyID)
}

func (s *service) GetAuditLogs(ctx context.Context, pharmacyID, stockOutID uuid.UUID) ([]StockOutAuditLog, error) {
	return s.repo.GetAuditLogs(ctx, pharmacyID, stockOutID)
}
