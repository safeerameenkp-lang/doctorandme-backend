# Consultation Types - Complete Guide 🏥

## 🎯 Overview

We have **4 consultation types**: 2 main types + 2 follow-up sub-types.

---

## 📋 Complete Consultation Types

| Type | Category | Description | Payment | UI Label |
|------|----------|-------------|---------|----------|
| `clinic_visit` | Main | Regular in-person clinic visit | Required | 🏥 Clinic Visit |
| `video_consultation` | Main | Regular remote video call | Required | 💻 Video Consultation |
| `follow-up-via-clinic` | Follow-Up | Follow-up in-person visit | FREE | 🔄 Follow-Up (Clinic) |
| `follow-up-via-video` | Follow-Up | Follow-up video call | FREE | 🔄 Follow-Up (Video) |

---

## 🔄 Auto-Detection Logic

The system **automatically detects** if it's a follow-up:

```go
// ✅ Auto-detect in backend
if consultation_type == "follow-up-via-clinic" || consultation_type == "follow-up-via-video" {
    is_follow_up = true
    payment_status = "waived"
    fee_amount = 0.0
}
```

**Result:** You don't need to manually set `is_follow_up` - it's automatic! ✅

---

## 📝 Request Examples

### 1. Regular Clinic Visit ✅

```json
POST /api/appointments/simple
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 10:00:00",
  "consultation_type": "clinic_visit",     // ✅ Main type
  "payment_method": "pay_now",             // ✅ Payment required
  "payment_type": "cash"
}
```

**Result:**
- ✅ Regular appointment
- ✅ Payment required
- ✅ Fee charged

---

### 2. Regular Video Consultation ✅

```json
{
  "consultation_type": "video_consultation",  // ✅ Main type
  "payment_method": "pay_later"              // ✅ Payment required
}
```

**Result:**
- ✅ Regular appointment
- ✅ Payment required
- ✅ Fee charged

---

### 3. Follow-Up Clinic Visit (FREE) ✅

```json
{
  "consultation_type": "follow-up-via-clinic",  // ✅ Follow-up sub-type
  // ✅ NO payment_method needed - auto-waived!
}
```

**Result:**
- ✅ `is_follow_up` = true (auto-set)
- ✅ Payment = "waived" (automatic)
- ✅ Fee = 0.0 (FREE)
- ✅ Validates: 5-day window, doctor match, department match

---

### 4. Follow-Up Video Consultation (FREE) ✅

```json
{
  "consultation_type": "follow-up-via-video",  // ✅ Follow-up sub-type
  // ✅ NO payment_method needed - auto-waived!
}
```

**Result:**
- ✅ `is_follow_up` = true (auto-set)
- ✅ Payment = "waived" (automatic)
- ✅ Fee = 0.0 (FREE)
- ✅ Validates: 5-day window, doctor match, department match

---

## 💻 Flutter Integration

### Step 1: Define Consultation Types

```dart
class ConsultationType {
  // Main types
  static const String clinicVisit = 'clinic_visit';
  static const String videoConsultation = 'video_consultation';
  
  // Follow-up sub-types
  static const String followUpClinic = 'follow-up-via-clinic';
  static const String followUpVideo = 'follow-up-via-video';
  
  // Helper: Check if follow-up
  static bool isFollowUp(String type) {
    return type == followUpClinic || type == followUpVideo;
  }
}
```

---

### Step 2: UI Dropdown

