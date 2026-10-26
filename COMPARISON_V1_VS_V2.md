# Perbandingan Integration Testing: V1 vs V2 vs V3 (Single-Binary Suite)

Dokumen ini membandingkan 3 pendekatan arsitektur integration testing yang diimplementasikan di repository ini:
- **Metode V1 (`integration-tests/`)**: Pola `be-groapp-accounting` (Container fisik baru per folder modul).
- **Metode V2 (`integration-tests-v2/`)**: Pola **Laravel Parallel Testing** (Reusable Container + 1 Database terisolasi per folder modul).
- **Metode V3 (`integration-tests-v3/`)**: Pola **Laravel Feature Suite (Single Binary)** (Seluruh modul dalam 1 package test suite, kompilasi hanya 1 biner, support partial run via filter `-run`).

---

## 1. Ringkasan Perbandingan (Matrix)

| Parameter | **Metode V1 (Tradisional)** | **Metode V2 (Multi-Folder DB)** | **Metode V3 (Single-Binary Suite)** |
| :--- | :--- | :--- | :--- |
| **Filosofi Desain** | Standar Go / Testcontainers | Laravel Parallel Testing | Laravel Feature Test Suite |
| **Jumlah Biner Go** | **7 Biner Terpisah** | **7 Biner Terpisah** | **HANYA 1 BINER TUNGGAL** |
| **Kompilasi CPU (Windows)** | ~12.5 detik (compile 7x) | ~12.5 detik (compile 7x) | **~3 detik (compile 1x)** |
| **Koneksi Database** | Docker Named Pipe (`npipe`) | Fast-Path TCP (< 30 ms) | Fast-Path TCP (< 30 ms) |
| **Container Docker** | 7 Container Fisik | 1 Container Bersama | 1 Container Bersama |
| **Isolasi Database** | Container berbeda | Database berbeda (`v2_xxx_test`) | Transaksi auto-rollback (`tx.Rollback()`) |
| **Waktu Full Suite 7 Modul** | **~16.6 detik** | **~13.0 detik** | **~5.4 detik** (Paling Cepat! ⚡) |
| **Partial Run (1 Modul)** | `make test-product` (~5.5s) | `make test-v2-product` (~4s) | `make test-v3-product` (**~0.05s**) |
| **Filter Test Parsial** | Per folder | Per folder | Flag `-run TestModul` (seperti `--filter` di Laravel) |

---

## 2. Arsitektur & Cara Kerja

### Metode V1: Container-per-Folder (Pola `be-groapp-accounting`)

```
go test ./integration-tests/...
      │
      ├──> [Folder: product/]  ──> Docker Named Pipe Handshake (3.5s)
      │                            ──> Spawn Container Postgres A (4.5s)
      │                            ──> Run Test (0.01s)
      │                            ──> Destroy Container A (2.3s)
      │
      └──> [Folder: purchase/] ──> Docker Named Pipe Handshake (3.5s)
                                   ──> Spawn Container Postgres B (4.5s)
                                   ──> Run Test (0.01s)
                                   ──> Destroy Container B (2.3s)
```

**Karakteristik V1:**
1. Tiap package Go menyalakan container PostgreSQL baru dari image `postgres:16-alpine`.
2. Menghabiskan waktu startup & teardown Docker (~8–10 detik) di **setiap kali** test package dijalankan.
3. Di dalam 1 folder, data yang dibuat oleh satu fungsi test akan tetap tersimpan di tabel database sampai dihapus manual atau container dimatikan.

---

### Metode V2: Laravel Parallel Testing + Fast-Path TCP Pattern

