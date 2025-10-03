package controllers

import (
    "organization-service/config"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "strings"
    "time"
    "fmt"
    "database/sql"
    "shared-security"
)

// ==================== STAFF MANAGEMENT ====================

type CreateStaffInput struct {
    FirstName   string   `json:"first_name" binding:"required,min=2,max=50"`
    LastName    string   `json:"last_name" binding:"required,min=2,max=50"`
    Email       *string  `json:"email" binding:"omitempty,email"`
    Username    string   `json:"username" binding:"required,min=3,max=30"`
    Phone       *string  `json:"phone" binding:"omitempty,len=10"`
    Password    string   `json:"password" binding:"required,min=8"`
    StaffType   string   `json:"staff_type" binding:"required,oneof=receptionist doctor lab_tech pharmacist billing"`
    Permissions []string `json:"permissions"`
    ClinicID    string   `json:"clinic_id" binding:"required"`
}

func CreateStaff(c *gin.Context) {
    var input CreateStaffInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Start transaction
    tx, err := config.DB.Begin()
    if err != nil {
        security.SendDatabaseError(c, "Failed to start transaction")
        return
    }
    defer tx.Rollback()

    // Check if username/email already exists
    var existingUserID string
    err = tx.QueryRow(`SELECT id FROM users WHERE username = $1 OR email = $2`, input.Username, input.Email).Scan(&existingUserID)
    if err == nil {
        security.SendValidationError(c, "User already exists", "Username or email already exists")
        return
    }

    // Hash password
    passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        security.SendDatabaseError(c, "Failed to hash password")
        return
    }

    // Create user
    var userID string
    err = tx.QueryRow(`
        INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, string(passHash)).Scan(&userID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create user")
        return
    }

    // Create staff record
    permissionsJSON := fmt.Sprintf(`["%s"]`, strings.Join(input.Permissions, `","`))
    var staffID string
    err = tx.QueryRow(`
        INSERT INTO staff (user_id, clinic_id, staff_type, permissions)
        VALUES ($1, $2, $3, $4) RETURNING id
    `, userID, input.ClinicID, input.StaffType, permissionsJSON).Scan(&staffID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create staff record")
        return
    }

    // Assign role based on staff type
    var roleID string
    err = tx.QueryRow(`SELECT id FROM roles WHERE name = $1`, input.StaffType).Scan(&roleID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to find role")
        return
    }

    _, err = tx.Exec(`
        INSERT INTO user_roles (user_id, role_id, clinic_id)
        VALUES ($1, $2, $3)
    `, userID, roleID, input.ClinicID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to assign role")
        return
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        security.SendDatabaseError(c, "Failed to commit transaction")
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "id": staffID,
        "user_id": userID,
        "message": "Staff member created successfully",
    })
}

func GetClinicStaff(c *gin.Context) {
    clinicID := c.Param("clinic_id")

    rows, err := config.DB.Query(`
        SELECT s.id, s.user_id, u.first_name, u.last_name, u.email, u.username, u.phone,
               s.staff_type, s.permissions, s.is_active, s.created_at
        FROM staff s
        JOIN users u ON s.user_id = u.id
        WHERE s.clinic_id = $1
        ORDER BY s.created_at DESC
    `, clinicID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch staff")
        return
    }
    defer rows.Close()

    var staff []map[string]interface{}
    for rows.Next() {
        var s map[string]interface{} = make(map[string]interface{})
        var id, userID, firstName, lastName, email, username, phone, staffType, permissionsJSON string
        var isActive bool
        var createdAt string
        
        err := rows.Scan(&id, &userID, &firstName, &lastName, &email, 
                        &username, &phone, &staffType, &permissionsJSON, 
                        &isActive, &createdAt)
        if err != nil {
            continue
        }
        
        s["id"] = id
        s["user_id"] = userID
        s["first_name"] = firstName
        s["last_name"] = lastName
        s["email"] = email
        s["username"] = username
        s["phone"] = phone
        s["staff_type"] = staffType
        s["permissions"] = permissionsJSON
        s["is_active"] = isActive
        s["created_at"] = createdAt
        
        staff = append(staff, s)
    }

    c.JSON(http.StatusOK, staff)
}

type UpdateStaffRoleInput struct {
    StaffType   string   `json:"staff_type" binding:"required,oneof=receptionist doctor lab_tech pharmacist billing"`
    Permissions []string `json:"permissions"`
}

func UpdateStaffRole(c *gin.Context) {
    clinicID := c.Param("clinic_id")
    userID := c.Param("user_id")
    
    var input UpdateStaffRoleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    permissionsJSON := fmt.Sprintf(`["%s"]`, strings.Join(input.Permissions, `","`))
    
    result, err := config.DB.Exec(`
        UPDATE staff SET staff_type = $1, permissions = $2, updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $3 AND clinic_id = $4
    `, input.StaffType, permissionsJSON, userID, clinicID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to update staff role")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "staff member")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Staff role updated successfully"})
}

func DeactivateStaff(c *gin.Context) {
    clinicID := c.Param("clinic_id")
    userID := c.Param("user_id")
    
    result, err := config.DB.Exec(`
        UPDATE staff SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $1 AND clinic_id = $2
    `, userID, clinicID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to deactivate staff")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "staff member")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Staff member deactivated successfully"})
}

// ==================== QUEUE MANAGEMENT ====================

type CreateQueueInput struct {
    ClinicID  string  `json:"clinic_id" binding:"required"`
    QueueType string  `json:"queue_type" binding:"required,oneof=doctor lab pharmacy"`
    DoctorID  *string `json:"doctor_id"` // required for doctor queues
}

func CreateQueue(c *gin.Context) {
    var input CreateQueueInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    if input.QueueType == "doctor" && input.DoctorID == nil {
        security.SendValidationError(c, "Doctor ID required", "Doctor ID is required for doctor queues")
        return
    }

    var queueID string
    err := config.DB.QueryRow(`
        INSERT INTO queues (clinic_id, queue_type, doctor_id)
        VALUES ($1, $2, $3) RETURNING id
    `, input.ClinicID, input.QueueType, input.DoctorID).Scan(&queueID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create queue")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": queueID, "message": "Queue created successfully"})
}

func GetQueues(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    queueType := c.Query("queue_type")

    query := `
        SELECT q.id, q.clinic_id, q.queue_type, q.doctor_id, q.is_active, q.is_paused, 
               q.current_token, q.created_at,
               COALESCE(d.doctor_code, '') as doctor_code,
               COALESCE(u.first_name || ' ' || u.last_name, '') as doctor_name
        FROM queues q
        LEFT JOIN doctors d ON q.doctor_id = d.id
        LEFT JOIN users u ON d.user_id = u.id
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND q.clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }

    if queueType != "" {
        query += fmt.Sprintf(" AND q.queue_type = $%d", argIndex)
        args = append(args, queueType)
        argIndex++
    }

    query += " ORDER BY q.created_at DESC"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch queues")
        return
    }
    defer rows.Close()

    var queues []map[string]interface{}
    for rows.Next() {
        var id, clinicID, queueType, doctorID string
        var isActive, isPaused bool
        var currentToken sql.NullInt32
        var createdAt time.Time
        var doctorCode, doctorName sql.NullString
        
        err := rows.Scan(&id, &clinicID, &queueType, &doctorID, 
                        &isActive, &isPaused, &currentToken, &createdAt,
                        &doctorCode, &doctorName)
        if err != nil {
            continue
        }
        
        q := map[string]interface{}{
            "id": id,
            "clinic_id": clinicID,
            "queue_type": queueType,
            "doctor_id": doctorID,
            "is_active": isActive,
            "is_paused": isPaused,
            "current_token": currentToken,
            "created_at": createdAt,
            "doctor_code": doctorCode,
            "doctor_name": doctorName,
        }
        queues = append(queues, q)
    }

    c.JSON(http.StatusOK, queues)
}

