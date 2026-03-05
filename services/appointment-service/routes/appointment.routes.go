package routes

import (
	"appointment-service/config"
	"appointment-service/controllers"
	"appointment-service/middleware"

	"github.com/gin-gonic/gin"
)

func AppointmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", controllers.HealthCheck)
	rg.Use(middleware.AuthMiddleware(config.DB))

	appointments := rg.Group("/appointments")
	{
		appointments.POST("/simple", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateSimpleAppointment)
		appointments.GET("/simple-list", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.GetSimpleAppointmentList)
		appointments.GET("/simple/:id", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.GetSimpleAppointmentDetails)
		appointments.POST("/simple/:id/reschedule", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.RescheduleAppointmentDetails)
		appointments.POST("/:id/reschedule-simple", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.RescheduleSimpleAppointment)
		appointments.GET("/followup-eligibility", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.CheckFollowUpEligibility)
		appointments.GET("/followup-eligibility/active", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.ListActiveFollowUps)
		appointments.POST("/followup-eligibility/expire-old", middleware.RequireRole(config.DB, "clinic_admin"), controllers.ExpireOldFollowUps)
		appointments.POST("", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateAppointment)
		appointments.POST("/with-patient", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreatePatientWithAppointment)
		appointments.GET("", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetAppointments)
		appointments.GET("/list", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetAppointmentList)
		appointments.GET("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetAppointment)
		appointments.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.UpdateAppointment)
		appointments.POST("/:id/reschedule", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.RescheduleAppointment)
		appointments.POST("/:id/cancel", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CancelAppointment)
		appointments.GET("/slots/available", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.GetAvailableTimeSlots)
		appointments.GET("/summary", middleware.RequireRole(config.DB, "clinic_admin", "receptionist", "doctor"), controllers.GetAppointmentSummary)
	}

	checkins := rg.Group("/checkins")
	{
		checkins.POST("", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.CreateCheckin)
		checkins.GET("", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetCheckins)
		checkins.GET("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetCheckin)
		checkins.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.UpdateCheckin)
		checkins.GET("/doctor/:doctor_id/queue", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetDoctorQueue)
	}

	vitals := rg.Group("/vitals")
	{
		vitals.POST("", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.CreateVitals)
		vitals.GET("", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetVitals)
		vitals.GET("/appointment/:appointment_id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetVitalsByAppointment)
		vitals.GET("/appointment/:appointment_id/history", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetVitalsHistoryByAppointment)
		vitals.PUT("/:id", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.UpdateVitals)
		vitals.GET("/patient/:patient_id/history", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientVitalsHistory)
		vitals.GET("/clinic-patient/:patient_id", middleware.RequireRole(config.DB, "clinic_admin", "doctor", "receptionist"), controllers.GetPatientVitalsHistory)
	}

	reports := rg.Group("/reports")
	{
		reports.GET("/daily-collection", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetDailyCollectionReport)
		reports.GET("/pending-payments", middleware.RequireRole(config.DB, "clinic_admin", "receptionist"), controllers.GetPendingPaymentsReport)
		reports.GET("/utilization", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetUtilizationReport)
		reports.GET("/no-show", middleware.RequireRole(config.DB, "clinic_admin", "doctor"), controllers.GetNoShowReport)
	}
}
