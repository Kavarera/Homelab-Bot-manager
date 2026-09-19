# 🤖 HS1 Telegram Server & Invoice Management Bot

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Telegram Bot API](https://img.shields.io/badge/Telegram_Bot_API-v5-2CA5E0?style=flat&logo=telegram)](https://core.telegram.org/bots/api)
[![SQLite](https://img.shields.io/badge/SQLite-Pure_Go-003B57?style=flat&logo=sqlite)](https://modernc.org/sqlite)
[![Architecture](https://img.shields.io/badge/Architecture-Clean_Architecture-brightgreen?style=flat)](#-arsitektur-proyek)

Bot Telegram berbasis Golang yang dirancang untuk mengelola penagihan invoice otomatis, manajemen data client dan katalog produk, pengiriman email invoice profesional beserta lampiran PDF resmi, serta monitoring status server / homelab secara terpusat dan aman.

---

## 📌 Daftar Isi

- [✨ Fitur Utama](#-fitur-utama)
- [🏗️ Arsitektur Proyek](#-arsitektur-proyek)
- [📋 Prasyarat Sistem](#-prasyarat-sistem)
- [⚙️ Panduan Instalasi & Konfigurasi](#️-panduan-instalasi--konfigurasi)
- [🚀 Menjalankan Bot](#-menjalankan-bot)
- [📖 Panduan Penggunaan (Workflow Bot)](#-panduan-penggunaan-workflow-bot)
- [🧪 Menjalankan Pengujian (Testing)](#-menjalankan-pengujian-testing)
- [🛠️ Deployment dengan Systemd (Linux/VPS)](#️-deployment-dengan-systemd-linuxvps)

---

## ✨ Fitur Utama

### 1. 📄 Penerbitan & Pengiriman Invoice (Multi-Produk)

- **Multi-Product Selection**: Memilih satu atau beberapa produk layanan sekaligus dalam satu invoice menggunakan Reply Keyboard interaktif.
- **Anti-Duplikasi Bulanan**: Mencegah penerbitan invoice ganda untuk produk yang sama kepada client yang sama pada bulan berjalan.
- **Format Penomoran Standar**: Format nomor invoice otomatis:
  $$\text{INV/}\{\text{ID\_CLIENT}\}/\{\text{TAHUN}\}/\{\text{BULAN}\}/\{\text{TGL}\}/\{\text{ID\_INVOICE}\}$$
  _(Contoh: `INV/1/2026/09/19/1`)_
- **Kalkulasi Jatuh Tempo Otomatis**: Default jatuh tempo $+14$ hari. Jika jatuh pada hari Sabtu atau Minggu, otomatis digeser ke hari Senin.
- **Pembuatan PDF Resmi (Pure-Go CGO-Free)**: Otomatis merender dokumen PDF berisi tabel rincian produk, informasi rekening pembayaran, logo perusahaan, stempel, dan tanda tangan resmi.
- **Pengiriman Email Otomatis**: Mengirim email berformat HTML profesional dengan lampiran PDF ke email perusahaan client via SMTP.
- **Pengiriman PDF ke Telegram**: File PDF invoice juga langsung dikirimkan ke chat Telegram admin sebagai arsip instan.
- **Proteksi Transaksi DB**: Invoice hanya disimpan ke database jika email berhasil terkirim. Jika SMTP gagal, pembuatan invoice dibatalkan secara aman (tidak ada data korup di DB).

### 2. 🏢 Manajemen Client

- **Tambah Client Baru**: Input Nama Perusahaan, Nama PIC, Alamat, Email, serta pemilihan produk-produk yang dilanggan.
- **Edit Client**: Pembaruan data fleksibel per kolom dengan dukungan tombol `⏩ Skip`, serta fitur tambah produk langganan (`➕ Tambah Produk`) dan hapus produk (`➖ Hapus Produk`).
- **Hapus Client (Soft Delete)**: Menghapus client dari daftar aktif tanpa menghilangkan riwayat invoice terdahulu di database (`deleted_at` timestamp).

### 3. 📦 Manajemen Produk

- **Tambah Produk**: Input nama produk dan harga dengan parser harga cerdas (misal: `400000`, `400k`, `1.5M`, `Rp 400.000`).
- **Edit Produk**: Perbarui nama atau harga produk yang sudah ada secara mandiri.
- **Hapus Produk (Soft Delete)**: Menghapus produk dari katalog aktif tanpa merusak integritas relasi item invoice lama (`deleted_at` timestamp).

### 4. 💾 Backup Database Remote VPS (PostgreSQL)

- **Deteksi Otomatis Kontainer PostgreSQL**: Bot mengecek container Docker yang berjalan di VPS target melalui SSH (Tailscale VPN) dan secara cerdas memfilter kontainer yang berbasis image `postgres`.
- **Zero Disk Garbage (Direct Streaming)**: Dump PostgreSQL dialirkan secara *real-time* via SSH `stdout` (`docker exec ... pg_dumpall`). Tidak ada file temporary atau sampah dump yang ditinggalkan di disk VPS.
- **Kompresi Gzip & Enkripsi AES-256-GCM**: Stream dump dikompresi dengan Gzip dan dienkripsi menggunakan AES-256-GCM dengan kunci acak 32-byte berkekuatan militer.
- **Penyimpanan Lokal Homelab Terstruktur**: File backup `.sql.gz.enc` disimpan di server homelab dengan struktur direktori rapi:
  $$\text{/home/kava/backup\_db/}\{\text{container\_name}\}/\{\text{YYYYMMDD}\}/\text{backup\_}\{\text{container}\}\_\{\text{timestamp}\}\text{.sql.gz.enc}$$
- **Kirim Dokumen ke Telegram**: Bot langsung mengirimkan file backup terenkripsi ke chat Telegram admin.
- **Self-Destruct Encryption Key Message**: Kunci dekripsi dikirimkan dalam pesan terpisah dengan tombol **`🗑️ Hapus Pesan Kunci Sekarang`** untuk penghapusan instan, serta **Timer Goroutine 5 Menit** yang otomatis menghapus pesan kunci jika admin lupa menghapusnya secara manual.

### 5. 🖥️ Monitoring Server

- Command `/status` untuk memeriksa status host, uptime sistem, dan informasi lingkungan server.

### 6. 🛡️ Keamanan & Antarmuka Interaktif

- **Whitelist Authorization**: Hanya Telegram User ID yang terdaftar di `.env` yang dapat mengakses bot.
- **Panic Recovery & Structured Logging**: Dilengkapi middleware panic recovery dan logging terstruktur via `log/slog`.
- **Stateful Flow Engine**: Dialog multi-langkah berbasis sesi in-memory dengan auto-cleanup dan tombol pembatalan (`❌ Batalkan`).
- **Reply Keyboard Cepat**: Akses seluruh menu melalui tombol keyboard tanpa perlu menghafal perintah teks.
- **Tombol Tutup Menu**: Tombol `🔽 Tutup Menu` untuk menyembunyikan keyboard saat sedang tidak digunakan.

---

## 🏗️ Arsitektur Proyek

Proyek ini dibangun mengikuti prinsip **Clean Architecture & Separation of Concerns**:

```
hs1-bot/
├── assets/                  # Asset fisik (logo.png, signature.png)
├── cmd/
│   └── bot/
│       └── main.go          # Entry point aplikasi & dependency wiring
├── config/                  # Konfigurasi & loader .env
├── internal/
│   ├── assets/              # Embedded assets via //go:embed
│   ├── bot/                 # Telegram Bot router, middleware, poller, & context
│   │   ├── flow/            # State machine multi-step wizard (Invoice, Client, Produk, Delete)
│   │   ├── handlers/        # Command handlers (Help, System Status)
│   │   ├── middleware/      # Auth, Logger, & Recover middleware
│   │   ├── session/         # In-memory session store dengan TTL & auto-cleanup
│   │   └── ui/              # Reply keyboard builders & helper parsing
│   ├── domain/              # Entity domain & business rules (Client, Invoice, Product)
│   ├── executor/            # Eksekutor perintah OS lokal
│   ├── repository/          # SQLite database repository & data seeders
│   └── service/             # Business services (Email, PDF Generator, System)
│       └── template/        # HTML Invoice email template renderer
├── pkg/
│   └── logger/              # Inisialisasi structured logger (slog)
├── .env.example             # Template variabel environment
├── .gitignore               # Konfigurasi file ignore Git
├── go.mod                   # Dependency Go modules
└── README.md                # Dokumentasi proyek
```

---

## 📋 Prasyarat Sistem

Sebelum menjalankan bot, pastikan telah menyiapkan:

1. **Go (Golang)**: Versi 1.22 atau yang lebih baru ([Unduh Go](https://go.dev/dl/)).
2. **Akun & Bot Telegram**:
   - Buat bot baru melalui [@BotFather](https://t.me/BotFather) di Telegram untuk mendapatkan **Bot Token**.
   - Dapatkan Telegram User ID Anda melalui [@userinfobot](https://t.me/userinfobot).
3. **Akun SMTP Email**:
   - Host SMTP, Port (misal: port `465` dengan SSL atau `587` dengan STARTTLS), Username, dan Password (atau App Password untuk Gmail).

---

## ⚙️ Panduan Instalasi & Konfigurasi

### 1. Clone Repository

```bash
git clone https://github.com/username/Golang-Server-Bot-Tele.git
cd Golang-Server-Bot-Tele/hs1-bot
```

### 2. Konfigurasi Environment (`.env`)

Salin file `.env.example` menjadi `.env`:

```bash
cp .env.example .env
```

Buka dan sesuaikan file `.env`:

```ini
# ==========================================
# KONFIGURASI TELEGRAM BOT
# ==========================================
# Token Bot yang didapatkan dari @BotFather
TELEGRAM_TOKEN=123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ

# Telegram User ID Anda (hanya user ini yang diizinkan mengakses bot)
ALLOWED_USER_ID=123456789

# ==========================================
# KONFIGURASI SERVER, SSH & DATABASE
# ==========================================
# Nama server / homelab yang ditampilkan di bot
SERVER_NAME=HS1 Homelab Server

# Lokasi penyimpanan file database SQLite internal bot
DATABASE_PATH=data/bot.db

# Koneksi SSH ke VPS target (menggunakan IP Tailscale VPN / Host SSH)
SSH_HOST=payrollpro@100.x.x.x
SSH_PRIVATE_KEY_PATH=~/.ssh/id_rsa

# Folder penyimpanan arsip backup database di homelab
BACKUP_BASE_PATH=/home/kava/backup_db

# ==========================================
# KONFIGURASI SMTP EMAIL
# ==========================================
SMTP_HOST=smtp.example.com
SMTP_PORT=465
SMTP_SSL=true
SMTP_USERNAME=billing@example.com
SMTP_PASSWORD=RahasiaPasswordSMTP123
SMTP_SENDER_NAME=PT. Kavarera Kreasi Teknologi
SMTP_SENDER_EMAIL=billing@example.com
```

> [!NOTE]
> File `.env`, database `*.db`, dan file backup `*.enc` telah otomatis dikecualikan (`.gitignore`) agar data rahasia tidak ter-commit ke Git.

---

## 🚀 Menjalankan Bot

### Mode Development

Jalankan bot secara langsung menggunakan `go run`:

```bash
go run ./cmd/bot
```

### Mode Production (Compile Binary)

Build binary yang dapat dieksekusi:

```bash
# Build untuk Linux / macOS
go build -o hs1-bot ./cmd/bot

# Cross-compile dari Windows untuk Linux (Ubuntu Homelab)
# PowerShell:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o hs1-bot ./cmd/bot

# Build untuk Windows
go build -o hs1-bot.exe ./cmd/bot
```

Jalankan binary hasil build:

```bash
./hs1-bot
```

---

## 📖 Panduan Penggunaan (Workflow Bot)

Buka Telegram dan kirim perintah `/start` atau `/menu` ke bot Anda.

### 📱 Menu Utama

```
+-------------------------------------------------------------+
|                      📄 Kirim Invoice                       |
+---------------------+---------------------+-----------------+
| 🏢 Tambah Client    | ✏️ Edit Client      | 🗑️ Hapus Client  |
+---------------------+---------------------+-----------------+
| 📦 Tambah Produk    | ⚙️ Edit Produk      | 🗑️ Hapus Produk  |
+---------------------+---------------------+-----------------+
| 💾 Backup Database  | 🔽 Tutup Menu                         |
+---------------------+---------------------------------------+
```

---

### 1. 📄 Alur Kirim Invoice (Multi-Produk)

1. Klik tombol **`📄 Kirim Invoice`**.
2. **Pilih Client**: Klik tombol nama perusahaan client yang dituju.
3. **Pilih Produk**:
   - Klik nama produk untuk memilih atau membatalkan pilihan (indikator `✅` akan muncul pada produk terpilih).
   - Anda dapat memilih 1 atau lebih produk sekaligus.
   - Klik tombol **`✅ Submit`** jika sudah selesai memilih.
4. **Pemrosesan Otomatis**:
   - Sistem memvalidasi agar tidak ada produk yang ditagihkan dua kali pada bulan yang sama.
   - Merender file PDF invoice dengan kop surat, logo resmi, dan stempel tanda tangan.
   - Mengirim email penagihan berformat HTML beserta lampiran PDF ke email client.
   - Mengirimkan dokumen PDF ke chat Telegram Anda.
   - Menyimpan riwayat invoice ke database SQLite.

---

### 2. 🏢 Alur Manajemen Client (Tambah, Edit, Hapus)

- **Tambah Client**:
  1. Masukkan Nama Perusahaan (contoh: `PT. Maju Bersama`).
  2. Masukkan Nama PIC (contoh: `Budi Santoso`).
  3. Masukkan Alamat Kantor.
  4. Masukkan Email Perusahaan (contoh: `finance@majubersama.com`).
  5. Pilih produk langganan client $\rightarrow$ Klik **`✅ Konfirmasi`**.
- **Edit Client**:
  1. Pilih client yang akan diedit.
  2. Masukkan nilai baru atau klik **`⏩ Skip`** untuk melewati field tertentu.
  3. Kelola produk langganan: klik **`➕ Tambah Produk`**, **`➖ Hapus Produk`**, atau **`⏩ Skip`**.
  4. Konfirmasi perubahan.
- **Hapus Client (Soft Delete)**:
  1. Klik tombol **`🗑️ Hapus Client`**.
  2. Pilih client dari daftar tombol.
  3. Periksa ringkasan client $\rightarrow$ Klik **`✅ Konfirmasi`** untuk menonaktifkan.

---

### 3. 📦 Alur Manajemen Produk (Tambah, Edit, Hapus)

- **Tambah Produk**:
  1. Masukkan Nama Produk / Layanan (contoh: `Dedicated Server Epyc`).
  2. Masukkan Harga Produk:
     - Mendukung format angka langsung: `500000`
     - Mendukung singkatan ribuan/jutaan: `500k`, `1.5M`, `2.5jt`
     - Mendukung format mata uang: `Rp 500.000`
  3. Konfirmasi ringkasan harga $\rightarrow$ Tersimpan di database.
- **Edit Produk**:
  1. Pilih produk yang ingin diubah.
  2. Ketik nama baru atau klik **`⏩ Skip`**.
  3. Ketik harga baru atau klik **`⏩ Skip`**.
  4. Konfirmasi perubahan.
- **Hapus Produk (Soft Delete)**:
  1. Klik tombol **`🗑️ Hapus Produk`**.
  2. Pilih produk dari daftar tombol.
  3. Periksa ringkasan produk & harga $\rightarrow$ Klik **`✅ Konfirmasi`** untuk menonaktifkan.

---

### 4. 💾 Alur Backup Database VPS (PostgreSQL)

1. Klik tombol **`💾 Backup Database`** atau ketik command `/backup`.
2. **Pilih Kontainer PostgreSQL**:
   - Bot otomatis menginspeksi VPS melalui SSH dan hanya menampilkan container yang berbasis image PostgreSQL (misal: `payroll_db_postgres`).
   - Klik nama container target.
3. **Konfirmasi & Eksekusi**:
   - Bot menampilkan detail container (Nama & Image ID) $\rightarrow$ Klik **`✅ Konfirmasi`**.
4. **Proses Backup Otomatis**:
   - Bot melakukan *direct streaming* `pg_dumpall` via SSH tanpa meninggalkan file temporary di VPS.
   - Stream data dikompresi (Gzip) dan dienkripsi dengan AES-256-GCM menggunakan kunci 32-byte unik.
   - File tersimpan di homelab: `/home/kava/backup_db/{container}/{YYYYMMDD}/backup_{container}_{timestamp}.sql.gz.enc`.
   - Bot mengirim dokumen `.sql.gz.enc` ke chat Telegram.
   - Bot mengirim pesan kunci dekripsi rahasia dengan tombol **`🗑️ Hapus Pesan Kunci Sekarang`**.
   - **Auto Self-Destruct**: Pesan kunci akan otomatis terhapus dalam waktu 5 menit jika tidak dihapus manual.

---

### 5. ⌨️ Tombol Navigasi & Bantuan

- **`❌ Batalkan`**: Membatalkan flow/wizard yang sedang berjalan kapan saja.
- **`⏩ Skip`**: Melewati langkah edit tanpa mengubah data sebelumnya.
- **`🔽 Tutup Menu`**: Menyembunyikan Reply Keyboard saat tidak digunakan.
- **`/status`**: Memeriksa kondisi sistem dan uptime server.
- **`/backup`**: Memulai wizard backup database remote VPS.
- **`/help`**: Menampilkan daftar perintah dan panduan bantuan.

---

## 🧪 Menjalankan Pengujian (Testing)

Proyek ini dilengkapi dengan rangkaian unit test menyeluruh untuk memastikan keandalan logika bisnis:

```bash
# Jalankan seluruh unit test
go test -v ./...

# Jalankan pengujian dengan kalkulasi coverage
go test -cover ./...
```

---

## 🛠️ Deployment dengan Systemd (Linux/VPS)

Untuk menjalankan bot sebagai service background di Linux (Ubuntu/Debian) yang otomatis aktif saat booting:

1. Buat file service systemd:

   ```bash
   sudo nano /etc/systemd/system/hs1-bot.service
   ```

2. Tambahkan konfigurasi berikut:

   ```ini
   [Unit]
   Description=HS1 Telegram Server & Invoice Bot
   After=network.target

   [Service]
   Type=simple
   User=ubuntu
   WorkingDirectory=/home/ubuntu/hs1-bot
   ExecStart=/home/ubuntu/hs1-bot/hs1-bot
   Restart=always
   RestartSec=5s
   EnvironmentFile=/home/ubuntu/hs1-bot/.env

   [Install]
   WantedBy=multi-user.target
   ```

3. Reload daemon dan jalankan service:

   ```bash
   # Reload systemd configuration
   sudo systemctl daemon-reload

   # Enable agar service berjalan otomatis saat startup
   sudo systemctl enable hs1-bot

   # Jalankan service bot
   sudo systemctl start hs1-bot

   # Cek status bot
   sudo systemctl status hs1-bot
   ```

4. Untuk melihat log aktivitas bot secara realtime:
   ```bash
   journalctl -u hs1-bot -f
   ```

---
