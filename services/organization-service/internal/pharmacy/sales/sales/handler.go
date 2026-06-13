package sales

import (
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) GetSale(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))

	sale, err := h.svc.GetSaleWithDetails(c.Request.Context(), pharmacyID, saleID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, sale)
}

func (h *Handler) CreateDraft(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	rxId := c.Param("rxId")

	if rxId == "" {
		h.respondError(c, http.StatusBadRequest, "Missing prescription ID")
		return
	}

	sale, err := h.svc.CreateDraft(c.Request.Context(), pharmacyID, rxId)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, sale)
}

func (h *Handler) CreateWalkInDraft(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	var patient Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if patient.Name == "" || patient.Phone == "" {
		h.respondError(c, http.StatusBadRequest, "Name and Phone are required")
		return
	}

	sale, err := h.svc.CreateWalkInDraft(c.Request.Context(), pharmacyID, patient)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, sale)
}

func (h *Handler) AddItem(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	items, err := h.svc.AddItemToDraft(c.Request.Context(), pharmacyID, saleID, req)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, items)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))
	itemID, _ := uuid.Parse(c.Param("itemId"))

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err2 := h.svc.UpdateItem(c.Request.Context(), pharmacyID, saleID, itemID, req)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]string{"message": "Item updated"})
}

func (h *Handler) RemoveItem(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))
	itemID, _ := uuid.Parse(c.Param("itemId"))

	err2 := h.svc.RemoveItem(c.Request.Context(), pharmacyID, saleID, itemID)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]string{"message": "Item removed"})
}

func (h *Handler) FinalizeSale(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))

	var req FinalizeSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	sale, err2 := h.svc.FinalizeSale(c.Request.Context(), pharmacyID, saleID, req)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, sale)
}

func (h *Handler) DispatchSale(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	saleID, _ := uuid.Parse(c.Param("id"))

	sale, err2 := h.svc.DispatchSale(c.Request.Context(), pharmacyID, saleID)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, sale)
}

func (h *Handler) SearchPatients(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	phone := c.Query("phone")

	patients, err2 := h.svc.SearchPatientsByPhone(c.Request.Context(), pharmacyID, phone)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}
	h.respondJSON(c, http.StatusOK, patients)
}

func (h *Handler) ListPatients(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	limit := 25
	offset := 0
	search := c.Query("search")

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

	patients, total, err2 := h.svc.ListPatients(c.Request.Context(), pharmacyID, limit, offset, search)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    patients,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *Handler) GetPatientByID(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	patientID, err2 := uuid.Parse(c.Param("id"))
	if err2 != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	patient, err2 := h.svc.GetPatientByID(c.Request.Context(), pharmacyID, patientID)
	if err2 != nil {
		h.respondError(c, http.StatusNotFound, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, patient)
}

func (h *Handler) GetPatientSales(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	patientID, err2 := uuid.Parse(c.Param("id"))
	if err2 != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	limit := 6
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

	purchases, total, err2 := h.svc.GetPatientSales(c.Request.Context(), pharmacyID, patientID, limit, offset)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    purchases,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *Handler) GetPatientReturns(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	patientID, err2 := uuid.Parse(c.Param("id"))
	if err2 != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	limit := 6
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

	returns, total, err2 := h.svc.GetPatientReturns(c.Request.Context(), pharmacyID, patientID, limit, offset)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    returns,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *Handler) GetStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	granularity := c.Query("granularity")
	if granularity == "" {
		granularity = "day"
	}

	dateStr := c.Query("date")
	var targetDate time.Time
	if dateStr != "" {
		targetDate, _ = time.Parse("2006-01-02", dateStr)
	} else {
		targetDate = time.Now()
	}

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	} else {
		switch granularity {
		case "week":
			now := time.Now()
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
			endDate = startDate.AddDate(0, 1, -1)
		case "month":
			startDate = time.Now().AddDate(-1, 0, 0)
		default:
			startDate = time.Now().AddDate(0, 0, -6)
		}
	}

	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	} else if endDate.IsZero() {
		endDate = time.Now()
	}

	stats, err2 := h.svc.GetStats(c.Request.Context(), pharmacyID, targetDate, startDate, endDate, granularity)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, stats)
}

func (h *Handler) ProcessReturn(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	var req CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	handledBy := "Pharmacist"

	ret, err2 := h.svc.ProcessReturn(c.Request.Context(), pharmacyID, handledBy, req)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, ret)
}

func (h *Handler) ListReturns(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	returns, err2 := h.svc.ListReturns(c.Request.Context(), pharmacyID)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, returns)
}

func (h *Handler) GetReturn(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}
	returnID, _ := uuid.Parse(c.Param("id"))

	ret, err2 := h.svc.GetReturnDetails(c.Request.Context(), pharmacyID, returnID)
	if err2 != nil {
		h.respondError(c, http.StatusNotFound, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, ret)
}

func (h *Handler) GetPatientStats(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	stats, err2 := h.svc.GetPatientStats(c.Request.Context(), pharmacyID)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}
	h.respondJSON(c, http.StatusOK, stats)
}

func (h *Handler) ListSales(c *gin.Context) {
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

	search := c.Query("search")
	paymentMode := c.Query("paymentMode")

	var startDate, endDate time.Time
	if sd := c.Query("startDate"); sd != "" {
		startDate, _ = time.Parse("2006-01-02", sd)
	}
	if ed := c.Query("endDate"); ed != "" {
		endDate, _ = time.Parse("2006-01-02", ed)
		endDate = endDate.Add(24*time.Hour - time.Second)
	}

	salesList, total, err2 := h.svc.ListSales(c.Request.Context(), pharmacyID, limit, offset, startDate, endDate, paymentMode, search)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    salesList,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *Handler) GetRecurringRefillsReport(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy context")
		return
	}

	report, err2 := h.svc.GetRecurringRefillsReport(c.Request.Context(), pharmacyID)
	if err2 != nil {
		h.respondError(c, http.StatusInternalServerError, err2.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, report)
}

func (h *Handler) respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
	})
}
