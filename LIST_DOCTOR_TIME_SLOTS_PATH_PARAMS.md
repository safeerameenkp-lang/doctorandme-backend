# List Doctor Time Slots API - Path Parameter Version

## 🎯 Updated Endpoint
**GET** `/api/organizations/doctor-time-slots/:doctor_id/:clinic_id/:slot_type`

## 📝 Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | **Yes** | The doctor's unique identifier |
| `clinic_id` | UUID | **Yes** | The clinic's unique identifier |
| `slot_type` | String | **Yes** | Type of consultation: `offline` or `online` |

## 🔍 Query Parameters (Optional)

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `date` | String | Optional | Filter by specific date (YYYY-MM-DD format) to calculate availability |

## 🔐 Headers
```
Authorization: Bearer YOUR_AUTH_TOKEN
```

---

## 📌 Example Requests

### **1. Get Offline Slots for Doctor at Specific Clinic**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline
```

### **2. Get Online Slots for Doctor at Specific Clinic**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/online
```

### **3. Get Slots with Date Filter (for Availability Calculation)**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
```

### **4. Get Slots for Monday (using Date)**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-14
```
*Note: October 14, 2024 is a Monday, so only Monday slots will be returned*

---

## ✅ Success Response (200 OK)

```json
{
  "slots": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 3,
      "available_spots": 7,
      "is_available": true,
      "status": "available",
      "notes": "Morning shift - Monday",
      "is_active": true,
      "created_at": "2024-10-13T10:00:00Z",
      "updated_at": "2024-10-13T10:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "booked_patients": 10,
      "available_spots": 0,
      "is_available": false,
      "status": "booking_full",
      "notes": "Afternoon shift - Monday",
      "is_active": true,
      "created_at": "2024-10-13T10:00:00Z",
      "updated_at": "2024-10-13T10:00:00Z"
    }
  ],
  "total": 2
}
```

---

## 📊 Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique slot identifier |
| `doctor_id` | UUID | Doctor's unique identifier |
| `clinic_id` | UUID | Clinic's unique identifier |
| `day_of_week` | Integer | Day number: 0=Sunday, 1=Monday, 2=Tuesday, ..., 6=Saturday |
| `day_name` | String | Day name in English (Sunday, Monday, Tuesday, etc.) |
| `slot_type` | String | Type of consultation: `offline` or `online` |
| `start_time` | String | Start time in HH:MM format (24-hour) |
| `end_time` | String | End time in HH:MM format (24-hour) |
| `max_patients` | Integer | Maximum patients allowed for this slot |
| `booked_patients` | Integer | Number of confirmed/completed appointments for the selected date |
| `available_spots` | Integer | Remaining spots (max_patients - booked_patients) |
| `is_available` | Boolean | `true` if available_spots > 0, `false` otherwise |
| `status` | String | `"available"` or `"booking_full"` |
| `notes` | String | Optional notes about the slot |
| `is_active` | Boolean | Whether the slot is active |
| `created_at` | DateTime | When the slot was created |
| `updated_at` | DateTime | Last update timestamp |

---

## 🚨 Error Responses

### **Missing doctor_id (400)**
```json
{
  "error": "doctor_id is required"
}
```

### **Invalid UUID Format (400)**
```json
{
  "error": "Invalid doctor_id format. Must be a valid UUID"
}
```

### **Invalid Slot Type (400)**
```json
{
  "error": "Invalid slot_type. Must be one of: in-person, online, video, offline"
}
```

### **Invalid Date Format (400)**
```json
{
  "error": "Invalid date format. Use YYYY-MM-DD"
}
```

### **Unauthorized (401)**
```json
{
  "error": "Unauthorized"
}
```

---

## 🔧 cURL Examples

### **Basic Request**
```bash
curl -X GET \
  "http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline" \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN"
```

### **With Date Filter**
```bash
curl -X GET \
  "http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15" \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN"
```

---

## 💡 Use Cases

