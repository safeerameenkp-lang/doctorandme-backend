# 🔄 Appointment Reschedule API Guide

This document details the **Reschedule Flow** for the frontend, ensuring proper slot management (releasing old slots and booking new ones).

---

## 📍 API Endpoint

**POST** `/appointments/simple/:id/reschedule`

### Base URL
`{{baseUrl}}/api/v1/appointments/simple/{appointment_id}/reschedule`

---

## 📥 Request Body (JSON)

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `doctor_id` | UUID | ✅ **Yes** | ID of the doctor (can be same or new). | `d123...` |
| `individual_slot_id` | UUID | ✅ **Yes** | ID of the **NEW** time slot selected. | `s456...` |
| `appointment_date` | Date | ✅ **Yes** | New date (YYYY-MM-DD). | `2026-02-25` |
| `appointment_time` | Time | ✅ **Yes** | New time (YYYY-MM-DD HH:MM:SS). | `2026-02-25 10:00:00` |
| `consultation_type` | String | ❌ No | Type: "clinic_visit", "video_consultation". | `clinic_visit` |
| `reason` | String | ❌ No | Reason for rescheduling. | `Patient request` |
| `notes` | String | ❌ No | Additional notes. | `Urgent` |

### Example Request
```json
{
  "doctor_id": "d12345-uuid",
  "individual_slot_id": "s98765-uuid",
  "appointment_date": "2026-02-25",
  "appointment_time": "2026-02-25 10:00:00",
  "consultation_type": "video_consultation",
  "reason": "Traffic delay",
  "notes": "Please inform doctor"
}
```

---

## 🧠 Backend Logic (What happens automatically)

When you call this API, the backend performs the following atomic operations:

1.  **Validate New Slot**: Checks if the target slot is truly available.
2.  **Release Old Slot** (The Fix):
    -   The PREVIOUS slot ID (from the appointment) is identified.
    -   Its `available_count` is **Incremented** (+1).
    -   Its `status` is forced to **'available'** (and `is_booked = false`).
    -   This ensures the old slot immediately becomes bookable for other patients.
3.  **Book New Slot**:
    -   The NEW slot's `available_count` is **Decremented** (-1).
    -   If count reaches 0, its `status` becomes **'booked'**.
4.  **Update Appointment**:
    -   The appointment record is updated with the new date, time, and slot ID.

---

## 📤 Response

Returns the updated appointment details.

```json
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "a123...",
    "status": "scheduled",
    "appointment_date_time": "2026-02-25 10:00:00",
    "doctor_name": "Dr. Smith",
    ...
  }
}
```

## ⚠️ Frontend Tips

-   **Don't** assume the old slot is free until the API returns success.
-   **Do** refresh the "Slots List" after rescheduling if you are keeping it open, as availabilities have changed.
