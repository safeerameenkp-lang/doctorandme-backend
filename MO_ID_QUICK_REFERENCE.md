# MO ID Auto-Generation - Quick Reference Card

## Feature: MO1001 ✅
**Status**: IMPLEMENTED  
**Date**: October 17, 2025

---

## What It Does

When creating a clinic patient, if `mo_id` is **not provided**, the system **automatically generates** a unique MO ID using the format:

```
{clinic_code}{sequential_number}
```

**Examples**:
- Clinic "MO" → MO0001, MO0002, MO0003...
- Clinic "ABC" → ABC0001, ABC0002, ABC0003...

---

## Quick Usage

### Create Patient (Auto-Generate MO ID)
```bash
POST /clinic-specific-patients
{
  "clinic_id": "uuid",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+91 1234567890"
}

# Response: mo_id = "MO0001" (auto-generated)
```

### Create Patient (Custom MO ID)
```bash
POST /clinic-specific-patients
{
  "clinic_id": "uuid",
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+91 9876543210",
  "mo_id": "CUSTOM123"
}

# Response: mo_id = "CUSTOM123" (uses provided)
```

---

## Key Features

✅ **Auto-Generated**: No manual tracking needed  
✅ **Sequential**: Proper numbering per clinic  
✅ **Unique**: Enforced per clinic  
✅ **Flexible**: Can provide custom IDs  
✅ **Isolated**: Each clinic has own sequence  

---

## Files Changed

**Code**: `services/organization-service/controllers/clinic_patient.controller.go`  
**Lines**: 148-184

---

## Documentation

📄 **Complete Guide**: `CLINIC_PATIENT_AUTO_MO_ID_GUIDE.md`  
📄 **Feature Summary**: `MO_ID_FEATURE_SUMMARY.md`  
📄 **Implementation**: `MO_ID_IMPLEMENTATION_COMPLETE.md`  
📄 **API Guide**: `CLINIC_PATIENT_ENHANCED_API_GUIDE.md`  
🧪 **Test Script**: `test-auto-mo-id.ps1`  

---

## Testing

```powershell
# Run test script
.\test-auto-mo-id.ps1

# Enter clinic ID when prompted
# Script tests all scenarios automatically
```

---

## Examples

### Scenario 1: Sequential Auto-Generation
```
Clinic "MO":
Patient 1 (no mo_id) → MO0001
Patient 2 (no mo_id) → MO0002
Patient 3 (no mo_id) → MO0003
```

### Scenario 2: Custom ID Doesn't Break Sequence
```
Clinic "ABC":
Patient 1 (no mo_id) → ABC0001
Patient 2 (mo_id: "SPECIAL") → SPECIAL
Patient 3 (no mo_id) → ABC0002  ✅ Continues sequence
```

### Scenario 3: Multiple Clinics (Isolated)
```
Clinic "MO":  MO0001, MO0002, MO0003
Clinic "ABC": ABC0001, ABC0002, ABC0003
✅ Each clinic has own sequence
```

---

## Error Handling

### Duplicate MO ID
**Status**: 409 Conflict
```json
{
  "error": "Mo ID exists in this clinic",
  "message": "A patient with this Mo ID already exists..."
}
```

---

## Deployment

✅ No migration required  
✅ No breaking changes  
✅ Backward compatible  
✅ Ready for production  

---

## Technical Details

**Format**: `fmt.Sprintf("%s%04d", clinicCode, maxNumber+1)`  
**Database**: Uses existing `clinic_patients.mo_id` column  
**Constraint**: `UNIQUE (clinic_id, mo_id)`  
**Index**: `idx_clinic_patients_mo_id`  

---

## Quick Commands

```bash
# Create patient (auto MO ID)
curl -X POST http://localhost:8080/clinic-specific-patients \
  -H "Content-Type: application/json" \
  -d '{"clinic_id":"UUID","first_name":"John","last_name":"Doe","phone":"+911234567890"}'

# List patients (see MO IDs)
curl "http://localhost:8080/clinic-specific-patients?clinic_id=UUID"

# Get single patient
curl "http://localhost:8080/clinic-specific-patients/PATIENT_UUID"
```

---

**Version**: 1.0  
**Status**: ✅ COMPLETE  
**Last Updated**: October 17, 2025







