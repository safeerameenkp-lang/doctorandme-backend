package main

import (
	"context"
	"log"
	"net/http"
	"organization-service/config"
	"organization-service/internal/patient"
	"organization-service/routes"
	"os"
	"os/signal"
	"syscall"
	"time"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	r := gin.Default()

	// Add CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Serve uploaded files (images, documents, etc.) - accessible at /uploads/...
	// This route is outside /api to match the Kong routing configuration
	r.Static("/uploads", "./uploads")

	// Initialize Domain Dependencies
	patientRepo := patient.NewPatientRepository(config.DB)
	patientService := patient.NewPatientService(patientRepo)
	patientHandler := patient.NewPatientHandler(patientService)

	api := r.Group("/api")
	routes.OrganizationRoutes(api, patientHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Organization service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down organization service...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Organization service forced to shutdown:", err)
	}

	log.Println("Organization service exited")
}
