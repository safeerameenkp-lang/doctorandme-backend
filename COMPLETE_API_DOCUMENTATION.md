# Complete API Documentation - Microservices

## 🌐 Base URLs

- **Kong Gateway (Main Entry Point)**: `http://localhost:8000`
- **Auth Service**: `http://localhost:8000/api/auth`
- **Organization Service**: `http://localhost:8000/api/organizations`
- **Appointment Service**: `http://localhost:8000/api/v1`

---

## 🔐 Authentication

All protected endpoints require JWT token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Getting Access Token

Use the Login API to get an access token:

```javascript
// Login Request
POST http://localhost:8000/api/auth/login
Content-Type: application/json

{
  "login": "user@example.com",  // email, phone, or username
  "password": "password123"
}

// Response
{
  "id": "user-uuid",
  "firstName": "John",
  "lastName": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "roles": ["user", "doctor"],
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600
}
```

---

## 📡 Auth Service API

### Base Path: `/api/auth`

#### Public Endpoints

##### 1. Health Check
```
GET /api/auth/health
```
**Response:**
```json
{
  "status": "healthy",
  "service": "auth-service",
  "timestamp": 1704067200
}
```

##### 2. Register
```
POST /api/auth/register
```
**Request:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "phone": "1234567890",
  "password": "password123"
}
```

##### 3. Login ⭐
```
POST /api/auth/login
```
**Request:**
```json
{
  "login": "user@example.com",  // email, phone, or username
  "password": "password123"
}
```
**Response:**
```json
{
  "id": "user-uuid",
  "firstName": "John",
  "lastName": "Doe",
  "email": "user@example.com",
  "username": "johndoe",
  "phone": "1234567890",
  "roles": ["user", "doctor"],
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600
}
```

##### 4. Refresh Token
```
POST /api/auth/refresh
```
**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

##### 5. Logout
```
POST /api/auth/logout
```
**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Protected Endpoints (Require Auth Token)

##### 6. Get Profile
```
GET /api/auth/profile
Headers: Authorization: Bearer <token>
```

##### 7. Update Profile
```
PUT /api/auth/profile
Headers: Authorization: Bearer <token>
```
**Request:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "newemail@example.com",
  "phone": "9876543210"
}
```

##### 8. Change Password
```
POST /api/auth/change-password
Headers: Authorization: Bearer <token>
```
**Request:**
```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

---

## 🏢 Organization Service API

### Base Path: `/api/organizations`

#### Public Endpoints

##### 1. Health Check
```
GET /api/organizations/health
```

#### Protected Endpoints (Require Auth Token)

##### Organizations

##### 2. Create Organization
```
POST /api/organizations/organizations
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 3. Create Organization with Admin
```
POST /api/organizations/organizations/with-admin
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 4. List Organizations
```
GET /api/organizations/organizations
Headers: Authorization: Bearer <token>
```

##### 5. Get Organization
```
GET /api/organizations/organizations/:id
Headers: Authorization: Bearer <token>
```

##### 6. Update Organization
```
PUT /api/organizations/organizations/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, organization_admin
```

##### 7. Delete Organization
```
DELETE /api/organizations/organizations/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

#### Clinics

##### 8. Create Clinic
```
POST /api/organizations/clinics
Headers: Authorization: Bearer <token>
Roles: super_admin, organization_admin
```

##### 9. Create Clinic with Admin
```
POST /api/organizations/clinics/with-admin
Headers: Authorization: Bearer <token>
Roles: super_admin, organization_admin
```

##### 10. List Clinics
```
GET /api/organizations/clinics
Headers: Authorization: Bearer <token>
Query Parameters:
  - organization_id (optional)
  - is_active (optional)
```

##### 11. Get Clinic
```
GET /api/organizations/clinics/:id
Headers: Authorization: Bearer <token>
```

##### 12. Update Clinic
```
PUT /api/organizations/clinics/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, organization_admin, clinic_admin
```

