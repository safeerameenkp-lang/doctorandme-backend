# Follow-Up Reset Per Regular Appointment ✅

## 🎯 **Core Rule**

**Each NEW regular appointment RESETS follow-up eligibility!**

---

## ✅ **How It Works**

### Rule:
- Each **regular appointment** (clinic_visit, video_consultation) grants **ONE free follow-up within 5 days**
- This eligibility **RESETS** with each new regular appointment
- Patient can have **multiple free follow-ups** with the same doctor+department (one per regular appointment)

---

## 📊 **Complete Example**

### Timeline:

```
📅 Oct 1: Regular Appointment #1 with Dr. ABC (Cardiology)
          ↓ (Grants 1 free follow-up, valid until Oct 6)
          
  Oct 2: ✅ Follow-up #1 (FREE) - Uses the free follow-up from Oct 1
  Oct 3: ⚠️ Follow-up #2 (PAID ₹200) - Free already used for Oct 1 appointment

---

📅 Oct 10: Regular Appointment #2 with Dr. ABC (Cardiology) ← NEW BASE!
           ↓ (Grants NEW free follow-up, valid until Oct 15)
           
  Oct 11: ✅ Follow-up #1 (FREE) ← RESET! New free follow-up available!
  Oct 12: ⚠️ Follow-up #2 (PAID ₹200) - Free already used for Oct 10 appointment

---

📅 Oct 20: Regular Appointment #3 with Dr. ABC (Cardiology) ← NEW BASE!
           ↓ (Grants ANOTHER free follow-up, valid until Oct 25)
           
  Oct 21: ✅ Follow-up #1 (FREE) ← RESET AGAIN!
  Oct 22: ⚠️ Follow-up #2 (PAID ₹200) - Free already used for Oct 20 appointment
```

---

## 🔑 **Key Points**

1. ✅ **Each regular appointment = NEW eligibility**
2. ✅ **Free follow-up counter resets** with each regular appointment
3. ✅ **No limit** on how many regular appointments patient can book
4. ✅ **Each regular appointment grants ONE free follow-up**

---

## 🧪 **Test Scenarios**

### Scenario A: Multiple Regular Appointments in Short Time ✅

**Timeline:**
```
Oct 1:  Regular (Cardiology) - Paid ₹500
Oct 2:  Follow-up (FREE) ✅ - Uses Oct 1 eligibility
Oct 5:  Regular (Cardiology) - Paid ₹500 ← NEW
Oct 6:  Follow-up (FREE) ✅ - Uses Oct 5 eligibility (RESET!)
Oct 8:  Regular (Cardiology) - Paid ₹500 ← NEW
Oct 9:  Follow-up (FREE) ✅ - Uses Oct 8 eligibility (RESET!)
```

**Result:** 3 free follow-ups! ✅ (One per regular appointment)

---

### Scenario B: Regular Appointment After Free Period Expires ✅

**Timeline:**
```
Oct 1:  Regular (Cardiology) - Paid ₹500
Oct 2:  Follow-up (FREE) ✅
Oct 8:  Try follow-up → PAID ₹200 (expired - more than 5 days)
Oct 10: Regular (Cardiology) - Paid ₹500 ← NEW BASE
Oct 11: Follow-up (FREE) ✅ - Eligibility RESET!
```

**Result:** Even though free period expired, new regular appointment grants new free follow-up! ✅

---

### Scenario C: Patient Never Used Free Follow-Up ✅

**Timeline:**
```
Oct 1:  Regular (Cardiology) - Paid ₹500
        (Patient doesn't book follow-up)
Oct 10: Regular (Cardiology) - Paid ₹500 ← NEW BASE
Oct 11: Follow-up (FREE) ✅ - Uses Oct 10 eligibility
```

**Result:** Even though patient never used Oct 1's free follow-up, Oct 10 grants a NEW one! ✅

---

### Scenario D: Multiple Doctors (Independent) ✅

**Timeline:**
```
Oct 1:  Dr. A, Cardiology - Regular
Oct 2:  Dr. A, Cardiology - Follow-up (FREE)
Oct 3:  Dr. B, Cardiology - Regular ← Different doctor
Oct 4:  Dr. B, Cardiology - Follow-up (FREE) ← Independent!
Oct 5:  Dr. A, Cardiology - Regular ← NEW for Dr. A
Oct 6:  Dr. A, Cardiology - Follow-up (FREE) ← RESET for Dr. A!
```

**Result:** Each doctor tracks independently! ✅

---

## 🔍 **How Backend Implements This**

### Step 1: Find Last REGULAR Appointment

```sql
-- Only finds REGULAR appointments (not follow-ups)
SELECT appointment_date
FROM appointments
WHERE clinic_patient_id = ?
  AND doctor_id = ?
  AND department_id = ?
  AND consultation_type IN ('clinic_visit', 'video_consultation')  -- ✅ Only regular
  AND status IN ('completed', 'confirmed')
ORDER BY appointment_date DESC
LIMIT 1
```

**Result:** Gets the date of the LAST regular appointment (e.g., Oct 10)

---

### Step 2: Check Free Follow-Ups SINCE That Date

