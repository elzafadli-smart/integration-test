package feature_test

import (
	"testing"
	"github.com/google/uuid"
)

func TestInventory_CanAdjustStock(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	productID := uuid.New()
	payload := map[string]any{
		"productId": productID,
		"warehouse": "GUDANG-PUSAT",
		"quantity":  120,
	}

	this.PostJSON("/api/v1/inventory/adjust", payload).
		AssertOK().
		AssertJsonPath("message", "inventory adjusted successfully")

	this.AssertDatabaseHas("inventories", map[string]any{
		"company_id": this.CompanyID,
		"product_id": productID,
		"warehouse":  "GUDANG-PUSAT",
		"quantity":   120,
	})
}

