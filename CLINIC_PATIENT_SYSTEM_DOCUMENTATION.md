# 🏥 CLINIC PATIENT SYSTEM - COMPLETE DOCUMENTATION

## 🎯 **Overview**
This document provides complete API documentation for the clinic patient system, including follow-up status display, search functionality, and UI integration examples.

---

## 🔗 **API Endpoints**

### 1. **Get Clinic Patient List**
**Endpoint:** `GET /api/v1/clinic-specific-patients`  
**Description:** Retrieve all patients for a clinic with follow-up status and search functionality

#### **Query Parameters:**
- `clinic_id` (required): Clinic ID
- `search` (optional): Search term (searches first_name, last_name, phone, mo_id)
- `only_active` (optional): Show only active patients (default: true)
- `doctor_id` (optional): Filter follow-ups by specific doctor
- `department_id` (optional): Filter follow-ups by specific department

#### **Request Examples:**
```
# Get all active patients
GET /api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2

# Search for specific patient
GET /api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&search=ashiq

# Get patients with follow-ups for specific doctor
GET /api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&doctor_id=ef378478-1091-472e-af40-1655e77985b3

# Get patients with follow-ups for specific department
GET /api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&department_id=ad958b90-d383-4478-bfe3-08b53b8eeef7
```

#### **Response - Patient List with Follow-ups:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "total": 3,
  "patients": [
    {
      "id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
      "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
      "first_name": "ashiq",
      "last_name": "m",
      "phone": "+1234567890",
      "email": "ashiq@example.com",
      "date_of_birth": "1990-01-01",
      "age": 34,
      "gender": "male",
      "address1": "123 Main St",
      "address2": "Apt 4B",
      "district": "Downtown",
      "state": "CA",
      "mo_id": "MO123456",
      "medical_history": "Diabetes",
      "allergies": "None",
      "blood_group": "O+",
      "smoking_status": "non_smoker",
      "alcohol_use": "none",
      "height_cm": 175,
      "weight_kg": 70,
      "is_active": true,
      "created_at": "2025-10-20T08:00:00Z",
      "updated_at": "2025-10-25T10:30:00Z",
      "eligible_follow_ups": [
        {
          "appointment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "appointment_date": "2025-10-27",
          "remaining_days": 3,
          "next_follow_up_expiry": "2025-10-30",
          "note": "Free follow-up available"
        }
      ],
      "expired_followups": [],
      "appointment_history": [
        {
          "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "consultation_type": "clinic_visit",
          "appointment_date": "2025-10-27"
        }
      ]
    },
    {
      "id": "b2c3d4e5-f6g7-8901-bcde-f23456789012",
      "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
      "first_name": "john",
      "last_name": "doe",
      "phone": "+1987654321",
      "email": "john@example.com",
      "date_of_birth": "1985-05-15",
      "age": 39,
      "gender": "male",
      "address1": "456 Oak Ave",
      "address2": "",
      "district": "Uptown",
      "state": "CA",
      "mo_id": "MO789012",
      "medical_history": "Hypertension",
      "allergies": "Penicillin",
      "blood_group": "A+",
      "smoking_status": "former_smoker",
      "alcohol_use": "occasional",
      "height_cm": 180,
      "weight_kg": 85,
      "is_active": true,
      "created_at": "2025-10-22T09:15:00Z",
      "updated_at": "2025-10-25T11:45:00Z",
      "eligible_follow_ups": [],
      "expired_followups": [
        {
          "appointment_id": "",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "expired_on": "2025-10-25",
          "note": "Follow-up expired - create new regular appointment to renew"
        }
      ],
      "appointment_history": [
        {
          "id": "c3d4e5f6-g7h8-9012-cdef-345678901234",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "consultation_type": "clinic_visit",
          "appointment_date": "2025-10-20"
        }
      ]
    },
    {
      "id": "c3d4e5f6-g7h8-9012-cdef-345678901234",
      "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
      "first_name": "jane",
      "last_name": "smith",
      "phone": "+1555123456",
      "email": "jane@example.com",
      "date_of_birth": "1992-12-10",
      "age": 32,
      "gender": "female",
      "address1": "789 Pine St",
      "address2": "Unit 2A",
      "district": "Suburbs",
      "state": "CA",
      "mo_id": "MO345678",
      "medical_history": "Asthma",
      "allergies": "Dust",
      "blood_group": "B+",
      "smoking_status": "non_smoker",
      "alcohol_use": "none",
      "height_cm": 165,
      "weight_kg": 60,
      "is_active": true,
      "created_at": "2025-10-24T14:20:00Z",
      "updated_at": "2025-10-25T16:30:00Z",
      "eligible_follow_ups": [],
      "expired_followups": [],
      "appointment_history": []
    }
  ]
}
```

---

### 2. **Get Single Patient Details**
**Endpoint:** `GET /api/v1/clinic-specific-patients/:id`  
**Description:** Get detailed information for a specific patient

#### **Response:**
```json
{
  "patient": {
    "id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "first_name": "ashiq",
    "last_name": "m",
    "phone": "+1234567890",
    "email": "ashiq@example.com",
    "date_of_birth": "1990-01-01",
    "age": 34,
    "gender": "male",
    "address1": "123 Main St",
    "address2": "Apt 4B",
    "district": "Downtown",
    "state": "CA",
    "mo_id": "MO123456",
    "medical_history": "Diabetes",
    "allergies": "None",
    "blood_group": "O+",
    "smoking_status": "non_smoker",
    "alcohol_use": "none",
    "height_cm": 175,
    "weight_kg": 70,
    "is_active": true,
    "created_at": "2025-10-20T08:00:00Z",
    "updated_at": "2025-10-25T10:30:00Z",
    "eligible_follow_ups": [
      {
        "appointment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
        "doctor_name": "Dr. Smith",
        "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
        "department": "Cardiology",
        "appointment_date": "2025-10-27",
        "remaining_days": 3,
        "next_follow_up_expiry": "2025-10-30",
        "note": "Free follow-up available"
      }
    ],
    "expired_followups": [],
    "appointment_history": [
      {
        "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
        "doctor_name": "Dr. Smith",
        "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
        "department": "Cardiology",
        "consultation_type": "clinic_visit",
        "appointment_date": "2025-10-27"
      }
    ]
  }
}
```

---

### 3. **Create New Patient**
**Endpoint:** `POST /api/v1/clinic-specific-patients`  
**Description:** Create a new clinic-specific patient

#### **Request Body:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "first_name": "new",
  "last_name": "patient",
  "phone": "+1555999888",
  "email": "newpatient@example.com",
  "date_of_birth": "1988-03-20",
  "gender": "female",
  "address1": "123 New St",
  "address2": "Apt 1",
  "district": "New District",
  "state": "CA",
  "mo_id": "MO999999",
  "medical_history": "None",
  "allergies": "None",
  "blood_group": "O+",
  "smoking_status": "non_smoker",
  "alcohol_use": "none",
  "height_cm": 170,
  "weight_kg": 65
}
```