```dart
class AppointmentBookingPage extends StatefulWidget {
  final Patient patient;
  
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  String selectedConsultationType = ConsultationType.clinicVisit;
  
  // ✅ Build consultation type options based on patient eligibility
  List<Map<String, String>> get consultationOptions {
    List<Map<String, String>> options = [
      // Main types (always available)
      {
        'value': ConsultationType.clinicVisit,
        'label': '🏥 Clinic Visit',
        'description': 'In-person visit to clinic',
        'payment': 'Payment required',
      },
      {
        'value': ConsultationType.videoConsultation,
        'label': '💻 Video Consultation',
        'description': 'Online video call',
        'payment': 'Payment required',
      },
    ];
    
    // ✅ Add follow-up options if eligible
    if (widget.patient.followUpEligibility?.eligible == true) {
      options.addAll([
        {
          'value': ConsultationType.followUpClinic,
          'label': '🔄 Follow-Up (Clinic Visit)',
          'description': 'Return visit in-person',
          'payment': '✅ FREE - No payment required',
        },
        {
          'value': ConsultationType.followUpVideo,
          'label': '🔄 Follow-Up (Video)',
          'description': 'Return visit via video',
          'payment': '✅ FREE - No payment required',
        },
      ]);
    }
    
    return options;
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // ✅ Consultation Type Dropdown
          DropdownButtonFormField<String>(
            decoration: InputDecoration(
              labelText: 'Consultation Type',
              border: OutlineInputBorder(),
            ),
            value: selectedConsultationType,
            items: consultationOptions.map((option) {
              return DropdownMenuItem<String>(
                value: option['value'],
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      option['label']!,
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    Text(
                      option['description']!,
                      style: TextStyle(fontSize: 11, color: Colors.grey[600]),
                    ),
                    Text(
                      option['payment']!,
                      style: TextStyle(
                        fontSize: 10,
                        color: option['payment']!.contains('FREE') 
                          ? Colors.green 
                          : Colors.orange,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              );
            }).toList(),
            onChanged: (value) {
              setState(() {
                selectedConsultationType = value!;
              });
            },
          ),
          
          SizedBox(height: 20),
          
          // ✅ Show payment section only for non-follow-up types
          if (!ConsultationType.isFollowUp(selectedConsultationType))
            _buildPaymentSection(),
          
          SizedBox(height: 20),
          
          // Book button
          ElevatedButton(
            onPressed: _bookAppointment,
            child: Text(
              ConsultationType.isFollowUp(selectedConsultationType)
                ? 'Book Follow-Up (Free)'
                : 'Book Appointment'
            ),
          ),
        ],
      ),
    );
  }
  
  Widget _buildPaymentSection() {
    return Column(
      children: [
        Text('Payment Required', style: TextStyle(fontWeight: FontWeight.bold)),
        // ... payment method dropdown
        // ... payment type dropdown
      ],
    );
  }
  
  Future<void> _bookAppointment() async {
    final body = {
      'clinic_patient_id': widget.patient.id,
      'doctor_id': selectedDoctorId,
      'clinic_id': clinicId,
      'individual_slot_id': selectedSlotId,
      'appointment_date': selectedDate,
      'appointment_time': selectedTime,
      'consultation_type': selectedConsultationType,  // ✅ One of 4 types
    };
    
    // ✅ Only add payment for non-follow-up types
    if (!ConsultationType.isFollowUp(selectedConsultationType)) {
      body['payment_method'] = selectedPaymentMethod;
      if (selectedPaymentMethod == 'pay_now') {
        body['payment_type'] = selectedPaymentType;
      }
    }
    
    // ✅ Debug print
    print('📤 Booking: ${selectedConsultationType}');
    print('📤 Is Follow-Up: ${ConsultationType.isFollowUp(selectedConsultationType)}');
    
    final response = await http.post(
      Uri.parse('$baseUrl/appointments/simple'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
      body: jsonEncode(body),
    );
    
    if (response.statusCode == 201) {
      _showSuccess();
    } else {
      _showError(response.body);
    }
  }
}
```

---

## 🔄 Backend Flow

### For Regular Types (clinic_visit, video_consultation):

```
1. Receive request with consultation_type
2. is_follow_up remains false (default)
3. Validate payment_method is provided
4. Calculate fee from doctor settings
5. Set payment_status based on payment_method
6. Create appointment with fee
```

---

### For Follow-Up Types (follow-up-via-clinic, follow-up-via-video):

```
1. Receive request with consultation_type
2. Auto-detect: is_follow_up = true ✅
3. Skip payment validation (not required)
4. Validate follow-up eligibility:
   ✅ Check previous appointment exists
   ✅ Check 5-day window
   ✅ Validate doctor match
   ✅ Validate department match
5. Auto-set: payment_status = "waived", fee = 0.0
6. Create FREE appointment
```

---

## 📊 Slot Type Filtering

When listing slots for appointment, filter by the base type:

