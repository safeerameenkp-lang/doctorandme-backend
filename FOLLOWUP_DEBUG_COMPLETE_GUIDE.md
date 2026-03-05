# Follow-Up Reset Debug Guide 🔍

## 🎯 **Your Issue**

After booking a new regular appointment, follow-up still shows **ORANGE (paid)** instead of **GREEN (free/renewed)**.

---

## 🔍 **Debugging Steps**

### **Step 1: Check Backend API Response**

**Correct API URL:**
```
GET http://localhost:8081/api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz&search=patient_name
```

**⚠️ Note:** This requires authentication token in the header.

**Expected Response:**
```json
{
  "clinic_id": "xxx",
  "total": 1,
  "patients": [
    {
      "id": "patient-uuid",
      "first_name": "John",
      "last_name": "Doe",
      "appointments": [
        {
          "appointment_id": "a003",
          "appointment_date": "2025-10-15",
          "doctor_id": "doctor-abc",
          "department": "Cardiology",
          "consultation_type": "clinic_visit",
          "status": "active",
          "follow_up_eligible": true,
          "free_follow_up_used": false
        }
      ],
      "eligible_follow_ups": [
        {
          "appointment_id": "a003",
          "doctor_id": "doctor-abc",
          "department": "Cardiology",
          "remaining_days": 5
        }
      ]
    }
  ]
}
```

**Key Check:** `eligible_follow_ups[]` array should NOT be empty!

---

### **Step 2: Check Database Directly**

**Query 1: Get Latest Regular Appointment**
```sql
SELECT 
    a.id,
    a.appointment_date,
    a.consultation_type,
    a.status,
    COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
    dept.name as department
FROM appointments a
JOIN doctors d ON d.id = a.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = a.department_id
WHERE a.clinic_patient_id = 'PATIENT_ID'
  AND a.clinic_id = 'CLINIC_ID'
  AND a.doctor_id = 'DOCTOR_ID'
  AND a.department_id = 'DEPARTMENT_ID'
  AND a.consultation_type IN ('clinic_visit', 'video_consultation')
  AND a.status IN ('completed', 'confirmed')
ORDER BY a.appointment_date DESC, a.appointment_time DESC;
```

**Query 2: Count Free Follow-Ups from Latest Date**
```sql
SELECT COUNT(*) as free_follow_up_count
FROM appointments
WHERE clinic_patient_id = 'PATIENT_ID'
  AND clinic_id = 'CLINIC_ID'
  AND doctor_id = 'DOCTOR_ID'
  AND department_id = 'DEPARTMENT_ID'
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'
  AND appointment_date >= 'LATEST_DATE'
  AND status NOT IN ('cancelled', 'no_show');
```

**Expected Result:**
- If `COUNT = 0` → Should show **FREE (GREEN)**
- If `COUNT > 0` → Should show **PAID (ORANGE)**

---

### **Step 3: Check Frontend Console**

**Look for these debug messages:**

```
🔍 getFollowUpStatus called:
   doctorId: doctor-abc-uuid
   departmentId: dept-cardio-uuid
   eligibleFollowUps.length: 1          ← Should be 1 if eligible
   appointments.length: 3               ← Should include new appointment

📋 Patient Card Debug:
   Patient: John Doe
   Total appointments: 3
   Total eligibleFollowUps: 1            ← Should be 1 if eligible
   Card Status: free                     ← Should be 'free' not 'paid_expired'
   Will show: GREEN                      ← Should say GREEN
```

---

## 🚨 **Common Issues & Solutions**

### **Issue 1: API Returns Empty eligible_follow_ups**

**Symptoms:**
- Frontend shows ORANGE
- Console shows `eligibleFollowUps.length: 0`
- Database query shows `COUNT = 0` but API doesn't return it

**Possible Causes:**
1. **Authentication issue** - API requires valid token
2. **Backend logic bug** - Query not finding appointments
3. **Service not restarted** - Old code still running

**Solutions:**
1. ✅ Check authentication token in frontend
2. ✅ Restart organization service: `docker-compose restart organization-service`
3. ✅ Check backend logs for errors

