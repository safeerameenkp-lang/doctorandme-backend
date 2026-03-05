# Day of Week Format - Complete Reference

## 🎯 Important: Day of Week Format

The system uses **0-6 format** (NOT 1-7) to match database constraints and Go's standard.

---

## 📅 Day of Week Values

```
0 = Sunday
1 = Monday
2 = Tuesday
3 = Wednesday
4 = Thursday
5 = Friday
6 = Saturday
```

**Example:**
- October 18, 2025 = Saturday = `day_of_week: 6`
- October 19, 2025 = Sunday = `day_of_week: 0`
- October 20, 2025 = Monday = `day_of_week: 1`

---

## ✅ Auto-Calculation

When you create session-based slots, the system **automatically calculates** `day_of_week` from the date:

**Request:**
```json
{
  "date": "2025-10-18",
  ...
}
```

**System Calculates:**
```
October 18, 2025 = Saturday
day_of_week = 6 ✅ (auto-calculated)
```

**Database Stores:**
```sql
INSERT INTO doctor_time_slots (
    specific_date,  -- "2025-10-18"
    day_of_week,    -- 6 (auto-calculated)
    ...
)
```

---

## 🔍 Complete Examples

### Example 1: Monday Slots
```json
POST /doctor-session-slots
{
  "date": "2025-10-20",  // Monday
  "sessions": [...]
}
```

**Auto-calculated:** `day_of_week = 1` ✅

---

### Example 2: Saturday Slots
```json
POST /doctor-session-slots
{
  "date": "2025-10-18",  // Saturday
  "sessions": [...]
}
```

**Auto-calculated:** `day_of_week = 6` ✅

---

### Example 3: Sunday Slots
```json
POST /doctor-session-slots
{
  "date": "2025-10-19",  // Sunday
  "sessions": [...]
}
```

**Auto-calculated:** `day_of_week = 0` ✅

---

## 📊 October 2025 Calendar Reference

| Date | Day | day_of_week |
|------|-----|-------------|
| Oct 13 | Monday | 1 |
| Oct 14 | Tuesday | 2 |
| Oct 15 | Wednesday | 3 |
| Oct 16 | Thursday | 4 |
| Oct 17 | Friday | 5 |
| **Oct 18** | **Saturday** | **6** |
| **Oct 19** | **Sunday** | **0** |
| Oct 20 | Monday | 1 |
| Oct 21 | Tuesday | 2 |
| Oct 22 | Wednesday | 3 |

---

## 💻 JavaScript/TypeScript Conversion

### JavaScript Date to day_of_week (0-6)
```javascript
const date = new Date('2025-10-18');
const dayOfWeek = date.getDay();  // Returns 6 (Saturday)
// No conversion needed! JavaScript uses 0-6 format ✅
```

**JavaScript day values:**
```javascript
date.getDay() returns:
0 = Sunday
1 = Monday
2 = Tuesday
3 = Wednesday
4 = Thursday
5 = Friday
6 = Saturday
```

### Display Day Name
```javascript
const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const date = new Date('2025-10-18');
const dayOfWeek = date.getDay();  // 6
const dayName = dayNames[dayOfWeek];  // "Saturday"

console.log(`${date.toISOString().split('T')[0]} is a ${dayName} (day_of_week: ${dayOfWeek})`);
// Output: "2025-10-18 is a Saturday (day_of_week: 6)"
```

---

## ⚠️ Common Confusion

### ❌ ISO 8601 Format (1-7)
Some systems use ISO 8601 where:
```
1 = Monday
7 = Sunday  ❌ This will FAIL in our system!
```

### ✅ Our System (0-6)
We use standard format:
```
0 = Sunday  ✅
1 = Monday
6 = Saturday
```

---

## 🔍 Why 0-6 Format?

### Reasons:
1. ✅ Matches **Go's time.Weekday()** - Returns 0-6
2. ✅ Matches **JavaScript's getDay()** - Returns 0-6
3. ✅ Matches **PostgreSQL's EXTRACT(DOW)** - Returns 0-6
4. ✅ Matches **Database constraint** - Allows 0-6
5. ✅ **Most common format** in programming

### Go Example:
```go
date := time.Parse("2006-01-02", "2025-10-18")
dayOfWeek := int(date.Weekday())  // Returns 6 (Saturday)
// Perfect! No conversion needed ✅
```

---

## 📋 Quick Reference Table

| Day | 0-6 Format (✅ Our System) | 1-7 Format (❌ ISO 8601) |
|-----|---------------------------|--------------------------|
| Sunday | 0 | 7 |
| Monday | 1 | 1 |
| Tuesday | 2 | 2 |
| Wednesday | 3 | 3 |
| Thursday | 4 | 4 |
| Friday | 5 | 5 |
| Saturday | 6 | 6 |

---

## ✅ Code Fix Applied

**Before (Wrong - Used ISO 8601):**
```go
goWeekday := int(parsedDate.Weekday())
dayOfWeek := goWeekday
if goWeekday == 0 {
    dayOfWeek = 7  // ❌ This violates constraint!
}
```

**After (Correct - Uses 0-6):**
```go
// Auto-calculate day_of_week (0=Sunday to 6=Saturday)
dayOfWeek := int(parsedDate.Weekday())  // ✅ Simple and correct!
```

---

## 🚀 Now Your Request Will Work

**Request:**
```json
{
  "date": "2025-10-18"  // Saturday
}
```

**Auto-calculated:**
```
day_of_week = 6 ✅ (Matches constraint 0-6)
```

**Database:**
```sql
INSERT INTO doctor_time_slots (
    specific_date,  -- '2025-10-18'
    day_of_week,    -- 6 (Saturday, passes constraint!)
    ...
)
```

---

## 📖 Summary

| Aspect | Format | Example |
|--------|--------|---------|
| **Database Constraint** | 0-6 | Sunday = 0, Saturday = 6 |
| **Go time.Weekday()** | 0-6 | Sunday = 0, Saturday = 6 |
| **JavaScript getDay()** | 0-6 | Sunday = 0, Saturday = 6 |
| **Our API** | 0-6 | Sunday = 0, Saturday = 6 |
| **Auto-calculation** | 0-6 | Direct from Go, no conversion |

---

**Status:** ✅ Fixed - Uses Standard 0-6 Format  
**Last Updated:** October 15, 2025

