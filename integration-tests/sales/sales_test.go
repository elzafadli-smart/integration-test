package sales_test

import (
	"encoding/json"
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
	Gorm      *gorm.DB
	App       *fiber.App
	Repo      repository.SalesOrderRepository
	CompanyID uuid.UUID
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

	salesRepo := repository.NewSalesOrderRepository(db.Gorm)
	salesHandler := handler.NewSalesOrderHandler(salesRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	salesHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:      db.Gorm,
		App:       app,
		Repo:      salesRepo,
		CompanyID: internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestSalesOrderCRUD_FullFlow(t *testing.T) {
	customerID := uuid.New()

	// 1. CREATE
	createPayload := model.CreateSalesOrderRequest{
		CustomerID:  customerID,
		OrderNumber: "SO-TEST-001",
		TotalAmount: 250000.0,
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	status, respBody := authedRequest("POST", "/api/v1/sales-orders", body)
	require.Equal(t, 201, status)

	var createResp struct {
		Data model.SalesOrder `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	assert.Equal(t, "SO-TEST-001", createResp.Data.OrderNumber)
	assert.Equal(t, "DRAFT", createResp.Data.Status)
	soID := createResp.Data.ID

	// 2. UPDATE STATUS
	updatePayload := model.UpdateSalesOrderStatusRequest{
		Status: "CONFIRMED",
	}
	body, _ = json.Marshal(updatePayload)
	status, _ = authedRequest("PATCH", "/api/v1/sales-orders/"+soID.String()+"/status", body)
	require.Equal(t, 200, status)

	// 3. GET BY ID
	status, respBody = authedRequest("GET", "/api/v1/sales-orders/"+soID.String(), nil)
	require.Equal(t, 200, status)
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	assert.Equal(t, "CONFIRMED", createResp.Data.Status)

	// 4. DELETE
	status, _ = authedRequest("DELETE", "/api/v1/sales-orders/"+soID.String(), nil)
	require.Equal(t, 200, status)
}
