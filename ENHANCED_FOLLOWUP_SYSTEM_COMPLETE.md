# Enhanced Follow-Up System - Complete Implementation 📋

## 🎯 **Your Requirements**

> "The API should list all doctors along with their departments and saved follow-up records. Each patient can see multiple doctors, so every doctor should have their own independent follow-up condition. The API should list all follow-ups per doctor."

**Status:** ✅ **IMPLEMENTING NOW!**

---

## 🔄 **Enhanced API Design**

### **New API Endpoint: Patient Follow-Up Status by Doctor**

```http
GET /api/organizations/patient-followup-status/{patient_id}
```

**Response Structure:**
```json
{
  "patient": {
    "id": "patient-uuid",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
  },
  "doctors": [
    {
      "doctor_id": "doctor-uuid",
      "doctor_name": "Dr. Smith",
      "departments": [
        {
          "department_id": "dept-uuid",
          "department_name": "Cardiology",
          "follow_up_status": {
            "status": true,
            "is_free": true,
            "status_label": "free",
            "color_code": "green",
            "message": "Free follow-up available (3 days remaining)",
            "days_remaining": 3,
            "last_appointment_date": "2025-10-20",
            "follow_up_expiry": "2025-10-25",
            "free_follow_up_used": false
          },
          "appointment_history": [
            {
              "appointment_id": "appointment-uuid",
              "appointment_date": "2025-10-20",
              "appointment_type": "clinic_visit",
              "status": "completed",
              "days_since": 2
            }
          ]
        }
      ]
    }
  ]
}
```

---

## 🔧 **Implementation**

### **1. Enhanced Organization Service Controller**

