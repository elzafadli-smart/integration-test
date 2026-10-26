package feature_test

import (
	"os"
	"testing"
	"gorm.io/gorm"
)

// sharedDB is the singleton connection pool for the entire feature test suite.
// This mimics Laravel's single-process test environment.
var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	// Initialize single database for the entire feature test suite
	db, cleanup := SetupPostgres("v3_feature_test")
	sharedDB = db
	defer cleanup()

	os.Exit(m.Run())
}

