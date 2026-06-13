package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	client *http.Client
}

func NewHandler() *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *Handler) GetSummary(c *gin.Context) {
	token := middleware.GetRawToken(c.Request.Context())
	pharmacyID := c.GetString("pharmacy_id")

	// Get trend parameters from query
	granularity := c.Query("granularity")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	date := c.Query("date")
	expiryFilter := c.Query("expiryFilter")
	if expiryFilter == "" {
		expiryFilter = "expiring"
	}

	salesParams := fmt.Sprintf("?granularity=%s&startDate=%s&endDate=%s&date=%s", granularity, startDate, endDate, date)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	baseURL := fmt.Sprintf("http://localhost:%s/api/pharmacy", port)

	var wg sync.WaitGroup
	wg.Add(3)

	var salesData map[string]interface{}
	var invData map[string]interface{}
	var supData map[string]interface{}

	// Fetch Sales Stats
	go func() {
		defer wg.Done()
		salesData = h.fetchStats(baseURL+"/sales/stats"+salesParams, token, pharmacyID)
	}()

	// Fetch Inventory Stats (Purchase Financials)
	go func() {
		defer wg.Done()
		invData = h.fetchStats(baseURL+"/inventory/stock-in/stats", token, pharmacyID)
	}()

	// Fetch Supplier Stats
	go func() {
		defer wg.Done()
		supData = h.fetchStats(baseURL+"/supplier/stats", token, pharmacyID)
	}()

	wg.Wait()

	resp := SummaryResponse{}

	// Map Sales
	if d, ok := salesData["data"].(map[string]interface{}); ok {
		resp.DailySales, _ = d["daily_sales"].(float64)
		if val, ok := d["sales_volume"].(float64); ok {
			resp.SalesVolume = int(val)
		}
		if val, ok := d["new_patients"].(float64); ok {
			resp.NewPatients = int(val)
		}
		if val, ok := d["total_patients"].(float64); ok {
			resp.TotalPatients = int(val)
		}
		if val, ok := d["recurring_patients"].(float64); ok {
			resp.RecurringPatients = int(val)
		}
		resp.RecurringSales, _ = d["recurring_sales"].(float64)

		// Map Trend
		if amounts, ok := d["trend_amounts"].([]interface{}); ok {
			for _, a := range amounts {
				if val, ok := a.(float64); ok {
					resp.TrendAmounts = append(resp.TrendAmounts, val)
				}
			}
		}
		if dates, ok := d["trend_dates"].([]interface{}); ok {
			for _, dVal := range dates {
				if val, ok := dVal.(string); ok {
					resp.TrendDates = append(resp.TrendDates, val)
				}
			}
		}
	}

	// Map Inventory (Stock In Financials)
	if d, ok := invData["data"].(map[string]interface{}); ok {
		resp.SupplierPayment, _ = d["paid_amount"].(float64)
		resp.SupplierDue, _ = d["due_amount"].(float64)
	}

	// Fetch Medicine/Batch stats for the other cards
	batchStats := h.fetchStats(baseURL+"/inventory/batches/stats", token, pharmacyID)
	if d, ok := batchStats["data"].(map[string]interface{}); ok {
		if val, ok := d["total_stocks"].(float64); ok {
			resp.TotalMedicines = int(val)
		}
		resp.TotalStockValue, _ = d["total_stock_value"].(float64)
		if val, ok := d["expired_stock"].(float64); ok {
			resp.ExpiredStockCount = int(val)
		}
		if val, ok := d["expiring_soon"].(float64); ok {
			resp.ExpiringSoonCount = int(val)
		}
		resp.ExpiringSoonValue, _ = d["expiring_soon_value"].(float64)
		if val, ok := d["high_risk_count"].(float64); ok {
			resp.HighRiskCount = int(val)
		}
	}

	// Fetch Expiring Batches
	expiringBatches := h.fetchStats(baseURL+"/inventory/batches?limit=5&offset=0&filter="+expiryFilter, token, pharmacyID)
	if d, ok := expiringBatches["data"].([]interface{}); ok {
		resp.ExpiringBatches = d
	}

	// Map Suppliers
	if d, ok := supData["data"].(map[string]interface{}); ok {
		if val, ok := d["total_suppliers"].(float64); ok {
			resp.TotalSuppliers = int(val)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

func (h *Handler) fetchStats(url string, token string, pharmacyID string) map[string]interface{} {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", url, err)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if pharmacyID != "" {
		req.Header.Set("X-Pharmacy-ID", pharmacyID)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("Error fetching stats from %s: %v", url, err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response from %s: %v", url, err)
		return nil
	}
	return result
}
