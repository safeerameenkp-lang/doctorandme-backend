# Slot Capacity Management System 📊

## 🎯 Overview

This system implements **count-based slot availability** where each 5-minute slot supports **1 patient** (configurable). The system automatically:
- ✅ Decreases `available_count` when booking
- ✅ Marks slot as "Fully Booked" when `available_count = 0`
- ✅ Prevents double booking with race condition protection
- ✅ Auto-syncs with appointments table

---

## 📋 Key Concepts

### Each 5-Minute Slot Structure

```
┌─────────────────────────────────────┐
│ Slot: 09:00 - 09:05                 │
├─────────────────────────────────────┤
│ max_patients: 1                     │  ← Total capacity
│ available_count: 1                  │  ← Available spots
│ booked_count: 0                     │  ← Booked spots
│ is_bookable: true                   │  ← Can book?
│ status: "available"                 │  ← Current status
└─────────────────────────────────────┘
```

**Formula:**  
`booked_count = max_patients - available_count`

---

## 🔄 Booking Flow

### Step-by-Step Process

```
User selects slot 09:00-09:05
         ↓
1. Check available_count > 0
         ↓
2. If YES → Create appointment
         ↓
3. Decrease available_count by 1
   available_count = available_count - 1
         ↓
4. If available_count = 0
   ├─ Set is_booked = true
   ├─ Set status = 'booked'
   └─ Set display_message = "Fully Booked"
         ↓
5. ✅ Slot now shows as RED (unavailable)
```

---

## 📊 Database Schema

### Migration 021: Capacity Tracking

```sql
ALTER TABLE doctor_individual_slots 
ADD COLUMN max_patients INTEGER NOT NULL DEFAULT 1,
ADD COLUMN available_count INTEGER NOT NULL DEFAULT 1;

-- Constraint: available_count must be between 0 and max_patients
ALTER TABLE doctor_individual_slots
ADD CONSTRAINT check_available_count_valid 
CHECK (available_count >= 0 AND available_count <= max_patients);
```

---

## 🔐 Race Condition Prevention

### Atomic Update with WHERE Clause

```sql
UPDATE doctor_individual_slots
SET available_count = available_count - 1,
    is_booked = CASE WHEN available_count - 1 <= 0 THEN true ELSE is_booked END,
    status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END
WHERE id = $1
AND available_count > 0      -- ⚠️ Only update if spots available
AND status = 'available'      -- ⚠️ Only update if still available
```

**How It Prevents Double Booking:**
- If 2 users try to book simultaneously
- Only the first UPDATE will match the WHERE conditions
- Second UPDATE finds `available_count = 0` → No rows affected → Returns error

---

## 📝 API Examples

### Example 1: Check Slot Availability

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
          "session_name": "Morning",
          "slots": [
            {
              "id": "slot-1",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "max_patients": 1,
              "available_count": 1,      // ✅ 1 spot available
              "booked_count": 0,          // 0 booked
              "is_bookable": true,        // ✅ Can book
              "status": "available",
              "display_message": "Available"
            },
            {
              "id": "slot-2",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "max_patients": 1,
              "available_count": 0,      // ⛔ No spots
              "booked_count": 1,          // 1 booked
              "is_bookable": false,       // ⛔ Cannot book
              "status": "booked",
              "display_message": "Fully Booked"
            }
          ]
        }
      ]
    }
  ]
}
```

---

### Example 2: Book Appointment (Success)

**Request:**
```json
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

**Response (201 Created):**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid",
    "booking_number": "BN202510180001",
    "token_number": 1,
    "status": "confirmed"
  }
}
```

**What Happened:**
1. ✅ Checked `available_count = 1` (available)
2. ✅ Created appointment
3. ✅ Decreased `available_count` from 1 to 0
4. ✅ Set `is_booked = true`, `status = 'booked'`
5. ✅ Slot now shows as "Fully Booked"

---

### Example 3: Try to Book Full Slot (Error)

**Request:**
```json
POST /api/appointments/simple
{
  "individual_slot_id": "slot-2"  // Already fully booked
}
```

**Response (409 Conflict):**
```json
{
  "error": "Slot not available",
  "message": "This slot is fully booked. Please select another slot.",
  "details": {
    "max_patients": 1,
    "available_count": 0,
    "booked_count": 1
  }
}
```

**UI Should:** Show error and keep slot in RED ⛔

---

### Example 4: Race Condition Handled

**Scenario:** Two users (A & B) click same slot simultaneously

```
Time  | User A                    | User B
------+---------------------------+---------------------------
T1    | Check: available_count=1  | Check: available_count=1
      | ✅ Available              | ✅ Available
------+---------------------------+---------------------------
T2    | UPDATE available_count-1  | (waiting...)
      | WHERE available_count>0   |
      | ✅ SUCCESS (rows=1)        |
------+---------------------------+---------------------------
T3    | (booking complete)        | UPDATE available_count-1
      |                           | WHERE available_count>0
      |                           | ❌ FAIL (rows=0)
------+---------------------------+---------------------------
T4    | ✅ Appointment created    | ❌ Error: "Slot just got booked"
```

**Result:** Only User A gets the slot. User B gets clear error message.

---

## 🎨 UI Integration

### Slot Display Based on Capacity

```dart
class SlotCard extends StatelessWidget {
  final IndividualSlot slot;
  
