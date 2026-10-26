package sales_test

import (
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_sales_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

func TestSalesOrder_CanBeCreated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	customerID := uuid.New()
	payload := map[string]any{
		"customerId":  customerID,
		"orderNumber": "SO-2026-001",
		"totalAmount": 150000.0,
	}

	this.PostJSON("/api/v1/sales-orders", payload).
		AssertCreated().
		AssertJsonPath("data.orderNumber", "SO-2026-001").
		AssertJsonPath("data.status", "DRAFT")

	this.AssertDatabaseHas("sales_orders", map[string]any{
		"company_id":   this.CompanyID,
		"order_number": "SO-2026-001",
		"status":       "DRAFT",
	})
}

func TestSalesOrder_CanUpdateStatus(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	customerID := uuid.New()
	res := this.PostJSON("/api/v1/sales-orders", map[string]any{
		"customerId":  customerID,
		"orderNumber": "SO-2026-002",
		"totalAmount": 300000.0,
	}).AssertCreated()

	orderID := res.GetJsonString("data.id")

	this.PatchJSON("/api/v1/sales-orders/"+orderID+"/status", map[string]any{
		"status": "COMPLETED",
	}).AssertOK()

	this.AssertDatabaseHas("sales_orders", map[string]any{
		"id":     orderID,
		"status": "COMPLETED",
	})
}
