package customer_test

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
	Repo      repository.CustomerRepository
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

	customerRepo := repository.NewCustomerRepository(db.Gorm)
	customerHandler := handler.NewCustomerHandler(customerRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	customerHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:      db.Gorm,
		App:       app,
		Repo:      customerRepo,
		CompanyID: internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestCustomerCRUD_FullFlow(t *testing.T) {
	// 1. CREATE
	createPayload := model.CreateCustomerRequest{
		Name:  "Budi Santoso",
		Email: "budi@example.com",
		Phone: "08123456789",
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	status, respBody := authedRequest("POST", "/api/v1/customers", body)
	require.Equal(t, 201, status)

	var createResp struct {
		Data model.Customer `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	assert.Equal(t, "Budi Santoso", createResp.Data.Name)
	assert.Equal(t, "budi@example.com", createResp.Data.Email)
	customerID := createResp.Data.ID

	// 2. GET BY ID
	status, respBody = authedRequest("GET", "/api/v1/customers/"+customerID.String(), nil)
	require.Equal(t, 200, status)

	// 3. UPDATE
	newName := "Budi Santoso Updated"
	updatePayload := model.UpdateCustomerRequest{
		Name: &newName,
	}
	body, _ = json.Marshal(updatePayload)
	status, _ = authedRequest("PUT", "/api/v1/customers/"+customerID.String(), body)
	require.Equal(t, 200, status)

	// 4. DELETE
	status, _ = authedRequest("DELETE", "/api/v1/customers/"+customerID.String(), nil)
	require.Equal(t, 200, status)

	// Verify not found
	status, _ = authedRequest("GET", "/api/v1/customers/"+customerID.String(), nil)
	require.Equal(t, 404, status)
}
