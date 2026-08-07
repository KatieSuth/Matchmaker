locals {
  name_prefix = "matchmaker"

  frontend_url = "https://${var.domain}"
  api_public   = "https://${var.domain}/api"

  discord_redirect_uri = "https://${var.domain}/api/auth/discord_redirect"

  # Placeholder until you push real images (see infra/README). ignore_changes keeps digests.
  placeholder_image = "us-docker.pkg.dev/cloudrun/container/hello"

  labels = var.labels
}