#### **Response:**
```json
{
  "message": "Patient created successfully",
  "patient": {
    "id": "e4f5g6h7-i8j9-0123-efgh-456789012345",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "first_name": "new",
    "last_name": "patient", 
    "phone": "+1555999888",
    "email": "newpatient@example.com",
    "date_of_birth": "1988-03-20",
    "age": 36,
    "gender": "female",
    "address1": "123 New St",
    "address2": "Apt 1",
    "district": "New District",
    "state": "CA",
    "mo_id": "MO999999",
    "medical_history": "None",
    "allergies": "None",
    "blood_group": "O+",
    "smoking_status": "non_smoker",
    "alcohol_use": "none",
    "height_cm": 170,
    "weight_kg": 65,
    "is_active": true,
    "created_at": "2025-10-25T12:00:00Z",
    "updated_at": "2025-10-25T12:00:00Z",
    "eligible_follow_ups": [],
    "expired_followups": [],
    "appointment_history": []
  }
}
```

---

## 🎨 **UI Integration Guide**

### **1. Patient List Component**

```jsx
import React, { useState, useEffect } from 'react';

const PatientList = ({ clinicId }) => {
  const [patients, setPatients] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(false);
  const [filterDoctor, setFilterDoctor] = useState('');
  const [filterDepartment, setFilterDepartment] = useState('');

  // Fetch patients with follow-up data
  const fetchPatients = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        clinic_id: clinicId,
        ...(searchTerm && { search: searchTerm }),
        ...(filterDoctor && { doctor_id: filterDoctor }),
        ...(filterDepartment && { department_id: filterDepartment })
      });

      const response = await fetch(`/api/v1/clinic-specific-patients?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      const data = await response.json();
      setPatients(data.patients || []);
    } catch (error) {
      console.error('Error fetching patients:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPatients();
  }, [clinicId, searchTerm, filterDoctor, filterDepartment]);

  return (
    <div className="patient-list">
      {/* Search and Filters */}
      <div className="filters">
        <input
          type="text"
          placeholder="Search patients (name, phone, MO ID)..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />
        
        <select
          value={filterDoctor}
          onChange={(e) => setFilterDoctor(e.target.value)}
          className="filter-select"
        >
          <option value="">All Doctors</option>
          <option value="ef378478-1091-472e-af40-1655e77985b3">Dr. Smith</option>
          {/* Add more doctors */}
        </select>

        <select
          value={filterDepartment}
          onChange={(e) => setFilterDepartment(e.target.value)}
          className="filter-select"
        >
          <option value="">All Departments</option>
          <option value="ad958b90-d383-4478-bfe3-08b53b8eeef7">Cardiology</option>
          {/* Add more departments */}
        </select>
      </div>

      {/* Patient Cards */}
      <div className="patient-cards">
        {loading ? (
          <div className="loading">Loading patients...</div>
        ) : patients.length === 0 ? (
          <div className="no-patients">No patients found</div>
        ) : (
          patients.map(patient => (
            <PatientCard key={patient.id} patient={patient} />
          ))
        )}
      </div>
    </div>
  );
};