```
go test ./integration-tests-v2/...
      │
      ▼
1. Cek TCP port lokal (port 5434 / 5432)  ──>  Terhubung dalam 26 ms! ⚡
   (Bypass Docker Named Pipe & Testcontainers 100%)
      │
      ▼
┌────────────────────────────────────────────────────────────────────────┐
│               1 SHARED DOCKER CONTAINER POSTGRES                       │
│                     "v2-shared-test-postgres"                          │
│               (Standby di port 5434, dipakai semua folder)             │
└───────────────────┬────────────────────────────────┬───────────────────┘
                    │                                │
                    ▼                                ▼
       Database: v2_product_test        Database: v2_purchase_test
       (Dibuat otomatis via SQL)         (Dibuat otomatis via SQL)
                    │                                │
                    ▼                                ▼
        [Test A: tx.Rollback()]          [Test X: tx.Rollback()]
        [Test B: tx.Rollback()]          [Test Y: tx.Rollback()]
```

**Karakteristik V2:**
1. **Fast-Path TCP Bypass**: Menguji koneksi TCP langsung ke port `5434` / `5432`. Jika container test sudah standby, Go langsung connect dalam **26–48 ms** tanpa menyentuh Windows Named Pipe Docker.
2. **Dynamic Database Provisioning**: Secara otomatis mengeksekusi `CREATE DATABASE v2_product_test` jika belum ada. Folder `product` dan `purchase` memiliki database fisik masing-masing di dalam 1 container yang sama.
3. **`DatabaseTransactions` Trait**: Setiap test method diawali dengan `tx := db.Begin()` dan diakhiri dengan `tx.Rollback()`. Database selalu kembali bersih 100% seketika tanpa perlu truncate tabel.
4. **Fallback Testcontainers**: Jika belum ada container Postgres yang menyala di lokal, sistem otomatis menyalakan container `v2-shared-test-postgres` via Testcontainers.

---

### Metode V3: Laravel Feature Suite (Single Binary Pattern)

```
go test ./integration-tests-v3/...
      │
      ▼
1. Kompilasi Go: Hanya 1 Biner Tunggal ("feature.test.exe") ──> Sangat Cepat (~3s)!
      │
      ▼
2. TestMain(): Konek Fast-Path TCP ke 1 Container Reusable (Port 5434) (24 ms)
   dan pastikan Database "v3_feature_test" siap (1x migrate untuk seluruh test)
      │
      ▼
┌────────────────────────────────────────────────────────────────────────┐
│                        1 PROSES BINER PENGUJIAN                        │
│                   (Menjalankan Seluruh Modul Berurutan)                │
├───────────────────┬───────────────────┬───────────────────┬────────────┤
│  product_test.go  │ purchase_test.go  │ customer_test.go  │  sales...  │
│  [tx.Rollback()]  │  [tx.Rollback()]  │  [tx.Rollback()]  │  [Roll...] │
└───────────────────┴───────────────────┴───────────────────┴────────────┘
```

**Karakteristik V3:**
1. **Single Binary Compilation**: Seluruh modul berada dalam satu folder/package test (`package feature_test`), persis seperti folder `tests/Feature/` di Laravel. Go **hanya mengompilasi 1 file biner**, menghemat waktu CPU dari 12.5s menjadi hanya **~3s**.
2. **One-Time Setup**: Inisialisasi TCP database dan `AutoMigrate` hanya dijalankan **1 kali** di awal `TestMain()`. Seluruh modul langsung bergantian menggunakan database `v3_feature_test` yang bersih berkat `tx.Rollback()`.
3. **Fleksibilitas Parsial (`-run`)**: Meskipun berada dalam 1 biner, pengembang tetap bisa menjalankan modul terpisah menggunakan flag filter bawaan Go (contoh: `go test ./integration-tests-v3 -run TestCustomer`), sama seperti opsi `--filter=CustomerTest` di Laravel.

---

## 3. Perbandingan Kode (Side-by-Side)

### Skenario: Membuat Produk (POST `/api/v1/products`) & Verifikasi Database

