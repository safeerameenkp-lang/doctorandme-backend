@echo off
echo 🐳 Docker Build Retry Script for DrAndMe Backend
echo ===============================================

echo.
echo 🔧 Quick Fix Options:
echo.

echo Option 1: Try building with PowerShell script
echo   powershell -ExecutionPolicy Bypass -File docker-build-retry.ps1
echo.

echo Option 2: Manual Docker commands
echo   docker pull alpine:3.18
echo   docker-compose build
echo.

echo Option 3: Clear Docker cache and retry
echo   docker system prune -a
echo   docker-compose build --no-cache
echo.

echo Option 4: Use alternative registry
echo   docker pull registry.k8s.io/alpine:3.18
echo   docker-compose build
echo.

echo Option 5: Build individual services
echo   docker build -f services/auth-service/Dockerfile -t drandme-auth-service .
echo   docker build -f services/organization-service/Dockerfile -t drandme-organization-service .
echo   docker build -f services/appointment-service/Dockerfile -t drandme-appointment-service .
echo.

echo 🚀 Running PowerShell script...
powershell -ExecutionPolicy Bypass -File docker-build-retry.ps1

pause
