package supplier

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"organization-service/middleware"
)

type SupplierService interface {
	CreateSuppliers(ctx context.Context, pharmacyID uuid.UUID, reqs []*CreateSupplierRequest) ([]*Supplier, error)
	GetSupplier(ctx context.Context, id, pharmacyID uuid.UUID) (*Supplier, error)
	ListSuppliers(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Supplier, int, error)
	SearchSuppliers(ctx context.Context, pharmacyID uuid.UUID, search string, limit, offset int) ([]*Supplier, int, error)
	UpdateSupplier(ctx context.Context, id, pharmacyID uuid.UUID, req *UpdateSupplierRequest) (*Supplier, error)
	GetStats(ctx context.Context, pharmacyID uuid.UUID) (*SupplierStats, error)
	GetHistory(ctx context.Context, id, pharmacyID uuid.UUID) ([]*SupplierAuditLog, error)
}

type supplierService struct {
	repo SupplierRepository
}

func NewSupplierService(repo SupplierRepository) SupplierService {
	return &supplierService{repo: repo}
}

func (s *supplierService) CreateSuppliers(ctx context.Context, pharmacyID uuid.UUID, reqs []*CreateSupplierRequest) ([]*Supplier, error) {
	// 1. Get User Info from Context
	userIDStr, userName, _ := middleware.GetUserInfo(ctx)
	userID, _ := uuid.Parse(userIDStr)

	var createdSuppliers []*Supplier
	now := time.Now().UTC()

	for _, req := range reqs {
		// 2. Validate and Sanitize
		if err := s.validateAndSanitize(req); err != nil {
			return nil, err
		}

		// 3. Map to model
		supplier := &Supplier{
			ID:            uuid.New(),
			PharmacyID:    pharmacyID,
			Name:          req.Name,
			SupplierType:  req.SupplierType,
			ContactPerson: req.ContactPerson,
			ContactNumber: req.ContactNumber,
			Website:       req.Website,
			Email:         req.Email,
			Address:       req.Address,
			State:         req.State,
			Pincode:       req.Pincode,
			GSTNumber:     req.GSTNumber,
			PANNumber:     req.PANNumber,
			LicenseNumber: req.LicenseNumber,
			BankDetails:   req.BankDetails,
			CreditTerms:   req.CreditTerms,
			IsActive:      true,
			CreatedAt:     now,
			UpdatedAt:     now,
			CreatedBy:     userID,
			UpdatedBy:     userID,
		}

		if err := s.repo.Create(ctx, supplier); err != nil {
			return nil, err
		}

		// Record audit log
		_ = s.repo.CreateAuditLog(ctx, &SupplierAuditLog{
			ID:            uuid.New(),
			PharmacyID:    pharmacyID,
			SupplierID:    supplier.ID,
			ActionType:    "CREATE",
			ChangedBy:     userID,
			ChangedByName: userName,
			ChangedAt:     now,
		})

		createdSuppliers = append(createdSuppliers, supplier)
	}

	return createdSuppliers, nil
}

func (s *supplierService) GetSupplier(ctx context.Context, id, pharmacyID uuid.UUID) (*Supplier, error) {
	return s.repo.FindByID(ctx, id, pharmacyID)
}

func (s *supplierService) ListSuppliers(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]*Supplier, int, error) {
	return s.repo.FindAll(ctx, pharmacyID, limit, offset)
}

func (s *supplierService) SearchSuppliers(ctx context.Context, pharmacyID uuid.UUID, search string, limit, offset int) ([]*Supplier, int, error) {
	return s.repo.Search(ctx, pharmacyID, search, limit, offset)
}

