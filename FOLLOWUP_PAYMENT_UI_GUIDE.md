# Follow-Up Payment UI Logic - Complete Guide ✅

## 🎯 **Core Rule**

**One FREE follow-up per doctor+department within 5 days. After that, payment required.**

---

## ✅ **Follow-Up Payment Logic**

| Condition | Eligibility | Payment | UI Behavior |
|-----------|------------|---------|-------------|
| **Within 5 days** + First follow-up | ✅ Eligible | ❌ **FREE** | **HIDE payment section** |
| **Within 5 days** + Already used free | ✅ Eligible | ✅ **Required** | **SHOW payment section** |
| **After 5 days** (expired) | ✅ Eligible | ✅ **Required** | **SHOW payment section** |
| **No previous appointment** | ❌ Not eligible | N/A | Show "New Appointment" |

---

## 📊 **API Response Fields**

Your API already provides all the needed information:

```json
{
  "appointments": [
    {
      "appointment_id": "a001",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "days_since": 2,
      "remaining_days": 3,
      "status": "active",
      "follow_up_eligible": true,
      "free_follow_up_used": false,
      "note": "Eligible for free follow-up..."
    }
  ],
  "eligible_follow_ups": [
    {
      "doctor_id": "doctor-a",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "remaining_days": 3
    }
  ]
}
```

---

## 🎨 **Frontend UI Logic**

### Decision Tree:

```dart
bool shouldHidePayment(Appointment apt) {
  // ✅ HIDE payment if: Active + Free not used
  if (apt.status == 'active' && 
      apt.followUpEligible == true && 
      apt.freeFollowUpUsed == false) {
    return true;  // FREE follow-up - HIDE payment
  }
  
  // ⚠️ SHOW payment for all other cases
  return false;
}
```

---

## 📋 **Detailed Examples**

### Example 1: Within 5 Days, First Follow-Up ✅ FREE

**Patient Appointment:**
```json
{
  "appointment_date": "2025-10-18",
  "days_since": 2,
  "remaining_days": 3,
  "status": "active",
  "follow_up_eligible": true,
  "free_follow_up_used": false
}
```

**UI Behavior:**
```dart
if (appointment.status == 'active' && 
    !appointment.freeFollowUpUsed) {
  // ✅ FREE FOLLOW-UP
  hidePaymentSection();
  showMessage('🎉 This is a FREE follow-up!');
  bookingButton.text = 'Book FREE Follow-Up';
}
```

**Booking Request:**
```json
{
  "consultation_type": "follow-up-via-clinic",
  // ❌ NO payment_method field!
}
```

**Result:** ✅ Booked FREE, payment_status = "waived", fee = 0

---

### Example 2: Within 5 Days, Already Used Free ⚠️ PAID

**Patient Appointment:**
```json
{
  "appointment_date": "2025-10-18",
  "days_since": 2,
  "remaining_days": 3,
  "status": "active",
  "follow_up_eligible": true,
  "free_follow_up_used": true,  // ⚠️ Already used!
  "note": "Free follow-up already used..."
}
```

**UI Behavior:**
```dart
if (appointment.status == 'active' && 
    appointment.freeFollowUpUsed) {
  // ⚠️ PAID FOLLOW-UP
  showPaymentSection();
  showMessage('⚠️ Free follow-up already used. Payment required (₹200)');
  bookingButton.text = 'Book Follow-Up (₹200)';
}
```

**Booking Request:**
```json
{
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",  // ✅ Payment required!
  "payment_type": "cash"
}
```

**Result:** ✅ Booked with payment, payment_status = "paid", fee = 200

---

### Example 3: After 5 Days (Expired) ⚠️ PAID

**Patient Appointment:**
```json
{
  "appointment_date": "2025-10-10",
  "days_since": 10,
  "remaining_days": null,
  "status": "expired",
  "follow_up_eligible": true,
  "free_follow_up_used": false,
  "note": "Follow-up period expired..."
}
```