##### 13. Delete Clinic
```
DELETE /api/organizations/clinics/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, organization_admin
```

#### Doctors

##### 14. Create Doctor
```
POST /api/organizations/doctors
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 15. List Doctors
```
GET /api/organizations/doctors
Headers: Authorization: Bearer <token>
Query Parameters:
  - clinic_id (optional)
  - organization_id (optional)
```

##### 16. List All Doctors
```
GET /api/organizations/doctors/all
Headers: Authorization: Bearer <token>
```

##### 17. Get Doctor
```
GET /api/organizations/doctors/:id
Headers: Authorization: Bearer <token>
```

##### 18. Get Doctors by Clinic
```
GET /api/organizations/doctors/clinic/:clinic_id
Headers: Authorization: Bearer <token>
```

##### 19. Update Doctor
```
PUT /api/organizations/doctors/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

##### 20. Delete Doctor
```
DELETE /api/organizations/doctors/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

#### Doctor Time Slots

##### 21. Create Doctor Time Slots (Date-Specific)
```
POST /api/organizations/doctor-time-slots
Headers: Authorization: Bearer <token>
Roles: doctor, clinic_admin
```

##### 22. List Doctor Time Slots
```
GET /api/organizations/doctor-time-slots
Headers: Authorization: Bearer <token>
Query Parameters:
  - doctor_id (required)
  - clinic_id (optional)
  - slot_type (optional)
  - date (optional)
```

##### 23. Get Doctor Time Slot
```
GET /api/organizations/doctor-time-slots/:id
Headers: Authorization: Bearer <token>
```

##### 24. Update Doctor Time Slot
```
PUT /api/organizations/doctor-time-slots/:id
Headers: Authorization: Bearer <token>
Roles: doctor, clinic_admin
```

##### 25. Delete Doctor Time Slot
```
DELETE /api/organizations/doctor-time-slots/:id
Headers: Authorization: Bearer <token>
Roles: doctor, clinic_admin
```

#### Doctor Session Slots (Auto-Generated Individual Slots)

##### 26. Create Doctor Session Slots
```
POST /api/organizations/doctor-session-slots
Headers: Authorization: Bearer <token>
Roles: doctor, clinic_admin
```

##### 27. List Doctor Session Slots
```
GET /api/organizations/doctor-session-slots
Headers: Authorization: Bearer <token>
Query Parameters:
  - doctor_id (required)
  - clinic_id (optional)
  - date (optional)
  - slot_type (optional: clinic_visit, video_consultation, follow-up-via-clinic, follow-up-via-video)
```

##### 28. Sync Slot Booking Status
```
POST /api/organizations/doctor-session-slots/sync-booking-status
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

#### Doctor Leaves

##### 29. Apply for Leave
```
POST /api/organizations/doctor-leaves
Headers: Authorization: Bearer <token>
Roles: doctor, clinic_admin, receptionist
```

##### 30. List Doctor Leaves
```
GET /api/organizations/doctor-leaves
Headers: Authorization: Bearer <token>
Query Parameters:
  - clinic_id (optional)
  - doctor_id (optional)
  - status (optional)
  - leave_type (optional)
```

##### 31. Get Doctor Leave
```
GET /api/organizations/doctor-leaves/:id
Headers: Authorization: Bearer <token>
```

##### 32. Review Leave
```
POST /api/organizations/doctor-leaves/:id/review
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 33. Cancel Leave
```
POST /api/organizations/doctor-leaves/:id/cancel
Headers: Authorization: Bearer <token>
Roles: doctor
```

##### 34. Get Doctor Leave Stats
```
GET /api/organizations/doctor-leaves/stats/:doctor_id
Headers: Authorization: Bearer <token>
```

#### Clinic Doctor Links

##### 35. Create Clinic Doctor Link
```
POST /api/organizations/clinic-doctor-links
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 36. List Clinic Doctor Links
```
GET /api/organizations/clinic-doctor-links
Headers: Authorization: Bearer <token>
```