type AssignTokenInput struct {
    QueueID       string `json:"queue_id" binding:"required"`
    PatientID     string `json:"patient_id" binding:"required"`
    AppointmentID string `json:"appointment_id" binding:"required"`
    Priority      bool   `json:"priority"`
}

func AssignToken(c *gin.Context) {
    var input AssignTokenInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Get next token number
    var tokenNumber int
    err := config.DB.QueryRow(`
        SELECT COALESCE(MAX(token_number), 0) + 1 FROM queue_tokens WHERE queue_id = $1
    `, input.QueueID).Scan(&tokenNumber)
    if err != nil {
        security.SendDatabaseError(c, "Failed to get token number")
        return
    }

    var tokenID string
    err = config.DB.QueryRow(`
        INSERT INTO queue_tokens (queue_id, patient_id, appointment_id, token_number, priority)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `, input.QueueID, input.PatientID, input.AppointmentID, tokenNumber, input.Priority).Scan(&tokenID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to assign token")
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "id": tokenID,
        "token_number": tokenNumber,
        "message": "Token assigned successfully",
    })
}

func ReassignToken(c *gin.Context) {
    tokenID := c.Param("token_id")
    newQueueID := c.PostForm("queue_id")

    if newQueueID == "" {
        security.SendValidationError(c, "Queue ID required", "New queue ID is required")
        return
    }

    result, err := config.DB.Exec(`
        UPDATE queue_tokens SET queue_id = $1 WHERE id = $2
    `, newQueueID, tokenID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to reassign token")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "token")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Token reassigned successfully"})
}

func PauseQueue(c *gin.Context) {
    queueID := c.Param("queue_id")
    
    result, err := config.DB.Exec(`
        UPDATE queues SET is_paused = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = $1
    `, queueID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to pause queue")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "queue")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Queue paused successfully"})
}

func ResumeQueue(c *gin.Context) {
    queueID := c.Param("queue_id")
    
    result, err := config.DB.Exec(`
        UPDATE queues SET is_paused = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = $1
    `, queueID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to resume queue")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "queue")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Queue resumed successfully"})
}

// ==================== PHARMACY MANAGEMENT ====================

type CreateMedicineInput struct {
    ClinicID      string  `json:"clinic_id" binding:"required"`
    MedicineName  string  `json:"medicine_name" binding:"required,min=2,max=255"`
    GenericName   *string `json:"generic_name"`
    MedicineCode  string  `json:"medicine_code" binding:"required,min=2,max=50"`
    Category      *string `json:"category"`
    Unit          string  `json:"unit" binding:"required,min=1,max=20"`
    CurrentStock  int     `json:"current_stock" binding:"min=0"`
    MinStockLevel int     `json:"min_stock_level" binding:"min=0"`
    MaxStockLevel int     `json:"max_stock_level" binding:"min=0"`
    UnitPrice     float64 `json:"unit_price" binding:"required,min=0"`
    ExpiryDate    *string `json:"expiry_date"` // YYYY-MM-DD format
    SupplierName  *string `json:"supplier_name"`
    BatchNumber   *string `json:"batch_number"`
}

