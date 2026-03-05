# Follow-Up Reset - Quick Reference ⚡

## 🎯 **Simple Rule**

**Each regular appointment = NEW free follow-up!**

---

## 📊 **Visual Example**

```
Regular Appointment #1 (Oct 1)
    ↓ Grants: 1 FREE follow-up (valid 5 days)
    ├─ Follow-up (Oct 2) → FREE ✅
    └─ Follow-up (Oct 3) → PAID ❌ (already used)

Regular Appointment #2 (Oct 10) ← NEW BASE
    ↓ Grants: 1 NEW FREE follow-up (valid 5 days)
    ├─ Follow-up (Oct 11) → FREE ✅ RESET!
    └─ Follow-up (Oct 12) → PAID ❌ (already used)

Regular Appointment #3 (Oct 20) ← NEW BASE
    ↓ Grants: 1 NEW FREE follow-up (valid 5 days)
    └─ Follow-up (Oct 21) → FREE ✅ RESET AGAIN!
```

---

## ✅ **Key Points**

1. Each **regular appointment** grants **1 free follow-up**
2. **Unlimited** regular appointments = **unlimited** free follow-ups
3. Eligibility **automatically resets** with each regular visit
4. Old appointments don't affect new eligibility

---

## 🧪 **Quick Test**

```
Day 1:  Regular (paid ₹500)
Day 2:  Follow-up (FREE) ✅
Day 5:  Regular (paid ₹500) ← NEW
Day 6:  Follow-up (FREE) ✅ RESET!
Day 10: Regular (paid ₹500) ← NEW
Day 11: Follow-up (FREE) ✅ RESET!
```

**Total:** 3 regular appointments = 3 free follow-ups! ✅

---

## 💻 **How It Works (Backend)**

```sql
-- Step 1: Find LAST regular appointment
SELECT date FROM appointments
WHERE type IN ('clinic_visit', 'video_consultation')
ORDER BY date DESC LIMIT 1
→ Result: Oct 10

-- Step 2: Count free follow-ups SINCE that date
SELECT COUNT(*) FROM appointments
WHERE type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'
  AND date >= Oct 10  ← Key: Only from last regular
→ Result: 0 = FREE available!
```

---

## ✅ **Status**

**Already implemented!** No changes needed! 🎉

The system automatically resets eligibility with each regular appointment because it only counts follow-ups AFTER the last regular visit.

---

**Remember:** Book regular → Get free follow-up → Repeat! ✅