##### 37. Get Clinic Doctor Links by Doctor
```
GET /api/organizations/clinic-doctor-links/doctor/:doctor_id
Headers: Authorization: Bearer <token>
```

##### 38. Update Clinic Doctor Link
```
PUT /api/organizations/clinic-doctor-links/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 39. Delete Clinic Doctor Link
```
DELETE /api/organizations/clinic-doctor-links/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

#### Doctor Schedules

##### 40. Create Doctor Schedule
```
POST /api/organizations/doctor-schedules
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

##### 41. List Doctor Schedules
```
GET /api/organizations/doctor-schedules
Headers: Authorization: Bearer <token>
```

##### 42. Get Doctor Schedule
```
GET /api/organizations/doctor-schedules/:id
Headers: Authorization: Bearer <token>
```

##### 43. Update Doctor Schedule
```
PUT /api/organizations/doctor-schedules/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

##### 44. Delete Doctor Schedule
```
DELETE /api/organizations/doctor-schedules/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

#### Departments

##### 45. Create Department
```
POST /api/organizations/departments
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 46. List Departments
```
GET /api/organizations/departments
Headers: Authorization: Bearer <token>
```

##### 47. Get Department
```
GET /api/organizations/departments/:id
Headers: Authorization: Bearer <token>
```

##### 48. Update Department
```
PUT /api/organizations/departments/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 49. Delete Department
```
DELETE /api/organizations/departments/:id
Headers: Authorization: Bearer <token>
Roles: super_admin, clinic_admin
```

##### 50. Get Doctors by Department
```
GET /api/organizations/departments/:id/doctors
Headers: Authorization: Bearer <token>
```

#### Patients (Global)

##### 51. Create Patient (Global)
```
POST /api/organizations/patients
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 52. List Patients (Global)
```
GET /api/organizations/patients
Headers: Authorization: Bearer <token>
```

##### 53. Get Patient (Global)
```
GET /api/organizations/patients/:id
Headers: Authorization: Bearer <token>
```

##### 54. Update Patient (Global)
```
PUT /api/organizations/patients/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 55. Delete Patient (Global)
```
DELETE /api/organizations/patients/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

#### Clinic-Specific Patients ⭐ (Recommended)

##### 56. Create Clinic Patient
```
POST /api/organizations/clinic-specific-patients
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 57. List Clinic Patients
```
GET /api/organizations/clinic-specific-patients
Headers: Authorization: Bearer <token>
Query Parameters:
  - clinic_id (required)
  - search (optional)
  - only_active (optional)
```

##### 58. Get Clinic Patient
```
GET /api/organizations/clinic-specific-patients/:id
Headers: Authorization: Bearer <token>
```

##### 59. Update Clinic Patient
```
PUT /api/organizations/clinic-specific-patients/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 60. Delete Clinic Patient
```
DELETE /api/organizations/clinic-specific-patients/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

#### Patient-Clinic Assignments

##### 61. Assign Patient to Clinic
```
POST /api/organizations/patient-clinics
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 62. List Patient-Clinic Assignments
```
GET /api/organizations/patient-clinics
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 63. Get Patient-Clinic Assignment
```
GET /api/organizations/patient-clinics/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 64. Update Patient-Clinic Assignment
```
PUT /api/organizations/patient-clinics/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 65. Remove Patient from Clinic
```
DELETE /api/organizations/patient-clinics/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 66. Get Clinics by Patient
```
GET /api/organizations/patient-clinics/patient/:patient_id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

#### External Services

##### 67. Create External Service
```
POST /api/organizations/services
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 68. List External Services
```
GET /api/organizations/services
Headers: Authorization: Bearer <token>
```

##### 69. Get External Service
```
GET /api/organizations/services/:id
Headers: Authorization: Bearer <token>
```

##### 70. Update External Service
```
PUT /api/organizations/services/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 71. Delete External Service
```
DELETE /api/organizations/services/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

#### Clinic Service Links

