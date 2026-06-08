# 🚀 Panduan Deployment SIGAP2

Panduan ini menjelaskan cara men-deploy aplikasi SIGAP2 ke server production/cloud.

## Prasyarat

- **Docker** & **Docker Compose** terinstall di server
- Akses SSH ke server VPS (atau akun di platform cloud)
- Repository GitHub sudah di-push

---

## Opsi 1: Deploy ke Railway.app (Rekomendasi — Gratis)

[Railway](https://railway.app) mendukung deployment Docker langsung dari GitHub.

### Langkah-langkah:

1. **Buat akun Railway** di https://railway.app (login via GitHub)

2. **New Project → Deploy from GitHub repo**
   - Pilih repository `sigap2`
   - Railway akan otomatis mendeteksi `Dockerfile`

3. **Tambahkan MySQL Database**
   - Klik **"+ New"** → **"Database"** → **"MySQL"**
   - Railway akan otomatis membuat instance MySQL

4. **Set Environment Variables** di Railway dashboard:
   ```
   DB_HOST=<railway-mysql-host>
   DB_PORT=3306
   DB_USER=root
   DB_PASS=<railway-mysql-password>
   DB_NAME=railway
   JWT_SECRET=your-super-secret-jwt-key-production
   APP_PORT=3000
   GEMINI_API_KEY=your_gemini_api_key
   ```
   > **Tip**: Railway menyediakan variable referensi otomatis untuk MySQL, gunakan `${{MySQL.MYSQL_HOST}}` dll.

5. **Generate Domain**
   - Settings → Networking → Generate Domain
   - Aplikasi akan tersedia di: `https://sigap2-production.up.railway.app`

---

## Opsi 2: Deploy ke VPS (DigitalOcean / Hetzner / IDCloudHost)

### 1. Siapkan VPS

Buat VPS dengan spesifikasi minimum:
- **OS**: Ubuntu 22.04 LTS
- **RAM**: 1GB
- **Storage**: 25GB SSD
- **Harga**: ~$5-6/bulan (DigitalOcean), ~€3.29/bulan (Hetzner)

### 2. Install Docker di VPS

```bash
# SSH ke server
ssh root@YOUR_SERVER_IP

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verifikasi
docker --version
docker compose version
```

### 3. Clone & Deploy

```bash
# Clone repository
git clone https://github.com/YOUR_USERNAME/sigap2.git
cd sigap2

# Buat file .env untuk production
cp .env.example .env
nano .env
```

Edit file `.env`:
```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=sigap
DB_PASS=sigap_secret_production_strong_password
DB_NAME=sigap2
JWT_SECRET=your-super-secret-jwt-key-production-change-this
APP_PORT=3000
GEMINI_API_KEY=your_gemini_api_key_here
```

```bash
# Build & jalankan dengan Docker Compose
docker compose up -d --build

# Cek status
docker compose ps
docker compose logs -f app
```

### 4. Setup Reverse Proxy (Nginx) — Opsional tapi Rekomendasi

```bash
# Install Nginx
sudo apt install nginx -y

# Buat konfigurasi
sudo nano /etc/nginx/sites-available/sigap2
```

Isi konfigurasi:
```nginx
server {
    listen 80;
    server_name sigap2.yourdomain.com;  # Ganti dengan domain Anda

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    client_max_body_size 10M;  # Untuk upload foto bukti
}
```

```bash
# Aktifkan konfigurasi
sudo ln -s /etc/nginx/sites-available/sigap2 /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 5. Setup SSL (HTTPS) — Opsional

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Generate SSL certificate
sudo certbot --nginx -d sigap2.yourdomain.com
```

---

## Opsi 3: Deploy ke Render.com (Gratis)

1. Buat akun di https://render.com
2. **New → Web Service** → Connect GitHub repo
3. Pilih **Docker** sebagai Environment
4. Set Environment Variables (sama seperti Railway)
5. Klik **Deploy**
6. Untuk database MySQL, gunakan layanan external seperti:
   - **PlanetScale** (MySQL gratis)
   - **Aiven** (MySQL gratis tier)

---

## Verifikasi Deployment

Setelah deploy, pastikan:

1. **Halaman login** bisa diakses: `https://YOUR_URL/login`
2. **SOS form** bisa diakses: `https://YOUR_URL/sos`
3. **API endpoint** merespons: `curl https://YOUR_URL/api/sos -X POST -d "latitude=-6.2&longitude=106.8&description=test"`
4. **Docker containers** berjalan: `docker compose ps` (harus menunjukkan `Up`)

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| Container restart loop | Cek logs: `docker compose logs app` — biasanya masalah koneksi DB |
| Database connection refused | Pastikan `DB_HOST=mysql` (bukan localhost) di docker-compose |
| Port sudah dipakai | Ubah port mapping di `docker-compose.yml` |
| Upload foto gagal | Pastikan volume `./web:/root/web` ter-mount dengan benar |
