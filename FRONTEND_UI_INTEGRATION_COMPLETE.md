# Frontend UI Integration - Complete Guide 🎨

## 📱 **UI Components & Models**

### **1. Patient List View Component**

```typescript
// PatientListViewModel.ts
export interface PatientListViewModel {
  patients: PatientCardViewModel[];
  searchQuery: string;
  filterActive: boolean;
  selectedClinicId: string;
  loading: boolean;
  error: string | null;
}

export interface PatientCardViewModel {
  // Patient Info
  id: string;
  name: string;                      // "First Last"
  phone: string;
  email: string | null;
  moId: string | null;
  
  // Follow-Up Status
  followUpStatus: FollowUpStatus;
  followUpStatusBadge: StatusBadgeViewModel;
  daysRemaining: number | null;
  hasActiveFollowUp: boolean;
  canBookFreeFollowUp: boolean;
  
  // Quick Actions
  canBookAppointment: boolean;
  canViewHistory: boolean;
  hasActiveAppointment: boolean;
}

export enum FollowUpStatus {
  NONE = 'none',
  ACTIVE = 'active',
  USED = 'used',
  EXPIRED = 'expired',
  RENEWED = 'renewed'
}

export interface StatusBadgeViewModel {
  label: string;        // "Active", "Used", etc.
  color: string;       // "#10B981"
  bgColor: string;    // "#D1FAE5"
  icon: string;       // "check-circle", "clock", etc.
}

// Component Usage
export const PatientListComponent = () => {
  const viewModel = usePatientListViewModel();
  
  return (
    <div>
      {/* Search Bar */}
      <SearchBar 
        value={viewModel.searchQuery}
        onChange={handleSearch}
        placeholder="Search by name, phone, or MO ID"
      />
      
      {/* Filter Toggle */}
      <Toggle 
        label="Active Only"
        checked={viewModel.filterActive}
        onChange={handleToggleFilter}
      />
      
      {/* Patient Cards */}
      {viewModel.patients.map(patient => (
        <PatientCard key={patient.id} patient={patient} />
      ))}
    </div>
  );
};
```

### **2. Patient Card Component**

```typescript
// PatientCardViewModel.ts
export interface PatientCardViewModel {
  // Basic Info
  id: string;
  name: string;
  phone: string;
  moId: string | null;
  
  // Status Badge
  statusBadge: {
    text: string;          // "Active Follow-Up"
    color: string;          // "#10B981"
    backgroundColor: string; // "#D1FAE5"
    daysRemaining: number | null;
  };
  
  // Actions
  actions: {
    onBookAppointment: () => void;
    onViewHistory: () => void;
    onEditPatient: () => void;
  };
}

// PatientCard.tsx
export const PatientCard = ({ patient }: { patient: PatientCardViewModel }) => {
  return (
    <Card>
      <CardHeader>
        <Avatar>{patient.name.charAt(0)}</Avatar>
        <div>
          <Text bold>{patient.name}</Text>
          <Text size="sm" color="gray">{patient.phone}</Text>
          {patient.moId && <Badge>{patient.moId}</Badge>}
        </div>
      </CardHeader>
      
      <CardBody>
        {/* Follow-Up Status */}
        <StatusBadge {...patient.statusBadge} />
        
        {/* Days Remaining */}
        {patient.statusBadge.daysRemaining !== null && (
          <Text size="sm">
            {patient.statusBadge.daysRemaining} days remaining
          </Text>
        )}
      </CardBody>
      
      <CardFooter>
        <Button onClick={patient.actions.onBookAppointment}>
          Book Appointment
        </Button>
        <Button variant="outline" onClick={patient.actions.onViewHistory}>
          View History
        </Button>
      </CardFooter>
    </Card>
  );
};
```

### **3. Appointment Booking Form View Model**