  Widget build(BuildContext context) {
    Color bgColor;
    Color borderColor;
    String statusText;
    bool isClickable;
    
    if (slot.availableCount <= 0 || !slot.isBookable) {
      // 🔴 RED - Fully Booked
      bgColor = Color(0xFFFFEBEE);
      borderColor = Color(0xFFF44336);
      statusText = "Fully Booked (${slot.bookedCount}/${slot.maxPatients})";
      isClickable = false;
    } else {
      // 🟢 GREEN - Available
      bgColor = Color(0xFFE8F5E9);
      borderColor = Color(0xFF4CAF50);
      
      if (slot.maxPatients == 1) {
        statusText = "Available";
      } else {
        statusText = "${slot.availableCount}/${slot.maxPatients} Available";
      }
      isClickable = true;
    }
    
    return GestureDetector(
      onTap: isClickable ? () => bookSlot() : null,
      child: Container(
        decoration: BoxDecoration(
          color: bgColor,
          border: Border.all(color: borderColor, width: 2),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Column(
          children: [
            Text('${slot.slotStart} - ${slot.slotEnd}'),
            Text(statusText, style: TextStyle(color: borderColor)),
          ],
        ),
      ),
    );
  }
}
```

---

## 📊 Capacity Formulas

### Basic Formulas

```
booked_count = max_patients - available_count

is_bookable = (available_count > 0 AND status = 'available')

is_full = (available_count = 0)
```

### Example Calculations

| max_patients | available_count | booked_count | is_bookable | Status |
|--------------|-----------------|--------------|-------------|--------|
| 1 | 1 | 0 | ✅ true | 🟢 Available |
| 1 | 0 | 1 | ❌ false | 🔴 Fully Booked |
| 5 | 5 | 0 | ✅ true | 🟢 5/5 Available |
| 5 | 3 | 2 | ✅ true | 🟡 3/5 Available |
| 5 | 1 | 4 | ✅ true | 🟡 1/5 Available |
| 5 | 0 | 5 | ❌ false | 🔴 Fully Booked |

---

## 🔄 Auto-Sync System

### Sync with Appointments Table

The system automatically counts actual appointments and syncs `available_count`:

```sql
-- Count actual appointments
SELECT COUNT(*) 
FROM appointments 
WHERE individual_slot_id = 'slot-id'
AND status NOT IN ('cancelled', 'no_show')

-- Calculate correct available_count
new_available_count = max_patients - actual_appointments_count

-- Update if mismatch
UPDATE doctor_individual_slots
SET available_count = new_available_count,
    is_booked = CASE WHEN new_available_count <= 0 THEN true ELSE false END
WHERE id = 'slot-id'
```

**When Sync Happens:**
- ✅ Every time slots are listed (`GET /doctor-session-slots`)
- ✅ Manual sync API call (`POST /sync-booking-status`)

---

## ⚙️ Configuration

### Change Slot Capacity

To allow multiple patients per 5-minute slot:

**In `CreateDoctorSessionSlots`:**
```go
// Change this line:
maxPatientsPerSlot := 1  // Current: 1 patient per slot

// To:
maxPatientsPerSlot := 3  // New: 3 patients per slot
```

**Result:**
- Each 5-minute slot can now handle 3 patients
- `available_count` starts at 3
- Slot marked as "Fully Booked" after 3 bookings

---

## 🧪 Testing Scenarios

### Test 1: Normal Booking
```
1. Create slot: max_patients=1, available_count=1
2. Book appointment
3. Verify: available_count=0, is_booked=true
4. Try booking again → Should fail with "Fully Booked"
```

### Test 2: Multiple Bookings (if max_patients > 1)
```
1. Create slot: max_patients=3, available_count=3
2. Book 1st appointment → available_count=2
3. Book 2nd appointment → available_count=1
4. Book 3rd appointment → available_count=0, is_booked=true
5. Try 4th booking → Should fail
```

### Test 3: Concurrent Bookings
```
1. Start 2 booking requests simultaneously for same slot
2. Only 1 should succeed
3. Other should get "Slot just got booked" error
4. Verify: Only 1 appointment created
```

### Test 4: Sync After Manual DB Update
```
1. Manually delete an appointment from DB
2. Call GET /doctor-session-slots
3. Verify: available_count increased by 1
4. Verify: is_booked=false if count > 0
```

---

## ✅ Summary

### Key Features

| Feature | Implementation |
|---------|----------------|
| **Slot Duration** | ✅ 5 minutes per slot |
| **Patients per Slot** | ✅ 1 patient (configurable) |
| **Capacity Tracking** | ✅ `available_count` decreases on booking |
| **Auto-disable** | ✅ Slot disabled when `available_count = 0` |
| **Double Booking Prevention** | ✅ Atomic UPDATE with WHERE clause |
| **Race Condition Handling** | ✅ Database-level locking |
| **Auto-sync** | ✅ Syncs with appointments table |
| **UI Status** | ✅ GREEN (available) / RED (fully booked) |

---

### Benefits

**For Patients:**
- ✅ Clear availability indicators
- ✅ No confusion about slot status
- ✅ Immediate feedback if slot becomes unavailable

**For Clinics:**
- ✅ Accurate slot utilization tracking
- ✅ No double bookings
- ✅ Automatic capacity management

**For Developers:**
- ✅ Simple API integration
- ✅ Automatic sync, no manual updates
- ✅ Built-in race condition protection

---

## 📋 Quick Reference

### Slot Status Indicators

| available_count | is_bookable | Display | Color |
|-----------------|-------------|---------|-------|
| > 0 | true | "Available" | 🟢 GREEN |
| = 0 | false | "Fully Booked" | 🔴 RED |

### API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /doctor-session-slots` | List slots with capacity info |
| `POST /appointments/simple` | Book appointment (decreases count) |
| `POST /sync-booking-status` | Manual sync all slots |

---

**Status:** ✅ **Count-based slot system active!**

**Result:** Each 5-minute slot tracks capacity, automatically disables when full, and prevents double booking! 📊🔒

