# 📊 Appointment Summary API Guide

This document details the newly created Appointment Summary API, designed for the Clinic Dashboard to show real-time appointment counts.

---

## 📍 API Endpoint

**GET** `/appointments/summary`

### Base URL
`{{baseUrl}}/api/v1/appointments/summary`

---

## 📥 Request Parameters (Query)

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `clinic_id` | UUID | ✅ **Yes** | The ID of the clinic. | `e93b535c...` |
| `date` | Date (YYYY-MM-DD) | ❌ No | Filter by date. Defaults to **Today**. | `2026-02-19` |
| `doctor_id` | UUID | ❌ No | Filter by specific doctor. Use "all" or omit for all doctors. | `d123...` |
| `status` | String | ❌ No | Filter by status (e.g., "arrived"). Use "all" or omit for generally grouped counts. | `arrived` |

---

## 📤 Response Structure

The API returns a consolidated summary map with counts for each status.

```json
{
  "success": true,
  "clinic_id": "e93b535c-e349-4a65-b450-0e785c4f3254",
  "date": "2026-02-19",
  "doctor_id": "",
  "status": "",
  "summary": {
    "total": 50,
    "confirmed": 20,
    "arrived": 10,
    "completed": 15,
    "cancelled": 2,
    "no_show": 1,
    "pending": 0
  }
}
```

### 🔑 Key Fields in `summary` object:
- **`total`**: The grand total of appointments matching the filter.
- **`confirmed`**: Confirmed future appointments.
- **`arrived`**: Patients who have arrived at the clinic.
- **`completed`**: Completed consultations.
- **`cancelled`**: Cancelled appointments.
- **`no_show`**: Patients who missed their slot.

---

## 💡 Frontend Integration Logic

### 1. Default Dashboard Load
Call without `doctor_id` or `status` to get the full clinic overview.
```http
GET /appointments/summary?clinic_id={clinicId}&date={today}
```

### 2. Filter by Doctor
When a specific doctor is selected in the dropdown:
```http
GET /appointments/summary?clinic_id={clinicId}&date={today}&doctor_id={selectedDoctorId}
```
The summary breakdown will update to reflect **only** that doctor's appointments.

### 3. Filter by Status (List + Summary)
If you filter the list by "Arrived", you might still want to see the **Full Summary** (to see how many are completed vs arrived).
- Ideally, **do NOT pass `status` to the summary API** even if the list is filtered, so the dashboard cards remain complete.
- Only pass `status` if you specifically want to count *only* arrived appointments (which will result in `total: N, arrived: N, completed: 0`).

---

## 🔄 Dynamic Updates

When an appointment status changes (e.g., **Arrived** -> **Completed**):
1. Call the Update Status API (`PUT /appointments/{id}`).
2. **Re-fetch** this Summary API.
   - `arrived` count will decrease by 1.
   - `completed` count will increase by 1.

This ensures the dashboard numbers always match the real-time database state.
