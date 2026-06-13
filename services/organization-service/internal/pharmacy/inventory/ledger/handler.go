package ledger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"organization-service/middleware"
)

type Handler struct {
	svc  Service
	repo Repository // Using repo for queries to keep it simple
}

func NewHandler(svc Service, repo Repository) *Handler {
	return &Handler{svc: svc, repo: repo}
}

func (h *Handler) GetByBatch(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	batchIDStr := c.Param("batchId")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	logs, err := h.repo.GetByBatch(c.Request.Context(), pharmacyID, batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

func (h *Handler) GetByMedicine(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	medicineIDStr := c.Param("medicineId")
	medicineID, err := uuid.Parse(medicineIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid medicine ID"})
		return
	}

	logs, err := h.repo.GetByMedicine(c.Request.Context(), pharmacyID, medicineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}
