package models

import (
    "time"
)

type User struct {
    ID           string     `json:"id" db:"id"`
    Email        *string    `json:"email" db:"email"`
    Username     string     `json:"username" db:"username"`
    PasswordHash string     `json:"-" db:"password_hash"`
    FirstName    string     `json:"first_name" db:"first_name"`
    LastName     string     `json:"last_name" db:"last_name"`
    Phone        *string    `json:"phone" db:"phone"`
    DateOfBirth  *time.Time `json:"date_of_birth" db:"date_of_birth"`
    Gender       *string    `json:"gender" db:"gender"`
    IsActive     bool       `json:"is_active" db:"is_active"`
    LastLogin    *time.Time `json:"last_login" db:"last_login"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type Role struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    Permissions map[string]interface{} `json:"permissions" db:"permissions"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

type UserRole struct {
    ID             string     `json:"id" db:"id"`
    UserID         string     `json:"user_id" db:"user_id"`
    RoleID         string     `json:"role_id" db:"role_id"`
    OrganizationID *string    `json:"organization_id" db:"organization_id"`
    ClinicID       *string    `json:"clinic_id" db:"clinic_id"`
    ServiceID      *string    `json:"service_id" db:"service_id"`
    IsActive       bool       `json:"is_active" db:"is_active"`
    AssignedAt     time.Time  `json:"assigned_at" db:"assigned_at"`
}

type RefreshToken struct {
    ID        string     `json:"id" db:"id"`
    UserID    string     `json:"user_id" db:"user_id"`
    Token     string     `json:"token" db:"token"`
    ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
}