func CreateMedicine(c *gin.Context) {
    var input CreateMedicineInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var medicineID string
    err := config.DB.QueryRow(`
        INSERT INTO pharmacy_inventory (clinic_id, medicine_name, generic_name, medicine_code, 
                                       category, unit, current_stock, min_stock_level, max_stock_level,
                                       unit_price, expiry_date, supplier_name, batch_number)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id
    `, input.ClinicID, input.MedicineName, input.GenericName, input.MedicineCode, input.Category,
       input.Unit, input.CurrentStock, input.MinStockLevel, input.MaxStockLevel, input.UnitPrice,
       input.ExpiryDate, input.SupplierName, input.BatchNumber).Scan(&medicineID)
    if err != nil {
        if strings.Contains(err.Error(), "unique constraint") {
            security.SendValidationError(c, "Medicine code already exists", "Medicine code must be unique")
            return
        }
        security.SendDatabaseError(c, "Failed to create medicine")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": medicineID, "message": "Medicine created successfully"})
}

func GetPharmacyInventory(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    category := c.Query("category")
    lowStock := c.Query("low_stock") == "true"
    expired := c.Query("expired") == "true"

    query := `
        SELECT id, clinic_id, medicine_name, generic_name, medicine_code, category, unit,
               current_stock, min_stock_level, max_stock_level, unit_price, expiry_date,
               supplier_name, batch_number, is_active, created_at, updated_at
        FROM pharmacy_inventory
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }

    if category != "" {
        query += fmt.Sprintf(" AND category = $%d", argIndex)
        args = append(args, category)
        argIndex++
    }

    if lowStock {
        query += " AND current_stock <= min_stock_level"
    }

    if expired {
        query += " AND expiry_date < CURRENT_DATE"
    }

    query += " ORDER BY medicine_name"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch inventory")
        return
    }
    defer rows.Close()

    var inventory []map[string]interface{}
    for rows.Next() {
        var id, clinicID, medicineName, genericName, medicineCode, category, unit string
        var currentStock, minStockLevel, maxStockLevel sql.NullInt32
        var unitPrice sql.NullFloat64
        var expiryDate sql.NullTime
        var supplierName, batchNumber sql.NullString
        var isActive bool
        var createdAt, updatedAt time.Time
        
        err := rows.Scan(&id, &clinicID, &medicineName, &genericName,
                        &medicineCode, &category, &unit, &currentStock,
                        &minStockLevel, &maxStockLevel, &unitPrice,
                        &expiryDate, &supplierName, &batchNumber,
                        &isActive, &createdAt, &updatedAt)
        if err != nil {
            continue
        }
        
        item := map[string]interface{}{
            "id": id,
            "clinic_id": clinicID,
            "medicine_name": medicineName,
            "generic_name": genericName,
            "medicine_code": medicineCode,
            "category": category,
            "unit": unit,
            "current_stock": currentStock,
            "min_stock_level": minStockLevel,
            "max_stock_level": maxStockLevel,
            "unit_price": unitPrice,
            "expiry_date": expiryDate,
            "supplier_name": supplierName,
            "batch_number": batchNumber,
            "is_active": isActive,
            "created_at": createdAt,
            "updated_at": updatedAt,
        }
        inventory = append(inventory, item)
    }

    c.JSON(http.StatusOK, inventory)
}

type UpdateStockInput struct {
    CurrentStock int `json:"current_stock" binding:"required,min=0"`
}

func UpdateMedicineStock(c *gin.Context) {
    medicineID := c.Param("medicine_id")
    
    var input UpdateStockInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    result, err := config.DB.Exec(`
        UPDATE pharmacy_inventory SET current_stock = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, input.CurrentStock, medicineID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to update stock")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "medicine")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}

type CreatePharmacyDiscountInput struct {
    ClinicID            string  `json:"clinic_id" binding:"required"`
    DiscountName        string  `json:"discount_name" binding:"required,min=2,max=255"`
    DiscountType        string  `json:"discount_type" binding:"required,oneof=percentage fixed_amount"`
    DiscountValue       float64 `json:"discount_value" binding:"required,min=0"`
    MinPurchaseAmount   float64 `json:"min_purchase_amount" binding:"min=0"`
    MaxDiscountAmount   *float64 `json:"max_discount_amount"`
    ValidFrom           string  `json:"valid_from" binding:"required"` // YYYY-MM-DD
    ValidTo             string  `json:"valid_to" binding:"required"`   // YYYY-MM-DD
}

func CreatePharmacyDiscount(c *gin.Context) {
    var input CreatePharmacyDiscountInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var discountID string
    err := config.DB.QueryRow(`
        INSERT INTO pharmacy_discounts (clinic_id, discount_name, discount_type, discount_value,
                                       min_purchase_amount, max_discount_amount, valid_from, valid_to)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
    `, input.ClinicID, input.DiscountName, input.DiscountType, input.DiscountValue,
       input.MinPurchaseAmount, input.MaxDiscountAmount, input.ValidFrom, input.ValidTo).Scan(&discountID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create discount")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": discountID, "message": "Pharmacy discount created successfully"})
}

// ==================== LAB MANAGEMENT ====================

type CreateLabTestInput struct {
    ClinicID                string  `json:"clinic_id" binding:"required"`
    TestCode                string  `json:"test_code" binding:"required,min=2,max=50"`
    TestName                string  `json:"test_name" binding:"required,min=2,max=255"`
    TestCategory            *string `json:"test_category"`
    Description             *string `json:"description"`
    SampleType              *string `json:"sample_type"`
    PreparationInstructions *string `json:"preparation_instructions"`
    NormalRange             *string `json:"normal_range"`
    Unit                    *string `json:"unit"`
    Price                   float64 `json:"price" binding:"required,min=0"`
    TurnaroundTimeHours     int     `json:"turnaround_time_hours" binding:"min=1"`
}

func CreateLabTest(c *gin.Context) {
    var input CreateLabTestInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var testID string
    err := config.DB.QueryRow(`
        INSERT INTO lab_tests (clinic_id, test_code, test_name, test_category, description,
                              sample_type, preparation_instructions, normal_range, unit, price, turnaround_time_hours)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id
    `, input.ClinicID, input.TestCode, input.TestName, input.TestCategory, input.Description,
       input.SampleType, input.PreparationInstructions, input.NormalRange, input.Unit, input.Price, input.TurnaroundTimeHours).Scan(&testID)
    if err != nil {
        if strings.Contains(err.Error(), "unique constraint") {
            security.SendValidationError(c, "Test code already exists", "Test code must be unique")
            return
        }
        security.SendDatabaseError(c, "Failed to create lab test")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": testID, "message": "Lab test created successfully"})
}

func GetLabTests(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    category := c.Query("category")

    query := `
        SELECT id, clinic_id, test_code, test_name, test_category, description, sample_type,
               preparation_instructions, normal_range, unit, price, turnaround_time_hours,
               is_active, created_at, updated_at
        FROM lab_tests
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }

    if category != "" {
        query += fmt.Sprintf(" AND test_category = $%d", argIndex)
        args = append(args, category)
        argIndex++
    }

    query += " ORDER BY test_name"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch lab tests")
        return
    }
    defer rows.Close()

    var tests []map[string]interface{}
    for rows.Next() {
        var id, clinicID, testCode, testName, testCategory, description, sampleType, preparationInstructions, normalRange, unit string
        var price sql.NullFloat64
        var turnaroundTimeHours sql.NullInt32
        var isActive bool
        var createdAt, updatedAt time.Time
        
        err := rows.Scan(&id, &clinicID, &testCode, &testName,
                        &testCategory, &description, &sampleType,
                        &preparationInstructions, &normalRange, &unit,
                        &price, &turnaroundTimeHours, &isActive,
                        &createdAt, &updatedAt)
        if err != nil {
            continue
        }
        
        test := map[string]interface{}{
            "id": id,
            "clinic_id": clinicID,
            "test_code": testCode,
            "test_name": testName,
            "test_category": testCategory,
            "description": description,
            "sample_type": sampleType,
            "preparation_instructions": preparationInstructions,
            "normal_range": normalRange,
            "unit": unit,
            "price": price,
            "turnaround_time_hours": turnaroundTimeHours,
            "is_active": isActive,
            "created_at": createdAt,
            "updated_at": updatedAt,
        }
        tests = append(tests, test)
    }

    c.JSON(http.StatusOK, tests)
}

type CreateSampleCollectorInput struct {
    UserID         string  `json:"user_id" binding:"required"`
    ClinicID       string  `json:"clinic_id" binding:"required"`
    CollectorCode  *string `json:"collector_code"`
    Specialization *string `json:"specialization"`
}

func CreateSampleCollector(c *gin.Context) {
    var input CreateSampleCollectorInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var collectorID string
    err := config.DB.QueryRow(`
        INSERT INTO lab_sample_collectors (user_id, clinic_id, collector_code, specialization)
        VALUES ($1, $2, $3, $4) RETURNING id
    `, input.UserID, input.ClinicID, input.CollectorCode, input.Specialization).Scan(&collectorID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create sample collector")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": collectorID, "message": "Sample collector created successfully"})
}

type UploadLabResultInput struct {
    OrderID      string  `json:"order_id" binding:"required"`
    TestID       string  `json:"test_id" binding:"required"`
    ResultValue  *string `json:"result_value"`
    ResultUnit   *string `json:"result_unit"`
    NormalRange  *string `json:"normal_range"`
    Status       string  `json:"status" binding:"required,oneof=normal abnormal critical"`
    Notes        *string `json:"notes"`
    VisibleToPatient bool `json:"visible_to_patient"`
}

func UploadLabResult(c *gin.Context) {
    var input UploadLabResultInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("user_id")
    if !exists {
        security.SendError(c, http.StatusUnauthorized, "USER_NOT_AUTHENTICATED", "User not authenticated", "User authentication is required to access this resource", nil)
        return
    }

    var resultID string
    err := config.DB.QueryRow(`
        INSERT INTO lab_results (order_id, test_id, result_value, result_unit, normal_range,
                                status, notes, uploaded_by, is_visible_to_patient)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
    `, input.OrderID, input.TestID, input.ResultValue, input.ResultUnit, input.NormalRange,
       input.Status, input.Notes, userID, input.VisibleToPatient).Scan(&resultID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to upload lab result")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": resultID, "message": "Lab result uploaded successfully"})
}

// ==================== INSURANCE PROVIDER MANAGEMENT ====================

type CreateInsuranceProviderInput struct {
    ClinicID            string  `json:"clinic_id" binding:"required"`
    ProviderName        string  `json:"provider_name" binding:"required,min=2,max=255"`
    ProviderCode        *string `json:"provider_code"`
    ContactDetails      map[string]interface{} `json:"contact_details"`
    ConsultationCovered bool    `json:"consultation_covered"`
    MedicinesCovered    bool    `json:"medicines_covered"`
    LabTestsCovered     bool    `json:"lab_tests_covered"`
    CoveragePercentage  float64 `json:"coverage_percentage" binding:"min=0,max=100"`
    MaxCoverageAmount   *float64 `json:"max_coverage_amount"`
}

func CreateInsuranceProvider(c *gin.Context) {
    var input CreateInsuranceProviderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var providerID string
    err := config.DB.QueryRow(`
        INSERT INTO insurance_providers (clinic_id, provider_name, provider_code, contact_details,
                                        consultation_covered, medicines_covered, lab_tests_covered,
                                        coverage_percentage, max_coverage_amount)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
    `, input.ClinicID, input.ProviderName, input.ProviderCode, input.ContactDetails,
       input.ConsultationCovered, input.MedicinesCovered, input.LabTestsCovered,
       input.CoveragePercentage, input.MaxCoverageAmount).Scan(&providerID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create insurance provider")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": providerID, "message": "Insurance provider created successfully"})
}

func GetInsuranceProviders(c *gin.Context) {
    clinicID := c.Query("clinic_id")

    query := `
        SELECT id, clinic_id, provider_name, provider_code, contact_details,
               consultation_covered, medicines_covered, lab_tests_covered,
               coverage_percentage, max_coverage_amount, is_active, created_at, updated_at
        FROM insurance_providers
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }

    query += " ORDER BY provider_name"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch insurance providers")
        return
    }
    defer rows.Close()

    var providers []map[string]interface{}
    for rows.Next() {
        var id, clinicID, providerName, providerCode string
        var contactDetailsJSON string
        var consultationCovered, medicinesCovered, labTestsCovered bool
        var coveragePercentage sql.NullFloat64
        var maxCoverageAmount sql.NullFloat64
        var isActive bool
        var createdAt, updatedAt time.Time
        
        err := rows.Scan(&id, &clinicID, &providerName,
                        &providerCode, &contactDetailsJSON, &consultationCovered,
                        &medicinesCovered, &labTestsCovered,
                        &coveragePercentage, &maxCoverageAmount,
                        &isActive, &createdAt, &updatedAt)
        if err != nil {
            continue
        }
        
        provider := map[string]interface{}{
            "id": id,
            "clinic_id": clinicID,
            "provider_name": providerName,
            "provider_code": providerCode,
            "contact_details": contactDetailsJSON,
            "consultation_covered": consultationCovered,
            "medicines_covered": medicinesCovered,
            "lab_tests_covered": labTestsCovered,
            "coverage_percentage": coveragePercentage,
            "max_coverage_amount": maxCoverageAmount,
            "is_active": isActive,
            "created_at": createdAt,
            "updated_at": updatedAt,
        }
        providers = append(providers, provider)
    }

    c.JSON(http.StatusOK, providers)
}