```typescript
// AppointmentBookingViewModel.ts
export interface AppointmentBookingViewModel {
  // Step 1: Patient Selection
  selectedPatient: PatientViewModel | null;
  availablePatients: PatientViewModel[];
  
  // Step 2: Doctor & Department Selection
  selectedDoctor: DoctorViewModel | null;
  selectedDepartment: DepartmentViewModel | null;
  availableDoctors: DoctorViewModel[];
  availableDepartments: DepartmentViewModel[];
  
  // Step 3: Date & Time Selection
  selectedDate: string | null;         // YYYY-MM-DD
  selectedTime: string | null;          // HH:MM
  selectedSlot: SlotViewModel | null;
  availableSlots: SlotViewModel[];
  
  // Step 4: Follow-Up Eligibility
  followUpEligibility: FollowUpEligibilityViewModel | null;
  
  // Step 5: Payment
  consultationType: ConsultationTypeViewModel;
  paymentMethod: PaymentMethodViewModel;
  
  // Form State
  currentStep: number;
  loading: boolean;
  error: string | null;
}

export interface FollowUpEligibilityViewModel {
  isFree: boolean;
  isEligible: boolean;
  daysRemaining: number | null;
  validUntil: string;                   // YYYY-MM-DD
  message: string;
}

export interface ConsultationTypeViewModel {
  type: 'clinic_visit' | 'video_consultation' | 'follow-up-via-clinic' | 'follow-up-via-video';
  label: string;
  icon: string;
  color: string;
}

export interface SlotViewModel {
  id: string;
  time: string;                         // "10:30"
  availableCount: number;
  maxPatients: number;
  isAvailable: boolean;
}

// Component
export const AppointmentBookingForm = () => {
  const viewModel = useAppointmentBookingViewModel();
  
  return (
    <Stepper currentStep={viewModel.currentStep}>
      {/* Step 1: Patient */}
      <PatientSelector /> 
      
      {/* Step 2: Doctor & Department */}
      <DoctorSelector />
      
      {/* Step 3: Date & Time */}
      <DateSlotSelector />
      
      {/* Step 4: Follow-Up Check */}
      {viewModel.followUpEligibility && (
        <FollowUpEligibilityCard 
          eligibility={viewModel.followUpEligibility}
          onUseFreeFollowUp={() => selectConsultationType('follow-up-via-clinic')}
        />
      )}
      
      {/* Step 5: Payment */}
      <PaymentForm />
      
      <Button onClick={handleSubmit}>Book Appointment</Button>
    </Stepper>
  );
};
```

### **4. Follow-Up Eligibility Card**

```typescript
// FollowUpEligibilityCard.tsx
interface FollowUpEligibilityViewModel {
  isFree: boolean;
  isEligible: boolean;
  daysRemaining: number | null;
  validUntil: string;
  message: string;
}

export const FollowUpEligibilityCard = ({ 
  eligibility,
  onUseFreeFollowUp 
}: {
  eligibility: FollowUpEligibilityViewModel;
  onUseFreeFollowUp: () => void;
}) => {
  if (!eligibility.isEligible) {
    return null;
  }
  
  if (eligibility.isFree && eligibility.isEligible) {
    return (
      <Alert type="success">
        <AlertIcon icon="check-circle" />
        <AlertTitle>Free Follow-Up Available!</AlertTitle>
        <AlertDescription>
          You have a FREE follow-up appointment available.
          {eligibility.daysRemaining && `${eligibility.daysRemaining} days remaining.`}
        </AlertDescription>
        <Button onClick={onUseFreeFollowUp} colorScheme="green">
          Use Free Follow-Up
        </Button>
      </Alert>
    );
  }
  
  return (
    <Alert type="info">
      <AlertIcon icon="info" />
      <AlertTitle>Follow-Up Available</AlertTitle>
      <AlertDescription>
        You can book a follow-up appointment (payment required).
      </AlertDescription>
    </Alert>
  );
};
```

---

