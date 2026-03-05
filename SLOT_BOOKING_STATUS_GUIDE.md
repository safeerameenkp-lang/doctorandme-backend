# Slot Booking Status & UI Color Guide 🎨

## 🎯 Overview

This guide explains how slot booking status works and how the UI should display slots based on their availability.

---

## 📊 API Response Structure

### List Slots API
**Endpoint:** `GET /api/organizations/doctor-session-slots`

**Response:**
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "date": "2025-10-18",
  "slot_type": "offline",
  "slots": [
    {
      "id": "time-slot-uuid",
      "date": "2025-10-18",
      "day_of_week": 5,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        {
          "id": "session-uuid",
          "session_name": "Morning",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 20,
          "generated_slots": 36,
          "available_slots": 30,
          "booked_slots": 6,
          "slots": [
            {
              "id": "individual-slot-1",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": false,
              "is_bookable": true,          // ✅ Can book
              "status": "available",
              "display_message": "Available"
            },
            {
              "id": "individual-slot-2",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "is_booked": true,
              "is_bookable": false,         // ⛔ Cannot book
              "booked_patient_id": "patient-uuid",
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

---

## 🔐 Slot Status Fields

| Field | Type | Description | Use in UI |
|-------|------|-------------|-----------|
| `is_booked` | boolean | True if slot has appointment | Check for booking status |
| `is_bookable` | boolean | False if already booked | **Use to enable/disable button** |
| `status` | string | "available", "booked", "blocked" | Display status badge |
| `display_message` | string | "Available" or "Booked" | **Show to user** |
| `booked_patient_id` | string? | Patient who booked (if booked) | For admin info |
| `booked_appointment_id` | string? | Appointment ID (if booked) | For linking |

---

## 🎨 UI Color Coding

### Color Scheme

| Status | Color | Background | Border | Text | Clickable |
|--------|-------|------------|--------|------|-----------|
| **Available** | 🟢 Green | `#E8F5E9` | `#4CAF50` | `#1B5E20` | ✅ Yes |
| **Booked** | 🔴 Red | `#FFEBEE` | `#F44336` | `#B71C1C` | ❌ No |
| **Selected** | 🔵 Blue | `#E3F2FD` | `#2196F3` | `#0D47A1` | ✅ Yes |

---

## 📱 Flutter UI Implementation

### Slot Card Widget

```dart
class SlotCard extends StatelessWidget {
  final IndividualSlot slot;
  final bool isSelected;
  final VoidCallback? onTap;
  
  const SlotCard({
    required this.slot,
    this.isSelected = false,
    this.onTap,
  });
  
  @override
  Widget build(BuildContext context) {
    // Determine colors based on status
    Color backgroundColor;
    Color borderColor;
    Color textColor;
    bool isClickable = slot.isBookable && !isSelected;
    
    if (slot.isBooked || !slot.isBookable) {
      // 🔴 RED - Booked slots (not clickable)
      backgroundColor = Color(0xFFFFEBEE);
      borderColor = Color(0xFFF44336);
      textColor = Color(0xFFB71C1C);
    } else if (isSelected) {
      // 🔵 BLUE - Selected slot
      backgroundColor = Color(0xFFE3F2FD);
      borderColor = Color(0xFF2196F3);
      textColor = Color(0xFF0D47A1);
    } else {
      // 🟢 GREEN - Available slots
      backgroundColor = Color(0xFFE8F5E9);
      borderColor = Color(0xFF4CAF50);
      textColor = Color(0xFF1B5E20);
    }
    
    return GestureDetector(
      onTap: isClickable ? onTap : null,
      child: Container(
        padding: EdgeInsets.symmetric(vertical: 12, horizontal: 16),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border.all(color: borderColor, width: 2),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Column(
          children: [
            // Time display
            Text(
              '${slot.slotStart} - ${slot.slotEnd}',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: textColor,
              ),
            ),
            SizedBox(height: 4),
            
            // Status badge
            Container(
              padding: EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: borderColor.withOpacity(0.2),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                slot.displayMessage, // "Available" or "Booked"
                style: TextStyle(
                  fontSize: 12,
                  color: textColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
            
            // Show icon if booked
            if (slot.isBooked) ...[
              SizedBox(height: 4),
              Icon(
                Icons.block,
                color: borderColor,
                size: 16,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
```

---

### Slot Model

```dart
class IndividualSlot {
  final String id;
  final String slotStart;
  final String slotEnd;
  final bool isBooked;
  final bool isBookable;           // ⭐ Key field
  final String? bookedPatientId;
  final String? bookedAppointmentId;
  final String status;
  final String displayMessage;     // ⭐ Key field
  
  IndividualSlot({
    required this.id,
    required this.slotStart,
    required this.slotEnd,
    required this.isBooked,
    required this.isBookable,
    this.bookedPatientId,
    this.bookedAppointmentId,
    required this.status,
    required this.displayMessage,
  });
  
  factory IndividualSlot.fromJson(Map<String, dynamic> json) {
    return IndividualSlot(
      id: json['id'],
      slotStart: json['slot_start'],
      slotEnd: json['slot_end'],
      isBooked: json['is_booked'] ?? false,
      isBookable: json['is_bookable'] ?? true,
      bookedPatientId: json['booked_patient_id'],
      bookedAppointmentId: json['booked_appointment_id'],
      status: json['status'] ?? 'available',
      displayMessage: json['display_message'] ?? 'Available',
    );
  }
}
```

---

### Usage in Slot List

```dart
class SlotListView extends StatefulWidget {
  final List<IndividualSlot> slots;
  
  @override
  _SlotListViewState createState() => _SlotListViewState();
}

class _SlotListViewState extends State<SlotListView> {
  String? selectedSlotId;
  
  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
        childAspectRatio: 2,
      ),
      itemCount: widget.slots.length,
      itemBuilder: (context, index) {
        final slot = widget.slots[index];
        
        return SlotCard(
          slot: slot,
          isSelected: selectedSlotId == slot.id,
          onTap: () {
            // ⛔ Only allow booking if slot is bookable
            if (slot.isBookable && !slot.isBooked) {
              setState(() {
                selectedSlotId = slot.id;
              });
            } else {
              // Show error message
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('This slot is already booked by another patient'),
                  backgroundColor: Colors.red,
                ),
              );
            }
          },
        );
      },
    );
  }
}
```

---

## ⛔ Booking Validation

### Client-Side (Flutter)

```dart
Future<void> bookAppointment(String slotId) async {
  // Get the slot
  final slot = slots.firstWhere((s) => s.id == slotId);
  
  // ⚠️ VALIDATE BEFORE API CALL
  if (!slot.isBookable || slot.isBooked) {
    throw Exception('This slot is not available for booking');
  }
  
  // Proceed with API call
  final response = await http.post(
    Uri.parse('/api/appointments/simple'),
    body: jsonEncode({
      'individual_slot_id': slotId,
      // ... other fields
    }),
  );
  
  // Handle response
  if (response.statusCode == 409) {
    // Slot was booked by someone else between list and booking
    throw Exception('Slot just got booked by another patient');
  }
}
```

---

### Server-Side (API Validation)

```go
// ⛔ Block booking if slot is already booked
if isBooked || slotStatus != "available" {
    c.JSON(http.StatusConflict, gin.H{
        "error":   "Slot already booked",
        "message": "This slot is already booked by another patient.",
    })
    return
}
```

**HTTP 409 Conflict** is returned if slot is already booked.

---

## 🔄 Double Booking Prevention

### How It Works

1. **List Slots** → API returns `is_bookable: false` for booked slots
2. **UI Renders** → Booked slots shown in RED, button disabled
3. **User Clicks** → Only green (available) slots are clickable
4. **API Validation** → Backend verifies slot is still available
5. **Database Update** → Slot marked as booked, locked for others

### Race Condition Handling

```
User A selects slot → API validates → Still available → Book ✅
User B selects SAME slot → API validates → Now booked → Reject ❌
```

The API always validates slot availability at booking time!

---

## 📋 Quick Checklist

### UI Developer Checklist

- [ ] Use `is_bookable` to enable/disable slot buttons
- [ ] Show booked slots in RED color (`#FFEBEE` background)
- [ ] Show available slots in GREEN color (`#E8F5E9` background)
- [ ] Display `display_message` ("Available" or "Booked")
- [ ] Disable click on slots where `is_bookable = false`
- [ ] Show error snackbar if user tries to book unavailable slot
- [ ] Handle HTTP 409 error for race condition cases
- [ ] Refresh slot list after successful booking

---

## 🎯 Example Response Values

### Available Slot ✅
```json
{
  "id": "slot-uuid-1",
  "slot_start": "09:00",
  "slot_end": "09:05",
  "is_booked": false,
  "is_bookable": true,        // ✅ Green - Clickable
  "status": "available",
  "display_message": "Available"
}
```

### Booked Slot ⛔
```json
{
  "id": "slot-uuid-2",
  "slot_start": "09:05",
  "slot_end": "09:10",
  "is_booked": true,
  "is_bookable": false,       // ⛔ Red - Not clickable
  "booked_patient_id": "patient-uuid",
  "booked_appointment_id": "appointment-uuid",
  "status": "booked",
  "display_message": "Booked"
}
```

---

## ✅ Summary

| Slot Status | `is_bookable` | UI Color | Clickable | Action |
|-------------|---------------|----------|-----------|--------|
| Available | `true` | 🟢 Green | ✅ Yes | Allow booking |
| Booked | `false` | 🔴 Red | ❌ No | Show error |
| Selected | `true` | 🔵 Blue | ✅ Yes | Proceed to book |

---

**Key Points:**
1. ✅ Use `is_bookable` field to control button state
2. 🔴 Show booked slots in RED - never allow clicking
3. ⛔ API validates again on booking to prevent race conditions
4. 💾 Once booked, slot locked for all other patients

**Status:** ✅ Double booking prevention active! 🎉

