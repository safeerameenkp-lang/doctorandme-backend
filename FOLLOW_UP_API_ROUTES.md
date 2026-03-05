# Follow-Up API Routes - Quick Reference

## 🛣️ Routes to Register

### Appointment Service Routes

Add these to your appointment service router:

```go
// Follow-Up Eligibility Routes
followUpGroup := router.Group("/appointments/followup-eligibility")
{
    // Check eligibility for specific doctor+department
    followUpGroup.GET("", controllers.CheckFollowUpEligibility)
    
    // List all active follow-ups for a patient
    followUpGroup.GET("/active", controllers.ListActiveFollowUps)
    
    // Manually expire old follow-ups (for cron jobs)
    followUpGroup.POST("/expire-old", controllers.ExpireOldFollowUps)
}
```

---

## 📋 Complete API Reference

### 1. Check Follow-Up Eligibility

**Endpoint:** `GET /appointments/followup-eligibility`

**Query Parameters:**
- `clinic_patient_id` (required): Patient UUID
- `clinic_id` (required): Clinic UUID
- `doctor_id` (required): Doctor UUID
- `department_id` (optional): Department UUID

**Response:**
```json
{
  "eligibility": {
    "eligible": true,
    "is_free": true,
    "message": "Free follow-up available (3 days remaining)",
    "valid_until": "2025-10-25",
    "days_remaining": 3,
    "doctor_name": "Dr. John Smith",
    "department_name": "Cardiology"
  }
}
```

**Usage:**
```bash
curl "http://localhost:8081/appointments/followup-eligibility?\
clinic_patient_id=abc-123&\
clinic_id=clinic-456&\
doctor_id=doc-789&\
department_id=dept-012"
```

---

### 2. List Active Follow-Ups

**Endpoint:** `GET /appointments/followup-eligibility/active`

**Query Parameters:**
- `clinic_patient_id` (required): Patient UUID
- `clinic_id` (required): Clinic UUID

**Response:**
```json
{
  "total": 2,
  "active_followups": [
    {
      "followup_id": "fu-001",
      "doctor_id": "doc-789",
      "doctor_name": "Dr. John Smith",
      "department_id": "dept-012",
      "department_name": "Cardiology",
      "is_free": true,
      "valid_from": "2025-10-20",
      "valid_until": "2025-10-25",
      "days_remaining": 3,
      "message": "Free follow-up available"
    }
  ]
}
```

**Usage:**
```bash
curl "http://localhost:8081/appointments/followup-eligibility/active?\
clinic_patient_id=abc-123&\
clinic_id=clinic-456"
```

---

### 3. Expire Old Follow-Ups (Maintenance)

**Endpoint:** `POST /appointments/followup-eligibility/expire-old`

**Authorization:** Admin only (recommended)

**Response:**
```json
{
  "message": "Successfully expired old follow-ups",
  "expired_count": 15
}
```

**Usage:**
```bash
# Manual trigger
curl -X POST http://localhost:8081/appointments/followup-eligibility/expire-old

# Or schedule as cron job (daily at midnight)
0 0 * * * curl -X POST http://localhost:8081/appointments/followup-eligibility/expire-old
```

---

## 🔄 Existing Routes (Updated Behavior)

### Create Appointment

**Endpoint:** `POST /appointments/simple`

**NEW BEHAVIOR:**

1. **When creating REGULAR appointment** (`clinic_visit` or `video_consultation`):
   - ✅ Creates follow-up record in `follow_ups` table
   - ✅ Auto-renews any existing follow-ups for same doctor+department

2. **When creating FOLLOW-UP appointment** (`follow-up-via-clinic` or `follow-up-via-video`):
   - ✅ Checks `follow_ups` table for eligibility
   - ✅ If free follow-up available → No payment required
   - ✅ Marks follow-up as `used` in `follow_ups` table