```go
// services/organization-service/controllers/clinic_patient.controller.go

// PatientFollowUpStatusResponse represents follow-up status for all doctors
type PatientFollowUpStatusResponse struct {
	Patient PatientInfo                    `json:"patient"`
	Doctors []DoctorFollowUpStatus        `json:"doctors"`
}

type PatientInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     *string `json:"email,omitempty"`
}

type DoctorFollowUpStatus struct {
	DoctorID    string                `json:"doctor_id"`
	DoctorName  string                `json:"doctor_name"`
	Departments []DepartmentFollowUp  `json:"departments"`
}

type DepartmentFollowUp struct {
	DepartmentID     string                `json:"department_id"`
	DepartmentName   string                `json:"department_name"`
	FollowUpStatus   FollowUpStatus        `json:"follow_up_status"`
	AppointmentHistory []AppointmentHistoryItem `json:"appointment_history"`
}

type FollowUpStatus struct {
	Status              bool    `json:"status"`               // true = free, false = paid
	IsFree              bool    `json:"is_free"`              // same as status
	StatusLabel         string  `json:"status_label"`         // "free", "paid", "none"
	ColorCode           string  `json:"color_code"`           // "green", "orange", "gray"
	Message             string  `json:"message"`
	DaysRemaining       *int    `json:"days_remaining,omitempty"`
	LastAppointmentDate *string `json:"last_appointment_date,omitempty"`
	FollowUpExpiry      *string `json:"follow_up_expiry,omitempty"`
	FreeFollowUpUsed    bool    `json:"free_follow_up_used"`
}

// GetPatientFollowUpStatus - Get follow-up status for all doctors and departments
// GET /api/organizations/patient-followup-status/:patient_id
func GetPatientFollowUpStatus(c *gin.Context) {
	patientID := c.Param("patient_id")

	if _, err := uuid.Parse(patientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid patient_id format",
		})
		return
	}

	// Get patient basic info
	var patient PatientInfo
	err := config.DB.QueryRow(`
		SELECT id, first_name, last_name, phone, email
		FROM clinic_patients
		WHERE id = $1
	`, patientID).Scan(
		&patient.ID, &patient.FirstName, &patient.LastName, &patient.Phone, &patient.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Patient not found",
			})
			return
		}
		security.SendDatabaseError(c, "Failed to fetch patient")
		return
	}

	// Get all doctors and their departments with follow-up status
	doctors, err := getPatientFollowUpStatusByDoctors(patientID, config.DB)
	if err != nil {
		security.SendDatabaseError(c, "Failed to fetch follow-up status")
		return
	}

	response := PatientFollowUpStatusResponse{
		Patient: patient,
		Doctors: doctors,
	}

	c.JSON(http.StatusOK, response)
}

// getPatientFollowUpStatusByDoctors gets follow-up status for all doctors
func getPatientFollowUpStatusByDoctors(patientID string, db *sql.DB) ([]DoctorFollowUpStatus, error) {
	// Get all doctors who have appointments with this patient
	query := `
		SELECT DISTINCT 
			d.id as doctor_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN users u ON u.id = d.user_id
		WHERE a.clinic_patient_id = $1
		  AND a.status IN ('completed', 'confirmed')
		ORDER BY doctor_name
	`

	rows, err := db.Query(query, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []DoctorFollowUpStatus
	for rows.Next() {
		var doctor DoctorFollowUpStatus
		err := rows.Scan(&doctor.DoctorID, &doctor.DoctorName)
		if err != nil {
			continue
		}

		// Get departments for this doctor
		departments, err := getDoctorDepartmentsFollowUpStatus(patientID, doctor.DoctorID, db)
		if err != nil {
			continue
		}

		doctor.Departments = departments
		doctors = append(doctors, doctor)
	}

	return doctors, nil
}

// getDoctorDepartmentsFollowUpStatus gets follow-up status for all departments of a doctor
func getDoctorDepartmentsFollowUpStatus(patientID, doctorID string, db *sql.DB) ([]DepartmentFollowUp, error) {
	// Get all departments for this doctor (including NULL for general)
	query := `
		SELECT DISTINCT 
			COALESCE(dept.id, '') as department_id,
			COALESCE(dept.name, 'General') as department_name
		FROM appointments a
		LEFT JOIN departments dept ON dept.id = a.department_id
		WHERE a.clinic_patient_id = $1
		  AND a.doctor_id = $2
		  AND a.status IN ('completed', 'confirmed')
		ORDER BY department_name
	`

	rows, err := db.Query(query, patientID, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []DepartmentFollowUp
	for rows.Next() {
		var dept DepartmentFollowUp
		err := rows.Scan(&dept.DepartmentID, &dept.DepartmentName)
		if err != nil {
			continue
		}

		// Get follow-up status for this doctor+department combination
		followUpStatus, err := getFollowUpStatusForDoctorDepartment(patientID, doctorID, dept.DepartmentID, db)
		if err != nil {
			continue
		}

		dept.FollowUpStatus = followUpStatus

		// Get appointment history for this doctor+department
		appointmentHistory, err := getAppointmentHistoryForDoctorDepartment(patientID, doctorID, dept.DepartmentID, db)
		if err != nil {
			continue
		}

		dept.AppointmentHistory = appointmentHistory
		departments = append(departments, dept)
	}

	return departments, nil
}

// getFollowUpStatusForDoctorDepartment gets follow-up status for specific doctor+department
func getFollowUpStatusForDoctorDepartment(patientID, doctorID, departmentID string, db *sql.DB) (FollowUpStatus, error) {
	var status FollowUpStatus

	// Check if there's an active follow-up in follow_ups table
	query := `
		SELECT 
			is_free,
			valid_until,
			valid_from
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND doctor_id = $2
		  AND status = 'active'
		  AND valid_until >= CURRENT_DATE
	`

	args := []interface{}{patientID, doctorID}
	if departmentID != "" {
		query += ` AND (department_id = $3 OR (department_id IS NULL AND $3 = ''))`
		args = append(args, departmentID)
	} else {
		query += ` AND (department_id IS NULL OR department_id = '')`
	}

	query += ` ORDER BY created_at DESC LIMIT 1`

	var isFree bool
	var validUntil, validFrom time.Time
	err := db.QueryRow(query, args...).Scan(&isFree, &validUntil, &validFrom)

	if err == nil {
		// Active follow-up found
		daysRemaining := int(time.Until(validUntil).Hours() / 24)
		
		status.Status = isFree
		status.IsFree = isFree
		status.StatusLabel = "free"
		status.ColorCode = "green"
		status.Message = fmt.Sprintf("Free follow-up available (%d days remaining)", daysRemaining)
		status.DaysRemaining = &daysRemaining
		
		validFromStr := validFrom.Format("2006-01-02")
		status.LastAppointmentDate = &validFromStr
		
		validUntilStr := validUntil.Format("2006-01-02")
		status.FollowUpExpiry = &validUntilStr
		
		status.FreeFollowUpUsed = false
		
		return status, nil
	}

	// No active follow-up - check if patient has previous appointment
	checkQuery := `
		SELECT 
			appointment_date,
			COUNT(*) as total_appointments
		FROM appointments
		WHERE clinic_patient_id = $1
		  AND doctor_id = $2
		  AND consultation_type IN ('clinic_visit', 'video_consultation')
		  AND status IN ('completed', 'confirmed')
	`

	checkArgs := []interface{}{patientID, doctorID}
	if departmentID != "" {
		checkQuery += ` AND (department_id = $3 OR (department_id IS NULL AND $3 = ''))`
		checkArgs = append(checkArgs, departmentID)
	} else {
		checkQuery += ` AND (department_id IS NULL OR department_id = '')`
	}

	checkQuery += ` GROUP BY appointment_date ORDER BY appointment_date DESC LIMIT 1`

	var lastAppointmentDate time.Time
	var totalAppointments int
	err = db.QueryRow(checkQuery, checkArgs...).Scan(&lastAppointmentDate, &totalAppointments)

	if err == nil && totalAppointments > 0 {
		// Has previous appointment - check if follow-up period expired
		daysSince := int(time.Since(lastAppointmentDate).Hours() / 24)
		
		if daysSince <= 5 {
			// Still within follow-up period but no active follow-up record
			// This means free follow-up was used or not created
			status.Status = false
			status.IsFree = false
			status.StatusLabel = "paid"
			status.ColorCode = "orange"
			status.Message = "Follow-up available (payment required)"
			status.FreeFollowUpUsed = true
			
			lastAppointmentStr := lastAppointmentDate.Format("2006-01-02")
			status.LastAppointmentDate = &lastAppointmentStr
			
			expiryDate := lastAppointmentDate.AddDate(0, 0, 5)
			expiryStr := expiryDate.Format("2006-01-02")
			status.FollowUpExpiry = &expiryStr
			
			return status, nil
		} else {
			// Follow-up period expired
			status.Status = false
			status.IsFree = false
			status.StatusLabel = "paid"
			status.ColorCode = "orange"
			status.Message = "Follow-up period expired (payment required)"
			status.FreeFollowUpUsed = false
			
			lastAppointmentStr := lastAppointmentDate.Format("2006-01-02")
			status.LastAppointmentDate = &lastAppointmentStr
			
			expiryDate := lastAppointmentDate.AddDate(0, 0, 5)
			expiryStr := expiryDate.Format("2006-01-02")
			status.FollowUpExpiry = &expiryStr
			
			return status, nil
		}
	}

	// No previous appointment
	status.Status = false
	status.IsFree = false
	status.StatusLabel = "none"
	status.ColorCode = "gray"
	status.Message = "No previous appointment with this doctor and department"
	status.FreeFollowUpUsed = false

	return status, nil
}

// getAppointmentHistoryForDoctorDepartment gets appointment history for specific doctor+department
func getAppointmentHistoryForDoctorDepartment(patientID, doctorID, departmentID string, db *sql.DB) ([]AppointmentHistoryItem, error) {
	query := `
		SELECT 
			a.id,
			a.doctor_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			a.department_id,
			dept.name as department,
			a.consultation_type,
			a.appointment_date,
			a.status,
			a.payment_status,
			a.created_at
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = a.department_id
		WHERE a.clinic_patient_id = $1
		  AND a.doctor_id = $2
		  AND a.status IN ('completed', 'confirmed')
	`

	args := []interface{}{patientID, doctorID}
	if departmentID != "" {
		query += ` AND (a.department_id = $3 OR (a.department_id IS NULL AND $3 = ''))`
		args = append(args, departmentID)
	} else {
		query += ` AND (a.department_id IS NULL OR a.department_id = '')`
	}

	query += ` ORDER BY a.appointment_date DESC, a.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []AppointmentHistoryItem
	for rows.Next() {
		var item AppointmentHistoryItem
		var appointmentDate time.Time
		var createdAt time.Time

		err := rows.Scan(
			&item.ID,
			&item.DoctorID,
			&item.DoctorName,
			&item.DepartmentID,
			&item.Department,
			&item.AppointmentType,
			&appointmentDate,
			&item.Status,
			&item.PaymentStatus,
			&createdAt,
		)
		if err != nil {
			continue
		}

		item.AppointmentDate = appointmentDate.Format("2006-01-02")
		item.DaysSince = int(time.Since(appointmentDate).Hours() / 24)
		item.ValidityDays = 5
		item.FollowUpEligible = true
		item.FollowUpStatus = "active"
		item.RenewalStatus = "valid"
		item.FreeFollowUpUsed = false
		item.Note = fmt.Sprintf("Appointment with %s (%s)", item.DoctorName, item.Department)

		history = append(history, item)
	}

	return history, nil
}
```

---

### **2. Enhanced Patient List API**

```go
// Enhanced ListClinicPatients with doctor-specific follow-up status
func ListClinicPatientsEnhanced(c *gin.Context) {
	clinicID := c.Query("clinic_id")
	search := c.Query("search")
	onlyActive := c.DefaultQuery("only_active", "true")

	if clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "clinic_id is required",
		})
		return
	}

	if _, err := uuid.Parse(clinicID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clinic_id format",
		})
		return
	}

	// Build query
	query := `
		SELECT id, clinic_id, first_name, last_name, phone, email, date_of_birth, age, gender,
		       address1, address2, district, state, mo_id, medical_history, allergies, 
		       blood_group, smoking_status, alcohol_use, height_cm, weight_kg, 
		       is_active, created_at, updated_at
		FROM clinic_patients
		WHERE clinic_id = $1
	`
	args := []interface{}{clinicID}
	argIndex := 2

	if onlyActive == "true" {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	if search != "" {
		query += fmt.Sprintf(` AND (
			LOWER(first_name) LIKE LOWER($%d) OR 
			LOWER(last_name) LIKE LOWER($%d) OR 
			LOWER(phone) LIKE LOWER($%d) OR 
			LOWER(mo_id) LIKE LOWER($%d) OR
			LOWER(address1) LIKE LOWER($%d) OR
			LOWER(district) LIKE LOWER($%d) OR
			LOWER(state) LIKE LOWER($%d)
		)`, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY created_at DESC"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		security.SendDatabaseError(c, "Failed to fetch patients")
		return
	}
	defer rows.Close()

	var patients []ClinicPatientResponse
	for rows.Next() {
		var patient ClinicPatientResponse
		err := rows.Scan(
			&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
			&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
			&patient.Address1, &patient.Address2, &patient.District, &patient.State,
			&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
			&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
			&patient.IsActive, &patient.CreatedAt, &patient.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Get follow-up status for all doctors
		doctors, err := getPatientFollowUpStatusByDoctors(patient.ID, config.DB)
		if err != nil {
			continue
		}

		// Set overall follow-up eligibility based on any doctor having free follow-up
		hasFreeFollowUp := false
		for _, doctor := range doctors {
			for _, dept := range doctor.Departments {
				if dept.FollowUpStatus.Status {
					hasFreeFollowUp = true
					break
				}
			}
			if hasFreeFollowUp {
				break
			}
		}

		// Set patient follow-up eligibility
		if hasFreeFollowUp {
			patient.FollowUpEligibility = &FollowUpEligibility{
				Eligible:      true,
				IsFree:        true,
				StatusLabel:   "free",
				ColorCode:     "green",
				Message:       "Free follow-up available with some doctors",
				DaysRemaining: nil,
			}
		} else {
			patient.FollowUpEligibility = &FollowUpEligibility{
				Eligible:      true,
				IsFree:        false,
				StatusLabel:   "paid",
				ColorCode:     "orange",
				Message:       "Follow-up available (payment required)",
				DaysRemaining: nil,
			}
		}

		patients = append(patients, patient)
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_id": clinicID,
		"total":     len(patients),
		"patients":  patients,
	})
}
```

---

### **3. Enhanced Appointment Creation**

```go
// Enhanced CreateSimpleAppointment with automatic follow-up creation
func CreateSimpleAppointmentEnhanced(c *gin.Context) {
	var input CreateSimpleAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"details": err.Error(),
		})
		return
	}

	// ... existing validation code ...

	// Create appointment
	var appointment models.Appointment
	var appointmentDate time.Time
	var isFreeFollowUp bool = false

	// ... existing appointment creation code ...

	// Enhanced follow-up creation logic
	if input.ConsultationType == "clinic_visit" || input.ConsultationType == "video_consultation" {
		// This is a regular appointment - create follow-up eligibility
		err = followUpManager.CreateFollowUp(
			input.ClinicPatientID,
			input.ClinicID,
			input.DoctorID,
			input.DepartmentID,
			appointment.ID,
			appointmentDate,
		)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to create follow-up record: %v", err)
		} else {
			log.Printf("✅ Created follow-up eligibility for patient %s with doctor %s", 
				input.ClinicPatientID, input.DoctorID)
		}
	}

	// Enhanced follow-up validation
	if input.IsFollowUp {
		// Check follow-up eligibility using the enhanced system
		followUpStatus, err := getFollowUpStatusForDoctorDepartment(
			input.ClinicPatientID,
			input.DoctorID,
			func() string {
				if input.DepartmentID != nil {
					return *input.DepartmentID
				}
				return ""
			}(),
			config.DB,
		)
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check follow-up eligibility",
				"details": err.Error(),
			})
			return
		}

		if !followUpStatus.Status {
			// Paid follow-up - require payment
			if input.PaymentMethod == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Payment method required",
					"message": "This follow-up requires payment",
				})
				return
			}
			isFreeFollowUp = false
		} else {
			// Free follow-up
			isFreeFollowUp = true
			
			// Mark follow-up as used
			err = followUpManager.MarkFollowUpAsUsed(
				input.ClinicPatientID,
				input.ClinicID,
				input.DoctorID,
				input.DepartmentID,
				appointment.ID,
			)
			if err != nil {
				log.Printf("⚠️ Warning: Failed to mark follow-up as used: %v", err)
			}
		}
	}

	// Enhanced response
	response := gin.H{
		"message":     "Appointment created successfully",
		"appointment": appointment,
	}

	if input.IsFollowUp {
		response["is_free_followup"] = isFreeFollowUp
		response["followup_type"] = func() string {
			if isFreeFollowUp {
				return "free"
			}
			return "paid"
		}()
		response["followup_message"] = func() string {
			if isFreeFollowUp {
				return "This is a FREE follow-up"
			}
			return "This is a PAID follow-up"
		}()
	} else {
		response["is_regular_appointment"] = true
		response["followup_granted"] = true
		response["followup_message"] = "Free follow-up eligibility granted (valid for 5 days)"
		
		expiryDate := appointmentDate.AddDate(0, 0, 5)
		response["followup_valid_until"] = expiryDate.Format("2006-01-02")
	}

	c.JSON(http.StatusCreated, response)
}
```

---

### **4. Routes Configuration**

```go
// services/organization-service/routes/organization.routes.go

