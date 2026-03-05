# Slot Type - Quick Reference Card 🏥

## 🎯 New Naming Convention

| Old (❌) | New (✅) | Meaning |
|----------|----------|---------|
| `offline` | `clinic_visit` | Patient visits clinic |
| `online` | `video_consultation` | Remote video call |

---

## 📝 API Values

### Create Slots
```json
{
  "slot_type": "clinic_visit"  // or "video_consultation"
}
```

### List Slots
```bash
GET /doctor-session-slots?slot_type=clinic_visit
```

---

## 🔄 Follow-Up Types

| Value | Shows |
|-------|-------|
| `follow-up-via-clinic` | Clinic visit slots |
| `follow-up-via-video` | Video consultation slots |

---

## 💻 Flutter Dropdown

```dart
DropdownMenuItem(
  value: 'clinic_visit',
  child: Text('🏥 Clinic Visit'),
),
DropdownMenuItem(
  value: 'video_consultation',
  child: Text('💻 Video Consultation'),
),
DropdownMenuItem(
  value: 'follow-up-via-clinic',
  child: Text('🔄 Follow-Up (Clinic)'),
),
DropdownMenuItem(
  value: 'follow-up-via-video',
  child: Text('🔄 Follow-Up (Video)'),
),
```

---

## 🧪 Test

```bash
# ✅ New values work
GET /doctor-session-slots?slot_type=clinic_visit

# ❌ Old values fail
GET /doctor-session-slots?slot_type=offline
# Returns: 400 Invalid slot_type
```

---

## 🚀 Deploy

1. Run migration: `023_rename_slot_types.sql`
2. Deploy services
3. Update Flutter constants

---

**Done!** ✅