**UI Behavior:**
```dart
if (appointment.status == 'expired') {
  // ⚠️ EXPIRED - PAID FOLLOW-UP
  showPaymentSection();
  showMessage('🕒 Free period expired. Payment required (₹200)');
  bookingButton.text = 'Book Follow-Up (₹200)';
}
```

**Booking Request:**
```json
{
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",  // ✅ Payment required!
  "payment_type": "cash"
}
```

**Result:** ✅ Booked with payment, payment_status = "paid", fee = 200

---

### Example 4: Future Appointment ⏳ NOT ELIGIBLE

**Patient Appointment:**
```json
{
  "appointment_date": "2025-10-22",
  "days_since": -2,
  "remaining_days": null,
  "status": "future",
  "follow_up_eligible": false,
  "note": "Appointment scheduled for future..."
}
```

**UI Behavior:**
```dart
if (appointment.status == 'future') {
  // ⏳ FUTURE - NOT ELIGIBLE YET
  disableFollowUpButton();
  showMessage('⏳ Follow-up available after appointment date');
  bookingButton.text = 'Follow-Up (After Appointment)';
  bookingButton.enabled = false;
}
```

---

## 🎨 **Complete UI Component**

```dart
Widget buildFollowUpBookingSection(Appointment appointment) {
  // ✅ Case 1: FREE Follow-Up (Active + Not Used)
  if (appointment.status == 'active' && 
      appointment.followUpEligible && 
      !appointment.freeFollowUpUsed) {
    return Column(
      children: [
        // ✅ Show FREE badge
        Container(
          color: Colors.green[50],
          padding: EdgeInsets.all(16),
          child: Row(
            children: [
              Icon(Icons.check_circle, color: Colors.green),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  '🎉 FREE Follow-Up Available (${appointment.remainingDays} days left)',
                  style: TextStyle(color: Colors.green, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
        ),
        
        // ❌ NO PAYMENT SECTION
        
        // Book button
        ElevatedButton(
          onPressed: () => bookFollowUp(
            appointment: appointment,
            isFree: true,  // ✅ No payment
          ),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.green),
          child: Text('Book FREE Follow-Up'),
        ),
      ],
    );
  }
  
  // ⚠️ Case 2: PAID Follow-Up (Active but Used OR Expired)
  if (appointment.followUpEligible) {
    String reason = appointment.freeFollowUpUsed 
        ? 'Free follow-up already used'
        : 'Free period expired';
    
    return Column(
      children: [
        // Show warning
        Container(
          color: Colors.orange[50],
          padding: EdgeInsets.all(16),
          child: Row(
            children: [
              Icon(Icons.warning, color: Colors.orange),
              SizedBox(width: 8),
              Text(reason, style: TextStyle(color: Colors.orange)),
            ],
          ),
        ),
        
        // ✅ SHOW PAYMENT SECTION
        PaymentMethodSelector(
          onSelected: (method, type) {
            setState(() {
              selectedPaymentMethod = method;
              selectedPaymentType = type;
            });
          },
        ),
        
        // Book button
        ElevatedButton(
          onPressed: () => bookFollowUp(
            appointment: appointment,
            isFree: false,  // ✅ Payment required
            paymentMethod: selectedPaymentMethod,
            paymentType: selectedPaymentType,
          ),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.orange),
          child: Text('Book Follow-Up (₹200)'),
        ),
      ],
    );
  }
  
  // ❌ Case 3: Not Eligible
  return Text('Follow-up not available for this appointment');
}
```

---

## 🔧 **Payment Section Component**

