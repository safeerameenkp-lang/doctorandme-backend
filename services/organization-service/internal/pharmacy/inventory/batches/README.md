# Batch Management & Inventory Engine

## Overview
The `batches` module is the "Live Engine" of the pharmaceutical inventory system. It maintains the real-time state of all medicinal stock, categorized by individual batch identifiers. This module ensures that the pharmacy operates on a **FEFO (First Expired, First Out)** principle by providing deep visibility into batch lifecycles.

## Directory Structure
```text
batches/
├── handler.go       # HTTP interface for live inventory queries
├── service.go       # Inventory business logic & Stock orchestration
├── repository.go    # High-performance atomic Upsert logic (PostgreSQL)
└── models.go        # Batch entities & Update DTOs
```

## Core Responsibilities
- **Live Stock Tracking**: Maintaining accurate counts of individual tablets, bottles, and pieces.
- **Expiry Monitoring**: Tracking `expiry_date` at the batch level to trigger early warning alerts.
- **Profitability Mapping**: Storing the specific `cost_price` and `MRP` for every unique batch to ensure precise margin tracking.
- **Physical Mapping**: Syncing digital inventory with physical storage using `rack_no` assignments.

## The "Upsert" Protocol
The Batch module utilizes a sophisticated **Upsert (Update or Insert)** mechanism to handle stock-in transactions:

1. **Identification**: Checks for an existing record matching `TenantID` + `MedicineID` + `BatchNo`.
2. **Dynamic Update**: If a match is found, the engine:
   - Increments the `quantity_available`.
   - Updates the Pricing metadata (MRP, Cost) to the latest invoice values.
   - Refreshes the `rack_no` location.
3. **Atomic Initialization**: If no match exists, a new UUID-gated batch record is created.

## Key Calculations
- **Unit Price Derivation**: MRP and Cost are stored at the **Base Unit** level (per tablet/piece) to ensure consistency across different dispensing modes.
- **Inventory valuation**: Aggregates batch quantities and unit MRPs to provide real-time dashboard analytics.

## Technical Specifications
- **Indexing**: Optimized using GIN and composite B-Tree indexes on `(tenant_id, medicine_id, batch_no)` for sub-millisecond query performance.
- **Concurrency**: Leverages database-level transactions (SQL Tx) to ensure stock totals never drift during simultaneous stock-in events.

## API Integration
- `GET /inventory/batches`: Returns the live inventory list for the tenant.
- `GET /inventory/batches?medicine_id={uuid}`: Returns all active batches for a specific pharmaceutical product.
