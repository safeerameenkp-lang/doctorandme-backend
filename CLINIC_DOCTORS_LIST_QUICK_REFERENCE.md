# Clinic Doctors List API - Quick Reference

## 🎯 Recommended API

### Get Doctors for a Clinic
```http
GET /api/doctors/clinic/:clinic_id
Authorization: Bearer {token}
```

**Example:**
```powershell
$clinicId = "abc-123-def-456"
$token = "your-jwt-token"

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/doctors/clinic/$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

# Display results
$response.doctors | ForEach-Object {
    Write-Host "$($_.full_name) - $($_.specialization)"
    Write-Host "Offline: ₹$($_.clinic_specific_fees.consultation_fee_offline)"
    Write-Host "Online: ₹$($_.clinic_specific_fees.consultation_fee_online)"
}
```

**Response:**
```json
{
  "clinic_id": "abc-123",
  "doctors": [
    {
      "doctor_id": "doc-123",
      "full_name": "Dr. John Doe",
      "specialization": "Cardiology",
      "email": "john@example.com",
      "phone": "+1234567890",
      "clinic_specific_fees": {
        "consultation_fee_offline": 500.00,
        "consultation_fee_online": 300.00,
        "follow_up_fee": 250.00,
        "follow_up_days": 7
      }
    }
  ],
  "total_doctors": 1
}
```

---

## 🔗 Alternative API (Admin)

### Get Clinic-Doctor Links
```http
GET /api/clinic-doctor-links?clinic_id={clinic_id}
Authorization: Bearer {token}
```

**Example:**
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/clinic-doctor-links?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
    }
```

---

## 📊 Comparison

| Feature | `/doctors/clinic/:id` | `/clinic-doctor-links` |
|---------|----------------------|------------------------|
| Clinic-specific fees | ✅ | ✅ |
| Default fees | ✅ | ❌ |
| Link metadata | ✅ | ✅ |
| Clinic details | ❌ | ✅ |
| **Best for** | **Appointment booking** | **Admin management** |

---

## 🧪 Test Script

```powershell
# Run the test script
.\scripts\test-clinic-doctors-list.ps1 `
    -Token "your-jwt-token" `
    -ClinicId "clinic-uuid"
```

---

## 📚 Full Documentation

See `CLINIC_DOCTORS_LIST_API_COMPLETE.md` for:
- Complete API documentation
- All available endpoints
- Error handling
- Advanced examples
- Use cases

---

## 🔑 Key Points

1. **Always use the clinic-specific fees** from the API response for billing
2. **Default fees** are fallback values if clinic-specific fees are not set
3. Only **active** doctors, users, and links are returned
4. Results are **sorted by doctor name**
5. **Authentication required** for all endpoints

---

## 💡 Common Use Cases

### Display doctors on appointment booking page
```powershell
$doctors = (Invoke-RestMethod -Uri "http://localhost:8080/api/doctors/clinic/$clinicId" `
    -Headers @{"Authorization" = "Bearer $token"}).doctors

# Use: $doctors[0].full_name, $doctors[0].clinic_specific_fees.consultation_fee_offline
```

### Check if a doctor works at a clinic
```powershell
$links = (Invoke-RestMethod -Uri "http://localhost:8080/api/clinic-doctor-links?clinic_id=$clinicId&doctor_id=$doctorId" `
    -Headers @{"Authorization" = "Bearer $token"}).links

if ($links.count -gt 0) {
    Write-Host "Doctor works at this clinic"
}
```

### Get all clinics where a doctor works
```powershell
$clinics = (Invoke-RestMethod -Uri "http://localhost:8080/api/clinic-doctor-links?doctor_id=$doctorId" `
    -Headers @{"Authorization" = "Bearer $token"}).links

$clinics | ForEach-Object {
    Write-Host "$($_.clinic.name) - Fee: ₹$($_.fees.consultation_fee_offline)"
}
```

