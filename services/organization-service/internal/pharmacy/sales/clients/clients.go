package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"organization-service/internal/pharmacy/sales/prescriptions"
	"organization-service/middleware"

	"github.com/google/uuid"
)

type Reservation struct {
	Quantity int `json:"quantity"`
}

type ReturnItemRequest struct {
	BatchID  uuid.UUID `json:"batch_id"`
	Quantity int       `json:"quantity"`
	Reason   string    `json:"reason"`
}

type StockAvailability struct {
	BatchID            uuid.UUID `json:"id"`
	BatchNo            string    `json:"batch_no"`
	MedicineName       string    `json:"medicine_name"`
	MedicineBrand      string    `json:"medicine_brand"`
	Quantity           int       `json:"quantity_available"`
	MRP                float64   `json:"mrp"`
	UnitPrice          float64   `json:"unit_price"`
	ExpiryDate         time.Time `json:"expiry_date"`
	CGSTRate           float64   `json:"cgst_rate"`
	SGSTRate           float64   `json:"sgst_rate"`
	TotalTaxPercentage float64   `json:"total_tax_percentage"`
	RetailDiscPerc     float64   `json:"retail_disc_perc"`
	StaffDiscPerc      float64   `json:"staff_disc_perc"`
	SpecialDiscPerc    float64   `json:"special_disc_perc"`
	MaxDiscPerc        float64   `json:"max_disc_perc"`
	RackNo             string    `json:"rack_no"`
}

type InventoryClient interface {
	GetAvailability(ctx context.Context, pharmacyID, productID uuid.UUID) ([]StockAvailability, error)
	ReserveStock(ctx context.Context, pharmacyID, productID, batchID uuid.UUID, quantity int) (string, error)
	UpdateReservation(ctx context.Context, pharmacyID uuid.UUID, reservationID string, newQuantity int) error
	ConfirmStock(ctx context.Context, pharmacyID uuid.UUID, reservationID string) error
	ReleaseStock(ctx context.Context, pharmacyID uuid.UUID, reservationID string) error
	ReturnItems(ctx context.Context, pharmacyID uuid.UUID, items []ReturnItemRequest) error
}

type httpInventoryClient struct {
	baseURL string
	client  *http.Client
}

