# One-Time Free Follow-Up System - Complete Guide 🔄

## 🎯 New Follow-Up Logic

**Follow-ups are ALWAYS available**, but payment depends on eligibility:

| Scenario | Free or Paid? |
|----------|--------------|
| **First follow-up** within 5 days | ✅ **FREE** (one time only) |
| **Second+ follow-up** within 5 days | 💰 **PAID** (free already used) |
| **Any follow-up** after 5 days | 💰 **PAID** (free period expired) |

---

## 📊 Complete Scenarios

### Scenario 1: First Follow-Up (Within 5 Days) ✅ FREE

```
Last Appointment: 2025-10-17
Current Date: 2025-10-19 (2 days later)
Free Follow-Ups Used: 0

Result:
✅ Can book follow-up
✅ FREE (payment waived)
✅ Fee = 0.00
```

**API Response:**
```json
{
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": true,
    "days_remaining": 3,
    "message": "You have one FREE follow-up available"
  }
}
```

---

### Scenario 2: Second Follow-Up (Within 5 Days) 💰 PAID

```
Last Appointment: 2025-10-17
Current Date: 2025-10-19 (2 days later)
Free Follow-Ups Used: 1 (already used)

Result:
✅ Can book follow-up
💰 PAID (free already used)
💰 Fee = doctor's follow_up_fee
```

**API Response:**
```json
{
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": false,
    "message": "Free follow-up already used. Additional follow-ups require payment."
  }
}
```

---

### Scenario 3: Follow-Up After 5 Days 💰 PAID

```
Last Appointment: 2025-10-10
Current Date: 2025-10-20 (10 days later)

Result:
✅ Can book follow-up
💰 PAID (5-day free period expired)
💰 Fee = doctor's follow_up_fee
```

**API Response:**
```json
{
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": false,
    "message": "Follow-up available but payment required (5-day free period expired)"
  }
}
```

---

### Scenario 4: No Previous Appointment ❌ NOT AVAILABLE

```
Total Appointments: 0

Result:
❌ Cannot book follow-up
```

**API Response:**
```json
{
  "follow_up_eligibility": {
    "eligible": false,
    "is_free": false,
    "reason": "No previous appointment found"
  }
}
```

---

## 🔄 Backend Logic Flow

### 1. Patient List API (GET /clinic-specific-patients)

```go
For each patient:
  1. Query last appointment
  2. If no appointment:
     → eligible = false, is_free = false
  
  3. If has appointment:
     → eligible = true (always!)
     
  4. Calculate days_since
  
  5. If days_since <= 5:
     a. Check if free follow-up already used:
        - Query count of follow-ups with payment_status = 'waived'
        - After last appointment date
     
     b. If count = 0:
        → is_free = true
        → message = "You have one FREE follow-up available"
     
     c. If count > 0:
        → is_free = false
        → message = "Free follow-up already used..."
  
  6. If days_since > 5:
     → is_free = false
     → message = "...payment required (5-day free period expired)"
```

---

### 2. Create Appointment API (POST /appointments/simple)

```go
If consultation_type starts with "follow-up-":
  1. Auto-set: is_follow_up = true
  
  2. Validate previous appointment exists
  3. Validate doctor match
  4. Validate department match
  
  4. Check if FREE or PAID:
     a. Calculate days_since
     
     b. If days_since <= 5:
        - Query: Has free follow-up been used?
        - If NO → isFreeFollowUp = true
        - If YES → isFreeFollowUp = false (paid)
     
     c. If days_since > 5:
        → isFreeFollowUp = false (paid)
  
  5. Set payment:
     - If isFreeFollowUp = true:
       → payment_status = "waived"
       → fee_amount = 0.0
     
     - If isFreeFollowUp = false:
       → Require payment_method
       → fee_amount = doctor's follow_up_fee
```

---

## 💻 Flutter Integration

### Step 1: Patient Model

```dart
class FollowUpEligibility {
  final bool eligible;    // Can book follow-up?
  final bool isFree;      // Is it free?
  final String? reason;
  final int? daysRemaining;
  final String? message;
  
  FollowUpEligibility({
    required this.eligible,
    required this.isFree,
    this.reason,
    this.daysRemaining,
    this.message,
  });
  
  factory FollowUpEligibility.fromJson(Map<String, dynamic> json) {
    return FollowUpEligibility(
      eligible: json['eligible'] ?? false,
      isFree: json['is_free'] ?? false,
      reason: json['reason'],
      daysRemaining: json['days_remaining'],
      message: json['message'],
    );
  }
}
```

