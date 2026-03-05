# Follow-Up Slot Type - Quick Reference 🔄

## 🎯 New Slot Type Options

| User Selects | API Value | Shows |
|--------------|-----------|-------|
| Regular Offline | `offline` | Offline slots |
| Regular Online | `online` | Online slots |
| **Follow-Up Offline** | `follow-up-via-offline` | Offline slots |
| **Follow-Up Online** | `follow-up-via-online` | Online slots |

---

## 📝 API Call

```bash
GET /api/organizations/doctor-session-slots?
  doctor_id={uuid}&
  clinic_id={uuid}&
  date=2025-10-20&
  slot_type=follow-up-via-offline  # ✅ New option

Authorization: Bearer {token}
```

---

## 🔄 Mapping

```
follow-up-via-offline  →  Shows offline slots
follow-up-via-online   →  Shows online slots
```

---

## 💻 Flutter Code

```dart
DropdownButtonFormField<String>(
  decoration: InputDecoration(labelText: 'Appointment Type'),
  items: [
    DropdownMenuItem(value: 'offline', child: Text('Regular Offline')),
    DropdownMenuItem(value: 'online', child: Text('Regular Online')),
    DropdownMenuItem(value: 'follow-up-via-offline', child: Text('Follow-Up (Offline)')),
    DropdownMenuItem(value: 'follow-up-via-online', child: Text('Follow-Up (Online)')),
  ],
  onChanged: (value) {
    setState(() => selectedSlotType = value);
    _loadSlots(); // Re-fetch with filter
  },
);
```

---

## ✅ Result

- **Follow-Up Offline** = Only shows offline slots
- **Follow-Up Online** = Only shows online slots
- **No filter** = Shows all slots

---

## 🧪 Quick Test

```bash
# Test 1: Follow-up offline
GET /doctor-session-slots?doctor_id=xxx&slot_type=follow-up-via-offline
# ✅ Returns only offline slots

# Test 2: Follow-up online
GET /doctor-session-slots?doctor_id=xxx&slot_type=follow-up-via-online
# ✅ Returns only online slots
```

---

**Status:** ✅ Ready to use!

