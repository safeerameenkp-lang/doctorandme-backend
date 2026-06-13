package reservations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, r *Reservation) error
	GetByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Reservation, error)
	Update(ctx context.Context, r *Reservation) error
	Delete(ctx context.Context, pharmacyID, id uuid.UUID) error
	
	// GetReservedQuantity returns the sum of all pending reservations for a batch
	GetReservedQuantity(ctx context.Context, pharmacyID, batchID uuid.UUID) (int, error)

	// PurgeOldReservations removes confirmed/cancelled reservations older than given duration
	PurgeOldReservations(ctx context.Context, olderThan time.Duration) (int64, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, res *Reservation) error {
	query := `
		INSERT INTO inventory.reservations (id, pharmacy_id, product_id, batch_id, quantity, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query, res.ID, res.PharmacyID, res.ProductID, res.BatchID, res.Quantity, res.Status, res.ExpiresAt, res.CreatedAt, res.UpdatedAt)
	return err
}

func (r *postgresRepository) GetByID(ctx context.Context, pharmacyID, id uuid.UUID) (*Reservation, error) {
	query := `
		SELECT id, pharmacy_id, product_id, batch_id, quantity, status, expires_at, created_at, updated_at
		FROM inventory.reservations
		WHERE id = $1 AND pharmacy_id = $2
	`
	res := &Reservation{}
	err := r.db.QueryRowContext(ctx, query, id, pharmacyID).Scan(
		&res.ID, &res.PharmacyID, &res.ProductID, &res.BatchID, &res.Quantity, &res.Status, &res.ExpiresAt, &res.CreatedAt, &res.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reservation not found")
	}
	return res, err
}

func (r *postgresRepository) Update(ctx context.Context, res *Reservation) error {
	query := `
		UPDATE inventory.reservations
		SET quantity = $1, status = $2, updated_at = $3
		WHERE id = $4 AND pharmacy_id = $5
	`
	_, err := r.db.ExecContext(ctx, query, res.Quantity, res.Status, time.Now(), res.ID, res.PharmacyID)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, pharmacyID, id uuid.UUID) error {
	query := `DELETE FROM inventory.reservations WHERE id = $1 AND pharmacy_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, pharmacyID)
	return err
}

func (r *postgresRepository) GetReservedQuantity(ctx context.Context, pharmacyID, batchID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM inventory.reservations
		WHERE batch_id = $1 AND pharmacy_id = $2 AND status = 'PENDING' AND expires_at > $3
	`
	var total int
	err := r.db.QueryRowContext(ctx, query, batchID, pharmacyID, time.Now()).Scan(&total)
	return total, err
}

func (r *postgresRepository) PurgeOldReservations(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM inventory.reservations
		WHERE status IN ('CONFIRMED', 'CANCELLED', 'EXPIRED')
		AND created_at < $1
	`
	cutoff := time.Now().Add(-olderThan)
	res, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
