.PHONY: test test-v2 test-all test-v2-up test-v2-clean \
	test-product test-purchase test-customer test-supplier test-category test-sales test-inventory \
	test-v2-product test-v2-purchase test-v2-customer test-v2-supplier test-v2-category test-v2-sales test-v2-inventory

# V1: Testcontainers per-folder pattern (7 Modules, Fresh Container per folder)
test:
	go test -v -count=1 ./integration-tests/...

test-product:
	go test -v -count=1 ./integration-tests/product/...
test-purchase:
	go test -v -count=1 ./integration-tests/purchase/...
test-customer:
	go test -v -count=1 ./integration-tests/customer/...
test-supplier:
	go test -v -count=1 ./integration-tests/supplier/...
test-category:
	go test -v -count=1 ./integration-tests/category/...
test-sales:
	go test -v -count=1 ./integration-tests/sales/...
test-inventory:
	go test -v -count=1 ./integration-tests/inventory/...

# V2: Laravel-style Integration Testing (7 Modules, Reused Container + Per-Folder DB + Auto-Rollback)
test-v2:
	go test -v -count=1 ./integration-tests-v2/...

test-v2-product:
	go test -v -count=1 ./integration-tests-v2/product/...
test-v2-purchase:
	go test -v -count=1 ./integration-tests-v2/purchase/...
test-v2-customer:
	go test -v -count=1 ./integration-tests-v2/customer/...
test-v2-supplier:
	go test -v -count=1 ./integration-tests-v2/supplier/...
test-v2-category:
	go test -v -count=1 ./integration-tests-v2/category/...
test-v2-sales:
	go test -v -count=1 ./integration-tests-v2/sales/...
test-v2-inventory:
	go test -v -count=1 ./integration-tests-v2/inventory/...

# V3: Single-Binary Feature Suite (1 Biner untuk SELURUH 7 Modul - 100% Identik Laravel)
test-v3:
	go test -v -count=1 ./integration-tests-v3/...

test-v3-product:
	go test -v ./integration-tests-v3 -run TestProduct
test-v3-purchase:
	go test -v ./integration-tests-v3 -run TestPurchase
test-v3-customer:
	go test -v ./integration-tests-v3 -run TestCustomer
test-v3-supplier:
	go test -v ./integration-tests-v3 -run TestSupplier
test-v3-category:
	go test -v ./integration-tests-v3 -run TestCategory
test-v3-sales:
	go test -v ./integration-tests-v3 -run TestSales
test-v3-inventory:
	go test -v ./integration-tests-v3 -run TestInventory

# Start the shared V2/V3 test container on port 5434 (enables instant Fast-Path TCP)
test-v2-up:
	-docker run -d --name v2-shared-test-postgres -p 5434:5432 -e POSTGRES_USER=groapp -e POSTGRES_PASSWORD=groapp -e POSTGRES_DB=postgres postgres:16-alpine

# Clean up the reused V2 test container and cached URL if needed
test-v2-clean:
	-docker rm -f v2-shared-test-postgres
	-powershell -Command "Remove-Item -Force -ErrorAction SilentlyContinue \"$([IO.Path]::GetTempPath())v2_shared_test_postgres.url\""

test-all: test test-v2
