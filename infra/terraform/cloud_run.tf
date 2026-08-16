resource "google_cloud_run_v2_service" "api" {
  name     = "${local.name_prefix}-api"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  labels = local.labels

  template {
    service_account = google_service_account.api.email

    scaling {
      min_instance_count = 0
      max_instance_count = var.cloud_run_max_instances
    }

    vpc_access {
      egress = "PRIVATE_RANGES_ONLY"
      network_interfaces {
        network    = google_compute_network.main.id
        subnetwork = google_compute_subnetwork.main.id
      }
    }

    containers {
      name  = "api"
      image = local.placeholder_image

      resources {
        limits = {
          cpu    = var.cloud_run_cpu
          memory = var.cloud_run_memory
        }
      }

      ports {
        container_port = 8080
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 5
        period_seconds        = 5
        failure_threshold     = 12
      }

      liveness_probe {
        http_get {
          path = "/health"
        }
        period_seconds = 30
      }

      env {
        name  = "GIN_MODE"
        value = "release"
      }
      env {
        name  = "FRONTEND_URL"
        value = local.frontend_url
      }
      env {
        name  = "COOKIE_DOMAIN"
        value = var.domain
      }
      env {
        name  = "DISCORD_CLIENT_ID"
        value = var.discord_client_id
      }
      env {
        name  = "DISCORD_REDIRECT_URI"
        value = local.discord_redirect_uri
      }
      env {
        name  = "DB_MAX_CONNS"
        value = "5"
      }
      # Local Compose leaves AUTO_MIGRATE unset (migrate on API start).
      # Production migrates via the matchmaker-migrate Cloud Run Job instead.
      env {
        name  = "AUTO_MIGRATE"
        value = "false"
      }
      # Cloud Run sits behind Cloudflare; trust X-Forwarded-For from the platform.
      env {
        name  = "TRUSTED_PROXIES"
        value = "0.0.0.0/0"
      }

      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.database_url.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "JWT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.jwt.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "COOKIE_HASH_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.cookie_hash.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "COOKIE_ENCRYPT_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.cookie_encrypt.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "DISCORD_CLIENT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.discord_client_secret.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "ORIGIN_VERIFY_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.origin_verify.secret_id
            version = "latest"
          }
        }
      }

      dynamic "env" {
        for_each = var.sentry_dsn_api != "" ? [1] : []
        content {
          name = "SENTRY_DSN"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.sentry_dsn_api[0].secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.database_url,
    google_secret_manager_secret_version.jwt,
    google_secret_manager_secret_version.cookie_hash,
    google_secret_manager_secret_version.cookie_encrypt,
    google_secret_manager_secret_version.discord_client_secret,
    google_secret_manager_secret_version.origin_verify,
  ]

  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].containers[0].image,
    ]
  }
}

# One-shot goose Up against private Cloud SQL. Execute after pushing a new API image
# and before (or as part of) rolling the API service — see `make gcp-deploy`.
resource "google_cloud_run_v2_job" "migrate" {
  name     = "${local.name_prefix}-migrate"
  location = var.region

  labels = local.labels

  template {
    template {
      service_account = google_service_account.api.email
      timeout         = "300s"
      max_retries     = 1

      vpc_access {
        egress = "PRIVATE_RANGES_ONLY"
        network_interfaces {
          network    = google_compute_network.main.id
          subnetwork = google_compute_subnetwork.main.id
        }
      }

      containers {
        name  = "migrate"
        image = local.placeholder_image
        args  = ["migrate"]

        resources {
          limits = {
            cpu    = var.cloud_run_cpu
            memory = var.cloud_run_memory
          }
        }

        env {
          name  = "GIN_MODE"
          value = "release"
        }

        env {
          name = "DATABASE_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.database_url_migrate.secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.database_url_migrate,
    google_sql_database_instance.main,
  ]

  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].template[0].containers[0].image,
    ]
  }
}

# One-shot least-privilege role bootstrap (CREATE ROLE matchmaker_app / matchmaker_migrator).
# Run once via `make gcp-db-bootstrap` before setting db_roles_bootstrapped=true.
resource "google_cloud_run_v2_job" "db_bootstrap" {
  name     = "${local.name_prefix}-db-bootstrap"
  location = var.region

  labels = local.labels

  template {
    template {
      service_account = google_service_account.api.email
      timeout         = "300s"
      max_retries     = 1

      vpc_access {
        egress = "PRIVATE_RANGES_ONLY"
        network_interfaces {
          network    = google_compute_network.main.id
          subnetwork = google_compute_subnetwork.main.id
        }
      }

      containers {
        name  = "db-bootstrap"
        image = local.placeholder_image
        args  = ["db-bootstrap"]

        resources {
          limits = {
            cpu    = var.cloud_run_cpu
            memory = var.cloud_run_memory
          }
        }

        env {
          name  = "GIN_MODE"
          value = "release"
        }

        env {
          name = "DATABASE_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.database_url_admin.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "DB_APP_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.db_app_password.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "DB_MIGRATOR_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.db_migrator_password.secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.database_url_admin,
    google_secret_manager_secret_version.db_app_password,
    google_secret_manager_secret_version.db_migrator_password,
    google_sql_database_instance.main,
  ]

  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_service" "frontend" {
  name     = "${local.name_prefix}-frontend"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  labels = local.labels

  template {
    service_account = google_service_account.frontend.email

    scaling {
      min_instance_count = 0
      max_instance_count = var.cloud_run_max_instances
    }

    containers {
      name  = "frontend"
      image = local.placeholder_image

      resources {
        limits = {
          cpu    = var.cloud_run_cpu
          memory = var.cloud_run_memory
        }
      }

      ports {
        container_port = 3000
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 5
        period_seconds        = 5
        failure_threshold     = 12
      }

      liveness_probe {
        http_get {
          path = "/health"
        }
        period_seconds = 30
      }

      env {
        name  = "HOSTNAME"
        value = "0.0.0.0"
      }

      env {
        name = "ORIGIN_VERIFY_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.origin_verify.secret_id
            version = "latest"
          }
        }
      }

      dynamic "env" {
        for_each = var.sentry_dsn_frontend != "" ? [1] : []
        content {
          name = "SENTRY_DSN"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.sentry_dsn_frontend[0].secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.origin_verify,
  ]

  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].containers[0].image,
    ]
  }
}
