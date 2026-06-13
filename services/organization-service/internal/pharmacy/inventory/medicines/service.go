package medicines

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

type Service interface {
	CreateMedicines(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, reqs []*CreateMedicineRequest) ([]*Medicine, error)
	GetMedicine(ctx context.Context, id, pharmacyID uuid.UUID) (*Medicine, error)
	ListMedicines(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Medicine, int, error)
	SearchMedicines(ctx context.Context, pharmacyID uuid.UUID, search, brandName, category, barcode, supplierID string, isActive *bool, hasStock bool, limit, offset int) ([]*Medicine, int, error)
	UpdateMedicine(ctx context.Context, id, pharmacyID, userID uuid.UUID, userName string, req *UpdateMedicineRequest) (*Medicine, error)
	ValidateSupplierOwnership(ctx context.Context, pharmacyID uuid.UUID, supplierID *uuid.UUID) error
	GetMedicineStats(ctx context.Context, pharmacyID uuid.UUID) (*MedicineStats, error)
	GetMedicineHistory(ctx context.Context, medicineID, pharmacyID uuid.UUID) ([]*MedicineAuditLog, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateMedicines(ctx context.Context, pharmacyID, userID uuid.UUID, userName string, reqs []*CreateMedicineRequest) ([]*Medicine, error) {
	// 1. Batch Validation of Suppliers
	supplierMap := make(map[uuid.UUID]bool)
	var supplierIDs []uuid.UUID
	for _, req := range reqs {
		if !supplierMap[req.SupplierID] {
			supplierMap[req.SupplierID] = true
			supplierIDs = append(supplierIDs, req.SupplierID)
		}
	}

	isValid, err := s.repo.ValidateSuppliers(ctx, pharmacyID, supplierIDs)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("invalid supplier: one or more suppliers do not belong to your pharmacy")
	}

	// 2. Prepare Data
	var medicines []*Medicine
	for _, req := range reqs {
		med := &Medicine{
			ID:               uuid.New(),
			PharmacyID:       pharmacyID,
			CreatedBy:        userID,
			CreatedByName:    userName,
			UpdatedBy:        userID,
			UpdatedByName:    userName,
			Name:             req.Name,
			BrandName:        req.BrandName,
			DosageForm:       req.DosageForm,
			Category:         req.Category,
			Manufacturer:     req.Manufacturer,
			MfgLicense:       req.MfgLicense,
			SupplierID:       req.SupplierID,
			HSNCode:          req.HSNCode,
			ScheduleType:     req.ScheduleType,
			IsRxRequired:     req.IsRxRequired,
			UnitType:         req.UnitType,
			Barcode:          req.Barcode,
			StorageCondition: req.StorageCondition,
			CGSTRate:         req.CGSTRate,
			SGSTRate:         req.SGSTRate,
			IsActive:         true,
		}
		medicines = append(medicines, med)
	}

	// 3. Atomic Bulk Storage
	if err := s.repo.Create(ctx, medicines); err != nil {
		return nil, err
	}

	return medicines, nil
}

func (s *service) GetMedicine(ctx context.Context, id, pharmacyID uuid.UUID) (*Medicine, error) {
	return s.repo.GetByID(ctx, id, pharmacyID)
}

func (s *service) ListMedicines(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Medicine, int, error) {
	return s.repo.List(ctx, pharmacyID, limit, offset)
}

func (s *service) SearchMedicines(ctx context.Context, pharmacyID uuid.UUID, search, brandName, category, barcode, supplierID string, isActive *bool, hasStock bool, limit, offset int) ([]*Medicine, int, error) {
	return s.repo.Search(ctx, pharmacyID, search, brandName, category, barcode, supplierID, isActive, hasStock, limit, offset)
}

func (s *service) UpdateMedicine(ctx context.Context, id, pharmacyID, userID uuid.UUID, userName string, req *UpdateMedicineRequest) (*Medicine, error) {
	existing, err := s.repo.GetByID(ctx, id, pharmacyID)
	if err != nil {
		return nil, err
	}

	// Validate supplier ownership if supplier_id is being updated
	if req.SupplierID != nil {
		if err := s.ValidateSupplierOwnership(ctx, pharmacyID, req.SupplierID); err != nil {
			return nil, err
		}
	}

	if req.Name != nil { existing.Name = *req.Name }
	if req.BrandName != nil { existing.BrandName = *req.BrandName }
	if req.DosageForm != nil { existing.DosageForm = *req.DosageForm }
	if req.Category != nil { existing.Category = *req.Category }
	if req.Manufacturer != nil { existing.Manufacturer = *req.Manufacturer }
	if req.MfgLicense != nil { existing.MfgLicense = *req.MfgLicense }
	if req.SupplierID != nil { existing.SupplierID = *req.SupplierID }
	if req.HSNCode != nil { existing.HSNCode = *req.HSNCode }
	if req.ScheduleType != nil { existing.ScheduleType = *req.ScheduleType }
	if req.IsRxRequired != nil { existing.IsRxRequired = *req.IsRxRequired }
	if req.UnitType != nil { existing.UnitType = *req.UnitType }
	if req.Barcode != nil { existing.Barcode = *req.Barcode }
	if req.StorageCondition != nil { existing.StorageCondition = *req.StorageCondition }
	if req.CGSTRate != nil { existing.CGSTRate = *req.CGSTRate }
	if req.SGSTRate != nil { existing.SGSTRate = *req.SGSTRate }
	if req.IsActive != nil { existing.IsActive = *req.IsActive }

	existing.UpdatedBy = userID
	existing.UpdatedByName = userName

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *service) GetMedicineHistory(ctx context.Context, medicineID, pharmacyID uuid.UUID) ([]*MedicineAuditLog, error) {
	return s.repo.GetAuditLogs(ctx, medicineID, pharmacyID)
}

func (s *service) GetMedicineStats(ctx context.Context, pharmacyID uuid.UUID) (*MedicineStats, error) {
	return s.repo.GetStats(ctx, pharmacyID)
}

// ValidateSupplierOwnership checks if the supplier belongs to the same pharmacy
func (s *service) ValidateSupplierOwnership(ctx context.Context, pharmacyID uuid.UUID, supplierID *uuid.UUID) error {
	// supplier_id is now required (NOT NULL in database)
	// But we still check for nil here for API validation
	if supplierID == nil {
		return fmt.Errorf("supplier_id is required")
	}

	// Check if supplier exists and belongs to this pharmacy
	isValid, err := s.repo.ValidateSuppliers(ctx, pharmacyID, []uuid.UUID{*supplierID})
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("invalid supplier: supplier does not belong to your pharmacy")
	}
	return nil
}
