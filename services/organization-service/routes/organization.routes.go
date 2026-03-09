package routes

import (
	"organization-service/config"
	"organization-service/controllers"
	"organization-service/internal/patient"
	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

func OrganizationRoutes(rg *gin.RouterGroup, patientHandler *patient.PatientHandler) {
	// Health check endpoint (no auth required)
	rg.GET("/health", controllers.HealthCheck)
	// Serve static uploads (public access for logos)
	// accessible at /api/organizations/uploads/...
	rg.Static("/uploads", "./uploads")

	rg.Use(middleware.AuthMiddleware(config.DB))

	// Organizations
	org := rg.Group("/organizations")
	{
		org.POST("", middleware.RequireRole(config.DB, "super_admin"), controllers.CreateOrganization)
		org.POST("/with-admin", middleware.RequireRole(config.DB, "super_admin"), controllers.CreateOrganizationWithAdmin)
		org.GET("", controllers.GetOrganizations)
		org.GET("/:id", controllers.GetOrganization)
		org.PUT("/:id", middleware.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.UpdateOrganization)
		org.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin"), controllers.DeleteOrganization)
	}

	// Clinics
	clinics := rg.Group("/clinics")
	{
		clinics.POST("", middleware.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.CreateClinic)
		clinics.POST("/with-admin", middleware.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.CreateClinicWithAdmin)
		clinics.GET("", controllers.GetClinics)
		clinics.GET("/:id", controllers.GetClinic)
		clinics.PUT("/:id", middleware.RequireRole(config.DB, "super_admin", "organization_admin", "clinic_admin"), controllers.UpdateClinic)
		clinics.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.DeleteClinic)
		// Get doctors by clinic
		clinics.GET("/:id/doctors", controllers.GetDoctorsByClinic)
	}

	// Doctor Management
	doctors := rg.Group("/doctors")
	{
		// Create doctor profile only (no clinic assignment)
		// Use /clinic-doctor-links to assign doctor to multiple clinics
		doctors.POST("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.CreateDoctor)
		doctors.GET("", controllers.GetDoctors)
		doctors.GET("all", controllers.GetAllDoctors)
		doctors.GET("/:id", controllers.GetDoctor)
		doctors.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateDoctor)
		doctors.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin"), controllers.DeleteDoctor)

		// Get doctors by clinic (role-scoped) - returns clinic-specific fees
		doctors.GET("/clinic/:clinic_id", controllers.GetDoctorsByClinic)
	}

	// Doctor Leave Management
	leaves := rg.Group("/doctor-leaves")
	leaves.Use(middleware.AuthMiddleware(config.DB)) // All leave routes require authentication
	{
		// Apply for leave (Doctor, Clinic Admin, Receptionist)
		leaves.POST("", middleware.RequireRole(config.DB, "doctor", "clinic_admin", "receptionist"), controllers.ApplyLeave)

		// List leaves (with query filters: clinic_id, doctor_id, status, leave_type)
		leaves.GET("", controllers.ListDoctorLeaves)

		// Get single leave details
		leaves.GET("/:id", controllers.GetDoctorLeave)

		// Update leave (Doctor can update their own pending leave)
		leaves.PUT("/:id", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.UpdateDoctorLeave)

		// Review leave (Clinic Admin/Receptionist)
		leaves.POST("/:id/review", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.ReviewLeave)

		// Cancel leave (Doctor cancels their own, or Admin cancels)
		leaves.POST("/:id/cancel", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.CancelLeave)

		// Get leave statistics for a doctor
		leaves.GET("/stats/:doctor_id", controllers.GetDoctorLeaveStats)
	}

	// Doctor Time Slots Management (Date-Specific Slots)
	timeSlots := rg.Group("/doctor-time-slots")
	timeSlots.Use(middleware.AuthMiddleware(config.DB)) // All time slot routes require authentication
	{
		// Create date-specific time slots (bulk create) - Doctor, Clinic Admin
		timeSlots.POST("", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.CreateDoctorTimeSlots)

		// List time slots with query filtering - Query params: doctor_id (required), clinic_id, slot_type, date
		timeSlots.GET("", controllers.ListDoctorTimeSlots)

		// Get single time slot by ID
		timeSlots.GET("/:id", controllers.GetDoctorTimeSlot)

		// Update time slot - Doctor, Clinic Admin
		timeSlots.PUT("/:id", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.UpdateDoctorTimeSlot)

		// Delete time slot (soft delete) - Doctor, Clinic Admin
		timeSlots.DELETE("/:id", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.DeleteDoctorTimeSlot)
	}

	// Session-Based Doctor Time Slots (Auto-generates individual bookable slots)
	sessionSlots := rg.Group("/doctor-session-slots")
	sessionSlots.Use(middleware.AuthMiddleware(config.DB))
	{
		// Create session-based time slots with auto-generated individual slots
		sessionSlots.POST("", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.CreateDoctorSessionSlots)

		// List session-based slots - Query params: doctor_id (required), clinic_id, date, slot_type (clinic_visit/video_consultation/follow-up-via-clinic/follow-up-via-video)
		sessionSlots.GET("", controllers.ListDoctorSessionSlots)

		// Sync slot booking status with appointments table
		sessionSlots.POST("/sync-booking-status", middleware.RequireRole(config.DB, "clinic_admin"), controllers.SyncSlotBookingStatus)

		// Update existing session times within a time slot
		sessionSlots.PUT("/:id", middleware.RequireRole(config.DB, "doctor", "clinic_admin"), controllers.UpdateSessionSlotSessions)
	}

	// Clinic Doctor Links (link any doctor to multiple clinics with clinic-specific fees)
	links := rg.Group("/clinic-doctor-links")
	{
		links.POST("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.CreateClinicDoctorLink)
		links.GET("", controllers.GetClinicDoctorLinks)
		links.GET("/doctor/:doctor_id", controllers.GetClinicDoctorLinksByDoctor)
		links.PUT("/:id", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.UpdateClinicDoctorLink)
		links.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin"), controllers.DeleteClinicDoctorLink)
	}

	// Doctor Schedule Management
	schedules := rg.Group("/doctor-schedules")
	{
		schedules.POST("", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.CreateDoctorSchedule)
		schedules.GET("", controllers.GetDoctorSchedules)
		schedules.GET("/:id", controllers.GetDoctorSchedule)
		schedules.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateDoctorSchedule)
		schedules.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.DeleteDoctorSchedule)
	}

	// Doctor Consultation Fees
	consultationFees := rg.Group("/doctor-consultation-fees")
	consultationFees.Use(middleware.AuthMiddleware(config.DB))
	{
		consultationFees.GET("", controllers.GetDoctorConsultationFees)
		consultationFees.POST("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin", "doctor"), controllers.AddConsultationFees)
		consultationFees.PUT("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin", "doctor"), controllers.UpdateConsultationFees)
	}

	// External Services
	services := rg.Group("/services")
	{
		services.POST("", middleware.RequireRole(config.DB, "super_admin"), controllers.CreateExternalService)
		services.GET("", controllers.GetExternalServices)
		services.GET("/:id", controllers.GetExternalService)
		services.PUT("/:id", middleware.RequireRole(config.DB, "super_admin"), controllers.UpdateExternalService)
		services.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin"), controllers.DeleteExternalService)
	}

	// Clinic Service Links
	serviceLinks := rg.Group("/links")
	{
		serviceLinks.POST("", middleware.RequireRole(config.DB, "super_admin"), controllers.CreateClinicServiceLink)
		serviceLinks.GET("", controllers.GetClinicServiceLinks)
		serviceLinks.GET("/:id", controllers.GetClinicServiceLink)
		serviceLinks.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin"), controllers.DeleteClinicServiceLink)
	}

	// Patient-Clinic Assignments
	patientClinics := rg.Group("/patient-clinics")
	{
		patientClinics.POST("", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), patientHandler.AssignPatientToClinic)
		patientClinics.GET("", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientClinicAssignments)
		patientClinics.GET("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientClinicAssignment)
		patientClinics.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin"), controllers.UpdatePatientClinicAssignment)
		patientClinics.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin"), controllers.RemovePatientFromClinic)
		patientClinics.GET("/patient/:patient_id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetClinicsByPatient)
	}

	// ==================== ADMIN PANEL ROUTES ====================

	// Staff Management (Clinic Admin only) - Enhanced Admin Panel
	adminStaff := rg.Group("/admin/staff")
	adminStaff.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminStaff.GET("/roles", controllers.ListRoles)
		adminStaff.POST("", controllers.CreateStaff)
		adminStaff.GET("/clinic/:clinic_id", controllers.GetClinicStaff)
		adminStaff.GET("/clinic/:clinic_id/:staff_id", controllers.GetStaffDetails)
		adminStaff.PUT("/clinic/:clinic_id/:user_id/role", controllers.UpdateStaffRole)
		adminStaff.PUT("/clinic/:clinic_id/:user_id", controllers.UpdateStaff)
		adminStaff.DELETE("/clinic/:clinic_id/:user_id", controllers.DeactivateStaff)
	}

	// Queue Management
	adminQueues := rg.Group("/admin/queues")
	adminQueues.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminQueues.POST("", controllers.CreateQueue)
		adminQueues.GET("", controllers.GetQueues)
		adminQueues.POST("/tokens", controllers.AssignToken)
		adminQueues.PUT("/tokens/:token_id/reassign", controllers.ReassignToken)
		adminQueues.PUT("/:queue_id/pause", controllers.PauseQueue)
		adminQueues.PUT("/:queue_id/resume", controllers.ResumeQueue)
	}

	// Pharmacy Management
	adminPharmacy := rg.Group("/admin/pharmacy")
	adminPharmacy.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminPharmacy.POST("/medicines", controllers.CreateMedicine)
		adminPharmacy.GET("/inventory", controllers.GetPharmacyInventory)
		adminPharmacy.PUT("/medicines/:medicine_id/stock", controllers.UpdateMedicineStock)
		adminPharmacy.POST("/discounts", controllers.CreatePharmacyDiscount)
	}

	// Lab Management
	adminLab := rg.Group("/admin/lab")
	adminLab.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminLab.POST("/tests", controllers.CreateLabTest)
		adminLab.GET("/tests", controllers.GetLabTests)
		adminLab.POST("/collectors", controllers.CreateSampleCollector)
	}

	// Lab Results Upload (Lab Technicians can also upload)
	labResults := rg.Group("/admin/lab/results")
	labResults.Use(middleware.RequireRole(config.DB, "clinic_admin", "lab_technician"))
	{
		labResults.POST("", controllers.UploadLabResult)
	}

	// Insurance Provider Management
	adminInsurance := rg.Group("/admin/insurance")
	adminInsurance.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminInsurance.POST("/providers", controllers.CreateInsuranceProvider)
		adminInsurance.GET("/providers", controllers.GetInsuranceProviders)
	}

	// Reports & Analytics
	adminReports := rg.Group("/admin/reports")
	adminReports.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminReports.GET("/daily-stats", controllers.GetDailyStats)
		adminReports.GET("/doctor-stats", controllers.GetDoctorStats)
		adminReports.GET("/financial", controllers.GetFinancialReport)
	}

	// Patient Management (Admin)
	adminPatients := rg.Group("/admin/patients")
	adminPatients.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminPatients.POST("/merge", controllers.MergePatients)
		adminPatients.GET("/:patient_id/history", controllers.GetPatientHistory)
	}

	// Billing & Fee Management
	adminBilling := rg.Group("/admin/billing")
	adminBilling.Use(middleware.RequireRole(config.DB, "clinic_admin"))
	{
		adminBilling.POST("/fee-structures", controllers.CreateFeeStructure)
		adminBilling.GET("/fee-structures", controllers.GetFeeStructures)
		adminBilling.POST("/discounts", controllers.CreateBillingDiscount)
	}

	// Department Management
	departments := rg.Group("/departments")
	departments.Use(middleware.AuthMiddleware(config.DB))
	{
		departments.POST("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.CreateDepartment)
		departments.GET("", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.ListDepartments)
		departments.GET("/:id", controllers.GetDepartment)
		departments.PUT("/:id", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.UpdateDepartment)
		departments.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.DeleteDepartment)
		departments.GET("/:id/doctors", controllers.GetDoctorsByDepartment)
	}

	// Patient Management - Super Admin (Global)
	patientsGlobal := rg.Group("/patients")
	patientsGlobal.Use(middleware.AuthMiddleware(config.DB))
	{
		patientsGlobal.POST("", middleware.RequireRole(config.DB, "super_admin"), patientHandler.CreatePatient)
		patientsGlobal.GET("", patientHandler.ListPatients)
		patientsGlobal.GET("/:id", patientHandler.GetPatient)
		patientsGlobal.PUT("/:id", middleware.RequireRole(config.DB, "super_admin"), patientHandler.UpdatePatient)
		patientsGlobal.DELETE("/:id", middleware.RequireRole(config.DB, "super_admin"), patientHandler.DeletePatient)
	}

	// Patient Management - Clinic Admin (Clinic-specific)
	patientsClinic := rg.Group("/clinic-patients")
	patientsClinic.Use(middleware.AuthMiddleware(config.DB))
	{
		patientsClinic.POST("", middleware.RequireRole(config.DB, "clinic_admin"), patientHandler.CreatePatientWithClinic)
		patientsClinic.GET("", patientHandler.ListPatients)
		patientsClinic.GET("/:id", patientHandler.GetPatient)
		patientsClinic.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin"), patientHandler.UpdatePatient)
		patientsClinic.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin"), patientHandler.DeletePatient)
		patientsClinic.POST("/:id/assign-clinic", middleware.RequireRole(config.DB, "clinic_admin"), patientHandler.AssignPatientToClinic)
	}

	// Clinic-Specific Patients (Isolated per clinic, no global users)
	clinicSpecificPatients := rg.Group("/clinic-specific-patients")
	clinicSpecificPatients.Use(middleware.AuthMiddleware(config.DB))
	{
		// Create patient for specific clinic (no global user creation)
		clinicSpecificPatients.POST("", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateClinicPatient)

		// List patients for clinic - Query param: clinic_id (required), search, only_active
		clinicSpecificPatients.GET("", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.ListClinicPatients)

		// Get single clinic patient
		clinicSpecificPatients.GET("/:id", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.GetClinicPatient)

		// Get single clinic patient full details with doctor mappings and complete history
		clinicSpecificPatients.GET("/:id/details", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.GetClinicPatientFullDetails)

		// Update clinic patient
		clinicSpecificPatients.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.UpdateClinicPatient)

		// Delete clinic patient (soft delete)
		clinicSpecificPatients.DELETE("/:id", middleware.RequireRole(config.DB, "clinic_admin"), controllers.DeleteClinicPatient)
	}

	// Follow-up status is now integrated into GetClinicPatient endpoint
}
