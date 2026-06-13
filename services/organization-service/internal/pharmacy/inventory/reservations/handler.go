package reservations

import (
	"net/http"
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
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}

	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	res, err := h.svc.Reserve(c.Request.Context(), pharmacyID, req)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(c, http.StatusCreated, res)
}

func (h *Handler) Update(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}
	id, _ := uuid.Parse(c.Param("id"))

	var req UpdateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	if err := h.svc.Update(c.Request.Context(), pharmacyID, id, req); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *Handler) Confirm(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Confirm(c.Request.Context(), pharmacyID, id); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]string{"message": "confirmed"})
}

func (h *Handler) Cancel(c *gin.Context) {
	pharmacyIDStr := middleware.GetPharmacyInfo(c.Request.Context())
	pharmacyID, err := uuid.Parse(pharmacyIDStr)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid pharmacy")
		return
	}
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Cancel(c.Request.Context(), pharmacyID, id); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(c, http.StatusOK, map[string]string{"message": "cancelled"})
}

func (h *Handler) respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func (h *Handler) respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": message})
}
