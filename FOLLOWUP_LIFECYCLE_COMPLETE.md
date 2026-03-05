# Follow-Up Lifecycle - Complete Documentation 🔄

## 🎯 **Complete Follow-Up Logic**

Your system now implements the complete follow-up lifecycle!

---

## 📊 **Complete Status Flow**

### **Flow 1: Regular Appointment → Free Follow-Up → Use → Paid**
```
1. Book Regular Appointment
   ↓
2. Follow-up created: status="active", is_free=true
   valid_from: appointment_date
   valid_until: appointment_date + 5 days
   patient status: "active"
   
3. Book Free Follow-Up (within 5 days)
   ↓
4. Follow-up updated: status="used"
   patient status: "used"
   
5. Try to Book Another Follow-Up
   ↓
6. CheckFollowUpEligibility: status="used"
   Result: isFree=false (PAID) ✅
```

### **Flow 2: Regular Appointment → Let Expire → Paid**
```
1. Book Regular Appointment
   ↓
2. Follow-up created: status="active"
   valid_until: appointment_date + 5 days
   
3. Wait 5+ Days
   ↓
4. CheckFollowUpEligibility called
   Auto-Expire: status="expired" ✅
   patient status: "expired" ✅
   
5. Try to Book Follow-Up
   ↓
6. CheckFollowUpEligibility: status="expired"
   Result: isFree=false (PAID) ✅
```

### **Flow 3: Expiry → Renewal → New Free Follow-Up**
```
1. Follow-up status: "expired"
   patient status: "expired"
   
2. Book New Regular Appointment (same doctor+dept)
   ↓
3. Auto-Expire old follow-ups ✅
   old follow-up: status="renewed" ✅
   
4. Create New Follow-Up: status="active" ✅
   patient status: "renewed" → "active" ✅
   
5. Result: New free follow-up available (5 days) ✅
```

---

## ✅ **Implementation Details**

### **1. Auto-Expiration**
```go
// Called automatically in CheckFollowUpEligibility
func (fm *FollowUpManager) ExpireOldFollowUps() {
    // Find expired follow-ups
    UPDATE follow_ups
    SET status = 'expired'
    WHERE status = 'active'
      AND valid_until < CURRENT_DATE
    
    // Update patient status
    UPDATE clinic_patients
    SET current_followup_status = 'expired'
    WHERE last_followup_id = expired_followup_id
}
```

### **2. Mark as Used**
```go
// When booking free follow-up
func (fm *FollowUpManager) MarkFollowUpAsUsed(...) {
    UPDATE follow_ups
    SET status = 'used',
        used_at = CURRENT_TIMESTAMP,
        used_appointment_id = appointment_id
    WHERE status = 'active'
      AND valid_until >= CURRENT_DATE
    
    // Update patient status
    UPDATE clinic_patients
    SET current_followup_status = 'used'
}
```

### **3. Check Eligibility**
```go
// Before returning eligibility, auto-expire old ones
func CheckFollowUpEligibility(...) {
    // FIRST: Expire old follow-ups
    fm.ExpireOldFollowUps()
    
    // THEN: Check if has active follow-up
    activeFollowUp = GetActiveFollowUp()
    
    if activeFollowUp && isFree {
        return isFree=true ✅
    }
    
    if status == "used" || status == "expired" {
        return isFree=false (PAID) ✅
    }
}
```

---

## 🎯 **Rules Enforced**

### **Rule 1: Only ONE Free Follow-Up**
- ✅ First appointment → Creates ONE free follow-up
- ✅ Using it → Status becomes "used"
- ✅ Next follow-up → Must be PAID
- ✅ Backend guarantees this!

### **Rule 2: 5-Day Validity**
- ✅ Created on appointment date
- ✅ Valid until: appointment date + 5 days
- ✅ After 5 days → Auto-expires
- ✅ Status becomes "expired"

### **Rule 3: Auto-Expiration**
- ✅ Called automatically in CheckFollowUpEligibility
- ✅ Expires any follow-up past valid_until
- ✅ Updates patient status
- ✅ Can also be called manually via API

### **Rule 4: Renewal**
- ✅ Booking new regular appointment (same doctor+dept) after expiry
- ✅ Old follow-up marked as "renewed"
- ✅ New free follow-up created
- ✅ Patient gets another 5 days

---

## 📋 **API Endpoints**

### **1. Check Eligibility (Auto-Expires Old Ones)**
```
GET /api/v1/appointments/followup-eligibility?clinic_patient_id=xxx&doctor_id=xxx
→ Automatically expires old follow-ups
→ Returns: { eligible, is_free, message }
```

### **2. Manual Expiry Trigger**
```
POST /api/v1/appointments/followup-eligibility/expire-old
→ Manually expire old follow-ups
→ Returns: { message, expired_count }
```

---

## ✅ **Frontend Must Check**

### **Before Showing Follow-Up Option**

```typescript
const eligibility = await checkEligibility(patientId, doctorId, deptId);

if (eligibility.status === 'active' && eligibility.is_free) {
    // Show: FREE follow-up available
    daysRemaining = eligibility.days_remaining;
}

if (eligibility.status === 'used' || eligibility.status === 'expired') {
    // Show: This follow-up requires payment
}
```

---

## 🎉 **Complete System**

**Backend:**
- ✅ Creates follow-ups
- ✅ Auto-expires old ones
- ✅ Marks as used
- ✅ Tracks status
- ✅ Only ONE free follow-up

**Frontend:**
- ✅ Check eligibility before showing options
- ✅ Show FREE if available
- ✅ Show PAID if used/expired
- ✅ Display days remaining

**Your follow-up system is complete with full expiry logic! 🎉**

