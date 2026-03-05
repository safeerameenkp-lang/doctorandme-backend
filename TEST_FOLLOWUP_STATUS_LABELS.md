# Test Follow-Up Status Labels ⚡

## 🎯 **Quick Test Scenarios**

### **Scenario 1: Free Follow-Up (Green)** 🟢

**Setup:**
1. Book regular appointment (Oct 20, 2025)
   - Doctor: Dr. AB
   - Department: Cardiology
   - Type: Clinic Visit
   - Payment: Pay now

**Test API:**
```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-ab&department_id=cardiology
```

**Expected Response:**
```json
{
  "patients": [
    {
      "id": "patient-123",
      "first_name": "John",
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": true,
        "status_label": "free",
        "color_code": "green",
        "message": "Free follow-up available (5 days remaining)",
        "days_remaining": 5
      }
    }
  ]
}
```

**UI Should Show:**
- ✅ Green avatar
- ✅ "Free Follow-Up Eligible"
- ✅ "5 days remaining"
- ✅ Follow-up button visible
- ✅ Payment NOT required

---

### **Scenario 2: Paid Follow-Up (Orange)** 🟠

**Setup:**
1. Find patient with appointment from Oct 10, 2025 (>5 days ago)
   - OR patient who already used free follow-up

**Test API:**
```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-ab&department_id=cardiology
```

**Expected Response:**
```json
{
  "patients": [
    {
      "id": "patient-456",
      "first_name": "Jane",
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": false,
        "status_label": "paid",
        "color_code": "orange",
        "message": "Follow-up available (payment required)",
        "days_remaining": null
      }
    }
  ]
}
```

**UI Should Show:**
- ✅ Orange avatar
- ✅ "Paid Follow-Up Available"
- ✅ "Payment required"
- ✅ Follow-up button visible
- ✅ Payment section REQUIRED

---

### **Scenario 3: No History (Gray)** ⚪

**Setup:**
1. Search for patient who NEVER visited this doctor+department

**Test API:**
```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-xyz&department_id=neurology
```

**Expected Response:**
```json
{
  "patients": [
    {
      "id": "patient-789",
      "first_name": "Bob",
      "follow_up_eligibility": {
        "eligible": false,
        "is_free": false,
        "status_label": "none",
        "color_code": "gray",
        "message": "No previous appointment with this doctor and department",
        "reason": "No previous appointment found"
      }
    }
  ]
}
```

**UI Should Show:**
- ✅ Gray avatar
- ✅ "No Previous Appointment"
- ✅ "Book regular appointment first"
- ✅ Follow-up button HIDDEN

---

### **Scenario 4: Doctor Not Selected (Gray)** ⚪

**Setup:**
1. Search patients WITHOUT selecting doctor

**Test API:**
```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx
```

**Expected Response:**
```json
{
  "patients": [
    {
      "id": "patient-123",
      "first_name": "John",
      "follow_up_eligibility": {
        "eligible": false,
        "is_free": false,
        "status_label": "needs_selection",
        "color_code": "gray",
        "message": "Please select a doctor to check follow-up eligibility",
        "reason": "Doctor not selected"
      }
    }
  ]
}
```

**UI Should Show:**
- ✅ Gray avatar
- ✅ "Select Doctor"
- ✅ "Select a doctor to check eligibility"
- ✅ Prompt to select doctor

---

## 🔍 **Database Verification**

### **Check Follow-Ups Table:**

```sql
-- List active follow-ups
SELECT 
    cp.first_name || ' ' || cp.last_name as patient_name,
    u.first_name || ' ' || u.last_name as doctor_name,
    dept.name as department,
    f.is_free,
    f.status,
    f.valid_from,
    f.valid_until,
    CURRENT_DATE - f.valid_from as days_since,
    f.valid_until - CURRENT_DATE as days_remaining
FROM follow_ups f
JOIN clinic_patients cp ON cp.id = f.clinic_patient_id
JOIN doctors d ON d.id = f.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = f.department_id
WHERE f.clinic_id = 'your-clinic-id'
  AND f.status = 'active'
ORDER BY f.valid_until DESC;
```

