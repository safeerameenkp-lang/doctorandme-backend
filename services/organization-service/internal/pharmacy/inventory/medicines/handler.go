package medicines

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"organization-service/middleware"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	svc      Service
	validate *validator.Validate
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (h *Handler) Create(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	userIDStr, userName, _ := middleware.GetUserInfo(c.Request.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid user")
		return
	}

	// Read raw body to determine if it's a single object or an array
	body, err := c.GetRawData()
	if err != nil {
		respondError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		respondError(c, http.StatusBadRequest, "Empty request body")
		return
	}

	var reqs []*CreateMedicineRequest

	// Polymorphic Check: Starts with '[' means it's an array
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(body, &reqs); err != nil {
			respondError(c, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
			return
		}
	} else {
		// Single object case
		var single CreateMedicineRequest
		if err := json.Unmarshal(body, &single); err != nil {
			respondError(c, http.StatusBadRequest, "Invalid JSON object: "+err.Error())
			return
		}
		reqs = append(reqs, &single)
	}

	if len(reqs) == 0 {
		respondError(c, http.StatusBadRequest, "At least one medicine record is required")
		return
	}

	// Validate all items
	for i, req := range reqs {
		if err := h.validate.Struct(req); err != nil {
			msg := fmt.Sprintf("Validation failed at item %d: %v", i+1, err)
			respondError(c, http.StatusBadRequest, msg)
			return
		}
	}

	meds, err := h.svc.CreateMedicines(c.Request.Context(), pharmacyID, userID, userName, reqs)
	if err != nil {
		if errors.Is(err, ErrDuplicateMedicine) {
			respondError(c, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid supplier") {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "Storage failure: "+err.Error())
		return
	}

	// For a single item, return the object. For multiple, return the array.
	if !strings.HasPrefix(trimmed, "[") && len(meds) == 1 {
		respondJSON(c, http.StatusCreated, meds[0])
	} else {
		respondJSON(c, http.StatusCreated, meds)
	}
}

func (h *Handler) GetOne(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid medicine ID")
		return
	}

	med, err := h.svc.GetMedicine(c.Request.Context(), id, pharmacyID)
	if err != nil {
		if errors.Is(err, ErrMedicineNotFound) {
			respondError(c, http.StatusNotFound, "Medicine not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondJSON(c, http.StatusOK, med)
}

func (h *Handler) GetAll(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	// Enforce pagination defaults and safety caps
	if limit <= 0 {
		limit = 25
	} else if limit > 1000 {
		limit = 1000
	}

	// Check for search filters
	search := c.Query("search")
	brandName := c.Query("brand_name")
	category := c.Query("category")
	barcode := c.Query("barcode")
	isActiveStr := c.Query("is_active")
	var isActive *bool
	if isActiveStr != "" {
		val, _ := strconv.ParseBool(isActiveStr)
		isActive = &val
	}

	hasStock, _ := strconv.ParseBool(c.Query("has_stock"))
	supplierID := c.Query("supplier_id")

	var medicines []*Medicine
	var total int
	// Use Search if any filter is provided, otherwise use List
	if search != "" || brandName != "" || category != "" || barcode != "" || supplierID != "" || isActive != nil || hasStock {
		medicines, total, err = h.svc.SearchMedicines(c.Request.Context(), pharmacyID, search, brandName, category, barcode, supplierID, isActive, hasStock, limit, offset)
	} else {
		medicines, total, err = h.svc.ListMedicines(c.Request.Context(), pharmacyID, limit, offset)
	}

	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    medicines,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *Handler) Update(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	userIDStr, userName, _ := middleware.GetUserInfo(c.Request.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid user")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(c, http.StatusBadRequest, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	med, err := h.svc.UpdateMedicine(c.Request.Context(), id, pharmacyID, userID, userName, &req)
	if err != nil {
		if errors.Is(err, ErrMedicineNotFound) {
			respondError(c, http.StatusNotFound, "Medicine not found")
			return
		}
		if errors.Is(err, ErrDuplicateMedicine) {
			respondError(c, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid supplier") {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondJSON(c, http.StatusOK, med)
}

func (h *Handler) GetStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	stats, err := h.svc.GetMedicineStats(c.Request.Context(), pharmacyID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch stats: "+err.Error())
		return
	}

	respondJSON(c, http.StatusOK, stats)
}

func (h *Handler) GetHistory(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid medicine ID")
		return
	}

	logs, err := h.svc.GetMedicineHistory(c.Request.Context(), id, pharmacyID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch audit logs: "+err.Error())
		return
	}

	respondJSON(c, http.StatusOK, logs)
}

func respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{Success: true, Data: data})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{Success: false, Error: message})
}