**Example - Regular Appointment:**
```json
POST /appointments/simple
{
  "clinic_patient_id": "abc-123",
  "clinic_id": "clinic-456",
  "doctor_id": "doc-789",
  "department_id": "dept-012",
  "consultation_type": "clinic_visit",
  "individual_slot_id": "slot-001",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 10:00:00",
  "payment_method": "pay_now",
  "payment_type": "cash"
}

// Response includes:
{
  "message": "Appointment created successfully",
  "appointment": {...},
  "followup_granted": true,
  "followup_valid_until": "2025-10-25"
}
```

**Example - Free Follow-Up:**
```json
POST /appointments/simple
{
  "clinic_patient_id": "abc-123",
  "clinic_id": "clinic-456",
  "doctor_id": "doc-789",
  "department_id": "dept-012",
  "consultation_type": "follow-up-via-clinic",
  "individual_slot_id": "slot-002",
  "appointment_date": "2025-10-22",
  "appointment_time": "2025-10-22 14:00:00"
  // NO payment_method required!
}

// Response:
{
  "message": "Appointment created successfully",
  "appointment": {...},
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up"
}
```

---

### Get Patient Details

**Endpoint:** `GET /clinic-specific-patients/:id`

**NEW RESPONSE FIELDS:**

```json
{
  "patient": {
    "id": "abc-123",
    "first_name": "John",
    ...
    
    // NEW: Follow-up eligibility info
    "follow_up_eligibility": {
      "eligible": true,
      "is_free": true,
      "message": "Free follow-up available",
      "days_remaining": 3
    },
    
    // NEW: All active follow-ups (can have multiple for different doctors)
    "eligible_follow_ups": [
      {
        "appointment_id": "appt-001",
        "doctor_id": "doc-789",
        "doctor_name": "Dr. John Smith",
        "department_name": "Cardiology",
        "appointment_date": "2025-10-20",
        "remaining_days": 3,
        "next_followup_expiry": "2025-10-25",
        "note": "Eligible for FREE follow-up"
      }
    ],
    
    // NEW: Expired follow-ups (need renewal)
    "expired_followups": [
      {
        "doctor_id": "doc-999",
        "doctor_name": "Dr. Jane Doe",
        "department_name": "General",
        "expired_on": "2025-10-15",
        "note": "Book new regular appointment to renew"
      }
    ]
  }
}
```

---

## 🎨 Frontend Integration Examples

### 1. Check Eligibility Before Booking

```javascript
// When user selects doctor+department on booking page
async function checkEligibility(patientId, clinicId, doctorId, deptId) {
  const params = new URLSearchParams({
    clinic_patient_id: patientId,
    clinic_id: clinicId,
    doctor_id: doctorId,
    department_id: deptId || ''
  });
  
  const response = await fetch(`/appointments/followup-eligibility?${params}`);
  const { eligibility } = await response.json();
  
  if (eligibility.is_free) {
    showSuccessMessage(`🎉 Free Follow-Up Available! (${eligibility.days_remaining} days left)`);
    hidePaymentSection();
    setConsultationType('follow-up-via-clinic');
  } else if (eligibility.eligible) {
    showWarningMessage('⚠️ Follow-up available but requires payment');
    showPaymentSection();
  } else {
    showInfoMessage('Book as regular appointment');
    setConsultationType('clinic_visit');
  }
}
```

### 2. Display Active Follow-Ups Dashboard

```javascript
// Dashboard widget showing all active follow-ups
async function loadActiveFollowUps(patientId, clinicId) {
  const params = new URLSearchParams({
    clinic_patient_id: patientId,
    clinic_id: clinicId
  });
  
  const response = await fetch(`/appointments/followup-eligibility/active?${params}`);
  const { active_followups } = await response.json();
  
  const html = active_followups.map(fu => `
    <div class="followup-card">
      <h3>${fu.doctor_name}</h3>
      <p>${fu.department_name || 'General'}</p>
      <p class="expires">Expires: ${fu.valid_until} (${fu.days_remaining} days left)</p>
      <button onclick="bookFollowUp('${fu.doctor_id}', '${fu.department_id}')">
        Book Free Follow-Up
      </button>
    </div>
  `).join('');
  
  document.getElementById('followups-widget').innerHTML = html;
}
```

