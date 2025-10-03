package routes

import (
    "organization-service/config"
    "organization-service/controllers"
    "github.com/gin-gonic/gin"
    "shared-security"
)

func OrganizationRoutes(rg *gin.RouterGroup) {
    // Health check endpoint (no auth required)
    rg.GET("/health", controllers.HealthCheck)
    
    // Apply authentication middleware to all other routes using shared module
    rg.Use(security.AuthMiddleware(config.DB))
    
    // Organizations
    org := rg.Group("/organizations")
    {
        org.POST("", security.RequireRole(config.DB, "super_admin"), controllers.CreateOrganization)
        org.POST("/with-admin", security.RequireRole(config.DB, "super_admin"), controllers.CreateOrganizationWithAdmin)
        org.GET("", controllers.GetOrganizations)
        org.GET("/:id", controllers.GetOrganization)
        org.PUT("/:id", security.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.UpdateOrganization)
        org.DELETE("/:id", security.RequireRole(config.DB, "super_admin"), controllers.DeleteOrganization)
    }
    
    // Clinics
    clinics := rg.Group("/clinics")
    {
        clinics.POST("", security.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.CreateClinic)
        clinics.POST("/with-admin", security.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.CreateClinicWithAdmin)
        clinics.GET("", controllers.GetClinics)
        clinics.GET("/:id", controllers.GetClinic)
        clinics.PUT("/:id", security.RequireRole(config.DB, "super_admin", "organization_admin", "clinic_admin"), controllers.UpdateClinic)
        clinics.DELETE("/:id", security.RequireRole(config.DB, "super_admin", "organization_admin"), controllers.DeleteClinic)
    }
    
    // Doctor Management
    doctors := rg.Group("/doctors")
    {
        // Create doctor profile only (no clinic assignment)
        // Use /clinic-doctor-links to assign doctor to multiple clinics
        doctors.POST("", security.RequireRole(config.DB, "super_admin", "clinic_admin"), controllers.CreateDoctor)
        doctors.GET("", controllers.GetDoctors)
        doctors.GET("all", controllers.GetAllDoctors)
        doctors.GET("/:id", controllers.GetDoctor)
        doctors.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateDoctor)
        doctors.DELETE("/:id", security.RequireRole(config.DB, "clinic_admin"), controllers.DeleteDoctor)
    }
    
    // Clinic Doctor Links (link any doctor to multiple clinics)
    links := rg.Group("/clinic-doctor-links")
    {
        links.POST("", security.RequireRole(config.DB, "super_admin","clinic_admin"), controllers.CreateClinicDoctorLink)
        links.GET("", controllers.GetClinicDoctorLinks)
      
        links.DELETE("/:id", security.RequireRole(config.DB, "clinic_admin"), controllers.DeleteClinicDoctorLink)
    }
    
    // Doctor Schedule Management
    schedules := rg.Group("/doctor-schedules")
    {
        schedules.POST("", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.CreateDoctorSchedule)
        schedules.GET("", controllers.GetDoctorSchedules)
        schedules.GET("/:id", controllers.GetDoctorSchedule)
        schedules.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateDoctorSchedule)
        schedules.DELETE("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.DeleteDoctorSchedule)
    }
    
    // External Services
    services := rg.Group("/services")
    {
        services.POST("", security.RequireRole(config.DB, "super_admin"), controllers.CreateExternalService)
        services.GET("", controllers.GetExternalServices)
        services.GET("/:id", controllers.GetExternalService)
        services.PUT("/:id", security.RequireRole(config.DB, "super_admin"), controllers.UpdateExternalService)
        services.DELETE("/:id", security.RequireRole(config.DB, "super_admin"), controllers.DeleteExternalService)
    }
    
    // Clinic Service Links
    serviceLinks := rg.Group("/links")
    {
        serviceLinks.POST("", security.RequireRole(config.DB, "super_admin"), controllers.CreateClinicServiceLink)
        serviceLinks.GET("", controllers.GetClinicServiceLinks)
        serviceLinks.GET("/:id", controllers.GetClinicServiceLink)
        serviceLinks.DELETE("/:id", security.RequireRole(config.DB, "super_admin"), controllers.DeleteClinicServiceLink)
    }
    
    // Patient Management
    patients := rg.Group("/patients")
    {
        patients.POST("", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreatePatient)
        patients.GET("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatients)
        patients.GET("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatient)
        patients.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.UpdatePatient)
        patients.DELETE("/:id", security.RequireRole(config.DB, "clinic_admin"), controllers.DeletePatient)
        patients.GET("/clinic/:clinic_id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientsByClinic)
    }
    
    // Patient-Clinic Assignments
    patientClinics := rg.Group("/patient-clinics")
    {
        patientClinics.POST("", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.AssignPatientToClinic)
        patientClinics.GET("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientClinicAssignments)
        patientClinics.GET("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientClinicAssignment)
        patientClinics.PUT("/:id", security.RequireRole(config.DB, "clinic_admin"), controllers.UpdatePatientClinicAssignment)
        patientClinics.DELETE("/:id", security.RequireRole(config.DB, "clinic_admin"), controllers.RemovePatientFromClinic)
        patientClinics.GET("/patient/:patient_id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetClinicsByPatient)
    }
    
    // ==================== ADMIN PANEL ROUTES ====================
    
    // Staff Management (Clinic Admin only) - Enhanced Admin Panel
    adminStaff := rg.Group("/admin/staff")
    adminStaff.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminStaff.POST("", controllers.CreateStaff)
        adminStaff.GET("/clinic/:clinic_id", controllers.GetClinicStaff)
        adminStaff.PUT("/clinic/:clinic_id/:user_id/role", controllers.UpdateStaffRole)
        adminStaff.DELETE("/clinic/:clinic_id/:user_id", controllers.DeactivateStaff)
    }
    
    // Queue Management
    adminQueues := rg.Group("/admin/queues")
    adminQueues.Use(security.RequireRole(config.DB, "clinic_admin"))
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
    adminPharmacy.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminPharmacy.POST("/medicines", controllers.CreateMedicine)
        adminPharmacy.GET("/inventory", controllers.GetPharmacyInventory)
        adminPharmacy.PUT("/medicines/:medicine_id/stock", controllers.UpdateMedicineStock)
        adminPharmacy.POST("/discounts", controllers.CreatePharmacyDiscount)
    }
    
    // Lab Management
    adminLab := rg.Group("/admin/lab")
    adminLab.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminLab.POST("/tests", controllers.CreateLabTest)
        adminLab.GET("/tests", controllers.GetLabTests)
        adminLab.POST("/collectors", controllers.CreateSampleCollector)
    }
    
    // Lab Results Upload (Lab Technicians can also upload)
    labResults := rg.Group("/admin/lab/results")
    labResults.Use(security.RequireRole(config.DB, "clinic_admin", "lab_technician"))
    {
        labResults.POST("", controllers.UploadLabResult)
    }
    
    // Insurance Provider Management
    adminInsurance := rg.Group("/admin/insurance")
    adminInsurance.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminInsurance.POST("/providers", controllers.CreateInsuranceProvider)
        adminInsurance.GET("/providers", controllers.GetInsuranceProviders)
    }
    
    // Reports & Analytics
    adminReports := rg.Group("/admin/reports")
    adminReports.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminReports.GET("/daily-stats", controllers.GetDailyStats)
        adminReports.GET("/doctor-stats", controllers.GetDoctorStats)
        adminReports.GET("/financial", controllers.GetFinancialReport)
    }
    
    // Patient Management (Admin)
    adminPatients := rg.Group("/admin/patients")
    adminPatients.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminPatients.POST("/merge", controllers.MergePatients)
        adminPatients.GET("/:patient_id/history", controllers.GetPatientHistory)
    }
    
    // Billing & Fee Management
    adminBilling := rg.Group("/admin/billing")
    adminBilling.Use(security.RequireRole(config.DB, "clinic_admin"))
    {
        adminBilling.POST("/fee-structures", controllers.CreateFeeStructure)
        adminBilling.GET("/fee-structures", controllers.GetFeeStructures)
        adminBilling.POST("/discounts", controllers.CreateBillingDiscount)
    }
}
