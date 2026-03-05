Write-Host "Testing hash-password utility endpoint..." -ForegroundColor Cyan

$body = @{
    password = "Test123456"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/auth/hash-password" -Method POST -Body $body -ContentType "application/json"
    
    Write-Host "Success!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Password: $($response.password)" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Generated Hash:" -ForegroundColor Cyan
    Write-Host $response.hash -ForegroundColor White
    Write-Host ""
    Write-Host "Note: $($response.note)" -ForegroundColor Gray
    
} catch {
    Write-Host "Error:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}

