resource "google_artifact_registry_repository" "docker" {
  location      = var.region
  repository_id = "${local.name_prefix}-docker"
  description   = "Matchmaker container images"
  format        = "DOCKER"
  labels        = local.labels

  # A pushed x.x.x tag always means the same image; re-runs reuse it instead of overwriting.
  docker_config {
    immutable_tags = true
  }

  depends_on = [google_project_service.services]
}
