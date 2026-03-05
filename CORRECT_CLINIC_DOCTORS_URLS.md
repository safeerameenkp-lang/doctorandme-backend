# ✅ Correct URLs for Clinic Doctors APIs

## ❌ WRONG URL (404 Error)
```
http://localhost:8081/api/organizations/doctors/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/doctors
```

## ✅ CORRECT URLS

### Option 1: Get Doctors by Clinic (Recommended)
```
http://localhost:8081/api/doctors/clinic/{clinic_id}
```

**Example:**
```
http://localhost:8081/api/doctors/clinic/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2
```

**PowerShell:**
```powershell
$clinicId = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$token = "your-jwt-token"

$response = Invoke-RestMethod -Uri "http://localhost:8081/api/doctors/clinic/$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

Write-Host "Total Doctors: $($response.total_doctors)"
$response.doctors
```

**cURL:**
```bash
curl -X GET "http://localhost:8081/api/doctors/clinic/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

### Option 2: Get Clinic-Doctor Links
```
http://localhost:8081/api/clinic-doctor-links?clinic_id={clinic_id}
```

**Example:**
```
http://localhost:8081/api/clinic-doctor-links?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2
```

**PowerShell:**
```powershell
$clinicId = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$token = "your-jwt-token"

$response = Invoke-RestMethod -Uri "http://localhost:8081/api/clinic-doctor-links?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

Write-Host "Total Links: $($response.count)"
$response.links
```

---

## 🔄 Restart Services

After the code changes, restart the services:

```powershell
# Stop services
docker-compose down

# Rebuild and start
docker-compose up --build -d

# Check logs
docker-compose logs -f organization-service
```

---

## 📋 All Available Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/doctors/clinic/:clinic_id` | GET | Get all doctors for a clinic with fees ✅ |
| `/api/clinic-doctor-links?clinic_id=...` | GET | Get clinic-doctor links |
| `/api/clinic-doctor-links?doctor_id=...` | GET | Get all clinics for a doctor |
| `/api/doctors` | GET | Get all doctors |
| `/api/doctors?clinic_id=...` | GET | Get doctors filtered by clinic |

---

## 🧪 Quick Test

```powershell
# Replace with your actual values
$baseUrl = "http://localhost:8081/api"
$token = "your-jwt-token"
$clinicId = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"

# Test the endpoint
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/doctors/clinic/$clinicId" `
        -Method GET `
        -Headers @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
    
    Write-Host "✓ Success! Found $($response.total_doctors) doctors" -ForegroundColor Green
    
    $response.doctors | ForEach-Object {
        Write-Host "`nDoctor: $($_.full_name)"
        Write-Host "Specialization: $($_.specialization)"
        Write-Host "Offline Fee: ₹$($_.clinic_specific_fees.consultation_fee_offline)"
        Write-Host "Online Fee: ₹$($_.clinic_specific_fees.consultation_fee_online)"
    }
    
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Host "The clinic might not exist or has no doctors linked." -ForegroundColor Yellow
    }
    
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "Your token might be invalid or expired." -ForegroundColor Yellow
    }
}
```

---

## 🐛 Troubleshooting

### 404 Error
- ✅ Check you're using the correct URL: `/api/doctors/clinic/{clinic_id}`
- ✅ NOT `/api/organizations/doctors/{clinic_id}/doctors`
- ✅ Verify the clinic ID is correct (UUID format)
- ✅ Make sure services are restarted after code changes

### 401 Unauthorized
- ✅ Include `Authorization: Bearer {token}` header
- ✅ Token might be expired - get a new one from login

### Empty Results
- ✅ The clinic might not have any doctors linked yet
- ✅ Use POST `/api/clinic-doctor-links` to link doctors to the clinic

---

## 📚 Related Documentation

- `CLINIC_DOCTORS_LIST_API_COMPLETE.md` - Full API documentation
- `CLINIC_DOCTORS_LIST_QUICK_REFERENCE.md` - Quick reference guide
- `DOCTOR_LINKING_WORKFLOW.md` - How to link doctors to clinics
- `scripts/test-clinic-doctors-list.ps1` - Complete test script