#### Metode V1 (Tradisional Go)
```go
func TestProductCRUD_Create(t *testing.T) {
    // 1. Marshalling JSON manual
    createPayload := model.CreateProductRequest{
        Name:  "Test Espresso Beans",
        SKU:   "SKU-001",
        Price: 75000.0,
        Stock: 50,
    }
    body, err := json.Marshal(createPayload)
    require.NoError(t, err)

    // 2. Request HTTP manual
    status, respBody := authedRequest("POST", "/api/v1/products", body)
    require.Equal(t, 201, status)

    // 3. Unmarshalling response manual
    var createResp struct {
        Data model.Product `json:"data"`
    }
    require.NoError(t, json.Unmarshal(respBody, &createResp))
    assert.Equal(t, "Test Espresso Beans", createResp.Data.Name)

    // 4. Verifikasi langsung ke DB dengan manual GORM query
    var dbProduct model.Product
    err = env.Gorm.Where("id = ? AND company_id = ?", createResp.Data.ID, env.CompanyID).First(&dbProduct).Error
    require.NoError(t, err)
    assert.Equal(t, "Test Espresso Beans", dbProduct.Name)
}
```

#### Metode V2 & V3 (Laravel Style Chaining)
```go
func TestProduct_CanBeCreated(t *testing.T) {
    // 1. Inisialisasi test case dengan auto-rollback transaction
    this := internal.NewTestCase(t, sharedDB) // Di V3: NewTestCase(t, sharedDB)

    payload := map[string]any{
        "name":  "Kopi Gayo Arabica",
        "sku":   "GAYO-001",
        "price": 85000.0,
        "stock": 25,
    }

    // 2. Request dan Response Assertions secara Fluent Chaining
    this.PostJSON("/api/v1/products", payload).
        AssertCreated().
        AssertJsonPath("data.name", "Kopi Gayo Arabica").
        AssertJsonPath("data.sku", "GAYO-001")

    // 3. Database Assertion ala Laravel: $this->assertDatabaseHas()
    this.AssertDatabaseHas("products", map[string]any{
        "company_id": this.CompanyID,
        "sku":        "GAYO-001",
        "name":       "Kopi Gayo Arabica",
    })
    // Transaksi di-rollback otomatis begitu fungsi selesai!
}
```

---

## 4. Kelebihan & Kekurangan Masing-Masing

### Metode V1 (Multi-Container Tradisional)
- **Kelebihan**:
  - Isolasi tingkat infrastruktur paling kuat (karena container Docker terpisah).
- **Kekurangan**:
  - Waktu tunggu sangat lama (16+ detik untuk 7 modul).
  - Boros RAM (menyalakan 7 instance container PostgreSQL sekaligus).
  - Rentan tabrakan data antar-method dalam 1 folder jika tidak dibersihkan manual.

### Metode V2 (Multi-Folder Parallel DB)
- **Kelebihan**:
  - Tiap modul memiliki database fisik terpisah (`v2_product_test`, `v2_customer_test`).
  - Paket Go terpisah per folder (modularitas paket Go standar).
  - Transaksi otomatis rollback per method test.
- **Kekurangan**:
  - Kompilasi Go di Windows memakan waktu ~12.5 detik saat `full suite` karena mengompilasi 7 file `.exe` secara bersamaan.

### Metode V3 (Single-Binary Suite ala Laravel Feature Test)
- **Kelebihan**:
  - **Paling Cepat untuk Full Suite**: Hanya butuh **5.4 detik** untuk seluruh 7 modul (menghemat waktu kompilasi 70%).
  - **Sangat Mirip Laravel**: Seluruh test file berada di satu folder `tests/Feature`, dieksekusi dalam 1 biner.
  - **Tetap Bisa Partial Run**: Cukup gunakan flag `-run TestCustomer` atau target Makefile `make test-v3-customer`.
  - Inisialisasi database dan AutoMigrate hanya terjadi 1 kali saja.
- **Kekurangan**:
  - Seluruh fungsi test berada dalam satu package Go (`feature_test`), sehingga nama fungsi test harus unik (misal `TestProduct_...`, `TestCustomer_...`).

---

## 5. Panduan Pemilihan (Kapan Menggunakan V1, V2, atau V3?)

