# Follow-Up Slot Type Filter - Complete Guide 🔄

## 🎯 Overview

The `ListDoctorSessionSlots` API now supports **follow-up appointment types** that automatically filter slots based on consultation mode.

---

## 📋 Slot Type Options

| Slot Type | Description | Shows |
|-----------|-------------|-------|
| `offline` | Regular offline slots | Only offline slots |
| `online` | Regular online slots | Only online slots |
| `follow-up-via-offline` | Follow-up offline appointment | Only offline slots |
| `follow-up-via-online` | Follow-up online appointment | Only online slots |

---

## 🔄 How It Works

### Mapping Logic:

```go
switch slotType {
case "offline":
    actualSlotType = "offline"
case "online":
    actualSlotType = "online"
case "follow-up-via-offline":
    actualSlotType = "offline"  // ✅ Maps to offline
case "follow-up-via-online":
    actualSlotType = "online"   // ✅ Maps to online
}
```

**Result:** Follow-up types automatically show the correct slots!

---

## 🌐 API Usage

### Endpoint:
```
GET /api/organizations/doctor-session-slots
```

### Query Parameters:

| Parameter | Type | Required | Values |
|-----------|------|----------|--------|
| `doctor_id` | UUID | ✅ Yes | Doctor's UUID |
| `clinic_id` | UUID | ❌ No | Clinic's UUID |
| `date` | Date | ❌ No | YYYY-MM-DD format |
| `slot_type` | String | ❌ No | `offline`, `online`, `follow-up-via-offline`, `follow-up-via-online` |
| `appointment_id` | UUID | ❌ No | For reschedule mode |

---

## 📝 Examples

### Example 1: Regular Offline Appointment

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20&
  slot_type=offline

Authorization: Bearer {token}
```

**Result:** ✅ Shows only **offline** slots

---

### Example 2: Regular Online Appointment

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20&
  slot_type=online

Authorization: Bearer {token}
```

**Result:** ✅ Shows only **online** slots

---

### Example 3: Follow-Up Offline Appointment ⭐

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20&
  slot_type=follow-up-via-offline

Authorization: Bearer {token}
```

**Result:** ✅ Shows only **offline** slots (follow-up mode)

---

### Example 4: Follow-Up Online Appointment ⭐

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20&
  slot_type=follow-up-via-online

Authorization: Bearer {token}
```

**Result:** ✅ Shows only **online** slots (follow-up mode)

---

### Example 5: All Slots (No Filter)

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20

Authorization: Bearer {token}
```

**Result:** ✅ Shows **both offline and online** slots

---

## 📱 Flutter UI Integration

### Dropdown for Appointment Type

```dart
class AppointmentBookingPage extends StatefulWidget {
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  String? selectedSlotType;
  
  final List<Map<String, String>> appointmentTypes = [
    {'value': 'offline', 'label': '🏥 Regular Offline Appointment'},
    {'value': 'online', 'label': '💻 Regular Online Appointment'},
    {'value': 'follow-up-via-offline', 'label': '🔄 Follow-Up (Offline)'},
    {'value': 'follow-up-via-online', 'label': '🔄 Follow-Up (Online)'},
  ];
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Book Appointment')),
      body: Column(
        children: [
          // Appointment Type Dropdown
          DropdownButtonFormField<String>(
            decoration: InputDecoration(
              labelText: 'Appointment Type',
              border: OutlineInputBorder(),
            ),
            value: selectedSlotType,
            items: appointmentTypes.map((type) {
              return DropdownMenuItem<String>(
                value: type['value'],
                child: Text(type['label']!),
              );
            }).toList(),
            onChanged: (value) {
              setState(() {
                selectedSlotType = value;
                // Re-fetch slots with the new filter
                _loadSlots();
              });
            },
          ),
          
          SizedBox(height: 20),
          
          // Slots List
          Expanded(
            child: _buildSlotsList(),
          ),
        ],
      ),
    );
  }
  
  Future<void> _loadSlots() async {
    if (selectedSlotType == null) {
      // Show error or instruction
      return;
    }
    
    final url = Uri.parse(
      '$baseUrl/organizations/doctor-session-slots?'
      'doctor_id=$doctorId&'
      'clinic_id=$clinicId&'
      'date=$selectedDate&'
      'slot_type=$selectedSlotType'  // ✅ Pass selected type
    );
    
    final response = await http.get(
      url,
      headers: {'Authorization': 'Bearer $token'},
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      setState(() {
        // Update slots
        slots = parseSlots(data['time_slots']);
      });
    }
  }
  
  Widget _buildSlotsList() {
    if (selectedSlotType == null) {
      return Center(
        child: Text('Please select appointment type'),
      );
    }
    
    return ListView.builder(
      itemCount: slots.length,
      itemBuilder: (context, index) {
        final slot = slots[index];
        return ListTile(
          title: Text(slot.slotStart),
          subtitle: Text(slot.isBookable 
            ? '${slot.availableCount} spots available' 
            : 'Fully booked'),
          trailing: ElevatedButton(
            onPressed: slot.isBookable 
              ? () => _bookSlot(slot) 
              : null,
            child: Text('Book'),
          ),
        );
      },
    );
  }
}
```

---

## 🎨 UI Flow Examples

### Scenario 1: Patient Selects Follow-Up Offline

```
User Flow:
1. Opens appointment booking
2. Selects "🔄 Follow-Up (Offline)" from dropdown
   ↓
API Call:
  slot_type=follow-up-via-offline
   ↓
Backend Mapping:
  actualSlotType = "offline"
   ↓
Result:
  Shows only offline slots ✅
   ↓
User:
  Selects an available offline slot
  Books appointment
