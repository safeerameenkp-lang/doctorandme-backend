# Stock-In & Procurement Engine

## Overview
The `stockin` module is the primary gateway for inventory ingestion within the Pharmacy SaaS platform. It handles the complex business logic of converting supplier invoices into live medicinal batches, ensuring financial integrity and pharmaceutical standard compliance.

## Directory Structure
```text
stockin/
├── handler.go       # HTTP Entry points & Request decoding
├── service.go       # Core Business Logic & 11 Validation Rules
├── repository.go    # Database mapping layer (Purchases & Items)
├── models.go        # Request/Response DTOs & DB Entities
└── README.md        # Technical Documentation
```

## Core Responsibilities
- **Invoice Harmonization**: Validating physical supplier bills against digital records.
- **Inventory Generation**: Orchestrating the creation and incrementing of medicine batches.
- **Financial Control**: Enforcing strict "Manual Entry vs System Calculation" parity to prevent accounting discrepancies.
- **Data Isolation**: Guaranteeing that procurement data is strictly bounded by `TenantID`.

## Business Logic & Validation Rules
The module implements the following "11 Critical Rules" to ensure operational safety:

### 1. Security & Ownership
- **Supplier Verification**: Only suppliers belonging to the active pharmacy (Tenant) can be selected.
- **Bill Idempotency**: Prevents duplicate invoice entry for the same supplier to avoid inventory inflation.
- **Medicine Status**: Only `IsActive` medicines can be procured.

### 2. Pharmaceutical Unit Standards
- **Fetch from Master**: Packaging modes (Strip, Box, etc.) are dictated by the Medicine Master table.
- **Mode Locking**: Non-"Strip" modes are hard-locked to a multiplier of 1 to ensure unit count accuracy.
- **Base Unit Mapping**: Automatic conversion of packaging types to granular units (tablets, bottles, pieces).

### 3. Inventory Calculation Engine
- **Stock Multiplier**: Uses the formula `(ReceivedQty + BonusQty) * UnitsPerPack` for inventory updates.
- **Landing Cost Derivation**: Calculates the true landing cost per unit by prioritizing Net Total over raw purchase price, independent of tax/discount variables.

### 4. Financial Integrity
- **Grand Total Guard**: Rejects transactions if the manual grand total header differs from the sum of item net amounts (Tolerance: 0.01).
- **Payment Status Automation**: Dynamically assigns `PENDING`, `PARTIAL`, or `COMPLETED` status based on the payment vs. total delta.

## Technical Architecture

### Workflow Details
1. **Entry**: Receives a `CreatePurchaseRequest` via the API.
2. **Validation**: Parallel checks for supplier ownership and invoice uniqueness.
3. **Drafting**: Iterates through purchase items to perform derived cost calculations.
4. **Transaction**: Initiates an atomic PostgreSQL transaction.
5. **Upsert**: Calls the `Batches` service to increment stock or create new batch records.
6. **Finalization**: Records the Purchase Header and commits the transaction.

### Data Structures
- `Purchase`: Represents the overall invoice (Header).
- `PurchaseItem`: Represents individual medicine entries within an invoice.
- `UpdateBatchDTO`: The payload transformed for the Batch Engine.

## API Endpoints
- `POST /inventory/stock-in`: Create a new purchase and update inventory.
- `GET /inventory/stock-in`: List all procurement history for the tenant.
- `GET /inventory/stock-in/{id}`: Detailed view of a specific invoice and its items.
