Spy Cat Agency (SCA) — REST API in Go

Overview
- CRUD for spy cats, missions, and targets.
- Storage: PostgreSQL 15+ via GORM; migrations are raw SQL in `migrations`.
- External service TheCatAPI for breed list (in‑memory caching).
- Swagger documentation and lightweight middleware/logging for development.

Quick Start (Docker)
- Command: `docker compose up --build` or `make up`
- App: `http://localhost:888`
- Health: `GET http://localhost:888/healthz` → `{ "ok": true }`
- Swagger: `http://localhost:888/swagger/index.html`

Local Run (without Docker)
- Requirements: `Go 1.23+`, running PostgreSQL 15+
- DB environment variables are optional (defaults are used).
- Apply migrations: `go run sca/cmd/sca/main.go --migrate-only`
- Start API: `APP_ENV=dev go run sca/cmd/sca/main.go`

Environment Variables
- `POSTGRES_HOST`: DB host (default `localhost`)
- `POSTGRES_PORT`: DB port (default `5432`)
- `POSTGRES_USER`: DB user (default `sca`)
- `POSTGRES_PASSWORD`: DB password (default `sca`)
- `POSTGRES_DB`: DB name (default `sca`)
- `APP_ENV`: mode (`dev` enables detailed logs)
- `THECATAPI_KEY`: optional API key for https://thecatapi.com (raises limits)

Base URL and Health
- All REST endpoints are under: `/api/v1`
- Health check: `GET /healthz` returns `{ "ok": true }`

Endpoints (summary)
- Cats:
  - `POST /api/v1/cats` — create (name, years_of_experience, breed, salary_cents)
  - `GET /api/v1/cats` — list
  - `GET /api/v1/cats/{id}` — get by ID
  - `PUT /api/v1/cats/{id}` — update salary (`salary_cents`)
  - `GET /api/v1/breeds` — list breeds from TheCatAPI
- Missions and targets:
  - `POST /api/v1/missions` — create mission with targets (1–3, names unique within a mission)
  - `GET /api/v1/missions` — list (with targets)
  - `GET /api/v1/missions/{id}` — get (with targets)
  - `PATCH /api/v1/missions/{id}` — mark mission completed
  - `DELETE /api/v1/missions/{id}` — delete (forbidden if a cat is assigned)
  - `POST /api/v1/missions/{id}/assign_cat` — assign a cat (a cat may have only one active mission)
  - `POST /api/v1/missions/{id}/targets` — add new targets (up to 3 total, names unique per mission)
  - `PATCH /api/v1/missions/{id}/targets/{tid}` — update a target (notes/status; notes cannot be changed after completion)
  - `DELETE /api/v1/missions/{id}/targets/{tid}` — delete a target (cannot delete completed targets)

Business Rules and Invariants
- At most one active (non‑completed) mission per cat: enforced by a DB unique index.
- Max 3 targets per mission; target names are unique within the mission.
- Completed target’s notes are frozen (no edits allowed).
- A mission can be explicitly completed (`PATCH /missions/{id}`); deletion is forbidden if a cat is assigned.

Examples (curl)
```
# health
curl -s http://localhost:888/healthz

# create a cat
curl -s -X POST http://localhost:888/api/v1/cats \
  -H 'content-type: application/json' \
  -d '{
    "name":"Whiskers",
    "years_of_experience":2,
    "breed":"siamese",
    "salary_cents":500000
  }'

# create a mission with two targets
curl -s -X POST http://localhost:888/api/v1/missions \
  -H 'content-type: application/json' \
  -d '{
    "targets":[
      {"name":"Mouse A","country":"US","notes":""},
      {"name":"Mouse B","country":"UK","notes":""}
    ]
  }'

# assign a cat to mission 1
curl -s -X POST http://localhost:888/api/v1/missions/1/assign_cat \
  -H 'content-type: application/json' \
  -d '{"cat_id":1}'

# mark target 1 as completed
curl -s -X PATCH http://localhost:888/api/v1/missions/1/targets/1 \
  -H 'content-type: application/json' \
  -d '{"completed":true}'
```

Project Structure
- `sca/cmd/sca`: application entrypoint
- `sca/internal/server`: routing, middleware, swagger
- `sca/internal/handlers`: HTTP handlers (cats, missions, targets, breeds)
- `sca/internal/models`: data models (GORM)
- `sca/internal/storage`: DB initialization and migrations
- `sca/internal/clients/thecatapi`: TheCatAPI client (HTTP + mock)
- `migrations`: PostgreSQL SQL migrations
- `docs`: Swagger (generated via `swag init`)

Development
- Generate Swagger locally: `swag init -g sca/cmd/sca/main.go -o docs`
- Apply migrations and run API:
  - `go run sca/cmd/sca/main.go --migrate-only`
  - `APP_ENV=dev go run sca/cmd/sca/main.go`
