package prescriptions

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, pharmacyID uuid.UUID, req CreatePrescriptionRequest) (*Prescription, error)
	Get(ctx context.Context, pharmacyID uuid.UUID, id string) (*Prescription, error)
	List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Prescription, int, SalesHistoryStats, error)
}

type prescriptionsService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &prescriptionsService{repo: repo}
}

func (s *prescriptionsService) Create(ctx context.Context, pharmacyID uuid.UUID, req CreatePrescriptionRequest) (*Prescription, error) {
	// Generate a simple human-readable ID for testing (e.g. RX-1234)
	id := fmt.Sprintf("RX-%d", time.Now().UnixNano()/1000000)

	p := &Prescription{
		ID:           id,
		PharmacyID:   pharmacyID,
		TokenNo:      req.TokenNo,
		PatientName:  req.PatientName,
		PatientPhone: req.PatientPhone,
		DoctorName:   req.DoctorName,
		Date:         time.Now(),
		Status:       "PENDING",
	}

	for _, item := range req.Items {
		quantity := item.Quantity
		dosagePerDay := item.DosagePerDay

		if dosagePerDay == 0 {
			command := item.Morning + item.Noon + item.Night
			dosagePerDay = command
		}

		if quantity == 0 && dosagePerDay > 0 && item.DurationDays > 0 {
			// Auto-calculate and round UP to nearest whole number
			calc := dosagePerDay * float64(item.DurationDays)
			quantity = int(math.Ceil(calc))
		}

		p.Items = append(p.Items, PrescriptionItem{
			ID:             uuid.New(),
			PrescriptionID: id,
			ProductID:      item.ProductID,
			MedicineName:   item.MedicineName,
			MedicineBrand:  item.MedicineBrand,
			Quantity:       quantity,
			DurationDays:   item.DurationDays,
			DosagePerDay:   dosagePerDay,
			Morning:        item.Morning,
			Noon:           item.Noon,
			Night:          item.Night,
			Instructions:   item.Instructions,
		})
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *prescriptionsService) Get(ctx context.Context, pharmacyID uuid.UUID, id string) (*Prescription, error) {
	return s.repo.GetByID(ctx, pharmacyID, id)
}

func (s *prescriptionsService) List(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]Prescription, int, SalesHistoryStats, error) {
	return s.repo.List(ctx, pharmacyID, limit, offset)
}