---

### **Issue 2: Database Shows COUNT > 0**

**Symptoms:**
- Database query shows free follow-ups already used
- API correctly returns empty array
- Frontend correctly shows ORANGE

**Analysis:**
- ✅ **This is CORRECT behavior!**
- Patient already used their free follow-up
- Need to book ANOTHER regular appointment to reset

**Solution:**
- Book another regular appointment with same doctor+department
- This will create a new base date for follow-up eligibility

---

### **Issue 3: Different Doctor/Department**

**Symptoms:**
- Booked with Dr. ABC, Cardiology
- Searching with Dr. XYZ, Cardiology
- Shows ORANGE (correct - different doctor)

**Solution:**
- ✅ Search with the SAME doctor+department as the appointment

---

### **Issue 4: Auto-Refresh Not Working**

**Symptoms:**
- Books appointment successfully
- No auto-refresh message in console
- UI doesn't update

**Solutions:**
1. ✅ Use manual refresh button (🔄)
2. ✅ Clear search and search again
3. ✅ Wait 2-3 seconds after booking

---

## 🧪 **Test Scripts**

### **Script 1: Quick API Test**
```bash
# Run this script to test the API
chmod +x quick-api-test.sh
./quick-api-test.sh
```

### **Script 2: Comprehensive Test**
```bash
# Run this for detailed debugging
chmod +x test-followup-reset-comprehensive.sh
./test-followup-reset-comprehensive.sh
```

### **Script 3: Database Check**
```bash
# Run this to check database directly
chmod +x debug-followup-status.sh
./debug-followup-status.sh
```

---

## 📊 **Expected Flow (Success)**

### **Before Booking New Regular:**
```
Database: COUNT = 1 (free follow-up used)
API: eligible_follow_ups = []
Frontend: ORANGE avatar
```

### **After Booking New Regular:**
```
Database: COUNT = 0 (from new appointment date)
API: eligible_follow_ups = [new_appointment]
Frontend: GREEN avatar ✅
```

### **After Booking FREE Follow-Up:**
```
Database: COUNT = 1 (free follow-up used again)
API: eligible_follow_ups = []
Frontend: ORANGE avatar
```

---

## 🔧 **Manual Testing Steps**

### **Step 1: Book Regular Appointment**
1. Select doctor+department
2. Select patient
3. Choose **🏥 Clinic Visit** (not follow-up)
4. Fill payment details
5. Click "Book Now"

### **Step 2: Check Console**
Look for:
```
🔄 Auto-refreshing patient search...
✅ Patient search refreshed
```

### **Step 3: Search Patient Again**
1. Clear search box
2. Type patient name
3. Check first patient card

**Expected:**
- 🟢 GREEN avatar
- "Free Follow-Up Eligible" text

### **Step 4: Book FREE Follow-Up**
1. Select same patient
2. Choose **🔄 Follow-Up (Clinic)**
3. Should NOT require payment
4. Click "Book Now"

**Expected:**
- ✅ Books successfully
- ✅ No payment required

### **Step 5: Check Again**
Search patient again - should show ORANGE (free used)

### **Step 6: Book Another Regular**
Repeat Step 1 - should reset to GREEN again!

---

## 🎯 **Summary**

**The system should work like this:**

```
Regular #1 → FREE Follow-Up → Regular #2 → FREE Follow-Up → Regular #3 → FREE Follow-Up
   ↓            ↓                 ↓            ↓                 ↓            ↓
  Paid       FREE               Paid         FREE              Paid        FREE
            (RESET!)                        (RESET!)                      (RESET!)
```

**Each regular appointment resets the follow-up eligibility!**

---

## 📞 **Next Steps**

1. ✅ **Run the test scripts** to check backend
2. ✅ **Check frontend console** for debug messages
3. ✅ **Verify database queries** return correct counts
4. ✅ **Test the complete flow** manually

**If still not working, share:**
- Console output from frontend
- API response from backend
- Database query results

**The debug scripts will help identify the exact issue!** 🔍✅