func SetupOrganizationRoutes(router *gin.Engine) {
	api := router.Group("/api/organizations")
	{
		// Existing routes
		api.GET("/clinic-specific-patients", ListClinicPatients)
		api.GET("/clinic-specific-patients/:id", GetClinicPatient)
		
		// Enhanced routes
		api.GET("/clinic-specific-patients-enhanced", ListClinicPatientsEnhanced)
		api.GET("/patient-followup-status/:patient_id", GetPatientFollowUpStatus)
	}
}
```

---

### **5. Flutter API Integration**

```dart
// Enhanced Flutter API Service
class EnhancedFollowUpApiService {
  // Get patient follow-up status for all doctors
  static Future<Map<String, dynamic>> getPatientFollowUpStatus({
    required String patientId,
    required String token,
  }) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/organizations/patient-followup-status/$patientId'),
        headers: getHeaders(token),
      );
      
      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load follow-up status: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching follow-up status: $e');
    }
  }

  // Enhanced patient list with follow-up status
  static Future<Map<String, dynamic>> getPatientsWithFollowUpStatus({
    required String clinicId,
    String? search,
    required String token,
  }) async {
    try {
      String url = '$baseUrl/api/organizations/clinic-specific-patients-enhanced?clinic_id=$clinicId';
      
      if (search != null && search.isNotEmpty) url += '&search=$search';
      
      final response = await http.get(
        Uri.parse(url),
        headers: getHeaders(token),
      );
      
      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load patients: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching patients: $e');
    }
  }
}
```

---

### **6. Enhanced Flutter UI**

```dart
// Enhanced Patient Card with Doctor-Specific Follow-Up Status
class EnhancedPatientCard extends StatelessWidget {
  final Patient patient;
  final VoidCallback onViewFollowUpStatus;
  final VoidCallback onBookFollowUp;
  final VoidCallback onBookRegular;

