locals {
  name_prefix = "matchmaker"

  frontend_url = "https://${var.domain}"
  api_public   = "https://${var.domain}/api"

  discord_redirect_uri = "https://${var.domain}/api/auth/discord_redirect"

  # Placeholder until you push real images (see infra/README). ignore_changes keeps digests.
  placeholder_image = "us-docker.pkg.dev/cloudrun/container/hello"

  labels = var.labels

  # Postgres role names created by `server db-bootstrap` (not cloudsqlsuperuser).
  db_app_role      = "matchmaker_app"
  db_migrator_role = "matchmaker_migrator"
  db_name          = "matchmaker"

  db_host = google_sql_database_instance.main.private_ip_address

  # Until bootstrap: API + migrate share legacy cloudsqlsuperuser `matchmaker`.
  # After: API → matchmaker_app, migrate → matchmaker_migrator.
  database_url_app = var.db_roles_bootstrapped ? format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    local.db_app_role,
    var.db_app_password,
    local.db_host,
    local.db_name,
    ) : format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    "matchmaker",
    var.db_app_password,
    local.db_host,
    local.db_name,
  )

  database_url_migrate = var.db_roles_bootstrapped ? format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    local.db_migrator_role,
    var.db_migrator_password,
    local.db_host,
    local.db_name,
    ) : format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    "matchmaker",
    var.db_app_password,
    local.db_host,
    local.db_name,
  )

  database_url_admin = format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    "postgres",
    var.db_admin_password,
    local.db_host,
    local.db_name,
  )
}