func (s *supplierService) UpdateSupplier(ctx context.Context, id, pharmacyID uuid.UUID, req *UpdateSupplierRequest) (*Supplier, error) {
	existing, err := s.repo.FindByID(ctx, id, pharmacyID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.SupplierType != nil {
		existing.SupplierType = *req.SupplierType
	}
	if req.ContactPerson != nil {
		existing.ContactPerson = *req.ContactPerson
	}
	if req.ContactNumber != nil {
		existing.ContactNumber = strings.TrimSpace(*req.ContactNumber)
	}
	if req.Website != nil {
		existing.Website = *req.Website
	}
	if req.Email != nil {
		existing.Email = strings.TrimSpace(*req.Email)
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.State != nil {
		existing.State = *req.State
	}
	if req.Pincode != nil {
		existing.Pincode = strings.TrimSpace(*req.Pincode)
	}
	if req.GSTNumber != nil {
		existing.GSTNumber = strings.TrimSpace(*req.GSTNumber)
	}
	if req.PANNumber != nil {
		existing.PANNumber = strings.TrimSpace(*req.PANNumber)
	}
	if req.LicenseNumber != nil {
		existing.LicenseNumber = strings.TrimSpace(*req.LicenseNumber)
	}

	// Granular Update for Bank Details
	if req.BankDetails != nil {
		if req.BankDetails.BankName != "" {
			existing.BankDetails.BankName = req.BankDetails.BankName
		}
		if req.BankDetails.AccountName != "" {
			existing.BankDetails.AccountName = req.BankDetails.AccountName
		}
		if req.BankDetails.AccountNumber != "" {
			existing.BankDetails.AccountNumber = strings.TrimSpace(req.BankDetails.AccountNumber)
		}
		if req.BankDetails.IFSCCode != "" {
			existing.BankDetails.IFSCCode = strings.TrimSpace(req.BankDetails.IFSCCode)
		}
	}

	// Granular Update for Credit Terms
	if req.CreditTerms != nil {
		existing.CreditTerms.CreditPeriodDays = req.CreditTerms.CreditPeriodDays
		existing.CreditTerms.CreditLimit = req.CreditTerms.CreditLimit
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Final Validation State
	validationReq := &CreateSupplierRequest{
		Name:        existing.Name,
		GSTNumber:   existing.GSTNumber,
		PANNumber:   existing.PANNumber,
		BankDetails: existing.BankDetails,
	}

	// Re-validate the final state
	if err := s.validateAndSanitize(validationReq); err != nil {
		return nil, err
	}

	// Update auditing
	userIDStr, _, _ := middleware.GetUserInfo(ctx)
	userID, _ := uuid.Parse(userIDStr)

	existing.UpdatedAt = time.Now().UTC()
	existing.UpdatedBy = userID

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// Record audit log
	_, userName, _ := middleware.GetUserInfo(ctx)
	_ = s.repo.CreateAuditLog(ctx, &SupplierAuditLog{
		ID:            uuid.New(),
		PharmacyID:    pharmacyID,
		SupplierID:    existing.ID,
		ActionType:    "UPDATE",
		ChangedBy:     userID,
		ChangedByName: userName,
		ChangedAt:     existing.UpdatedAt,
	})

	return existing, nil
}

func (s *supplierService) GetHistory(ctx context.Context, id, pharmacyID uuid.UUID) ([]*SupplierAuditLog, error) {
	return s.repo.GetHistory(ctx, id, pharmacyID)
}

func (s *supplierService) GetStats(ctx context.Context, pharmacyID uuid.UUID) (*SupplierStats, error) {
	return s.repo.GetStats(ctx, pharmacyID)
}

// Internal Helper for Validation and Sanitization
// Internal Helper for Validation and Sanitization
func (s *supplierService) validateAndSanitize(req *CreateSupplierRequest) error {
	// 1. Sanitize (Trim spaces)
	req.Name = strings.TrimSpace(req.Name)
	req.GSTNumber = strings.TrimSpace(req.GSTNumber)
	req.PANNumber = strings.TrimSpace(req.PANNumber)
	req.Email = strings.TrimSpace(req.Email)
	req.BankDetails.AccountNumber = strings.TrimSpace(req.BankDetails.AccountNumber)
	req.BankDetails.IFSCCode = strings.TrimSpace(req.BankDetails.IFSCCode)

	// 2. Fundamental Checks
	if req.Name == "" {
		return errors.New("supplier name is required")
	}

	return nil
}
