package feature_test

import (
	"testing"
)

func TestCustomer_CanBeCreated(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	payload := map[string]any{
		"name":  "PT Maju Bersama",
		"email": "info@majubersama.com",
		"phone": "081299887766",
	}

	this.PostJSON("/api/v1/customers", payload).
		AssertCreated().
		AssertJsonPath("data.name", "PT Maju Bersama").
		AssertJsonPath("data.email", "info@majubersama.com")

	this.AssertDatabaseHas("customers", map[string]any{
		"company_id": this.CompanyID,
		"name":       "PT Maju Bersama",
		"email":      "info@majubersama.com",
	})
}

