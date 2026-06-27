package main

import (
	"context"
	"log"
	"net/http"
	"organization-service/config"
	"organization-service/internal/patient"
	"organization-service/internal/pharmacy/inventory/batches"
	"organization-service/internal/pharmacy/inventory/ledger"
	"organization-service/internal/pharmacy/inventory/medicines"
	"organization-service/internal/pharmacy/inventory/reservations"
	"organization-service/internal/pharmacy/inventory/stockin"
	"organization-service/internal/pharmacy/inventory/stockouts"
	"organization-service/internal/pharmacy/notification"
	"organization-service/internal/pharmacy/sales/clients"
	"organization-service/internal/pharmacy/sales/prescriptions"
	"organization-service/internal/pharmacy/sales/sales"
	"organization-service/internal/pharmacy/supplier"
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

	// Initialize Pharmacy Inventory dependencies
	medsRepo := medicines.NewRepository(config.DB)
	medsSvc := medicines.NewService(medsRepo)
	medsHandler := medicines.NewHandler(medsSvc)

	batchesRepo := batches.NewRepository(config.DB)
	batchesSvc := batches.NewService(batchesRepo)
	batchesHandler := batches.NewHandler(batchesSvc)

	ledgerRepo := ledger.NewRepository(config.DB)
	ledgerSvc := ledger.NewService(ledgerRepo)
	ledgerHandler := ledger.NewHandler(ledgerSvc, ledgerRepo)

	resRepo := reservations.NewRepository(config.DB)
	resSvc := reservations.NewService(resRepo, batchesRepo)
	resHandler := reservations.NewHandler(resSvc)
	resSvc.StartPurgeWorker(context.Background()) // Start the background reservations purge worker

	stockInRepo := stockin.NewRepository(config.DB)
	stockInSvc := stockin.NewService(stockInRepo, medsRepo, batchesSvc)
	stockInHandler := stockin.NewHandler(stockInSvc)

	stockOutRepo := stockouts.NewRepository(config.DB)
	stockOutSvc := stockouts.NewService(stockOutRepo)
	stockOutHandler := stockouts.NewHandler(stockOutSvc)

	inventoryHandlers := routes.InventoryHandlers{
		Meds:     medsHandler,
		Batches:  batchesHandler,
		Ledger:   ledgerHandler,
		StockIn:  stockInHandler,
		StockOut: stockOutHandler,
		Res:      resHandler,
	}

	// Initialize Pharmacy Sales dependencies
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	inventoryBaseURL := "http://localhost:" + port + "/api/pharmacy/inventory"

	rxRepo := prescriptions.NewRepository(config.DB)
	rxSvc := prescriptions.NewService(rxRepo)
	rxHandler := prescriptions.NewHandler(rxSvc)

	salesRepo := sales.NewRepository(config.DB)
	invClient := clients.NewInventoryClient(inventoryBaseURL)
	rxClient := clients.NewLocalPrescriptionClient(rxRepo)
	salesSvc := sales.NewService(salesRepo, invClient, rxClient)
	salesHandler := sales.NewHandler(salesSvc)

	salesHandlers := routes.SalesHandlers{
		Sales: salesHandler,
		Rx:    rxHandler,
	}

	// Initialize Pharmacy Supplier dependencies
	supRepo := supplier.NewPostgresSupplierRepository(config.DB)
	supService := supplier.NewSupplierService(supRepo)
	supHandler := supplier.NewSupplierHandler(supService)

	supplierHandlersBundle := routes.SupplierHandlers{
		Supplier: supHandler,
	}

	// Initialize Pharmacy Notification dependencies
	notifService := notification.NewNotificationService()
	notifHandler := notification.NewNotificationHandler(notifService)

	notificationHandlersBundle := routes.NotificationHandlers{
		Notification: notifHandler,
	}

	api := r.Group("/api")
	routes.OrganizationRoutes(api, patientHandler, inventoryHandlers, salesHandlers, supplierHandlersBundle, notificationHandlersBundle)

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
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
