package internal

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"time"

	"be-integration-example/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var FixedCompanyID = uuid.MustParse("01900000-0000-0000-0000-000000000001")

type IntegrationDB struct {
	DB      *sql.DB
	Gorm    *gorm.DB
	Cleanup func()
}

// SetupPostgres spins up a clean PostgreSQL testcontainer, waits for ready,
// and runs schema auto-migration across all models.
func SetupPostgres() (*IntegrationDB, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("groapp_test"),
		postgres.WithUsername("groapp"),
		postgres.WithPassword("groapp"),
	)
	if err != nil {
		fmt.Printf("[setup] Skipping integration test, Docker run failed: %v\n", err)
		return nil, true
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("[setup] ConnectionString failed: %v\n", err)
		_ = container.Terminate(context.Background())
		return nil, false
	}

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("[setup] sql.Open failed: %v\n", err)
		_ = container.Terminate(context.Background())
		return nil, false
	}

	if err := waitForDB(ctx, sqlDB); err != nil {
		fmt.Printf("[setup] waitForDB failed: %v\n", err)
		_ = container.Terminate(context.Background())
		return nil, false
	}

	gormDB, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		fmt.Printf("[setup] gorm.Open failed: %v\n", err)
		_ = container.Terminate(context.Background())
		return nil, false
	}

	// Migrate all shared schemas
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
		fmt.Printf("[setup] AutoMigrate failed: %v\n", err)
		_ = container.Terminate(context.Background())
		return nil, false
	}

	cleanup := func() {
		_ = sqlDB.Close()
		_ = container.Terminate(context.Background())
	}

	return &IntegrationDB{
		DB:      sqlDB,
		Gorm:    gormDB,
		Cleanup: cleanup,
	}, false
}

func waitForDB(ctx context.Context, db *sql.DB) error {
	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}
	for time.Now().Before(deadline) {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("db ping timeout")
}

// DoRequest executes an in-memory HTTP request against a Fiber app
func DoRequest(app *fiber.App, method, path string, body []byte, headers map[string]string) (int, []byte) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		panic(fmt.Sprintf("app.Test failed: %v", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

// AuthedRequest executes an in-memory HTTP request including X-Company-Id header
func AuthedRequest(app *fiber.App, companyID uuid.UUID, method, path string, body []byte) (int, []byte) {
	headers := map[string]string{
		"X-Company-Id": companyID.String(),
	}
	return DoRequest(app, method, path, body, headers)
}
