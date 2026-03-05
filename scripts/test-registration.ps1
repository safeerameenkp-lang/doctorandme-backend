# Test Registration with Super Admin Role

Write-Host "=== Testing Registration API with Super Admin Role ===" -ForegroundColor Cyan

$AUTH_URL = "http://localhost:8080/api/auth"

# Test registration payload
$registrationBody = @{
    first_name = "Super"
    last_name = "Admin"
    email = "superadmin@drandme.com"
    username = "superadmin"
    phone = "9999999999"
    password = "superadmin123"
} | ConvertTo-Json

Write-Host "`nRegistering new user..." -ForegroundColor Yellow
Write-Host "POST $AUTH_URL/register" -ForegroundColor Gray
Write-Host $registrationBody -ForegroundColor Gray

try {
    $response = Invoke-RestMethod -Uri "$AUTH_URL/register" -Method Post `
        -Body $registrationBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "`n✅ Registration successful!" -ForegroundColor Green
    Write-Host "`nUser Details:" -ForegroundColor Cyan
    Write-Host "ID: $($response.user.id)" -ForegroundColor White
    Write-Host "Username: $($response.user.username)" -ForegroundColor White
    Write-Host "Email: $($response.user.email)" -ForegroundColor White
    Write-Host "Name: $($response.user.first_name) $($response.user.last_name)" -ForegroundColor White
    
    Write-Host "`nTokens received:" -ForegroundColor Cyan
    Write-Host "Access Token: $($response.accessToken.Substring(0, 50))..." -ForegroundColor White
    Write-Host "Refresh Token: $($response.refreshToken.Substring(0, 50))..." -ForegroundColor White

    # Now login to check the role
    Write-Host "`n--- Testing Login to Verify Role ---" -ForegroundColor Yellow
    
    $loginBody = @{
        login = "superadmin"
        password = "superadmin123"
    } | ConvertTo-Json

    Write-Host "POST $AUTH_URL/login" -ForegroundColor Gray
    
    $loginResponse = Invoke-RestMethod -Uri "$AUTH_URL/login" -Method Post `
        -Body $loginBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "`n✅ Login successful!" -ForegroundColor Green
    Write-Host "`nAssigned Roles:" -ForegroundColor Cyan
    foreach ($role in $loginResponse.roles) {
        Write-Host "  - Role: $($role.name)" -ForegroundColor $(if ($role.name -eq "super_admin") { "Green" } else { "Yellow" })
        Write-Host "    ID: $($role.id)" -ForegroundColor Gray
        Write-Host "    Permissions: $($role.permissions | ConvertTo-Json -Compress)" -ForegroundColor Gray
    }

    if ($loginResponse.roles[0].name -eq "super_admin") {
        Write-Host "`n🎉 SUCCESS! Default role is now SUPER_ADMIN!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ FAILED! Default role is: $($loginResponse.roles[0].name)" -ForegroundColor Red
    }

} catch {
    $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json -ErrorAction SilentlyContinue
    if ($errorDetails) {
        Write-Host "`n❌ Registration failed!" -ForegroundColor Red
        Write-Host "Error: $($errorDetails.error)" -ForegroundColor Red
    } else {
        Write-Host "`n❌ Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan

