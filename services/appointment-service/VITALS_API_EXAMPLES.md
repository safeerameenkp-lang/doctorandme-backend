# Vitals API Examples

Here are the valid JSON request bodies for the `POST /api/v1/vitals` endpoint.

## 1. Full Payload (All Fields)
Use this when you have all vital signs recorded.

```json
{
  "appointment_id": "82d15ae7-a179-4502-9a4c-56240b407766",
  "recorded_by": "e8d1a760-e15d-4797-be87-2758736e39cc",
  "systolic_bp": 120,
  "diastolic_bp": 80,
  "temperature": 98.6,
  "pulse_rate": 72,
  "height_cm": 175,
  "weight_kg": 70.5
}
```

## 2. Partial Payload (Blood Pressure Only)
Use this if you only captured BP.

```json
{
  "appointment_id": "82d15ae7-a179-4502-9a4c-56240b407766",
  "recorded_by": "e8d1a760-e15d-4797-be87-2758736e39cc",
  "systolic_bp": 120,
  "diastolic_bp": 80
}
```

## 3. Partial Payload (Vitals Only - Temperature & Weight)
Note: Values can be integers or decimals (floats).

```json
{
  "appointment_id": "82d15ae7-a179-4502-9a4c-56240b407766",
  "recorded_by": "e8d1a760-e15d-4797-be87-2758736e39cc",
  "temperature": 12.2,
  "weight_kg": 65
}
```

## 4. Minimal Payload (Required Fields Only)
Technically valid if no actual vitals are recorded yet (but usually you'd want at least one value).

```json
{
  "appointment_id": "82d15ae7-a179-4502-9a4c-56240b407766",
  "recorded_by": "e8d1a760-e15d-4797-be87-2758736e39cc"
}
```

## 5. Invalid Payload (Common Mistakes)
*   **Strings instead of numbers**: `"systolic_bp": "120"` (Should be `120`)
*   **Wrong field names**: `"bp_systolic"` (Should be `systolic_bp`)
*   **Missing required fields**: Missing `appointment_id` or `recorded_by`.

## Frontend Integration Tips
*   Ensure `appointment_id` is a valid UUID string.
*   Ensure `recorded_by` is a valid UUID string (usually the logged-in user's ID).
*   Send `Content-Type: application/json` header.
