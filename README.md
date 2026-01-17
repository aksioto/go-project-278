## Link Shortener Service

[![Hexlet](https://github.com/aksioto/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/aksioto/go-project-278/actions)
[![CI](https://github.com/aksioto/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/aksioto/go-project-278/actions/workflows/ci.yml)

Production-ready HTTP API for creating short links, tracking visits, and managing redirects. Built with Gin, PostgreSQL, pgx, and Sentry instrumentation.

### Demo
[Onrender](https://go-project-278-4s76.onrender.com/)

## Features
1. Create, list, update, and delete short links
2. Redirect endpoint (`/r/:code`) that records visits (IP, UA, referer, status)
3. Pagination responses include `Content-Range start-end/total` headers
4. Validation powered by `go-playground/validator` (RFC 3986 parsing + scheme whitelist)
5. Structured logging (`log/slog`) and centralized error handling
6. Sentry integration with request context, safe header filtering, and graceful flush

## Tech Stack
- Go 1.25+
- Gin, pgx, sqlc, gomock
- PostgreSQL
- Sentry (optional)
- Docker + Makefile helpers
- Caddy

## Getting Started

### Setup
```bash
git clone https://github.com/aksioto/go-project-278.git
cd go-project-278
cp .env.sample .env          # configure DATABASE_URL, BASE_URL, etc.
go mod download
```

### Apply database migrations
1. **Install goose (one time):**
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```
2. **Create the target database** in PostgreSQL (e.g., `createdb urlshortener`).
3. **Configure migration variables** in your `.env` file, then reload the environment:
   ```bash
   MIGRATIONS_DIR=internal/db/migrations
   DATABASE_URL="postgres://user:pass@localhost:5432/urlshortener?sslmode=disable"
   ```
4. **Run migrations up:**
   ```bash
   make migrate-up
   ```
5. Optional helpers:
   ```bash
   make migrate-status   # show applied migrations
   make migrate-down     # rollback latest migration
   make migrate-new name=add_links_table  # scaffold new migration
   ```

### Run the service
```bash
go run .
```

### Run backend + Hexlet frontend (dev)
```bash
npm install
npm run dev   # spins up Go API and @hexlet/project-url-shortener-frontend via concurrently
```
Backend: `localhost:8080`, Frontend: `localhost:5173`

### Run frontend only
```bash
npm run frontend  # launches @hexlet/project-url-shortener-frontend (preview mode) on port 5173
```

### Run tests / lint (optional)
```bash
make test
make lint
```

## Configuration
Minimum required environment variables (see `.env.sample` for the full list and descriptions):

| Variable | Description |
| --- | --- |
| `ENV` | Deployment environment (dev/prod) |
| `SERVICE_NAME` | Service identifier used for logs/Sentry |
| `APP_VERSION` | Semantic version exposed in logs/Sentry |
| `APP_PORT` | HTTP server port for the Go application |
| `BASE_URL` | Base URL used to build absolute short links |
| `DATABASE_URL` | PostgreSQL DSN for the application |

## Observability
- `middleware.ErrorsMiddleware` maps known domain errors to 4xx and reports only 5xx to Sentry.
- `middleware.SentryMiddleware` attaches request context; when disabled it becomes a no-op.
- `internal/infra/sentry` filters sensitive headers before sending events.

## Frontend & Caddy
- The repository already includes `@hexlet/project-url-shortener-frontend`; run `npm run dev` (see above) to start the SPA on port `5173` together with the API.
- A sample `Caddyfile` is provided to serve the built frontend from `/app/public` and reverse-proxy API calls to the Go service (`APP_PORT`). Keep `ALLOWED_ORIGINS` in sync with your frontend origin.

## Deployment (Render)
- The provided `Dockerfile` and `bin/run.sh`: run goose migrations, launch Caddy (serving static UI) and the Go binary.
- On Render, set `PORT=80`, `APP_PORT=8080`, `DATABASE_URL`, `SENTRY_DSN` (optional), and any other variables from `.env.sample`. The public URL is handled by Render; use `BASE_URL` accordingly.

## API Overview
| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/links` | `POST` | Create link |
| `/api/links` | `GET` | List links (supports `range` query) |
| `/api/links/:id` | `GET/PUT/DELETE` | Get, update, delete link |
| `/api/link_visits` | `GET` | List visits with pagination |
| `/api/link_visits/:id` | `DELETE` | Delete visit by ID |
| `/r/:code` | `GET` | Redirect to original URL |
| `/ping` | `GET` | Health check |

Each `GET` with pagination returns `Content-Range` header (`resource start-end/total`) to comply with front-end data grids.
