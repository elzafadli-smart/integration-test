package supplier_test

import (
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_supplier_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

func TestSupplier_CanBeCreated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

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

func TestSupplier_ValidationFails(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostJSON("/api/v1/suppliers", map[string]any{
		"name": "",
	}).AssertBadRequest()
}
