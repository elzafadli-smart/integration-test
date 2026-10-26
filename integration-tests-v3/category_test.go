package feature_test

import (
	"testing"
)

func TestCategory_CanBeCreated(t *testing.T) {
	this := NewTestCase(t, sharedDB)

	payload := map[string]any{
		"name":        "Coffee Beans",
		"description": "Arabica and Robusta whole beans",
	}

	this.PostJSON("/api/v1/categories", payload).
		AssertCreated().
		AssertJsonPath("data.name", "Coffee Beans")

	this.AssertDatabaseHas("categories", map[string]any{
		"company_id": this.CompanyID,
		"name":       "Coffee Beans",
	})
}

