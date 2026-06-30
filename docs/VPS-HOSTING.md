# IRoom VPS Hosting Guide

This guide covers deploying IRoom to a VPS (Ubuntu/Debian) with Docker, HTTPS via Caddy, and WebRTC configuration.

---

## Prerequisites

- A VPS with Ubuntu 22.04+ or Debian 12+ (minimum 1 vCPU, 1GB RAM)
- A domain name pointing to your VPS IP
- SSH access to the VPS

---

## 1. Server Setup

SSH into your VPS and install Docker:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect
exit
```

Verify Docker is working:

```bash
docker compose version
```

---

## 2. Clone and Configure

```bash
# Clone the repo
git clone <repo-url> iroom
cd iroom

# Create environment file
cp .env.example .env
```

Edit `.env` with your production values:

```bash
nano .env
```

**Critical variables to change:**

```env
# Generate a random secret: openssl rand -hex 32
JWT_SECRET=your-random-64-char-secret-here

# Your VPS public IP (required for WebRTC to work)
WEBRTC_PUBLIC_IP=203.0.113.50

# Optional: TURN server credentials
WEBRTC_TURN_SECRET=your-turn-secret

# External API key
EXTERNAL_API_KEY=change-me-external-api-key
```

---

## 3. Configure Caddy for HTTPS

Edit the `Caddyfile` to use your domain:

```bash
nano Caddyfile
```

Replace the contents with:

```Caddyfile
your-domain.com {
    encode gzip

    handle {
        reverse_proxy app:8080
    }
}
```

Caddy will automatically provision a Let's Encrypt SSL certificate.

---

## 4. Open Required Ports

Make sure these ports are open on your VPS firewall:

```bash
# SSH
sudo ufw allow 22/tcp

# HTTP/HTTPS (Caddy)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# WebRTC UDP (Pion)
sudo ufw allow 8081/udp

# Enable firewall
sudo ufw enable
```

If your VPS provider has a separate firewall (e.g., AWS Security Groups, Hetzner Cloud Firewall), open the same ports there.

---

## 5. Build and Start

```bash
docker compose up -d --build
```

Check that both containers are running:

```bash
docker compose ps
```

You should see `app` and `caddy` containers with status `Up`.

Check logs if something is wrong:

```bash
docker compose logs -f
```

---

## 6. Access IRoom

Open `https://your-domain.com` in your browser.

**Default admin login:**
- Email: `admin@iroom.local`
- Password: `admin123`

**Change the default password immediately** after first login via the admin panel.

---

## 7. WebRTC Configuration

WebRTC requires your server's public IP to be set for peers to connect. Without it, video/audio will only work on local networks.

### Option A: Direct Pion (simple setup)

Set in `.env`:

```env
WEBRTC_PUBLIC_IP=your-vps-ip
WEBRTC_UDP_PORT=8081
```

This is sufficient for most VPS setups where UDP traffic is not blocked.

### Option B: With TURN Server (for restrictive networks)

If users are behind strict NATs or firewalls, add a TURN server. You can use Coturn on the same VPS:

```bash
sudo apt install coturn

# Enable the service
sudo systemctl enable coturn
```

Edit `/etc/turnserver.conf`:

```conf
listening-port=3478
tls-listening-port=5349
fingerprint
lt-cred-mech
user=turnuser:turnpassword
realm=your-domain.com
total-quota=100
stale-nonce=600
log-file=/var/log/turnserver.log
```

Set in `.env`:

```env
WEBRTC_TURN_SECRET=turnpassword
```

Open TURN ports:

```bash
sudo ufw allow 3478/tcp
sudo ufw allow 3478/udp
sudo ufw allow 5349/tcp
sudo ufw allow 5349/udp
```

---

## 8. Data Persistence

All persistent data is stored in Docker volumes mapped to local directories:

| Path | Purpose |
|------|---------|
| `./data/` | SQLite database (`iroom.db`) |
| `./uploads/` | User uploads (files, avatars) |
| `./recordings/` | Session recordings |

**Back up these directories regularly.** The database is a single SQLite file.

### Backup Example

```bash
# Stop the app briefly for a consistent backup
docker compose stop app

# Backup
tar czf iroom-backup-$(date +%Y%m%d).tar.gz data/ uploads/ recordings/

# Restart
docker compose start app
```

### Automated Backup with Cron

```bash
crontab -e
```

Add:

```cron
0 3 * * * cd /home/user/iroom && tar czf /backups/iroom-$(date +\%Y\%m\%d).tar.gz data/ uploads/ recordings/
```

---

## 9. Updating

```bash
cd iroom
git pull
docker compose up -d --build
```

The database and uploads persist across updates since they are in mounted volumes.

---

## 10. Reverse Proxy with Nginx (Alternative)

If you prefer Nginx over Caddy, replace the Caddy service in `docker-compose.yml` with:

```yaml
services:
  app:
    build: .
    ports:
      - "127.0.0.1:8080:8080"
    env_file:
      - .env
    volumes:
      - ./data:/app/data
      - ./uploads:/app/uploads
      - ./recordings:/app/recordings
    restart: unless-stopped
```

Then install Nginx + Certbot on the host:

```bash
sudo apt install nginx certbot python3-certbot-nginx
```

Create `/etc/nginx/sites-available/iroom`:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/iroom /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d your-domain.com
```

---

## Troubleshooting

**WebRTC video/audio not working:**
- Verify `WEBRTC_PUBLIC_IP` is set to your VPS public IP
- Check that UDP port 8081 is open
- Test with `nc -uzv your-vps-ip 8081`

**Caddy not getting SSL certificate:**
- Ensure your domain's A record points to the VPS IP
- Ports 80 and 443 must be open
- Check `docker compose logs caddy`

**App won't start:**
- Check logs: `docker compose logs app`
- Verify `.env` has valid values
- Ensure `data/`, `uploads/`, `recordings/` directories exist and are writable

**Database locked errors:**
- SQLite handles one writer at a time; this is normal under low traffic
- For high traffic, consider the write-ahead log (already enabled by default with `.db-wal`)
