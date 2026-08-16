output "artifact_registry_repo" {
  description = "Artifact Registry repository ID"
  value       = google_artifact_registry_repository.docker.repository_id
}

output "artifact_registry_url" {
  description = "Docker push base (region-docker.pkg.dev/project/repo)"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker.repository_id}"
}

output "cloud_run_api_uri" {
  description = "Cloud Run API URL (Worker origin; use public domain for clients)"
  value       = google_cloud_run_v2_service.api.uri
}

output "cloud_run_frontend_uri" {
  description = "Cloud Run frontend URL (Worker origin; use public domain for clients)"
  value       = google_cloud_run_v2_service.frontend.uri
}

output "cloud_run_api_name" {
  value = google_cloud_run_v2_service.api.name
}

output "cloud_run_frontend_name" {
  value = google_cloud_run_v2_service.frontend.name
}

output "cloud_run_migrate_job_name" {
  description = "Cloud Run Job that runs `server migrate` (goose Up)"
  value       = google_cloud_run_v2_job.migrate.name
}

output "cloud_run_db_bootstrap_job_name" {
  description = "Cloud Run Job that runs `server db-bootstrap` (least-privilege roles)"
  value       = google_cloud_run_v2_job.db_bootstrap.name
}

output "cloud_sql_connection_name" {
  value = google_sql_database_instance.main.connection_name
}

output "cloud_sql_private_ip" {
  value     = google_sql_database_instance.main.private_ip_address
  sensitive = true
}

output "vpc_network" {
  value = google_compute_network.main.name
}