```

---

### Scenario 2: Patient Selects Follow-Up Online

```
User Flow:
1. Opens appointment booking
2. Selects "🔄 Follow-Up (Online)" from dropdown
   ↓
API Call:
  slot_type=follow-up-via-online
   ↓
Backend Mapping:
  actualSlotType = "online"
   ↓
Result:
  Shows only online slots ✅
   ↓
User:
  Selects an available online slot
  Books video/online appointment
```

---

## 📊 Response Structure

Same as regular `ListDoctorSessionSlots` response, but filtered:

```json
{
  "time_slots": [
    {
      "id": "slot-uuid",
      "doctor_id": "doctor-uuid",
      "clinic_id": "clinic-uuid",
      "date": "2025-10-20",
      "slot_type": "offline",  // ✅ Only shows requested type
      "is_available": true,
      "sessions": [
        {
          "id": "session-uuid",
          "session_name": "Morning Session",
          "start_time": "09:00:00",
          "end_time": "12:00:00",
          "slots": [
            {
              "id": "individual-slot-uuid",
              "slot_start": "09:00:00",
              "slot_end": "09:05:00",
              "is_booked": false,
              "is_bookable": true,
              "max_patients": 1,
              "available_count": 1,
              "booked_count": 0,
              "status": "available",
              "display_message": "Available"
            }
            // ... more slots
          ]
        }
      ]
    }
  ]
}
```

---

## ❌ Error Cases

### Invalid Slot Type

**Request:**
```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  slot_type=follow-up-via-phone  ❌ Invalid
```

**Response:**
```json
{
  "error": "Invalid slot_type. Must be 'offline', 'online', 'follow-up-via-offline', or 'follow-up-via-online'"
}
```

---

## 🔍 Validation Rules

| Rule | Status | Description |
|------|--------|-------------|
| `doctor_id` required | ✅ | Must be valid UUID |
| `slot_type` optional | ✅ | If provided, must be one of 4 values |
| `date` validation | ✅ | Cannot be in the past |
| `clinic_id` optional | ✅ | Filters by clinic |

---

## 🧪 Testing

### Test 1: Follow-Up Offline Shows Only Offline Slots

```bash
# Create both offline and online slots for a doctor
POST /api/organizations/doctor-session-slots
{
  "slot_type": "offline",
  "date": "2025-10-20",
  ...
}

POST /api/organizations/doctor-session-slots
{
  "slot_type": "online",
  "date": "2025-10-20",
  ...
}

# Query with follow-up-via-offline
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  date=2025-10-20&
  slot_type=follow-up-via-offline

# ✅ Should show ONLY offline slots
```

---

### Test 2: Follow-Up Online Shows Only Online Slots

```bash
# Query with follow-up-via-online
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  date=2025-10-20&
  slot_type=follow-up-via-online

# ✅ Should show ONLY online slots
```

---

### Test 3: No Filter Shows All Slots

```bash
# Query without slot_type
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  date=2025-10-20

# ✅ Should show BOTH offline and online slots
```

---

## 📋 Summary Table

| User Selects | `slot_type` Value | Backend Filters | Slots Shown |
|--------------|------------------|----------------|-------------|
| Regular Offline | `offline` | `slot_type = 'offline'` | Offline only |
| Regular Online | `online` | `slot_type = 'online'` | Online only |
| Follow-Up Offline | `follow-up-via-offline` | `slot_type = 'offline'` | Offline only |
| Follow-Up Online | `follow-up-via-online` | `slot_type = 'online'` | Online only |
| No selection | (empty) | No filter | All slots |

---

## 🎯 Key Benefits

✅ **Clearer Intent:** "Follow-up" clearly indicates appointment type in UI  
✅ **Correct Filtering:** Automatically shows only relevant slots  
✅ **Backward Compatible:** Existing `offline`/`online` still work  
✅ **No Database Changes:** Pure API-level mapping  
✅ **Easy UI Integration:** Simple dropdown selection  

---

## 📝 Complete Example Flow

### 1️⃣ User Opens Appointment Booking

```dart
// UI shows dropdown
appointmentTypes: [
  'Regular Offline',
  'Regular Online',
  'Follow-Up (Offline)',
  'Follow-Up (Online)',
]
```

---

### 2️⃣ User Selects "Follow-Up (Offline)"

```dart
selectedSlotType = 'follow-up-via-offline';
```

---

### 3️⃣ API Call

```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&
  date=2025-10-20&
  slot_type=follow-up-via-offline
```

---

### 4️⃣ Backend Mapping

```go
actualSlotType = "offline"  // Mapped automatically
```

---

### 5️⃣ Database Query

```sql
SELECT * FROM doctor_time_slots
WHERE doctor_id = $1
  AND slot_type = 'offline'  -- ✅ Filtered
  AND specific_date = $2
```

---

### 6️⃣ Response

```json
{
  "time_slots": [
    // ✅ Only offline slots returned
  ]
}
```

---

### 7️⃣ UI Display

```
Available Offline Slots:
  ✅ 09:00 - 09:05 (Available)
  ✅ 09:05 - 09:10 (Available)
  ✅ 09:10 - 09:15 (Booked)
  ...
```

---

## ✅ Implementation Status

| Component | Status | Notes |
|-----------|--------|-------|
| API Parameter | ✅ Done | Accepts 4 slot_type values |
| Validation | ✅ Done | Validates all 4 types |
| Mapping Logic | ✅ Done | Maps follow-up to offline/online |
| Database Query | ✅ Done | Uses mapped value |
| Error Handling | ✅ Done | Clear error messages |
| Documentation | ✅ Done | This guide |

---

**Status:** ✅ **Follow-up slot type filtering fully implemented!**

**Use Cases:** Regular appointments + Follow-up appointments  
**Filtering:** Automatic based on selected type  
**Ready for:** Flutter UI integration! 🎉

