# Appointment Token System Documentation

This document outlines the updated token generation logic for the appointment booking system.

## 1. Token Format
All appointment tokens now follow a standardized alphanumeric format:
**`[DoctorCode]-[TokenNumber]`**

*   **DoctorCode**: A unique 2-letter uppercase code (e.g., `RA`, `AS`).
*   **TokenNumber**: A 2-digit (or more) sequential number (e.g., `01`, `15`, `102`).

### Examples:
| Doctor Name | Department | Token Samples |
| :--- | :--- | :--- |
| Dr. Rahman | Cardiology | `RA-01`, `RA-02`, `RA-03` |
| Dr. Rakesh | Orthopedics | `RK-01`, `RK-02` |
| Dr. Arun | General | `AR-01`, `AR-02` |

## 2. Token Logic Rules

### A. Doctor-Specific Sequences
Each doctor maintains their own independent token sequence. A booking for Doctor A does not affect the next token for Doctor B.

### B. Persistent Numbering (No Daily Reset)
Tokens **do not reset to 01** at the start of a new day. They continue to increment indefinitely for that specific doctor/clinic combination to ensure a clear historical sequence.

### C. Unique Doctor Codes
Codes are automatically generated based on the doctor's name:
1. **Primary Strategy**: First Initial + Last Initial (e.g., `Rahman` -> `RA`).
2. **Collision Strategy (Same Name)**: If `RA` is already taken by another doctor, the system uses the **Department Initial** (e.g., `Rahman` from `Cardiology` -> `RC`, `Rahman` from `Neurology` -> `RN`).
3. **Fallback Strategy**: If initials still clash, a numeric suffix is appended (e.g., `RA1`, `RA2`).

## 3. Frontend Integration Requirements

### Data Type Change
- The `token_number` field in the API responses has changed from **Integer** to **String**.
- **Action Required**: Update Flutter/React models to expect a `String` for the `token_number` field.

### Field Mapping
| Component | Field | Type |
| :--- | :--- | :--- |
| Appointment List | `token_number` | `String` (e.g., "RA-05") |
| Booking Request | `booking_mode` | `String` ("slot" or "walk_in") |
| Booking Request | `individual_slot_id` | `String?` (null for "walk_in") |

## 4. Backend Implementation Details
- **Migration 014**: Changed `token_number` column to `VARCHAR(20)`.
- **Logic Location**: `GenerateTokenNumber` and `GetOrGenerateDoctorCode` in `services/appointment-service/utils/appointment_utils.go`.
