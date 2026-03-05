# Follow-Up Renewal Test - Verify Fix ✅

## 🎯 **Your Requirement (Confirmed)**

> "Whenever the same patient books a new regular appointment with the same doctor and department, the system must restart or renew the free follow-up period for 5 days from the new appointment's date — even if the old follow-up expired before."

**Status:** ✅ **This is EXACTLY what the fix does!**

---

## 📊 **How It Works Now**

### **The Logic:**

1. **Get most recent regular appointment** (by doctor+department)
   ```sql
   SELECT * FROM appointments
   WHERE consultation_type IN ('clinic_visit', 'video_consultation')
   ORDER BY appointment_date DESC LIMIT 1
   ```

2. **Count free follow-ups from that date onward**
   ```sql
   SELECT COUNT(*) FROM appointments
   WHERE consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
     AND payment_status = 'waived'
     AND appointment_date >= NEW_APPOINTMENT_DATE
   ```

3. **Grant eligibility if count = 0**
   - If no free follow-ups exist from the new appointment date → **ELIGIBLE** ✅
   - If free follow-ups exist from the new appointment date → **NOT ELIGIBLE** ❌

---

## 🧪 **Complete Test Case**

### **Scenario:**

```
📅 Oct 10:  Regular Appointment #1
            ↓ Free follow-up granted (valid Oct 10-15)

📅 Oct 12:  FREE Follow-Up #1 (used within 5 days)
            ↓ Free follow-up used

📅 Oct 16:  Follow-up expired (5 days passed)
            ↓ Status: Expired ❌

📅 Oct 20:  NEW Regular Appointment #2 (same doctor + same dept)
            ✅ System checks: Any free follow-ups from Oct 20 onward?
            ✅ Answer: NO (old one was Oct 12, before Oct 20)
            ✅ Result: FREE FOLLOW-UP RENEWED! ✅
            ✅ New validity: Oct 20-25 (5 days)
```

---

## ✅ **Expected Behavior After Fix**

### **Timeline:**

| Date | Action | Free Follow-Up Status | UI Color |
|------|--------|----------------------|----------|
| Oct 10 | Regular #1 booked | Active (Days 1-5) | 🟢 GREEN |
| Oct 12 | Follow-Up #1 used (FREE) | Used | 🟠 ORANGE |
| Oct 13-16 | Wait (expired) | Expired | 🟠 ORANGE |
| Oct 20 | Regular #2 booked | **RENEWED! (Days 1-5)** | **🟢 GREEN** ✅ |
| Oct 21 | Follow-Up #2 (FREE) | Can book FREE | 🟢 GREEN |
| Oct 22 | Follow-Up #2 used (FREE) | Used | 🟠 ORANGE |

---

## 🧪 **Test Steps**

### **Step 1: Create Expired Scenario**

**Option A: Wait for expiration** (if you already have expired follow-up)
- Skip to Step 2

**Option B: Create test data**
```sql
-- 1. Book regular appointment 10 days ago
INSERT INTO appointments (...) VALUES (
  ...,
  appointment_date = CURRENT_DATE - INTERVAL '10 days',
  consultation_type = 'clinic_visit',
  status = 'confirmed'
);

-- 2. Book free follow-up 8 days ago
INSERT INTO appointments (...) VALUES (
  ...,
  appointment_date = CURRENT_DATE - INTERVAL '8 days',
  consultation_type = 'follow-up-via-clinic',
  payment_status = 'waived',
  status = 'confirmed'
);
```

**Result:** Follow-up is now expired (more than 5 days)

---

### **Step 2: Book New Regular Appointment**

**In your UI:**
```
Doctor: Dr. AB (same as before)
Department: AC (same as before)
Type: 🏥 Clinic Visit (REGULAR, not follow-up)
Date: Today or Tomorrow
Payment: Pay Now (Cash/Card/UPI)
```

**Click:** "Book Now"

**Wait:** 2-3 seconds for auto-refresh

---

### **Step 3: Check Patient Eligibility**

**Search for patient:**
```
Search with Dr. AB + Department AC
```

**Expected Frontend Console:**
```
📋 Patient Card Debug:
   Patient: John Doe
   Total appointments: 2 (or more)
   Total eligibleFollowUps: 1          ✅ MUST BE 1!
   Card Status: free                   ✅ MUST BE 'free'!
   Will show: GREEN                    ✅ MUST SAY GREEN!
   Eligible follow-ups:
      - Dr. AB (AC) - 5 days          ✅ NEW 5-day window!
```

**Expected UI:**
- 🟢 **GREEN avatar** ✅
- **"Free Follow-Up Eligible"** text ✅
- Patient is selectable ✅

---

### **Step 4: Book NEW FREE Follow-Up**

