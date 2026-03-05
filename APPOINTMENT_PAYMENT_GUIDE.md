# Appointment Payment Method Guide 💳

## 🎯 Payment Method Options

The appointment creation API supports **3 payment methods**:

| Payment Method | Description | Requires Payment Type | Payment Status |
|---------------|-------------|----------------------|----------------|
| **pay_now** | Payment collected at booking | ✅ Yes (cash/card/upi) | `paid` |
| **pay_later** | Payment deferred | ❌ No | `pending` |
| **way_off** | Payment waived/exempted | ❌ No | `waived` |

---

## 📝 API Request Structure

### Endpoint
```
POST /api/appointments/simple
```

### Required Fields

```json
{
  "clinic_patient_id": "UUID",
  "doctor_id": "UUID",
  "clinic_id": "UUID",
  "individual_slot_id": "UUID",
  "appointment_date": "YYYY-MM-DD",
  "appointment_time": "YYYY-MM-DD HH:MM:SS",
  "consultation_type": "offline|online",
  "payment_method": "pay_now|pay_later|way_off",  // Required
  "payment_type": "cash|card|upi"                  // Required only if pay_now
}
```

---

## ✅ Example Requests

### 1. Pay Now with Cash

```json
{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "cash",
  "reason": "Regular checkup"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid",
    "booking_number": "BN202510180001",
    "token_number": 1,
    "fee_amount": 500.00,
    "payment_status": "paid",           // ✅ Marked as paid
    "payment_mode": "cash",             // ✅ Cash recorded
    "status": "confirmed"
  }
}
```

---

### 2. Pay Now with Card

```json
{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "card"               // Card payment
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "booking_number": "BN202510180002",
    "token_number": 2,
    "payment_status": "paid",
    "payment_mode": "card"              // ✅ Card recorded
  }
}
```

---

### 3. Pay Now with UPI

```json
{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 10:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "upi"                // UPI payment
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "booking_number": "BN202510180003",
    "token_number": 3,
    "payment_status": "paid",
    "payment_mode": "upi"               // ✅ UPI recorded
  }
}
```

---

### 4. Pay Later

```json
{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 11:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_later"        // No payment_type needed
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "booking_number": "BN202510180004",
    "token_number": 4,
    "payment_status": "pending",        // ✅ Marked as pending
    "payment_mode": null,               // ✅ No payment mode
    "fee_amount": 500.00
  }
}
```

---

### 5. Way Off (Payment Waived)

```json
{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 14:30:00",
  "consultation_type": "offline",
  "payment_method": "way_off"          // Payment exempted
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "booking_number": "BN202510180005",
    "token_number": 5,
    "payment_status": "waived",         // ✅ Marked as waived
    "payment_mode": null,               // ✅ No payment mode
    "fee_amount": 0.00
  }
}
```

---

## ❌ Error Cases

### Error 1: Missing payment_type for pay_now

**Request:**
```json
{
  "payment_method": "pay_now"
  // Missing payment_type
}
```

**Response (400):**
```json
{
  "error": "Payment type required",
  "message": "When payment_method is 'pay_now', you must provide payment_type (cash, card, or upi)"
}
```

---

### Error 2: Invalid payment_method

**Request:**
```json
{
  "payment_method": "invalid_method"
}
```

**Response (400):**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.PaymentMethod' Error:Field validation for 'PaymentMethod' failed on the 'oneof' tag"
}
```

---

### Error 3: Invalid payment_type

**Request:**
```json
{
  "payment_method": "pay_now",
  "payment_type": "bitcoin"  // Invalid
}
```

**Response (400):**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.PaymentType' Error:Field validation for 'PaymentType' failed on the 'oneof' tag"
}
```

---

## 📱 Flutter Integration

### Payment Method Model

```dart
enum PaymentMethod {
  payNow,
  payLater,
  wayOff
}

enum PaymentType {
  cash,
  card,
  upi
}

class AppointmentPayment {
  final PaymentMethod method;
  final PaymentType? type;  // Required only for payNow
  
  AppointmentPayment({
    required this.method,
    this.type,
  });
  
  // Validation
  bool isValid() {
    if (method == PaymentMethod.payNow) {
      return type != null;
    }
    return true;
  }
  
  Map<String, dynamic> toJson() {
    final json = {
      'payment_method': method.name.toLowerCase().replaceAll('_', '_'),
    };
    
    if (type != null) {
      json['payment_type'] = type!.name;
    }
    
    return json;
  }
}
```

---

### UI Widget Example

