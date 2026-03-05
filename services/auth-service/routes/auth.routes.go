package routes

import (
    "auth-service/config"
    "auth-service/controllers"
    "auth-service/middleware"
    "github.com/gin-gonic/gin"
)

func AuthRoutes(rg *gin.RouterGroup) {
    // Health check endpoint
    rg.GET("/health", controllers.HealthCheck)
    
    // Public endpoints
    rg.POST("/register", controllers.Register)
    rg.POST("/login", controllers.Login)
    rg.POST("/refresh", controllers.Refresh)
    rg.POST("/logout", controllers.Logout)
    
    // Utility endpoint for password hashing (remove in production)
    rg.POST("/hash-password", controllers.HashPasswordUtility)
    
    // Protected endpoints (all authenticated users)
    protected := rg.Group("")
    protected.Use(middleware.AuthMiddleware(config.DB))
    {
        protected.GET("/profile", controllers.GetProfile)
        protected.PUT("/profile", controllers.UpdateProfile)
        protected.POST("/change-password", controllers.ChangePassword)
    }
    
    // Super Admin only endpoints (Platform-wide access)
    superAdmin := rg.Group("/admin")
    superAdmin.Use(middleware.AuthMiddleware(config.DB))
    superAdmin.Use(middleware.RequireSuperAdmin(config.DB))
    {
        // User Management (Platform-wide)
        superAdmin.GET("/users", controllers.ListUsers)
        superAdmin.GET("/users/:id", controllers.GetUser)
        superAdmin.POST("/users", controllers.CreateUser)
        superAdmin.PUT("/users/:id", controllers.UpdateUser)
        superAdmin.DELETE("/users/:id", controllers.DeleteUser)
        
        // User Status Management
        superAdmin.POST("/users/:id/block", controllers.BlockUser)
        superAdmin.POST("/users/:id/unblock", controllers.UnblockUser)
        superAdmin.POST("/users/:id/activate", controllers.ActivateUser)
        superAdmin.POST("/users/:id/deactivate", controllers.DeactivateUser)
        
        // Password Management
        superAdmin.POST("/users/:id/change-password", controllers.AdminChangePassword)
        
        // User Role Assignment (Platform-wide)
        superAdmin.POST("/users/:id/roles", controllers.AssignRole)
        superAdmin.DELETE("/users/:id/roles/:role_id", controllers.RemoveRole)
        
        // Activity Logs
        superAdmin.GET("/users/:id/activity-logs", controllers.GetUserActivityLogs)
        
        // Role Management (Platform-wide)
        superAdmin.GET("/roles", controllers.ListRoles)
        superAdmin.GET("/roles/:id", controllers.GetRole)
        superAdmin.POST("/roles", controllers.CreateRole)
        superAdmin.PUT("/roles/:id", controllers.UpdateRole)
        superAdmin.DELETE("/roles/:id", controllers.DeleteRole)
        
        // Role Status Management
        superAdmin.POST("/roles/:id/activate", controllers.ActivateRole)
        superAdmin.POST("/roles/:id/deactivate", controllers.DeactivateRole)
        
        // Role Permissions Management
        superAdmin.PUT("/roles/:id/permissions", controllers.UpdateRolePermissions)
        
        // Role Users
        superAdmin.GET("/roles/:id/users", controllers.GetRoleUsers)
        
        // Permission Templates
        superAdmin.GET("/permission-templates", controllers.GetPermissionTemplates)
    }
    
    // Organization Admin endpoints (Scoped to their organization)
    orgAdmin := rg.Group("/org-admin")
    orgAdmin.Use(middleware.AuthMiddleware(config.DB))
    orgAdmin.Use(middleware.RequireOrganizationAdmin(config.DB))
    {
        // User Management (Organization scope)
        orgAdmin.GET("/users", controllers.ScopedListUsers)
        orgAdmin.GET("/users/:id", controllers.GetUser)
        orgAdmin.POST("/users", controllers.CreateUser)
        orgAdmin.PUT("/users/:id", controllers.UpdateUser)
        
        // User Status Management (Limited)
        orgAdmin.POST("/users/:id/activate", controllers.ActivateUser)
        orgAdmin.POST("/users/:id/deactivate", controllers.DeactivateUser)
        
        // Role Assignment (Within organization)
        orgAdmin.POST("/users/:id/roles", controllers.AssignRole)
        orgAdmin.DELETE("/users/:id/roles/:role_id", controllers.RemoveRole)
        
        // View roles (cannot create/modify)
        orgAdmin.GET("/roles", controllers.ListRoles)
        orgAdmin.GET("/roles/:id", controllers.GetRole)
    }
    
    // Clinic Admin endpoints (Scoped to their clinic)
    clinicAdmin := rg.Group("/clinic-admin")
    clinicAdmin.Use(middleware.AuthMiddleware(config.DB))
    clinicAdmin.Use(middleware.RequireClinicAdmin(config.DB))
    {
        // User Management (Clinic scope)
        clinicAdmin.GET("/users", controllers.ScopedListUsers)
        clinicAdmin.GET("/users/:id", controllers.GetUser)
        clinicAdmin.POST("/users", controllers.CreateUser)
        clinicAdmin.PUT("/users/:id", controllers.UpdateUser)
        
        // User Status Management (Limited)
        clinicAdmin.POST("/users/:id/activate", controllers.ActivateUser)
        clinicAdmin.POST("/users/:id/deactivate", controllers.DeactivateUser)
        
        // Role Assignment (Within clinic)
        clinicAdmin.POST("/users/:id/roles", controllers.AssignRole)
        clinicAdmin.DELETE("/users/:id/roles/:role_id", controllers.RemoveRole)
        
        // View roles (cannot create/modify)
        clinicAdmin.GET("/roles", controllers.ListRoles)
        clinicAdmin.GET("/roles/:id", controllers.GetRole)
    }
    
    // === SCOPED RESOURCE ENDPOINTS ===
    // These endpoints work across all admin levels with automatic role-based filtering
    
    // Super Admin - Resource Management (Platform-wide)
    superAdminResources := rg.Group("/admin/resources")
    superAdminResources.Use(middleware.AuthMiddleware(config.DB))
    superAdminResources.Use(middleware.RequireSuperAdmin(config.DB))
    {
        superAdminResources.GET("/clinics", controllers.ListClinics)
        superAdminResources.GET("/patients", controllers.ListPatients)
        superAdminResources.GET("/doctors", controllers.ListDoctors)
        superAdminResources.GET("/staff", controllers.ListStaff)
    }
    
    // Organization Admin - Resource Management (Organization scope)
    orgAdminResources := rg.Group("/org-admin/resources")
    orgAdminResources.Use(middleware.AuthMiddleware(config.DB))
    orgAdminResources.Use(middleware.RequireOrganizationAdmin(config.DB))
    {
        orgAdminResources.GET("/clinics", controllers.ListClinics)
        orgAdminResources.GET("/patients", controllers.ListPatients)
        orgAdminResources.GET("/doctors", controllers.ListDoctors)
        orgAdminResources.GET("/staff", controllers.ListStaff)
    }
    
    // Clinic Admin - Resource Management (Clinic scope)
    clinicAdminResources := rg.Group("/clinic-admin/resources")
    clinicAdminResources.Use(middleware.AuthMiddleware(config.DB))
    clinicAdminResources.Use(middleware.RequireClinicAdmin(config.DB))
    {
        clinicAdminResources.GET("/clinics", controllers.ListClinics)
        clinicAdminResources.GET("/patients", controllers.ListPatients)
        clinicAdminResources.GET("/doctors", controllers.ListDoctors)
        clinicAdminResources.GET("/staff", controllers.ListStaff)
    }
    
    // Any authenticated user - Resource Management (Role-based scope)
    // Doctors, Receptionists, Pharmacy, Lab staff can access this
    resources := rg.Group("/resources")
    resources.Use(middleware.AuthMiddleware(config.DB))
    {
        resources.GET("/clinics", controllers.ListClinics)
        resources.GET("/patients", controllers.ListPatients)
        resources.GET("/doctors", controllers.ListDoctors)
        resources.GET("/staff", controllers.ListStaff)
    }
}
