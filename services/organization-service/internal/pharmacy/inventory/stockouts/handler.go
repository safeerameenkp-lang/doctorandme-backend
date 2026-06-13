package stockouts

import (
	"fmt"
	"net/http"
	"organization-service/middleware"
	"strconv"

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
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	userIDStr, userName, _ := middleware.GetUserInfo(c.Request.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid user")
		return
	}

	var req CreateStockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondError(c, http.StatusBadRequest, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	stockOut, err := h.svc.CreateStockOut(c.Request.Context(), pharmacyID, userID, userName, req)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, stockOut)
}

func (h *Handler) GetByID(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid stock-out ID")
		return
	}

	stockOut, items, err := h.svc.GetStockOutDetails(c.Request.Context(), pharmacyID, id)
	if err != nil {
		h.respondError(c, http.StatusNotFound, "Stock-out record not found")
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]interface{}{
		"stock_out": stockOut,
		"items":     items,
	})
}

func (h *Handler) GetStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	stats, err := h.svc.GetStats(c.Request.Context(), pharmacyID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, stats)
}

func (h *Handler) List(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	stockOuts, total, err := h.svc.ListStockOuts(c.Request.Context(), pharmacyID, page, pageSize)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stockOuts,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  pageSize,
			Offset: (page - 1) * pageSize,
		},
	})
}

func (h *Handler) GetHistory(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid stock-out ID")
		return
	}

	history, err := h.svc.GetAuditLogs(c.Request.Context(), pharmacyID, id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, history)
}

func (h *Handler) respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{Success: true, Data: data})
}

func (h *Handler) respondError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{Success: false, Error: message})
}
