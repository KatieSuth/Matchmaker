# Production infrastructure (GCP + Cloudflare)

This runbook deploys Matchmaker to Google Cloud with a Cloudflare Free edge.
**You create all SaaS accounts yourself** — nothing in this repo signs up for GCP,
Cloudflare, Sentry, or Discord.

For **what** Terraform sets up, **why**, day-2 operations, and app integration contracts
(env vars, URLs, IAM, failure modes), see [INFRASTRUCTURE.md](INFRASTRUCTURE.md).

Local development stays on Docker Compose (Caddy + Postgres). See the root
[README](../README.md).

Typical idle cost is roughly **$10–15/month**, dominated by Cloud SQL.

---

## Architecture

| Piece    | Choice                                                            |
|----------|-------------------------------------------------------------------|
| Compute  | Cloud Run (API + frontend), scale-to-zero, max 2 instances        |
| Database | Cloud SQL Postgres (`db-f1-micro`), **private IP only**           |
| Edge     | Cloudflare Worker: `/api/*` → API (strip prefix), `/*` → frontend |
| Images   | Artifact Registry; build/push **manually**                        |
| Secrets  | Secret Manager → Cloud Run                                        |
| Alerts   | Billing budgets ($15 / $30 / $50 email) + uptime on `/api/health` |

Origin protection: Worker injects `X-Origin-Verify`; API and frontend reject
mismatches when `ORIGIN_VERIFY_SECRET` is set. Use `https://{domain}` as the public
entrypoint — not raw `*.run.app` origins (those are for the Worker/operators; this repo
is public, so do not commit real origin URLs or secrets).

---

## Phase 0 — Accounts (manual)

Do these in the browser / vendor consoles. Terraform starts only after you have IDs and tokens.

### 1. Google Cloud

