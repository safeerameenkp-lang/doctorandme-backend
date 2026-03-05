# Follow-Up Status Labels - Complete UI Guide 🎨

## ✅ **Problem FIXED!**

The patient list was showing **green (free follow-up)** for ALL patients regardless of actual eligibility.

**Now:** Each patient shows the correct status based on their actual follow-up eligibility with the selected doctor and department.

---

## 🎯 **New Response Structure**

### **API Response Fields:**

```json
{
  "patient": {
    "id": "patient-123",
    "first_name": "John",
    "last_name": "Doe",
    "follow_up_eligibility": {
      "eligible": true,
      "is_free": true,
      "status_label": "free",        // ✅ NEW: "free", "paid", "none", "needs_selection", "error"
      "color_code": "green",          // ✅ NEW: "green", "orange", "gray"
      "message": "Free follow-up available (3 days remaining)",
      "days_remaining": 3,
      "reason": null
    }
  }
}
```

---

## 🎨 **Status Labels (Use These in UI)**

### **1. 🟢 `status_label: "free"` (Green)**

**Meaning:** Patient has an active FREE follow-up with this doctor+department

**Display:**
- Avatar color: Green
- Text: "Free Follow-Up Eligible"
- Subtext: "X days remaining"
- Action: Show follow-up option (NO payment required)

**Example:**
```json
{
  "eligible": true,
  "is_free": true,
  "status_label": "free",
  "color_code": "green",
  "message": "Free follow-up available (3 days remaining)",
  "days_remaining": 3
}
```

---

### **2. 🟠 `status_label: "paid"` (Orange)**

**Meaning:** Patient can book a follow-up BUT must pay (free period expired or already used)

**Display:**
- Avatar color: Orange
- Text: "Paid Follow-Up Available"
- Subtext: "Payment required"
- Action: Show follow-up option (WITH payment required)

**Example:**
```json
{
  "eligible": true,
  "is_free": false,
  "status_label": "paid",
  "color_code": "orange",
  "message": "Follow-up available (payment required)",
  "days_remaining": null
}
```

---

### **3. ⚪ `status_label: "none"` (Gray)**

**Meaning:** No previous appointment with this doctor+department - cannot book follow-up

**Display:**
- Avatar color: Gray
- Text: "No Previous Appointment"
- Subtext: "Book a regular appointment first"
- Action: Hide follow-up option (show regular appointment only)

**Example:**
```json
{
  "eligible": false,
  "is_free": false,
  "status_label": "none",
  "color_code": "gray",
  "message": "No previous appointment with this doctor and department",
  "reason": "No previous appointment found"
}
```

---

### **4. ⚪ `status_label: "needs_selection"` (Gray)**

**Meaning:** Doctor not selected yet - cannot determine eligibility

**Display:**
- Avatar color: Gray
- Text: "Select Doctor"
- Subtext: "Select a doctor to check eligibility"
- Action: Prompt user to select doctor first

**Example:**
```json
{
  "eligible": false,
  "is_free": false,
  "status_label": "needs_selection",
  "color_code": "gray",
  "message": "Please select a doctor to check follow-up eligibility",
  "reason": "Doctor not selected"
}
```

---

### **5. ⚪ `status_label: "error"` (Gray)**

**Meaning:** Error checking eligibility (database issue, etc.)

**Display:**
- Avatar color: Gray
- Text: "Status Unknown"
- Subtext: "Try again"
- Action: Show generic error state

**Example:**
```json
{
  "eligible": false,
  "is_free": false,
  "status_label": "error",
  "color_code": "gray",
  "message": "Could not check follow-up eligibility"
}
```

---

## 🎯 **Frontend Implementation (Recommended)**

### **React/TypeScript Example:**