```sql
-- Only counts follow-ups AFTER the last regular appointment
SELECT COUNT(*)
FROM appointments
WHERE clinic_patient_id = ?
  AND doctor_id = ?
  AND department_id = ?
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'
  AND appointment_date >= ?  -- ✅ Key: >= last regular appointment date
  AND status NOT IN ('cancelled', 'no_show')
```

**Result:**
- If COUNT = 0 → ✅ FREE follow-up available
- If COUNT > 0 → ⚠️ Free already used

**Key:** The `>=` ensures we only count follow-ups AFTER the last regular appointment!

---

## 📊 **Database Example**

### Appointments Table:

| Date | Type | Payment | Notes |
|------|------|---------|-------|
| Oct 1 | clinic_visit | paid | Regular #1 ← Base |
| Oct 2 | follow-up-via-clinic | waived | FREE (for Oct 1) |
| Oct 3 | follow-up-via-clinic | paid | PAID (already used) |
| Oct 10 | clinic_visit | paid | Regular #2 ← NEW Base |
| Oct 11 | follow-up-via-clinic | waived | FREE (for Oct 10) ✅ RESET! |
| Oct 12 | follow-up-via-clinic | paid | PAID (already used for Oct 10) |

---

### Query When Booking Oct 11 Follow-Up:

```sql
-- Step 1: Find last regular
SELECT appointment_date FROM appointments
WHERE ... AND consultation_type IN ('clinic_visit', 'video_consultation')
ORDER BY appointment_date DESC LIMIT 1
```
**Result:** Oct 10 ← Last regular

```sql
-- Step 2: Count free follow-ups since Oct 10
SELECT COUNT(*) FROM appointments
WHERE ... 
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'
  AND appointment_date >= 'Oct 10'  -- ✅ Only from Oct 10 onward
```
**Result:** COUNT = 0 (no follow-ups since Oct 10) → ✅ FREE!

---

## ✅ **UI Impact**

### Patient API Response:

```json
{
  "appointments": [
    {
      "appointment_id": "a003",
      "appointment_date": "2025-10-10",
      "appointment_type": "clinic_visit",
      "days_since": 1,
      "remaining_days": 4,
      "status": "active",
      "follow_up_eligible": true,
      "free_follow_up_used": false,  // ✅ Shows fresh eligibility!
      "note": "Eligible for free follow-up with Dr. ABC (Cardiology)"
    },
    {
      "appointment_id": "a002",
      "appointment_date": "2025-10-03",
      "appointment_type": "follow-up-via-clinic",
      "days_since": 8,
      "status": "expired"
    },
    {
      "appointment_id": "a001",
      "appointment_date": "2025-10-01",
      "appointment_type": "clinic_visit",
      "days_since": 10,
      "status": "expired"
    }
  ],
  "eligible_follow_ups": [
    {
      "appointment_id": "a003",
      "doctor_name": "Dr. ABC",
      "department": "Cardiology",
      "appointment_date": "2025-10-10",
      "remaining_days": 4,
      "note": "Eligible for free follow-up..."
    }
  ]
}
```

**Key:** Even though patient had appointments on Oct 1 and Oct 3, the Oct 10 regular appointment shows as a **fresh eligible follow-up**! ✅

---

## 🎯 **Business Logic Summary**

| Action | Effect on Eligibility |
|--------|---------------------|
| **Book regular appointment** | ✅ Grants NEW free follow-up (resets counter) |
| **Use free follow-up** | ⚠️ Counter increments (no more free for THIS regular appointment) |
| **Book another regular appointment** | ✅ Resets counter (new free follow-up available) |
| **Let 5 days expire** | ⏰ Follow-up still possible but requires payment |
| **Book paid follow-up** | No effect on counter (doesn't reset eligibility) |

---

## ✅ **Advantages**

| Benefit | Description |
|---------|-------------|
| **Fair to patients** | Each visit deserves a follow-up |
| **Encourages follow-ups** | Patients more likely to return for check-up |
| **Clear tracking** | Each regular appointment = separate eligibility |
| **No confusion** | Old appointments don't affect new ones |
| **Business friendly** | Encourages regular appointments |

---

## 🧪 **Testing Checklist**

- [ ] Book regular appointment → Check 1 free follow-up available
- [ ] Use free follow-up → Check next follow-up requires payment
- [ ] Book NEW regular appointment → Check free follow-up RESET
- [ ] Use new free follow-up → Check counter increments again
- [ ] Let 5 days expire → Book new regular → Check fresh eligibility

---

## ✅ **Summary**

**Current Implementation:** ✅ **ALREADY CORRECT!**

The system already implements this correctly:
1. ✅ Each regular appointment is a new base
2. ✅ Free follow-up counter only counts from last regular appointment
3. ✅ New regular appointment resets eligibility
4. ✅ Patient can have unlimited free follow-ups (one per regular visit)

**No changes needed! The system works exactly as you described!** 🎉

---

## 📝 **Key SQL Logic**

```sql
-- This line ensures reset behavior:
AND appointment_date >= $lastRegularAppointmentDate

-- Because it only counts follow-ups AFTER the most recent regular appointment,
-- each new regular appointment automatically "resets" the counter!
```

---

**Result:** Follow-up eligibility automatically resets with each new regular appointment! ✅🎉