### 3. Patient List with Follow-Up Badges

```javascript
// Show follow-up status in patient list
function renderPatientList(patients) {
  return patients.map(patient => {
    let badge = '';
    
    if (patient.eligible_follow_ups && patient.eligible_follow_ups.length > 0) {
      const count = patient.eligible_follow_ups.length;
      badge = `<span class="badge badge-success">
                 ✅ ${count} Free Follow-Up${count > 1 ? 's' : ''} Available
               </span>`;
    } else if (patient.expired_followups && patient.expired_followups.length > 0) {
      badge = `<span class="badge badge-warning">
                 ⚠️ Follow-Up Expired - Book Regular
               </span>`;
    }
    
    return `
      <tr>
        <td>${patient.first_name} ${patient.last_name}</td>
        <td>${patient.phone}</td>
        <td>${badge}</td>
        <td>
          <button onclick="viewPatient('${patient.id}')">View</button>
        </td>
      </tr>
    `;
  }).join('');
}
```

---

## 🔧 Admin/Maintenance

### Setup Cron Job for Auto-Expiration

Add to your server's crontab:

```bash
# Expire old follow-ups daily at midnight
0 0 * * * curl -X POST http://localhost:8081/appointments/followup-eligibility/expire-old

# Or with authentication if you add it:
0 0 * * * curl -X POST -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  http://localhost:8081/appointments/followup-eligibility/expire-old
```

### Manual Database Queries (for debugging)

```sql
-- Check active follow-ups for a patient
SELECT 
  f.id,
  f.status,
  f.is_free,
  f.valid_from,
  f.valid_until,
  d.doctor_code,
  u.first_name || ' ' || u.last_name as doctor_name,
  dept.name as department
FROM follow_ups f
JOIN doctors d ON d.id = f.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = f.department_id
WHERE f.clinic_patient_id = 'YOUR_PATIENT_ID'
  AND f.clinic_id = 'YOUR_CLINIC_ID'
ORDER BY f.created_at DESC;

-- Count follow-ups by status
SELECT status, COUNT(*) 
FROM follow_ups 
GROUP BY status;

-- Find follow-ups expiring soon
SELECT 
  cp.first_name || ' ' || cp.last_name as patient_name,
  u.first_name || ' ' || u.last_name as doctor_name,
  f.valid_until,
  CURRENT_DATE - f.valid_until as days_remaining
FROM follow_ups f
JOIN clinic_patients cp ON cp.id = f.clinic_patient_id
JOIN doctors d ON d.id = f.doctor_id
JOIN users u ON u.id = d.user_id
WHERE f.status = 'active'
  AND f.valid_until BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '2 days'
ORDER BY f.valid_until ASC;
```

---

## ✅ Testing Checklist

- [ ] Run migration: `025_create_follow_ups_table.sql`
- [ ] Register new routes in appointment service
- [ ] Test: Create regular appointment → Follow-up record created
- [ ] Test: Book free follow-up → No payment required
- [ ] Test: Book another regular → Old follow-up renewed
- [ ] Test: Check eligibility API → Correct response
- [ ] Test: List active follow-ups → Shows all
- [ ] Test: Expire old follow-ups → Status updated
- [ ] Update frontend to use new APIs
- [ ] Setup cron job for auto-expiration

---

## 📊 Monitoring Queries

```sql
-- Daily follow-up statistics
SELECT 
  DATE(created_at) as date,
  status,
  COUNT(*) as count
FROM follow_ups
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY DATE(created_at), status
ORDER BY date DESC, status;

-- Follow-up usage rate
SELECT 
  COUNT(CASE WHEN status = 'used' THEN 1 END) as used_count,
  COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_unused_count,
  ROUND(
    COUNT(CASE WHEN status = 'used' THEN 1 END)::numeric / 
    NULLIF(COUNT(*), 0) * 100, 
    2
  ) as usage_percentage
FROM follow_ups
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days';
```

---

**Ready to deploy! 🚀**

