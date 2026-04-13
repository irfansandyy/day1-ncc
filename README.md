# Production-Ready Llama Chat App

This repository contains a full-stack chat application with:

- Frontend: Next.js App Router
- Backend: Go REST API
- Database: PostgreSQL
- Local LLM Inference: Docker Model Runner using Hugging Face model source
- Reverse Proxy: Caddy with automatic TLS (Let's Encrypt)

You can also run with host-level Nginx on a VPS (recommended when Nginx is already installed on the host).

The application supports JWT authentication, per-user persistent chat history, creating new chats, and AI responses using `hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K` via Docker Model Runner.

## 1. Project Structure

```text
.
├── backend/
│   ├── config/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repositories/
│   ├── services/
│   ├── Dockerfile
│   └── main.go
├── frontend/
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── Dockerfile
├── database/
│   └── init/
│       └── 001_init.sql
├── deploy/
│   ├── Caddyfile
│   └── nginx/
│       ├── day1-ncc-http-bootstrap.conf
│       └── day1-ncc.conf
├── docker-compose.yml
├── .env.example
└── REPORT_ID.md
```

## 2. Core Features

- User registration and login (`/api/auth/register`, `/api/auth/login`)
- JWT authentication middleware for protected chat routes
- Password hashing using bcrypt
- Persistent chat sessions per user in PostgreSQL
- New chat creation endpoint (`POST /api/chats`)
- Message persistence for both user and AI
- Local LLM integration with Docker Model Runner

## 3. LLM Requirements Compliance

- Model family: `meta-llama/Llama-3.2-1B-Instruct`
- Runtime: Docker Model Runner (OpenAI-compatible endpoint)
- Model source: Hugging Face (`hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K`)
- Context limiter: `LLM_CTX_SIZE=4096` enforced in backend request assembly
- No model reload per request:
  - Model is loaded and managed by Docker Model Runner.
  - Backend uses a singleton LLM client (`services.GetLLMService`) that reuses the same HTTP client and base URL.

### Default startup flow (Docker Model Runner)

```bash
export HF_TOKEN=$(cat ~/.cache/huggingface/token)
./scripts/docker-model-run.sh hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K
```
If your machine has limited RAM/VRAM, use the lower-memory quantization:

```bash
./scripts/docker-model-run.sh hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q4_K_M
```

The startup script now also auto-fallbacks from `Q6_K` to `Q4_K_M` when model initialization fails (can be disabled with `AUTO_FALLBACK_LOW_MEM=0`).
It also unloads previously running models before startup to avoid hidden memory contention (can be disabled with `UNLOAD_EXISTING_MODELS=0`).

Alternative login method:

```bash
env PATH="$HOME/.local/bin:$PATH" hf auth login
```

Then start the application stack:

```bash
docker compose --profile caddy --env-file .env up -d --build
```

Or run the default one-command bootstrap:

```bash
./scripts/up-with-dmr.sh
```

To use host Nginx (instead of Caddy container):

```bash
USE_HOST_NGINX=1 ./scripts/up-with-dmr.sh
```

This starts `db`, `backend`, and `frontend` only, with loopback host bindings:

- Frontend: `127.0.0.1:3000`
- Backend: `127.0.0.1:8080`

Use these Nginx templates:

- `deploy/nginx/day1-ncc-http-bootstrap.conf` for first certificate issuance on HTTP.
- `deploy/nginx/day1-ncc.conf` for final HTTPS (443) + redirect.

VPS host Nginx setup (Ubuntu):

```bash
sudo mkdir -p /var/www/certbot
sudo cp deploy/nginx/day1-ncc-http-bootstrap.conf /etc/nginx/sites-available/day1-ncc.conf
sudo sed -i 's/chat.example.com/YOUR_DOMAIN/g' /etc/nginx/sites-available/day1-ncc.conf
sudo ln -sf /etc/nginx/sites-available/day1-ncc.conf /etc/nginx/sites-enabled/day1-ncc.conf
sudo nginx -t
sudo systemctl reload nginx
```

Issue TLS certificate (Let's Encrypt) and reload Nginx:

```bash
sudo apt-get update && sudo apt-get install -y certbot
sudo certbot certonly --webroot -w /var/www/certbot -d YOUR_DOMAIN --email YOUR_EMAIL --agree-tos --no-eff-email
sudo cp deploy/nginx/day1-ncc.conf /etc/nginx/sites-available/day1-ncc.conf
sudo sed -i 's/chat.example.com/YOUR_DOMAIN/g' /etc/nginx/sites-available/day1-ncc.conf
sudo nginx -t
sudo systemctl reload nginx
```

After this, Nginx serves HTTPS on `443` and redirects `80 -> 443`.

Default `.env` values are already configured for this flow:

```bash
LLM_BASE_URL=http://model-runner.docker.internal/engines
LLM_MODEL_NAME=hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K
LLM_CTX_SIZE=4096
```
Important memory note:

- `LLM_CTX_SIZE` in this project limits backend prompt assembly.
- Increasing it (for example to `16384`) increases runtime KV cache pressure during inference and can cause out-of-memory on smaller machines.
- If you get `inference backend took too long to initialize`, use a smaller quantization (`Q4_K_M`) and keep context in the `2048-4096` range.

With this setup, backend calls Docker Model Runner OpenAI-compatible endpoint at:

```text
http://model-runner.docker.internal/engines/v1/chat/completions
```

## 4. API Endpoints

### Public

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /health`

### Protected (JWT Bearer)

- `GET /api/chats`
- `POST /api/chats`
- `GET /api/chats/{chatID}/messages`
- `POST /api/chats/{chatID}/messages`

## 5. Health Endpoint

### Route

- `GET /health`

### Success Response (extended)

```json
{
  "status": "ok",
  "services": {
    "database": "ok",
    "llm": "ok"
  }
}
```

### Simple Response

```text
GET /health?simple=true
```

Returns:

```json
{
  "status": "ok"
}
```

### What is checked

- Database readiness via `db.PingContext`
- LLM readiness via lightweight Docker Model Runner model check (`/v1/models`)

The endpoint always responds with HTTP 200 and reports service state details.

## 6. Environment Setup

1. Copy environment template:

```bash
cp .env.example .env
```

1. Configure reverse proxy domain and ACME email in `.env`:

```bash
DOMAIN=chat.example.com
ACME_EMAIL=admin@example.com
```

For local testing without public DNS, keep `DOMAIN=localhost`.

1. Start Docker Model Runner model:

```bash
hf auth login
./scripts/docker-model-run.sh hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K
```

## 7. Run With Docker Compose

```bash
docker compose --profile caddy --env-file .env up -d --build
```

Or host Nginx mode (no Caddy container):

```bash
docker compose --env-file .env up -d --build db backend frontend
```

Check status:

```bash
docker compose ps
```

Verify health endpoint:

```bash
curl -k https://localhost/health
```

Application URL:

```text
https://localhost
```

All public traffic is served through the reverse proxy on port 443.

## 8. Local Development (Without Docker)

### Backend

```bash
cd backend
go mod tidy
go run .
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## 9. Production Deployment to DigitalOcean VPS

1. Provision an Ubuntu droplet (recommended minimum 4 vCPU / 8 GB RAM).
1. SSH into VPS and install Docker + Compose plugin.
1. Clone repository on VPS.
1. Create `.env` from `.env.example` and set secure values: strong `JWT_SECRET`, `DOMAIN` set to your public domain (required for valid Let's Encrypt certificate), and `ACME_EMAIL` set to your email for certificate registration.
1. On VPS, authenticate Hugging Face and start the Docker Model Runner model:

```bash
hf auth login
./scripts/docker-model-run.sh hf.co/bartowski/Llama-3.2-1B-Instruct-GGUF:Q6_K
```

1. Start services:

```bash
docker compose --env-file .env up -d --build
```

1. Open firewall ports (DigitalOcean Cloud Firewall + host UFW): `22/tcp`, `80/tcp` (ACME challenge + redirect), and `443/tcp` (public HTTPS).
1. Verify external health check:

```bash
curl https://<YOUR_DOMAIN>/health
```

Routing is handled by Caddy:

- `/` to frontend service
- `/api/*` and `/health` to backend service

Alternative with host Nginx:

- `/` to `http://127.0.0.1:3000`
- `/api/*` and `/health` to `http://127.0.0.1:8080`

## 10. Performance Notes

- CPU-friendly local inference via Docker Model Runner
- Connection pooling for PostgreSQL in backend
- Centralized singleton LLM client in backend
- Lightweight health checks
- Graceful shutdown in backend server
- Basic request rate limiting middleware
