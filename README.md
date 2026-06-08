# 🆘 SIGAP2 — Sistem Informasi Tanggap Darurat Bencana

<p align="center">
  <strong>Platform tanggap darurat bencana dengan AI-powered NLP untuk analisis urgensi otomatis dan live GPS tracking distribusi bantuan.</strong>
</p>

---

## 📋 Deskripsi

SIGAP2 adalah sistem informasi berbasis web yang dirancang untuk mengkoordinasikan respons terhadap bencana alam. Platform ini menghubungkan **korban bencana**, **relawan lapangan**, dan **administrator** dalam ekosistem yang terintegrasi.

### Fitur Utama

| Fitur | Deskripsi |
|-------|-----------|
| 🆘 **SOS Report** | Korban mengirim laporan darurat dengan lokasi GPS |
| 🤖 **AI NLP Analysis** | Analisis otomatis urgensi & kebutuhan menggunakan Google Gemini |
| 📦 **Distribusi Logistik** | Manajemen stok gudang & pengiriman bantuan |
| 📍 **Live GPS Tracking** | Pelacakan posisi relawan secara real-time di peta |
| ✅ **Verifikasi Pengiriman** | Upload foto bukti & verifikasi oleh admin |
| 👥 **Multi-Role** | 3 role: Admin, Relawan, Korban |
| 📊 **Export CSV** | Export laporan ke format CSV |

## 🏗️ Tech Stack

| Layer | Teknologi |
|-------|-----------|
| **Backend** | Go (Golang) + [Go Fiber v2](https://gofiber.io/) |
| **Database** | MySQL 8.0 + [GORM](https://gorm.io/) ORM |
| **Frontend** | HTML Templates (Server-Side Rendered) + JavaScript |
| **Maps** | Leaflet.js |
| **AI/NLP** | Google Gemini API + Regex Fallback |
| **Auth** | JWT (HTTP-Only Cookie) + bcrypt |
| **Containerization** | Docker + Docker Compose |

## 📁 Struktur Project

```
sigap2/
├── cmd/
│   ├── server/          # Entry point aplikasi
│   ├── seed_logistics/  # Seeder data logistik
│   └── test_nlp/        # Testing NLP service
├── internal/
│   ├── config/          # Konfigurasi aplikasi (.env)
│   ├── database/        # Koneksi & migrasi database
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/       # Auth & logging middleware
│   ├── models/          # Data models (GORM)
│   ├── routes/          # Route definitions
│   └── services/        # Business logic layer
├── web/
│   ├── static/          # CSS, JS, Images, Uploads
│   └── templates/       # HTML templates (10 modul)
├── docs/
│   ├── swagger.yaml     # API Documentation (OpenAPI 3.0)
│   └── SIGAP2_Postman_Collection.json
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose (app + MySQL)
├── DEPLOYMENT.md        # Panduan deployment
└── README.md
```

## 🚀 Quick Start

### Prasyarat
- Go 1.21+
- MySQL 8.0
- (Opsional) Docker & Docker Compose

### Menjalankan dengan Docker (Rekomendasi)

```bash
# Clone repository
git clone https://github.com/YOUR_USERNAME/sigap2.git
cd sigap2

# Copy environment file
cp .env.example .env
# Edit .env sesuai kebutuhan

# Jalankan
docker compose up -d --build

# Akses di http://localhost:3000
```

### Menjalankan Lokal (Tanpa Docker)

```bash
# Clone & masuk direktori
git clone https://github.com/YOUR_USERNAME/sigap2.git
cd sigap2

# Copy dan edit .env
cp .env.example .env

# Pastikan MySQL sudah running dan database sudah dibuat
mysql -u root -e "CREATE DATABASE sigap2;"

# Jalankan
go run ./cmd/server

# Akses di http://localhost:3000
```

### Akun Default

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@sigap.id | admin123 |
| Relawan | relawan@sigap.id | relawan123 |

## 📖 Dokumentasi API

Dokumentasi API tersedia dalam dua format:

1. **Swagger/OpenAPI 3.0**: [`docs/swagger.yaml`](docs/swagger.yaml)
   - Buka di [Swagger Editor](https://editor.swagger.io/) dengan import file
2. **Postman Collection**: [`docs/SIGAP2_Postman_Collection.json`](docs/SIGAP2_Postman_Collection.json)
   - Import langsung ke Postman

### Endpoint Utama

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| `POST` | `/api/sos` | ❌ | Kirim laporan SOS darurat |
| `GET` | `/api/reports/markers` | ✅ | Ambil marker peta |
| `GET` | `/api/notifications` | ✅ | Notifikasi laporan pending |
| `POST` | `/api/deliveries/{id}/location` | ✅ | Update lokasi GPS relawan |
| `GET` | `/api/deliveries/active` | ✅ | Daftar pengiriman aktif |
| `GET` | `/api/deliveries/{id}/location` | ✅ | Lokasi pengiriman tertentu |

## 🌐 Deployment

Lihat panduan lengkap di [`DEPLOYMENT.md`](DEPLOYMENT.md).

Opsi deployment:
- **Railway.app** — Deploy dari GitHub, gratis
- **Render.com** — Docker support, gratis
- **VPS** (DigitalOcean/Hetzner) — Full control, ~$5/bulan

## 👥 Tim Pengembang

| Nama | Role | Kontribusi |
|------|------|------------|
| - | - | - |

## 📄 Lisensi

MIT License