  const EnhancedPatientCard({
    Key? key,
    required this.patient,
    required this.onViewFollowUpStatus,
    required this.onBookFollowUp,
    required this.onBookRegular,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Patient Header
            Row(
              children: [
                CircleAvatar(
                  backgroundColor: _getStatusColor(patient.followUpEligibility.colorCode),
                  radius: 24,
                  child: Text(
                    '${patient.firstName[0]}${patient.lastName[0]}',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                  ),
                ),
                SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${patient.firstName} ${patient.lastName}',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        patient.phone,
                        style: TextStyle(fontSize: 14, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  onPressed: onViewFollowUpStatus,
                  icon: Icon(Icons.info_outline),
                  tooltip: 'View Follow-Up Status',
                ),
              ],
            ),
            
            SizedBox(height: 12),
            
            // Follow-Up Status Summary
            Container(
              padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: _getStatusColor(patient.followUpEligibility.colorCode).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: _getStatusColor(patient.followUpEligibility.colorCode),
                  width: 1,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    _getStatusIcon(patient.followUpEligibility.statusLabel).icon,
                    color: _getStatusColor(patient.followUpEligibility.colorCode),
                    size: 20,
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      patient.followUpEligibility.message,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: _getStatusColor(patient.followUpEligibility.colorCode),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            
            SizedBox(height: 12),
            
            // Action Buttons
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: onBookFollowUp,
                    icon: Icon(Icons.refresh, size: 18),
                    label: Text('Follow-Up'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _getStatusColor(patient.followUpEligibility.colorCode),
                      foregroundColor: Colors.white,
                      padding: EdgeInsets.symmetric(vertical: 12),
                    ),
                  ),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: onBookRegular,
                    icon: Icon(Icons.calendar_today, size: 18),
                    label: Text('Regular'),
                    style: OutlinedButton.styleFrom(
                      padding: EdgeInsets.symmetric(vertical: 12),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Color _getStatusColor(String colorCode) {
    switch (colorCode) {
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'gray':
      default:
        return Colors.grey;
    }
  }

  Icon _getStatusIcon(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return Icon(Icons.check_circle, color: Colors.green, size: 24);
      case 'paid':
        return Icon(Icons.payment, color: Colors.orange, size: 24);
      case 'none':
        return Icon(Icons.info_outline, color: Colors.grey, size: 24);
      default:
        return Icon(Icons.help_outline, color: Colors.grey, size: 24);
    }
  }
}

// Follow-Up Status Detail Dialog
class FollowUpStatusDialog extends StatefulWidget {
  final String patientId;
  final String token;

  const FollowUpStatusDialog({
    Key? key,
    required this.patientId,
    required this.token,
  }) : super(key: key);

  @override
  _FollowUpStatusDialogState createState() => _FollowUpStatusDialogState();
}

class _FollowUpStatusDialogState extends State<FollowUpStatusDialog> {
  Map<String, dynamic>? followUpData;
  bool isLoading = true;
  String? errorMessage;

  @override
  void initState() {
    super.initState();
    loadFollowUpStatus();
  }

  Future<void> loadFollowUpStatus() async {
    try {
      final data = await EnhancedFollowUpApiService.getPatientFollowUpStatus(
        patientId: widget.patientId,
        token: widget.token,
      );
      
      setState(() {
        followUpData = data;
        isLoading = false;
      });
    } catch (e) {
      setState(() {
        errorMessage = e.toString();
        isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Follow-Up Status'),
      content: SizedBox(
        width: double.maxFinite,
        height: 400,
        child: isLoading
            ? Center(child: CircularProgressIndicator())
            : errorMessage != null
                ? Center(child: Text('Error: $errorMessage'))
                : followUpData == null
                    ? Center(child: Text('No data available'))
                    : ListView.builder(
                        itemCount: followUpData!['doctors'].length,
                        itemBuilder: (context, index) {
                          final doctor = followUpData!['doctors'][index];
                          return DoctorFollowUpCard(doctor: doctor);
                        },
                      ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text('Close'),
        ),
      ],
    );
  }
}

// Doctor Follow-Up Card
class DoctorFollowUpCard extends StatelessWidget {
  final Map<String, dynamic> doctor;

  const DoctorFollowUpCard({Key? key, required this.doctor}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              doctor['doctor_name'],
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 8),
            ...(doctor['departments'] as List).map((dept) {
              return DepartmentFollowUpTile(department: dept);
            }).toList(),
          ],
        ),
      ),
    );
  }
}

