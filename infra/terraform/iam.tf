resource "google_service_account" "api" {
  account_id   = "${local.name_prefix}-api"
  display_name = "Matchmaker API Cloud Run"
}

resource "google_service_account" "frontend" {
  account_id   = "${local.name_prefix}-frontend"
  display_name = "Matchmaker Frontend Cloud Run"
}

resource "google_project_iam_member" "api_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "api_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "frontend_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.frontend.email}"
}

# Identity the Cloudflare Worker uses to invoke Cloud Run. The Worker signs a JWT with this
# key, exchanges it for a Google ID token, and sends it as X-Serverless-Authorization, so
# Cloud Run IAM rejects direct *.run.app traffic before any container starts.
resource "google_service_account" "edge" {
  account_id   = "${local.name_prefix}-edge"
  display_name = "Matchmaker Cloudflare Worker (Cloud Run invoker)"
}

# Key material lands in Terraform state (like the other secrets) and in the Worker secret
# binding. Rotate with: terraform apply -replace=google_service_account_key.edge
resource "google_service_account_key" "edge" {
  service_account_id = google_service_account.edge.name
}

resource "google_cloud_run_v2_service_iam_member" "api_invoker_edge" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.edge.email}"
}

resource "google_cloud_run_v2_service_iam_member" "frontend_invoker_edge" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.edge.email}"
}

# Legacy public invoke. Kept only while var.cloud_run_public_invoker is true (rollout step 1);
# origin-verify still gates app traffic either way.
resource "google_cloud_run_v2_service_iam_member" "api_invoker" {
  count    = var.cloud_run_public_invoker ? 1 : 0
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "frontend_invoker" {
  count    = var.cloud_run_public_invoker ? 1 : 0
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

moved {
  from = google_cloud_run_v2_service_iam_member.api_invoker
  to   = google_cloud_run_v2_service_iam_member.api_invoker[0]
}

moved {
  from = google_cloud_run_v2_service_iam_member.frontend_invoker
  to   = google_cloud_run_v2_service_iam_member.frontend_invoker[0]
}