// ==================== REPORTS & ANALYTICS ====================

func GetDailyStats(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    dateStr := c.Query("date") // YYYY-MM-DD format
    
    if clinicID == "" || dateStr == "" {
        security.SendValidationError(c, "Missing parameters", "clinic_id and date are required")
        return
    }

    var totalPatients, newPatients, totalAppointments, completedAppointments, cancelledAppointments sql.NullInt32
    var totalRevenue, consultationRevenue, labRevenue, pharmacyRevenue sql.NullFloat64
    var avgWaitTimeMinutes sql.NullFloat64
    
    err := config.DB.QueryRow(`
        SELECT total_patients, new_patients, total_appointments, completed_appointments,
               cancelled_appointments, total_revenue, consultation_revenue, lab_revenue,
               pharmacy_revenue, avg_wait_time_minutes
        FROM analytics_daily_stats
        WHERE clinic_id = $1 AND stat_date = $2
    `, clinicID, dateStr).Scan(&totalPatients, &newPatients, 
                               &totalAppointments, &completedAppointments,
                               &cancelledAppointments, &totalRevenue,
                               &consultationRevenue, &labRevenue,
                               &pharmacyRevenue, &avgWaitTimeMinutes)
    if err != nil {
        if err == sql.ErrNoRows {
            // Return zero stats if no data found
            stats := map[string]interface{}{
                "total_patients": 0, "new_patients": 0, "total_appointments": 0,
                "completed_appointments": 0, "cancelled_appointments": 0,
                "total_revenue": 0, "consultation_revenue": 0, "lab_revenue": 0,
                "pharmacy_revenue": 0, "avg_wait_time_minutes": 0,
            }
            c.JSON(http.StatusOK, stats)
            return
        } else {
            security.SendDatabaseError(c, "Failed to fetch daily stats")
            return
        }
    }
    
    stats := map[string]interface{}{
        "total_patients": totalPatients,
        "new_patients": newPatients,
        "total_appointments": totalAppointments,
        "completed_appointments": completedAppointments,
        "cancelled_appointments": cancelledAppointments,
        "total_revenue": totalRevenue,
        "consultation_revenue": consultationRevenue,
        "lab_revenue": labRevenue,
        "pharmacy_revenue": pharmacyRevenue,
        "avg_wait_time_minutes": avgWaitTimeMinutes,
    }

    c.JSON(http.StatusOK, stats)
}

