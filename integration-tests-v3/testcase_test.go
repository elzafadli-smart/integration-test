package feature_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"be-integration-example/internal/handler"
	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var FixedCompanyID = uuid.MustParse("01900000-0000-0000-0000-000000000001")

func getCacheFilePath() string {
	return filepath.Join(os.TempDir(), "v2_shared_test_postgres.url")
}

// tryDirectPostgres checks if an active PostgreSQL instance is already listening locally
// (e.g. from cache file, env TEST_DB_URL, or port 5434/5432). This completely bypasses
// the Windows Docker Named Pipe overhead.
func tryDirectPostgres() (string, bool) {
	candidates := []string{}
	if custom := os.Getenv("TEST_DB_URL"); custom != "" {
		candidates = append(candidates, custom)
	}

	// Check cached URL from previous run
	if cachedData, err := os.ReadFile(getCacheFilePath()); err == nil {
		if trimmed := strings.TrimSpace(string(cachedData)); trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}

	candidates = append(candidates,
		"postgres://groapp:groapp@localhost:5434/postgres?sslmode=disable",
		"postgres://groapp:groapp@localhost:5432/postgres?sslmode=disable",
	)

	for _, connStr := range candidates {
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err = db.PingContext(ctx)
		cancel()
		_ = db.Close()
		if err == nil {
			return connStr, true
		}
	}
	return "", false
}