export default PatientList;
```

### **2. Patient Card Component**

```jsx
import React from 'react';

const PatientCard = ({ patient }) => {
  const getFollowUpStatus = () => {
    if (patient.eligible_follow_ups && patient.eligible_follow_ups.length > 0) {
      const followUp = patient.eligible_follow_ups[0];
      return {
        type: 'active',
        message: `✅ Free follow-up available (${followUp.remaining_days} days remaining)`,
        doctor: followUp.doctor_name,
        department: followUp.department,
        expiry: followUp.next_follow_up_expiry
      };
    } else if (patient.expired_followups && patient.expired_followups.length > 0) {
      const expired = patient.expired_followups[0];
      return {
        type: 'expired',
        message: '⚠️ Follow-up expired - create new regular appointment to renew',
        doctor: expired.doctor_name,
        department: expired.department,
        expiredOn: expired.expired_on
      };
    } else {
      return {
        type: 'none',
        message: '❌ No follow-up available'
      };
    }
  };

  const followUpStatus = getFollowUpStatus();

  return (
    <div className="patient-card">
      {/* Patient Basic Info */}
      <div className="patient-header">
        <h3>{patient.first_name} {patient.last_name}</h3>
        <div className="patient-meta">
          <span className="age">{patient.age} years</span>
          <span className="gender">{patient.gender}</span>
          <span className="mo-id">MO: {patient.mo_id}</span>
        </div>
        <div className="contact-info">
          <span className="phone">{patient.phone}</span>
          <span className="email">{patient.email}</span>
        </div>
      </div>

      {/* Follow-up Status */}
      <div className={`followup-status ${followUpStatus.type}`}>
        <div className="status-message">{followUpStatus.message}</div>
        
        {followUpStatus.type === 'active' && (
          <div className="followup-details">
            <div className="doctor">Doctor: {followUpStatus.doctor}</div>
            <div className="department">Department: {followUpStatus.department}</div>
            <div className="expiry">Valid until: {followUpStatus.expiry}</div>
          </div>
        )}

        {followUpStatus.type === 'expired' && (
          <div className="expired-details">
            <div className="doctor">Doctor: {followUpStatus.doctor}</div>
            <div className="department">Department: {followUpStatus.department}</div>
            <div className="expired-date">Expired on: {followUpStatus.expiredOn}</div>
          </div>
        )}
      </div>

      {/* Medical Info */}
      <div className="medical-info">
        <div className="blood-group">Blood Group: {patient.blood_group}</div>
        <div className="medical-history">History: {patient.medical_history}</div>
        <div className="allergies">Allergies: {patient.allergies}</div>
      </div>

      {/* Actions */}
      <div className="patient-actions">
        <button 
          className="btn-primary"
          onClick={() => handleCreateAppointment(patient)}
        >
          Create Appointment
        </button>
        
        {followUpStatus.type === 'active' && (
          <button 
            className="btn-success"
            onClick={() => handleCreateFollowUp(patient)}
          >
            Create Follow-up
          </button>
        )}
        
        <button 
          className="btn-secondary"
          onClick={() => handleViewDetails(patient)}
        >
          View Details
        </button>
      </div>
    </div>
  );
};