func GetDoctorStats(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    dateStr := c.Query("date")
    
    if clinicID == "" || dateStr == "" {
        security.SendValidationError(c, "Missing parameters", "clinic_id and date are required")
        return
    }

    rows, err := config.DB.Query(`
        SELECT ds.doctor_id, d.doctor_code, u.first_name || ' ' || u.last_name as doctor_name,
               ds.total_appointments, ds.completed_appointments, ds.avg_consultation_time_minutes,
               ds.total_revenue, ds.patient_satisfaction_score
        FROM analytics_doctor_stats ds
        JOIN doctors d ON ds.doctor_id = d.id
        JOIN users u ON d.user_id = u.id
        WHERE ds.clinic_id = $1 AND ds.stat_date = $2
        ORDER BY ds.total_revenue DESC
    `, clinicID, dateStr)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch doctor stats")
        return
    }
    defer rows.Close()

    var doctorStats []map[string]interface{}
    for rows.Next() {
        var stats map[string]interface{} = make(map[string]interface{})
        var doctorID, doctorCode, doctorName string
        var totalAppointments, completedAppointments int
        var avgConsultationTime, totalRevenue, patientSatisfactionScore float64
        
        err := rows.Scan(&doctorID, &doctorCode, &doctorName,
                        &totalAppointments, &completedAppointments,
                        &avgConsultationTime, &totalRevenue,
                        &patientSatisfactionScore)
        if err != nil {
            continue
        }
        
        stats["doctor_id"] = doctorID
        stats["doctor_code"] = doctorCode
        stats["doctor_name"] = doctorName
        stats["total_appointments"] = totalAppointments
        stats["completed_appointments"] = completedAppointments
        stats["avg_consultation_time_minutes"] = avgConsultationTime
        stats["total_revenue"] = totalRevenue
        stats["patient_satisfaction_score"] = patientSatisfactionScore
        
        doctorStats = append(doctorStats, stats)
    }

    c.JSON(http.StatusOK, doctorStats)
}

