# IRoom Deployment Guide

## Architecture

```
User Browser ←── WebSocket (binary media + JSON chat) ──→ Go Server
```

- **Video/Audio**: MediaRecorder (VP8+Opus) → WebSocket binary chunks → MediaSource playback
- **Chat**: JSON messages over same WebSocket
- **Whiteboard**: Canvas strokes over same WebSocket
- **No WebRTC, no STUN, no TURN, no ICE**

## Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 512MB | 1GB |
| CPU | 1 core | 2 cores |
| Bandwidth | 50 Mbps | 100 Mbps |
| OS | Ubuntu 20.04+ / Debian 11+ / any Linux |

## Capacity (on 1 core, 1GB RAM)

| Setup | Users | Bandwidth Used |
|-------|-------|---------------|
| Teacher video + student audio | 3 classes × 20 = 60 | ~50 Mbps |
| Teacher video + student audio | 5 classes × 20 = 100 | ~83 Mbps |
| All video (250kbps each) | 20 users | ~100 Mbps |
| Audio only (32kbps) | 200+ users | ~40 Mbps |

---

## Deployment Options

### Option 1: Docker (Recommended)

```bash
# 1. Clone the repo
git clone <repo-url> iroom
cd iroom

# 2. Configure
cp .env.example .env
nano .env   # Set JWT_SECRET at minimum

# 3. Start
docker compose up -d

# 4. Open
# http://YOUR_VPS_IP
```

Docker Compose exposes:
- **80** → HTTP (Caddy reverse proxy)
- **443** → HTTPS (Caddy, auto-TLS with domain)

No other ports needed! Media travels over the same WebSocket as chat.

### Option 2: Manual (Development/Custom)

#### Prerequisites
- Go 1.22+
- Node.js 18+
- npm

#### Steps

```bash
# 1. Build frontend
cd web
npm ci
npm run build
cd ..

# 2. Build backend
go build -o server ./cmd/server

# 3. Configure
cp .env.example .env
nano .env

# 4. Run
./server
# → http://localhost:8080
```

### Option 3: systemd service (Production without Docker)

```bash
# Build
cd web && npm ci && npm run build && cd ..
go build -o /usr/local/bin/iroom ./cmd/server

# Create service
cat > /etc/systemd/system/iroom.service << 'EOF'
[Unit]
Description=IRoom Server
After=network.target

[Service]
Type=simple
User=iroom
WorkingDirectory=/opt/iroom
ExecStart=/usr/local/bin/iroom
Restart=always
RestartSec=3
EnvironmentFile=/opt/iroom/.env

[Install]
WantedBy=multi-user.target
EOF

# Setup
useradd -r -s /bin/false iroom
mkdir -p /opt/iroom
cp server config.yaml .env static/ /opt/iroom/
chown -R iroom:iroom /opt/iroom

systemctl daemon-reload
systemctl enable --now iroom
```

---

## Environment Variables

```bash
# Required
JWT_SECRET=your-random-secret-here-change-this

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DATABASE_PATH=iroom.db

# Upload
UPLOAD_MAX_SIZE=52428800        # 50MB
UPLOAD_DIR=uploads

# SMTP (optional, for password reset emails)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=you@gmail.com
SMTP_PASSWORD=app-password-here
SMTP_ENABLED=true
```

**No WebRTC config needed!** No `WEBRTC_PUBLIC_IP`, no STUN ports, no TURN servers.

---

## Firewall Rules

Only **one port** is needed:

```bash
# Ubuntu/Debian (ufw)
ufw allow 80/tcp
ufw allow 443/tcp

# CentOS/RHEL (firewalld)
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --reload
```

If not using Caddy (no reverse proxy), open port 8080 directly:
```bash
ufw allow 8080/tcp
```

---

## Domain & HTTPS (Production)

The Caddyfile is pre-configured for automatic HTTPS:

```bash
# Edit Caddyfile
nano Caddyfile
```

```caddy
your-domain.com {
    encode gzip
    handle {
        reverse_proxy app:8080
    }
}
```

Then restart:
```bash
docker compose restart caddy
```

Caddy auto-obtains a Let's Encrypt SSL certificate.

---

## Video Quality (Adaptive)

Quality automatically adjusts based on user count:

| Users | Resolution | Video Bitrate | Audio Bitrate |
|-------|-----------|---------------|---------------|
| 1–5 | 720p | 1.5 Mbps | 64 kbps |
| 6–15 | 540p | 800 kbps | 48 kbps |
| 16–30 | 360p | 400 kbps | 32 kbps |
| 31–60 | 270p | 250 kbps | 24 kbps |
| 61+ | 180p | 128 kbps | 16 kbps |

When a user joins and the count crosses a threshold, all participants' video quality adjusts automatically (resolution + bitrate). This happens transparently — no restart needed.

To customize, edit `QUALITY_TIERS` in `web/src/lib/classroom/media-client.ts`.

---

## Monitoring

```bash
# Check if running
curl http://localhost:8080/api/v1/health

# Docker logs
docker compose logs -f app

# Check WebSocket connections (if you add /metrics)
ss -s | grep estab
```

---

## Troubleshooting

### Video not showing
- Open browser console (F12)
- Check for `[Media]` log messages
- Ensure `getUserMedia` permission is granted
- Verify WebSocket is connected (check `[chat] WS connected`)

### High latency (>2 seconds)
- Check server bandwidth: `iftop` or `nethogs`
- Reduce quality tiers in `media-client.ts`
- Ensure VPS is in same region as users (Iran → Iran)

### WebSocket disconnections
- Check `nginx`/Caddy proxy timeout settings (increase to 3600s)
- Ensure WebSocket upgrade headers are passed through
- Check `docker compose logs` for errors

### Browser compatibility
- **Chrome/Edge**: Full support (MediaSource + MediaRecorder)
- **Firefox**: Partial (MediaRecorder works, MediaSource has limitations)
- **Safari**: Not supported (no MediaSource for WebM)

---

## Scaling

### Vertical (bigger server)
| VPS | Max Users (video+audio) | Cost Range |
|-----|------------------------|------------|
| 1 core / 1GB / 100Mbps | ~60 (3 classes × 20) | $5/mo |
| 2 cores / 2GB / 200Mbps | ~120 (6 classes × 20) | $10/mo |
| 4 cores / 4GB / 500Mbps | ~300 (15 classes × 20) | $20/mo |

### Horizontal (multiple servers)
For 500+ users, split classes across multiple VPS instances. Each VPS runs one `iroom` process. Users in the same class connect to the same VPS.

---

## Backup

```bash
# Database
cp iroom.db iroom.db.backup

# Uploads
tar czf uploads-backup.tar.gz uploads/

# Full backup
tar czf iroom-backup.tar.gz iroom.db uploads/ recordings/ config.yaml .env
```
