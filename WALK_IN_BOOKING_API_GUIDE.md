# Walk-in Booking API Guide (Simplified Mode)

This guide describes how to implement Walk-in bookings in the frontend. Walk-in bookings allow clinics to book appointments even when all predefined session slots for a doctor are full or have already passed.

## Overview (New Permissive Logic)

Walk-in bookings are enabled for a specific date if **sessions are defined**. Walk-in is now a parallel booking option and is **NOT blocked** by slot availability.

### Business Cases:

| Case | Scenario | Walk-in Option? |
| :--- | :--- | :--- |
| **1** | Morning session is full, but Afternoon has slots. | **SHOW** (Walk-in allowed at user discretion). |
| **2** | Morning time has passed, but Afternoon has slots. | **SHOW** (Walk-in allowed at user discretion). |
| **3** | All sessions (Morning & Afternoon) are fully booked. | **SHOW** (Walk-in allowed). |
| **4** | Current time has passed all sessions for today. | **SHOW** (Walk-in allowed). |

The Backend **DOES NOT** enforce slot exhaustion. The frontend can choose to show/hide the button, but the API will accept walk-ins even if slots are available.

## 1. Checking Walk-in Availability

The Slot Listing API now returns a direct indication of whether a walk-in is allowed for the day.

### GET /doctor-session-slots
- **Query Params**: `doctor_id`, `clinic_id`, `date`, `slot_type`
- **New Response Fields**:
  - `walkin_available`: `boolean` (True if walk-in can be booked)
  - `walkin_reason`: `string` (Explanation of why walk-in is available or disabled)

#### Example Response (Walk-in Allowed)
```json
{
  "doctor_id": "...",
  "date": "2024-03-20",
  "walkin_available": true,
  "walkin_reason": "All slots are fully booked or time has passed for this date.",
  "slots": [...]
}
```

#### Example Response (Walk-in Blocked)
```json
{
  "doctor_id": "...",
  "date": "2024-03-20",
  "walkin_available": false,
  "walkin_reason": "15 slots are still available. Please select a specific time slot.",
  "slots": [...]
}
```

## 2. Request Parameters (Booking)

When creating a walk-in appointment, include the `booking_mode` field and leave `slot_id` as null.

| Field | Type | Description |
| :--- | :--- | :--- |
| `booking_mode` | `string` | **Optional**. Values: `"slot"` (default) or `"walk_in"`. |
| `individual_slot_id` | `uuid` | **Required** for `"slot"`. **Must be null** for `"walk_in"`. |

### Example Request (`POST /appointments/simple`)

```json
{
  "clinic_patient_id": "...",
  "doctor_id": "...",
  "clinic_id": "...",
  "appointment_date": "2024-03-20",
  "appointment_time": "2024-03-20 18:30:00",
  "booking_mode": "walk_in"
}
```

## 3. Validation & Error Handling

### Day-wide Availability Check
If you try to book a walk-in when there are still available future slots anywhere in the day for that doctor, the API will return a `400 Bad Request`.

**Response:**
```json
{
  "error": "Slots available",
  "message": "There are still 5 available slots for this date. Please select a specific time slot.",
  "available_slots": 5
}
```

## 4. Implementation Strategy for Frontend

1.  **Call List API**: Call `GET /doctor-session-slots` for the selected date.
2.  **Evaluate Mode**: 
    - If `walkin_available` is `true`, show a single "Book as Walk-in" button/panel.
    - If `walkin_available` is `false`, show the normal slot grid and hide the walk-in option (or show the `walkin_reason` on hover).
3.  **Submission**: On submission, set `booking_mode: "walk_in"` and ensure `individual_slot_id` is null.
4.  **Display**: In the appointment lists, show a "Walk-in" badge for appointments where `booking_mode == "walk_in"`.
