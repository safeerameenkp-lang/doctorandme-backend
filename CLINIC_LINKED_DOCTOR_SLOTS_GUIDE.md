# Clinic-Linked Doctor Time Slots - Complete Guide

## 🎯 Overview

The time slot system now **only shows slots for doctors who are linked to clinics** via the `clinic_doctor_links` table. This ensures proper clinic-doctor relationships.

---

## 📋 How It Works

### 1. **Doctor Must Be Linked to Clinic First**

Before creating time slots, you must link the doctor to the clinic:

```powershell
# Step 1: Link doctor to clinic
$linkPayload = @{
    doctor_id = "doctor-uuid"
    clinic_id = "clinic-uuid"
    consultation_fee_offline = 500.00
    consultation_fee_online = 300.00
    follow_up_fee = 250.00
    follow_up_days = 7
} | ConvertTo-Json

$link = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/clinic-doctor-links" `
    -Method POST `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    } `
    -Body $linkPayload

Write-Host "✅ Doctor linked to clinic: $($link.id)"
```

### 2. **Create Time Slots for Linked Doctor**

```powershell
# Step 2: Create time slot for this doctor at this clinic
$slotPayload = @{
    doctor_id = "doctor-uuid"
    clinic_id = "clinic-uuid"
    day_of_week = 1  # Monday
    slot_type = "offline"
    start_time = "09:00"
    end_time = "12:00"
    max_patients = 20
    notes = "Morning shift"
} | ConvertTo-Json

$slot = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots" `
    -Method POST `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    } `
    -Body $slotPayload

Write-Host "✅ Time slot created: $($slot.slot_id)"
```

### 3. **List Time Slots (Only Shows Linked Doctors)**

```powershell
# Get all time slots for a clinic (only linked doctors)
$clinicId = "clinic-uuid"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

Write-Host "Found $($response.total_count) slots for linked doctors"
```

---

## 🔄 Complete Workflow Example

### Scenario: Dr. John Works at ABC Clinic (Morning) and XYZ Clinic (Afternoon)

```powershell
$token = "your-jwt-token"
$doctorId = "john-doctor-uuid"
$abcClinicId = "abc-clinic-uuid"
$xyzClinicId = "xyz-clinic-uuid"

# ===== STEP 1: Link Doctor to ABC Clinic =====
Write-Host "`n📌 Linking Dr. John to ABC Clinic..." -ForegroundColor Cyan