| consultation_type | Show Slots of Type |
|-------------------|-------------------|
| `clinic_visit` | `clinic_visit` |
| `video_consultation` | `video_consultation` |
| `follow-up-via-clinic` | `clinic_visit` |
| `follow-up-via-video` | `video_consultation` |

**Flutter Example:**
```dart
String getSlotTypeFilter(String consultationType) {
  if (consultationType == 'follow-up-via-clinic') {
    return 'clinic_visit';
  } else if (consultationType == 'follow-up-via-video') {
    return 'video_consultation';
  } else {
    return consultationType;
  }
}

// Use in API call
final slotType = getSlotTypeFilter(selectedConsultationType);
final url = '$baseUrl/organizations/doctor-session-slots?'
  'doctor_id=$doctorId&'
  'clinic_id=$clinicId&'
  'date=$date&'
  'slot_type=$slotType';
```

---

## ✅ Validation Matrix

| Type | Payment Required | Follow-Up Validation | Fee Charged |
|------|-----------------|---------------------|-------------|
| `clinic_visit` | ✅ Yes | ❌ No | ✅ Yes |
| `video_consultation` | ✅ Yes | ❌ No | ✅ Yes |
| `follow-up-via-clinic` | ❌ No | ✅ Yes | ❌ No (Free) |
| `follow-up-via-video` | ❌ No | ✅ Yes | ❌ No (Free) |

---

## 🧪 Testing Examples

### Test 1: Regular Clinic Visit ✅
```bash
POST /api/appointments/simple
{
  "consultation_type": "clinic_visit",
  "payment_method": "pay_now",
  "payment_type": "cash"
}

# Expected: Success, fee charged
```

---

### Test 2: Regular Video ✅
```bash
{
  "consultation_type": "video_consultation",
  "payment_method": "pay_later"
}

# Expected: Success, fee charged
```

---

### Test 3: Follow-Up Clinic (Eligible) ✅
```bash
{
  "consultation_type": "follow-up-via-clinic"
  // No payment fields
}

# Expected: Success, FREE, payment waived
```

---

### Test 4: Follow-Up Video (Eligible) ✅
```bash
{
  "consultation_type": "follow-up-via-video"
  // No payment fields
}

# Expected: Success, FREE, payment waived
```

---

### Test 5: Follow-Up (Not Eligible) ❌
```bash
{
  "consultation_type": "follow-up-via-clinic"
}

# Patient has no previous appointment
# Expected: Error 400 - "No previous appointment found"
```

---

### Test 6: Follow-Up (Expired) ❌
```bash
{
  "consultation_type": "follow-up-via-video"
}

# Last appointment was 7 days ago
# Expected: Error 400 - "Follow-up period expired"
```

---

## 📋 Complete Request Format

```typescript
interface AppointmentRequest {
  clinic_patient_id: string;
  doctor_id: string;
  clinic_id: string;
  department_id?: string;
  individual_slot_id: string;
  appointment_date: string;  // YYYY-MM-DD
  appointment_time: string;  // YYYY-MM-DD HH:MM:SS
  
  // ✅ One of 4 types
  consultation_type: 
    | 'clinic_visit'           // Regular
    | 'video_consultation'     // Regular
    | 'follow-up-via-clinic'   // Follow-up
    | 'follow-up-via-video';   // Follow-up
  
  reason?: string;
  notes?: string;
  
  // ✅ Required ONLY for regular types (not follow-ups)
  payment_method?: 'pay_now' | 'pay_later' | 'way_off';
  payment_type?: 'cash' | 'card' | 'upi';  // Required if pay_now
}
```

---

## ✅ Summary

| Feature | Status |
|---------|--------|
| Total consultation types | 4 |
| Main types | 2 (clinic_visit, video_consultation) |
| Follow-up sub-types | 2 (follow-up-via-clinic, follow-up-via-video) |
| Auto-detection | ✅ Yes (backend auto-sets is_follow_up) |
| Payment handling | ✅ Auto-waived for follow-ups |
| Validation | ✅ Follow-up rules enforced |
| Slot filtering | ✅ Maps to base types |
| Linter errors | ✅ None |

---

**Status:** ✅ **All 4 consultation types working perfectly!** 🎉

Now frontend can send any of these 4 values and backend will handle them correctly! 🏥💻🔄

