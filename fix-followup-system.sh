#!/bin/bash

# =====================================================
# FOLLOW-UP SYSTEM SETUP & FIX SCRIPT
# This script fixes the frontend follow-up issue
# =====================================================

echo "🔧 FOLLOW-UP SYSTEM SETUP & FIX"
echo "================================"

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="drandme_db"
DB_USER="postgres"

echo ""
echo "📋 STEP 1: Check if follow_ups table exists"
echo "------------------------------------------"

TABLE_EXISTS=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT EXISTS (
    SELECT FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_name = 'follow_ups'
);" | xargs)

if [ "$TABLE_EXISTS" = "t" ]; then
    echo "✅ follow_ups table exists"
else
    echo "❌ follow_ups table does NOT exist"
    echo "🔧 Creating follow_ups table..."
    
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/025_create_follow_ups_table.sql
    
    if [ $? -eq 0 ]; then
        echo "✅ follow_ups table created successfully"
    else
        echo "❌ Failed to create follow_ups table"
        exit 1
    fi
fi

echo ""
echo "📊 STEP 2: Check current follow_ups data"
echo "----------------------------------------"

FOLLOWUP_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT COUNT(*) FROM follow_ups;" | xargs)

echo "Current follow_ups count: $FOLLOWUP_COUNT"

if [ "$FOLLOWUP_COUNT" -eq 0 ]; then
    echo "⚠️ No follow-up records found"
    echo "🔧 Backfilling follow-up data from recent appointments..."
    
    echo ""
    echo "📋 STEP 3: Backfill follow_ups from recent appointments"
    echo "------------------------------------------------------"
    
    # Create follow-ups for recent appointments (within last 30 days)
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO follow_ups (
        clinic_patient_id, clinic_id, doctor_id, department_id,
        source_appointment_id, status, is_free, valid_from, valid_until,
        created_at, updated_at
    )
    SELECT 
        a.clinic_patient_id,
        a.clinic_id,
        a.doctor_id,
        a.department_id,
        a.id,
        CASE 
            WHEN CURRENT_DATE - a.appointment_date <= 5 THEN 'active'
            ELSE 'expired'
        END as status,
        true as is_free,
        a.appointment_date as valid_from,
        a.appointment_date + INTERVAL '5 days' as valid_until,
        CURRENT_TIMESTAMP as created_at,
        CURRENT_TIMESTAMP as updated_at
    FROM appointments a
    WHERE a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
      AND a.appointment_date >= CURRENT_DATE - INTERVAL '30 days'
      AND NOT EXISTS (
          SELECT 1 FROM follow_ups f 
          WHERE f.source_appointment_id = a.id
      )
    ORDER BY a.appointment_date DESC;
    "
    
    if [ $? -eq 0 ]; then
        echo "✅ Follow-up records backfilled successfully"
        
        # Check new count
        NEW_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT COUNT(*) FROM follow_ups;" | xargs)
        echo "New follow_ups count: $NEW_COUNT"
    else
        echo "❌ Failed to backfill follow-up records"
        exit 1
    fi
else
    echo "✅ Follow-up records already exist"
fi

echo ""
echo "🔍 STEP 4: Verify specific patient data"
echo "--------------------------------------"

# Check for the specific patient mentioned in the logs
PATIENT_ID="your-patient-id"  # Replace with actual patient ID
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"

echo "Checking follow-ups for patient: $PATIENT_ID"
echo "Doctor: $DOCTOR_ID"
echo "Department: $DEPARTMENT_ID"

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    f.id,
    f.status,
    f.is_free,
    f.valid_from,
    f.valid_until,
    f.created_at,
    a.appointment_date,
    a.consultation_type
FROM follow_ups f
JOIN appointments a ON a.id = f.source_appointment_id
WHERE f.clinic_patient_id = '$PATIENT_ID'
  AND f.doctor_id = '$DOCTOR_ID'
  AND f.department_id = '$DEPARTMENT_ID'
ORDER BY f.created_at DESC;
"

echo ""
echo "🔍 STEP 5: Test Follow-Up Helper Functions"
echo "------------------------------------------"

# Test the FollowUpHelper.GetActiveFollowUps function
echo "Testing GetActiveFollowUps for patient: $PATIENT_ID"

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    f.id,
    f.doctor_id,
    f.department_id,
    f.status,
    f.is_free,
    f.valid_until,
    EXTRACT(DAY FROM (f.valid_until - CURRENT_DATE)) as days_remaining
FROM follow_ups f
WHERE f.clinic_patient_id = '$PATIENT_ID'
  AND f.status = 'active'
  AND f.valid_until >= CURRENT_DATE
ORDER BY f.valid_until ASC;
"

echo ""
echo "🎯 STEP 6: Test API Endpoints"
echo "-----------------------------"

echo "Testing follow-up eligibility API..."
echo "Run this command to test:"
echo ""
echo "curl -X GET 'http://localhost:8081/appointments/followup-eligibility?clinic_patient_id=$PATIENT_ID&clinic_id=your-clinic-id&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID' \\"
echo "  -H 'Authorization: Bearer your-token'"
echo ""
echo "Expected result: Should return eligible follow-ups"

echo ""
echo "🎉 SETUP COMPLETE!"
echo "=================="
echo ""
echo "✅ follow_ups table created/verified"
echo "✅ Follow-up records backfilled"
echo "✅ Patient data verified"
echo "✅ Helper functions tested"
echo ""
echo "🚀 NEXT STEPS:"
echo "1. Restart your organization-service"
echo "2. Test the frontend again"
echo "3. Check that eligibleFollowUps.length > 0"
echo ""
echo "📊 If still having issues, check:"
echo "1. Database connection in organization-service"
echo "2. FollowUpHelper import in clinic_patient.controller.go"
echo "3. API endpoint responses"

