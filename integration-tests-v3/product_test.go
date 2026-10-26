package feature_test

import (
	"testing"
)

func TestProduct_CanBeCreated(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	payload := map[string]any{
		"name":  "Kopi Gayo Arabica",
		"sku":   "GAYO-V3-001",
		"price": 85000.0,
		"stock": 25,
	}

	this.PostJSON("/api/v1/products", payload).
		AssertCreated().
		AssertJsonPath("data.name", "Kopi Gayo Arabica").
		AssertJsonPath("data.sku", "GAYO-V3-001")

	this.AssertDatabaseHas("products", map[string]any{
		"company_id": this.CompanyID,
		"sku":        "GAYO-V3-001",
		"name":       "Kopi Gayo Arabica",
	})
}

func TestProduct_ValidationFails(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	this.PostJSON("/api/v1/products", map[string]any{
		"name": "",
		"sku":  "",
	}).AssertBadRequest()
}

func TestProduct_RequiresAuthentication(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	this.PostUnauthenticated("/api/v1/products", map[string]any{
		"name": "Kopi",
		"sku":  "KOP-1",
	}).AssertUnauthorized()
}

