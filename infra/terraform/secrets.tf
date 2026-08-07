resource "google_secret_manager_secret" "db_password" {
  secret_id = "${local.name_prefix}-db-password"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.db_password
}

resource "google_secret_manager_secret" "jwt" {
  secret_id = "${local.name_prefix}-jwt-secret"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "jwt" {
  secret      = google_secret_manager_secret.jwt.id
  secret_data = var.jwt_secret
}

resource "google_secret_manager_secret" "cookie_hash" {
  secret_id = "${local.name_prefix}-cookie-hash"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "cookie_hash" {
  secret      = google_secret_manager_secret.cookie_hash.id
  secret_data = var.cookie_hash_key
}

resource "google_secret_manager_secret" "cookie_encrypt" {
  secret_id = "${local.name_prefix}-cookie-encrypt"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "cookie_encrypt" {
  secret      = google_secret_manager_secret.cookie_encrypt.id
  secret_data = var.cookie_encrypt_key
}

resource "google_secret_manager_secret" "origin_verify" {
  secret_id = "${local.name_prefix}-origin-verify"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "origin_verify" {
  secret      = google_secret_manager_secret.origin_verify.id
  secret_data = var.origin_verify_secret
}

resource "google_secret_manager_secret" "discord_client_secret" {
  secret_id = "${local.name_prefix}-discord-client-secret"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "discord_client_secret" {
  secret      = google_secret_manager_secret.discord_client_secret.id
  secret_data = var.discord_client_secret
}

resource "google_secret_manager_secret" "sentry_dsn_api" {
  count     = var.sentry_dsn_api != "" ? 1 : 0
  secret_id = "${local.name_prefix}-sentry-dsn-api"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "sentry_dsn_api" {
  count       = var.sentry_dsn_api != "" ? 1 : 0
  secret      = google_secret_manager_secret.sentry_dsn_api[0].id
  secret_data = var.sentry_dsn_api
}

resource "google_secret_manager_secret" "sentry_dsn_frontend" {
  count     = var.sentry_dsn_frontend != "" ? 1 : 0
  secret_id = "${local.name_prefix}-sentry-dsn-frontend"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "sentry_dsn_frontend" {
  count       = var.sentry_dsn_frontend != "" ? 1 : 0
  secret      = google_secret_manager_secret.sentry_dsn_frontend[0].id
  secret_data = var.sentry_dsn_frontend
}

resource "google_secret_manager_secret" "database_url" {
  secret_id = "${local.name_prefix}-database-url"
  labels    = local.labels
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "database_url" {
  secret = google_secret_manager_secret.database_url.id
  secret_data = format(
    "postgres://%s:%s@%s:5432/%s?sslmode=disable",
    google_sql_user.app.name,
    var.db_password,
    google_sql_database_instance.main.private_ip_address,
    google_sql_database.app.name,
  )

  depends_on = [
    google_sql_database_instance.main,
    google_sql_user.app,
    google_sql_database.app,
  ]
}
