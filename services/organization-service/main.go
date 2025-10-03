package main

import (
    "organization-service/config"
    "organization-service/routes"
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "shared-security"
)

func main() {
    config.ConnectDB()
    
    r := gin.Default()
    
    // Add CORS middleware from shared module
    r.Use(security.CORSMiddleware())
    
    api := r.Group("/api/organizations")
    routes.OrganizationRoutes(api)

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
