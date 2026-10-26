package category_test

import (
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_category_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

func TestCategory_CanBeCreated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

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

func TestCategory_ValidationFails(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostJSON("/api/v1/categories", map[string]any{
		"name": "",
	}).AssertBadRequest()
}
