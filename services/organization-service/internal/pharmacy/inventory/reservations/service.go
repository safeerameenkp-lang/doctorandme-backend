package reservations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"organization-service/internal/pharmacy/inventory/batches"
)

type Service interface {
	Reserve(ctx context.Context, pharmacyID uuid.UUID, req CreateReservationRequest) (*Reservation, error)
	Update(ctx context.Context, pharmacyID, id uuid.UUID, req UpdateReservationRequest) error
	Confirm(ctx context.Context, pharmacyID, id uuid.UUID) error
	Cancel(ctx context.Context, pharmacyID, id uuid.UUID) error
	StartPurgeWorker(ctx context.Context) // Production-grade background cleanup
}

type reservationsService struct {
	repo      Repository
	batchRepo batches.Repository
}

func NewService(repo Repository, batchRepo batches.Repository) Service {
	return &reservationsService{
		repo:      repo,
		batchRepo: batchRepo,
	}
}

func (s *reservationsService) Reserve(ctx context.Context, pharmacyID uuid.UUID, req CreateReservationRequest) (*Reservation, error) {
	// 1. Get Batch and check availability
	batch, err := s.batchRepo.GetBatch(ctx, pharmacyID, req.BatchID)
	if err != nil {
		return nil, err
	}

	// Calculate net available = Current in batch - existing active reservations
	reserved, err := s.repo.GetReservedQuantity(ctx, pharmacyID, req.BatchID)
	if err != nil {
		return nil, err
	}

	available := batch.QuantityAvailable - reserved
	if available < req.Quantity {
		return nil, fmt.Errorf("insufficient stock: %d available, %d requested", available, req.Quantity)
	}

	res := &Reservation{
		ID:         uuid.New(),
		PharmacyID: pharmacyID,
		ProductID:  req.ProductID,
		BatchID:    req.BatchID,
		Quantity:   req.Quantity,
		Status:     StatusPending,
		ExpiresAt:  time.Now().Add(1 * time.Hour), // TTL for reservation
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *reservationsService) Update(ctx context.Context, pharmacyID, id uuid.UUID, req UpdateReservationRequest) error {
	res, err := s.repo.GetByID(ctx, pharmacyID, id)
	if err != nil {
		return err
	}

	if res.Status != StatusPending {
		return fmt.Errorf("cannot update non-pending reservation")
	}

	// Check if new quantity is available
	batch, err := s.batchRepo.GetBatch(ctx, pharmacyID, res.BatchID)
	if err != nil {
		return err
	}

	reserved, _ := s.repo.GetReservedQuantity(ctx, pharmacyID, res.BatchID)
	// Subtract current reservation from the reserved sum to check headroom
	headroom := batch.QuantityAvailable - (reserved - res.Quantity)

	if headroom < req.Quantity {
		return fmt.Errorf("insufficient stock for updated quantity")
	}

	res.Quantity = req.Quantity
	return s.repo.Update(ctx, res)
}

func (s *reservationsService) Confirm(ctx context.Context, pharmacyID, id uuid.UUID) error {
	res, err := s.repo.GetByID(ctx, pharmacyID, id)
	if err != nil {
		return err
	}

	if res.Status != StatusPending {
		return fmt.Errorf("reservation is already %s", res.Status)
	}

	// Start Transaction to deduct stock for real
	tx, err := s.batchRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deduct stock from batch
	err = s.batchRepo.UpdateBatchQuantity(ctx, tx, pharmacyID, res.BatchID, -res.Quantity)
	if err != nil {
		return err
	}

	// Mark reservation as confirmed
	res.Status = StatusConfirmed
	if err := s.repo.Update(ctx, res); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *reservationsService) Cancel(ctx context.Context, pharmacyID, id uuid.UUID) error {
	res, err := s.repo.GetByID(ctx, pharmacyID, id)
	if err != nil {
		return err
	}

	res.Status = StatusCancelled
	return s.repo.Update(ctx, res)
}

// StartPurgeWorker runs in the background to keep the database clean
func (s *reservationsService) StartPurgeWorker(ctx context.Context) {
	// For production: Purge once every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	
	// Also run once immediately on startup
	go func() {
		s.runPurge(ctx)
		
		for {
			select {
			case <-ticker.C:
				s.runPurge(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *reservationsService) runPurge(ctx context.Context) {
	// We keep data for 7 days as discussed for audit/safety
	count, err := s.repo.PurgeOldReservations(ctx, 7*24*time.Hour)
	if err != nil {
		fmt.Printf("[Purge-Worker] Error: %v\n", err)
	} else if count > 0 {
		fmt.Printf("[Purge-Worker] Cleaned up %d old reservation records\n", count)
	}
}
