# Token Number System Guide 🎫

## 🎯 Overview

The token number system automatically generates **sequential token numbers** for appointments. Each token is unique per **doctor + clinic + date**.

---

## 📊 How It Works

### Token Number Generation

```
Token Number = Auto-incrementing integer per doctor/clinic/date
```

**Example for Doctor A at Clinic X on Oct 17:**
- 1st appointment → Token #1
- 2nd appointment → Token #2
- 3rd appointment → Token #3
- ...and so on

---

## 🗄️ Database Structure

### `doctor_tokens` Table

```sql
CREATE TABLE doctor_tokens (
    id UUID PRIMARY KEY,
    doctor_id UUID NOT NULL,          -- Which doctor
    clinic_id UUID NOT NULL,           -- Which clinic
    token_date DATE NOT NULL,          -- Which date
    current_token INTEGER NOT NULL,    -- Last token number assigned
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    
    UNIQUE (doctor_id, clinic_id, token_date)  -- One record per doctor/clinic/date
);
```

---

## 🔄 Token Generation Flow

### When Creating Appointment:

```
Step 1: Call GenerateTokenNumber(doctor_id, clinic_id, date)
         ↓
Step 2: Check if record exists in doctor_tokens
         ↓
    ┌────────┴────────┐
    │                 │
 NO │                 │ YES
    ↓                 ↓
Create record    Increment existing
token = 1        token = token + 1
    │                 │
    └────────┬────────┘
             ↓
Step 3: Return token number
         ↓
Step 4: Save to appointment.token_number
```

---

## 📝 Example Scenarios

### Scenario 1: First Appointment of the Day

**Input:**
- Doctor: Dr. Ahmed (id: doctor-123)
- Clinic: City Clinic (id: clinic-456)
- Date: 2025-10-17

**Database State (Before):**
```
doctor_tokens table: EMPTY for this doctor/clinic/date
```

**What Happens:**
1. Check doctor_tokens → No record found
2. INSERT new record with token = 1
3. Return token_number = 1

**Database State (After):**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      1
```

**Appointment Created:**
```json
{
  "token_number": 1,
  "booking_number": "BN202510170001"
}
```

---

### Scenario 2: Second Appointment (Same Day)

**Input:**
- Same doctor, same clinic, same date

**Database State (Before):**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      1
```

**What Happens:**
1. Check doctor_tokens → Record exists, current_token = 1
2. Increment: new_token = 1 + 1 = 2
3. UPDATE record SET current_token = 2
4. Return token_number = 2

**Database State (After):**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      2
```

**Appointment Created:**
```json
{
  "token_number": 2,
  "booking_number": "BN202510170002"
}
```

---

### Scenario 3: Different Doctor, Same Clinic, Same Day

**Input:**
- Doctor: Dr. Sara (id: doctor-789)
- Clinic: City Clinic (id: clinic-456)
- Date: 2025-10-17

**Database State:**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      2      ← Dr. Ahmed
doctor-789 | clinic-456 | 2025-10-17 |      1      ← Dr. Sara (NEW)
```

**Result:**
- Dr. Ahmed's next token: 3
- Dr. Sara's next token: 2 ✅ (Independent counter)

---

### Scenario 4: Same Doctor, Different Clinic

**Input:**
- Doctor: Dr. Ahmed (id: doctor-123)
- Clinic: East Clinic (id: clinic-999)
- Date: 2025-10-17

**Database State:**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      2      ← City Clinic
doctor-123 | clinic-999 | 2025-10-17 |      1      ← East Clinic (NEW)
```

**Result:**
- City Clinic: Next token = 3
- East Clinic: Next token = 2 ✅ (Independent counter)

---

### Scenario 5: Next Day Reset

**Input:**
- Doctor: Dr. Ahmed (id: doctor-123)
- Clinic: City Clinic (id: clinic-456)
- Date: 2025-10-18 (NEW DAY)

**Database State:**
```sql
doctor_id  | clinic_id  | token_date | current_token
doctor-123 | clinic-456 | 2025-10-17 |      5      ← Yesterday
doctor-123 | clinic-456 | 2025-10-18 |      1      ← Today (NEW)
```

**Result:**
- Oct 17: Tokens 1-5 were used
- Oct 18: Starts fresh from token 1 ✅

---

## 🔐 Race Condition Prevention

### Problem:
What if 2 appointments are created simultaneously?

### Solution: Transaction with SELECT FOR UPDATE

```go
// Start transaction
tx.Begin()

// Lock the row to prevent concurrent updates
SELECT current_token 
FROM doctor_tokens 
WHERE doctor_id = $1 AND clinic_id = $2 AND token_date = $3
FOR UPDATE  ← ⚠️ Locks the row