```dart
class PaymentMethodSelector extends StatefulWidget {
  final Function(PaymentMethod, PaymentType?) onPaymentSelected;
  
  const PaymentMethodSelector({required this.onPaymentSelected});
  
  @override
  _PaymentMethodSelectorState createState() => _PaymentMethodSelectorState();
}

class _PaymentMethodSelectorState extends State<PaymentMethodSelector> {
  PaymentMethod? selectedMethod;
  PaymentType? selectedType;
  
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Payment Method Selection
        Text('Payment Method', style: TextStyle(fontWeight: FontWeight.bold)),
        SizedBox(height: 8),
        
        RadioListTile<PaymentMethod>(
          title: Text('Pay Now'),
          value: PaymentMethod.payNow,
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() {
              selectedMethod = value;
              selectedType = null; // Reset payment type
            });
          },
        ),
        
        RadioListTile<PaymentMethod>(
          title: Text('Pay Later'),
          value: PaymentMethod.payLater,
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() {
              selectedMethod = value;
              selectedType = null;
              widget.onPaymentSelected(value!, null);
            });
          },
        ),
        
        RadioListTile<PaymentMethod>(
          title: Text('Way Off (Free)'),
          value: PaymentMethod.wayOff,
          groupValue: selectedMethod,
          onChanged: (value) {
            setState(() {
              selectedMethod = value;
              selectedType = null;
              widget.onPaymentSelected(value!, null);
            });
          },
        ),
        
        // Show payment type selection only if Pay Now is selected
        if (selectedMethod == PaymentMethod.payNow) ...[
          SizedBox(height: 16),
          Text('Payment Type *', style: TextStyle(fontWeight: FontWeight.bold, color: Colors.red)),
          SizedBox(height: 8),
          
          Row(
            children: [
              Expanded(
                child: ElevatedButton.icon(
                  icon: Icon(Icons.money),
                  label: Text('Cash'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: selectedType == PaymentType.cash 
                        ? Colors.green 
                        : Colors.grey[300],
                  ),
                  onPressed: () {
                    setState(() {
                      selectedType = PaymentType.cash;
                    });
                    widget.onPaymentSelected(selectedMethod!, selectedType);
                  },
                ),
              ),
              SizedBox(width: 8),
              Expanded(
                child: ElevatedButton.icon(
                  icon: Icon(Icons.credit_card),
                  label: Text('Card'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: selectedType == PaymentType.card 
                        ? Colors.green 
                        : Colors.grey[300],
                  ),
                  onPressed: () {
                    setState(() {
                      selectedType = PaymentType.card;
                    });
                    widget.onPaymentSelected(selectedMethod!, selectedType);
                  },
                ),
              ),
              SizedBox(width: 8),
              Expanded(
                child: ElevatedButton.icon(
                  icon: Icon(Icons.phone_android),
                  label: Text('UPI'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: selectedType == PaymentType.upi 
                        ? Colors.green 
                        : Colors.grey[300],
                  ),
                  onPressed: () {
                    setState(() {
                      selectedType = PaymentType.upi;
                    });
                    widget.onPaymentSelected(selectedMethod!, selectedType);
                  },
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }
}
```

---

### API Call with Payment

```dart
Future<Map<String, dynamic>> bookAppointment({
  required String clinicPatientId,
  required String doctorId,
  required String clinicId,
  required String slotId,
  required String date,
  required String time,
  required String consultationType,
  required PaymentMethod paymentMethod,
  PaymentType? paymentType,
}) async {
  // Validate
  if (paymentMethod == PaymentMethod.payNow && paymentType == null) {
    throw Exception('Payment type required for Pay Now');
  }
  
  final body = {
    'clinic_patient_id': clinicPatientId,
    'doctor_id': doctorId,
    'clinic_id': clinicId,
    'individual_slot_id': slotId,
    'appointment_date': date,
    'appointment_time': '$date $time:00',
    'consultation_type': consultationType,
    'payment_method': _paymentMethodToString(paymentMethod),
  };
  
  // Add payment_type only if provided
  if (paymentType != null) {
    body['payment_type'] = paymentType.name;
  }
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $token',
    },
    body: jsonEncode(body),
  );
  
  if (response.statusCode == 201) {
    return jsonDecode(response.body);
  } else {
    final error = jsonDecode(response.body);
    throw Exception(error['error']);
  }
}

String _paymentMethodToString(PaymentMethod method) {
  switch (method) {
    case PaymentMethod.payNow:
      return 'pay_now';
    case PaymentMethod.payLater:
      return 'pay_later';
    case PaymentMethod.wayOff:
      return 'way_off';
  }
}
```

---

## 📊 Payment Status Mapping

| Payment Method | Payment Type | Payment Status | Payment Mode | Use Case |
|---------------|-------------|----------------|--------------|----------|
| `pay_now` | `cash` | `paid` | `cash` | Cash paid at counter |
| `pay_now` | `card` | `paid` | `card` | Card payment done |
| `pay_now` | `upi` | `paid` | `upi` | UPI payment done |
| `pay_later` | - | `pending` | `null` | Deferred payment |
| `way_off` | - | `waived` | `null` | Free/exempted |

---

## 🔄 Payment Workflow

```
┌─────────────────────────────────────────────────────────┐
│ APPOINTMENT CREATION WITH PAYMENT                       │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
           ┌────────────────────────┐
           │ Select Payment Method  │
           └────────────────────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
         ▼              ▼              ▼
   ┌─────────┐   ┌──────────┐   ┌─────────┐
   │ Pay Now │   │Pay Later │   │ Way Off │
   └─────────┘   └──────────┘   └─────────┘
         │              │              │
         ▼              ▼              ▼
┌───────────────┐      │              │
│Select Payment │      │              │
│Type (Required)│      │              │
│ • Cash        │      │              │
│ • Card        │      │              │
│ • UPI         │      │              │
└───────────────┘      │              │
         │              │              │
         ▼              ▼              ▼
    payment_status  payment_status payment_status
      = "paid"        = "pending"     = "waived"
```

---

## ✅ Summary

### Payment Method Rules

| Rule | Validation |
|------|-----------|
| **payment_method** | ✅ Required, must be: `pay_now`, `pay_later`, or `way_off` |
| **payment_type** | ✅ Required **only** when `payment_method = pay_now` |
| **payment_type** | ✅ Must be: `cash`, `card`, or `upi` |
| **payment_status** | ✅ Auto-set based on payment_method |
| **payment_mode** | ✅ Auto-set from payment_type (only for pay_now) |

---

**API Endpoint:** `POST /api/appointments/simple`  
**Status:** ✅ Ready with Payment Method Support! 💳🎉

