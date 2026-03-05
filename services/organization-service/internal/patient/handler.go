package patient

import (
	"fmt"
	"net/http"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

// PatientHandler handles all HTTP routing related to the Patient context
type PatientHandler struct {
	service PatientService
}

// NewPatientHandler creates a new handler struct with the injected service
func NewPatientHandler(svc PatientService) *PatientHandler {
	return &PatientHandler{service: svc}
}

// CreatePatient - Create a new patient (Super Admin - Global)
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var input CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Delegate processing to the pure service layer
	patient, err := h.service.CreatePatient(c.Request.Context(), input)
	if err != nil {
		if err.Error() == "phone_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Phone number exists",
				"message": "A user with this phone number already exists",
			})
			return
		}
		if err.Error() == "mo_id_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Mo ID exists",
				"message": "A patient with this Mo ID already exists",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to create patient: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Patient created successfully",
		"patient": patient,
	})
}

// CreatePatientWithClinic - Create a new patient and assign to clinic (Clinic Admin)
func (h *PatientHandler) CreatePatientWithClinic(c *gin.Context) {
	var input CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Let the service handle validations and transactional logic
	patient, clinicName, err := h.service.CreatePatientWithClinic(c.Request.Context(), input)
	if err != nil {
		if err.Error() == "missing_clinic_id" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing clinic_id",
				"message": "clinic_id is required for clinic admin patient creation",
			})
			return
		}
		if err.Error() == "clinic_not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Clinic not found",
				"message": "Clinic not found or is inactive",
			})
			return
		}
		if err.Error() == "phone_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Phone number exists",
				"message": "A user with this phone number already exists",
			})
			return
		}
		if err.Error() == "mo_id_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Mo ID exists",
				"message": "A patient with this Mo ID already exists",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to create patient: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Patient created and assigned to clinic successfully",
		"patient": patient,
		"clinic": gin.H{
			"id":   *input.ClinicID,
			"name": clinicName,
		},
	})
}

// ListPatients - List patients for a clinic
func (h *PatientHandler) ListPatients(c *gin.Context) {
	clinicID := c.Query("clinic_id")
	onlyActive := c.DefaultQuery("only_active", "true")
	search := c.Query("search") // Search by name, phone, or mo_id

	patients, err := h.service.ListPatients(c.Request.Context(), clinicID, search, onlyActive)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch patients")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patients":    patients,
		"total_count": len(patients),
	})
}

// GetPatient - Get single patient details
func (h *PatientHandler) GetPatient(c *gin.Context) {
	patientID := c.Param("id")

	patient, err := h.service.GetPatient(c.Request.Context(), patientID)
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Patient not found",
				"message": "The specified patient does not exist",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch patient")
		return
	}

	c.JSON(http.StatusOK, patient)
}

// UpdatePatient - Update patient details
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	patientID := c.Param("id")

	var input UpdatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	if err := h.service.UpdatePatient(c.Request.Context(), patientID, input); err != nil {
		if err.Error() == "mo_id_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Mo ID exists",
				"message": "A patient with this Mo ID already exists",
			})
			return
		}
		if err.Error() == "phone_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Phone number exists",
				"message": "A user with this phone number already exists",
			})
			return
		}
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Patient not found",
				"message": "The specified patient does not exist",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to update patient: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient updated successfully",
	})
}

// DeletePatient - Delete a patient (soft delete)
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	patientID := c.Param("id")

	if err := h.service.DeletePatient(c.Request.Context(), patientID); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Patient not found",
				"message": "The specified patient does not exist",
			})
			return
		}
		if fmt.Sprintf("%v", err)[0:7] == "patient" { // error formatting matching appointments check
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Cannot delete patient",
				"message": fmt.Sprintf("Please handle appointments before deleting. Reason: %v", err),
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to delete patient")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient deleted successfully",
	})
}

// AssignPatientToClinic - Assign existing patient to another clinic
func (h *PatientHandler) AssignPatientToClinic(c *gin.Context) {
	patientID := c.Param("id")

	var input AssignClinicInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	if err := h.service.AssignPatientToClinic(c.Request.Context(), patientID, input); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Patient not found",
				"message": "Patient not found or is inactive",
			})
			return
		}
		if err.Error() == "clinic_not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Clinic not found",
				"message": "Clinic not found or is inactive",
			})
			return
		}
		if err.Error() == "already_assigned" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Patient already assigned",
				"message": "Patient is already assigned to this clinic",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to assign patient: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient assigned to clinic successfully",
	})
}
