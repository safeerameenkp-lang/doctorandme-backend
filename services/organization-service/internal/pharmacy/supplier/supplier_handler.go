package supplier

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SupplierHandler struct {
	service SupplierService
}

func NewSupplierHandler(service SupplierService) *SupplierHandler {
	return &SupplierHandler{service: service}
}

func (h *SupplierHandler) Create(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	// Read raw body to determine if it's a single object or an array
	body, err := c.GetRawData()
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		h.respondError(c, http.StatusBadRequest, "Empty request body")
		return
	}

	var reqs []*CreateSupplierRequest

	// Polymorphic Check: Starts with '[' means it's an array
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(body, &reqs); err != nil {
			h.respondError(c, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
			return
		}
	} else {
		// Single object case
		var single CreateSupplierRequest
		if err := json.Unmarshal(body, &single); err != nil {
			h.respondError(c, http.StatusBadRequest, "Invalid JSON object: "+err.Error())
			return
		}
		reqs = append(reqs, &single)
	}

	if len(reqs) == 0 {
		h.respondError(c, http.StatusBadRequest, "At least one supplier record is required")
		return
	}

	suppliers, err := h.service.CreateSuppliers(c.Request.Context(), pharmacyID, reqs)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// For a single item, return the object. For multiple, return the array.
	if !strings.HasPrefix(trimmed, "[") && len(suppliers) == 1 {
		h.respondJSON(c, http.StatusCreated, suppliers[0])
	} else {
		h.respondJSON(c, http.StatusCreated, suppliers)
	}
}

func (h *SupplierHandler) GetOne(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	supplier, err := h.service.GetSupplier(c.Request.Context(), id, pharmacyID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, supplier)
}

func (h *SupplierHandler) GetAll(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	search := c.Query("search")

	var (
		suppliers []*Supplier
		total     int
		err2      error
	)

	if search != "" {
		suppliers, total, err2 = h.service.SearchSuppliers(c.Request.Context(), pharmacyID, search, limit, offset)
	} else {
		suppliers, total, err2 = h.service.ListSuppliers(c.Request.Context(), pharmacyID, limit, offset)
	}

	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSONWithMeta(c, http.StatusOK, suppliers, PaginationMeta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *SupplierHandler) Update(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	var req UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	supplier, err := h.service.UpdateSupplier(c.Request.Context(), id, pharmacyID, &req)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, supplier)
}

func (h *SupplierHandler) GetStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), pharmacyID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, stats)
}

func (h *SupplierHandler) GetHistory(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid supplier ID")
		return
	}

	history, err := h.service.GetHistory(c.Request.Context(), id, pharmacyID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, history)
}

func (h *SupplierHandler) HealthCheck(c *gin.Context) {
	h.respondJSON(c, http.StatusOK, map[string]string{"status": "up"})
}

func (h *SupplierHandler) respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *SupplierHandler) respondJSONWithMeta(c *gin.Context, status int, data interface{}, meta PaginationMeta) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

func (h *SupplierHandler) respondError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error:   message,
	})
}
