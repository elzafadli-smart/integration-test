package feature_test

import (
	"testing"
	"github.com/google/uuid"
)

func TestSalesOrder_CanBeCreated(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	customerID := uuid.New()
	payload := map[string]any{
		"customerId":  customerID,
		"orderNumber": "SO-V3-001",
		"totalAmount": 150000.0,
	}

	this.PostJSON("/api/v1/sales-orders", payload).
		AssertCreated().
		AssertJsonPath("data.orderNumber", "SO-V3-001").
		AssertJsonPath("data.status", "DRAFT")

	this.AssertDatabaseHas("sales_orders", map[string]any{
		"company_id":   this.CompanyID,
		"order_number": "SO-V3-001",
		"status":       "DRAFT",
	})
}

