# Consultation Types - Quick Reference 🏥

## 📋 4 Valid Types

| Send This | Category | Payment | Fee |
|-----------|----------|---------|-----|
| `clinic_visit` | Regular | Required | Yes |
| `video_consultation` | Regular | Required | Yes |
| `follow-up-via-clinic` | Follow-Up | FREE | No |
| `follow-up-via-video` | Follow-Up | FREE | No |

---

## 💻 Flutter Constants

```dart
class ConsultationType {
  static const String clinicVisit = 'clinic_visit';
  static const String videoConsultation = 'video_consultation';
  static const String followUpClinic = 'follow-up-via-clinic';
  static const String followUpVideo = 'follow-up-via-video';
}
```

---

## 📝 Request Examples

### Regular Clinic:
```json
{
  "consultation_type": "clinic_visit",
  "payment_method": "pay_now"
}
```

### Regular Video:
```json
{
  "consultation_type": "video_consultation",
  "payment_method": "pay_later"
}
```

### Follow-Up Clinic (FREE):
```json
{
  "consultation_type": "follow-up-via-clinic"
  // No payment needed!
}
```

### Follow-Up Video (FREE):
```json
{
  "consultation_type": "follow-up-via-video"
  // No payment needed!
}
```

---

## 🔄 Auto-Detection

Backend **automatically** sets `is_follow_up = true` for:
- `follow-up-via-clinic`
- `follow-up-via-video`

No need to manually set it! ✅

---

## 🎯 Slot Filtering

| consultation_type | Get slots for |
|-------------------|---------------|
| `clinic_visit` | `clinic_visit` |
| `video_consultation` | `video_consultation` |
| `follow-up-via-clinic` | `clinic_visit` |
| `follow-up-via-video` | `video_consultation` |

---

**Status:** ✅ Ready to use!