##### 72. Create Clinic Service Link
```
POST /api/organizations/links
Headers: Authorization: Bearer <token>
Roles: super_admin
```

##### 73. List Clinic Service Links
```
GET /api/organizations/links
Headers: Authorization: Bearer <token>
```

##### 74. Get Clinic Service Link
```
GET /api/organizations/links/:id
Headers: Authorization: Bearer <token>
```

##### 75. Delete Clinic Service Link
```
DELETE /api/organizations/links/:id
Headers: Authorization: Bearer <token>
Roles: super_admin
```

#### Admin Panel Routes

##### 76. Create Staff
```
POST /api/organizations/admin/staff
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 77. Get Clinic Staff
```
GET /api/organizations/admin/staff/clinic/:clinic_id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 78. Update Staff Role
```
PUT /api/organizations/admin/staff/clinic/:clinic_id/:user_id/role
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 79. Deactivate Staff
```
DELETE /api/organizations/admin/staff/clinic/:clinic_id/:user_id
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 80. Create Queue
```
POST /api/organizations/admin/queues
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 81. Get Queues
```
GET /api/organizations/admin/queues
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 82. Assign Token
```
POST /api/organizations/admin/queues/tokens
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 83. Reassign Token
```
PUT /api/organizations/admin/queues/tokens/:token_id/reassign
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 84. Pause Queue
```
PUT /api/organizations/admin/queues/:queue_id/pause
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### 85. Resume Queue
```
PUT /api/organizations/admin/queues/:queue_id/resume
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

---

## 📅 Appointment Service API

### Base Path: `/api/v1`

#### Public Endpoints

##### 1. Health Check
```
GET /api/v1/health
```

#### Protected Endpoints (Require Auth Token)

##### Appointments - Simple (Recommended for Frontend)

##### 2. Create Simple Appointment ⭐
```
POST /api/v1/appointments/simple
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```
**Request:**
```json
{
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_patient_id": "patient-uuid",
  "appointment_date": "2024-01-15",
  "appointment_time": "10:00:00",
  "consultation_type": "clinic_visit",
  "reason": "Regular checkup"
}
```

##### 3. Get Simple Appointment List
```
GET /api/v1/appointments/simple-list
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist, doctor
Query Parameters:
  - clinic_id (optional)
  - doctor_id (optional)
  - date (optional, format: YYYY-MM-DD)
  - status (optional)
```

##### 4. Get Simple Appointment Details
```
GET /api/v1/appointments/simple/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist, doctor
```

##### 5. Reschedule Simple Appointment
```
POST /api/v1/appointments/simple/:id/reschedule
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```
**Request:**
```json
{
  "appointment_date": "2024-01-20",
  "appointment_time": "14:00:00",
  "individual_slot_id": "slot-uuid"
}
```

##### Appointments - Advanced

##### 6. Create Appointment (Multiple Patient Types)
```
POST /api/v1/appointments
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```
**Request:**
```json
{
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  "user_id": "user-uuid",  // Optional - one of: user_id, patient_id, clinic_patient_id, mobile_no, mo_id
  "appointment_date": "2024-01-15",
  "appointment_time": "2024-01-15 10:00:00",
  "consultation_type": "clinic_visit",  // video, in_person, offline, online, follow_up, clinic_visit
  "duration_minutes": 30,
  "reason": "Regular checkup",
  "slot_id": "slot-uuid",  // Optional
  "individual_slot_id": "individual-slot-uuid"  // Optional
}
```

##### 7. Create Patient with Appointment
```
POST /api/v1/appointments/with-patient
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 8. List Appointments
```
GET /api/v1/appointments
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
Query Parameters:
  - doctor_id (optional)
  - clinic_id (optional)
  - patient_id (optional)
  - status (optional)
  - date (optional)
