# IRoom

Reusable deployment config for hosting IRoom. The app is a Go + WebSocket
binary, so it needs a container host and — because free hosts wipe their disk
— an **external Postgres database**.

| File | Host | Card required? |
|------|------|----------------|
| `railway.json` | **Railway — $0/month, free tier** | **No** |
| `render.yaml` | Render — $0/month, free tier | Usually yes |
| `.github/workflows/deploy-huggingface.yml` | HF Spaces — **$9/mo PRO required** | Yes |

Start with **Railway** if you want $0 and no credit card.

---

## 1. Railway (free, no credit card) — recommended

Railway's **Free** plan is `$0/month` and their pricing page states
**"No credit card required"**. It supports Docker, WebSockets, and
auto-deploys from GitHub.

1. Sign up at [railway.app](https://railway.app) (GitHub login).
2. **New Project → Deploy from GitHub repo** → pick iroom.
3. Railway detects the root `Dockerfile` automatically; `railway.json` pins the
   behaviour.
4. Add the variables (Dashboard → service → **Variables**):

   | Key | Value |
   |---|---|
   | `DB_DRIVER` | `postgres` |
   | `DATABASE_URL` | your Supabase URI |
   | `SERVER_HOST` | `0.0.0.0` |
   | `JWT_SECRET` | long random string (else every redeploy logs everyone out) |

5. Railway injects `PORT`; the app already reads it.

### Free-tier limits — read these

| Limit | Value | Consequence |
|---|---|---|
| RAM | **0.5 GB** | Tight but workable for the Go server |
| vCPU | 1 | Fine for a demo |
| Volume | **0.5 GB** | ⚠️ **Deleted 30 days after your trial credits expire** |
| Replicas | 1 | No scaling |
| Egress | $0.05/GB | Billed — avoid heavy video |

> **Volumes are not permanent on the free tier.** Railway's docs: *"Railway
> deletes stateful volumes created by Trial accounts 30 days after the
> expiration of your credits."* Do **not** store anything you can't lose —
> keep the database in Supabase (which persists) and treat files as scratch.

> **Verify your GitHub is connected.** Unverified accounts get a "Limited
> Trial" with **restricted outbound network and ports**, which can break the
> connection to Supabase. Connecting GitHub (which you need to deploy anyway)
> gives full network access. Check at `railway.app/verify` if unsure.

---

## 2. Render (free, but usually asks for a card)

`render.yaml` configures Render's free tier. If Render's signup demands a
card on your account, use Railway instead — the app is identical either way.

```bash
# Render Dashboard -> New -> Blueprint -> connect the repo
# Paste DATABASE_URL when prompted
```

Render's free tier **spins down after 15 minutes idle** and has **no
persistent disk**, so it is a worse fit than Railway for this app.

---

## 3. Hugging Face Spaces — no longer free

Since **July 2026**, HF moved Gradio and **Docker** Spaces behind the **PRO
plan ($9/month)** with no announcement, while the pricing page still lists
"CPU Basic — FREE". The CLI returns:

> *"Static Spaces are free for everyone, but hosting Gradio and Docker Spaces
> on free cpu-basic requires a PRO subscription."*

Pushing to an existing Docker Space also now requires PRO. The workflow and
the `sdk: docker` frontmatter in `README.md` remain in place if you pay for PRO.

---

## 4. Database (required on all free hosts)

Free container hosts wipe their disk, so the data must live outside the
container. Use **Supabase** (free).

1. Create a project, then **Project Settings → Database → Connection string → URI**.
2. Use the **Session pooler** on **port 5432**:

   ```
   postgresql://postgres.REF:PASSWORD@aws-0-us-east-1.pooler.supabase.com:5432/postgres
   ```

> ### ⚠️ Port 5432, not 6543
>
> IRoom uses prepared statements
> (`internal/adapter/repository/sqlite/settings.go`), and Supabase's
> **transaction** pooler (6543) does not support them — saving settings from
> the admin panel would fail.

3. The schema is created **automatically on first boot** from the embedded
   Postgres migration. Nothing else to run.

**Free tier:** 500 MB database, 1 GB storage, 5 GB egress/month. The project
pauses after 1 week idle and wakes on the next request (slow, not broken).

---

## 5. Verify a deployment

```bash
curl https://<your-host>/api/v1/health
```

```json
{ "success": true,
  "data": { "status": "ok", "db_driver": "postgres", "db_size": "n/a" } }
```

`db_driver: postgres` confirms the external database is live. Then log in with
`admin@iroom.local` / `iroomteppasword123` — **change that password.**

---

## 6. Files (uploads / recordings)

On any free host the local disk is disposable, so uploads and recordings are
lost while their database rows remain. To persist them, move them to
**Supabase Storage** (1 GB free) — this is the main outstanding piece of work
if you need shared files to survive.
