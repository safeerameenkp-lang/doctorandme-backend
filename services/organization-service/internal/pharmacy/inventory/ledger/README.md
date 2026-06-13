# Inventory Ledger Module

## Overview
The **Inventory Ledger** is the central accountability layer for the Pharmacy SaaS application. It acts as an immutable audit log, recording every stock movement (Purchase, Sale, Return, Adjustment) across all tenants. 

Its primary purpose is to ensure **medical and financial traceability**, providing a point-in-time snapshot of inventory levels.

---

## 🛠 Core Features

### 1. Atomic Integrity
The Ledger is designed to work within parent transactions. It uses a synchronous update pattern:
- **No Ghost Movements**: Stock is only deducted/added if the Ledger record is successfully committed.
- **ACID Compliant**: Integrated with the `billing-service` and `stock-in` flows.

### 2. Concurrency Control (Safe-by-Design)
To handle high-traffic pharmacies with multiple staff members:
- **Pessimistic Locking**: Every movement triggers a `SELECT ... FOR UPDATE` on the specific medicine batch.
- **Race Condition Prevention**: Prevents "double-deduction" or negative stock scenarios.

### 3. Financial Snapshots
Every entry stores the `balance_after` quantity. This allows for:
- Historic stock reconstruction.
- Discrepancy detection between digital records and physical counts.
- Seamless end-of-day auditing.

---

## 🏗 Data Structure

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | `UUID` | Primary Key. |
| `transaction_type` | `Enum` | `PURCHASE`, `SALE`, `SALE_RETURN`, `ADJUSTMENT`. |
| `quantity_change` | `Int` | Positive (IN) or Negative (OUT). |
| `balance_after` | `Int` | Final quantity after the movement. |
| `reference_id` | `UUID` | ID of the source Invoice or Purchase Order. |
| `notes` | `NullString` | Optional auditor notes. |

---

## 🚀 API Endpoints

### Get Ledger by Batch
`GET /inventory/ledger/batch/{batchId}`
Retrieves the full movement history for a specific batch. Useful for tracking expiry-related issues.

### Get Ledger by Medicine
`GET /inventory/ledger/medicine/{medicineId}`
Retrieves the movement history for a specific drug across all its batches.

---

## 🔧 Technical Notes

- **Schema**: Located in `inventory.stock_ledger` table.