func NewInventoryClient(baseURL string) InventoryClient {
	return &httpInventoryClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *httpInventoryClient) GetAvailability(ctx context.Context, pharmacyID, productID uuid.UUID) ([]StockAvailability, error) {
	url := fmt.Sprintf("%s/batches?medicine_id=%s", c.baseURL, productID)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service error: %d", resp.StatusCode)
	}

	var result struct {
		Data []StockAvailability `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *httpInventoryClient) ReserveStock(ctx context.Context, pharmacyID, productID, batchID uuid.UUID, quantity int) (string, error) {
	url := fmt.Sprintf("%s/reservations", c.baseURL)
	body, _ := json.Marshal(map[string]interface{}{
		"product_id": productID,
		"batch_id":   batchID,
		"quantity":   quantity,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return "", fmt.Errorf("inventory service error: %s", errResp.Error)
		}
		return "", fmt.Errorf("inventory service error: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.ID, nil
}

func (c *httpInventoryClient) UpdateReservation(ctx context.Context, pharmacyID uuid.UUID, reservationID string, newQuantity int) error {
	url := fmt.Sprintf("%s/reservations/%s", c.baseURL, reservationID)
	body, _ := json.Marshal(map[string]interface{}{"quantity": newQuantity})
	req, _ := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory service update error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpInventoryClient) ConfirmStock(ctx context.Context, pharmacyID uuid.UUID, reservationID string) error {
	url := fmt.Sprintf("%s/reservations/%s/confirm", c.baseURL, reservationID)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory service confirm error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpInventoryClient) ReleaseStock(ctx context.Context, pharmacyID uuid.UUID, reservationID string) error {
	url := fmt.Sprintf("%s/reservations/%s", c.baseURL, reservationID)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("inventory service release error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpInventoryClient) ReturnItems(ctx context.Context, pharmacyID uuid.UUID, items []ReturnItemRequest) error {
	url := fmt.Sprintf("%s/batches/return", c.baseURL)
	body, _ := json.Marshal(map[string]interface{}{"items": items})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-ID", pharmacyID.String())
	req.Header.Set("Authorization", "Bearer "+middleware.GetRawToken(ctx))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory service return error: status %d", resp.StatusCode)
	}
	return nil
}

// Prescription Client

type PrescriptionItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	Name         string    `json:"medicine_name"`
	Brand        string    `json:"medicine_brand"`
	Quantity     int       `json:"quantity"`
	DurationDays int       `json:"duration_days"`
	DosagePerDay float64   `json:"dosage_per_day"`
	Morning      float64   `json:"morning"`
	Noon         float64   `json:"noon"`
	Night        float64   `json:"night"`
	Instructions string    `json:"instructions"`
}

type PrescriptionData struct {
	ID             string             `json:"id"`
	PatientName    string             `json:"patient_name"`
	PatientPhone   string             `json:"patient_phone"`
	DoctorName     string             `json:"doctor_name"`
	TokenNo        string             `json:"token_no"`
	Status         string             `json:"status"`
	Date           time.Time          `json:"date"`
	TotalMedicines int                `json:"total_medicines"`
	Items          []PrescriptionItem `json:"items"`
	LatestSaleID   *uuid.UUID         `json:"latest_sale_id"`
}

type PrescriptionClient interface {
	GetPrescription(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string) (*PrescriptionData, error)
	MarkAsDispensed(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string) error
	UpdateStatus(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, status string) error
	UpdateLatestSaleID(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, saleID uuid.UUID) error
	UpdateBillingInfo(ctx context.Context, pharmacyID uuid.UUID, id string, amount float64, method string, handledBy string, invoiceNo string) error
}

type MockPrescriptionClient struct{}

func (m *MockPrescriptionClient) GetPrescription(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string) (*PrescriptionData, error) {
	if prescriptionID == "" {
		return nil, fmt.Errorf("prescription id required")
	}
	return &PrescriptionData{
		ID:          prescriptionID,
		PatientName: "John Doe",
		Items: []PrescriptionItem{
			{ProductID: uuid.New(), Name: "Paracetamol", Quantity: 10},
		},
	}, nil
}

func (m *MockPrescriptionClient) MarkAsDispensed(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string) error {
	return nil
}

func (m *MockPrescriptionClient) UpdateStatus(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, status string) error {
	return nil
}

func (m *MockPrescriptionClient) UpdateLatestSaleID(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, saleID uuid.UUID) error {
	return nil
}

func (m *MockPrescriptionClient) UpdateBillingInfo(ctx context.Context, pharmacyID uuid.UUID, id string, amount float64, method string, handledBy string, invoiceNo string) error {
	return nil
}

type LocalPrescriptionClient struct {
	repo prescriptions.Repository
}

func NewLocalPrescriptionClient(repo prescriptions.Repository) PrescriptionClient {
	return &LocalPrescriptionClient{repo: repo}
}

func (l *LocalPrescriptionClient) GetPrescription(ctx context.Context, pharmacyID uuid.UUID, id string) (*PrescriptionData, error) {
	p, err := l.repo.GetByID(ctx, pharmacyID, id)
	if err != nil {
		return nil, err
	}

	data := &PrescriptionData{
		ID:             p.ID,
		PatientName:    p.PatientName,
		PatientPhone:   p.PatientPhone,
		DoctorName:     p.DoctorName,
		TokenNo:        p.TokenNo,
		Status:         p.Status,
		Date:           p.Date,
		TotalMedicines: p.TotalMedicines,
		LatestSaleID:   p.LatestSaleID,
	}

	for _, item := range p.Items {
		data.Items = append(data.Items, PrescriptionItem{
			ProductID:    item.ProductID,
			Name:         item.MedicineName,
			Brand:        item.MedicineBrand,
			Quantity:     item.Quantity,
			DurationDays: item.DurationDays,
			DosagePerDay: item.DosagePerDay,
			Morning:      item.Morning,
			Noon:         item.Noon,
			Night:        item.Night,
			Instructions: item.Instructions,
		})
	}
	return data, nil
}

func (l *LocalPrescriptionClient) MarkAsDispensed(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string) error {
	return l.repo.UpdateStatus(ctx, pharmacyID, prescriptionID, "DISPENSED")
}

func (l *LocalPrescriptionClient) UpdateStatus(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, status string) error {
	return l.repo.UpdateStatus(ctx, pharmacyID, prescriptionID, status)
}

func (l *LocalPrescriptionClient) UpdateLatestSaleID(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, saleID uuid.UUID) error {
	return l.repo.UpdateLatestSaleID(ctx, pharmacyID, prescriptionID, saleID)
}

func (l *LocalPrescriptionClient) UpdateBillingInfo(ctx context.Context, pharmacyID uuid.UUID, prescriptionID string, amount float64, method string, handledBy string, invoiceNo string) error {
	return l.repo.UpdateBillingInfo(ctx, pharmacyID, prescriptionID, amount, method, handledBy, invoiceNo)
}