// Increment safely
UPDATE doctor_tokens SET current_token = current_token + 1

// Commit
tx.Commit()
```

**Timeline:**
```
Time | User A                     | User B
-----+----------------------------+---------------------------
T1   | BEGIN TRANSACTION          | BEGIN TRANSACTION
T2   | SELECT FOR UPDATE (Lock)   | SELECT FOR UPDATE (Wait...)
T3   | token = 5, new = 6         |
T4   | UPDATE token = 6           |
T5   | COMMIT (Release lock)      |
T6   |                            | SELECT FOR UPDATE (Gets lock)
T7   |                            | token = 6, new = 7
T8   |                            | UPDATE token = 7
T9   |                            | COMMIT

Result: User A gets token 6, User B gets token 7 ✅
No conflicts!
```

---

## 📊 Token Number vs Booking Number

| Field | Example | Purpose | Format |
|-------|---------|---------|--------|
| **token_number** | 5 | Queue position | Integer (1, 2, 3...) |
| **booking_number** | BN202510170005 | Unique identifier | BN + Date + Sequence |

**Both are related but different:**
- `token_number`: Simple queue number (1, 2, 3...)
- `booking_number`: Globally unique booking ID

---

## 🧪 Testing

### Test 1: Single Doctor, Single Clinic

```bash
# Appointment 1
POST /appointments/simple
→ token_number: 1

# Appointment 2
POST /appointments/simple
→ token_number: 2

# Appointment 3
POST /appointments/simple
→ token_number: 3
```

**Expected:** Sequential tokens ✅

---

### Test 2: Multiple Doctors, Same Clinic

```bash
# Dr. Ahmed - Appointment 1
POST /appointments/simple (doctor: ahmed, clinic: city)
→ token_number: 1

# Dr. Sara - Appointment 1
POST /appointments/simple (doctor: sara, clinic: city)
→ token_number: 1  ← Independent counter ✅

# Dr. Ahmed - Appointment 2
POST /appointments/simple (doctor: ahmed, clinic: city)
→ token_number: 2  ← Dr. Ahmed's counter continues ✅
```

---

### Test 3: Next Day Reset

```bash
# Oct 17 - Appointments
POST /appointments/simple (date: 2025-10-17)
→ token_number: 1

POST /appointments/simple (date: 2025-10-17)
→ token_number: 2

# Oct 18 - New day
POST /appointments/simple (date: 2025-10-18)
→ token_number: 1  ← Reset to 1 ✅
```

---

## 📋 Code Location

### Token Generation Function

**File:** `services/appointment-service/utils/appointment_utils.go`

**Function:** `GenerateTokenNumber(doctorID, clinicID string, appointmentDate time.Time)`

**Returns:** `(tokenNumber int, error)`

---

## 🔍 Troubleshooting

### Issue 1: All tokens are 1

**Cause:** `doctor_tokens` table doesn't exist

**Solution:** Run migration 022
```bash
Get-Content migrations/022_create_doctor_tokens_table.sql | 
docker exec -i postgres psql -U postgres -d drandme
```

---

### Issue 2: Token numbers not incrementing

**Check:**
```sql
SELECT * FROM doctor_tokens 
WHERE doctor_id = 'your-doctor-id' 
AND clinic_id = 'your-clinic-id' 
AND token_date = CURRENT_DATE;
```

**Should show:** `current_token` incrementing

---

### Issue 3: Duplicate token numbers

**Check unique constraint:**
```sql
SELECT constraint_name 
FROM information_schema.table_constraints 
WHERE table_name = 'doctor_tokens' 
AND constraint_type = 'UNIQUE';
```

**Should exist:** `unique_doctor_clinic_date`

---

## ✅ Summary

### Key Points

| Aspect | Detail |
|--------|--------|
| **Scope** | Per doctor + clinic + date |
| **Start** | Token 1 each day |
| **Increment** | Auto +1 for each appointment |
| **Reset** | Daily (new date → starts from 1) |
| **Uniqueness** | Unique per doctor/clinic/date combo |
| **Concurrency** | Protected by transaction + lock |

---

### Formula

```
Token Number = Count of appointments for:
  - Same doctor
  - Same clinic
  - Same date
  + 1
```

---

### Migration Applied

**File:** `migrations/022_create_doctor_tokens_table.sql`

**Status:** ✅ Applied

**Result:** Token numbers now auto-increment correctly! 🎫

---

**Now your token numbers will be:** 1, 2, 3, 4... ✅ (not all 1s!)