```

##### 9. Get Appointment List (Table Format)
```
GET /api/v1/appointments/list
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 10. Get Single Appointment
```
GET /api/v1/appointments/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 11. Update Appointment
```
PUT /api/v1/appointments/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 12. Reschedule Appointment
```
POST /api/v1/appointments/:id/reschedule
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 13. Cancel Appointment
```
POST /api/v1/appointments/:id/cancel
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 14. Get Available Time Slots
```
GET /api/v1/appointments/slots/available
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
Query Parameters:
  - doctor_id (required)
  - clinic_id (required)
  - date (required, format: YYYY-MM-DD)
```

##### Follow-up Management

##### 15. Check Follow-up Eligibility
```
GET /api/v1/appointments/followup-eligibility
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist, doctor
Query Parameters:
  - patient_id (required)
  - appointment_id (required)
```

##### 16. List Active Follow-ups
```
GET /api/v1/appointments/followup-eligibility/active
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist, doctor
```

##### 17. Expire Old Follow-ups
```
POST /api/v1/appointments/followup-eligibility/expire-old
Headers: Authorization: Bearer <token>
Roles: clinic_admin
```

##### Check-ins

##### 18. Create Check-in
```
POST /api/v1/checkins
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```
**Request:**
```json
{
  "appointment_id": "appointment-uuid",
  "check_in_time": "2024-01-15 09:45:00"
}
```

##### 19. List Check-ins
```
GET /api/v1/checkins
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 20. Get Check-in
```
GET /api/v1/checkins/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 21. Update Check-in
```
PUT /api/v1/checkins/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 22. Get Doctor Queue
```
GET /api/v1/checkins/doctor/:doctor_id/queue
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### Vitals

##### 23. Record Vitals
```
POST /api/v1/vitals
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```
**Request:**
```json
{
  "appointment_id": "appointment-uuid",
  "blood_pressure": "120/80",
  "temperature": 98.6,
  "pulse": 72,
  "weight": 70.5,
  "height": 175.0,
  "spo2": 98,
  "notes": "Patient appears healthy"
}
```

##### 24. List Vitals
```
GET /api/v1/vitals
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 25. Get Vitals by Appointment
```
GET /api/v1/vitals/appointment/:appointment_id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### 26. Update Vitals
```
PUT /api/v1/vitals/:id
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

##### 27. Get Patient Vitals History
```
GET /api/v1/vitals/patient/:patient_id/history
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor, receptionist
```

##### Reports

##### 28. Daily Collection Report
```
GET /api/v1/reports/daily-collection
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
Query Parameters:
  - date (optional, format: YYYY-MM-DD)
  - clinic_id (optional)
```

##### 29. Pending Payments Report
```
GET /api/v1/reports/pending-payments
Headers: Authorization: Bearer <token>
Roles: clinic_admin, receptionist
```

##### 30. Utilization Report
```
GET /api/v1/reports/utilization
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

##### 31. No-Show Report
```
GET /api/v1/reports/no-show
Headers: Authorization: Bearer <token>
Roles: clinic_admin, doctor
```

---

## 🔄 Error Responses

All APIs return standardized error responses:

### Validation Error (400)
```json
{
  "error": "Validation failed",
  "message": "Invalid input data",
  "code": "VALIDATION_ERROR",
  "details": "Field validation error details"
}
```

### Unauthorized (401)
```json
{
  "error": "Authentication required",
  "message": "Please provide a valid authorization token",
  "code": "MISSING_TOKEN"
}
```

