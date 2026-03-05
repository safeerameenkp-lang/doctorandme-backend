# 🚀 Frontend Integration Guide: Optimized Slot System

This guide details how to integrate the newly optimized **Doctor Session Slots API** into the frontend (Flutter/React). 
The backend now handles **time validation (past slots)**, **booking counts**, and **reschedule logic** efficiently.

---

## 1️⃣ API Endpoint

**GET** `/doctor-session-slots`

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | ✅ Yes | The doctor's ID |
| `date` | Date (YYYY-MM-DD) | ✅ Yes | The selected date |
| `clinic_id` | UUID | ❌ No | Filter by clinic |
| `slot_type` | String | ❌ No | Filter by type (clinic_visit, video_consultation) |
| `appointment_id` | UUID | ❌ No | **(NEW)** For **Reschedule Flow**. Pass the *current* appointment ID. |

---

## 2️⃣ Response Structure (Key Changes)

Each individual slot object now contains **pre-calculated availability flags**. You do **NOT** need to calculate availability in the frontend.

```json
{
  "id": "slot-uuid",
  "slot_start": "10:00",
  "slot_end": "10:15",
  "is_bookable": false,          // ⛔ MASTER FLAG: Use this to enable/disable button
  "status": "blocked",           // "available", "booked", "blocked"
  "display_message": "Time Passed", // "Available", "Fully Booked", "Time Passed"
  "available_count": 0,
  "max_patients": 1
}
```

### 🧠 Logic Handled by Backend
1.  **Past & Current Minute Slots**: Automatically marked `is_bookable: false`, `status: "blocked"`, `display_message: "Time Passed"`. (Strict Rule: You can only book future slots).
2.  **Full Slots**: Automatically marked `is_bookable: false`, `status: "booked"`, `display_message: "Fully Booked"`.
3.  **Reschedule**: If `appointment_id` is passed, the backend **excludes** that appointment from the count, attempting to make its own slot appear "Available" again (so the user can keep it or move).

---

## 3️⃣ Flutter Integration Example

### A. Update Data Model (`IndividualSlot`)

Ensure your model parses the new fields.

```dart
class IndividualSlot {
  final String id;
  final String slotStart;
  final String slotEnd;
  final bool isBookable;        // ✅ NEW
  final String status;          // ✅ NEW
  final String displayMessage;  // ✅ NEW
  final int availableCount;

  IndividualSlot({
    required this.id,
    required this.slotStart,
    required this.slotEnd,
    required this.isBookable,
    required this.status,
    required this.displayMessage,
    required this.availableCount,
  });

  factory IndividualSlot.fromJson(Map<String, dynamic> json) {
    return IndividualSlot(
      id: json['id'],
      slotStart: json['slot_start'],
      slotEnd: json['slot_end'],
      isBookable: json['is_bookable'] ?? false,
      status: json['status'] ?? 'unknown',
      displayMessage: json['display_message'] ?? '',
      availableCount: json['available_count'] ?? 0,
    );
  }
}
```

### B. Service Method (`SlotService`)

Add `appointmentId` parameter for rescheduling.

```dart
Future<List<TimeSlot>> getDoctorSessionSlots({
  required String doctorId,
  required String date,
  String? clinicId,
  String? appointmentId, // 🆕 Cancellation/Reschedule support
}) async {
  final uri = Uri.parse('$baseUrl/doctor-session-slots').replace(queryParameters: {
    'doctor_id': doctorId,
    'date': date,
    if (clinicId != null) 'clinic_id': clinicId,
    if (appointmentId != null) 'appointment_id': appointmentId, // 🆕 Pass it here
  });

  final response = await http.get(uri, headers: headers);
  // ... parse response ...
}
```

### C. UI Widget (`SlotItem`)

Use `is_bookable` to control interactivity and styling.

```dart
Widget _buildSlotItem(IndividualSlot slot) {
  // 🎨 Color Logic
  Color bgColor;
  Color textColor;
  
  if (!slot.isBookable) {
    if (slot.status == 'blocked') {
      // Time Passed / Blocked
      bgColor = Colors.grey.shade300;
      textColor = Colors.grey.shade600;
    } else {
      // Fully Booked
      bgColor = Colors.red.shade100;
      textColor = Colors.red.shade800;
    }
  } else {
    // Available
    bgColor = Colors.green.shade50;
    textColor = Colors.green.shade800;
  }

  return GestureDetector(
    onTap: slot.isBookable 
      ? () => onSlotSelected(slot) 
      : null, // ⛔ Disable interaction if not bookable
    child: Container(
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: slot.isBookable ? Colors.green : Colors.transparent),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            "${slot.slotStart} - ${slot.slotEnd}",
            style: TextStyle(color: textColor, fontWeight: FontWeight.bold),
          ),
          Text(
            slot.displayMessage, // 📝 Show "Time Passed" / "Available"
            style: TextStyle(fontSize: 10, color: textColor),
          ),
        ],
      ),
    ),
  );
}
```

## 4️⃣ Server-Side Validation Notice

⚠️ **Security Update**: Even if a user bypasses the UI checks, the backend `CreateAppointment` API performs a **Time Check**.
If a user tries to book a past slot, the API returns **400 Bad Request**.

**Error Response:**
```json
{
  "error": "Slot time has passed",
  "message": "You cannot book a past time slot"
}
```
Ensure your frontend handles this 400 error gracefully (e.g., show a toast "This slot just expired").