$abcLink = @{
    doctor_id = $doctorId
    clinic_id = $abcClinicId
    consultation_fee_offline = 500.00
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8081/api/clinic-doctor-links" `
    -Method POST `
    -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} `
    -Body $abcLink

Write-Host "✅ Linked to ABC Clinic" -ForegroundColor Green

# ===== STEP 2: Create Morning Slot at ABC Clinic =====
Write-Host "`n📅 Creating morning slot at ABC Clinic..." -ForegroundColor Cyan

$morningSlot = @{
    doctor_id = $doctorId
    clinic_id = $abcClinicId
    day_of_week = 1  # Monday
    slot_type = "offline"
    start_time = "09:00"
    end_time = "12:00"
    max_patients = 20
    notes = "Morning shift at ABC"
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots" `
    -Method POST `
    -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} `
    -Body $morningSlot

Write-Host "✅ Morning slot created" -ForegroundColor Green

# ===== STEP 3: Link Doctor to XYZ Clinic =====
Write-Host "`n📌 Linking Dr. John to XYZ Clinic..." -ForegroundColor Cyan

$xyzLink = @{
    doctor_id = $doctorId
    clinic_id = $xyzClinicId
    consultation_fee_offline = 600.00
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8081/api/clinic-doctor-links" `
    -Method POST `
    -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} `
    -Body $xyzLink

Write-Host "✅ Linked to XYZ Clinic" -ForegroundColor Green

# ===== STEP 4: Create Afternoon Slot at XYZ Clinic =====
Write-Host "`n📅 Creating afternoon slot at XYZ Clinic..." -ForegroundColor Cyan

$afternoonSlot = @{
    doctor_id = $doctorId
    clinic_id = $xyzClinicId
    day_of_week = 1  # Monday
    slot_type = "offline"
    start_time = "14:00"
    end_time = "17:00"
    max_patients = 15
    notes = "Afternoon shift at XYZ"
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots" `
    -Method POST `
    -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} `
    -Body $afternoonSlot

Write-Host "✅ Afternoon slot created" -ForegroundColor Green

# ===== STEP 5: Get All Slots for Dr. John =====
Write-Host "`n📋 Getting all slots for Dr. John..." -ForegroundColor Cyan

$allSlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId" `
    -Method GET `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "`n✅ Dr. John's Complete Schedule:" -ForegroundColor Green
$allSlots.time_slots | ForEach-Object {
    Write-Host "  🏥 $($_.clinic_name): $($_.start_time)-$($_.end_time) [$($_.slot_type)]"
}

# ===== STEP 6: Get Slots Only for ABC Clinic =====
Write-Host "`n📋 Getting slots for Dr. John at ABC Clinic only..." -ForegroundColor Cyan

$abcSlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$abcClinicId" `
    -Method GET `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "`n✅ Dr. John's Schedule at ABC Clinic:" -ForegroundColor Green
$abcSlots.time_slots | ForEach-Object {
    Write-Host "  📅 $($_.day_name) $($_.start_time)-$($_.end_time) - Max $($_.max_patients) patients"
}

# ===== STEP 7: Get Slots Only for XYZ Clinic =====
Write-Host "`n📋 Getting slots for Dr. John at XYZ Clinic only..." -ForegroundColor Cyan

$xyzSlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$xyzClinicId" `
    -Method GET `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "`n✅ Dr. John's Schedule at XYZ Clinic:" -ForegroundColor Green
$xyzSlots.time_slots | ForEach-Object {
    Write-Host "  📅 $($_.day_name) $($_.start_time)-$($_.end_time) - Max $($_.max_patients) patients"
}
```

---

## 🎯 Use Cases

### Use Case 1: Show Doctors at a Clinic (with Slots)

```powershell
# Get all doctors at ABC Clinic
$clinicId = "abc-clinic-uuid"

# Get doctors
$doctors = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctors/clinic/$clinicId" `
    -Headers @{"Authorization" = "Bearer $token"}

# For each doctor, get their time slots at THIS clinic
$doctors.doctors | ForEach-Object {
    $doctorName = $_.full_name
    $doctorId = $_.doctor_id
    
    Write-Host "`n👨‍⚕️ $doctorName" -ForegroundColor Cyan
    
    $slots = Invoke-RestMethod `
        -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId" `
        -Headers @{"Authorization" = "Bearer $token"}
    
    if ($slots.total_count -gt 0) {
        $slots.time_slots | ForEach-Object {
            Write-Host "   📅 $($_.day_name) $($_.start_time)-$($_.end_time) [$($_.slot_type)]"
        }
    } else {
        Write-Host "   ⚠️  No slots configured" -ForegroundColor Yellow
    }
}
```

### Use Case 2: User Clicks on Doctor - Show Slots at Selected Clinic

```powershell
# User selects: Dr. John at ABC Clinic
$selectedDoctorId = "john-doctor-uuid"
$selectedClinicId = "abc-clinic-uuid"

# Get ONLY that doctor's slots at THAT clinic
$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$selectedDoctorId&clinic_id=$selectedClinicId&only_active=true" `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "Available Appointment Slots:" -ForegroundColor Cyan
$response.time_slots | Group-Object day_name | ForEach-Object {
    Write-Host "`n📅 $($_.Name):"
    $_.Group | ForEach-Object {
        Write-Host "   ⏰ $($_.start_time) - $($_.end_time) [$($_.slot_type)] (Max: $($_.max_patients))"
    }
}
```

### Use Case 3: Get Today's Available Slots

```powershell
# Get current day (0=Sunday, 1=Monday, etc.)
$today = (Get-Date).DayOfWeek.value__

$doctorId = "doctor-uuid"
$clinicId = "clinic-uuid"

$todaySlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId&day_of_week=$today&only_active=true" `
    -Headers @{"Authorization" = "Bearer $token"}

if ($todaySlots.total_count -gt 0) {
    Write-Host "✅ Available today:" -ForegroundColor Green
    $todaySlots.time_slots | ForEach-Object {
        Write-Host "   $($_.start_time)-$($_.end_time) [$($_.slot_type)]"
    }
} else {
    Write-Host "❌ No slots available today" -ForegroundColor Red
}
```

---

## 🚫 What Happens If Doctor Is Not Linked?

### Scenario: Try to Get Slots for Unlinked Doctor

```powershell
# Dr. Sarah is NOT linked to ABC Clinic
$sarahId = "sarah-doctor-uuid"
$abcClinicId = "abc-clinic-uuid"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$sarahId&clinic_id=$abcClinicId" `
    -Headers @{"Authorization" = "Bearer $token"}

# Result: Empty!
# {
#   "time_slots": [],
#   "total_count": 0
# }
```

**Why?** The system now checks `clinic_doctor_links` table. If there's no active link, no slots are returned.

---

## ✅ Benefits of This Approach

1. **Data Integrity** - Only valid doctor-clinic combinations show slots
2. **Security** - Prevents showing slots for doctors not authorized at a clinic
3. **Flexibility** - Same doctor can have different slots at different clinics
4. **Clinic-Specific Fees** - Each link has its own fee structure
5. **Easy Management** - Deactivate link → slots automatically hidden

---

## 🔧 API Endpoints Summary

| Endpoint | Description | Filters |
|----------|-------------|---------|
| `GET /api/doctor-time-slots` | List all slots (only linked doctors) | `doctor_id`, `clinic_id`, `day_of_week`, `slot_type` |
| `GET /api/doctor-time-slots/doctor/:id` | Get specific doctor's slots | `clinic_id`, `day_of_week`, `slot_type` |
| `GET /api/doctors/clinic/:clinic_id` | Get all doctors at a clinic | - |
| `GET /api/clinic-doctor-links` | Get all doctor-clinic links | `doctor_id`, `clinic_id` |

---

## 📖 Quick Reference

### Get Slots for Doctor at Specific Clinic
```
GET /api/doctor-time-slots?doctor_id={uuid}&clinic_id={uuid}
```

### Get All Slots for a Doctor (All Clinics)
```
GET /api/doctor-time-slots?doctor_id={uuid}
```

### Get All Slots at a Clinic (All Doctors)
```
GET /api/doctor-time-slots?clinic_id={uuid}
```

### Get Today's Slots for Doctor at Clinic
```
GET /api/doctor-time-slots?doctor_id={uuid}&clinic_id={uuid}&day_of_week=1
```

---

## 🔄 Restart Services

After updating the code:

```powershell
docker-compose down
docker-compose up --build -d
docker-compose logs -f organization-service
```

---

## 📝 Summary

✅ Time slots only show for **clinic-linked doctors**  
✅ Doctor can have **different slots at different clinics**  
✅ Filter by `doctor_id` and/or `clinic_id`  
✅ Automatically checks `clinic_doctor_links` table  
✅ Deactivating link hides slots automatically  

This ensures your appointment booking system only shows valid, authorized doctor-clinic combinations!

