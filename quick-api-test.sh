#!/bin/bash

# Quick API Test for Follow-Up Reset
# Run this after booking a new regular appointment

echo "🔍 QUICK FOLLOW-UP API TEST"
echo "============================"

# Default values (change these to match your setup)
API_BASE="http://localhost:8080"
CLINIC_ID="7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
DOCTOR_ID="85394ce8-94f7-4dca-a536-34305c46a98e"
DEPARTMENT_ID="9a626b17-3ac0-44a5-bbee-402932f82337"

echo "Using default values:"
echo "API Base: $API_BASE"
echo "Clinic ID: $CLINIC_ID"
echo "Doctor ID: $DOCTOR_ID"
echo "Department ID: $DEPARTMENT_ID"
echo ""

# Test the patient API
echo "🔍 Testing Patient API with doctor+department filters..."
echo "URL: $API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID"
echo ""

response=$(curl -s "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID")

echo "📋 API Response:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

echo ""
echo "🎯 Key Points to Check:"
echo "1. Look for 'eligible_follow_ups' array"
echo "2. If array has entries → Should show GREEN"
echo "3. If array is empty → Should show ORANGE"
echo "4. Check 'appointments' array for latest regular appointment"

echo ""
echo "✅ Test complete!"


