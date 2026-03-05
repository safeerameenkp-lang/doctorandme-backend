# ✅ Updated: Time Slots for Clinic-Linked Doctors Only

## 🎯 What Changed?

The `ListDoctorTimeSlots` API now **only shows time slots for doctors who are linked to the clinic** via the `clinic_doctor_links` table.

---

## 📋 How to Use

### Get Slots for a Doctor at a Specific Clinic

```powershell
$doctorId = "doctor-uuid"
$clinicId = "clinic-uuid"
$token = "your-jwt-token"

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }

# Shows ONLY slots for this doctor at THIS clinic
Write-Host "Found $($response.total_count) slots"
$response.time_slots | ForEach-Object {
    Write-Host "$($_.day_name) $($_.start_time)-$($_.end_time) at $($_.clinic_name)"
}
```

---

## ✅ Example Response

```json
{
  "time_slots": [
    {
      "id": "slot-1",
      "doctor_id": "doctor-uuid",
      "doctor_name": "Dr. John Doe",
      "clinic_id": "abc-clinic-uuid",
      "clinic_name": "ABC Clinic",
      "day_of_week": 1,
      "day_name": "Monday",
      "start_time": "09:00",
      "end_time": "12:00",
      "slot_type": "offline",
      "max_patients": 20
    }
  ],
  "total_count": 1
}
```

---

## 🔑 Key Points

1. **Doctor must be linked to clinic first** (via `/api/clinic-doctor-links`)
2. **Slots are clinic-specific** - Same doctor can have different slots at different clinics
3. **Filter by `doctor_id` and `clinic_id`** to get exact slots
4. **Only active links show slots** - Deactivate link = hide slots

---

## 📖 Full Documentation

- **CLINIC_LINKED_DOCTOR_SLOTS_GUIDE.md** - Complete guide with examples
- **DOCTOR_TIME_SLOTS_API_GUIDE.md** - API reference

---

## 🔄 Restart Services

```powershell
docker-compose down
docker-compose up --build -d
```

Then test:

```powershell
.\test-time-slots-now.ps1 -Token "your-token" -DoctorId "doctor-uuid" -ClinicId "clinic-uuid"
```

---

## ✅ Done!

Your time slot system now correctly shows slots **only for clinic-linked doctors**! 🎉


