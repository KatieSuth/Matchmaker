resource "google_artifact_registry_repository" "docker" {
  location      = var.region
  repository_id = "${local.name_prefix}-docker"
  description   = "Matchmaker container images"
  format        = "DOCKER"
  labels        = local.labels

  depends_on = [google_project_service.services]
}
