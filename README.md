# Production-Ready Llama Chat App

This repository contains a full-stack chat application with:

- Frontend: Next.js App Router
- Backend: Go REST API
- Database: PostgreSQL
- Local LLM Inference: Docker Model Runner using Hugging Face model source
- Reverse Proxy: Caddy with automatic TLS (Let's Encrypt)

The application supports JWT authentication, per-user persistent chat history, creating new chats, and AI responses using `meta-llama/Llama-3.2-3B-Instruct` via Docker Model Runner.

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
│   └── Caddyfile
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

- Model family: `meta-llama/Llama-3.2-3B-Instruct`
- Runtime: Docker Model Runner (OpenAI-compatible endpoint)
- Model source: Hugging Face (`hf.co/meta-llama/Llama-3.2-3B-Instruct`)
- No model reload per request:
  - Model is loaded and managed by Docker Model Runner.
  - Backend uses a singleton LLM client (`services.GetLLMService`) that reuses the same HTTP client and base URL.

### Default startup flow (Docker Model Runner)

```bash
hf auth login
./scripts/docker-model-run.sh hf.co/meta-llama/Llama-3.2-3B-Instruct
```

Then start the application stack:

```bash
docker compose --env-file .env up -d --build
```

Or run the default one-command bootstrap:

```bash
./scripts/up-with-dmr.sh
```

Default `.env` values are already configured for this flow:

```bash
LLM_BASE_URL=http://model-runner.docker.internal/engines
LLM_MODEL_NAME=hf.co/meta-llama/Llama-3.2-3B-Instruct
```

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
./scripts/docker-model-run.sh hf.co/meta-llama/Llama-3.2-3B-Instruct
```

## 7. Run With Docker Compose

```bash
docker compose --env-file .env up -d --build
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
./scripts/docker-model-run.sh hf.co/meta-llama/Llama-3.2-3B-Instruct
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

## 10. Performance Notes

- CPU-friendly local inference via Docker Model Runner
- Connection pooling for PostgreSQL in backend
- Centralized singleton LLM client in backend
- Lightweight health checks
- Graceful shutdown in backend server
- Basic request rate limiting middleware
