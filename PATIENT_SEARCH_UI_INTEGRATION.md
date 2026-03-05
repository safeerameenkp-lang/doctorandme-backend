# 🎯 Patient Search with Follow-Up Status - UI Integration Guide

## 📱 **UI Flow: Doctor → Department → Patient Search → Follow-Up Status**

### **Step 1: User Selects Doctor and Department**
```
UI: Doctor Dropdown → Select "Dr. Smith"
UI: Department Dropdown → Select "Cardiology" 
```

### **Step 2: User Searches for Patient**
```
UI: Search Box → Type "John" or "1234567890"
UI: Click Search Button
```

### **Step 3: API Call Made**
```
GET /api/organizations/clinic-specific-patients?clinic_id=clinic-123&doctor_id=doctor-456&department_id=dept-789&search=John
```

### **Step 4: API Response with Follow-Up Status**
```json
{
  "clinic_id": "clinic-123",
  "total": 2,
  "patients": [
    {
      "id": "patient-001",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      "email": "john@example.com",
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": true,
        "status_label": "free",
        "color_code": "green",
        "message": "Free follow-up available (3 days left)",
        "days_remaining": 3
      }
    },
    {
      "id": "patient-002", 
      "first_name": "John",
      "last_name": "Smith",
      "phone": "9876543210",
      "email": "johnsmith@example.com",
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": false,
        "status_label": "paid",
        "color_code": "orange",
        "message": "Free follow-up already used (payment required)"
      }
    },
    {
      "id": "patient-003",
      "first_name": "John",
      "last_name": "Wilson", 
      "phone": "5555555555",
      "email": "johnwilson@example.com",
      "follow_up_eligibility": {
        "eligible": false,
        "is_free": false,
        "status_label": "none",
        "color_code": "gray",
        "message": "No previous appointment with this doctor and department"
      }
    }
  ]
}
```

---

## 🎨 **UI Display Examples**

### **Patient Card 1: Free Follow-Up Available**
```html
<div class="patient-card" style="border-left: 5px solid green;">
  <h3>John Doe</h3>
  <p>Phone: 1234567890</p>
  <div class="follow-up-status" style="background-color: #d4edda; color: #155724;">
    🟢 FREE Follow-Up Available (3 days left)
  </div>
  <button class="book-followup-btn" style="background-color: green;">
    Book Free Follow-Up
  </button>
</div>
```

### **Patient Card 2: Paid Follow-Up Required**
```html
<div class="patient-card" style="border-left: 5px solid orange;">
  <h3>John Smith</h3>
  <p>Phone: 9876543210</p>
  <div class="follow-up-status" style="background-color: #fff3cd; color: #856404;">
    🟠 Paid Follow-Up Required
  </div>
  <button class="book-followup-btn" style="background-color: orange;">
    Book Paid Follow-Up
  </button>
</div>
```

### **Patient Card 3: No Previous Appointment**
```html
<div class="patient-card" style="border-left: 5px solid gray;">
  <h3>John Wilson</h3>
  <p>Phone: 5555555555</p>
  <div class="follow-up-status" style="background-color: #f8f9fa; color: #6c757d;">
    ⚪ No Previous Appointment
  </div>
  <button class="book-regular-btn" style="background-color: blue;">
    Book Regular Appointment
  </button>
</div>
```

---

## 🔧 **Frontend JavaScript Implementation**

