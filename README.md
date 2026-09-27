---
title: IRoom
emoji: 🎓
colorFrom: indigo
colorTo: gray
sdk: docker
app_port: 7860
pinned: false
license: mit
---

# آی‌روم — IRoom
## EXPERIMENTAL (dont use it)

Open-source online classroom platform with live video/audio, chat, whiteboard, screen sharing, polls, recordings, and a full admin panel.

**WebSocket-based video/audio streaming** — no WebRTC, no STUN/TURN/ICE needed.

---

## Quick Start

### Docker (Recommended)

```bash
git clone <repo-url> iroom && cd iroom
cp .env.example .env
docker compose up -d
# → http://localhost:80
```

### Development

#### Prerequisites

- **Go 1.25+** — no C compiler needed. SQLite is provided by
  `modernc.org/sqlite` (pure Go), so the project builds with `CGO_ENABLED=0`
  and cross-compiles cleanly.
- **Node.js 18+**

#### Backend

**Option A — Manual build:**

```bash
cp .env.example .env
go build -o server ./cmd/server && ./server
# → http://localhost:8080
```

**Option B — Auto-reload with [air](https://github.com/air-verse/air) (recommended):**

```bash
# Install air (one-time)
go install github.com/air-verse/air@latest

# Run — watches .go/.sql/.yaml files and rebuilds on change
air
```

#### Frontend (new terminal)

```bash
cd web && npm install && npm run dev
# → http://localhost:5173
```

The dev server proxies `/api`, `/ws`, `/uploads` and `/recordings` to
`localhost:8080`, so run the backend and frontend side by side.

**Default login:** `admin@iroom.local` / `admin123`

#### Single-binary mode (no dev servers)

The backend also serves the built frontend, which is how it runs in Docker and
on Hugging Face Spaces. Build the frontend once, then run just the Go binary:

```bash
cd web && npm install && npm run build && cd ..
cp -r web/build static
go build -o server ./cmd/server && ./server
# → http://localhost:8080  (UI + API from one process)
```

---

## Architecture

```
┌─────────┐     ┌──────────┐     ┌──────────────────┐
│  Caddy   │────▶│   Go     │────▶│ SQLite (default) │
│  :80     │     │  :8080   │     │   or Postgres    │
└─────────┘     │  (Echo)  │     │ (Supabase/Neon)  │
                └──────────┘     └──────────────────┘
```

**Stack:** Go + Echo + SQLite/Postgres + SvelteKit + Tailwind CSS

The database layer is driver-agnostic: repositories depend on a portable
`database.DB` interface, and `DB_DRIVER` (plus `DATABASE_URL`) selects the
backend at startup. Nothing above `internal/database` needs to change to move
from SQLite to Postgres.

---

## Configuration

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Backend port |
| `PORT` | — | Listen port (takes precedence; set by HF Spaces/Render/Fly) |
| `DB_DRIVER` | `sqlite` | `sqlite` or `postgres` |
| `DATABASE_PATH` | `iroom.db` | SQLite file path |
| `DATABASE_URL` | — | Postgres DSN (implies `DB_DRIVER=postgres`) |
| `SUPABASE_DB_URL` | — | Alias for `DATABASE_URL` |
| `DB_SSLMODE` | — | e.g. `require`; only added if the URL has no `sslmode` |
| `DB_MAX_OPEN_CONNS` | `10` | Postgres pool size |
| `DB_MAX_IDLE_CONNS` | `5` | Postgres idle connections |
| `JWT_SECRET` | `change-me...` | JWT secret (change in prod!) |
| `JWT_ACCESS_EXPIRY` | `15` | Access token lifetime (min) |
| `UPLOAD_MAX_SIZE` | `52428800` | Max upload (50MB) |

### Free / ephemeral hosting

Containers on Railway, Render and similar hosts lose their local disk on every
restart, so a local `iroom.db` is not viable there. Point the app at an
external Postgres instead:

```bash
DB_DRIVER=postgres
DATABASE_URL=postgresql://postgres.REF:PASSWORD@aws-0-REGION.pooler.supabase.com:5432/postgres?sslmode=require
```

`railway.json` deploys to Railway's free tier (**$0, no credit card required**)
and the schema is created automatically on first boot. See
[DEPLOY.md](DEPLOY.md) for the full walkthrough and free-tier limits.

> **Hugging Face Spaces is no longer free** — since July 2026, Docker Spaces
> require the PRO plan ($9/month).

---

## Features

- **Classroom:** Live video/audio, screen sharing, whiteboard, chat, polls
- **Rooms:** Create/manage rooms with invite codes
- **Sessions:** Schedule and manage live sessions
- **Recording:** Cloud recording support
- **Admin Panel:** User management, room management, settings, logs
- **Auth:** JWT + optional TOTP 2FA
- **Persian:** Full RTL, Jalali calendar, Persian numbers

---

## Project Structure

```
iroom/
├── cmd/server/          # Go entrypoint
├── internal/
│   ├── handlers/        # HTTP handlers
│   ├── middleware/       # Auth, CORS, rate limiting
│   ├── models/          # Data models
│   ├── repository/      # Database queries
│   ├── services/        # Business logic
│   └── infrastructure/  # External integrations
├── web/src/             # SvelteKit frontend
├── config.yaml          # App configuration
├── docker-compose.yml   # Docker services
└── Dockerfile           # Multi-stage build
```

---

## License

MIT