// SetupPostgres connects to an existing local PostgreSQL container or spins up a shared one,
// and ensures an isolated database exists for each package (dbName).
func SetupPostgres(dbName string) (*gorm.DB, func()) {
	t0 := time.Now()
	var adminConnStr string
	var viaDocker bool

	// 1. Fast-Path: If PostgreSQL is already running locally, connect directly via TCP (< 5ms).
	// This gives identical instant speed to Laravel and completely avoids Windows Named Pipe latency!
	if directConnStr, ok := tryDirectPostgres(); ok {
		adminConnStr = directConnStr
	} else {
		// 2. Fallback: Spin up or reuse via Testcontainers
		viaDocker = true
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if runtime.GOOS == "windows" && os.Getenv("DOCKER_HOST") == "" {
			_ = os.Setenv("DOCKER_HOST", "npipe:////./pipe/docker_engine")
		}

		var container *postgres.PostgresContainer
		var err error
		for attempt := 1; attempt <= 5; attempt++ {
			container, err = postgres.Run(ctx,
				"postgres:16-alpine",
				postgres.WithDatabase("postgres"),
				postgres.WithUsername("groapp"),
				postgres.WithPassword("groapp"),
				testcontainers.WithReuseByName("v2-shared-test-postgres"),
			)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(attempt*300) * time.Millisecond)
		}
		if err != nil {
			panic(fmt.Sprintf("failed to get shared postgres container: %v", err))
		}

		adminConnStr, err = container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			panic(err)
		}

		// Cache connection string for future runs so they bypass Docker Named Pipe
		_ = os.WriteFile(getCacheFilePath(), []byte(adminConnStr), 0644)
	}
	tConnect := time.Since(t0)

	// Connect to default 'postgres' database to manage per-package databases
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		panic(err)
	}
	defer adminDB.Close()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := adminDB.Ping(); err == nil {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	// Ensure this package's isolated database exists (ala Laravel parallel test db: project_test_1)
	if err := ensureDatabaseExists(adminDB, dbName); err != nil {
		panic(fmt.Sprintf("failed to ensure database '%s': %v", dbName, err))
	}

	// Connect to this package's isolated database
	packageConnStr := strings.Replace(adminConnStr, "/postgres?", fmt.Sprintf("/%s?", dbName), 1)
	gormDB, err := gorm.Open(gormpostgres.Open(packageConnStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	tMigrateStart := time.Now()
	// Auto-migrate schema once for this isolated database
	if err := gormDB.AutoMigrate(
		&model.Product{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.Customer{},
		&model.Supplier{},
		&model.Category{},
		&model.SalesOrder{},
		&model.Inventory{},
	); err != nil {
		panic(err)
	}
	tMigrate := time.Since(tMigrateStart)

	source := "Fast-Path TCP (No Named Pipe)"
	if viaDocker {
		source = "Docker Testcontainers"
	}
	fmt.Printf("[V2 Setup %s] Source: %s (%v) | AutoMigrate: %v | Total setup: %v\n",
		dbName, source, tConnect.Round(time.Millisecond), tMigrate.Round(time.Millisecond), time.Since(t0).Round(time.Millisecond))

	cleanup := func() {
		if db, err := gormDB.DB(); err == nil {
			_ = db.Close()
		}
	}

	return gormDB, cleanup
}

func ensureDatabaseExists(adminDB *sql.DB, dbName string) error {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	if err := adminDB.QueryRow(query).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		_, err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

// TestCase implements the Laravel Feature Testing pattern:
// 1. DatabaseTransactions: Every single test method runs within an auto-rolled-back transaction
// 2. Fluent HTTP API: $this->postJson(), $response->assertCreated(), etc.
// 3. Database Assertions: $this->assertDatabaseHas(), $this->assertDatabaseMissing()
type TestCase struct {
	T         *testing.T
	DB        *gorm.DB // The active transaction
	App       *fiber.App
	CompanyID uuid.UUID
}

// NewTestCase begins a transaction for the test method and registers auto-rollback on test end.
func NewTestCase(t *testing.T, sharedDB *gorm.DB) *TestCase {
	// Laravel: DatabaseTransactions trait
	tx := sharedDB.Begin()

	productRepo := repository.NewProductRepository(tx)
	poRepo := repository.NewPurchaseOrderRepository(tx)
	customerRepo := repository.NewCustomerRepository(tx)
	supplierRepo := repository.NewSupplierRepository(tx)
	categoryRepo := repository.NewCategoryRepository(tx)
	salesRepo := repository.NewSalesOrderRepository(tx)
	invRepo := repository.NewInventoryRepository(tx)

	productHandler := handler.NewProductHandler(productRepo)
	poHandler := handler.NewPurchaseOrderHandler(poRepo, productRepo)
	customerHandler := handler.NewCustomerHandler(customerRepo)
	supplierHandler := handler.NewSupplierHandler(supplierRepo)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	salesHandler := handler.NewSalesOrderHandler(salesRepo)
	invHandler := handler.NewInventoryHandler(invRepo)

	app := fiber.New()
	api := app.Group("/api/v1")
	productHandler.RegisterRoutes(api)
	poHandler.RegisterRoutes(api)
	customerHandler.RegisterRoutes(api)
	supplierHandler.RegisterRoutes(api)
	categoryHandler.RegisterRoutes(api)
	salesHandler.RegisterRoutes(api)
	invHandler.RegisterRoutes(api)

	tc := &TestCase{
		T:         t,
		DB:        tx,
		App:       app,
		CompanyID: FixedCompanyID,
	}

	// Auto-rollback when the test method finishes
	t.Cleanup(func() {
		tx.Rollback()
	})

	return tc
}

// TestResponse wraps HTTP response with Laravel-style fluent assertion methods.
type TestResponse struct {
	t        *testing.T
	Status   int
	Body     []byte
	Headers  map[string]string
	jsonData map[string]any
}

func (tc *TestCase) doRequest(method, path string, body []byte, headers map[string]string) *TestResponse {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := tc.App.Test(req)
	require.NoError(tc.T, err, "app.Test should execute without error")
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var parsedJson map[string]any
	_ = json.Unmarshal(respBody, &parsedJson)

	return &TestResponse{
		t:        tc.T,
		Status:   resp.StatusCode,
		Body:     respBody,
		jsonData: parsedJson,
	}
}

// PostJSON simulates Laravel: $this->postJson($url, $data)
func (tc *TestCase) PostJSON(path string, data any) *TestResponse {
	body, _ := json.Marshal(data)
	headers := map[string]string{"X-Company-Id": tc.CompanyID.String()}
	return tc.doRequest("POST", path, body, headers)
}

// GetJSON simulates Laravel: $this->getJson($url)
func (tc *TestCase) GetJSON(path string) *TestResponse {
	headers := map[string]string{"X-Company-Id": tc.CompanyID.String()}
	return tc.doRequest("GET", path, nil, headers)
}

// PutJSON simulates Laravel: $this->putJson($url, $data)
func (tc *TestCase) PutJSON(path string, data any) *TestResponse {
	body, _ := json.Marshal(data)
	headers := map[string]string{"X-Company-Id": tc.CompanyID.String()}
	return tc.doRequest("PUT", path, body, headers)
}

// PatchJSON simulates Laravel: $this->patchJson($url, $data)
func (tc *TestCase) PatchJSON(path string, data any) *TestResponse {
	body, _ := json.Marshal(data)
	headers := map[string]string{"X-Company-Id": tc.CompanyID.String()}
	return tc.doRequest("PATCH", path, body, headers)
}

// DeleteJSON simulates Laravel: $this->deleteJson($url)
func (tc *TestCase) DeleteJSON(path string) *TestResponse {
	headers := map[string]string{"X-Company-Id": tc.CompanyID.String()}
	return tc.doRequest("DELETE", path, nil, headers)
}

// PostUnauthenticated simulates request without auth headers
func (tc *TestCase) PostUnauthenticated(path string, data any) *TestResponse {
	body, _ := json.Marshal(data)
	return tc.doRequest("POST", path, body, nil)
}

// --- Laravel Response Assertions ---

func (r *TestResponse) AssertStatus(expected int) *TestResponse {
	assert.Equal(r.t, expected, r.Status, "HTTP status mismatch. Response body: %s", string(r.Body))
	return r
}

func (r *TestResponse) AssertCreated() *TestResponse {
	return r.AssertStatus(201)
}

func (r *TestResponse) AssertOk() *TestResponse {
	return r.AssertStatus(200)
}

func (r *TestResponse) AssertOK() *TestResponse {
	return r.AssertOk()
}

func (r *TestResponse) GetJsonString(path string) string {
	parts := strings.Split(path, ".")
	var current any = r.jsonData

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return ""
		}
	}
	if current == nil {
		return ""
	}
	return fmt.Sprint(current)
}

func (r *TestResponse) AssertNotFound() *TestResponse {
	return r.AssertStatus(404)
}

func (r *TestResponse) AssertUnauthorized() *TestResponse {
	return r.AssertStatus(401)
}

func (r *TestResponse) AssertBadRequest() *TestResponse {
	return r.AssertStatus(400)
}

func (r *TestResponse) AssertJsonPath(path string, expected any) *TestResponse {
	parts := strings.Split(path, ".")
	var current any = r.jsonData

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			assert.Fail(r.t, fmt.Sprintf("Path '%s' not found in JSON response: %s", path, string(r.Body)))
			return r
		}
	}

	assert.Equal(r.t, expected, current, "JSON value at '%s' did not match", path)
	return r
}

// --- Laravel Database Assertions ---

func (tc *TestCase) AssertDatabaseHas(table string, conditions map[string]any) {
	var count int64
	query := tc.DB.Table(table)
	for k, v := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", k), v)
	}
	err := query.Count(&count).Error
	require.NoError(tc.T, err)
	assert.Greater(tc.T, count, int64(0), "Expected database to have record in table '%s' matching %v", table, conditions)
}

func (tc *TestCase) AssertDatabaseMissing(table string, conditions map[string]any) {
	var count int64
	query := tc.DB.Table(table)
	for k, v := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", k), v)
	}
	err := query.Count(&count).Error
	require.NoError(tc.T, err)
	assert.Equal(tc.T, int64(0), count, "Expected database missing record in table '%s' matching %v", table, conditions)
}

func (tc *TestCase) AssertDatabaseCount(table string, expected int64) {
	var count int64
	err := tc.DB.Table(table).Count(&count).Error
	require.NoError(tc.T, err)
	assert.Equal(tc.T, expected, count, "Expected table '%s' to have %d records, found %d", table, expected, count)
}