### **Patient Search Function**
```javascript
async function searchPatients(doctorId, departmentId, searchTerm) {
    try {
        // Show loading spinner
        showLoadingSpinner();
        
        // Make API call
        const response = await fetch(`/api/organizations/clinic-specific-patients?clinic_id=${clinicId}&doctor_id=${doctorId}&department_id=${departmentId}&search=${searchTerm}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`,
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        // Hide loading spinner
        hideLoadingSpinner();
        
        // Display patients with follow-up status
        displayPatients(data.patients);
        
    } catch (error) {
        console.error('Error searching patients:', error);
        showError('Failed to search patients');
    }
}

function displayPatients(patients) {
    const patientList = document.getElementById('patient-list');
    patientList.innerHTML = '';
    
    patients.forEach(patient => {
        const patientCard = createPatientCard(patient);
        patientList.appendChild(patientCard);
    });
}

function createPatientCard(patient) {
    const card = document.createElement('div');
    card.className = 'patient-card';
    
    // Set border color based on follow-up status
    const followUpStatus = patient.follow_up_eligibility;
    let borderColor = 'gray';
    let statusText = '';
    let statusIcon = '';
    let buttonText = '';
    let buttonColor = '';
    
    if (followUpStatus) {
        switch (followUpStatus.color_code) {
            case 'green':
                borderColor = 'green';
                statusText = followUpStatus.message;
                statusIcon = '🟢';
                buttonText = 'Book Free Follow-Up';
                buttonColor = 'green';
                break;
            case 'orange':
                borderColor = 'orange';
                statusText = followUpStatus.message;
                statusIcon = '🟠';
                buttonText = 'Book Paid Follow-Up';
                buttonColor = 'orange';
                break;
            case 'gray':
                borderColor = 'gray';
                statusText = followUpStatus.message;
                statusIcon = '⚪';
                buttonText = 'Book Regular Appointment';
                buttonColor = 'blue';
                break;
        }
    }
    
    card.style.borderLeft = `5px solid ${borderColor}`;
    
    card.innerHTML = `
        <div class="patient-info">
            <h3>${patient.first_name} ${patient.last_name}</h3>
            <p><strong>Phone:</strong> ${patient.phone}</p>
            <p><strong>Email:</strong> ${patient.email || 'N/A'}</p>
        </div>
        
        <div class="follow-up-status" style="background-color: ${getStatusBackgroundColor(followUpStatus?.color_code)}; color: ${getStatusTextColor(followUpStatus?.color_code)}; padding: 8px; border-radius: 4px; margin: 10px 0;">
            ${statusIcon} ${statusText}
        </div>
        
        <div class="patient-actions">
            <button class="book-appointment-btn" style="background-color: ${buttonColor}; color: white; padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer;" 
                    onclick="bookAppointment('${patient.id}', '${followUpStatus?.color_code || 'gray'}')">
                ${buttonText}
            </button>
        </div>
    `;
    
    return card;
}

function getStatusBackgroundColor(colorCode) {
    switch (colorCode) {
        case 'green': return '#d4edda';
        case 'orange': return '#fff3cd';
        case 'gray': return '#f8f9fa';
        default: return '#f8f9fa';
    }
}

function getStatusTextColor(colorCode) {
    switch (colorCode) {
        case 'green': return '#155724';
        case 'orange': return '#856404';
        case 'gray': return '#6c757d';
        default: return '#6c757d';
    }
}

function bookAppointment(patientId, followUpStatus) {
    if (followUpStatus === 'green') {
        // Book free follow-up
        showBookingModal(patientId, 'follow-up-via-clinic', true);
    } else if (followUpStatus === 'orange') {
        // Book paid follow-up
        showBookingModal(patientId, 'follow-up-via-clinic', false);
    } else {
        // Book regular appointment
        showBookingModal(patientId, 'clinic_visit', false);
    }
}
```

---

## 📱 **Complete UI Flow**

### **1. Doctor Selection**
```javascript
function onDoctorSelected(doctorId) {
    // Clear department and patient search
    clearDepartmentSelection();
    clearPatientSearch();
    
    // Load departments for this doctor
    loadDepartmentsForDoctor(doctorId);
}
```

### **2. Department Selection**
```javascript
function onDepartmentSelected(departmentId) {
    // Clear patient search
    clearPatientSearch();
    
    // Enable patient search
    enablePatientSearch();
}
```

### **3. Patient Search**
```javascript
function onSearchPatients(searchTerm) {
    const doctorId = getSelectedDoctorId();
    const departmentId = getSelectedDepartmentId();
    
    if (!doctorId || !departmentId) {
        showError('Please select both doctor and department first');
        return;
    }
    
    // Search patients with follow-up status
    searchPatients(doctorId, departmentId, searchTerm);
}
```

---

## 🎯 **Follow-Up Status Meanings**

### **🟢 Green - Free Follow-Up Available**
- **Meaning:** Patient had regular appointment within 5 days, free follow-up not used yet
- **Action:** Show "Book Free Follow-Up" button
- **Payment:** No payment required

### **🟠 Orange - Paid Follow-Up Required**
- **Meaning:** Either free follow-up already used OR 5+ days passed since regular appointment
- **Action:** Show "Book Paid Follow-Up" button  
- **Payment:** Payment required

### **⚪ Gray - No Previous Appointment**
- **Meaning:** Patient never had appointment with this doctor+department
- **Action:** Show "Book Regular Appointment" button
- **Payment:** Payment required

---

## 🚀 **Ready to Use!**

Your patient search API is already implemented and ready to show follow-up status! When you:

1. **Select Doctor** → Load departments
2. **Select Department** → Enable patient search  
3. **Search Patient** → API returns patients with follow-up status
4. **Display Results** → Show green/orange/gray status for each patient

The API will automatically show:
- ✅ **Free follow-up available** (green)
- ✅ **Paid follow-up required** (orange)  
- ✅ **No previous appointment** (gray)

**Your UI integration is ready!** 🎉
