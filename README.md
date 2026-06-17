# آی‌روم — IRoom

Open-source online classroom platform with live video/audio, chat, whiteboard, screen sharing, polls, recordings, and a full admin panel.

---

## Quick Start

### Docker (Recommended)

```bash
# Clone and configure
git clone <repo-url> iroom && cd iroom
cp .env.example .env

# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

**Services:**
| Service | Port | Description |
|---------|------|-------------|
| Caddy | `80` | Reverse proxy (entry point) |
| App | `8080` | Go backend + SvelteKit frontend |
| Janus | `8088` | WebRTC gateway for video/audio |

**Default login:** `admin@iroom.local` / `admin123`

### Development (Manual)

#### Prerequisites

- **Go** ≥ 1.24
- **Node.js** ≥ 20
- **Janus Gateway** (for video/audio)

#### 1. Backend (Go)

```bash
cd iroom

# Build and run
go build -o server ./cmd/server
./server
# → http://localhost:8080
```

#### 2. Frontend (SvelteKit)

```bash
cd iroom/web

npm install
npm run dev
# → http://localhost:5173 (proxied to backend)
```

#### 3. Janus Gateway (WebRTC)

```bash
# Via Docker
docker compose up janus -d

# Or manual install
# See: https://janus.conf.meetecho.com/docs/setup.html
# Config: janus-config/janus.jcfg
```

---

## Configuration

### Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Backend listen address |
| `SERVER_PORT` | `8080` | Backend port |
| `DATABASE_PATH` | `iroom.db` | SQLite database file |
| `JWT_SECRET` | `change-me-...` | JWT signing secret (change in production!) |
| `JWT_ACCESS_EXPIRY` | `15` | Access token lifetime (minutes) |
| `JWT_REFRESH_EXPIRY` | `10080` | Refresh token lifetime (minutes) |
| `JANUS_HTTP_URL` | `http://localhost:8088` | Janus HTTP API |
| `JANUS_WS_URL` | `ws://localhost:8188` | Janus WebSocket |
| `UPLOAD_MAX_SIZE` | `52428800` | Max upload size (50MB) |
| `EXTERNAL_API_KEY` | `change-me-...` | External API key |

### config.yaml

Alternative configuration file (env vars override yaml):

```yaml
server:
  host: "0.0.0.0"
  port: 8080
database:
  path: "iroom.db"
jwt:
  secret: "your-secret-key"
  access_expiry: 15
  refresh_expiry: 10080
janus:
  http_url: "http://localhost:8088"
  ws_url: "ws://localhost:8188"
```

---

## Architecture

```
┌─────────┐     ┌──────────┐     ┌──────────────┐
│  Caddy   │────▶│   Go     │────▶│   SQLite     │
│  :80     │     │  :8080   │     │  iroom.db    │
└─────────┘     └────┬─────┘     └──────────────┘
                     │
                     ▼
               ┌──────────┐
               │  Janus   │
               │  :8088   │
               │  :8188   │
               └──────────┘
```

**Stack:**
- **Backend:** Go + Echo framework + SQLite
- **Frontend:** SvelteKit + Tailwind CSS
- **WebRTC:** Janus Gateway (video/audio/screen share)
- **Proxy:** Caddy (production) or Vite dev server (development)

---

## Project Structure

```
iroom/
├── cmd/server/          # Go entrypoint
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # SQLite + migrations
│   ├── handlers/        # HTTP handlers (auth, classes, sessions, etc.)
│   ├── middleware/       # Auth, CORS, rate limiting
│   ├── models/          # Data models
│   ├── pkg/             # Utilities (hash, jwt, jalali, response)
│   ├── repository/      # Database queries
│   └── services/        # Business logic (LiveKit, TOTP, WebSocket)
├── web/
│   ├── src/
│   │   ├── lib/         # Shared components, stores, API client
│   │   └── routes/      # SvelteKit pages
│   └── static/          # Static assets
├── config.yaml          # App configuration
├── docker-compose.yml   # Docker services
├── Dockerfile           # Multi-stage build
└── Caddyfile            # Reverse proxy config
```

---

## Features

- **Classroom:** Live video/audio, screen sharing, whiteboard, chat, polls
- **Rooms:** Create/manage rooms with invite codes
- **Sessions:** Schedule and manage live sessions
- **Recording:** Cloud recording support
- **Admin Panel:** User management, room management, settings, logs
- **Authentication:** JWT + optional TOTP 2FA
- **File Upload:** Secure file sharing within sessions
- **Persian Support:** Full RTL layout, Jalali calendar, Persian numbers

---

## Development

```bash
# Install dependencies
cd web && npm install

# Run dev servers (backend + frontend)
go run ./cmd/server &
cd web && npm run dev

# Type check
cd web && npm run check

# Build
cd web && npm run build

# Run tests
go test ./...
```

---

## License

MIT