## 📤 **Upload & Reset Functions**

### **1. Patient Data Upload (CSV/Excel)**

```typescript
// PatientDataUploadViewModel.ts
export interface PatientUploadViewModel {
  file: File | null;
  parsedData: PatientRow[];
  errors: UploadError[];
  mapping: FieldMapping;
  preview: PatientViewModel[];
}

export interface PatientRow {
  rawData: Record<string, any>;
  mappedData: PatientViewModel | null;
  errors: string[];
}

export interface FieldMapping {
  [key: string]: string;  // e.g., "Full Name" => "name"
}

export interface UploadError {
  row: number;
  field: string;
  message: string;
}

// Usage
export const PatientUploadComponent = () => {
  const viewModel = usePatientUploadViewModel();
  
  const handleFileUpload = (file: File) => {
    viewModel.parseFile(file);
    viewModel.previewData();
  };
  
  const handleSave = async () => {
    const validPatients = viewModel.preview.filter(p => !p.errors.length);
    await uploadPatients(validPatients);
  };
  
  return (
    <div>
      <FileUploader 
        accept=".csv,.xlsx"
        onFileSelect={handleFileUpload}
      />
      
      <FieldMappingTable 
        availableFields={viewModel.availableFields}
        mapping={viewModel.mapping}
        onMappingChange={handleMappingChange}
      />
      
      <DataPreview 
        data={viewModel.preview}
        errors={viewModel.errors}
      />
      
      <Button onClick={handleSave} disabled={viewModel.errors.length > 0}>
        Upload {viewModel.preview.length} Patients
      </Button>
    </div>
  );
};
```

### **2. Appointment Data Upload**

```typescript
// AppointmentUploadViewModel.ts
export interface AppointmentUploadViewModel {
  file: File | null;
  parsedData: AppointmentRow[];
  errors: UploadError[];
  preview: AppointmentViewModel[];
  duplicates: string[];            // Existing booking numbers
}

export interface AppointmentRow {
  bookingNumber: string;
  patientId: string;
  doctorId: string;
  appointmentDate: string;
  appointmentTime: string;
  errors: string[];
}

// Component
export const AppointmentUploadComponent = () => {
  const viewModel = useAppointmentUploadViewModel();
  
  return (
    <div>
      {/* File Upload */}
      <FileUploader onFileSelect={handleFileUpload} />
      
      {/* Preview Table */}
      <DataTable 
        data={viewModel.preview}
        columns={['Booking #', 'Patient', 'Doctor', 'Date', 'Time']}
        onRemoveRow={handleRemoveRow}
      />
      
      {/* Duplicate Warning */}
      {viewModel.duplicates.length > 0 && (
        <Alert type="warning">
          {viewModel.duplicates.length} appointments with duplicate booking numbers found.
          These will be skipped.
        </Alert>
      )}
      
      <Button onClick={handleSave}>
        Upload {viewModel.preview.length - viewModel.duplicates.length} Appointments
      </Button>
    </div>
  );
};
```

### **3. Reset Functions**

