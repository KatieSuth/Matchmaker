# Production infrastructure — apply only after review

> Applies to: any command that changes live GCP, Terraform or Cloudflare state, from any directory.

The operator always reviews production changes. File edits in `infra/` are fine; **mutating prod is not**, until they have seen the plan and OK'd that run.

## Never run without an explicit OK

Do **not** execute any of these until the user has reviewed the proposed command/plan and replied with a clear OK for that run (e.g. "ok", "apply it"):

- `terraform apply`, `terraform destroy`, `terraform apply -auto-approve`
- `make infra-apply`, `make infra-destroy`
- `make gcp-push`, `make gcp-deploy`, `make gcp-migrate`, `make gcp-db-bootstrap`
- `gcloud run services update`, `gcloud run jobs execute`, or any other mutating `gcloud` against the prod project
- Cloudflare / GCP console-equivalent APIs that change live DNS, workers, SQL, or Run

Agent mode, "apply the change", or "update Terraform" means **edit the files** (and optionally `terraform plan` / `make infra-plan`). It is **not** permission to apply.

## Allowed without OK

- Edit `.tf` / Worker / docs
- `terraform fmt`, `terraform validate`, `make infra-plan` / `terraform plan` (read-only)
- Show the plan output and the exact apply command you would run

## Required workflow

1. Make the repo change.
2. Show `make infra-plan` (or the exact `gcloud`/`make gcp-*` command) and wait.
3. Apply only after they OK **that** plan/command. A new plan after extra edits needs a new OK.
4. Never pass `-auto-approve` unless they explicitly asked for it on that run.
