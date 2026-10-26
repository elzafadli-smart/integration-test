package purchase_test

import (
	"fmt"
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_purchase_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

// TestPurchaseOrder_CanBeCreatedWithLineItems tests creating a full PO with nested line items
// and validates both response and database persistence ala Laravel.
func TestPurchaseOrder_CanBeCreatedWithLineItems(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	// Fixture created inside this test's transaction
	product := model.Product{
		ID:        uuid.New(),
		CompanyID: this.CompanyID,
		Name:      "Biji Kopi Toraja 1kg",
		SKU:       "TOR-001",
		Price:     60000.0,
		Stock:     50,
	}
	require.NoError(t, this.DB.Create(&product).Error)

	payload := map[string]any{
		"vendorName": "PT Pangan Nusantara",
		"notes":      "Pengiriman mendesak",
		"items": []map[string]any{
			{
				"productId": product.ID,
				"quantity":  5,
				"unitPrice": 58000.0,
			},
		},
	}

	this.PostJSON("/api/v1/purchase-orders", payload).
		AssertCreated().
		AssertJsonPath("data.vendorName", "PT Pangan Nusantara").
		AssertJsonPath("data.status", "DRAFT").
		AssertJsonPath("data.totalAmount", 290000.0) // 5 * 58000

	this.AssertDatabaseHas("purchase_orders", map[string]any{
		"vendor_name":  "PT Pangan Nusantara",
		"total_amount": 290000.0,
		"status":       "DRAFT",
	})

	this.AssertDatabaseHas("purchase_order_items", map[string]any{
		"product_name": "Biji Kopi Toraja 1kg",
		"quantity":     5,
		"subtotal":     290000.0,
	})
}

// TestPurchaseOrder_CanUpdateStatus tests changing status to CONFIRMED
func TestPurchaseOrder_CanUpdateStatus(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	poID := uuid.New()
	po := model.PurchaseOrder{
		ID:          poID,
		CompanyID:   this.CompanyID,
		OrderNumber: "PO/202610/TEST01",
		VendorName:  "Supplier Beras",
		Status:      "DRAFT",
		TotalAmount: 100000.0,
	}
	require.NoError(t, this.DB.Create(&po).Error)

	this.PatchJSON(fmt.Sprintf("/api/v1/purchase-orders/%s/status", poID), map[string]any{
		"status": "CONFIRMED",
	}).AssertOk()

	this.AssertDatabaseHas("purchase_orders", map[string]any{
		"id":     poID,
		"status": "CONFIRMED",
	})
}

// TestPurchaseOrder_CascadeDeletesItems tests deleting a PO and asserting both header and items are removed
func TestPurchaseOrder_CascadeDeletesItems(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	poID := uuid.New()
	po := model.PurchaseOrder{
		ID:          poID,
		CompanyID:   this.CompanyID,
		OrderNumber: "PO/202610/DEL01",
		VendorName:  "Vendor to Delete",
		Status:      "DRAFT",
		TotalAmount: 50000.0,
		Items: []model.PurchaseOrderItem{
			{
				ID:              uuid.New(),
				CompanyID:       this.CompanyID,
				PurchaseOrderID: poID,
				ProductID:       uuid.New(),
				ProductName:     "Sample Item",
				Quantity:        1,
				UnitPrice:       50000.0,
				Subtotal:        50000.0,
			},
		},
	}
	require.NoError(t, this.DB.Create(&po).Error)

	// Assert records exist before delete
	this.AssertDatabaseHas("purchase_orders", map[string]any{"id": poID})
	this.AssertDatabaseHas("purchase_order_items", map[string]any{"purchase_order_id": poID})

	// Delete
	this.DeleteJSON(fmt.Sprintf("/api/v1/purchase-orders/%s", poID)).
		AssertOk()

	// Assert records are missing after delete
	this.AssertDatabaseMissing("purchase_orders", map[string]any{"id": poID})
	this.AssertDatabaseMissing("purchase_order_items", map[string]any{"purchase_order_id": poID})
}

// TestPurchaseOrder_ValidationFails tests input validation
func TestPurchaseOrder_ValidationFails(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	// Empty vendor name
	this.PostJSON("/api/v1/purchase-orders", map[string]any{
		"vendorName": "",
		"items":      []map[string]any{{"quantity": 1, "unitPrice": 100}},
	}).AssertBadRequest()

	// Empty items
	this.PostJSON("/api/v1/purchase-orders", map[string]any{
		"vendorName": "Valid Vendor",
		"items":      []any{},
	}).AssertBadRequest()
}