---

### Step 2: UI Display

```dart
Widget _buildFollowUpStatus(Patient patient) {
  if (!patient.followUpEligibility.eligible) {
    // No previous appointment
    return Container(
      padding: EdgeInsets.all(8),
      color: Colors.grey[200],
      child: Text('❌ No follow-up available (no previous appointment)'),
    );
  }
  
  if (patient.followUpEligibility.isFree) {
    // ✅ FREE follow-up available
    return Container(
      padding: EdgeInsets.all(8),
      color: Colors.green[50],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.check_circle, color: Colors.green, size: 16),
              SizedBox(width: 4),
              Text(
                'FREE Follow-Up Available',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.green[700],
                ),
              ),
            ],
          ),
          SizedBox(height: 4),
          Text(
            patient.followUpEligibility.message ?? '',
            style: TextStyle(fontSize: 12, color: Colors.green[600]),
          ),
          if (patient.followUpEligibility.daysRemaining != null)
            Text(
              '⏰ ${patient.followUpEligibility.daysRemaining} days remaining',
              style: TextStyle(fontSize: 11, color: Colors.orange[700]),
            ),
        ],
      ),
    );
  } else {
    // 💰 PAID follow-up available
    return Container(
      padding: EdgeInsets.all(8),
      color: Colors.orange[50],
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.payment, color: Colors.orange, size: 16),
              SizedBox(width: 4),
              Text(
                'Paid Follow-Up Available',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.orange[700],
                ),
              ),
            ],
          ),
          SizedBox(height: 4),
          Text(
            patient.followUpEligibility.message ?? '',
            style: TextStyle(fontSize: 12, color: Colors.grey[600]),
          ),
        ],
      ),
    );
  }
}
```

---

### Step 3: Booking Page

```dart
class AppointmentBookingPage extends StatefulWidget {
  final Patient patient;
  
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  String selectedConsultationType = 'clinic_visit';
  
  List<Map<String, dynamic>> get consultationOptions {
    List<Map<String, dynamic>> options = [
      {
        'value': 'clinic_visit',
        'label': '🏥 Clinic Visit',
        'payment': 'Payment required',
        'isFree': false,
      },
      {
        'value': 'video_consultation',
        'label': '💻 Video Consultation',
        'payment': 'Payment required',
        'isFree': false,
      },
    ];
    
    // ✅ Add follow-up options if patient has previous appointment
    if (widget.patient.followUpEligibility.eligible) {
      // Determine label based on free/paid
      String clinicLabel = widget.patient.followUpEligibility.isFree
        ? '🔄 Follow-Up Clinic (✅ FREE)'
        : '🔄 Follow-Up Clinic (💰 Paid)';
      
      String videoLabel = widget.patient.followUpEligibility.isFree
        ? '🔄 Follow-Up Video (✅ FREE)'
        : '🔄 Follow-Up Video (💰 Paid)';
      
      options.addAll([
        {
          'value': 'follow-up-via-clinic',
          'label': clinicLabel,
          'payment': widget.patient.followUpEligibility.isFree 
            ? '✅ FREE' 
            : '💰 Payment required',
          'isFree': widget.patient.followUpEligibility.isFree,
        },
        {
          'value': 'follow-up-via-video',
          'label': videoLabel,
          'payment': widget.patient.followUpEligibility.isFree 
            ? '✅ FREE' 
            : '💰 Payment required',
          'isFree': widget.patient.followUpEligibility.isFree,
        },
      ]);
    }
    
    return options;
  }
  
  // ✅ Check if selected type requires payment
  bool get requiresPayment {
    final option = consultationOptions.firstWhere(
      (opt) => opt['value'] == selectedConsultationType,
      orElse: () => {'isFree': false},
    );
    return !(option['isFree'] as bool);
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // Consultation type dropdown
          DropdownButtonFormField<String>(
            value: selectedConsultationType,
            items: consultationOptions.map((option) {
              return DropdownMenuItem<String>(
                value: option['value'],
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(option['label'], style: TextStyle(fontWeight: FontWeight.bold)),
                    Text(
                      option['payment'],
                      style: TextStyle(
                        fontSize: 11,
                        color: option['isFree'] ? Colors.green : Colors.orange,
                      ),
                    ),
                  ],
                ),
              );
            }).toList(),
            onChanged: (value) {
              setState(() => selectedConsultationType = value!);
            },
          ),
          
          // ✅ Show payment section only if required
          if (requiresPayment)
            _buildPaymentSection(),
          
          // Book button
          ElevatedButton(
            onPressed: _bookAppointment,
            child: Text(requiresPayment ? 'Book Appointment' : 'Book Free Follow-Up'),
          ),
        ],
      ),
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
      'consultation_type': selectedConsultationType,
    };
    
    // ✅ Add payment only if required
    if (requiresPayment) {
      body['payment_method'] = selectedPaymentMethod;
      if (selectedPaymentMethod == 'pay_now') {
        body['payment_type'] = selectedPaymentType;
      }
    }
    
    final response = await http.post(
      Uri.parse('$baseUrl/appointments/simple'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
      body: jsonEncode(body),
    );
    
    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      _showSuccess(data['appointment']);
    }
  }
}
```

