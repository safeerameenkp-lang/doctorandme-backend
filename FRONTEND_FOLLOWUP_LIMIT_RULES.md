# Frontend Follow-Up Limit Rules - Critical Implementation Guide ⚠️

## 🎯 **Important: Only ONE Free Follow-Up Per Doctor/Department**

Your appointment API enforces this rule. Here's what the frontend must implement:

---

## 📋 **Follow-Up Rules**

### **Rule 1: ONE Free Follow-Up Only**
- First regular appointment → Creates FREE follow-up (status: `active`)
- Booking that follow-up → Marks it as USED (status: `used`)
- **After using free follow-up → All subsequent follow-ups are PAID**

### **Rule 2: Per Doctor + Department**
- Free follow-up is specific to SAME doctor AND same department
- Different doctor or department = New regular appointment required

### **Rule 3: 5-Day Validity**
- Free follow-up valid for 5 days only
- After 5 days → Status becomes `expired` → Must be PAID

---

## 🔍 **How to Check Eligibility in Frontend**

### **Before Showing Follow-Up Option**

```typescript
// Check if patient has active free follow-up
const checkFollowUpEligibility = async (patientId: string, doctorId: string, deptId: string) => {
  const response = await fetch(
    `/api/appointments/check-follow-up-eligibility?clinic_patient_id=${patientId}&doctor_id=${doctorId}&department_id=${deptId}`,
    {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      }
    }
  );
  
  const data = await response.json();
  
  return {
    isEligible: data.is_eligible,
    isFree: data.is_free,           // true = FREE, false = PAID
    daysRemaining: data.days_remaining,
    status: data.status            // active, used, expired, renewed
  };
};

// Use this before showing follow-up option
const { isEligible, isFree, status, daysRemaining } = 
  await checkFollowUpEligibility(patientId, doctorId, departmentId);

if (!isEligible) {
  // Don't show follow-up option OR show as PAID only
  console.log('No follow-up available');
  return;
}

if (status === 'used') {
  // Free follow-up already used - Show PAID only
  console.log('Free follow-up already used. Next one is PAID.');
}

if (status === 'expired') {
  // Free follow-up expired - Show PAID only
  console.log('Free follow-up expired. Next one is PAID.');
}

if (status === 'active' && isFree) {
  // Has active free follow-up - Show FREE option
  console.log(`Free follow-up available! ${daysRemaining} days remaining`);
}
```

---

## 🎨 **Frontend UI Implementation**

### **Show/Hide Follow-Up Options Based on Status**

```typescript
// Component to display follow-up availability
const FollowUpOption: React.FC<{patientId: string, doctorId: string, deptId: string}> = 
  ({ patientId, doctorId, deptId }) => {
  const [followUpStatus, setFollowUpStatus] = useState<{
    isEligible: boolean;
    isFree: boolean;
    status: string;
    daysRemaining: number;
  } | null>(null);
  
  useEffect(() => {
    checkFollowUpStatus();
  }, [patientId, doctorId, deptId]);
  
  const checkFollowUpStatus = async () => {
    const status = await checkFollowUpEligibility(patientId, doctorId, deptId);
    setFollowUpStatus(status);
  };
  
  if (!followUpStatus || !followUpStatus.isEligible) {
    return <div>No follow-up available</div>;
  }
  
  if (followUpStatus.status === 'used') {
    return (
      <div className="follow-up-option paid">
        <p>⚠️ Free follow-up already used</p>
        <p>Next follow-up requires payment</p>
        <button onClick={() => bookPaidFollowUp()}>
          Book Paid Follow-Up
        </button>
      </div>
    );
  }
  
  if (followUpStatus.status === 'expired') {
    return (
      <div className="follow-up-option paid">
        <p>⏰ Free follow-up expired</p>
        <p>Next follow-up requires payment</p>
        <button onClick={() => bookPaidFollowUp()}>
          Book Paid Follow-Up
        </button>
      </div>
    );
  }
  
  if (followUpStatus.status === 'active' && followUpStatus.isFree) {
    return (
      <div className="follow-up-option free">
        <p>✅ FREE follow-up available!</p>
        <p>{followUpStatus.daysRemaining} days remaining</p>
        <button onClick={() => bookFreeFollowUp()}>
          Book FREE Follow-Up
        </button>
      </div>
    );
  }
  
  return null;
};
```