| Kebutuhan Pengujian | Rekomendasi | Alasan |
| :--- | :---: | :--- |
| **Ingin kecepatan eksekusi full suite paling maksimal & arsitektur persis Laravel** | **Metode V3** | Kompilasi hanya 1 biner, total test selesai dalam **~5.4 detik**. |
| **Ingin pemisahan package folder Go yang ketat per domain modul dengan database terisolasi** | **Metode V2** | Struktur folder modular per modul, database per folder aman dari tabrakan. |
| **Pengujian yang memodifikasi DDL skema DB secara runtime (e.g. migration runner engine)** | **Metode V1** | Membutuhkan container bersih baru untuk pengujian engine migrasi. |

---

## 6. Hasil Pengukuran & Estimasi Waktu (Benchmark 7 Modul Riil)

Pengujian telah divalidasi pada **7 modul lengkap** (`product`, `purchase`, `customer`, `supplier`, `category`, `sales`, `inventory`) di mesin lokal Windows dengan Docker Desktop:

### A. Waktu Eksekusi Riil 7 Modul Penuh (`go test ./...`)

| Modul | **Metode V1**<br>*(7 Container Docker Fisik)* | **Metode V2**<br>*(1 Container + 7 DB Terisolasi)* | **Metode V3**<br>*(1 Container + 1 Biner Suite)* |
| :--- | :---: | :---: | :---: |
| **1. `category`** | `16.48 detik` | `13.00 detik` | *(Dijalankan dalam 1 suite)* |
| **2. `customer`** | `16.25 detik` | `12.98 detik` | *(Dijalankan dalam 1 suite)* |
| **3. `inventory`** | `16.63 detik` | `12.98 detik` | *(Dijalankan dalam 1 suite)* |
| **4. `product`** | `16.55 detik` | `12.94 detik` | *(Dijalankan dalam 1 suite)* |
| **5. `purchase`** | `16.14 detik` | `12.97 detik` | *(Dijalankan dalam 1 suite)* |
| **6. `sales`** | `16.31 detik` | `12.88 detik` | *(Dijalankan dalam 1 suite)* |
| **7. `supplier`** | `16.39 detik` | `12.91 detik` | *(Dijalankan dalam 1 suite)* |
| **TOTAL WAKTU (7 Modul)** | **~16.6 detik** | **~13.0 detik** | **`5.43 detik` ⚡ (Paling Cepat)** |

---

### B. Pengujian Partial Harian (1 Modul saat Coding TDD)

Saat developer sedang fokus membuat fitur baru pada 1 modul (misal: `customer`):

| Skenario Pengujian | **Metode V1** | **Metode V2** | **Metode V3 (`-run`)** |
| :--- | :---: | :---: | :---: |
| **Command** | `make test-customer` | `make test-v2-customer` | `make test-v3-customer` |
| **Cold Run (`-count=1`)** | `5.98 detik` | `4.08 detik` | **`4.84 detik`** |
| **Inisialisasi DB** | `~4.5 detik` (spin up) | `0.023 detik (23 ms)` | **`0.024 detik (24 ms)`** |
| **Warm Run (Cached binary)** | `~5.5 detik` | `0.02 detik` | **`0.01 detik` (Instan)** |

---

### C. Rincian Penggunaan Resource Komputer

| Metrik Resource | **Metode V1** | **Metode V2** | **Metode V3** |
| :--- | :---: | :---: | :---: |
| **Jumlah File Biner Go** | 7 Biner `.exe` | 7 Biner `.exe` | **Hanya 1 Biner `.exe`** |
| **Jumlah Container Docker** | **7 Container Aktif Bersamaan** | **1 Container Bersama** | **1 Container Bersama** |
| **Database yang Dipakai** | 1 DB per container | 7 Database terpisah | 1 Database (`v3_feature_test`) |
| **Pembersihan Data** | Hancurkan 7 container | `tx.Rollback()` per test | `tx.Rollback()` per test |
| **Konsumsi RAM Docker** | **~1.4 GB – 2.1 GB** | **~150 MB – 220 MB** | **~150 MB – 200 MB** |
| **Beban CPU saat Kompilasi** | Sangat tinggi (compile 7x) | Sangat tinggi (compile 7x) | **Rendah (compile 1x saja)** |



