package product_test

import (
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
	Gorm      *gorm.DB
	App       *fiber.App
	Repo      repository.ProductRepository
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

	productRepo := repository.NewProductRepository(db.Gorm)
	productHandler := handler.NewProductHandler(productRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	productHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:      db.Gorm,
		App:       app,
		Repo:      productRepo,
		CompanyID: internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestProductCRUD_FullFlow(t *testing.T) {
	// 1. CREATE
	createPayload := model.CreateProductRequest{
		Name:  "Test Espresso Beans",
		SKU:   fmt.Sprintf("SKU-%s", uuid.NewString()[:8]),
		Price: 75000.0,
		Stock: 50,
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	status, respBody := authedRequest("POST", "/api/v1/products", body)
	require.Equal(t, 201, status, "expected 201 Created, got response: %s", string(respBody))

	var createResp struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Data    model.Product `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	product := createResp.Data

	assert.NotEqual(t, uuid.Nil, product.ID)
	assert.Equal(t, env.CompanyID, product.CompanyID)
	assert.Equal(t, createPayload.Name, product.Name)
	assert.Equal(t, createPayload.SKU, product.SKU)
	assert.Equal(t, createPayload.Price, product.Price)
	assert.Equal(t, createPayload.Stock, product.Stock)

	// Verify directly in DB
	var dbProduct model.Product
	err = env.Gorm.Where("id = ? AND company_id = ?", product.ID, env.CompanyID).First(&dbProduct).Error
	require.NoError(t, err)
	assert.Equal(t, createPayload.Name, dbProduct.Name)

	// 2. READ BY ID
	getURL := fmt.Sprintf("/api/v1/products/%s", product.ID.String())
	status, respBody = authedRequest("GET", getURL, nil)
	require.Equal(t, 200, status)

	var getResp struct {
		Code string        `json:"code"`
		Data model.Product `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &getResp))
	assert.Equal(t, product.ID, getResp.Data.ID)
	assert.Equal(t, createPayload.Name, getResp.Data.Name)

	// 3. LIST
	status, respBody = authedRequest("GET", "/api/v1/products", nil)
	require.Equal(t, 200, status)

	var listResp struct {
		Code string          `json:"code"`
		Data []model.Product `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &listResp))
	found := false
	for _, p := range listResp.Data {
		if p.ID == product.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created product must be in list")

	// 4. UPDATE
	newName := "Updated Arabica Beans"
	newPrice := 85000.0
	newStock := 80
	updatePayload := model.UpdateProductRequest{
		Name:  &newName,
		Price: &newPrice,
		Stock: &newStock,
	}
	updateBody, err := json.Marshal(updatePayload)
	require.NoError(t, err)

	status, respBody = authedRequest("PUT", getURL, updateBody)
	require.Equal(t, 200, status)

	var updateResp struct {
		Code string        `json:"code"`
		Data model.Product `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &updateResp))
	assert.Equal(t, newName, updateResp.Data.Name)

	// 5. DELETE
	status, respBody = authedRequest("DELETE", getURL, nil)
	require.Equal(t, 200, status)

	var count int64
	env.Gorm.Model(&model.Product{}).Where("id = ? AND company_id = ?", product.ID, env.CompanyID).Count(&count)
	assert.Equal(t, int64(0), count)

	// 6. VERIFY NOT FOUND
	status, _ = authedRequest("GET", getURL, nil)
	assert.Equal(t, 404, status)
}

func TestProductCRUD_Unauthorized_401(t *testing.T) {
	status, respBody := internal.DoRequest(env.App, "POST", "/api/v1/products", []byte(`{}`), nil)
	assert.Equal(t, 401, status)
	assert.Contains(t, string(respBody), "unauthorized")
}

func TestProductCRUD_Validation_400(t *testing.T) {
	invalidBody := []byte(`{"name":"","sku":"","price":-10}`)
	status, respBody := authedRequest("POST", "/api/v1/products", invalidBody)
	assert.Equal(t, 400, status)
	assert.Contains(t, string(respBody), "validation_error")
}

func TestProductCRUD_NotFound_404(t *testing.T) {
	nonExistentID := uuid.NewString()
	url := fmt.Sprintf("/api/v1/products/%s", nonExistentID)

	status, _ := authedRequest("GET", url, nil)
	assert.Equal(t, 404, status)
}
