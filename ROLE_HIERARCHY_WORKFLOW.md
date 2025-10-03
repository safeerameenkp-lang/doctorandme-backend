# Role Hierarchy Workflow Guide

This document demonstrates the complete workflow for implementing the hierarchical role management system with proper service separation.

## 🏗️ Service Architecture

```
Auth Service (Port 8080)
├── Authentication & Authorization
├── JWT Token Management
├── User Profile Management
└── Role-based Access Control

Organization Service (Port 8081)
├── Organization Management
├── Clinic Management
├── Admin Account Creation
└── Staff Management
```

## 📋 Complete Workflow Example

### Step 1: Super Admin Creates Organization with Admin Account

**1.1 Create Organization with Admin Account**
```bash
POST /api/organizations/organizations/with-admin
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
    "name": "MediCare Health Systems",
    "email": "admin@medicare.com",
    "phone": "+1234567890",
    "address": "123 Medical Plaza, Health City",
    "license_number": "MED-2024-001",
    "admin_first_name": "John",
    "admin_last_name": "Smith",
    "admin_email": "john.smith@medicare.com",
    "admin_username": "johnsmith_admin",
    "admin_phone": "+1234567891",
    "admin_password": "SecurePass123!"
}
```

### Step 2: Organization Admin Creates Clinic with Admin Account

**2.1 Create Clinic with Admin Account**
```bash
POST /api/organizations/clinics/with-admin
Authorization: Bearer <organization_admin_token>
Content-Type: application/json

{
    "organization_id": "<organization_id_from_step_1.1>",
    "clinic_code": "MC001",
    "name": "MediCare Downtown Clinic",
    "email": "downtown@medicare.com",
    "phone": "+1234567892",
    "address": "456 Main Street, Downtown",
    "license_number": "CLINIC-001",
    "admin_first_name": "Sarah",
    "admin_last_name": "Johnson",
    "admin_email": "sarah.johnson@medicare.com",
    "admin_username": "sarahjohnson_clinic",
    "admin_phone": "+1234567893",
    "admin_password": "ClinicPass123!"
}
```

### Step 3: Clinic Admin Creates Staff Members

**3.1 Create Doctor**
```bash
POST /api/organizations/staff
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id_from_step_2.1>",
    "role_name": "doctor",
    "first_name": "Dr. Michael",
    "last_name": "Brown",
    "email": "michael.brown@medicare.com",
    "username": "drbrown",
    "phone": "+1234567894",
    "password": "DoctorPass123!"
}
```

**3.2 Create Receptionist**
```bash
POST /api/organizations/staff
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id>",
    "role_name": "receptionist",
    "first_name": "Emily",
    "last_name": "Davis",
    "email": "emily.davis@medicare.com",
    "username": "emilydavis",
    "phone": "+1234567895",
    "password": "ReceptionPass123!"
}
```

**3.3 Create Pharmacist**
```bash
POST /api/organizations/staff
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id>",
    "role_name": "pharmacist",
    "first_name": "Robert",
    "last_name": "Wilson",
    "email": "robert.wilson@medicare.com",
    "username": "robertwilson",
    "phone": "+1234567896",
    "password": "PharmacyPass123!"
}
```

**3.4 Create Lab Technician**
```bash
POST /api/organizations/staff
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id>",
    "role_name": "lab_technician",
    "first_name": "Lisa",
    "last_name": "Anderson",
    "email": "lisa.anderson@medicare.com",
    "username": "lisaanderson",
    "phone": "+1234567897",
    "password": "LabPass123!"
}
```

**3.5 Create Billing Staff**
```bash
POST /api/organizations/staff
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id>",
    "role_name": "billing_staff",
    "first_name": "David",
    "last_name": "Miller",
    "email": "david.miller@medicare.com",
    "username": "davidmiller",
    "phone": "+1234567898",
    "password": "BillingPass123!"
}
```

### Step 4: Clinic Admin Manages Staff

**4.1 View All Staff**
```bash
GET /api/organizations/staff/clinic/<clinic_id>
Authorization: Bearer <clinic_admin_token>
```

**4.2 Update Staff Role**
```bash
PUT /api/organizations/staff/clinic/<clinic_id>/<user_id>/role
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "role_name": "doctor"
}
```

**4.3 Deactivate Staff Member**
```bash
DELETE /api/organizations/staff/clinic/<clinic_id>/<user_id>
Authorization: Bearer <clinic_admin_token>
```

### Step 5: Doctor Management

**5.1 Create Main Doctor (Super Admin only)**
```bash
POST /api/organizations/doctors/main
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
    "user_id": "<user_id_from_step_3.1>",
    "doctor_code": "MAIN001",
    "specialization": "Cardiology",
    "license_number": "MD-2024-001",
    "consultation_fee": 200.00,
    "follow_up_fee": 150.00,
    "follow_up_days": 7
}
```

**5.2 Link Main Doctor to Clinic**
```bash
POST /api/organizations/clinic-doctor-links
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "clinic_id": "<clinic_id>",
    "doctor_id": "<main_doctor_id_from_step_5.1>"
}
```

**5.3 Create Regular Doctor**
```bash
POST /api/organizations/doctors
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "user_id": "<user_id_from_step_3.2>",
    "clinic_id": "<clinic_id>",
    "doctor_code": "DR001",
    "specialization": "General Medicine",
    "license_number": "MD-2024-002",
    "consultation_fee": 150.00,
    "follow_up_fee": 100.00,
    "follow_up_days": 7,
    "is_main_doctor": false
}
```

**5.4 Create Doctor Schedule**
```bash
POST /api/organizations/doctor-schedules
Authorization: Bearer <clinic_admin_token>
Content-Type: application/json

{
    "doctor_id": "<doctor_id_from_step_5.3>",
    "day_of_week": 1,
    "start_time": "09:00",
    "end_time": "17:00",
    "slot_duration_minutes": 15
}
```

## 🔐 Service Responsibilities

### Auth Service (Port 8080)
- ✅ User authentication (login/logout)
- ✅ JWT token management
- ✅ User profile management
- ✅ Password changes
- ✅ Role-based access control verification

### Organization Service (Port 8081)
- ✅ Organization management
- ✅ Clinic management
- ✅ Admin account creation
- ✅ Staff management
- ✅ Role assignment within clinics
- ✅ Business entity operations

## 🚀 Getting Started

1. **Start the services**: `docker-compose up --build`
2. **Create a super admin user** through registration
3. **Follow the workflow** above to set up your organization
4. **Test the endpoints** using the provided test script

## 📊 Database Schema

The system uses the following key tables:
- `users` - All user accounts
- `roles` - Role definitions with permissions
- `organizations` - Organization entities
- `clinics` - Clinic entities linked to organizations
- `user_roles` - Links users to roles within specific contexts (organization/clinic/service)

## 🔒 Security Features

- **JWT Authentication** with access and refresh tokens
- **Role-based Access Control** with hierarchical permissions
- **Password Hashing** using bcrypt
- **Input Validation** on all endpoints
- **CORS Protection** with configurable origins
- **Token Expiration** and refresh token rotation
- **Service Separation** for better security and maintainability

This system provides a complete, scalable solution for managing healthcare organizations with proper service separation, role hierarchy, and security.
