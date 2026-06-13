package batches

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"organization-service/middleware"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	stats, err := h.svc.GetStats(c.Request.Context(), pharmacyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (h *Handler) List(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	var medicineID *uuid.UUID
	if medIDStr := c.Query("medicine_id"); medIDStr != "" {
		id, err := uuid.Parse(medIDStr)
		if err == nil {
			medicineID = &id
		}
	}

	limit := 25
	offset := 0
	if lStr := c.Query("limit"); lStr != "" {
		if val, err := strconv.Atoi(lStr); err == nil && val > 0 {
			limit = val
		}
	}
	if oStr := c.Query("offset"); oStr != "" {
		if val, err := strconv.Atoi(oStr); err == nil && val >= 0 {
			offset = val
		}
	}

	search := c.Query("search")
	supplierID := c.Query("supplier_id")
	filter := c.Query("filter")

	batches, total, err := h.svc.ListBatches(c.Request.Context(), pharmacyID, medicineID, limit, offset, search, supplierID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batches,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *Handler) ListSellable(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	search := c.Query("search")
	limit := 10
	if lStr := c.Query("limit"); lStr != "" {
		if val, err := strconv.Atoi(lStr); err == nil && val > 0 {
			limit = val
		}
	}

	batches, err := h.svc.ListSellableBatches(c.Request.Context(), pharmacyID, search, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batches,
	})
}

func (h *Handler) GetHistory(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	logs, err := h.svc.GetBatchAuditLogs(c.Request.Context(), pharmacyID, batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		"success": true,
		"data":    logs,
	}
	if len(logs) == 0 {
		resp["message"] = "History not found"
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	var req EditBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userIDStr, userName, _ := middleware.GetUserInfo(c.Request.Context())
	userID, _ := uuid.Parse(userIDStr)

	if err := h.svc.UpdateBatch(c.Request.Context(), pharmacyID, batchID, userID, userName, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update batch: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Batch updated successfully",
	})
}

func (h *Handler) ProcessReturn(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid pharmacy ID"})
		return
	}

	var req BatchReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userIDStr, userName, _ := middleware.GetUserInfo(c.Request.Context())
	userID, _ := uuid.Parse(userIDStr)

	if err := h.svc.ProcessReturn(c.Request.Context(), pharmacyID, userID, userName, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process return: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stock returned successfully",
	})
}
