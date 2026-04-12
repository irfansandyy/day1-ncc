# Laporan Implementasi Aplikasi Chat AI

## A. Deskripsi Service

Aplikasi yang dibuat adalah platform chat AI full-stack yang terdiri dari frontend Next.js (App Router), backend Golang REST API, PostgreSQL, dan model AI lokal menggunakan Docker Model Runner.

Untuk deployment publik, arsitektur sekarang menggunakan reverse proxy Caddy sehingga seluruh akses user masuk lewat HTTPS port 443 dengan sertifikat TLS otomatis dari Let's Encrypt.

Fitur utama:

- Registrasi dan login user
- Autentikasi JWT untuk endpoint terproteksi
- Penyimpanan riwayat chat per user secara persisten
- Fitur New Chat untuk memulai percakapan baru
- Penyimpanan pesan user dan AI ke database
- Integrasi model `meta-llama/Llama-3.2-3B-Instruct` dari Hugging Face melalui Docker Model Runner

## B. Penjelasan Endpoint /health

Fungsi endpoint `/health` adalah untuk memantau kesiapan service backend dan dependency utama.

Endpoint:

- `GET /health`

Contoh response extended:

```json
{
  "status": "ok",
  "services": {
    "database": "ok",
    "llm": "ok"
  }
}
```

Contoh response sederhana:

```json
{
  "status": "ok"
}
```

Cara kerja pengecekan service:

- Database dicek dengan `PingContext` ke PostgreSQL
- LLM dicek ringan ke endpoint Docker Model Runner (`/v1/models`)
- Endpoint tetap mengembalikan HTTP 200 agar orchestration health monitor tetap stabil

## C. Bukti Akses Endpoint

Endpoint dapat diakses publik melalui reverse proxy di port 443 (HTTPS). Service backend/frontend tidak diekspos langsung ke internet.

Contoh URL publik:

- `https://<domain-anda>/health`

Verifikasi cepat:

```bash
curl https://<domain-anda>/health
```

## D. Proses Build & Run Docker

Langkah build image dan menjalankan container:

1. Salin env file:
   - `cp .env.example .env`
2. Isi variabel reverse proxy/TLS pada `.env`:
   - `DOMAIN=<domain-anda>`
   - `ACME_EMAIL=<email-anda>`
3. Login ke Hugging Face dan jalankan model di Docker Model Runner:
   - `hf auth login`
   - `./scripts/docker-model-run.sh hf.co/meta-llama/Llama-3.2-3B-Instruct`
4. Build dan jalankan stack:
   - `docker compose --env-file .env up -d --build`
5. Cek status dan health:
   - `docker compose ps`
   - `curl https://<domain-anda>/health`

Penggunaan docker-compose:

- Menjalankan service `db`, `backend`, `frontend`, `reverse-proxy`
- Semua service memiliki `healthcheck`
- Dependency antar service menggunakan `depends_on` dengan `condition: service_healthy`
- Semua service menggunakan restart policy `unless-stopped`

## E. Proses Deployment ke VPS (DigitalOcean)

Langkah deploy ke server:

1. Buat droplet Ubuntu di DigitalOcean
2. Install Docker Engine dan Docker Compose plugin
3. Clone repository aplikasi ke VPS
4. Salin `.env.example` menjadi `.env`, lalu sesuaikan variabel produksi
5. Login Hugging Face dan jalankan model di Docker Model Runner:
   - `hf auth login`
   - `./scripts/docker-model-run.sh hf.co/meta-llama/Llama-3.2-3B-Instruct`
6. Jalankan:
   - `docker compose --env-file .env up -d --build`
7. Atur DNS domain ke IP VPS (A record)
8. Atur firewall DigitalOcean dan UFW untuk membuka port yang dibutuhkan:
   - `80/tcp` untuk challenge Let's Encrypt dan redirect
   - `443/tcp` untuk trafik HTTPS utama
9. Uji endpoint publik:
   - `curl https://<domain-anda>/health`

Konfigurasi yang dilakukan:

- Konfigurasi environment production
- Reverse proxy Caddy sebagai single entrypoint publik di port 443
- TLS otomatis Let's Encrypt
- Integrasi backend ke PostgreSQL internal compose
- Integrasi backend ke Docker Model Runner melalui `model-runner.docker.internal`

Cara menjalankan di VPS:

- Start: `docker compose --env-file .env up -d`
- Stop: `docker compose down`
- Logs: `docker compose logs -f`

## F. Kendala yang Dihadapi

Kendala utama:

- Ukuran model LLM cukup besar dan membutuhkan RAM yang memadai
- Waktu warm-up model lebih lama dibanding service biasa
- Sertifikat TLS valid butuh domain yang sudah mengarah ke VPS

Solusi:

- Menjalankan model lewat Docker Model Runner agar lifecycle model dikelola terpusat
- Menambahkan healthcheck dan dependency health agar backend hanya start saat dependency siap
- Menggunakan Caddy agar provisioning dan renew sertifikat Let's Encrypt berjalan otomatis
