#!/usr/bin/env pwsh

# Script to help fix passwords in the database
# This script will hash your password and provide SQL to update it

Write-Host "=== Password Hash Generator ===" -ForegroundColor Cyan
Write-Host ""

# Get password from user
$password = Read-Host "Enter the password you want to hash"

# Call the hash-password endpoint
$body = @{
    password = $password
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/auth/hash-password" -Method POST -Body $body -ContentType "application/json"
    
    Write-Host ""
    Write-Host "✓ Password hashed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Original Password: $password" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Bcrypt Hash:" -ForegroundColor Cyan
    Write-Host $response.hash -ForegroundColor White
    Write-Host ""
    Write-Host "SQL to update your user (replace YOUR_USERNAME):" -ForegroundColor Cyan
    Write-Host "UPDATE users SET password_hash = '$($response.hash)' WHERE username = 'YOUR_USERNAME';" -ForegroundColor White
    Write-Host ""
    Write-Host "Example with specific username:" -ForegroundColor Cyan
    $username = Read-Host "Enter your username (or press Enter to skip)"
    if ($username) {
        Write-Host ""
        Write-Host "UPDATE users SET password_hash = '$($response.hash)' WHERE username = '$username';" -ForegroundColor Green
        Write-Host ""
    }
    
} catch {
    Write-Host "Error: Could not connect to auth service" -ForegroundColor Red
    Write-Host "Make sure the auth service is running on http://localhost:8080" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Error details:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