func GetFinancialReport(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    fromDate := c.Query("from_date") // YYYY-MM-DD
    toDate := c.Query("to_date")     // YYYY-MM-DD
    
    if clinicID == "" || fromDate == "" || toDate == "" {
        security.SendValidationError(c, "Missing parameters", "clinic_id, from_date, and to_date are required")
        return
    }

    rows, err := config.DB.Query(`
        SELECT collection_date, consultation_amount, lab_amount, pharmacy_amount,
               procedure_amount, total_amount, cash_amount, card_amount, insurance_amount,
               outstanding_amount
        FROM daily_collections
        WHERE clinic_id = $1 AND collection_date BETWEEN $2 AND $3
        ORDER BY collection_date DESC
    `, clinicID, fromDate, toDate)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch financial report")
        return
    }
    defer rows.Close()

    var collections []map[string]interface{}
    var totalRevenue, totalCash, totalCard, totalInsurance, totalOutstanding float64

    for rows.Next() {
        var collection map[string]interface{} = make(map[string]interface{})
        var collectionDate string
        var consultation, lab, pharmacy, procedure, total, cash, card, insurance, outstanding float64
        
        err := rows.Scan(&collectionDate, &consultation, &lab, &pharmacy,
                        &procedure, &total, &cash, &card, &insurance, &outstanding)
        if err != nil {
            continue
        }
        
        collection["collection_date"] = collectionDate
        
        collection["consultation_amount"] = consultation
        collection["lab_amount"] = lab
        collection["pharmacy_amount"] = pharmacy
        collection["procedure_amount"] = procedure
        collection["total_amount"] = total
        collection["cash_amount"] = cash
        collection["card_amount"] = card
        collection["insurance_amount"] = insurance
        collection["outstanding_amount"] = outstanding
        
        collections = append(collections, collection)
        
        totalRevenue += total
        totalCash += cash
        totalCard += card
        totalInsurance += insurance
        totalOutstanding += outstanding
    }

    summary := map[string]interface{}{
        "total_revenue": totalRevenue,
        "total_cash": totalCash,
        "total_card": totalCard,
        "total_insurance": totalInsurance,
        "total_outstanding": totalOutstanding,
        "collections": collections,
    }

    c.JSON(http.StatusOK, summary)
}

// ==================== PATIENT MANAGEMENT (ADMIN) ====================

type MergePatientsInput struct {
    PrimaryPatientID   string `json:"primary_patient_id" binding:"required"`
    DuplicatePatientID string `json:"duplicate_patient_id" binding:"required"`
}

