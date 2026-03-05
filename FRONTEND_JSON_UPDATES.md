# Updated Frontend JSON Documentation

The following endpoints have been updated to make `first_name` and `last_name` optional. If these fields are omitted or left empty, the system will automatically use the `username` as the value for both.

## 1. Create Clinic Staff (Clinic Admin)
**Endpoint:** `POST /api/organizations/admin/staff`

### Minimal Request Body (Recommended)
```json
{
  "username": "reception_staff_01",
  "password": "SecurePassword123",
  "staff_type": "receptionist",
  "clinic_id": "your-clinic-uuid",
  "permissions": ["view_appointments", "manage_patients"]
}
```
*Result: `first_name` and `last_name` will both be set to "reception_staff_01" automatically.*

---

## 2. Create Platform User (Super Admin)
**Endpoint:** `POST /api/auth/admin/users`

### Minimal Request Body
```json
{
  "username": "new_platform_user",
  "password": "SecurePassword123",
  "role_ids": ["role-uuid-1"]
}
```

---

## 3. Public User Registration
**Endpoint:** `POST /api/auth/register`

### Minimal Request Body
```json
{
  "username": "patient_user_99",
  "password": "SecurePassword123"
}
```

---

## Summary of Changes
| Field | Status | Default Logic |
| :--- | :--- | :--- |
| `username` | **Required** | Must be unique (3-30 chars) |
| `password` | **Required** | Min 8 characters |
| `first_name` | Optional | Defaults to `username` if empty |
| `last_name` | Optional | Defaults to `username` if empty |
| `email` | Optional | Must be unique if provided |
| `phone` | Optional | Valid 10-digit number |