```typescript
// ResetViewModel.ts
export interface ResetViewModel {
  clinicId: string;
  resetType: 'appointments' | 'patients' | 'followups' | 'all';
  resetOptions: {
    deleteAppointments: boolean;
    deletePatients: boolean;
    resetFollowUps: boolean;
    resetStatus: boolean;
  };
  confirmPassword: string;
}

export const ResetComponent = () => {
  const viewModel = useResetViewModel();
  
  const handleReset = async () => {
    if (!viewModel.confirmPassword) {
      showError("Please enter your password to confirm");
      return;
    }
    
    await resetData({
      clinicId: viewModel.clinicId,
      resetType: viewModel.resetType,
      options: viewModel.resetOptions,
      password: viewModel.confirmPassword
    });
  };
  
  return (
    <Modal>
      <ModalHeader>Reset Data</ModalHeader>
      <ModalBody>
        <Text>Select what to reset:</Text>
        
        <CheckboxGroup>
          <Checkbox 
            checked={viewModel.resetOptions.deleteAppointments}
            onChange={handleOptionChange('deleteAppointments')}
          >
            Delete All Appointments
          </Checkbox>
          
          <Checkbox 
            checked={viewModel.resetOptions.deletePatients}
            onChange={handleOptionChange('deletePatients')}
          >
            Delete All Patients
          </Checkbox>
          
          <Checkbox 
            checked={viewModel.resetOptions.resetFollowUps}
            onChange={handleOptionChange('resetFollowUps')}
          >
            Reset Follow-Ups
          </Checkbox>
          
          <Checkbox 
            checked={viewModel.resetOptions.resetStatus}
            onChange={handleOptionChange('resetStatus')}
          >
            Reset Status Fields
          </Checkbox>
        </CheckboxGroup>
        
        <Input 
          type="password"
          placeholder="Enter password to confirm"
          value={viewModel.confirmPassword}
          onChange={handlePasswordChange}
        />
      </ModalBody>
      
      <ModalFooter>
        <Button colorScheme="red" onClick={handleReset}>
          Confirm Reset
        </Button>
        <Button variant="ghost" onClick={handleCancel}>
          Cancel
        </Button>
      </ModalFooter>
    </Modal>
  );
};
```

---

## 🔄 **State Management**

### **Redux Store Structure**

```typescript
// store.ts
export interface AppState {
  auth: AuthState;
  patients: PatientState;
  appointments: AppointmentState;
  followUps: FollowUpState;
}

export interface PatientState {
  items: ClinicPatient[];
  selectedPatient: ClinicPatient | null;
  searchQuery: string;
  filters: {
    onlyActive: boolean;
    clinicId: string;
  };
  loading: boolean;
  error: string | null;
}

export interface AppointmentState {
  items: Appointment[];
  selectedAppointment: Appointment | null;
  filters: {
    clinicId: string;
    doctorId: string;
    date: string;
  };
  loading: boolean;
  error: string | null;
}

export interface FollowUpState {
  activeFollowUps: FollowUp[];
  eligibility: {
    patientId: string;
    isFree: boolean;
    isEligible: boolean;
  } | null;
}
```

### **API Service Layer**

```typescript
// api/patientApi.ts
export class PatientApiService {
  async getPatients(params: GetPatientsParams): Promise<PatientResponse> {
    const response = await fetch(
      `/api/organizations/clinic-specific-patients?${buildQueryString(params)}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`
        }
      }
    );
    return response.json();
  }
  
  async createPatient(data: CreatePatientRequest): Promise<Patient> {
    const response = await fetch('/api/organizations/clinic-specific-patients', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    return response.json();
  }
  
  async uploadPatients(file: File): Promise<UploadResult> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await fetch('/api/organizations/clinic-specific-patients/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`
      },
      body: formData
    });
    return response.json();
  }
}

// api/appointmentApi.ts
export class AppointmentApiService {
  async createAppointment(data: CreateAppointmentRequest): Promise<Appointment> {
    const response = await fetch('/api/appointments/simple', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    return response.json();
  }
  
  async checkFollowUpEligibility(params: FollowUpEligibilityParams): Promise<FollowUpEligibility> {
    const response = await fetch(
      `/api/appointments/check-follow-up-eligibility?${buildQueryString(params)}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`
        }
      }
    );
    return response.json();
  }
}
```

---

## 📱 **Complete UI Flow Example**

```typescript
// App.tsx - Complete Flow
export const App = () => {
  return (
    <Router>
      <Routes>
        {/* Login */}
        <Route path="/login" element={<LoginPage />} />
        
        {/* Dashboard */}
        <Route path="/dashboard" element={<DashboardPage />}>
          {/* Patient Management */}
          <Route path="patients" element={<PatientListPage />} />
          <Route path="patients/new" element={<CreatePatientPage />} />
          <Route path="patients/:id" element={<PatientDetailPage />} />
          <Route path="patients/:id/edit" element={<EditPatientPage />} />
          <Route path="patients/upload" element={<PatientUploadPage />} />
          
          {/* Appointment Management */}
          <Route path="appointments" element={<AppointmentListPage />} />
          <Route path="appointments/new" element={<CreateAppointmentPage />} />
          <Route path="appointments/:id" element={<AppointmentDetailPage />} />
          <Route path="appointments/upload" element={<AppointmentUploadPage />} />
          
          {/* Settings */}
          <Route path="settings" element={<SettingsPage />} />
          <Route path="settings/reset" element={<ResetDataPage />} />
        </Route>
      </Routes>
    </Router>
  );
};

