package feature_test

import (
	"testing"
)

func TestSupplier_CanBeCreated(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	payload := map[string]any{
		"name":          "CV Kopi Nusantara",
		"contactPerson": "Hendra",
		"phone":         "081344556677",
	}

	this.PostJSON("/api/v1/suppliers", payload).
		AssertCreated().
		AssertJsonPath("data.name", "CV Kopi Nusantara").
		AssertJsonPath("data.contactPerson", "Hendra")

	this.AssertDatabaseHas("suppliers", map[string]any{
		"company_id": this.CompanyID,
		"name":       "CV Kopi Nusantara",
	})
}

