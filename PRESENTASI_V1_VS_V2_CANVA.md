# Modernisasi Integration Testing: V1 vs V2
## Menghilangkan Bottleneck Docker & Meningkatkan Produktivitas Developer
Tim Backend Engineering

---

# Problem Statement: Kendala pada Metode V1
- **Testing Lambat**: Full test suite memakan waktu ~16.6 detik hanya untuk 7 modul.
- **Overhead Docker Berat**: Setiap folder modul menyalakan & mematikan container Postgres baru (overhead 8–10 detik per modul).
- **Boros Resource**: Memakan RAM ~1.4 GB – 2.1 GB karena menjalankan 7 container sekaligus.
- **Rentan Tabrakan Data**: Data uji di dalam satu modul tidak otomatis bersih jika tidak dihapus manual.

---

# Pergeseran Paradigma Arsitektur
### Metode V1: Container per Modul
- 7 Modul = 7 Container Postgres Terpisah
- Menggunakan Docker Named Pipe (npipe) yang lambat di Windows
- Siklus berulang: Spawn Container -> Run Test -> Destroy Container

### Metode V2: Reusable Container + TCP Fast-Path
- 7 Modul = 1 Shared Container Postgres Standby
- Menggunakan Fast-Path TCP Port lokal (< 30 ms)
- Siklus efisien: Cek Port -> Auto-create Isolated DB -> Test with Auto-Rollback

---

# 3 Kunci Inovasi Metode V2
- **TCP Fast-Path Bypass**: Menghubungkan langsung ke port TCP 5434/5432 dalam ~26 ms tanpa melalui Docker Named Pipe.
- **Dynamic Database Provisioning**: Setiap modul otomatis dibuatkan database fisik terpisah (v2_product_test, v2_customer_test) di container yang sama.
- **DatabaseTransactions (Auto-Rollback)**: Setiap fungsi test dibungkus transaksi database dan di-rollback otomatis saat selesai. Data bersih 100% tanpa truncate table.

---

# Developer Experience: Perbandingan Kode
### Metode V1 (Tradisional)
- JSON marshaling & unmarshaling manual
- Validasi status code dan payload baris per baris
- Query database manual dengan GORM untuk verifikasi data

### Metode V2 (Fluent Chaining ala Laravel)
- Request JSON deklaratif: this.PostJSON("/api/v1/products", payload)
- Fluent assertion: .AssertCreated().AssertJsonPath("data.name", "...")
- Verifikasi DB instan: this.AssertDatabaseHas("products", map)
- Transaksi otomatis rollback setelah pengujian selesai

---

# Benchmark Head-to-Head (7 Modul Riil)
- **Container Docker**: V1 (7 Container Fisik) vs V2 (1 Container Bersama) -> Hemat 85% Container
- **Konsumsi RAM Docker**: V1 (~1.5–2 GB) vs V2 (~200 MB) -> Hemat RAM ~1.5 GB
- **Inisialisasi Database**: V1 (~4.5 detik) vs V2 (~0.023 detik / 23 ms) -> 200x Lebih Cepat
- **TDD Warm Run (Saat Coding)**: V1 (~5.5 detik) vs V2 (0.02 detik) -> Feedback Instan
- **Total Waktu 7 Modul**: V1 (~16.6 detik) vs V2 (~13.0 detik)

---

# Kesimpulan & Nilai Tambah
- **Tetap Modular**: Struktur package Go tetap rapi per folder modul (product, purchase, customer).
- **Isolasi Terjamin**: Database fisik terpisah per modul, bebas risiko tabrakan data antar-domain.
- **Laptop Developer Aman**: RAM dan CPU tidak terbebani tumpukan container Docker.
- **TDD Jauh Lebih Cepat**: Waktu tunggu berkurang dari hitungan detik menjadi milidetik.

---

# Rencana Aksi (Next Steps)
- Standarisasi helper testcase.go ke shared library tim backend.
- Sediakan konfigurasi 1 container Postgres reusable di lokal via Makefile / Docker Compose.
- Migrasi test suite modul yang ada secara bertahap ke pola fluent V2.
