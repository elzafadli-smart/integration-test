package product_test

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
	db, cleanup := internal.SetupPostgres("v2_product_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

// TestProduct_CanBeCreated demonstrates Laravel:
// $response = $this->postJson('/api/v1/products', $payload);
// $response->assertCreated()->assertJsonPath('data.name', '...');
// $this->assertDatabaseHas('products', ['sku' => '...']);
func TestProduct_CanBeCreated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	payload := map[string]any{
		"name":    "Kopi Gayo Arabica",
		"sku":     "GAYO-001",
		"barcode": "8991234567890",
		"price":   85000.0,
		"stock":   25,
	}

	this.PostJSON("/api/v1/products", payload).
		AssertCreated().
		AssertJsonPath("data.name", "Kopi Gayo Arabica").
		AssertJsonPath("data.sku", "GAYO-001").
		AssertJsonPath("data.barcode", "8991234567890")

	this.AssertDatabaseHas("products", map[string]any{
		"company_id": this.CompanyID,
		"sku":        "GAYO-001",
		"name":       "Kopi Gayo Arabica",
		"barcode":    "8991234567890",
	})
}

// TestProduct_CanBeFetched demonstrates Laravel:
// $response = $this->getJson("/api/v1/products/{$product->id}");
// $response->assertOk()->assertJsonPath(...);
func TestProduct_CanBeFetched(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	product := model.Product{
		ID:        uuid.New(),
		CompanyID: this.CompanyID,
		Name:      "Filter Paper V60",
		SKU:       "V60-PAP",
		Price:     35000.0,
		Stock:     100,
	}
	require.NoError(t, this.DB.Create(&product).Error)

	this.GetJSON(fmt.Sprintf("/api/v1/products/%s", product.ID)).
		AssertOk().
		AssertJsonPath("data.name", "Filter Paper V60").
		AssertJsonPath("data.sku", "V60-PAP")
}

// TestProduct_CanBeUpdated demonstrates Laravel:
// $this->putJson(...)->assertOk();
// $this->assertDatabaseHas('products', ['name' => 'Updated Name']);
func TestProduct_CanBeUpdated(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	product := model.Product{
		ID:        uuid.New(),
		CompanyID: this.CompanyID,
		Name:      "Cold Brew Bottle",
		SKU:       "CBB-250",
		Price:     25000.0,
		Stock:     15,
	}
	require.NoError(t, this.DB.Create(&product).Error)

	updatePayload := map[string]any{
		"name":  "Cold Brew Bottle 500ml",
		"price": 32000.0,
	}

	this.PutJSON(fmt.Sprintf("/api/v1/products/%s", product.ID), updatePayload).
		AssertOk().
		AssertJsonPath("data.name", "Cold Brew Bottle 500ml")

	this.AssertDatabaseHas("products", map[string]any{
		"id":   product.ID,
		"name": "Cold Brew Bottle 500ml",
	})
}

// TestProduct_CanBeDeleted demonstrates Laravel:
// $this->deleteJson(...)->assertOk();
// $this->assertDatabaseMissing('products', ['id' => $id]);
func TestProduct_CanBeDeleted(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	product := model.Product{
		ID:        uuid.New(),
		CompanyID: this.CompanyID,
		Name:      "Product To Delete",
		SKU:       "DEL-001",
		Price:     10000.0,
		Stock:     1,
	}
	require.NoError(t, this.DB.Create(&product).Error)

	this.DeleteJSON(fmt.Sprintf("/api/v1/products/%s", product.ID)).
		AssertOk()

	this.AssertDatabaseMissing("products", map[string]any{
		"id": product.ID,
	})
}

// TestProduct_ValidationFails demonstrates Laravel:
// $this->postJson(..., ['name' => ''])->assertBadRequest();
func TestProduct_ValidationFails(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostJSON("/api/v1/products", map[string]any{
		"name":  "",
		"sku":   "",
		"price": -10,
	}).AssertBadRequest()
}

// TestProduct_RequiresAuthentication demonstrates Laravel:
// $this->withoutHeaders()->postJson(...)->assertUnauthorized();
func TestProduct_RequiresAuthentication(t *testing.T) {
	this := internal.NewTestCase(t, sharedDB)

	this.PostUnauthenticated("/api/v1/products", map[string]any{"name": "Coffee"}).
		AssertUnauthorized()
}
