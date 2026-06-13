# Medicine Master - Domain Documentation

This module serves as the **Medicine Master Data (Catalog)**. It is the core source of truth for all pharmaceutical products within a specific pharmacy's inventory.

## 💊 Domain Concepts

### 1. Taxation (GST)
In the context of Indian Pharmacy SaaS, tax handling is critical.
- **CGST**: Central Goods and Services Tax.
- **SGST**: State Goods and Services Tax.
- *Logic*: These rates are stored at the product master level so that they can be automatically fetched during Sales/Billing.

### 2. HSN Code
Every medicine is assigned an **Harmonized System of Nomenclature** code. This is mandatory for legal compliance and tax reporting in international and local trade.

### 3. Schedule Type
Categorization of drugs based on their therapeutic effects and legal restrictions (e.g., Schedule H, H1, X).
- **Rx Required**: A boolean flag that ensures the frontend/POS system blocks the sale if a prescription isn't provided for scheduled drugs.

### 4. Unit & Dosage Management
Supports various forms of medication:
- **Dosage Forms**: Tablets, Capsules, Syrups, Injections, Ointments.
- **Unit Types**: Strips, Boxes, Bottles, Vials, Pieces.

## 🛡️ Business Rules & Validation

### Supplier Ownership Check
A medicine **must** be linked to a valid supplier. The backend performs an ownership check:
- Is the `supplier_id` valid?
- Does the supplier belong to the same `tenant_id`?
- Is the supplier currently `active`?
If any of these fail, the record is rejected to prevent data corruption.

### Duplicate Prevention Logic
We use a **Case-Insensitive Multi-Column Constraint**. A duplicate is detected if the following match (ignoring case):
`TenantID + Name + BrandName + DosageForm + UnitType + SupplierID`

## 📊 Database Schema Details

**Table**: `inventory.medicines`

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key |
| `tenant_id` | UUID | Data Isolation Key |
| `name` | VARCHAR | Generic name of the medicine |
| `brand_name` | VARCHAR | Commercial brand name |
| `dosage_form` | VARCHAR | e.g. "Tablet", "Liquid" |
| `supplier_id` | UUID | Link to validated supplier |
| `is_rx_required` | BOOLEAN | Prescription compliance flag |

### Critical Indexes
- `idx_medicines_unique_case_insensitive`: Ensures absolute uniqueness across critical identifying fields.
- `idx_medicines_barcode`: Fast lookup for pharmacy scanners.

## 📡 API Endpoints

| Method | Endpoint | Function |
| :--- | :--- | :--- |
| `POST` | `/inventory/medicines/` | **Register**: Create a new medicine master record. |
| `GET` | `/inventory/medicines/` | **List/Search**: Fetch medicines with optional search & filters. |
| `GET` | `/inventory/medicines/{id}` | **Retrieve**: Get full details of a specific medicine by ID. |
| `PUT` | `/inventory/medicines/{id}` | **Update**: Modify existing medicine details. |
| `PATCH` | `/inventory/medicines/{id}/status` | **Status Toggle**: Activate or deactivate a medicine record. |

---
*This module is designed for accuracy, legal compliance, and rapid pharmaceutical operations.*
