# Follow-Up Per Doctor + Department - Quick Reference ⚡

## 🎯 **Rule**

**One FREE follow-up per (Doctor + Department) within 5 days**

---

## ✅ **When Follow-Up is FREE**

**ALL must be true:**
- ✅ Same doctor
- ✅ Same department  
- ✅ Within 5 days
- ✅ First follow-up for this doctor+department

---

## ❌ **When Follow-Up is PAID**

**ANY is true:**
- ❌ Different doctor
- ❌ Different department
- ❌ After 5 days
- ❌ Already used free follow-up for this doctor+department

---

## 📊 **Examples**

### Scenario A: Same Everything ✅
```
Last: Doctor A → Cardiology (Oct 15)
New:  Doctor A → Cardiology (Oct 17)
Result: ✅ FREE
```

### Scenario B: Different Department ❌
```
Last: Doctor A → Cardiology (Oct 15)
New:  Doctor A → Neurology (Oct 17)
Result: ❌ PAID (different department)
```

### Scenario C: Different Doctor ❌
```
Last: Doctor A → Cardiology (Oct 15)
New:  Doctor B → Cardiology (Oct 17)
Result: ❌ PAID (different doctor)
```

### Scenario D: Already Used Free ❌
```
Oct 15: Doctor A → Cardiology (Paid)
Oct 16: Doctor A → Cardiology (FREE) ✅
Oct 17: Doctor A → Cardiology (PAID) ❌ Already used
```

### Scenario E: Each Department Gets One Free ✅
```
Oct 15: Doctor A → Cardiology (Paid)
Oct 16: Doctor A → Cardiology (FREE) ✅ First for Cardiology
Oct 18: Doctor A → Neurology (Paid)
Oct 19: Doctor A → Neurology (FREE) ✅ First for Neurology
```

---

## 🔍 **Query**

```sql
-- Count free follow-ups for THIS doctor + THIS department
SELECT COUNT(*)
FROM appointments
WHERE clinic_patient_id = ?
  AND doctor_id = ?          -- Same doctor
  AND department_id = ?       -- Same department
  AND payment_status = 'waived'
  AND appointment_date >= last_appointment_date
```

**Result:**
- COUNT = 0 → FREE ✅
- COUNT > 0 → PAID ❌

---

## 🧪 **Quick Tests**

| Test | Doctor | Dept | Days | Free Before? | Result |
|------|--------|------|------|--------------|--------|
| 1 | Same | Same | 3 | No | ✅ FREE |
| 2 | Same | Same | 3 | Yes | ❌ PAID |
| 3 | Same | Different | 3 | No | ❌ PAID |
| 4 | Different | Same | 3 | No | ❌ PAID |
| 5 | Same | Same | 6 | No | ❌ PAID |

---

## 📋 **Files Changed**

1. `appointment_simple.controller.go` (Line 138-165)
2. `clinic_patient.controller.go` (Line 709-746)

**Change:** Added `AND department_id = $5` to COUNT query

---

## ✅ **Status**

**COMPLETE** - Follow-ups tracked per (Doctor + Department)! 🎉

---

## 🚀 **Deploy**

```bash
docker-compose build appointment-service organization-service
docker-compose up -d appointment-service organization-service
```

---

**Summary:** Each (Doctor + Department) combination gets one free follow-up within 5 days! ✅