**Expected Output:**
```
patient_name | doctor_name | department | is_free | status | valid_from | valid_until | days_since | days_remaining
-------------|-------------|------------|---------|--------|------------|-------------|------------|----------------
John Doe     | Dr. AB      | Cardiology | true    | active | 2025-10-20 | 2025-10-25  | 0          | 5
```

**If `is_free = true` and `days_remaining > 0`:**
→ API should return `status_label: "free"` ✅

**If `is_free = false` or follow-up expired:**
→ API should return `status_label: "paid"` ✅

**If no row found:**
→ Check if patient has previous appointment
→ If yes: `status_label: "paid"`
→ If no: `status_label: "none"`

---

## 🧪 **Frontend Console Tests**

### **Test 1: Check Response**

```javascript
// In browser console after searching patients
fetch('/api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz', {
  headers: { 'Authorization': 'Bearer YOUR_TOKEN' }
})
.then(r => r.json())
.then(data => {
  console.log('📋 Patients with status:');
  data.patients.forEach(p => {
    console.log(`
      Name: ${p.first_name} ${p.last_name}
      Status Label: ${p.follow_up_eligibility.status_label}
      Color Code: ${p.follow_up_eligibility.color_code}
      Is Free: ${p.follow_up_eligibility.is_free}
      Eligible: ${p.follow_up_eligibility.eligible}
      Message: ${p.follow_up_eligibility.message}
    `);
  });
});
```

**Expected Console Output:**
```
📋 Patients with status:
Name: John Doe
Status Label: free
Color Code: green
Is Free: true
Eligible: true
Message: Free follow-up available (5 days remaining)

Name: Jane Smith
Status Label: paid
Color Code: orange
Is Free: false
Eligible: true
Message: Follow-up available (payment required)

Name: Bob Johnson
Status Label: none
Color Code: gray
Is Free: false
Eligible: false
Message: No previous appointment with this doctor and department
```

---

### **Test 2: Verify UI Rendering**

```javascript
// Check how many of each status
const statusCounts = data.patients.reduce((acc, p) => {
  const status = p.follow_up_eligibility.status_label;
  acc[status] = (acc[status] || 0) + 1;
  return acc;
}, {});

console.log('📊 Status Distribution:', statusCounts);
// Expected: { free: 1, paid: 2, none: 1 }
```

---

## ✅ **Success Criteria**

**Before (Broken):**
- ❌ All patients show green
- ❌ No way to distinguish free vs paid

**After (Fixed):**
- ✅ Each patient shows correct color based on actual status
- ✅ Green = Free follow-up available
- ✅ Orange = Paid follow-up (expired or used)
- ✅ Gray = No follow-up (no history or doctor not selected)
- ✅ `status_label` field makes UI logic simple
- ✅ `color_code` field provides styling

---

## 🚀 **Deploy & Test Flow**

```bash
# 1. Build service
docker-compose build organization-service

# 2. Deploy
docker-compose up -d organization-service

# 3. Check logs
docker-compose logs organization-service --tail=50

# 4. Test API
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:3002/api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz"

# 5. Verify response has status_label and color_code fields
```

---

## 📋 **Quick Checklist**

After deployment:

- [ ] API returns `status_label` field
- [ ] API returns `color_code` field
- [ ] Free follow-ups show `"free"` / `"green"`
- [ ] Paid follow-ups show `"paid"` / `"orange"`
- [ ] No history shows `"none"` / `"gray"`
- [ ] No doctor selected shows `"needs_selection"` / `"gray"`
- [ ] Frontend uses `status_label` for logic
- [ ] Frontend uses `color_code` for styling
- [ ] Follow-up button only shows for `"free"` and `"paid"`
- [ ] Payment section only shows for `"paid"`

**If all checked:** ✅ **Status labels working correctly!**

---

**Test it and verify each scenario!** 🧪✅

