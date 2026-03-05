# ✅ Simple Example: Clinic-Wise Doctor Slots

## 🎯 Your Need

**Same doctor, different clinics, different time slots**

```
Dr. John:
- ABC Clinic: 09:00-12:00 (Morning)
- XYZ Clinic: 14:00-17:00 (Afternoon)
```

---

## 📱 How to Use in Your UI

### When User Clicks ABC Clinic:

```
GET /api/doctor-time-slots?doctor_id={john-id}&clinic_id={abc-id}
```

**Response:**
```json
{
  "time_slots": [
    {
      "day_name": "Monday",
      "start_time": "09:00",
      "end_time": "12:00",
      "clinic_name": "ABC Clinic"
    }
  ]
}
```
✅ Shows ONLY morning slots

---

### When User Clicks XYZ Clinic:

```
GET /api/doctor-time-slots?doctor_id={john-id}&clinic_id={xyz-id}
```

**Response:**
```json
{
  "time_slots": [
    {
      "day_name": "Monday",
      "start_time": "14:00",
      "end_time": "17:00",
      "clinic_name": "XYZ Clinic"
    }
  ]
}
```
✅ Shows ONLY afternoon slots

---

## 💻 Code Example

```javascript
// User clicks ABC Clinic + Dr. John
const slots = await fetch(
  `/api/doctor-time-slots?doctor_id=${doctorId}&clinic_id=${abcClinicId}`
);
// Result: Morning slots (09:00-12:00)

// User clicks XYZ Clinic + Dr. John
const slots = await fetch(
  `/api/doctor-time-slots?doctor_id=${doctorId}&clinic_id=${xyzClinicId}`
);
// Result: Afternoon slots (14:00-17:00)
```

---

## 🔑 Important

**Always include BOTH parameters:**
- `doctor_id` = Which doctor
- `clinic_id` = Which clinic

The API will return **only** the slots for that doctor at that specific clinic!

---

## ✅ That's It!

Your API already works correctly. Just pass both `doctor_id` and `clinic_id` together, and you'll get the right slots for each clinic! 🎉


