package models

import "time"

type StockIn struct {
	ID          string    `json:"id"`
	PharmacyID  string    `json:"pharmacy_id"`
	SupplierID  string    `json:"supplier_id"`
	TotalAmount float64   `json:"total_amount"`
	DateIn      time.Time `json:"date_in"`
}
