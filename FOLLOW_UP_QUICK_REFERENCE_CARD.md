# Follow-Up System - Quick Reference Card 🎯

## ✅ **What's New**

1. ✅ **Same-day bug fixed** - Only ONE free follow-up per doctor+department
2. ✅ **Per-department tracking** - Each department gets its own free follow-up
3. ✅ **Context-aware** - Pass `doctor_id` & `department_id` to see accurate eligibility

---

## 🔑 **Core Rule**

**ONE FREE follow-up per (Doctor + Department) within 5 days**

---

## 📊 **Quick Decision Table**

| Condition | Result |
|-----------|--------|
| Same doctor + Same dept + ≤5 days + First | ✅ **FREE** |
| Same doctor + Same dept + ≤5 days + Already used | ❌ **PAID** |
| Different doctor OR Different dept | ❌ **PAID/NEW** |
| After 5 days | ❌ **PAID** |
| No previous appointment | ❌ **NEW** |

---

## 🚀 **Frontend Usage**

### API Call:
```
GET /clinic-specific-patients
  ?clinic_id=xxx
  &doctor_id=yyy       ← ADD THIS
  &department_id=zzz   ← ADD THIS
  &search=...
```

### Response:
```json
{
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": true,
    "message": "You have one FREE follow-up available..."
  }
}
```

### UI Logic:
```
if (is_free):       Show "FREE Follow-Up" button
elif (eligible):    Show "Follow-Up (₹200)" button  
else:               Show "New Appointment" button
```

---

## 🧪 **Test Cases**

```
✅ Same doctor, same dept, 2 days → FREE
❌ Same doctor, same dept, 2 days, already used → PAID
❌ Same doctor, diff dept, 2 days → PAID
❌ Diff doctor, same dept, 2 days → PAID
❌ Same doctor, same dept, 6 days → PAID
```

---

## 📝 **Files Changed**

- `appointment_simple.controller.go` (2 changes)
- `clinic_patient.controller.go` (5 changes)
- `024_fix_duplicate_free_followups.sql` (migration)

---

## 🚀 **Deploy**

```bash
docker-compose build appointment-service organization-service
docker-compose up -d
```

---

## ✅ **Status**

**COMPLETE & READY** 🎉

---

**Key Takeaway:** Always pass `doctor_id` and `department_id` when searching patients to get accurate follow-up eligibility! ✅