**In your UI:**
```
Doctor: Dr. AB (same)
Department: AC (same)
Type: 🔄 Follow-Up (Clinic)
Payment: NONE (should be FREE)
```

**Click:** "Book Now"

**Expected:** ✅ Books successfully without payment!

---

### **Step 5: Verify Eligibility Used**

**Search for patient again:**
```
Search with Dr. AB + Department AC
```

**Expected:**
- 🟠 **ORANGE avatar** (free follow-up used)
- Cannot book free follow-up again
- Can book paid follow-up

---

## 📋 **API Verification**

**Call the API directly:**

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8081/api/organizations/clinic-specific-patients?clinic_id=XXX&doctor_id=AB&department_id=AC&search=patient_name"
```

**Expected Response:**

```json
{
  "patients": [
    {
      "appointments": [
        {
          "appointment_id": "a002",
          "appointment_date": "2025-10-20",
          "doctor_id": "doctor-ab",
          "department": "AC",
          "consultation_type": "clinic_visit",
          "days_since": 0,
          "remaining_days": 5,
          "status": "active",
          "follow_up_eligible": true,
          "free_follow_up_used": false,
          "note": "Eligible for free follow-up with Dr. AB (AC)"
        },
        {
          "appointment_id": "a001",
          "appointment_date": "2025-10-10",
          "doctor_id": "doctor-ab",
          "department": "AC",
          "days_since": 10,
          "status": "expired",
          "follow_up_eligible": true,
          "free_follow_up_used": false,
          "note": "Older appointment - eligibility reset by newer appointment"
        }
      ],
      "eligible_follow_ups": [
        {
          "appointment_id": "a002",
          "doctor_id": "doctor-ab",
          "doctor_name": "Dr. AB",
          "department": "AC",
          "appointment_date": "2025-10-20",
          "remaining_days": 5,
          "note": "Eligible for free follow-up with Dr. AB (AC)"
        }
      ]
    }
  ]
}
```

**Key Checks:**
- ✅ `eligible_follow_ups` array has 1 entry
- ✅ Entry is for the NEW appointment (Oct 20)
- ✅ `remaining_days: 5` (new 5-day window)
- ✅ Old appointment (Oct 10) shows as "expired" and "superseded"

---

## 🔍 **Debug If Not Working**

### **1. Check Database - Most Recent Appointment**

```sql
SELECT 
    id,
    appointment_date,
    status,
    consultation_type
FROM appointments
WHERE clinic_patient_id = 'PATIENT_ID'
  AND doctor_id = 'DOCTOR_AB'
  AND department_id = 'DEPT_AC'
  AND consultation_type IN ('clinic_visit', 'video_consultation')
  AND status IN ('completed', 'confirmed')
ORDER BY appointment_date DESC, appointment_time DESC
LIMIT 1;
```

**Expected:**
- Most recent appointment should be the new one (Oct 20)
- Status should be 'confirmed'
- consultation_type should be 'clinic_visit' or 'video_consultation'

---

### **2. Check Database - Free Follow-Up Count**

```sql
SELECT COUNT(*) as free_count
FROM appointments
WHERE clinic_patient_id = 'PATIENT_ID'
  AND doctor_id = 'DOCTOR_AB'
  AND department_id = 'DEPT_AC'
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'
  AND appointment_date >= '2025-10-20'  -- NEW appointment date
  AND status NOT IN ('cancelled', 'no_show');
```

**Expected:** `free_count = 0` (no free follow-ups from Oct 20 onward)

**If free_count > 0:** There's a free follow-up from Oct 20 onward, which is why it's not showing as eligible. Check the appointment_date of those follow-ups.

---

### **3. Check Service Status**

```bash
docker-compose ps organization-service
```

**Expected:** Status should be "Up" and recently restarted

---

### **4. Check Service Logs**

```bash
docker-compose logs organization-service --tail=50
```

**Look for:** Any errors or warnings

---

## ✅ **Summary**

**Your requirement:**
> "Whenever the same patient books a new regular appointment with the same doctor and department, the system must restart or renew the free follow-up period for 5 days from the new appointment's date — even if the old follow-up expired before."

**Implementation:**
✅ Gets most recent regular appointment by doctor+department
✅ Counts free follow-ups from that appointment's date onward
✅ If count = 0, grants new free follow-up (RENEWAL)
✅ Old follow-ups before the new appointment date are ignored
✅ Appointments for today/tomorrow are now treated as eligible

**Test the fix:**
1. Book new regular appointment (expired follow-up)
2. Check patient card → Should show 🟢 GREEN
3. Book free follow-up → Should work without payment
4. Eligibility renewed! ✅

---

**The fix is deployed and should work! Test it now and let me know if you see any issues.** 🚀✅