func MergePatients(c *gin.Context) {
    var input MergePatientsInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Start transaction
    tx, err := config.DB.Begin()
    if err != nil {
        security.SendDatabaseError(c, "Failed to start transaction")
        return
    }
    defer tx.Rollback()

    // Verify both patients exist
    var primaryExists, duplicateExists bool
    err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM patients WHERE id = $1)`, input.PrimaryPatientID).Scan(&primaryExists)
    if err != nil || !primaryExists {
        security.SendNotFoundError(c, "primary patient")
        return
    }

    err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM patients WHERE id = $1)`, input.DuplicatePatientID).Scan(&duplicateExists)
    if err != nil || !duplicateExists {
        security.SendNotFoundError(c, "duplicate patient")
        return
    }

    // Update all appointments to point to primary patient
    _, err = tx.Exec(`
        UPDATE appointments SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge appointments")
        return
    }

    // Update all patient clinic assignments
    _, err = tx.Exec(`
        UPDATE patient_clinics SET patient_id = $1 
        WHERE patient_id = $2 AND clinic_id NOT IN (
            SELECT clinic_id FROM patient_clinics WHERE patient_id = $1
        )
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge clinic assignments")
        return
    }

    // Update all queue tokens
    _, err = tx.Exec(`
        UPDATE queue_tokens SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge queue tokens")
        return
    }

    // Update all lab orders
    _, err = tx.Exec(`
        UPDATE lab_orders SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge lab orders")
        return
    }

    // Update all pharmacy billing
    _, err = tx.Exec(`
        UPDATE pharmacy_billing SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge pharmacy billing")
        return
    }

    // Update all insurance records
    _, err = tx.Exec(`
        UPDATE patient_insurance SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge insurance records")
        return
    }

    // Update all insurance claims
    _, err = tx.Exec(`
        UPDATE insurance_claims SET patient_id = $1 WHERE patient_id = $2
    `, input.PrimaryPatientID, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to merge insurance claims")
        return
    }

    // Delete remaining clinic assignments for duplicate patient
    _, err = tx.Exec(`
        DELETE FROM patient_clinics WHERE patient_id = $1
    `, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to clean up duplicate clinic assignments")
        return
    }

    // Deactivate duplicate patient
    _, err = tx.Exec(`
        UPDATE patients SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP 
        WHERE id = $1
    `, input.DuplicatePatientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to deactivate duplicate patient")
        return
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        security.SendDatabaseError(c, "Failed to commit merge operation")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Patients merged successfully"})
}

func GetPatientHistory(c *gin.Context) {
    patientID := c.Param("patient_id")

    // Get patient basic info
    var patientInfo map[string]interface{} = make(map[string]interface{})
    var id, userID, moID string
    var medicalHistory, allergies, bloodGroup sql.NullString
    var firstName, lastName, email, phone string
    var dateOfBirth sql.NullTime
    var gender sql.NullString
    
    err := config.DB.QueryRow(`
        SELECT p.id, p.user_id, p.mo_id, p.medical_history, p.allergies, p.blood_group,
               u.first_name, u.last_name, u.email, u.phone, u.date_of_birth, u.gender
        FROM patients p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = $1
    `, patientID).Scan(&id, &userID, &moID,
                       &medicalHistory, &allergies, &bloodGroup,
                       &firstName, &lastName, &email,
                       &phone, &dateOfBirth, &gender)
    if err != nil {
        security.SendNotFoundError(c, "patient")
        return
    }
    
    patientInfo["id"] = id
    patientInfo["user_id"] = userID
    patientInfo["mo_id"] = moID
    patientInfo["medical_history"] = medicalHistory.String
    patientInfo["allergies"] = allergies.String
    patientInfo["blood_group"] = bloodGroup.String
    patientInfo["first_name"] = firstName
    patientInfo["last_name"] = lastName
    patientInfo["email"] = email
    patientInfo["phone"] = phone
    patientInfo["date_of_birth"] = dateOfBirth.Time
    patientInfo["gender"] = gender.String
    if err != nil {
        security.SendNotFoundError(c, "patient")
        return
    }

    // Get appointment history
    appointmentRows, err := config.DB.Query(`
        SELECT a.id, a.booking_number, a.appointment_time, a.consultation_type, a.status,
               a.fee_amount, a.payment_status, c.name as clinic_name,
               u.first_name || ' ' || u.last_name as doctor_name
        FROM appointments a
        JOIN clinics c ON a.clinic_id = c.id
        JOIN doctors d ON a.doctor_id = d.id
        JOIN users u ON d.user_id = u.id
        WHERE a.patient_id = $1
        ORDER BY a.appointment_time DESC
        LIMIT 50
    `, patientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch appointment history")
        return
    }
    defer appointmentRows.Close()

    var appointments []map[string]interface{}
    for appointmentRows.Next() {
        var id, bookingNumber, appointmentTime, consultationType, status, feeAmount, paymentStatus, clinicName, doctorName interface{}
        err := appointmentRows.Scan(&id, &bookingNumber, &appointmentTime, &consultationType,
                                   &status, &feeAmount, &paymentStatus, &clinicName, &doctorName)
        if err != nil {
            continue
        }
        appointment := map[string]interface{}{
            "id": id,
            "booking_number": bookingNumber,
            "appointment_time": appointmentTime,
            "consultation_type": consultationType,
            "status": status,
            "fee_amount": feeAmount,
            "payment_status": paymentStatus,
            "clinic_name": clinicName,
            "doctor_name": doctorName,
        }
        appointments = append(appointments, appointment)
    }

    // Get lab order history
    labRows, err := config.DB.Query(`
        SELECT lo.id, lo.order_number, lo.order_date, lo.status, lo.total_amount,
               lo.payment_status, c.name as clinic_name
        FROM lab_orders lo
        JOIN clinics c ON lo.clinic_id = c.id
        WHERE lo.patient_id = $1
        ORDER BY lo.order_date DESC
        LIMIT 20
    `, patientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch lab history")
        return
    }
    defer labRows.Close()

    var labOrders []map[string]interface{}
    for labRows.Next() {
        var id, orderNumber, orderDate, status, totalAmount, paymentStatus, clinicName interface{}
        err := labRows.Scan(&id, &orderNumber, &orderDate, &status, &totalAmount, &paymentStatus, &clinicName)
        if err != nil {
            continue
        }
        order := map[string]interface{}{
            "id": id,
            "order_number": orderNumber,
            "order_date": orderDate,
            "status": status,
            "total_amount": totalAmount,
            "payment_status": paymentStatus,
            "clinic_name": clinicName,
        }
        labOrders = append(labOrders, order)
    }

    history := map[string]interface{}{
        "patient_info": patientInfo,
        "appointments": appointments,
        "lab_orders":   labOrders,
    }

    c.JSON(http.StatusOK, history)
}

