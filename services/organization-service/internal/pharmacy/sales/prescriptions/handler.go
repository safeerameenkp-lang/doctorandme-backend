package prescriptions

import (
	"net/http"
	"strconv"

	"organization-service/middleware"

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

func (h *Handler) Create(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	var req CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	p, err := h.svc.Create(c.Request.Context(), pharmacyID, req)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, p)
}

func (h *Handler) List(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	limit := 25
	offset := 0

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		}
	}

	data, total, stats, err := h.svc.List(c.Request.Context(), pharmacyID, limit, offset)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"stats": gin.H{
				"total_sales":     stats.TotalSales,
				"total_amount":    stats.TotalAmount,
				"pending_sales":   stats.PendingSales,
				"completed_sales": stats.CompletedSales,
			},
		},
	})
}

func (h *Handler) Get(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	id := c.Param("id")

	p, err := h.svc.Get(c.Request.Context(), pharmacyID, id)
	if err != nil {
		h.respondError(c, http.StatusNotFound, "Prescription not found")
		return
	}

	h.respondJSON(c, http.StatusOK, p)
}

func (h *Handler) respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func (h *Handler) respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": message})
}
