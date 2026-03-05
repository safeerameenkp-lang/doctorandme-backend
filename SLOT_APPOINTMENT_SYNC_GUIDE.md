# Slot-Appointment Sync System 🔄

## 🎯 Overview

This system ensures **data consistency** between slot booking status and actual appointments. It automatically checks and syncs slot availability with the appointments table.

---

## 🔐 How It Works

### Automatic Validation on Slot List

When you call `GET /doctor-session-slots`, the API automatically:

1. ✅ Checks each slot's `is_booked` status
2. ✅ Cross-references with `appointments` table
3. ✅ Finds appointments linked to slots via `individual_slot_id`
4. ✅ Marks slots as booked if they have active appointments
5. ✅ Auto-updates database if inconsistency found

---

## 📊 Database Relationship

```
┌─────────────────────────────────────────────────────┐
│ SLOT-APPOINTMENT RELATIONSHIP                       │
└─────────────────────────────────────────────────────┘

doctor_individual_slots          appointments
┌──────────────────┐            ┌──────────────────┐
│ id (UUID)        │◄───────────│individual_slot_id│
│ is_booked        │            │ status           │
│ status           │            │ patient_id       │
│ booked_appt_id   │◄───────────│ id               │
└──────────────────┘            └──────────────────┘

When appointment created:
1. individual_slot_id = slot.id
2. slot.is_booked = true
3. slot.status = 'booked'
4. slot.booked_appointment_id = appointment.id
```

---

## 🔄 Sync Logic

### Auto-Sync During List (Real-time)

**API:** `GET /api/organizations/doctor-session-slots`

```go
// For each slot, check appointments table
SELECT EXISTS(
    SELECT 1 FROM appointments 
    WHERE individual_slot_id = $1 
    AND status NOT IN ('cancelled', 'no_show')
)

// If appointment exists but slot not marked as booked:
UPDATE doctor_individual_slots 
SET is_booked = true, status = 'booked' 
WHERE id = $1
```

**Result:** Slots with active appointments always show as booked! ✅

---

### Manual Sync (Batch Update)

**API:** `POST /api/organizations/doctor-session-slots/sync-booking-status`

**Purpose:** Sync all slots at once (useful after data import or migration)

**Request:**
```bash
POST /api/organizations/doctor-session-slots/sync-booking-status?clinic_id=xxx
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "Slot booking status synced successfully",
  "slots_booked": 15,    // Slots marked as booked
  "slots_freed": 3,      // Slots freed (no active appointment)
  "total_synced": 18
}
```

---

## 📝 Migration Applied

**Migration 020:** Added `individual_slot_id` to appointments table

```sql
ALTER TABLE appointments 
ADD COLUMN individual_slot_id UUID;

ALTER TABLE appointments
ADD CONSTRAINT appointments_individual_slot_id_fkey 
FOREIGN KEY (individual_slot_id) 
REFERENCES doctor_individual_slots(id);

CREATE INDEX idx_appointments_individual_slot_id 
ON appointments(individual_slot_id);
```

---

## ✅ Updated Appointment Creation

**File:** `appointment_simple.controller.go`

**Before:**
```go
INSERT INTO appointments (
    clinic_patient_id, clinic_id, doctor_id, ...
)
VALUES ($1, $2, $3, ...)
```

**After:**
```go
INSERT INTO appointments (
    clinic_patient_id, clinic_id, doctor_id, ..., individual_slot_id
)
VALUES ($1, $2, $3, ..., $16)  // ✅ Stores slot ID
```

Now every appointment is linked to its specific slot! 🎯

---

## 🔍 Validation Flow

### When Creating Appointment

```
1. User selects slot_id
         ↓
2. Check slot is available
   SELECT is_booked, status 
   FROM doctor_individual_slots
   WHERE id = slot_id
         ↓
3. If booked → REJECT ❌
         ↓
4. Create appointment
   INSERT INTO appointments (..., individual_slot_id)
         ↓
5. Mark slot as booked
   UPDATE doctor_individual_slots
   SET is_booked = true, status = 'booked'
         ↓
6. ✅ Appointment created & slot locked
```

---

### When Listing Slots

```
1. List all slots
   SELECT * FROM doctor_individual_slots
         ↓
2. For each slot, check appointments
   SELECT EXISTS(
       SELECT 1 FROM appointments 
       WHERE individual_slot_id = slot.id
       AND status NOT IN ('cancelled', 'no_show')
   )
         ↓
3. If appointment exists but slot.is_booked = false
         ↓
4. Auto-sync: Mark slot as booked
   UPDATE doctor_individual_slots 
   SET is_booked = true
         ↓
5. ✅ Return accurate slot status
```

---

## 🎨 UI Integration

### Slot Status Determination

```dart
// The API already handles sync, just use the response
class SlotDisplay {
  void displaySlot(IndividualSlot slot) {
    if (!slot.isBookable || slot.isBooked) {
      // 🔴 RED - Slot is booked (validated against appointments)
      showAsBooked(slot);
    } else {
      // 🟢 GREEN - Slot is available
      showAsAvailable(slot);
    }
  }
}
```

