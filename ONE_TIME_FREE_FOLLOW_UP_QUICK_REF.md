# One-Time Free Follow-Up - Quick Reference 🔄

## 🎯 Rules

| Scenario | Free or Paid? |
|----------|--------------|
| 1st follow-up within 5 days | ✅ **FREE** |
| 2nd+ follow-up within 5 days | 💰 **PAID** |
| Any follow-up after 5 days | 💰 **PAID** |

---

## 📊 Patient Response

```json
{
  "follow_up_eligibility": {
    "eligible": true,        // Can book follow-up?
    "is_free": true,         // Is it free?
    "days_remaining": 3,
    "message": "You have one FREE follow-up available"
  }
}
```

---

## 💻 Flutter Quick Check

```dart
// Check if free
if (patient.followUpEligibility.isFree) {
  // Show "FREE Follow-Up" - no payment needed
} else {
  // Show "Paid Follow-Up" - payment required
}
```

---

## 📝 API Request

### Free Follow-Up:
```json
{
  "consultation_type": "follow-up-via-clinic",
  // NO payment fields
}
```

### Paid Follow-Up:
```json
{
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",  // Required!
  "payment_type": "cash"
}
```

---

## ✅ States

| eligible | is_free | UI Shows |
|----------|---------|----------|
| true | true | ✅ FREE Follow-Up (Book Now!) |
| true | false | 💰 Paid Follow-Up (Payment Required) |
| false | false | ❌ No Follow-Up Available |

---

**Quick Rule:** Only the **FIRST** follow-up within **5 days** is FREE! ✅