// PatientListPage.tsx
export const PatientListPage = () => {
  const viewModel = usePatientListViewModel();
  
  // Fetch patients on mount
  useEffect(() => {
    viewModel.loadPatients();
  }, []);
  
  return (
    <PageLayout>
      <PageHeader>
        <Text fontSize="2xl" fontWeight="bold">Patients</Text>
        <ButtonGroup>
          <Button onClick={viewModel.handleCreateNew}>
            Create New Patient
          </Button>
          <Button onClick={viewModel.handleUpload}>
            Upload CSV
          </Button>
        </ButtonGroup>
      </PageHeader>
      
      {/* Search & Filter */}
      <PatientSearchBar 
        query={viewModel.searchQuery}
        onChange={viewModel.handleSearch}
        filterActive={viewModel.filterActive}
        onToggleFilter={viewModel.handleToggleFilter}
      />
      
      {/* Patient Grid */}
      <SimpleGrid columns={3} spacing={4}>
        {viewModel.patients.map(patient => (
          <PatientCard 
            key={patient.id}
            patient={patient}
            onSelect={viewModel.handleSelectPatient}
          />
        ))}
      </SimpleGrid>
      
      {viewModel.loading && <Spinner />}
      {viewModel.error && <Alert type="error">{viewModel.error}</Alert>}
    </PageLayout>
  );
};
```

---

## 🎨 **UI Component Library**

### **Status Badge Component**

```typescript
export const StatusBadge = ({ 
  status, 
  showDaysRemaining 
}: { 
  status: FollowUpStatus;
  showDaysRemaining?: boolean;
}) => {
  const config = STATUS_CONFIG[status];
  
  return (
    <Badge 
      colorScheme={config.colorScheme}
      variant="subtle"
    >
      <BadgeIcon icon={config.icon} />
      {config.label}
      {showDaysRemaining && config.daysRemaining && (
        <Text ml={2} size="xs">
          ({config.daysRemaining} days)
        </Text>
      )}
    </Badge>
  );
};

const STATUS_CONFIG = {
  [FollowUpStatus.NONE]: {
    label: 'No Follow-Up',
    colorScheme: 'gray',
    icon: 'info'
  },
  [FollowUpStatus.ACTIVE]: {
    label: 'Active',
    colorScheme: 'green',
    icon: 'check-circle'
  },
  [FollowUpStatus.USED]: {
    label: 'Used',
    colorScheme: 'blue',
    icon: 'done'
  },
  [FollowUpStatus.EXPIRED]: {
    label: 'Expired',
    colorScheme: 'red',
    icon: 'warning'
  },
  [FollowUpStatus.RENEWED]: {
    label: 'Renewed',
    colorScheme: 'purple',
    icon: 'refresh'
  }
};
```

---

## 🎉 **Complete Documentation!**

This document provides:
- ✅ View Models
- ✅ UI Components
- ✅ Upload Functions
- ✅ Reset Functions
- ✅ State Management
- ✅ API Services
- ✅ Complete UI Flow

Your frontend team can now implement the entire system! 🚀

