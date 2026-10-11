# GitHub Actions deploy identity (OIDC / Workload Identity Federation).
# No JSON keys: the Deploy workflow on x.x.x tags exchanges a GitHub OIDC token
# for a short-lived credential on matchmaker-github-deploy.

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-actions"
  display_name              = "GitHub Actions"
  description               = "OIDC pool for Matchmaker GitHub Actions deploys"

  depends_on = [google_project_service.services]
}

# Trust GitHub's OIDC issuer. Defense in depth only: whoever can push an x.x.x tag
# controls the workflow file that runs, so the real lock is the GitHub tag ruleset
# plus the `production` environment's deployment tag rules (see infra/README.md).
# Repository/owner IDs (not names) stop a renamed or re-registered repo inheriting trust.
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-oidc"
  display_name                       = "GitHub OIDC"
  description                        = "GitHub Actions OIDC for ${var.github_repository}"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_mapping = {
    "google.subject"                = "assertion.sub"
    "attribute.actor"               = "assertion.actor"
    "attribute.repository"          = "assertion.repository"
    "attribute.repository_id"       = "assertion.repository_id"
    "attribute.repository_owner_id" = "assertion.repository_owner_id"
    "attribute.ref"                 = "assertion.ref"
    "attribute.environment"         = "assertion.environment"
    "attribute.workflow_ref"        = "assertion.workflow_ref"
  }

  attribute_condition = join(" && ", [
    "assertion.repository_id == '${var.github_repository_id}'",
    "assertion.repository_owner_id == '${var.github_repository_owner_id}'",
    "assertion.ref.matches('^refs/tags/[0-9]+[.][0-9]+[.][0-9]+$')",
    "assertion.environment == 'production'",
    "assertion.workflow_ref.startsWith('${var.github_repository}/.github/workflows/deploy.yml@refs/tags/')",
  ])
}

resource "google_service_account" "github_deploy" {
  account_id   = "${local.name_prefix}-github-deploy"
  display_name = "GitHub Actions deploy"
  description  = "Push images and roll Cloud Run from GitHub Actions (WIF)"
}

# Allow the GitHub repo (via WIF) to impersonate the deploy SA.
resource "google_service_account_iam_member" "github_deploy_wif" {
  service_account_id = google_service_account.github_deploy.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository_id/${var.github_repository_id}"
}

# Push linux/amd64 API and frontend images during make gcp-deploy.
resource "google_artifact_registry_repository_iam_member" "github_deploy_writer" {
  project    = var.project_id
  location   = var.region
  repository = google_artifact_registry_repository.docker.repository_id
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.github_deploy.email}"
}

# Update the two Cloud Run services and update/execute the migrate Job. Granted per
# resource (not project-wide) so the deploy SA cannot touch db-bootstrap or create
# new services/jobs.
resource "google_cloud_run_v2_service_iam_member" "github_deploy_api" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.developer"
  member   = "serviceAccount:${google_service_account.github_deploy.email}"
}

resource "google_cloud_run_v2_service_iam_member" "github_deploy_frontend" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.developer"
  member   = "serviceAccount:${google_service_account.github_deploy.email}"
}

resource "google_cloud_run_v2_job_iam_member" "github_deploy_migrate" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_job.migrate.name
  role     = "roles/run.developer"
  member   = "serviceAccount:${google_service_account.github_deploy.email}"
}

# Pre-migration on-demand backup (`make gcp-db-backup`). Custom role so the deploy
# SA can create backups but not restore, delete, or otherwise administer Cloud SQL.
resource "google_project_iam_custom_role" "github_deploy_sql_backup" {
  role_id     = "matchmakerDeploySqlBackup"
  title       = "Matchmaker deploy SQL backup"
  description = "Create and read Cloud SQL on-demand backups before migrations"
  permissions = [
    "cloudsql.backupRuns.create",
    "cloudsql.backupRuns.get",
    "cloudsql.backupRuns.list",
    "cloudsql.instances.get",
  ]
}

resource "google_project_iam_member" "github_deploy_sql_backup" {
  project = var.project_id
  role    = google_project_iam_custom_role.github_deploy_sql_backup.id
  member  = "serviceAccount:${google_service_account.github_deploy.email}"
}

# Rolling a service that runs as another SA requires iam.serviceAccounts.actAs.
# Residual risk: an image deployed by this SA runs as the API SA, which can read
# its secrets and reach Cloud SQL — inherent to letting CI ship API code.
resource "google_service_account_iam_member" "github_deploy_act_as_api" {
  service_account_id = google_service_account.api.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_deploy.email}"
}

resource "google_service_account_iam_member" "github_deploy_act_as_frontend" {
  service_account_id = google_service_account.frontend.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_deploy.email}"
}
