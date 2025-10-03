package routes

import (
    "auth-service/config"
    "auth-service/controllers"
    "github.com/gin-gonic/gin"
    "shared-security"
)

func AuthRoutes(rg *gin.RouterGroup) {
    // Health check endpoint
    rg.GET("/health", controllers.HealthCheck)
    
    // Public endpoints
    rg.POST("/register", controllers.Register)
    rg.POST("/login", controllers.Login)
    rg.POST("/refresh", controllers.Refresh)
    rg.POST("/logout", controllers.Logout)
    
    // Protected endpoints (all authenticated users)
    protected := rg.Group("")
    protected.Use(security.AuthMiddleware(config.DB))
    {
        protected.GET("/profile", controllers.GetProfile)
        protected.PUT("/profile", controllers.UpdateProfile)
        protected.POST("/change-password", controllers.ChangePassword)
    }
}
