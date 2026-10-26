package inventory_test

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
	Repo      repository.InventoryRepository
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

	inventoryRepo := repository.NewInventoryRepository(db.Gorm)
	inventoryHandler := handler.NewInventoryHandler(inventoryRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	inventoryHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:      db.Gorm,
		App:       app,
		Repo:      inventoryRepo,
		CompanyID: internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestInventory_AdjustAndList(t *testing.T) {
	productID := uuid.New()

	// 1. ADJUST (Add Stock)
	adjustPayload := model.AdjustInventoryRequest{
		ProductID: productID,
		Warehouse: "WH-CENTRAL",
		Quantity:  100,
	}
	body, err := json.Marshal(adjustPayload)
	require.NoError(t, err)

	status, _ := authedRequest("POST", "/api/v1/inventory/adjust", body)
	require.Equal(t, 200, status)

	// 2. ADJUST AGAIN (Add 50)
	adjustPayload2 := model.AdjustInventoryRequest{
		ProductID: productID,
		Warehouse: "WH-CENTRAL",
		Quantity:  50,
	}
	body2, _ := json.Marshal(adjustPayload2)
	status, _ = authedRequest("POST", "/api/v1/inventory/adjust", body2)
	require.Equal(t, 200, status)

	// 3. GET BY PRODUCT
	status, respBody := authedRequest("GET", "/api/v1/inventory/product/"+productID.String(), nil)
	require.Equal(t, 200, status)

	var listResp struct {
		Data []model.Inventory `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &listResp))
	require.NotEmpty(t, listResp.Data)
	assert.Equal(t, 150, listResp.Data[0].Quantity)
}