---

## 🔄 **Complete Flow Example**

### **Scenario 1: First Time Patient**

```
1. Patient books first regular appointment
   → API creates follow-up with status="active"
   → Frontend shows: "Free follow-up available! 5 days remaining"
   
2. Patient books free follow-up (within 5 days)
   → API marks follow-up as "used"
   → Frontend shows: "Follow-up used"
   
3. Patient tries to book another follow-up
   → API returns: isEligible=false OR isFree=false
   → Frontend shows: "Next follow-up is PAID"
```

### **Scenario 2: Follow-Up Expires**

```
1. Patient has active follow-up
   → Frontend shows: "2 days remaining"
   
2. Patient doesn't book within 5 days
   → Status automatically becomes "expired"
   
3. Patient tries to book follow-up after 5 days
   → API returns: status="expired", isFree=false
   → Frontend shows: "Free follow-up expired. This is PAID"
```

### **Scenario 3: Multiple Free Follow-Ups Attempted**

```
1. Patient books first regular appointment → Gets free follow-up
2. Patient books that free follow-up → Status = "used"
3. Patient tries to book another follow-up immediately
   → API checks: status="used" → Returns: isFree=false
   → Frontend: Shows PAID follow-up only
4. Patient pays and books PAID follow-up
```

---

## ✅ **Frontend Validation Rules**

### **Display Logic**

```typescript
const shouldShowFreeFollowUp = (status: string, isFree: boolean): boolean => {
  // Only show FREE option if:
  // 1. Status is "active" 
  // 2. isFree = true
  // 3. daysRemaining > 0
  
  if (status === 'active' && isFree) {
    return true;  // ✅ Show FREE option
  }
  
  // Otherwise show PAID option only
  return false;
};

const shouldShowPaidFollowUp = (status: string): boolean => {
  // Show PAID option if:
  // - status = "used" (free follow-up already used)
  // - status = "expired" (free follow-up expired)
  // - No active follow-up
  
  return status === 'used' || status === 'expired';
};
```

### **User Messages**

```typescript
const getFollowUpMessage = (status: string, isFree: boolean, daysRemaining: number): string => {
  if (status === 'active' && isFree) {
    return `✅ FREE follow-up available! ${daysRemaining} days remaining. This is your ONE free follow-up with this doctor.`;
  }
  
  if (status === 'used') {
    return '⚠️ You have already used your free follow-up. Next follow-up requires payment.';
  }
  
  if (status === 'expired') {
    return '⏰ Your free follow-up expired. This follow-up requires payment.';
  }
  
  return 'This follow-up requires payment.';
};
```

---

## 🚨 **Critical Rules for Frontend**

### **1. Don't Allow Multiple Free Follow-Ups**

```typescript
// ❌ WRONG - Don't show free option if already used
if (lastFollowUpStatus === 'used') {
  showFreeOption = false;  // ✅ Correct
  showPaidOption = true;   // ✅ Correct
}

// ❌ WRONG - Don't allow booking multiple free follow-ups
// The backend will reject it, but frontend should prevent confusion
```

### **2. Check Status Before Showing Options**

```typescript
// ✅ Always check eligibility before showing follow-up option
const status = await checkFollowUpStatus();

if (status.status === 'active' && status.isFree) {
  // Show FREE option (this is the ONLY free follow-up)
} else {
  // Show PAID option only
}
```

### **3. Display Clear Messages**

```typescript
// ✅ Show clear message about the limitation
if (isFree) {
  message = "This is your ONE and ONLY free follow-up with Dr. [Name]. Use it within 5 days!";
} else {
  message = "Free follow-up already used/expired. This follow-up requires payment.";
}
```

---

## ✅ **Summary**

### **Backend Behavior (Automatic)**
- ✅ Creates ONE free follow-up per regular appointment
- ✅ Marks follow-up as "used" when booked
- ✅ Next follow-up becomes PAID automatically
- ✅ Tracks per doctor+department
- ✅ Expires after 5 days

### **Frontend Must:**
- ✅ Check eligibility before showing follow-up option
- ✅ Display correct status (FREE vs PAID)
- ✅ Show clear messages about limitation
- ✅ Prevent confusion about multiple free follow-ups

**This ensures only ONE free follow-up per doctor/department is ever used! 🎯**

