# آی‌روم — IRoom
## EXPERIMENTAL (dont use it)

Open-source online classroom platform with live video/audio, chat, whiteboard, screen sharing, polls, recordings, and a full admin panel.

**WebRTC powered by Pion (Go-native)** — no external Janus Gateway needed.

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

```bash
# Backend
cd iroom && go build -o server ./cmd/server && ./server

# Frontend (new terminal)
cd iroom/web && npm install && npm run dev
# → http://localhost:5173
```

**Default login:** `admin@iroom.local` / `admin123`

---

## Architecture

```
┌─────────┐     ┌──────────┐     ┌──────────────┐
│  Caddy   │────▶│   Go     │────▶│   SQLite     │
│  :80     │     │  :8080   │     │  iroom.db    │
└─────────┘     │  + Pion  │     └──────────────┘
                └──────────┘
```

**Stack:** Go + Echo + SQLite + Pion WebRTC + SvelteKit + Tailwind CSS

---

## Configuration

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Backend port |
| `JWT_SECRET` | `change-me...` | JWT secret (change in prod!) |
| `JWT_ACCESS_EXPIRY` | `15` | Access token lifetime (min) |
| `UPLOAD_MAX_SIZE` | `52428800` | Max upload (50MB) |

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
│   └── webrtc/          # Pion WebRTC (room, signaling)
├── web/src/             # SvelteKit frontend
├── config.yaml          # App configuration
├── docker-compose.yml   # Docker services
└── Dockerfile           # Multi-stage build
```

---

## License

MIT