---

## 📝 API Examples

### Example 1: First Free Follow-Up ✅

**Patient Status:**
```json
{
  "last_appointment": {
    "date": "2025-10-17",
    "days_since": 2
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": true,              // ✅ FREE
    "days_remaining": 3,
    "message": "You have one FREE follow-up available"
  }
}
```

**Request:**
```json
POST /api/appointments/simple
{
  "consultation_type": "follow-up-via-clinic",
  // ✅ NO payment_method needed - it's free!
  ...
}
```

**Response:**
```json
{
  "appointment": {
    "id": "uuid",
    "fee_amount": 0.00,           // ✅ Free
    "payment_status": "waived",    // ✅ Waived
    ...
  }
}
```

---

### Example 2: Second Follow-Up (Already Used Free) 💰

**Patient Status:**
```json
{
  "last_appointment": {
    "date": "2025-10-17",
    "days_since": 3
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": false,             // ❌ Not free (already used)
    "message": "Free follow-up already used. Additional follow-ups require payment."
  }
}
```

**Request:**
```json
POST /api/appointments/simple
{
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",   // 💰 Payment REQUIRED
  "payment_type": "cash",
  ...
}
```

**Response:**
```json
{
  "appointment": {
    "id": "uuid",
    "fee_amount": 200.00,          // 💰 Charged follow_up_fee
    "payment_status": "paid",       // 💰 Paid
    ...
  }
}
```

---

### Example 3: Follow-Up After 5 Days 💰

**Patient Status:**
```json
{
  "last_appointment": {
    "date": "2025-10-10",
    "days_since": 10
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": false,             // ❌ Not free (period expired)
    "message": "Follow-up available but payment required (5-day free period expired)"
  }
}
```

**Request:**
```json
POST /api/appointments/simple
{
  "consultation_type": "follow-up-via-video",
  "payment_method": "pay_later",  // 💰 Payment REQUIRED
  ...
}
```

**Response:**
```json
{
  "appointment": {
    "id": "uuid",
    "fee_amount": 200.00,          // 💰 Charged
    "payment_status": "pending",    // 💰 Pay later
    ...
  }
}
```

---

## 🧪 Test Timeline

### Patient's Journey:

```
Day 0: Regular appointment
  - Date: 2025-10-17
  - Fee: ₹500
  - Payment: Paid
  
Day 2: First follow-up (FREE) ✅
  - Date: 2025-10-19
  - Fee: ₹0 (FREE)
  - Payment: Waived
  - Within 5 days: ✅
  - Free used: 0 → Now 1
  
Day 4: Second follow-up (PAID) 💰
  - Date: 2025-10-21
  - Fee: ₹200 (follow_up_fee)
  - Payment: Required
  - Within 5 days: ✅ (doesn't matter)
  - Free used: 1 (already used!)
  
Day 12: Third follow-up (PAID) 💰
  - Date: 2025-10-29
  - Fee: ₹200
  - Payment: Required
  - After 5 days: ✅
```

---

