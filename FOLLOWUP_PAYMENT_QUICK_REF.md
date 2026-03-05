# Follow-Up Payment Logic - Quick Reference ⚡

## 🎯 **Simple Rule**

**FREE within 5 days (first time) → HIDE payment**
**ALL other cases → SHOW payment**

---

## 📊 **Decision Table**

| Days Since | Free Used? | Show Payment? | Fee |
|-----------|-----------|---------------|-----|
| ≤ 5 days | ❌ No | ❌ **HIDE** | ₹0 |
| ≤ 5 days | ✅ Yes | ✅ **SHOW** | ₹200 |
| > 5 days | Any | ✅ **SHOW** | ₹200 |

---

## 💻 **Code (Simple)**

```dart
bool shouldShowPayment(Appointment apt) {
  // HIDE only if: active + not used
  if (apt.status == 'active' && !apt.freeFollowUpUsed) {
    return false;  // ❌ HIDE (FREE)
  }
  return true;  // ✅ SHOW (PAID)
}
```

---

## 📤 **API Request**

### FREE (No payment):
```json
{
  "consultation_type": "follow-up-via-clinic"
  // ❌ NO payment_method
}
```

### PAID (With payment):
```json
{
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",  // ✅ Required
  "payment_type": "cash"
}
```

---

## ✅ **UI States**

```
🎉 FREE (days ≤ 5, not used)
   [Book FREE Follow-Up]
   ❌ Payment section hidden

⚠️ PAID (days ≤ 5, used)
   [Book Follow-Up (₹200)]
   ✅ Payment section visible

🕒 PAID (days > 5)
   [Book Follow-Up (₹200)]
   ✅ Payment section visible

⏳ FUTURE (days < 0)
   [Follow-Up After Appointment]
   ❌ Button disabled
```

---

## 🎨 **Quick Implementation**

```dart
// Step 1: Check eligibility
final isFree = (apt.status == 'active' && !apt.freeFollowUpUsed);

// Step 2: Show/hide payment
if (!isFree) {
  showPaymentSection();
}

// Step 3: Book
bookFollowUp(
  isFree: isFree,
  paymentMethod: isFree ? null : selectedMethod,
  paymentType: isFree ? null : selectedType,
);
```

---

**Remember:** Only ONE condition hides payment - everything else shows it! ✅

