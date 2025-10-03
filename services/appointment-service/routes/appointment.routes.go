package routes

import (
    "appointment-service/config"
    "appointment-service/controllers"
    "github.com/gin-gonic/gin"
    "shared-security"
)

func AppointmentRoutes(rg *gin.RouterGroup) {
    // Health check endpoint (no auth required)
    rg.GET("/health", controllers.HealthCheck)
    
    // Apply authentication middleware to all other routes using shared module
    rg.Use(security.AuthMiddleware(config.DB))
    
    // Appointments
    appointments := rg.Group("/appointments")
    {
        appointments.POST("", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateAppointment)
        appointments.POST("/with-patient", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreatePatientWithAppointment)
        appointments.GET("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetAppointments)
        appointments.GET("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetAppointment)
        appointments.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.UpdateAppointment)
        appointments.POST("/:id/reschedule", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.RescheduleAppointment)
        appointments.POST("/:id/cancel", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CancelAppointment)
        appointments.GET("/slots/available", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.GetAvailableTimeSlots)
    }
    
    // Patient Check-ins
    checkins := rg.Group("/checkins")
    {
        checkins.POST("", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateCheckin)
        checkins.GET("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetCheckins)
        checkins.GET("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetCheckin)
        checkins.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.UpdateCheckin)
        checkins.GET("/doctor/:doctor_id/queue", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetDoctorQueue)
    }
    
    // Patient Vitals
    vitals := rg.Group("/vitals")
    {
        vitals.POST("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.CreateVitals)
        vitals.GET("", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetVitals)
        vitals.GET("/appointment/:appointment_id", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetVitalsByAppointment)
        vitals.PUT("/:id", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateVitals)
        vitals.GET("/patient/:patient_id/history", security.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientVitalsHistory)
    }
    
    // Reports (Admin and Doctor access)
    reports := rg.Group("/reports")
    {
        reports.GET("/daily-collection", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetDailyCollectionReport)
        reports.GET("/pending-payments", security.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.GetPendingPaymentsReport)
        reports.GET("/utilization", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetUtilizationReport)
        reports.GET("/no-show", security.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetNoShowReport)
    }
}