### **1. Frontend Date Picker - User Selects October 15, 2024**
```
GET /doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
```
**Result:**
- Shows only Tuesday's slots (Oct 15 is a Tuesday)
- Shows real-time availability for that specific date
- `booked_patients` reflects confirmed/completed appointments for Oct 15

### **2. Show All Slots for a Doctor at a Clinic**
```
GET /doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline
```
**Result:**
- Returns all offline slots across all days of the week
- Uses current date for availability calculation

### **3. Toggle Between Offline and Online Consultations**
```
# Offline slots
GET /doctor-time-slots/{doctor_id}/{clinic_id}/offline

# Online slots
GET /doctor-time-slots/{doctor_id}/{clinic_id}/online
```

### **4. Check Next Monday's Availability**
```
GET /doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-21
```
**Result:**
- Shows only Monday slots
- Availability calculated for Oct 21, 2024

---

## 📅 Availability Status Explanation

### **`available`**
- `booked_patients < max_patients`
- There are spots remaining for booking
- Example: 3 booked out of 10 max → 7 available spots

### **`booking_full`**
- `booked_patients >= max_patients`
- No spots remaining for booking
- Example: 10 booked out of 10 max → 0 available spots

### **`unavailable`** (shown in grouped API only)
- Doctor has not created slots for that specific day
- Example: Doctor doesn't work on Sundays

---

## 🎨 Frontend Integration Example

### **React/Next.js Example**
```javascript
const fetchDoctorSlots = async (doctorId, clinicId, slotType, date) => {
  const url = date 
    ? `/api/organizations/doctor-time-slots/${doctorId}/${clinicId}/${slotType}?date=${date}`
    : `/api/organizations/doctor-time-slots/${doctorId}/${clinicId}/${slotType}`;
    
  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const data = await response.json();
  return data.slots;
};

// Usage
const slots = await fetchDoctorSlots(
  '3fd28e6d-7f9a-4dde-8172-d14a74a9b02d',
  '7a6c1211-c029-4923-a1a6-fe3dfe48bdf2',
  'offline',
  '2024-10-15'
);
```

### **Flutter/Dart Example**
```dart
Future<List<Slot>> fetchDoctorSlots(
  String doctorId,
  String clinicId,
  String slotType,
  String? date,
) async {
  final dateQuery = date != null ? '?date=$date' : '';
  final url = '$baseUrl/doctor-time-slots/$doctorId/$clinicId/$slotType$dateQuery';
  
  final response = await http.get(
    Uri.parse(url),
    headers: {'Authorization': 'Bearer $token'},
  );
  
  final data = jsonDecode(response.body);
  return (data['slots'] as List).map((s) => Slot.fromJson(s)).toList();
}
```

---

## 🔄 Complete URL Structure

```
http://localhost:8081/api/organizations/doctor-time-slots/:doctor_id/:clinic_id/:slot_type?date=YYYY-MM-DD
                                                           └─────┬────┘ └────┬───┘ └────┬────┘ └────┬────┘
                                                              Required     Required   Required   Optional
```

### **Example URL Breakdown**
```
http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
                                                           └──────────────────────┬──────────────────────┘ └──────────────────────┬──────────────────────┘ └──┬──┘ └────┬─────┘
                                                                            doctor_id                                        clinic_id                   slot  query param
                                                                                                                                                         type   (optional)
```

---

## ✅ Testing Checklist

- [ ] Test with valid doctor_id, clinic_id, slot_type
- [ ] Test with date query parameter
- [ ] Test without date query parameter
- [ ] Test with invalid UUID format
- [ ] Test with invalid slot_type
- [ ] Test with invalid date format
- [ ] Test without Authorization header
- [ ] Test with doctor who has no slots
- [ ] Test availability calculation (booked_patients, available_spots)
- [ ] Test status values (available, booking_full)

---

## 🎯 Your Specific Request

**URL:**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline
```

**With Date (for specific day):**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
```

**cURL Command:**
```bash
curl -X GET \
  "http://localhost:8081/api/organizations/doctor-time-slots/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

