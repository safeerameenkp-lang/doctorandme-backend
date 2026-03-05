# Test script for Doctor List API
# This shows the CORRECT URL format

$CLINIC_ID = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjAxOTUwNzgsImlhdCI6MTc2MDE5NDE3OCwic3ViIjoiZDJkOTYxYWYtZjA4OS00Yjc3LTllNDUtNjg2ZmIyZjY3YWRjIiwidHlwZSI6ImFjY2VzcyJ9.HLxotcE06HxHL8xnGzoB8wBDLM0i7nIeyE1v9H7xTj0"

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Testing Doctor List API" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "CORRECT URL FORMAT:" -ForegroundColor Green
Write-Host "http://localhost:8081/api/organizations/doctors?clinic_id=$CLINIC_ID" -ForegroundColor White
Write-Host ""

Write-Host "WRONG URL FORMAT (what you were using):" -ForegroundColor Red
Write-Host "http://localhost:8081/api/organizations/doctors?$CLINIC_ID" -ForegroundColor Gray
Write-Host ""

Write-Host "The difference:" -ForegroundColor Yellow
Write-Host "  ✓ CORRECT: ?clinic_id=<uuid>" -ForegroundColor Green
Write-Host "  ✗ WRONG:   ?<uuid> (missing parameter name)" -ForegroundColor Red
Write-Host ""

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Testing API..." -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

try {
    # Test 1: With clinic_id (should return doctors linked to that clinic)
    Write-Host "Test 1: Get doctors for clinic $CLINIC_ID" -ForegroundColor Yellow
    $response1 = Invoke-WebRequest -Uri "http://localhost:8081/api/organizations/doctors?clinic_id=$CLINIC_ID" `
        -Headers @{"Authorization"="Bearer $TOKEN"} `
        -UseBasicParsing
    
    Write-Host "Status: $($response1.StatusCode)" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor White
    $response1.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
    Write-Host ""
    
    # Test 2: Without clinic_id (should return all doctors)
    Write-Host "Test 2: Get all doctors (no clinic filter)" -ForegroundColor Yellow
    $response2 = Invoke-WebRequest -Uri "http://localhost:8081/api/organizations/doctors" `
        -Headers @{"Authorization"="Bearer $TOKEN"} `
        -UseBasicParsing
    
    Write-Host "Status: $($response2.StatusCode)" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor White
    $response2.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
    Write-Host ""
    
    Write-Host "===============================================" -ForegroundColor Cyan
    Write-Host "  Tests Complete!" -ForegroundColor Cyan
    Write-Host "===============================================" -ForegroundColor Cyan
    
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible issues:" -ForegroundColor Yellow
    Write-Host "  1. Service not running: docker-compose ps" -ForegroundColor Gray
    Write-Host "  2. Token expired: Get a new token from login" -ForegroundColor Gray
    Write-Host "  3. Service not updated: Rebuild with 'docker-compose up -d --build organization-service'" -ForegroundColor Gray
}

