package purchase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"be-integration-example/integration-tests/internal"
	"be-integration-example/internal/handler"
	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testEnv struct {
	Gorm        *gorm.DB
	App         *fiber.App
	PORepo      repository.PurchaseOrderRepository
	ProductRepo repository.ProductRepository
	CompanyID   uuid.UUID
}

var env *testEnv

func TestMain(m *testing.M) {
	db, skipped := internal.SetupPostgres()
	if skipped {
		os.Exit(0)
	}
	if db == nil {
		os.Exit(1)
	}

	productRepo := repository.NewProductRepository(db.Gorm)
	poRepo := repository.NewPurchaseOrderRepository(db.Gorm)
	poHandler := handler.NewPurchaseOrderHandler(poRepo, productRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	poHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:        db.Gorm,
		App:         app,
		PORepo:      poRepo,
		ProductRepo: productRepo,
		CompanyID:   internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestPurchaseOrder_CRUD_FullFlow(t *testing.T) {
	// 1. SEED PRODUCT FIXTURE
	product := model.Product{
		ID:        uuid.New(),
		CompanyID: env.CompanyID,
		Name:      "Biji Kopi Robusta 1kg",
		SKU:       fmt.Sprintf("ROB-%s", uuid.NewString()[:6]),
		Price:     50000.0,
		Stock:     100,
	}
	require.NoError(t, env.ProductRepo.Create(context.Background(), &product))

	// 2. CREATE PURCHASE ORDER
	createReq := model.CreatePurchaseOrderRequest{
		VendorName: "CV Sejahtera Pangan",
		Notes:      "Pengiriman batch pagi",
		Items: []model.CreatePOItemRequest{
			{
				ProductID: product.ID,
				Quantity:  10,
				UnitPrice: 48000.0, // Subtotal: 480,000
			},
			{
				ProductID: uuid.Nil, // Custom item
				Quantity:  2,
				UnitPrice: 25000.0, // Subtotal: 50,000
			},
		},
	}
	body, err := json.Marshal(createReq)
	require.NoError(t, err)

	status, respBody := authedRequest("POST", "/api/v1/purchase-orders", body)
	require.Equal(t, 201, status, "expected 201 Created, got response: %s", string(respBody))

	var createResp struct {
		Code    string              `json:"code"`
		Message string              `json:"message"`
		Data    model.PurchaseOrder `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	po := createResp.Data

	assert.NotEqual(t, uuid.Nil, po.ID)
	assert.Equal(t, env.CompanyID, po.CompanyID)
	assert.Equal(t, "CV Sejahtera Pangan", po.VendorName)
	assert.Equal(t, "DRAFT", po.Status)
	assert.Equal(t, 530000.0, po.TotalAmount)
	require.Len(t, po.Items, 2)
	assert.Regexp(t, `^PO/\d{6}/[A-Z0-9]{6}$`, po.OrderNumber)

	// Verify directly in DB
	var dbPO model.PurchaseOrder
	err = env.Gorm.Preload("Items").
		Where("id = ? AND company_id = ?", po.ID, env.CompanyID).
		First(&dbPO).Error
	require.NoError(t, err)
	assert.Equal(t, 530000.0, dbPO.TotalAmount)
	require.Len(t, dbPO.Items, 2)
	assert.Equal(t, "Biji Kopi Robusta 1kg", dbPO.Items[0].ProductName)
	assert.Equal(t, 480000.0, dbPO.Items[0].Subtotal)

	// 3. GET BY ID
	getURL := fmt.Sprintf("/api/v1/purchase-orders/%s", po.ID.String())
	status, respBody = authedRequest("GET", getURL, nil)
	require.Equal(t, 200, status)

	var getResp struct {
		Code string              `json:"code"`
		Data model.PurchaseOrder `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &getResp))
	assert.Equal(t, po.ID, getResp.Data.ID)
	require.Len(t, getResp.Data.Items, 2)

	// 4. LIST
	status, respBody = authedRequest("GET", "/api/v1/purchase-orders", nil)
	require.Equal(t, 200, status)

	var listResp struct {
		Code string                `json:"code"`
		Data []model.PurchaseOrder `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &listResp))
	found := false
	for _, p := range listResp.Data {
		if p.ID == po.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created purchase order should appear in list")

	// 5. UPDATE STATUS
	statusURL := fmt.Sprintf("/api/v1/purchase-orders/%s/status", po.ID.String())
	statusPayload := []byte(`{"status":"CONFIRMED"}`)
	status, respBody = authedRequest("PATCH", statusURL, statusPayload)
	require.Equal(t, 200, status)

	var updatedPO model.PurchaseOrder
	err = env.Gorm.Where("id = ? AND company_id = ?", po.ID, env.CompanyID).First(&updatedPO).Error
	require.NoError(t, err)
	assert.Equal(t, "CONFIRMED", updatedPO.Status)

	// 6. DELETE
	status, respBody = authedRequest("DELETE", getURL, nil)
	require.Equal(t, 200, status)

	var poCount, itemCount int64
	env.Gorm.Model(&model.PurchaseOrder{}).Where("id = ? AND company_id = ?", po.ID, env.CompanyID).Count(&poCount)
	assert.Equal(t, int64(0), poCount)

	env.Gorm.Model(&model.PurchaseOrderItem{}).Where("purchase_order_id = ?", po.ID).Count(&itemCount)
	assert.Equal(t, int64(0), itemCount)

	// 7. GET AFTER DELETE
	status, _ = authedRequest("GET", getURL, nil)
	assert.Equal(t, 404, status)
}

func TestPurchaseOrder_Validation_400(t *testing.T) {
	// Missing vendorName
	noVendorBody := []byte(`{"vendorName":"","items":[{"quantity":1,"unitPrice":100}]}`)
	status, respBody := authedRequest("POST", "/api/v1/purchase-orders", noVendorBody)
	assert.Equal(t, 400, status)
	assert.Contains(t, string(respBody), "vendorName is required")

	// Empty items
	noItemsBody := []byte(`{"vendorName":"Valid Vendor","items":[]}`)
	status, respBody = authedRequest("POST", "/api/v1/purchase-orders", noItemsBody)
	assert.Equal(t, 400, status)
	assert.Contains(t, string(respBody), "items cannot be empty")

	// Invalid quantity (0)
	invalidQtyBody := []byte(`{"vendorName":"Valid Vendor","items":[{"quantity":0,"unitPrice":100}]}`)
	status, respBody = authedRequest("POST", "/api/v1/purchase-orders", invalidQtyBody)
	assert.Equal(t, 400, status)
	assert.Contains(t, string(respBody), "quantity must be greater than zero")
}

func TestPurchaseOrder_NotFound_404(t *testing.T) {
	nonExistentID := uuid.NewString()
	url := fmt.Sprintf("/api/v1/purchase-orders/%s", nonExistentID)

	status, _ := authedRequest("GET", url, nil)
	assert.Equal(t, 404, status)
}
