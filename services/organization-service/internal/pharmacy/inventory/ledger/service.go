package ledger

import (
	"context"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Record(ctx context.Context, dto RecordMovementDTO) (int, error) {
	// Standard validation can be added here
	if dto.QuantityChange == 0 {
		return 0, nil // No movement to record
	}
	
	return s.repo.RecordMovement(ctx, dto)
}

// Additional business logic methods like "GenerateMonthlyAuditReport" 
// can be added here in the future.