### Forbidden (403)
```json
{
  "error": "Insufficient permissions",
  "message": "Access denied. This resource requires admin role",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

### Not Found (404)
```json
{
  "error": "Resource not found",
  "message": "The requested resource was not found",
  "code": "RESOURCE_NOT_FOUND"
}
```

### Database Error (500)
```json
{
  "error": "Database error",
  "message": "Database operation failed",
  "code": "DATABASE_ERROR"
}
```

---

## 💻 Frontend Integration Examples

### JavaScript/TypeScript

#### Setup Axios Instance
```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh token
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const response = await axios.post('http://localhost:8000/api/auth/refresh', {
            refresh_token: refreshToken,
          });
          localStorage.setItem('accessToken', response.data.accessToken);
          localStorage.setItem('refreshToken', response.data.refreshToken);
          // Retry original request
          error.config.headers.Authorization = `Bearer ${response.data.accessToken}`;
          return axios.request(error.config);
        } catch (refreshError) {
          // Redirect to login
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export default api;
```

#### Login Function
```javascript
async function login(login, password) {
  try {
    const response = await api.post('/api/auth/login', {
      login,
      password,
    });
    
    // Store tokens
    localStorage.setItem('accessToken', response.data.accessToken);
    localStorage.setItem('refreshToken', response.data.refreshToken);
    localStorage.setItem('user', JSON.stringify(response.data));
    
    return response.data;
  } catch (error) {
    console.error('Login failed:', error.response?.data);
    throw error;
  }
}
```

#### Get Profile
```javascript
async function getProfile() {
  try {
    const response = await api.get('/api/auth/profile');
    return response.data;
  } catch (error) {
    console.error('Get profile failed:', error.response?.data);
    throw error;
  }
}
```

#### List Organizations
```javascript
async function getOrganizations() {
  try {
    const response = await api.get('/api/organizations/organizations');
    return response.data;
  } catch (error) {
    console.error('Get organizations failed:', error.response?.data);
    throw error;
  }
}
```

#### Create Appointment
```javascript
async function createAppointment(appointmentData) {
  try {
    const response = await api.post('/api/v1/appointments', appointmentData);
    return response.data;
  } catch (error) {
    console.error('Create appointment failed:', error.response?.data);
    throw error;
  }
}
```

#### Get Time Slots
```javascript
async function getTimeSlots(doctorId, clinicId, date) {
  try {
    const response = await api.get('/api/v1/appointments/time-slots', {
      params: {
        doctor_id: doctorId,
        clinic_id: clinicId,
        date: date, // Format: YYYY-MM-DD
      },
    });
    return response.data;
  } catch (error) {
    console.error('Get time slots failed:', error.response?.data);
    throw error;
  }
}
```

---

## 📝 React Hook Examples

### useAuth Hook
```javascript
import { useState, useEffect } from 'react';
import api from './api';

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      api.get('/api/auth/profile')
        .then((response) => {
          setUser(response.data);
        })
        .catch(() => {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (login, password) => {
    const response = await api.post('/api/auth/login', { login, password });
    localStorage.setItem('accessToken', response.data.accessToken);
    localStorage.setItem('refreshToken', response.data.refreshToken);
    setUser(response.data);
    return response.data;
  };

  const logout = async () => {
    const refreshToken = localStorage.getItem('refreshToken');
    if (refreshToken) {
      await api.post('/api/auth/logout', { refresh_token: refreshToken });
    }
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    setUser(null);
  };

  return { user, loading, login, logout };
}
```

### useAppointments Hook
```javascript
import { useState, useEffect } from 'react';
import api from './api';

export function useAppointments(filters = {}) {
  const [appointments, setAppointments] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get('/api/v1/appointments', { params: filters })
      .then((response) => {
        setAppointments(response.data);
      })
      .catch((error) => {
        console.error('Failed to fetch appointments:', error);
      })
      .finally(() => setLoading(false));
  }, [JSON.stringify(filters)]);

  return { appointments, loading };
}
```

---

## 🔑 Important Notes for Frontend

1. **Token Storage**: Store `accessToken` and `refreshToken` securely (localStorage or httpOnly cookies)

2. **Token Expiry**: 
   - Access Token: 1 hour (3600 seconds)
   - Refresh Token: 7 days
   - Implement automatic token refresh before expiry

3. **CORS**: All services have CORS enabled for `*` origins

4. **Error Handling**: Always check `error.response?.data` for detailed error messages

5. **Date Formats**:
   - Date: `YYYY-MM-DD` (e.g., "2024-01-15")
   - DateTime: `YYYY-MM-DD HH:MM:SS` (e.g., "2024-01-15 10:00:00")

6. **Pagination**: Some list endpoints may support pagination (check response structure)

7. **Role-Based Access**: Some endpoints require specific roles (super_admin, organization_admin, clinic_admin)

---

## 🧪 Testing Endpoints

### Using Postman Collection

Import this collection structure:

```json
{
  "info": {
    "name": "DrAndMe Microservices API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8000"
    },
    {
      "key": "accessToken",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"login\": \"user@example.com\",\n  \"password\": \"password123\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/api/auth/login",
              "host": ["{{baseUrl}}"],
              "path": ["api", "auth", "login"]
            }
          }
        }
      ]
    }
  ]
}
```

---

## ✅ Quick Reference

| Service | Base Path | Main Endpoints | Total Endpoints |
|---------|-----------|----------------|----------------|
| Auth | `/api/auth` | login, register, profile, refresh | 8+ |
| Organization | `/api/organizations` | organizations, clinics, doctors, patients, time-slots | 85+ |
| Appointment | `/api/v1` | appointments, checkins, vitals, reports | 31+ |

## 📊 Endpoint Summary by Category

### Auth Service (8 endpoints)
- **Public**: health, register, login, refresh, logout
- **Protected**: profile (GET/PUT), change-password

### Organization Service (85+ endpoints)
- **Organizations**: 7 endpoints
- **Clinics**: 6 endpoints
- **Doctors**: 7 endpoints
- **Doctor Time Slots**: 5 endpoints
- **Doctor Session Slots**: 3 endpoints
- **Doctor Leaves**: 6 endpoints
- **Clinic Doctor Links**: 5 endpoints
- **Doctor Schedules**: 5 endpoints
- **Departments**: 6 endpoints
- **Patients (Global)**: 5 endpoints
- **Clinic-Specific Patients**: 5 endpoints ⭐
- **Patient-Clinic Assignments**: 6 endpoints
- **External Services**: 5 endpoints
- **Clinic Service Links**: 4 endpoints
- **Admin Panel**: 10+ endpoints (staff, queues, pharmacy, lab, etc.)

### Appointment Service (31+ endpoints)
- **Simple Appointments**: 5 endpoints ⭐ (Recommended)
- **Advanced Appointments**: 9 endpoints
- **Follow-ups**: 3 endpoints
- **Check-ins**: 5 endpoints
- **Vitals**: 5 endpoints
- **Reports**: 4 endpoints

## 🎯 Most Used Endpoints for Frontend

### Authentication Flow
1. `POST /api/auth/login` - Login
2. `GET /api/auth/profile` - Get user profile
3. `POST /api/auth/refresh` - Refresh token

### Appointment Booking Flow
1. `GET /api/organizations/clinics` - List clinics
2. `GET /api/organizations/doctors/clinic/:clinic_id` - Get doctors by clinic
3. `GET /api/organizations/doctor-session-slots` - Get available slots
4. `GET /api/organizations/clinic-specific-patients?clinic_id=xxx` - Search patients
5. `POST /api/v1/appointments/simple` - Create appointment ⭐

### Patient Management Flow
1. `POST /api/organizations/clinic-specific-patients` - Create patient
2. `GET /api/organizations/clinic-specific-patients?clinic_id=xxx` - List patients
3. `GET /api/organizations/clinic-specific-patients/:id` - Get patient details

### Appointment Management Flow
1. `GET /api/v1/appointments/simple-list` - List appointments
2. `GET /api/v1/appointments/simple/:id` - Get appointment details
3. `POST /api/v1/appointments/simple/:id/reschedule` - Reschedule
4. `POST /api/v1/checkins` - Check-in patient
5. `POST /api/v1/vitals` - Record vitals

---

## 📞 Support

For API issues or questions, check:
- Service health: `GET /api/{service}/health`
- Kong Admin: `http://localhost:8001`
- Service logs: `docker-compose logs {service-name}`