```typescript
interface FollowUpEligibility {
  eligible: boolean;
  is_free: boolean;
  status_label: 'free' | 'paid' | 'none' | 'needs_selection' | 'error';
  color_code: 'green' | 'orange' | 'gray';
  message?: string;
  days_remaining?: number;
  reason?: string;
}

function PatientCard({ patient }) {
  const eligibility = patient.follow_up_eligibility;
  
  // ✅ Use status_label for UI logic
  const getStatusDisplay = () => {
    switch (eligibility.status_label) {
      case 'free':
        return {
          color: 'green',
          icon: '🟢',
          label: 'Free Follow-Up',
          sublabel: `${eligibility.days_remaining} days remaining`,
          showFollowUp: true,
          requirePayment: false
        };
        
      case 'paid':
        return {
          color: 'orange',
          icon: '🟠',
          label: 'Paid Follow-Up',
          sublabel: 'Payment required',
          showFollowUp: true,
          requirePayment: true
        };
        
      case 'none':
        return {
          color: 'gray',
          icon: '⚪',
          label: 'No History',
          sublabel: 'Book regular appointment',
          showFollowUp: false,
          requirePayment: false
        };
        
      case 'needs_selection':
        return {
          color: 'gray',
          icon: '⚪',
          label: 'Select Doctor',
          sublabel: 'Check eligibility',
          showFollowUp: false,
          requirePayment: false
        };
        
      default:
        return {
          color: 'gray',
          icon: '⚪',
          label: 'Unknown',
          sublabel: 'Try again',
          showFollowUp: false,
          requirePayment: false
        };
    }
  };
  
  const status = getStatusDisplay();
  
  return (
    <div className={`patient-card status-${status.color}`}>
      <Avatar color={status.color}>
        {patient.first_name[0]}{patient.last_name[0]}
      </Avatar>
      
      <div className="patient-info">
        <h3>{patient.first_name} {patient.last_name}</h3>
        <p className="status-label">{status.icon} {status.label}</p>
        <p className="status-sublabel">{status.sublabel}</p>
      </div>
      
      {status.showFollowUp && (
        <button onClick={() => bookFollowUp(patient, status.requirePayment)}>
          Book Follow-Up {status.requirePayment && '(Pay)'}
        </button>
      )}
    </div>
  );
}
```

---

### **Vue.js Example:**

```vue
<template>
  <div :class="`patient-card status-${statusDisplay.color}`">
    <div class="avatar" :style="{ backgroundColor: statusDisplay.color }">
      {{ patient.first_name[0] }}{{ patient.last_name[0] }}
    </div>
    
    <div class="patient-info">
      <h3>{{ patient.first_name }} {{ patient.last_name }}</h3>
      <p class="status-label">{{ statusDisplay.icon }} {{ statusDisplay.label }}</p>
      <p class="status-sublabel">{{ statusDisplay.sublabel }}</p>
    </div>
    
    <button 
      v-if="statusDisplay.showFollowUp"
      @click="bookFollowUp"
      :class="{ 'payment-required': statusDisplay.requirePayment }"
    >
      Book Follow-Up {{ statusDisplay.requirePayment ? '(Pay)' : '' }}
    </button>
  </div>
</template>

<script>
export default {
  props: ['patient'],
  
  computed: {
    statusDisplay() {
      const eligibility = this.patient.follow_up_eligibility;
      
      switch (eligibility.status_label) {
        case 'free':
          return {
            color: 'green',
            icon: '🟢',
            label: 'Free Follow-Up',
            sublabel: `${eligibility.days_remaining} days remaining`,
            showFollowUp: true,
            requirePayment: false
          };
          
        case 'paid':
          return {
            color: 'orange',
            icon: '🟠',
            label: 'Paid Follow-Up',
            sublabel: 'Payment required',
            showFollowUp: true,
            requirePayment: true
          };
          
        case 'none':
          return {
            color: 'gray',
            icon: '⚪',
            label: 'No History',
            sublabel: 'Book regular appointment',
            showFollowUp: false,
            requirePayment: false
          };
          
        case 'needs_selection':
          return {
            color: 'gray',
            icon: '⚪',
            label: 'Select Doctor',
            sublabel: 'Check eligibility',
            showFollowUp: false,
            requirePayment: false
          };
          
        default:
          return {
            color: 'gray',
            icon: '⚪',
            label: 'Unknown',
            sublabel: 'Try again',
            showFollowUp: false,
            requirePayment: false
          };
      }
    }
  }
};
</script>
```

---

## 📋 **API Usage Flow**

### **Step 1: List Patients (With Doctor Selection)**

```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=xxx&department_id=xxx
```