```dart
class PaymentMethodSelector extends StatefulWidget {
  final Function(String method, String? type) onSelected;
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text('💳 Select Payment Method'),
        
        // Pay Now
        RadioListTile(
          title: Text('Pay Now'),
          value: 'pay_now',
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() => selectedMethod = value);
            showPaymentTypeOptions();  // Show cash/card/UPI
          },
        ),
        
        // Payment type options (if Pay Now selected)
        if (selectedMethod == 'pay_now') ...[
          RadioListTile(
            title: Text('💵 Cash'),
            value: 'cash',
            groupValue: selectedType,
            onChanged: (value) {
              setState(() => selectedType = value);
              onSelected('pay_now', 'cash');
            },
          ),
          RadioListTile(
            title: Text('💳 Card'),
            value: 'card',
            groupValue: selectedType,
            onChanged: (value) {
              setState(() => selectedType = value);
              onSelected('pay_now', 'card');
            },
          ),
          RadioListTile(
            title: Text('📱 UPI'),
            value: 'upi',
            groupValue: selectedType,
            onChanged: (value) {
              setState(() => selectedType = value);
              onSelected('pay_now', 'upi');
            },
          ),
        ],
        
        // Pay Later
        RadioListTile(
          title: Text('Pay Later'),
          value: 'pay_later',
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() => selectedMethod = value);
            onSelected('pay_later', null);
          },
        ),
        
        // Way Off (Waive payment)
        RadioListTile(
          title: Text('Way Off (Waive)'),
          value: 'way_off',
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() => selectedMethod = value);
            onSelected('way_off', null);
          },
        ),
      ],
    );
  }
}
```

---

## 📤 **Booking API Calls**

### FREE Follow-Up (No Payment):

```dart
Future<void> bookFreeFollowUp(Appointment apt) async {
  final body = {
    "clinic_patient_id": apt.patientId,
    "doctor_id": apt.doctorId,
    "clinic_id": clinicId,
    "department_id": apt.departmentId,
    "individual_slot_id": selectedSlotId,
    "appointment_date": selectedDate,
    "appointment_time": selectedTime,
    "consultation_type": "follow-up-via-clinic",
    // ❌ NO payment_method - Backend will auto-waive
  };
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {'Content-Type': 'application/json'},
    body: json.encode(body),
  );
  
  if (response.statusCode == 201) {
    showSuccess('🎉 FREE Follow-Up Booked!');
  }
}
```

---

### PAID Follow-Up (With Payment):

```dart
Future<void> bookPaidFollowUp(
  Appointment apt,
  String paymentMethod,
  String? paymentType,
) async {
  final body = {
    "clinic_patient_id": apt.patientId,
    "doctor_id": apt.doctorId,
    "clinic_id": clinicId,
    "department_id": apt.departmentId,
    "individual_slot_id": selectedSlotId,
    "appointment_date": selectedDate,
    "appointment_time": selectedTime,
    "consultation_type": "follow-up-via-clinic",
    "payment_method": paymentMethod,  // ✅ Required!
  };
  
  // Add payment_type if pay_now
  if (paymentMethod == 'pay_now' && paymentType != null) {
    body['payment_type'] = paymentType;
  }
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {'Content-Type': 'application/json'},
    body: json.encode(body),
  );
  
  if (response.statusCode == 201) {
    showSuccess('✅ Follow-Up Booked (₹200)!');
  }
}
```

---

## ✅ **Summary Table**

| Status | Days Since | Free Used? | Payment | UI Action |
|--------|-----------|-----------|---------|-----------|
| `active` | ≤ 5 | ❌ No | ❌ FREE | **HIDE** payment section |
| `active` | ≤ 5 | ✅ Yes | ✅ Required | **SHOW** payment section |
| `expired` | > 5 | Any | ✅ Required | **SHOW** payment section |
| `future` | < 0 | Any | N/A | **DISABLE** button |

---

## 🎯 **Key Points**

1. ✅ **FREE follow-up:** Hide payment, send NO payment_method
2. ⚠️ **PAID follow-up:** Show payment, send payment_method + payment_type
3. 🕒 **After 5 days:** Still eligible but requires payment
4. ⏳ **Future appointment:** Not eligible yet

---

## 📊 **Visual Flow**

```
User selects patient with appointment
         ↓
Check: appointment.status == 'active'?
         ↓
    YES ↓           NO → Show payment
         ↓
Check: appointment.freeFollowUpUsed == false?
         ↓
    YES ↓           NO → Show payment
         ↓
   ✅ HIDE PAYMENT SECTION
   Show "FREE Follow-Up" button
```

---

**Result:** Payment section shows/hides automatically based on follow-up eligibility! 🎉✅

