package batches

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Repo() Repository // Internal use
	ListBatches(ctx context.Context, pharmacyID uuid.UUID, medicineID *uuid.UUID, limit, offset int, search, supplierID, filter string) ([]Batch, int, error)
	ListSellableBatches(ctx context.Context, pharmacyID uuid.UUID, search string, limit int) ([]Batch, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (BatchStats, error)
	GetBatchAuditLogs(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]BatchAuditLog, error)
	UpdateBatch(ctx context.Context, pharmacyID, batchID, changedBy uuid.UUID, changedByName string, req EditBatchRequest) error
	ProcessReturn(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req BatchReturnRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListBatches(ctx context.Context, pharmacyID uuid.UUID, medicineID *uuid.UUID, limit, offset int, search, supplierID, filter string) ([]Batch, int, error) {
	return s.repo.ListBatches(ctx, pharmacyID, medicineID, limit, offset, search, supplierID, filter)
}
func (s *service) ListSellableBatches(ctx context.Context, pharmacyID uuid.UUID, search string, limit int) ([]Batch, error) {
	return s.repo.ListSellableBatches(ctx, pharmacyID, search, limit)
}

func (s *service) GetStats(ctx context.Context, pharmacyID uuid.UUID) (BatchStats, error) {
	return s.repo.GetStats(ctx, pharmacyID)
}

func (s *service) GetBatchAuditLogs(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]BatchAuditLog, error) {
	return s.repo.GetBatchAuditLogs(ctx, pharmacyID, batchID)
}

func (s *service) UpdateBatch(ctx context.Context, pharmacyID, batchID, changedBy uuid.UUID, changedByName string, req EditBatchRequest) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.Update(ctx, tx, pharmacyID, batchID, req); err != nil {
		return err
	}

	auditLog := BatchAuditLog{
		ID:            uuid.New(),
		PharmacyID:    pharmacyID,
		BatchID:       batchID,
		ActionType:    "UPDATE",
		ChangedBy:     changedBy,
		ChangedByName: changedByName,
		Notes:         "Manual inventory update",
		ChangedAt:     time.Now(),
	}

	if err := s.repo.CreateBatchLog(ctx, tx, auditLog); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) ProcessReturn(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, req BatchReturnRequest) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range req.Items {
		// 1. Fetch batch to ensure it exists and get medicine_id
		batch, err := s.repo.GetBatch(ctx, pharmacyID, item.BatchID)
		if err != nil {
			return fmt.Errorf("batch %s not found: %w", item.BatchID, err)
		}

		// 2. Add stock back using UpsertBatch (ledger tracking included)
		// We pass ALL existing batch fields to ensure they are NOT overwritten by zero values
		dto := UpdateBatchDTO{
			PharmacyID:         pharmacyID,
			MedicineID:         batch.MedicineID,
			BatchNo:            batch.BatchNo,
			MfgDate:            batch.MfgDate,
			ExpiryDate:         batch.ExpiryDate,
			RackNo:             batch.RackNo,
			QuantityToAdd:      item.Quantity,
			CostPrice:          batch.CostPrice,
			MRP:                batch.MRP,
			UnitPrice:          batch.UnitPrice,
			CGSTRate:           batch.CGSTRate,
			SGSTRate:           batch.SGSTRate,
			TotalTaxPercentage: batch.TotalTaxPercentage,
			RetailDiscPerc:     batch.RetailDiscPerc,
			StaffDiscPerc:      batch.StaffDiscPerc,
			SpecialDiscPerc:    batch.SpecialDiscPerc,
			MaxDiscPerc:        batch.MaxDiscPerc,
			TransactionType:    "SALE_RETURN",
			ReferenceType:      "CUSTOMER_RETURN",
			Notes:              item.Reason,
			PerformedBy:        &userID,
		}

		if batch.SupplierID != nil {
			dto.SupplierID = *batch.SupplierID
		}

		err = s.repo.AddReturnStock(ctx, tx, dto, item.BatchID)
		if err != nil {
			return fmt.Errorf("failed to update batch %s: %w", item.BatchID, err)
		}

		// 3. Create Audit Log
		auditLog := BatchAuditLog{
			ID:            uuid.New(),
			PharmacyID:    pharmacyID,
			BatchID:       item.BatchID,
			ActionType:    "SALE_RETURN",
			ChangedBy:     userID,
			ChangedByName: userName,
			Notes:         item.Reason,
			ChangedAt:     time.Now(),
		}

		if err := s.repo.CreateBatchLog(ctx, tx, auditLog); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *service) Repo() Repository {
	return s.repo
}
