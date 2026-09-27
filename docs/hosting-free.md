# Deploying IRoom on Free / Ephemeral Hosting

Platforms like **Hugging Face Spaces**, Render, Fly.io and Railway give you a
free container, but its **filesystem is wiped on every restart and redeploy**.
A local `iroom.db` therefore loses all data.

The solution used here: keep the app stateless and move the database to an
external Postgres (Supabase, Neon, ...). IRoom supports both drivers behind a
single abstraction, so switching is a config change — no code changes.

---

## 1. How the database layer works

```
                    ┌──────────────────────────────┐
   application ───▶│  internal/database           │
   (repositories,   │                              │
    handlers,       │  DB interface (portable)     │
    middleware)     │    • Exec / Query / QueryRow │
                    │    • Rebind  (?  →  $1)      │
                    │    • IsUniqueViolation       │
                    │    • DialectSQL upserts      │
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

### Switching drivers

```bash
# SQLite (default — local dev, VPS with a disk)
DB_DRIVER=sqlite
DATABASE_PATH=iroom.db

# Postgres / Supabase — ephemeral container hosting
DB_DRIVER=postgres          # optional: implied by DATABASE_URL
DATABASE_URL=postgresql://postgres.REF:PASSWORD@aws-0-REGION.pooler.supabase.com:5432/postgres?sslmode=require
```

`SUPABASE_DB_URL` is accepted as an alias, and `DB_SSLMODE=require` can be used
instead of baking `sslmode` into the URL.

---

## 2. Supabase setup

1. Create a project at [supabase.com](https://supabase.com) (free tier).
2. **Project Settings → Database → Connection string → URI**.
3. Copy the **Session pooler** URI (port `5432`). The transaction pooler
   (port `6543`) is fine for low traffic but does not support prepared
   statements.
4. Add `?sslmode=require` if it isn't already present.
5. The schema is created automatically on first boot — the Postgres migration
   baseline is embedded in the binary.

> **Note on the `postgres` role:** the baseline schema is created in the
> `public` schema owned by `postgres`. For stricter setups, create a dedicated
> role and set `search_path` in the connection string.

---

## 3. Hugging Face Spaces (Docker SDK)

### Repository setup

Add the Spaces metadata to the top of `README.md`:

```yaml
---
title: IRoom
emoji: 🎓
colorFrom: indigo
colorTo: gray
sdk: docker
app_port: 7860
pinned: false
---
```

### Space secrets

Under **Settings → Variables and secrets → New secret**:

| Variable | Value |
|----------|-------|
| `DB_DRIVER` | `postgres` |
| `DATABASE_URL` | *(your Supabase URI — mark as a secret)* |
| `JWT_SECRET` | *(long random string — mark as a secret)* |
| `EXTERNAL_API_KEY` | *(random string — mark as a secret)* |
| `RATE_LIMIT_DISABLED` | `false` |

The Dockerfile already binds `0.0.0.0`, reads `$PORT` and serves the built
frontend from the same process, so no extra wiring is needed. The build is
`CGO_ENABLED=0`, so there is no C toolchain layer — HF builds stay fast and the
image stays small.

### Why an external database is required

| Data | On a free Space |
|------|-----------------|
| Users, rooms, sessions, messages, settings | Must live in Postgres — the local disk is wiped on restart |
| Uploads / recordings | Ephemeral; lost on restart (use Supabase Storage if you need them) |
| JWT secret | Must be a Space secret, or every restart invalidates all sessions |

---

## 4. Verifying the deployment

```bash
curl https://<your-space>.hf.space/api/v1/health
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

`db_driver` confirms which backend is live. Restart the Space and check again —
if the data is still there, persistence is working.

---

## 5. Other platforms

The same `DB_DRIVER=postgres` + `DATABASE_URL` setup works on:

| Platform | Notes |
|----------|-------|
| Render | Free web service sleeps after 15 min; use the Postgres add-on or Supabase |
| Fly.io | Free allowance includes a volume, but an external DB is still safer |
| Railway | Free Postgres trial available |
| VPS (docker compose) | Keep SQLite — the disk persists, so the external DB is optional |

---

## 6. Adding a new database driver

1. Create `internal/database/<driver>.go` with an `open<Driver>` function that
   registers the driver, configures the pool, and calls
   `Migrate(conn, <fs>, "<dir>")`.
2. Add a case to the `switch` in `database.Open`.
3. Add SQL translation to `Rebind` / `isUniqueViolation` in `db.go` if the
   backend's placeholder or error style differs.
4. Add the migration directory (e.g. `migrations/<driver>/`).

Repositories, handlers and use cases need **no changes** — they only depend on
the `database.DB` interface.
