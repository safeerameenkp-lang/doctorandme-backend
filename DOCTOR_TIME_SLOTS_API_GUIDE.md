# Doctor Time Slots API - Complete Guide

## 📍 API Endpoint

```
GET http://localhost:8081/api/doctor-time-slots
```

---

## 🔍 Query Parameters (All Optional)

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `doctor_id` | UUID | Filter by specific doctor | `?doctor_id=abc-123-def-456` |
| `clinic_id` | UUID | Filter by specific clinic | `?clinic_id=xyz-789` |
| `day_of_week` | Integer (0-6) | Filter by day (0=Sunday, 6=Saturday) | `?day_of_week=1` |
| `slot_type` | String | Filter by type: `offline` or `online` | `?slot_type=offline` |
| `only_active` | Boolean | Show only active slots (default: `true`) | `?only_active=false` |

---

## 📊 Response Format

```json
{
  "time_slots": [
    {
      "id": "slot-uuid",
      "doctor_id": "doctor-uuid",
      "doctor_name": "Dr. John Doe",
      "clinic_id": "clinic-uuid",
      "clinic_name": "City Medical Center",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "is_active": true,
      "max_patients": 20,
      "notes": "Morning shift",
      "created_at": "2025-10-11T10:30:00Z",
      "updated_at": "2025-10-11T10:30:00Z"
    },
    {
      "id": "slot-uuid-2",
      "doctor_id": "doctor-uuid",
      "doctor_name": "Dr. John Doe",
      "clinic_id": "clinic-uuid",
      "clinic_name": "City Medical Center",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "online",
      "start_time": "14:00",
      "end_time": "17:00",
      "is_active": true,
      "max_patients": 15,
      "notes": "Online consultations",
      "created_at": "2025-10-11T11:00:00Z",
      "updated_at": "2025-10-11T11:00:00Z"
    }
  ],
  "total_count": 2
}
```

---

## 🧪 Usage Examples

### Example 1: Get All Time Slots for a Doctor

```powershell
$token = "your-jwt-token"
$doctorId = "doctor-uuid"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

Write-Host "Total Slots: $($response.total_count)"
$response.time_slots | ForEach-Object {
    Write-Host "$($_.day_name) $($_.start_time)-$($_.end_time) [$($_.slot_type)] - $($_.clinic_name)"
}
```

**Output:**
```
Total Slots: 5
Monday 09:00-12:00 [offline] - City Medical Center
Monday 14:00-17:00 [online] - City Medical Center
Wednesday 09:00-13:00 [offline] - Downtown Clinic
Friday 15:00-18:00 [offline] - City Medical Center
Saturday 10:00-14:00 [online] - Downtown Clinic
```

---

### Example 2: Get Time Slots for a Specific Clinic

```powershell
$clinicId = "clinic-uuid"

$response = Invoke-RestMethod `
    -Uri "ListDoctorTimeSlots?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

# Show all doctors' schedules at this clinic
$response.time_slots | Group-Object doctor_name | ForEach-Object {
    Write-Host "`n👨‍⚕️ $($_.Name):"
    $_.Group | ForEach-Object {
        Write-Host "   $($_.day_name) $($_.start_time)-$($_.end_time) [$($_.slot_type)]"
    }
}
```

---

### Example 3: Get Monday's Offline Slots

```powershell
$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?day_of_week=1&slot_type=offline" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

Write-Host "Monday Offline Consultations:"
$response.time_slots | ForEach-Object {
    Write-Host "🏥 $($_.clinic_name) - $($_.doctor_name) ($($_.start_time)-$($_.end_time))"
}
```

---

### Example 4: Get Specific Doctor's Slots at a Specific Clinic

```powershell
$doctorId = "doctor-uuid"
$clinicId = "clinic-uuid"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

Write-Host "Doctor's Schedule at This Clinic:"
$response.time_slots | Format-Table day_name, start_time, end_time, slot_type, max_patients -AutoSize
```

---

### Example 5: Get All Slots (Including Inactive)

```powershell
$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?only_active=false" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

# Show active vs inactive
$response.time_slots | Group-Object is_active | ForEach-Object {
    $status = if ($_.Name -eq "True") { "✅ Active" } else { "❌ Inactive" }
    Write-Host "$status: $($_.Count) slots"
}
```

---

## 📅 Day of Week Reference

| Value | Day |
|-------|-----|
| 0 | Sunday |
| 1 | Monday |
| 2 | Tuesday |
| 3 | Wednesday |
| 4 | Thursday |
| 5 | Friday |
| 6 | Saturday |

---

## 🧪 Complete Test Script

```powershell
# Test: List Doctor Time Slots
param(
    [string]$Token = "your-jwt-token",
    [string]$DoctorId = "",
    [string]$ClinicId = "",
    [string]$BaseUrl = "http://localhost:8081/api"
)