## 📊 Decision Table

| Days Since | Free Used? | Result | Fee | Payment |
|------------|-----------|--------|-----|---------|
| 2 days | No (0) | ✅ FREE | ₹0 | Waived |
| 2 days | Yes (1+) | 💰 PAID | ₹200 | Required |
| 7 days | No (0) | 💰 PAID | ₹200 | Required |
| 7 days | Yes (1+) | 💰 PAID | ₹200 | Required |

**Summary:**
- **FREE** = Only when (days ≤ 5 AND free_count = 0)
- **PAID** = All other cases

---

## 🔍 Free Follow-Up Detection Query

```sql
SELECT COUNT(*)
FROM appointments
WHERE clinic_patient_id = $1
  AND clinic_id = $2
  AND doctor_id = $3
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'              -- Only count free follow-ups
  AND appointment_date > $4                   -- After last regular appointment
  AND status NOT IN ('cancelled', 'no_show') -- Only active/completed
```

**Result:**
- Count = 0 → No free follow-up used → **Eligible for free**
- Count > 0 → Free follow-up already used → **Paid**

---

## 📱 Flutter UI States

### State 1: FREE Follow-Up Available

```dart
if (patient.followUpEligibility.eligible && patient.followUpEligibility.isFree) {
  return Container(
    color: Colors.green[50],
    child: Column(
      children: [
        Text('✅ FREE Follow-Up Available'),
        Text('Book within ${patient.followUpEligibility.daysRemaining} days'),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: Colors.green),
          onPressed: () => _bookFreeFollowUp(),
          child: Text('Book FREE Follow-Up'),
        ),
      ],
    ),
  );
}
```

---

### State 2: PAID Follow-Up Available

```dart
if (patient.followUpEligibility.eligible && !patient.followUpEligibility.isFree) {
  return Container(
    color: Colors.orange[50],
    child: Column(
      children: [
        Text('💰 Paid Follow-Up Available'),
        Text(patient.followUpEligibility.message ?? ''),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: Colors.orange),
          onPressed: () => _bookPaidFollowUp(),
          child: Text('Book Follow-Up (Paid)'),
        ),
      ],
    ),
  );
}
```

---

### State 3: No Follow-Up Available

```dart
if (!patient.followUpEligibility.eligible) {
  return Text(patient.followUpEligibility.reason ?? 'No follow-up available');
}
```

---

## ✅ Validation Rules

### For FREE Follow-Up:
1. ✅ Previous appointment exists
2. ✅ Within 5 days of last appointment
3. ✅ No free follow-up already used (since last appointment)
4. ✅ Same doctor as previous appointment
5. ✅ Same department as previous appointment
6. ✅ Slot available

### For PAID Follow-Up:
1. ✅ Previous appointment exists
2. ✅ Same doctor as previous appointment
3. ✅ Same department as previous appointment
4. ✅ Slot available
5. ✅ **Payment required** (can be pay_now, pay_later, way_off)

---

## 📊 Fee Calculation

```go
// For FREE follow-up
if isFreeFollowUp {
    feeAmount = 0.0
    paymentStatus = "waived"
}

// For PAID follow-up (or regular appointment)
else if input.IsFollowUp && followUpFee != nil {
    feeAmount = followUpFee  // Doctor's follow_up_fee
    paymentStatus = based on payment_method
}

else if consultationFee != nil {
    feeAmount = consultationFee  // Doctor's consultation_fee
    paymentStatus = based on payment_method
}
```

---

## ✅ Summary

| Aspect | Status |
|--------|--------|
| Follow-ups always available | ✅ Yes (if has previous appointment) |
| First follow-up within 5 days | ✅ FREE |
| Already used free follow-up | 💰 PAID |
| After 5-day period | 💰 PAID |
| Doctor validation | ✅ Enforced |
| Department validation | ✅ Enforced |
| Payment handling | ✅ Smart (free or paid) |
| No linter errors | ✅ Clean |

---

**Status:** ✅ **One-time free follow-up system complete!**

**Benefits:**
- ✅ Patients get one FREE follow-up within 5 days
- ✅ Additional follow-ups available but PAID
- ✅ Clear messaging in UI (free vs paid)
- ✅ Automatic detection and enforcement

**Done!** 🔄✅🎉