1. Sign up at [cloud.google.com](https://cloud.google.com) and enable billing.
2. Create a project (e.g. `matchmaker-prod`); note `PROJECT_ID`.
3. Attach a billing account; note `BILLING_ACCOUNT_ID` (`XXXXXX-XXXXXX-XXXXXX`).
4. Install [Cloud SDK](https://cloud.google.com/sdk/docs/install), then:

```bash
gcloud auth login
gcloud auth application-default login
gcloud config set project PROJECT_ID
```

Terraform’s Google providers set `user_project_override` + `billing_project` to your `project_id` so APIs that require a quota project (notably Billing Budgets) work with user ADC. You still need Billing Account permissions to create budgets (e.g. Billing Account Administrator or Costs Manager on the billing account).

### 2. Domain + Cloudflare

1. Confirm you own `matchmaker.games` (or your domain) at the registrar (wherever you bought/registered the domain).
2. Create a [Cloudflare](https://dash.cloudflare.com) Free account if you don’t have one.
3. **Add the zone** — in Cloudflare this means: onboard your domain so Cloudflare can manage its DNS.
   - Dashboard → **Add a domain** (or **Add site**) → enter `matchmaker.games`.
   - Choose the **Free** plan.
   - Cloudflare scans existing DNS and shows the **nameservers** it wants you to use (two hostnames like `*.ns.cloudflare.com`).
   - A “zone” is Cloudflare’s name for that domain’s DNS + proxy settings. One domain ≈ one zone.
4. At your **registrar**, replace the domain’s nameservers with the ones Cloudflare showed. Leave them until Cloudflare marks the zone **Active** (can take from minutes to a day).
5. Create a **custom** Cloudflare API token (do not use only the “Edit Cloudflare Workers” template — Terraform also creates a DNS record).
   - **My Profile → API Tokens → Create Token → Create Custom Token**
   - **Permissions:**
     - **Account** → **Workers Scripts** → **Edit**
     - **Zone** → **Workers Routes** → **Edit** (for the `matchmaker.games/*` route)
     - **Zone** → **DNS** → **Edit** (for the apex record)
   - **Zone Resources** → Include → Specific zone → `matchmaker.games` (or your domain)
   - **Account Resources** → Include → the account that owns that zone
   - Create the token, copy it once, and store it for `cloudflare_api_token` in `terraform.tfvars`
   - Prefer this scoped token over a global API key. If “Workers Routes” is missing in the UI, Account Workers Scripts Edit + Zone DNS Edit is the baseline; add Workers Routes if `terraform apply` fails on the route resource.
6. Note **Zone ID** and **Account ID** (Overview sidebar for the zone / account).

Avoid creating conflicting manual A/CNAME records for the apex before Terraform;
Terraform manages the proxied apex record and Worker route.

### 3. Sentry (optional)

Create an org and projects for API + frontend; copy DSNs (or leave empty in tfvars).

### 4. Discord

Create an application; copy Client ID / Secret. Add the redirect URI **after** HTTPS works:

`https://matchmaker.games/api/auth/discord_redirect`

### 5. Tools

You need **Terraform (>= 1.6)**, **Docker**, **make**, and the **Cloud SDK** (`gcloud`, already covered above). Pick an email for budget + uptime alerts.

#### Terraform on Ubuntu (including WSL2)

Install from HashiCorp’s official apt repo ([docs](https://developer.hashicorp.com/terraform/install)):

#### Other tools (Ubuntu)

```bash
sudo apt install -y make
# Docker: install Docker Engine or use Docker Desktop with WSL integration
# https://docs.docker.com/engine/install/ubuntu/
```

---

## Phase 1 — Terraform

### State bucket (operator IAM only)

Terraform stores its state (what it thinks exists in GCP) as objects in a **GCS bucket**.
That is separate from Matchmaker app storage: Cloud Run never reads this bucket.

**Versioned** means GCS keeps prior object generations when state is overwritten, so you
can recover from a bad write. **Operator-only IAM** means only the identity that runs
`terraform` (you) can read/write state — not the app service accounts, and not `allUsers`.

#### 1. Create the bucket

Use the same region you plan for the rest of the stack (e.g. `us-central1`). Bucket names
are globally unique; `${PROJECT_ID}-tf-state` is usually fine.

```bash
PROJECT_ID=your-gcp-project-id   # e.g. matchmaker-prod-123456
BUCKET="${PROJECT_ID}-tf-state"

gcloud storage buckets create "gs://${BUCKET}" \
  --project="${PROJECT_ID}" \
  --location=us-central1 \
  --uniform-bucket-level-access

# Turn on object versioning (state history / recovery)
gcloud storage buckets update "gs://${BUCKET}" --versioning
```

Confirm:

```bash
gcloud storage buckets describe "gs://${BUCKET}" --format="yaml(name,versioning_enabled,location)"
```

You should see `versioning_enabled: true`.

#### 2. Grant only your user on the bucket

Terraform needs permission to create/update/delete **objects** in this bucket (the state
files). App Cloud Run service accounts must **not** get that permission.

**Check whether you already have project Owner** (common if you created the project):

```bash
gcloud projects get-iam-policy "${PROJECT_ID}" \
  --flatten="bindings[].members" \
  --filter="bindings.role:roles/owner AND bindings.members:user:$(gcloud auth list --filter=status:ACTIVE --format='value(account)')" \
  --format="table(bindings.role, bindings.members)"
```

If you see a row with `roles/owner` and your `user:…@…`, you can already read/write the
bucket via that project role. You *could* skip the next command and still run Terraform.
**Still do the bucket binding below** so state access is an explicit, bucket-scoped grant
to you alone (clearer later, and not confused with whatever roles app SAs get on the project).

**Bind bucket-level object admin to your user** (recommended even if you are Owner):

```bash
# Email of the account you used for `gcloud auth login`
gcloud auth list --filter=status:ACTIVE --format="value(account)"

YOUR_EMAIL=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")

# Grant: create/read/update/delete objects in THIS bucket only
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member="user:${YOUR_EMAIL}" \
  --role="roles/storage.objectAdmin"
```

What that does:

| Piece | Meaning |
|-------|---------|
| `gs://${BUCKET}` | Only this state bucket |
| `member=user:YOU@…` | Your Google user (not a service account, not the public) |
| `roles/storage.objectAdmin` | Full control of objects in the bucket (needed for Terraform state) |

Optional equivalent in the console: **Cloud Storage → Buckets → your `…-tf-state` bucket →
Permissions → Grant access** → principal = your email → role = **Storage Object Admin** → Save.

`user:` is for a personal Google account. (A dedicated Terraform SA would be
`serviceAccount:terraform-runner@PROJECT_ID.iam.gserviceaccount.com` — optional; not
required for this runbook.)

**Verify the binding:**

```bash
gcloud storage buckets get-iam-policy "gs://${BUCKET}" \
  --format="table(bindings.role, bindings.members)"
```

You should see `roles/storage.objectAdmin` with `user:YOUR_EMAIL`. Do **not** add
`allUsers` or `allAuthenticatedUsers`. Other rows (e.g. project Owner inherited) can
appear on a solo project; that is OK.

Also ensure your identity can enable APIs and create Run/SQL/IAM resources on the project
(e.g. **Owner**, or an equivalent custom role set). Bucket access alone is not enough to
`terraform apply`.

### Configure and apply

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars   # edit with your values
cp backend.hcl.example backend.hcl             # then edit — see below
```

**`backend.hcl`** tells Terraform where to store remote state (the GCS bucket from the
previous section). Uncomment and set the values — use the **same** bucket name you
created (`${PROJECT_ID}-tf-state`):

```hcl
bucket = "your-gcp-project-id-tf-state"
prefix = "matchmaker/prod"
```

`prefix` is a folder-like path inside the bucket (keeps state objects namespaced). Leave
it as `matchmaker/prod` unless you have a reason to change it. `backend.hcl` is
gitignored; do not commit it.

**`terraform.tfvars`** — put secrets and project settings here (also gitignored):

```bash
# db_password, jwt_secret, cookie_hash_key, cookie_encrypt_key, origin_verify_secret,
# discord_client_secret, cloudflare_*, etc.
# Hex keys: make gen-keys  (or openssl rand -hex 32 each)
```

Then init (reads `backend.hcl`), plan, and apply:

```bash
make infra-init      # from repo root → terraform init -backend-config=backend.hcl
make infra-plan
make infra-apply     # starts Cloud SQL billing immediately
```

First apply may create Cloud Run services on a **placeholder** image. Probes can
fail until you push real images (Phase 2). Re-running `apply` is safe: Terraform
`ignore_changes` on container images so manual digests are not overwritten.

```bash
cd infra/terraform && terraform output
```

Save Artifact Registry URL, Cloud Run service names, and SQL connection name.

---

## Phase 2 — Images and go-live

```bash
REGION=us-central1
PROJECT_ID=your-gcp-project-id
AR="${REGION}-docker.pkg.dev/${PROJECT_ID}/matchmaker-docker"

gcloud auth configure-docker "${REGION}-docker.pkg.dev"

# API
docker build -t "${AR}/api:$(git rev-parse --short HEAD)" \
  -f backend/Dockerfile backend
docker push "${AR}/api:$(git rev-parse --short HEAD)"

# Frontend — bake public URL for your domain
docker build -t "${AR}/frontend:$(git rev-parse --short HEAD)" \
  -f frontend/Dockerfile frontend \
  --build-arg NEXT_PUBLIC_API_URL=https://matchmaker.games/api \
  --build-arg NEXT_PUBLIC_FRONTEND_DOMAIN=matchmaker.games \
  --build-arg NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT=604800 \
  --build-arg NEXT_PUBLIC_FEEDBACK_URL=https://github.com/KatieSuth/Matchmaker/issues
docker push "${AR}/frontend:$(git rev-parse --short HEAD)"

gcloud run jobs update matchmaker-migrate \
  --region="${REGION}" \
  --image="${AR}/api:$(git rev-parse --short HEAD)"

gcloud run jobs execute matchmaker-migrate \
  --region="${REGION}" --wait

gcloud run services update matchmaker-api \
  --region="${REGION}" \
  --image="${AR}/api:$(git rev-parse --short HEAD)"

gcloud run services update matchmaker-frontend \
  --region="${REGION}" \
  --image="${AR}/frontend:$(git rev-parse --short HEAD)"
```

Or use `make gcp-push` / `make gcp-deploy` after exporting `GCP_PROJECT`, `GCP_REGION`. Deploy order is push → migrate Job → update services (API does **not** auto-migrate in production).

### Checks

1. Cloudflare SSL/TLS preferably **Full (strict)** once origins serve HTTPS (`*.run.app` does).
2. Discord redirect URI as above.
3. Smoke:

```bash
curl -sS https://matchmaker.games/api/health
curl -sS https://matchmaker.games/health
```

4. Discord login; confirm Monitoring uptime; optional Sentry test event.

---

## Phase 3 — Day-2 and teardown

- **Infra change:** edit Terraform → `make infra-plan` → `make infra-apply`.
- **App only:** rebuild/push images → migrate (`make gcp-migrate` or via `make gcp-deploy`) → `gcloud run services update` (TF will not revert digests).
- **Teardown:**
  1. Export DB if needed (Cloud SQL export / dump via private path or temporary access).
  2. Set `deletion_protection = false` on the SQL instance (or in console) and apply.
  3. `make infra-destroy` (confirm).
  4. Delete the state bucket last.
  5. Confirm billing in the GCP console; optionally clean Cloudflare / Discord / Sentry.

---

## Makefile targets

| Target                | Purpose                                                 |
|-----------------------|---------------------------------------------------------|
| `make infra-init`     | `terraform init` with `backend.hcl`                     |
| `make infra-fmt`      | `terraform fmt`                                         |
| `make infra-validate` | `terraform validate`                                    |
| `make infra-plan`     | `terraform plan`                                        |
| `make infra-apply`    | `terraform apply`                                       |
| `make infra-destroy`  | `terraform destroy` (confirm)                           |
| `make gcp-push`       | Build/push amd64 images to Artifact Registry            |
| `make gcp-migrate`    | Update migrate Job image + execute goose Up             |
| `make gcp-deploy`     | Push → migrate → update Cloud Run API/frontend services |

---

## Troubleshooting

| Symptom                                          | Likely cause                                                                                                                    |
|--------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| Cloud Run unhealthy after first apply            | Still on placeholder image — push real images                                                                                   |
| Budget create 403 / quota project `764086051850` | Provider quota override missing or Billing Budgets API disabled; re-apply after `providers.tf` fix; confirm billing account IAM |
| 403 on API/pages via `*.run.app`                 | Origin-verify working as intended; use the Cloudflare domain                                                                    |
| 403 through the domain                           | Worker secret out of sync with Secret Manager — re-apply Cloudflare/secret bindings                                             |
| API cannot reach DB                              | Direct VPC / private IP; confirm same region and PSA connection                                                                 |
| Discord login fails                              | Wrong `FRONTEND_URL` / redirect URI / cookie domain                                                                             |
| `invalid timezone` on `/users/me/events`         | API image missing tzdata — fixed by embedding `time/tzdata`; redeploy API                                                       |

---

## Cost notes

- Cloud SQL bills from first successful apply (does not scale to zero).
- Budget alerts are email-only (no auto kill switch).
- Keep `max_instances`, disk size, and no-HA settings unless you accept higher cost.