function Test-TimeSlots {
    param([string]$Url, [string]$Description)
    
    Write-Host "`n=== $Description ===" -ForegroundColor Cyan
    Write-Host "URL: $Url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method GET -Headers @{
            "Authorization" = "Bearer $Token"
        }
        
        Write-Host "✓ Found $($response.total_count) time slots" -ForegroundColor Green
        
        if ($response.total_count -gt 0) {
            $response.time_slots | ForEach-Object {
                Write-Host "`n  📅 $($_.day_name) ($($_.start_time) - $($_.end_time))" -ForegroundColor White
                Write-Host "     Doctor: $($_.doctor_name)"
                Write-Host "     Clinic: $($_.clinic_name)"
                Write-Host "     Type: $($_.slot_type)"
                Write-Host "     Max Patients: $($_.max_patients)"
                Write-Host "     Active: $($_.is_active)"
                if ($_.notes) {
                    Write-Host "     Notes: $($_.notes)"
                }
            }
        }
    } catch {
        Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 1: All time slots
Test-TimeSlots "$BaseUrl/doctor-time-slots" "All Time Slots"

# Test 2: Filter by doctor
if ($DoctorId) {
    Test-TimeSlots "$BaseUrl/doctor-time-slots?doctor_id=$DoctorId" "Time Slots for Doctor"
}

# Test 3: Filter by clinic
if ($ClinicId) {
    Test-TimeSlots "$BaseUrl/doctor-time-slots?clinic_id=$ClinicId" "Time Slots for Clinic"
}

# Test 4: Filter by day (Monday)
Test-TimeSlots "$BaseUrl/doctor-time-slots?day_of_week=1" "Monday Time Slots"

# Test 5: Filter by type
Test-TimeSlots "$BaseUrl/doctor-time-slots?slot_type=offline" "Offline Time Slots"

# Test 6: Combine filters
if ($DoctorId -and $ClinicId) {
    Test-TimeSlots "$BaseUrl/doctor-time-slots?doctor_id=$DoctorId&clinic_id=$ClinicId&day_of_week=1" "Doctor's Monday Slots at Clinic"
}
```

**Run it:**
```powershell
.\test-time-slots.ps1 -Token "your-token" -DoctorId "doctor-uuid" -ClinicId "clinic-uuid"
```

---

## 🎯 Common Use Cases

### Use Case 1: Show Doctor's Weekly Schedule
```powershell
$doctorId = "doctor-uuid"
$response = Invoke-RestMethod -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId" `
    -Headers @{"Authorization" = "Bearer $token"}

# Group by day
$response.time_slots | Group-Object day_of_week | Sort-Object Name | ForEach-Object {
    $dayName = $_.Group[0].day_name
    Write-Host "`n📅 $dayName"
    $_.Group | ForEach-Object {
        Write-Host "   $($_.start_time)-$($_.end_time) at $($_.clinic_name) [$($_.slot_type)]"
    }
}
```

### Use Case 2: Find Available Appointment Slots for Today
```powershell
$today = (Get-Date).DayOfWeek.value__ # 0=Sunday, 1=Monday, etc.
$doctorId = "doctor-uuid"
$clinicId = "clinic-uuid"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId&day_of_week=$today&only_active=true" `
    -Headers @{"Authorization" = "Bearer $token"}

if ($response.total_count -gt 0) {
    Write-Host "Available slots today:"
    $response.time_slots | ForEach-Object {
        Write-Host "  $($_.start_time)-$($_.end_time) [$($_.slot_type)] - Max $($_.max_patients) patients"
    }
} else {
    Write-Host "No slots available today"
}
```

### Use Case 3: Show Clinic's Full Schedule
```powershell
$clinicId = "clinic-uuid"
$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?clinic_id=$clinicId&only_active=true" `
    -Headers @{"Authorization" = "Bearer $token"}

# Create a schedule grid
$response.time_slots | Group-Object day_of_week | Sort-Object Name | ForEach-Object {
    $dayName = $_.Group[0].day_name
    Write-Host "`n📅 $dayName" -ForegroundColor Cyan
    
    $_.Group | Group-Object doctor_name | ForEach-Object {
        Write-Host "  👨‍⚕️ $($_.Name)"
        $_.Group | ForEach-Object {
            Write-Host "     $($_.start_time)-$($_.end_time) [$($_.slot_type)]"
        }
    }
}
```

---

## 🔗 Related APIs

### Create Time Slot
```http
POST /api/doctor-time-slots
```

```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "day_of_week": 1,
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "12:00",
  "max_patients": 20,
  "notes": "Morning shift"
}
```

### Update Time Slot
```http
PUT /api/doctor-time-slots/:slot_id
```

### Delete Time Slot
```http
DELETE /api/doctor-time-slots/:slot_id
```

### Get Doctor's Available Slots (For Appointments)
```http
GET /api/doctor-time-slots/doctor/:doctor_id?clinic_id=...
```

---

## ⚠️ Error Handling

### 400 - Invalid Parameters
```json
{
  "error": "Invalid day_of_week. Must be 0-6 (0=Sunday, 6=Saturday)"
}
```

### 401 - Unauthorized
```json
{
  "error": "Unauthorized"
}
```

### Empty Results
```json
{
  "time_slots": [],
  "total_count": 0
}
```

---

## 💡 Tips

1. **Default Behavior**: By default, only active slots are returned (`only_active=true`)
2. **Time Format**: Times are in 24-hour format (HH:MM)
3. **Sorting**: Results are automatically sorted by day, then start time, then slot type
4. **Multiple Filters**: You can combine multiple query parameters
5. **Authentication**: Bearer token is required for all requests

---

## 📖 Summary

**Quick URLs:**
```
# All slots
GET /api/doctor-time-slots

# Doctor's slots
GET /api/doctor-time-slots?doctor_id={uuid}

# Clinic's slots
GET /api/doctor-time-slots?clinic_id={uuid}

# Specific day
GET /api/doctor-time-slots?day_of_week=1

# Offline only
GET /api/doctor-time-slots?slot_type=offline

# Combined filters
GET /api/doctor-time-slots?doctor_id={uuid}&clinic_id={uuid}&day_of_week=1&slot_type=offline
```

