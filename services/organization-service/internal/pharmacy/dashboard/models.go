package dashboard

type SummaryResponse struct {
	DailySales        float64       `json:"daily_sales"`
	SalesVolume       int           `json:"sales_volume"`
	NewPatients       int           `json:"new_patients"`
	TotalPatients     int           `json:"total_patients"`
	RecurringPatients int           `json:"recurring_patients"`
	RecurringSales    float64       `json:"recurring_sales"`
	TrendAmounts      []float64     `json:"trend_amounts"`
	TrendDates        []string      `json:"trend_dates"`
	TotalMedicines    int           `json:"total_medicines"`
	TotalStockValue   float64       `json:"total_stock_value"`
	ExpiredStockCount int           `json:"expired_stock_count"`
	ExpiringSoonCount int           `json:"expiring_soon_count"`
	ExpiringSoonValue float64       `json:"expiring_soon_value"`
	HighRiskCount     int           `json:"high_risk_count"`
	TotalSuppliers    int           `json:"total_suppliers"`
	SupplierPayment   float64       `json:"supplier_payment"`
	SupplierDue       float64       `json:"supplier_due"`
	ExpiringBatches   []interface{} `json:"expiring_batches"`
}
