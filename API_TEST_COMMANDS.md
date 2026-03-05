# 🧪 API Test Commands (Valid Data)

Here are the CURL commands to test `ListDoctorSessionSlots` using the **Valid Clinic IDs** you provided from your logs.

## 1️⃣ List Slots for "clinicq"
**Clinic ID:** `22f437fa-f927-445b-9808-51b57a1701c9`
```bash
curl "http://localhost:8081/doctor-session-slots?doctor_id=REPLACE_WITH_DOCTOR_ID&date=2026-02-18&clinic_id=22f437fa-f927-445b-9808-51b57a1701c9"
```

## 2️⃣ List Slots for "sabik"
**Clinic ID:** `d361bd8e-742c-4961-b940-7f3aa9aa6491`
```bash
curl "http://localhost:8081/doctor-session-slots?doctor_id=REPLACE_WITH_DOCTOR_ID&date=2026-02-18&clinic_id=d361bd8e-742c-4961-b940-7f3aa9aa6491"
```

## 3️⃣ List Slots for "alamala"
**Clinic ID:** `7a6c1211-c029-4923-a1a6-fe3dfe48bdf2`
```bash
curl "http://localhost:8081/doctor-session-slots?doctor_id=REPLACE_WITH_DOCTOR_ID&date=2026-02-18&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
```

## 4️⃣ List Slots (No Clinic Filter - All Sessions)
```bash
curl "http://localhost:8081/doctor-session-slots?doctor_id=REPLACE_WITH_DOCTOR_ID&date=2026-02-18"
```

---

## 🛠️ Debugging Steps
1.  **Replace** `REPLACE_WITH_DOCTOR_ID` with a real `doctor_id`.
2.  **Date Selection**:
    -   Use **Tomorrow** (`2026-02-18`) -> Expect `is_bookable: true`.
    -   Use **Today** (`2026-02-17`) -> Expect past times `is_bookable: false` ("Time Passed").
3.  **Check Output**:
    -   Look at the JSON response.
    -   Look at the Docker Logs (`docker logs organization-service --tail 50`) to see the `DEBUG` messages I added.
