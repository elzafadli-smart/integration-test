package category_test

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
	Repo      repository.CategoryRepository
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

	categoryRepo := repository.NewCategoryRepository(db.Gorm)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)

	app := fiber.New()
	apiGroup := app.Group("/api/v1")
	categoryHandler.RegisterRoutes(apiGroup)

	env = &testEnv{
		Gorm:      db.Gorm,
		App:       app,
		Repo:      categoryRepo,
		CompanyID: internal.FixedCompanyID,
	}

	code := m.Run()
	db.Cleanup()
	os.Exit(code)
}

func authedRequest(method, path string, body []byte) (int, []byte) {
	return internal.AuthedRequest(env.App, env.CompanyID, method, path, body)
}

func TestCategoryCRUD_FullFlow(t *testing.T) {
	// 1. CREATE
	createPayload := model.CreateCategoryRequest{
		Name:        "Beverages",
		Description: "All kinds of coffee and tea",
	}
	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	status, respBody := authedRequest("POST", "/api/v1/categories", body)
	require.Equal(t, 201, status)

	var createResp struct {
		Data model.Category `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &createResp))
	assert.Equal(t, "Beverages", createResp.Data.Name)
	categoryID := createResp.Data.ID

	// 2. GET BY ID
	status, respBody = authedRequest("GET", "/api/v1/categories/"+categoryID.String(), nil)
	require.Equal(t, 200, status)

	// 3. DELETE
	status, _ = authedRequest("DELETE", "/api/v1/categories/"+categoryID.String(), nil)
	require.Equal(t, 200, status)
}