export default PatientCard;
```

### **3. Follow-up Status Display Component**

```jsx
import React from 'react';

const FollowUpStatus = ({ patient }) => {
  const getStatusIcon = (type) => {
    switch (type) {
      case 'active': return '✅';
      case 'expired': return '⚠️';
      case 'none': return '❌';
      default: return '❓';
    }
  };

  const getStatusColor = (type) => {
    switch (type) {
      case 'active': return '#10b981'; // green
      case 'expired': return '#f59e0b'; // yellow
      case 'none': return '#ef4444'; // red
      default: return '#6b7280'; // gray
    }
  };

  const getFollowUpInfo = () => {
    if (patient.eligible_follow_ups && patient.eligible_follow_ups.length > 0) {
      return patient.eligible_follow_ups.map((followUp, index) => (
        <div key={index} className="followup-item active">
          <div className="followup-header">
            <span className="icon">✅</span>
            <span className="status">Free Follow-up Available</span>
            <span className="days-remaining">{followUp.remaining_days} days left</span>
          </div>
          <div className="followup-details">
            <div className="doctor">👨‍⚕️ {followUp.doctor_name}</div>
            <div className="department">🏥 {followUp.department}</div>
            <div className="expiry">📅 Valid until: {followUp.next_follow_up_expiry}</div>
          </div>
        </div>
      ));
    } else if (patient.expired_followups && patient.expired_followups.length > 0) {
      return patient.expired_followups.map((expired, index) => (
        <div key={index} className="followup-item expired">
          <div className="followup-header">
            <span className="icon">⚠️</span>
            <span className="status">Follow-up Expired</span>
            <span className="expired-date">Expired: {expired.expired_on}</span>
          </div>
          <div className="followup-details">
            <div className="doctor">👨‍⚕️ {expired.doctor_name}</div>
            <div className="department">🏥 {expired.department}</div>
            <div className="note">💡 Create new regular appointment to renew</div>
          </div>
        </div>
      ));
    } else {
      return (
        <div className="followup-item none">
          <div className="followup-header">
            <span className="icon">❌</span>
            <span className="status">No Follow-up Available</span>
          </div>
          <div className="followup-details">
            <div className="note">💡 Create regular appointment to get follow-up eligibility</div>
          </div>
        </div>
      );
    }
  };

  return (
    <div className="followup-status">
      <h4>Follow-up Status</h4>
      <div className="followup-list">
        {getFollowUpInfo()}
      </div>
    </div>
  );
};

export default FollowUpStatus;
```

### **4. Search Component**

```jsx
import React, { useState, useEffect } from 'react';

