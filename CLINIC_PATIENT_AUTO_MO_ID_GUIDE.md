# Automatic MO ID Generation for Clinic Patients

## Overview
When a new clinic patient is created, the system automatically assigns a unique "MO ID" (Medical Officer ID) based on the clinic's code. This ensures proper patient identification and tracking within each clinic.

## Feature Details

### Auto-Generation Logic
1. **Format**: `{clinic_code}{sequential_number}`
   - Example: If clinic code is "MO", patients get: MO0001, MO0002, MO0003, etc.
   - Example: If clinic code is "ABC", patients get: ABC0001, ABC0002, ABC0003, etc.

2. **Sequential Numbering**: 
   - Starts from 0001 for the first patient
   - Automatically increments for each new patient
   - Pads with zeros to maintain 4-digit format

3. **Unique Per Clinic**:
   - Each clinic maintains its own sequence
   - The same MO ID can exist in different clinics
   - Enforced by unique constraint: `(clinic_id, mo_id)`

### Implementation

#### When MO ID is Auto-Generated
- If `mo_id` is **not provided** or is **empty** in the request
- System fetches the clinic's `clinic_code`
- Finds the highest sequential number in existing MO IDs for that clinic
- Generates the next sequential MO ID

#### When MO ID is Manually Provided
- If `mo_id` **is provided** in the request
- System validates it doesn't already exist for that clinic
- Uses the provided MO ID (custom MO IDs are allowed)

### Database Structure
```sql
-- clinic_patients table
CREATE TABLE clinic_patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE NOT NULL,
    mo_id VARCHAR(50),  -- Auto-generated or manual
    
    -- Other fields...
    
    -- Unique constraint: mo_id per clinic
    CONSTRAINT unique_mo_id_per_clinic UNIQUE (clinic_id, mo_id)
);

-- Index for performance
CREATE INDEX idx_clinic_patients_mo_id ON clinic_patients(mo_id);
```

## API Usage

### Create Patient WITHOUT Mo ID (Auto-Generated)
```bash
POST /clinic-specific-patients
Content-Type: application/json

{
  "clinic_id": "abc-123-uuid",
  "first_name": "Safeer",
  "last_name": "Ameen",
  "phone": "+91 476738737",
  "email": "safeer@example.com",
  "age": 36,
  "gender": "Male"
  // mo_id is NOT provided - will be auto-generated
}
```

**Response:**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "patient-uuid",
    "clinic_id": "abc-123-uuid",
    "mo_id": "MO0001",  // ✅ Auto-generated
    "first_name": "Safeer",
    "last_name": "Ameen",
    "phone": "+91 476738737",
    "email": "safeer@example.com",
    "age": 36,
    "gender": "Male",
    "is_active": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### Create Patient WITH Custom Mo ID (Manual)
```bash
POST /clinic-specific-patients
Content-Type: application/json

{
  "clinic_id": "abc-123-uuid",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+91 9876543210",
  "mo_id": "CUSTOM123",  // ✅ Manually provided
  "age": 45,
  "gender": "Male"
}
```

**Response:**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "patient-uuid-2",
    "clinic_id": "abc-123-uuid",
    "mo_id": "CUSTOM123",  // ✅ Uses provided MO ID
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+91 9876543210",
    "age": 45,
    "gender": "Male",
    "is_active": true,
    "created_at": "2024-01-01T11:00:00Z",
    "updated_at": "2024-01-01T11:00:00Z"
  }
}
```

## Example Scenarios

### Scenario 1: Clinic "MO" Creating Patients
```
Clinic Code: "MO"
Existing Patients: None

1st Patient (no mo_id provided) → MO0001
2nd Patient (no mo_id provided) → MO0002
3rd Patient (mo_id: "MOXYZ") → MOXYZ  (custom)
4th Patient (no mo_id provided) → MO0003  (continues sequence)
5th Patient (no mo_id provided) → MO0004
```

### Scenario 2: Clinic "ABC" Creating Patients
```
Clinic Code: "ABC"
Existing Patients: None

1st Patient (no mo_id provided) → ABC0001
2nd Patient (no mo_id provided) → ABC0002
3rd Patient (no mo_id provided) → ABC0003
```

### Scenario 3: Multiple Clinics (Isolated Sequences)
```
Clinic "MO":  MO0001, MO0002, MO0003
Clinic "ABC": ABC0001, ABC0002, ABC0003

✅ Each clinic maintains its own sequence
✅ MO0001 and ABC0001 are different patients in different clinics
```

## Technical Implementation

### Code Location
- **File**: `services/organization-service/controllers/clinic_patient.controller.go`
- **Function**: `CreateClinicPatient`

### Algorithm
```go
// 1. Check if mo_id is provided
if input.MOID == nil || *input.MOID == "" {
    // 2. Get clinic's clinic_code
    clinicCode := getClinicCode(input.ClinicID)
    
    // 3. Find highest sequential number
    maxNumber := findMaxSequentialNumber(clinicCode, input.ClinicID)
    
    // 4. Generate next MO ID
    generatedMOID = fmt.Sprintf("%s%04d", clinicCode, maxNumber+1)
}