**Response:**
```json
{
  "clinic_id": "clinic-123",
  "total": 3,
  "patients": [
    {
      "id": "patient-1",
      "first_name": "John",
      "last_name": "Doe",
      "follow_up_eligibility": {
        "status_label": "free",
        "color_code": "green",
        "eligible": true,
        "is_free": true,
        "days_remaining": 3
      }
    },
    {
      "id": "patient-2",
      "first_name": "Jane",
      "last_name": "Smith",
      "follow_up_eligibility": {
        "status_label": "paid",
        "color_code": "orange",
        "eligible": true,
        "is_free": false
      }
    },
    {
      "id": "patient-3",
      "first_name": "Bob",
      "last_name": "Johnson",
      "follow_up_eligibility": {
        "status_label": "none",
        "color_code": "gray",
        "eligible": false,
        "is_free": false
      }
    }
  ]
}
```

---

### **Step 2: List Patients (Without Doctor Selection)**

```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx
```

**Response:**
```json
{
  "patients": [
    {
      "id": "patient-1",
      "first_name": "John",
      "follow_up_eligibility": {
        "status_label": "needs_selection",
        "color_code": "gray",
        "eligible": false,
        "is_free": false,
        "message": "Please select a doctor to check follow-up eligibility"
      }
    }
  ]
}
```

**UI Should Show:**
- Gray avatar
- "Select Doctor" label
- No follow-up button

---

## ✅ **Testing Checklist**

### **Test 1: Free Follow-Up**
- [ ] Book regular appointment (Doctor A, Cardiology)
- [ ] Search patient with Doctor A + Cardiology
- [ ] **Expected:** GREEN avatar, "Free Follow-Up Eligible"
- [ ] Book follow-up
- [ ] **Expected:** No payment required, fee = ₹0

---

### **Test 2: Paid Follow-Up (Expired)**
- [ ] Find patient with appointment > 5 days ago
- [ ] Search with same doctor+department
- [ ] **Expected:** ORANGE avatar, "Paid Follow-Up Available"
- [ ] Book follow-up
- [ ] **Expected:** Payment required

---

### **Test 3: Paid Follow-Up (Already Used Free)**
- [ ] Book regular appointment
- [ ] Book free follow-up
- [ ] Try to book another follow-up
- [ ] **Expected:** ORANGE avatar, "Paid Follow-Up Available"

---

### **Test 4: No History**
- [ ] Search patient who never visited this doctor
- [ ] **Expected:** GRAY avatar, "No Previous Appointment"
- [ ] Follow-up button hidden

---

### **Test 5: No Doctor Selected**
- [ ] Search patients without selecting doctor
- [ ] **Expected:** All patients show GRAY, "Select Doctor"

---

## 🚀 **Deployment**

```bash
# Build the service
docker-compose build organization-service

# Deploy
docker-compose up -d organization-service

# Check logs
docker-compose logs organization-service --tail=50
```

---

## 📊 **Status Label Decision Tree**

```
┌─────────────────────┐
│ Doctor Selected?    │
└──────┬──────────────┘
       │
       ├─ NO ──> status_label: "needs_selection" (gray)
       │
       ├─ YES
       │
       └──> Has previous appointment?
             │
             ├─ NO ──> status_label: "none" (gray)
             │
             ├─ YES
             │
             └──> Has active follow-up in follow_ups table?
                   │
                   ├─ YES + is_free = true ──> status_label: "free" (green)
                   │
                   └─ NO or is_free = false ──> status_label: "paid" (orange)
```

---

## ✅ **Summary**

**Before:**
- ❌ All patients showed green (incorrect)
- ❌ No clear way to distinguish free vs paid
- ❌ UI had to guess eligibility

**After:**
- ✅ Correct status for each patient
- ✅ Explicit `status_label` field for UI logic
- ✅ `color_code` for styling
- ✅ Clear messaging
- ✅ Works with doctor+department selection

**Frontend Action Required:**
1. Update UI to use `status_label` field (not just `eligible` or `is_free`)
2. Map `status_label` to colors and messages
3. Hide/show follow-up button based on status
4. Show/hide payment section based on `requirePayment` flag

---

**Deploy and test! The backend is ready!** 🚀✅

