package inventory_test

import (
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_inventory_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

func TestInventory_CanAdjustStock(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	productID := uuid.New()
	payload := map[string]any{
		"productId": productID,
		"warehouse": "GUDANG-BARAT",
		"quantity":  75,
	}

	this.PostJSON("/api/v1/inventory/adjust", payload).
		AssertOK().
		AssertJsonPath("message", "inventory adjusted successfully")

	this.AssertDatabaseHas("inventories", map[string]any{
		"company_id": this.CompanyID,
		"product_id": productID,
		"warehouse":  "GUDANG-BARAT",
		"quantity":   75,
	})
}
