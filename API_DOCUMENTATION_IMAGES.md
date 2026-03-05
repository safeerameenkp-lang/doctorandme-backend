# API Documentation - Image & Form Data

This document describes how to interact with APIs that require form data and image uploads.

## General Information

- **Content-Type**: `multipart/form-data`
- **Authentication**: Bearer Token required for all endpoints below.

---

## 1. Create Clinic with Admin

**Endpoint**: `POST /api/clinics/with-admin`
**Description**: Creates a new clinic along with a new clinic admin user and uploads a clinic logo.

### Form Fields (Body)

| Field Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `organization_id` | UUID | **Yes** | ID of the parent organization |
| `name` | String | **Yes** | Name of the clinic |
| `clinic_code` | String | No | Unique code for clinic (e.g., "CLI001"). **Auto-generated if left empty.** |
| `email` | String | No | Contact email for the clinic |
| `phone` | String | No | Contact phone for the clinic |
| `address` | String | No | Physical address |
| `license_number` | String | No | Clinic license number |
| `admin_first_name`| String | **Yes** | First name of the new admin user |
| `admin_last_name` | String | **Yes** | Last name of the new admin user |
| `admin_email` | String | **Yes** | Email for the new admin (used for login) |
| `admin_username` | String | **Yes** | Username for the new admin (used for login) |
| `admin_phone` | String | No | Phone number of the admin |
| `admin_password` | String | **Yes** | Password for the new admin account |
| `logo` | File | No | Image file (JPG, PNG) max 5MB |

### Frontend Implementation (Flutter/Dart) Example

```dart
Future<void> createClinicWithAdmin({
  required String organizationId,
  required String name,
  required String adminFirstName,
  required String adminLastName,
  required String adminEmail,
  required String adminUsername,
  required String adminPassword,
  File? logoFile, // The image file
  String? clinicCode,
  String? email,
  String? phone,
  String? address,
  String? licenseNumber,
  String? adminPhone,
}) async {
  var uri = Uri.parse('YOUR_API_URL/api/clinics/with-admin');
  var request = http.MultipartRequest('POST', uri);

  // Add Headers
  request.headers['Authorization'] = 'Bearer YOUR_TOKEN';

  // Add Fields
  request.fields['organization_id'] = organizationId;
  request.fields['name'] = name;
  request.fields['admin_first_name'] = adminFirstName;
  request.fields['admin_last_name'] = adminLastName;
  request.fields['admin_email'] = adminEmail;
  request.fields['admin_username'] = adminUsername;
  request.fields['admin_password'] = adminPassword;

  // Add Optional Fields if they exist
  if (clinicCode != null && clinicCode.isNotEmpty) request.fields['clinic_code'] = clinicCode;
  if (email != null) request.fields['email'] = email;
  if (phone != null) request.fields['phone'] = phone;
  if (address != null) request.fields['address'] = address;
  if (licenseNumber != null) request.fields['license_number'] = licenseNumber;
  if (adminPhone != null) request.fields['admin_phone'] = adminPhone;

  // Add File
  if (logoFile != null) {
      var stream = http.ByteStream(logoFile.openRead());
      var length = await logoFile.length();
      var multipartFile = http.MultipartFile(
        'logo', // This MUST match the backend field name
        stream,
        length,
        filename: logoFile.path.split('/').last,
      );
      request.files.add(multipartFile);
  }

  // Send
  var response = await request.send();
  
  if (response.statusCode == 201) {
    print('Clinic created successfully');
  } else {
    print('Failed to create clinic');
    // You can read the response body here
    final respStr = await response.stream.bytesToString();
    print(respStr);
  }
}
```

---

## 2. Create Doctor

**Endpoint**: `POST /api/doctors`
**Description**: Creates a new doctor profile (and optionally a new user account if user_id is not provided) and uploads a profile image.

### Form Fields (Body)

| Field Name | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `first_name` | String | **Yes** | Only if `user_id` is NOT provided |
| `last_name` | String | **Yes** | Only if `user_id` is NOT provided |
| `email` | String | **Yes** | Only if `user_id` is NOT provided |
| `username` | String | **Yes** | Only if `user_id` is NOT provided |
| `password` | String | **Yes** | Only if `user_id` is NOT provided |
| `phone` | String | No | Phone number |
| `user_id` | UUID | No | If providing an EXISTING user ID (replaces above user fields) |
| `doctor_code` | String | No | Unique doctor code |
| `specialization` | String | No | Eg. "Cardiologist" |
| `license_number` | String | No | Medical license number |
| `consultation_fee` | Float | No | Fee amount (e.g. "500.00") |
| `follow_up_fee` | Float | No | Follow up fee amount |
| `follow_up_days` | Int | No | Number of days for follow up validity |
| `profile_image` | File | No | Image file (JPG, PNG) max 5MB |

### Frontend Implementation (Flutter/Dart) Example

```dart
Future<void> createDoctor({
  required String firstName,
  required String lastName,
  required String email,
  required String username,
  required String password,
  File? profileImage, // The image file
  String? phone,
  String? doctorCode,
  String? specialization,
  String? licenseNumber,
  double? consultationFee,
  double? followUpFee,
  int? followUpDays,
}) async {
  var uri = Uri.parse('YOUR_API_URL/api/doctors');
  var request = http.MultipartRequest('POST', uri);

  // Add Headers
  request.headers['Authorization'] = 'Bearer YOUR_TOKEN';

  // Add Fields (Case: Creating New User)
  request.fields['first_name'] = firstName;
  request.fields['last_name'] = lastName;
  request.fields['email'] = email;
  request.fields['username'] = username;
  request.fields['password'] = password;

  // Add Optional Fields
  if (phone != null) request.fields['phone'] = phone;
  if (doctorCode != null) request.fields['doctor_code'] = doctorCode;
  if (specialization != null) request.fields['specialization'] = specialization;
  if (licenseNumber != null) request.fields['license_number'] = licenseNumber;
  if (consultationFee != null) request.fields['consultation_fee'] = consultationFee.toString();
  if (followUpFee != null) request.fields['follow_up_fee'] = followUpFee.toString();
  if (followUpDays != null) request.fields['follow_up_days'] = followUpDays.toString();

  // Add File
  if (profileImage != null) {
      var stream = http.ByteStream(profileImage.openRead());
      var length = await profileImage.length();
      var multipartFile = http.MultipartFile(
        'profile_image', // This MUST match the backend field name
        stream,
        length,
        filename: profileImage.path.split('/').last,
      );
      request.files.add(multipartFile);
  }

  // Send
  var response = await request.send();
  
  if (response.statusCode == 201) {
    print('Doctor created successfully');
  } else {
    print('Failed to create doctor');
    final respStr = await response.stream.bytesToString();
    print(respStr);
  }
}
```