// Department Follow-Up Tile
class DepartmentFollowUpTile extends StatelessWidget {
  final Map<String, dynamic> department;

  const DepartmentFollowUpTile({Key? key, required this.department}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final status = department['follow_up_status'];
    
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: _getStatusColor(status['color_code']),
        radius: 16,
        child: Icon(
          _getStatusIcon(status['status_label']).icon,
          color: Colors.white,
          size: 16,
        ),
      ),
      title: Text(department['department_name']),
      subtitle: Text(status['message']),
      trailing: status['status'] == true
          ? Chip(
              label: Text('FREE'),
              backgroundColor: Colors.green[100],
              labelStyle: TextStyle(color: Colors.green[800]),
            )
          : Chip(
              label: Text('PAID'),
              backgroundColor: Colors.orange[100],
              labelStyle: TextStyle(color: Colors.orange[800]),
            ),
    );
  }

  Color _getStatusColor(String colorCode) {
    switch (colorCode) {
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'gray':
      default:
        return Colors.grey;
    }
  }

  Icon _getStatusIcon(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return Icon(Icons.check_circle, color: Colors.green);
      case 'paid':
        return Icon(Icons.payment, color: Colors.orange);
      case 'none':
        return Icon(Icons.info_outline, color: Colors.grey);
      default:
        return Icon(Icons.help_outline, color: Colors.grey);
    }
  }
}
```

---

## 🚀 **Deployment**

### **1. Update Routes**

```go
// services/organization-service/routes/organization.routes.go
func SetupOrganizationRoutes(router *gin.Engine) {
	api := router.Group("/api/organizations")
	{
		// Existing routes
		api.GET("/clinic-specific-patients", ListClinicPatients)
		api.GET("/clinic-specific-patients/:id", GetClinicPatient)
		
		// New enhanced routes
		api.GET("/clinic-specific-patients-enhanced", ListClinicPatientsEnhanced)
		api.GET("/patient-followup-status/:patient_id", GetPatientFollowUpStatus)
	}
}
```

### **2. Build and Deploy**

```bash
# Build organization service
docker-compose build organization-service

# Deploy
docker-compose up -d organization-service

# Check logs
docker-compose logs organization-service --tail=50
```

---

## ✅ **Summary**

**Enhanced Features:**

1. **✅ Doctor-Specific Follow-Up Status** - Each doctor has independent follow-up conditions
2. **✅ Department-Specific Follow-Up** - Each department within a doctor has its own status
3. **✅ Complete Follow-Up History** - Shows all appointments and follow-up records
4. **✅ Enhanced Patient List** - Shows overall follow-up status with doctor details
5. **✅ Detailed Follow-Up Dialog** - Shows all doctors and their follow-up status
6. **✅ Automatic Follow-Up Creation** - Regular appointments automatically create follow-up eligibility
7. **✅ Enhanced Appointment Booking** - Automatically applies correct follow-up status

**API Endpoints:**
- `GET /api/organizations/patient-followup-status/{patient_id}` - Get all doctor follow-up status
- `GET /api/organizations/clinic-specific-patients-enhanced` - Enhanced patient list
- `POST /api/appointments/simple` - Enhanced appointment creation

**Ready to implement!** 🚀✅

**This enhanced system provides complete follow-up management per doctor and department!** 📋✨

