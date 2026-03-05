# 🎨 Reschedule UI Implementation Guide (Right Drawer)

This guide details how to implement the **Reschedule Flow** using a Right Drawer component, ensuring a real-time, production-grade experience.

---

## 📱 UI Flow Overview

1.  **Trigger**: User clicks "Reschedule" on an appointment card.
2.  **Action**: Open **Right Drawer**.
3.  **Drawer Content**:
    -   **Step 1**: Date Picker (Default to current appointment date or today).
    -   **Step 2**: Consultation Type Selector (Optional, default to current).
    -   **Step 3**: **Live Slot List** (Fetched based on Date + Doctor).
    -   **Step 4**: "Confirm Reschedule" Button (Disabled until slot selected).

---

## 🔄 Real-Time Data Handling

To ensure production-grade accuracy (minimizing "Slot Taken" errors):

### 1. Fetch Slots on Date Change
When the Drawer opens or Date is changed, immediately fetch fresh availability:
```javascript
// GET /organization/doctor-session-slots?doctor_id=...&date=YYYY-MM-DD
// OR the dedicated available slots endpoint
fetchAvailableSlots(doctorId, selectedDate);
```

### 2. Slot Display Logic
Render the slot list dynamically:
-   **Green / Clickable**: `available_count > 0` AND `status == 'available'`.
-   **Grey / Disabled**: `available_count == 0` OR `status == 'booked'`.
-   **Label**: Show time range (e.g., "10:00 AM - 10:15 AM") and optionally "X spots left".

### 3. Handle "Conflict" Gracefully
If the API returns `409 Conflict` (Slot just got booked by someone else):
-   **Show Alert**: "This slot was just taken. Please select another."
-   **Auto-Refresh**: Immediately re-fetch the slot list to update the UI.
-   **Keep Drawer Open**: Do not close the drawer; let the user pick again.

---

## 🛠️ API Integration Steps

### 1. Prepare Request Payload
Gather data from the UI state:
```javascript
const payload = {
  doctor_id: currentAppointment.doctor_id,
  individual_slot_id: selectedSlot.id,  // ID from the selected slot object
  appointment_date: selectedDate,       // YYYY-MM-DD
  appointment_time: selectedSlot.start_time, // "2026-02-25 10:00:00"
  consultation_type: selectedType,      // "clinic_visit" or "video_consultation"
  reason: reasonInput.value
};
```

### 2. Call Reschedule API
```javascript
POST /api/v1/appointments/simple/{appointmentId}/reschedule
Body: payload
```

### 3. Post-Success Action
-   **Close Drawer**.
-   **Show Toast**: "Rescheduled Successfully".
-   **Refresh Main Dashboard**: The main appointment list MUST be refreshed because the old slot is now free and the appointment details (date/time) have changed.
    -   Also refresh the **Summary API** (`GET /appointments/summary`) if the date changed (as today's counts might change).

---

## 🎨 Visual States (Production Quality)

-   **Loading State**: When changing date, show a skeleton loader or spinner in the slot area.
-   **Selected State**: Highlight the picked slot with a primary border/background.
-   **Current Slot**: If the user selects the *same day* as the original appointment, highlight their *current* slot differently (e.g., "Current Slot") or hide it if rescheduling typically implies changing it. (Usually better to show it as "Current").

---

## 🚀 "Live" Guidance

Since we don't have WebSockets for slots yet:
-   **Polling (Optional)**: If the user keeps the drawer open for > 1 minute, you *could* re-fetch slots every 30s.
-   **But best practice**: Just fetch on load/date-change. If a conflict happens (rare), handle the 409 error gracefully (as described above). This is the standard "Optimistic UI" approach.
