import requests
import time
import sys

BASE_URL = "http://localhost:8000"

def check_endpoint(name, url, expected_status=[200]):
    try:
        response = requests.get(url)
        if response.status_code in expected_status:
            print(f"[OK] {name}: OK ({response.status_code})")
            return True
        else:
            print(f"[FAIL] {name}: Failed. Status: {response.status_code}, Expected: {expected_status}")
            return False
    except Exception as e:
        print(f"[ERROR] {name}: Connection Error: {e}")
        return False

print("=== Verifying Backend System Health via Kong Gateway ===")

# Check Auth Service
check_endpoint("Auth Service Health", f"{BASE_URL}/api/auth/health")

# Check Organization Service
# Note: Organization Service base path changed to /api
check_endpoint("Organization Service Health", f"{BASE_URL}/api/health")

# Check Appointment Service
check_endpoint("Appointment Service Health", f"{BASE_URL}/api/v1/health")

# Check specific endpoints to ensure routing is correct
# Organization Service Resources
check_endpoint("List Clinics (Unauthorized check)", f"{BASE_URL}/api/clinics", [401])
check_endpoint("List Doctors (Unauthorized check)", f"{BASE_URL}/api/doctors", [401])

# Appointment Service Resources
check_endpoint("List Appointments (Unauthorized check)", f"{BASE_URL}/api/v1/appointments", [401])

print("\n=== Verification Complete ===")
print("If all health checks passed, the API Gateway routing is correct.")
print("If you see 404 errors, please rebuild containers: docker-compose up -d --build")
