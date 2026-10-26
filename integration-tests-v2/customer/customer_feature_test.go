package customer_test

import (
	"os"
	"testing"

	"be-integration-example/integration-tests-v2/internal"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	db, cleanup := internal.SetupPostgres("v2_customer_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

func TestCustomer_CanBeCreated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

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

func TestCustomer_ValidationFails(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostJSON("/api/v1/customers", map[string]any{
		"name": "",
	}).AssertBadRequest()
}

func TestCustomer_RequiresAuthentication(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostUnauthenticated("/api/v1/customers", map[string]any{"name": "No Auth"}).
		AssertUnauthorized()
}