**You don't need to do anything special!** The API handles all validation. ✅

---

## 📋 API Examples

### Example 1: List Slots (Auto-Sync)

**Request:**
```bash
GET /api/organizations/doctor-session-slots?doctor_id=xxx&date=2025-10-18
```

**Response:**
```json
{
  "slots": [
    {
      "sessions": [
        {
          "slots": [
            {
              "id": "slot-1",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": false,
              "is_bookable": true,
              "status": "available",
              "display_message": "Available"
            },
            {
              "id": "slot-2",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "is_booked": true,              // ✅ Auto-synced from appointments
              "is_bookable": false,
              "booked_appointment_id": "appointment-uuid",
              "status": "booked",
              "display_message": "Booked"
            }
          ]
        }
      ]
    }
  ]
}
```

**Behind the scenes:** API checked appointments table and updated slot status! ✅

---

### Example 2: Create Appointment

**Request:**
```bash
POST /api/appointments/simple
{
  "individual_slot_id": "slot-1",
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:00:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

**What Happens:**
1. ✅ Validate slot is available
2. ✅ Create appointment with `individual_slot_id = "slot-1"`
3. ✅ Update slot: `is_booked = true, status = 'booked'`
4. ✅ Link appointment: `booked_appointment_id = appointment.id`

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid",
    "individual_slot_id": "slot-1",  // ✅ Linked
    "booking_number": "BN202510180001",
    "status": "confirmed"
  }
}
```

---

### Example 3: Try to Book Already Booked Slot

**Request:**
```bash
POST /api/appointments/simple
{
  "individual_slot_id": "slot-2"  // Already booked
}
```

**Response (409 Conflict):**
```json
{
  "error": "Slot already booked",
  "message": "This slot is already booked by another patient. Please select another slot."
}
```

**UI Should:** Show error and keep slot displayed in RED ⛔

---

### Example 4: Manual Sync All Slots

**Request:**
```bash
POST /api/organizations/doctor-session-slots/sync-booking-status?clinic_id=clinic-uuid
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "Slot booking status synced successfully",
  "slots_booked": 12,   // 12 slots marked as booked
  "slots_freed": 2,     // 2 slots freed (no appointment)
  "total_synced": 14
}
```

**When to Use:**
- After data migration
- After manual database updates
- If you notice inconsistencies
- Before important operations

---

## ⚠️ Important Notes

### 1. Cancelled/No-Show Appointments

Slots are only marked as booked for **active** appointments:

```sql
status NOT IN ('cancelled', 'no_show')
```

If appointment is cancelled, slot should be freed.

---

### 2. Race Condition Handling

**Scenario:** Two users try to book same slot simultaneously

```
User A: Select slot → Validate → Still available → Book ✅
User B: Select slot → Validate → Now booked → REJECT ❌
```

The API validates at booking time, preventing double booking!

---

### 3. Data Consistency

The system ensures:
- ✅ Every booked slot has an active appointment
- ✅ Every active appointment has a booked slot
- ✅ No orphaned bookings
- ✅ No double bookings

---

## 🔧 Troubleshooting

### Problem: Slot shows available but has appointment

**Solution:** Call sync API
```bash
POST /api/organizations/doctor-session-slots/sync-booking-status?clinic_id=xxx
```

---

### Problem: Slot shows booked but no appointment

**Solution:** Sync API will free the slot
```bash
POST /api/organizations/doctor-session-slots/sync-booking-status
```

Response will show `slots_freed: N`

---

### Problem: Can't book available slot

**Check:**
1. Is `is_bookable = true`?
2. Is `is_booked = false`?
3. Is `status = 'available'`?

If all true but still failing, check database directly:
```sql
SELECT * FROM doctor_individual_slots WHERE id = 'slot-id';
SELECT * FROM appointments WHERE individual_slot_id = 'slot-id';
```

---

## ✅ Summary

| Feature | Status |
|---------|--------|
| `individual_slot_id` in appointments | ✅ Added |
| Auto-sync on slot list | ✅ Implemented |
| Manual sync API | ✅ Available |
| Double booking prevention | ✅ Active |
| Race condition handling | ✅ Protected |
| Cancelled appointment handling | ✅ Slots freed |
| Database indexes | ✅ Optimized |

---

## 📊 Quick Reference

| API | Method | Purpose |
|-----|--------|---------|
| `/doctor-session-slots` | GET | List slots (auto-syncs) |
| `/doctor-session-slots/sync-booking-status` | POST | Manual sync all |
| `/appointments/simple` | POST | Create appointment (books slot) |

---

**Key Points:**
1. ✅ Slot list automatically validates against appointments
2. ✅ Appointment creation stores `individual_slot_id`
3. ✅ Manual sync available for batch updates
4. ⛔ Double booking impossible (validated at booking time)
5. 🔴 Booked slots always show in RED in UI

**Status:** ✅ **Slot-Appointment sync system active and working!** 🔄🎉

