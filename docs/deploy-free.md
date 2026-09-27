# Deploying IRoom for Free

IRoom is a Go + WebSocket app that needs a container host and an **external
database** (free hosts wipe their local disk on every restart).

```
┌──────────────┐  push   ┌────────────────────┐   HTTPS   ┌──────────────────┐
│   GitHub     │ ──────▶ │ Render (free)      │ ────────▶ │ Supabase Postgres│
│  (source)    │         │  auto-deploys      │           │  (free, external)│
└──────────────┘         └────────────────────┘           └──────────────────┘
```

| Component | Service | Cost |
|---|---|---|
| App hosting | [Render](https://render.com) — free web service | **$0** |
| Database | [Supabase](https://supabase.com) — free Postgres | **$0** |
| Source + CI | [GitHub](https://github.com) | **$0** |

> ### ⚠️ Hugging Face Spaces is **not** free anymore
>
> As of **July 2026**, HF moved Gradio and **Docker** Spaces behind the **PRO
> plan ($9/month)**. The CLI returns:
> *"Static Spaces are free for everyone, but hosting Gradio and Docker Spaces
> on free cpu-basic requires a PRO subscription."*
>
> There was **no announcement**, and HF's pricing page still lists
> "CPU Basic — FREE", so this is easy to miss. Users also report that
> **pushing to an existing Docker Space now also demands PRO**.
>
> If you already have HF PRO, `.github/workflows/deploy-huggingface.yml` and
> the `sdk: docker` frontmatter in `README.md` still work unchanged.

---

## 1. What you need (all free, no credit card)

| # | Account | Purpose |
|---|---------|---------|
| 1 | [Supabase](https://supabase.com) | External Postgres database |
| 2 | [Render](https://render.com) | Hosts the app |

---

## 2. Create the Supabase database

1. Create a project (pick a region near your users) and wait ~2 minutes.
2. **Project Settings → Database → Connection string → URI**.
3. Copy the **Session pooler** URI:

   ```
   postgresql://postgres.PROJECTREF:PASSWORD@aws-0-us-east-1.pooler.supabase.com:5432/postgres
   ```

> ### ⚠️ Use port 5432 (Session pooler), NOT 6543
>
> IRoom uses prepared statements
> (`internal/adapter/repository/sqlite/settings.go`), and Supabase's
> **transaction** pooler on port `6543` does not support them — saving settings
> from the admin panel would fail. The **session** pooler on port `5432` is
> fully compatible.
>
> Add `?sslmode=require` if it isn't already there.

4. Nothing else — **the schema is created automatically on first boot** from the
   embedded Postgres migration.

**Free tier:** 500 MB database, 1 GB file storage, 5 GB egress/month. The
project pauses after 1 week of inactivity and wakes on the next request (that
first request is slow, not broken).

> **Want a database that never pauses?** [Neon](https://neon.com) free gives
> 0.5 GB, suspends after 5 idle minutes and wakes instantly — same
> `postgresql://` URL, no code changes.

---

## 3. Deploy to Render

1. Commit `render.yaml` to the repository root.
2. In the Render dashboard: **New → Blueprint** → connect this GitHub repo.
3. Render detects `render.yaml` and asks for the one secret it can't generate:
   **`DATABASE_URL`** (paste the Supabase URI).
4. Click **Apply**. The first build takes ~5 minutes (Node build → Go build →
   image). Subsequent deploys take ~2 minutes.

Render **auto-deploys on every push** to the connected branch, so no GitHub
Actions workflow is needed for this path.

### Verify

```bash
curl https://<your-service>.onrender.com/api/v1/health
```

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "db_driver": "postgres",
    "db_size": "n/a",
    "uptime": "0h 1m"
  }
}
```

`db_driver: postgres` confirms the external database is live. Then open the
site and log in with `admin@iroom.local` / `iroomteppasword123` — **change that
password immediately.**

### What `render.yaml` sets for you

| Setting | Why |
|---|---|
| `DB_DRIVER=postgres` | Free hosts have an ephemeral disk |
| `SERVER_HOST=0.0.0.0` | Required to be reachable behind the proxy |
| `JWT_SECRET` (generated) | Stays stable across redeploys, so logins survive |
| `DB_MAX_OPEN_CONNS=5` | Free instances overlap during deploys; keeps the DB connection count low |
| `healthCheckPath` | Lets Render detect readiness |

Render injects `PORT` automatically and the app already reads it.

---

## 4. Free-tier limits (read this before relying on it)

| Thing | Free-tier behaviour | Impact |
|-------|--------------------|--------|
| **App disk** | **Local files lost on every deploy/restart** | Database is external ✅, but **uploads and recordings are lost** |
| **Spin-down** | Service sleeps after 15 min idle | First visitor waits ~30–60 s while it wakes; **open WebSocket sessions drop** |
| **Bandwidth** | 5 GB/month included | Audio-only ≈ 40 kbps/user. **Video will exhaust this quickly** |
| **Supabase size** | 500 MB DB, 5 GB egress/mo | Fine for a classroom demo |
| **Supabase idle** | Pauses after 1 week | First request is slow, not broken |
| **Runtime** | Single instance, 512 MB | Fine for a demo, not a production classroom |

### Honest expectations

Render's own docs say free instances are **not for production** and that
Render may restart them at any time. This setup is best for a **demo,
prototype, or small classroom**. For real production traffic you need paid
compute, and for sustained video you need far more than 5 GB/month of egress.

### About uploads and recordings

The database persists, but **files on disk do not**. Shared files and session
recordings are wiped on redeploy, while their database rows remain (pointing at
missing files). Options:

1. Accept it for a demo.
2. Move files to **Supabase Storage** (1 GB free) — a small change to the
   file/recording handlers. Ask and I'll implement it.
3. Use a host with a persistent disk.

Live audio/video itself works — it streams over WebSocket, not the filesystem.

---

## 5. Troubleshooting

**`failed to init database: DATABASE_URL is required`**
`DATABASE_URL` wasn't entered in the Render dashboard
(**Environment** tab of the service).

**`FATAL: too many connections`**
Lower `DB_MAX_OPEN_CONNS` to `3`. A redeploy briefly overlaps with the running
instance.

**`prepared statement … does not exist`, or settings won't save**
You're on Supabase's transaction pooler (port 6543). Switch to the session
pooler on port 5432 — see [§2](#2-create-the-supabase-database).

**Everyone logged out after a deploy**
`JWT_SECRET` must be a stable generated value (as `render.yaml` does). If you
set it inline per-deploy it changes each time.

**Build fails at the Node step**
`npm ci` requires `web/package-lock.json`. Make sure `.dockerignore` does not
exclude it.

**`The service is live, but all API calls 404`**
Check the app is binding `$PORT`. `render.yaml` sets `SERVER_HOST=0.0.0.0` and
the app reads `PORT`, which Render injects automatically.

---

## 6. Ongoing workflow

```bash
git add -A && git commit -m "..." && git push origin main
# → Render auto-detects the push → rebuild → live in ~2 min
```

No workflow file to maintain. Manual deploy: **Render Dashboard → your service
→ Manual Deploy**.

---

## 7. Optional: Hugging Face Spaces (requires PRO, $9/mo)

If you have HF PRO, the included workflow still works:

| Component | Location |
|---|---|
| Push workflow | `.github/workflows/deploy-huggingface.yml` |
| Space config | `sdk: docker` / `app_port: 7860` frontmatter in `README.md` |
| GitHub secret | `HF_TOKEN` (HF **write** token) |
| GitHub variable | `HF_SPACE` = `username/space-name` |
| Space secrets | `DB_DRIVER`, `DATABASE_URL`, `JWT_SECRET`, `EXTERNAL_API_KEY` |

HF Spaces are otherwise similar to Render for this app: ephemeral disk,
external database required, single port, `PORT` respected.

---

## 8. Total cost

| Path | Cost |
|---|---|
| **Render + Supabase** | **$0 / month** |
| Hugging Face PRO + Supabase | $9 / month |

Nothing else should ever appear on a bill. If it does, check the instance
plan — it must be the **free** tier.

---

## 9. How the database layer works

```
                    ┌──────────────────────────────┐
   application ───▶│  internal/database           │
   (repositories,   │                              │
    handlers,       │  DB interface (portable)     │
    middleware)     │    • Exec / Query / QueryRow │
                    │    • Rebind  (?  →  $1)      │
                    │    • IsUniqueViolation       │
                    │                              │
                    │  ├── sqlite.go  → SQLite     │
                    │  └── postgres.go → Postgres  │
                    └──────────────┬───────────────┘
                                   │
              repositories use "?" placeholders and portable SQL;
              the driver rewrites them per backend
```

| Concern | How it stays portable |
|---------|----------------------|
| Placeholders | `?` everywhere; `Rebind()` converts to `$1,$2` for Postgres |
| Upserts | `INSERT ... ON CONFLICT (...) DO UPDATE` (supported by both) |
| Unique errors | `IsUniqueViolation()` maps SQLite messages / PG SQLSTATE `23505` |
| Schema | Separate embedded migration sets per driver |
| Pooling | SQLite → 1 connection; Postgres → configurable pool |

`SUPABASE_DB_URL` is accepted as an alias for `DATABASE_URL`, and
`DB_SSLMODE=require` can be used instead of baking `sslmode` into the URL.

---

## 10. Other platforms

The same `DB_DRIVER=postgres` + `DATABASE_URL` setup works on any host that
runs a Docker container and speaks WebSockets:

| Platform | Notes |
|----------|-------|
| **Render** | Free tier, see above; auto-deploys from GitHub |
| Fly.io | Free allowance; an external DB is still required |
| Railway | Free trial credit |
| Koyeb | Free tier exists, but **WebSockets are a paid feature** — not suitable |
| VPS (docker compose) | Keep SQLite — the disk persists, so an external DB is optional |

---

## 11. Adding a new database driver

1. Create `internal/database/<driver>.go` with an `open<Driver>` function that
   registers the driver, configures the pool, and calls
   `Migrate(conn, <fs>, "<dir>")`.
2. Add a case to the `switch` in `database.Open`.
3. Add SQL translation to `Rebind` / `isUniqueViolation` in `db.go` if the
   backend's placeholder or error style differs.
4. Add the migration directory (e.g. `migrations/<driver>/`).

Repositories, handlers and use cases need **no changes** — they only depend on
the `database.DB` interface.