// 5. Validate uniqueness
validateMOIDUnique(input.ClinicID, moID)

// 6. Create patient with MO ID
createPatient(patientData)
```

### SQL Query for Finding Max Number
```sql
SELECT COALESCE(MAX(
    CASE 
        WHEN mo_id ~ '^{clinic_code}[0-9]+$' 
        THEN CAST(SUBSTRING(mo_id FROM LENGTH($1) + 1) AS INTEGER)
        ELSE 0
    END
), 0) as max_num
FROM clinic_patients 
WHERE clinic_id = $2
```

**Explanation:**
- Uses regex to match MO IDs following the pattern: `{clinic_code}{numbers}`
- Extracts the numeric part
- Finds the maximum number
- Returns 0 if no matching MO IDs exist

## Validation & Error Handling

### Validation Rules
1. **Uniqueness**: MO ID must be unique within the clinic
2. **Format**: No format restrictions (allows auto-generated or custom)
3. **Length**: Maximum 50 characters

### Error Responses

#### Duplicate MO ID
```json
{
  "error": "Mo ID exists in this clinic",
  "message": "A patient with this Mo ID already exists in your clinic"
}
```

#### Database Error
```json
{
  "error": "Failed to generate MO ID",
  "message": "Database error occurred"
}
```

## Benefits

### For Clinics
✅ **Automatic Numbering**: No manual tracking needed
✅ **Consistency**: Standardized format across all patients
✅ **Flexibility**: Can override with custom IDs when needed
✅ **Scalability**: Supports up to 9999 patients with 4-digit format (expandable)

### For Patients
✅ **Unique Identity**: Each patient gets a unique clinic-specific ID
✅ **Easy Reference**: Simple format for verbal communication
✅ **Privacy**: No sensitive information in the ID

### For System
✅ **Data Integrity**: Database-level uniqueness constraint
✅ **Performance**: Indexed for fast lookups
✅ **Isolation**: Each clinic's sequence is independent

## Migration Path

### For Existing Patients (Without MO ID)
If you have existing patients without MO IDs, you can run a migration:

```sql
-- Update existing patients with auto-generated MO IDs
UPDATE clinic_patients cp
SET mo_id = (
    SELECT c.clinic_code || LPAD(
        ROW_NUMBER() OVER (PARTITION BY cp.clinic_id ORDER BY cp.created_at)::TEXT,
        4, '0'
    )
    FROM clinics c
    WHERE c.id = cp.clinic_id
)
WHERE mo_id IS NULL;
```

## Testing

### Test Case 1: Auto-Generated MO ID
```bash
# Create patient without mo_id
curl -X POST http://localhost:8080/clinic-specific-patients \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+91 1234567890"
  }'

# Expected: mo_id = "MO0001" (or {clinic_code}0001)
```

### Test Case 2: Custom MO ID
```bash
# Create patient with custom mo_id
curl -X POST http://localhost:8080/clinic-specific-patients \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+91 9876543210",
    "mo_id": "CUSTOM001"
  }'

# Expected: mo_id = "CUSTOM001"
```

### Test Case 3: Duplicate MO ID Error
```bash
# Try creating patient with existing mo_id
curl -X POST http://localhost:8080/clinic-specific-patients \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "first_name": "Bob",
    "last_name": "Brown",
    "phone": "+91 5555555555",
    "mo_id": "MO0001"
  }'

# Expected: 409 Conflict - "Mo ID exists in this clinic"
```

## Notes

1. **Thread Safety**: The current implementation may have race conditions in high-concurrency scenarios. Consider using database transactions or locks for production.

2. **Number Format**: The default format uses 4 digits (0001-9999). If you need more patients, you can modify the format in the code:
   ```go
   // Change %04d to %05d for 5 digits (00001-99999)
   generatedMOID = fmt.Sprintf("%s%05d", clinicCode, maxNumber+1)
   ```

3. **Performance**: The MO ID generation query is optimized with indexes. For very large patient databases (100K+ patients), consider caching or pre-generating ID blocks.

4. **Clinic Code Changes**: If a clinic's code changes, existing MO IDs remain unchanged (by design). New patients will use the new clinic code.

## Future Enhancements

### Possible Improvements
1. **Configurable Format**: Allow clinics to configure their own MO ID format
2. **ID Blocks**: Pre-allocate blocks of IDs for better performance
3. **Audit Trail**: Track when MO IDs are generated vs. manually provided
4. **Bulk Import**: Special handling for bulk patient imports with MO IDs

## Related Documentation
- [Clinic Patient API Guide](./CLINIC_PATIENT_ENHANCED_API_GUIDE.md)
- [Clinic Isolated System](./CLINIC_ISOLATED_PATIENTS_COMPLETE_GUIDE.md)
- [Database Migrations](./migrations/017_clinic_specific_patients.sql)

---

**Last Updated**: October 17, 2025
**Version**: 1.0
**Status**: Implemented ✅