const PatientSearch = ({ onSearch, onFilterChange }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedDoctor, setSelectedDoctor] = useState('');
  const [selectedDepartment, setSelectedDepartment] = useState('');
  const [doctors, setDoctors] = useState([]);
  const [departments, setDepartments] = useState([]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      onSearch(searchTerm);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm, onSearch]);

  // Load doctors and departments
  useEffect(() => {
    loadDoctors();
    loadDepartments();
  }, []);

  const loadDoctors = async () => {
    try {
      const response = await fetch('/api/v1/doctors', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      const data = await response.json();
      setDoctors(data.doctors || []);
    } catch (error) {
      console.error('Error loading doctors:', error);
    }
  };

  const loadDepartments = async () => {
    try {
      const response = await fetch('/api/v1/departments', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      const data = await response.json();
      setDepartments(data.departments || []);
    } catch (error) {
      console.error('Error loading departments:', error);
    }
  };

  const handleDoctorChange = (doctorId) => {
    setSelectedDoctor(doctorId);
    onFilterChange({ doctor_id: doctorId, department_id: selectedDepartment });
  };

  const handleDepartmentChange = (departmentId) => {
    setSelectedDepartment(departmentId);
    onFilterChange({ doctor_id: selectedDoctor, department_id: departmentId });
  };

  const clearFilters = () => {
    setSearchTerm('');
    setSelectedDoctor('');
    setSelectedDepartment('');
    onSearch('');
    onFilterChange({ doctor_id: '', department_id: '' });
  };

  return (
    <div className="patient-search">
      <div className="search-bar">
        <input
          type="text"
          placeholder="Search patients by name, phone, or MO ID..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />
        <button 
          className="clear-btn"
          onClick={clearFilters}
          disabled={!searchTerm && !selectedDoctor && !selectedDepartment}
        >
          Clear
        </button>
      </div>

      <div className="filters">
        <select
          value={selectedDoctor}
          onChange={(e) => handleDoctorChange(e.target.value)}
          className="filter-select"
        >
          <option value="">All Doctors</option>
          {doctors.map(doctor => (
            <option key={doctor.id} value={doctor.id}>
              {doctor.first_name} {doctor.last_name}
            </option>
          ))}
        </select>

        <select
          value={selectedDepartment}
          onChange={(e) => handleDepartmentChange(e.target.value)}
          className="filter-select"
        >
          <option value="">All Departments</option>
          {departments.map(dept => (
            <option key={dept.id} value={dept.id}>
              {dept.name}
            </option>
          ))}
        </select>
      </div>

      {/* Search Results Summary */}
      {(searchTerm || selectedDoctor || selectedDepartment) && (
        <div className="search-summary">
          <span>Filters applied:</span>
          {searchTerm && <span className="filter-tag">Search: "{searchTerm}"</span>}
          {selectedDoctor && (
            <span className="filter-tag">
              Doctor: {doctors.find(d => d.id === selectedDoctor)?.first_name} {doctors.find(d => d.id === selectedDoctor)?.last_name}
            </span>
          )}
          {selectedDepartment && (
            <span className="filter-tag">
              Department: {departments.find(d => d.id === selectedDepartment)?.name}
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default PatientSearch;
```

---

## 🎨 **CSS Styles**

### **Patient List Styles**

```css
.patient-list {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.filters {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.search-input {
  flex: 1;
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
}

.filter-select {
  padding: 10px 15px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  min-width: 150px;
}

.patient-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.patient-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s ease;
}

.patient-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.patient-header h3 {
  margin: 0 0 10px 0;
  color: #1f2937;
  font-size: 18px;
}

.patient-meta {
  display: flex;
  gap: 15px;
  margin-bottom: 10px;
  font-size: 14px;
  color: #6b7280;
}

.contact-info {
  display: flex;
  gap: 15px;
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 15px;
}

.followup-status {
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 15px;
}

.followup-status.active {
  background: #ecfdf5;
  border: 1px solid #10b981;
}

.followup-status.expired {
  background: #fffbeb;
  border: 1px solid #f59e0b;
}

.followup-status.none {
  background: #fef2f2;
  border: 1px solid #ef4444;
}

.status-message {
  font-weight: 600;
  margin-bottom: 8px;
}

.followup-details, .expired-details {
  font-size: 14px;
  color: #6b7280;
}

.followup-details div, .expired-details div {
  margin-bottom: 4px;
}

.medical-info {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 15px;
  font-size: 14px;
}

.medical-info div {
  margin-bottom: 4px;
}

.patient-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.btn-primary, .btn-success, .btn-secondary {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover {
  background: #059669;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background: #4b5563;
}

.loading, .no-patients {
  text-align: center;
  padding: 40px;
  color: #6b7280;
  font-size: 16px;
}
```

### **Follow-up Status Styles**

```css
.followup-status {
  margin-bottom: 20px;
}

.followup-status h4 {
  margin: 0 0 15px 0;
  color: #1f2937;
  font-size: 16px;
}

.followup-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.followup-item {
  padding: 15px;
  border-radius: 8px;
  border: 1px solid;
}

.followup-item.active {
  background: #ecfdf5;
  border-color: #10b981;
}

.followup-item.expired {
  background: #fffbeb;
  border-color: #f59e0b;
}

.followup-item.none {
  background: #fef2f2;
  border-color: #ef4444;
}

.followup-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.followup-header .icon {
  font-size: 16px;
}

.followup-header .status {
  font-weight: 600;
  flex: 1;
}

.followup-header .days-remaining,
.followup-header .expired-date {
  font-size: 14px;
  color: #6b7280;
}

.followup-details {
  font-size: 14px;
  color: #6b7280;
  margin-left: 26px;
}

.followup-details div {
  margin-bottom: 4px;
}
```

---

## 🔍 **Search Functionality**

### **Search Implementation**

```javascript
// Search patients with debouncing
const usePatientSearch = (clinicId) => {
  const [patients, setPatients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState({});

  const searchPatients = useCallback(async (term, filterOptions = {}) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        clinic_id: clinicId,
        ...(term && { search: term }),
        ...(filterOptions.doctor_id && { doctor_id: filterOptions.doctor_id }),
        ...(filterOptions.department_id && { department_id: filterOptions.department_id })
      });

      const response = await fetch(`/api/v1/clinic-specific-patients?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      const data = await response.json();
      setPatients(data.patients || []);
    } catch (error) {
      console.error('Search error:', error);
      setPatients([]);
    } finally {
      setLoading(false);
    }
  }, [clinicId]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      searchPatients(searchTerm, filters);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm, filters, searchPatients]);

  return {
    patients,
    loading,
    searchTerm,
    setSearchTerm,
    filters,
    setFilters,
    searchPatients
  };
};
```

### **Advanced Search Features**

```javascript
// Advanced search with multiple criteria
const advancedSearch = async (searchCriteria) => {
  const {
    clinicId,
    searchTerm,
    doctorId,
    departmentId,
    hasFollowUp,
    followUpStatus,
    dateRange,
    bloodGroup,
    ageRange
  } = searchCriteria;

  const params = new URLSearchParams({
    clinic_id: clinicId,
    ...(searchTerm && { search: searchTerm }),
    ...(doctorId && { doctor_id: doctorId }),
    ...(departmentId && { department_id: departmentId }),
    ...(hasFollowUp !== undefined && { has_followup: hasFollowUp }),
    ...(followUpStatus && { followup_status: followUpStatus }),
    ...(dateRange?.start && { created_after: dateRange.start }),
    ...(dateRange?.end && { created_before: dateRange.end }),
    ...(bloodGroup && { blood_group: bloodGroup }),
    ...(ageRange?.min && { age_min: ageRange.min }),
    ...(ageRange?.max && { age_max: ageRange.max })
  });

  const response = await fetch(`/api/v1/clinic-specific-patients?${params}`, {
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    }
  });

  return response.json();
};
```

---

## 📊 **Follow-up Status Indicators**

### **Status Types and Display**

```javascript
const getFollowUpStatusDisplay = (patient) => {
  const statuses = [];

  // Active follow-ups
  if (patient.eligible_follow_ups?.length > 0) {
    patient.eligible_follow_ups.forEach(followUp => {
      statuses.push({
        type: 'active',
        icon: '✅',
        color: '#10b981',
        message: `Free follow-up available (${followUp.remaining_days} days remaining)`,
        details: {
          doctor: followUp.doctor_name,
          department: followUp.department,
          expiry: followUp.next_follow_up_expiry,
          appointmentDate: followUp.appointment_date
        }
      });
    });
  }

  // Expired follow-ups
  if (patient.expired_followups?.length > 0) {
    patient.expired_followups.forEach(expired => {
      statuses.push({
        type: 'expired',
        icon: '⚠️',
        color: '#f59e0b',
        message: 'Follow-up expired - create new regular appointment to renew',
        details: {
          doctor: expired.doctor_name,
          department: expired.department,
          expiredOn: expired.expired_on
        }
      });
    });
  }

  // No follow-ups
  if (statuses.length === 0) {
    statuses.push({
      type: 'none',
      icon: '❌',
      color: '#ef4444',
      message: 'No follow-up available',
      details: {
        note: 'Create regular appointment to get follow-up eligibility'
      }
    });
  }

  return statuses;
};
```

---

## 🚨 **Error Handling**

### **Common Error Responses**

```json
{
  "error": "clinic_id is required"
}
```

```json
{
  "error": "Invalid clinic_id format"
}
```

```json
{
  "error": "Patient not found"
}
```

```json
{
  "error": "Authentication required",
  "message": "Please provide a valid authorization token in the request header",
  "code": "MISSING_TOKEN"
}
```

---

## 🔧 **Testing Examples**

### **Test Patient List API**

```bash
# Get all patients
curl -X GET "http://localhost:8081/api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Search patients
curl -X GET "http://localhost:8081/api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&search=ashiq" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Filter by doctor
curl -X GET "http://localhost:8081/api/v1/clinic-specific-patients?clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&doctor_id=ef378478-1091-472e-af40-1655e77985b3" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## ✅ **Summary**

This documentation provides complete integration guide for:

- ✅ **Patient List Display** with follow-up status
- ✅ **Search Functionality** (name, phone, MO ID)
- ✅ **Filter by Doctor/Department**
- ✅ **Follow-up Status Indicators**
- ✅ **UI Components** (React examples)
- ✅ **CSS Styling**
- ✅ **Error Handling**
- ✅ **API Testing Examples**

The system automatically populates follow-up data in patient responses, making it easy to display comprehensive patient information with follow-up status in your UI.
