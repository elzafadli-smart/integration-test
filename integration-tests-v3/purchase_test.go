package feature_test

import (
	"testing"

	"be-integration-example/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPurchaseOrder_CanBeCreatedWithLineItems(t *testing.T) {
	this := NewTestCase(t, sharedDB)

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

	res := this.PostJSON("/api/v1/purchase-orders", payload).
		AssertCreated().
		AssertJsonPath("data.vendorName", "PT Pangan Nusantara").
		AssertJsonPath("data.status", "DRAFT").
		AssertJsonPath("data.totalAmount", 290000.0)

	poID := res.GetJsonString("data.id")

	this.AssertDatabaseHas("purchase_orders", map[string]any{
		"vendor_name":  "PT Pangan Nusantara",
		"total_amount": 290000.0,
		"status":       "DRAFT",
	})

	this.AssertDatabaseHas("purchase_order_items", map[string]any{
		"purchase_order_id": poID,
		"product_id":        product.ID,
		"quantity":          5,
	})
}
