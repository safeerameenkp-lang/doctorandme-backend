# Backend Documentation (DrAndMe)

This document provides a comprehensive overview of the backend architecture, services, and features for the DrAndMe / Kwine platform.

## 1. Overview
The backend is a **Microservices Architecture** built with **Go (Golang)**, designed for scalability, modularity, and high performance. It uses **Docker** for containerization and **Kong** (implied) for API Gateway routing.

## 2. Tools & Technologies
- **Language**: Go (Golang)
- **Database**: PostgreSQL (implied by SQL files)
- **Containerization**: Docker & Docker Compose
- **API Gateway**: Kong
- **Authentication**: JWT (JSON Web Tokens)
- **Protocol**: REST API (JSON)

---

## 3. Service Breakdown

### 🏥 1. Appointment Service
Handles all logic related to booking, rescheduling, and managing appointments.

- **Location**: `services/appointment-service/`
- **Key Features**:
  - **Slot Management**: Generates and validates time slots (5-min, 15-min).
  - **Booking**: Creates appointments with validation (double-booking prevention).
  - **Follow-Ups**: Check eligibility for free follow-ups within validity period.
  - **Rescheduling**: Modifies existing appointments and updates slot status.
- **State/Storage**:
  - `appointments` table
  - `doctor_schedule` table

### 🏢 2. Organization Service
Manages clinic data, doctors, and patient relationships.

- **Location**: `services/organization-service/`
- **Key Features**:
  - **Clinic Management**: CRUD operations for clinics.
  - **Doctor-Clinic Linking**: Associates doctors with specific clinics.
  - **Patient Records**: Stores patient demographics and clinic associations.
- **State/Storage**:
  - `clinics` table
  - `clinic_doctors` table
  - `clinic_patients` table

### 🔐 3. Auth Service
Handles user authentication and authorization.

- **Location**: `services/auth-service/`
- **Key Features**:
  - **Login/Register**: Issues JWT tokens.
  - **Role-Based Access Control (RBAC)**: Validates permissions (Admin, Doctor, Patient).
  - **Token Refresh**: Manages session validity.
- **State/Storage**:
  - `users` table
  - `roles` table

---

## 4. Key Features & Implementation Details

### A. Follow-Up Logic
- **Tools**: Custom Go logic for date calculation.
- **Process**:
  1. Calculate days since last appointment.
  2. Check if within validity period (e.g., 7 days).
  3. Determine if "free" slot is available.
- **Endpoints**:
  - `GET /api/appointments/followup-eligibility`

### B. Time Slot Generation
- **Tools**: Go `time` package.
- **Process**:
  1. Fetch doctor's available hours.
  2. Generate discrete slots (e.g., 09:00, 09:15).
  3. Filter out booked slots.
- **Endpoints**:
  - `GET /api/appointments/slots`

### C. Security
- **Tools**: JWT Middleware.
- **Implementation**:
  - All protected routes require `Authorization: Bearer <token>` header.
  - Middleware validates token signature and expiration.

---

## 5. Deployment
- **Docker**: Each service runs in its own container.
- **Compose**: `docker-compose.yml` orchestrates all services and the database.
- **Build Command**: `./build-each-service.ps1` or `docker-compose up --build`.

## 6. Directory Structure

```
drandme-backend/
├── services/
│   ├── appointment-service/  # Booking Logic
│   ├── auth-service/         # User Auth
│   └── organization-service/ # Clinic/Doctor Data
├── shared/                   # Shared Code/Utils
├── docker-compose.yml        # Orchestration
├── Kong/                     # Gateway Config
└── scripts/                  # Build/Test Scripts
```
