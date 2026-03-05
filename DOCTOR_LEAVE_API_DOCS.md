# Doctor Leave Management API Documentation

This document outlines the API endpoints for managing doctor leaves, designed for the frontend team.

## Base URL
`/organizations/doctor-leaves`

---

## 1. Apply for Leave
Submit a new leave application for a doctor.

- **Endpoint**: `POST /organizations/doctor-leaves`
- **Access**: Doctor, Clinic Admin, Receptionist

### Request Body
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `doctor_id` | string (UUID) | Yes | The ID of the doctor applying for leave. |
| `clinic_id` | string (UUID) | Yes | The ID of the clinic where the leave applies. |
| `leave_type` | string | Yes | The category/reason for leave. See **Leave Types** below. |
| `leave_duration`| string | Yes | Which part of the day is affected. See **Leave Durations** below. |
| `from_date` | string (Date) | Yes | Start date in `YYYY-MM-DD` format. |
| `to_date` | string (Date) | Yes | End date in `YYYY-MM-DD` format. |
| `reason` | string | Yes | Detailed reason/notes for the leave (min 10 chars). |

### Allowed Leave Types (Reasons)
- `sick_leave` - Health related
- `casual_leave` - Personal/Casual reasons
- `vacation` - Planned vacation
- `emergency` - Urgent/Emergency leave
- `other` - Other reasons

### Allowed Leave Durations (Blocking Control)
- `morning` - Blocks slots **before 12:00 PM** only.
- `afternoon` - Blocks slots **from 12:00 PM onwards** only.
- `full_day` - Blocks **all** slots for the scheduled date(s).

### Example Request
```json
{
  "doctor_id": "a1b2c3d4-e5f6-...",
  "clinic_id": "c1c2c3c4-d5d6-...",
  "leave_type": "casual_leave",
  "leave_duration": "morning",
  "from_date": "2024-03-20",
  "to_date": "2024-03-20",
  "reason": "Personal work in the morning"
}
```

### Success Response (201 Created)
```json
{
  "message": "Leave application submitted successfully",
  "leave_id": "l1l2l3l4-...",
  "status": "pending",
  "total_days": 1
}
```

---

## 2. List Doctor Leaves
Retrieve a paginated list of leave applications with optional filtering.

- **Endpoint**: `GET /organizations/doctor-leaves`
- **Access**: Doctor (own leaves), Clinic Admin, Receptionist

### Query Parameters
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `clinic_id` | string (UUID) | No | Filter by clinic. |
| `doctor_id` | string (UUID) | No | Filter by doctor. |
| `status` | string | No | Filter by status (`pending`, `approved`, `rejected`, `cancelled`). |
| `leave_type` | string | No | Filter by leave type. |
| `page` | int | No | Page number (default: 1). |
| `page_size` | int | No | Items per page (default: 20). |

### Success Response (200 OK)
Includes `leave_duration` in the objects.

---

## 4. Update Leave Application
Update a pending or approved leave application.

- **Endpoint**: `PUT /organizations/doctor-leaves/:id`
- **Access**: Doctor (owner only), Clinic Admin

### Request Body
Same as **Apply for Leave**.

---

## 7. Slot Blocking Logic
The system automatically disables time slots based on the `leave_duration` field.

| Leave Duration | Blocked Time Range | Example |
| :--- | :--- | :--- |
| `morning` | **Before 12:00 PM** | Slots at 09:00, 10:30 are BLOCKED. Slots at 12:00, 14:00 are AVAILABLE. |
| `afternoon` | **12:00 PM and later** | Slots at 09:00 are AVAILABLE. Slots at 12:00, 14:00 are BLOCKED. |
| `full_day` | **All Day** | All slots for the date are BLOCKED. |

### Visual Identification
Blocked slots in `GET /doctor-session-slots`:
- `is_bookable`: `false`
- `status`: `"blocked"`
- `display_message`: `"Doctor on Leave"`