// ==================== BILLING & FEE MANAGEMENT ====================

type CreateFeeStructureInput struct {
    ClinicID     string   `json:"clinic_id" binding:"required"`
    ServiceType  string   `json:"service_type" binding:"required,oneof=consultation lab pharmacy procedure"`
    ServiceName  string   `json:"service_name" binding:"required,min=2,max=255"`
    BaseFee      float64  `json:"base_fee" binding:"required,min=0"`
    FollowUpFee  *float64 `json:"follow_up_fee"`
    FollowUpDays *int     `json:"follow_up_days"`
}

func CreateFeeStructure(c *gin.Context) {
    var input CreateFeeStructureInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var feeID string
    err := config.DB.QueryRow(`
        INSERT INTO fee_structures (clinic_id, service_type, service_name, base_fee, follow_up_fee, follow_up_days)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `, input.ClinicID, input.ServiceType, input.ServiceName, input.BaseFee, input.FollowUpFee, input.FollowUpDays).Scan(&feeID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create fee structure")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": feeID, "message": "Fee structure created successfully"})
}

func GetFeeStructures(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    serviceType := c.Query("service_type")

    query := `
        SELECT id, clinic_id, service_type, service_name, base_fee, follow_up_fee,
               follow_up_days, is_active, created_at, updated_at
        FROM fee_structures
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }

    if serviceType != "" {
        query += fmt.Sprintf(" AND service_type = $%d", argIndex)
        args = append(args, serviceType)
        argIndex++
    }

    query += " ORDER BY service_type, service_name"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch fee structures")
        return
    }
    defer rows.Close()

    var feeStructures []map[string]interface{}
    for rows.Next() {
        var id, clinicId, serviceType, serviceName, baseFee, followUpFee, followUpDays, isActive, createdAt, updatedAt interface{}
        err := rows.Scan(&id, &clinicId, &serviceType, &serviceName, &baseFee, &followUpFee, &followUpDays, &isActive, &createdAt, &updatedAt)
        if err != nil {
            continue
        }
        fee := map[string]interface{}{
            "id": id,
            "clinic_id": clinicId,
            "service_type": serviceType,
            "service_name": serviceName,
            "base_fee": baseFee,
            "follow_up_fee": followUpFee,
            "follow_up_days": followUpDays,
            "is_active": isActive,
            "created_at": createdAt,
            "updated_at": updatedAt,
        }
        feeStructures = append(feeStructures, fee)
    }

    c.JSON(http.StatusOK, feeStructures)
}

type CreateBillingDiscountInput struct {
    ClinicID           string   `json:"clinic_id" binding:"required"`
    DiscountName       string   `json:"discount_name" binding:"required,min=2,max=255"`
    DiscountType       string   `json:"discount_type" binding:"required,oneof=percentage fixed_amount"`
    DiscountValue      float64  `json:"discount_value" binding:"required,min=0"`
    ApplicableServices []string `json:"applicable_services"`
    MinAmount          float64  `json:"min_amount" binding:"min=0"`
    MaxDiscountAmount  *float64 `json:"max_discount_amount"`
    ValidFrom          string   `json:"valid_from" binding:"required"` // YYYY-MM-DD
    ValidTo            string   `json:"valid_to" binding:"required"`   // YYYY-MM-DD
}

func CreateBillingDiscount(c *gin.Context) {
    var input CreateBillingDiscountInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var discountID string
    err := config.DB.QueryRow(`
        INSERT INTO billing_discounts (clinic_id, discount_name, discount_type, discount_value,
                                      applicable_services, min_amount, max_discount_amount, valid_from, valid_to)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
    `, input.ClinicID, input.DiscountName, input.DiscountType, input.DiscountValue,
       input.ApplicableServices, input.MinAmount, input.MaxDiscountAmount, input.ValidFrom, input.ValidTo).Scan(&discountID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create billing discount")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": discountID, "message": "Billing discount created successfully"})
}
