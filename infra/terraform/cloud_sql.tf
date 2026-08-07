resource "google_sql_database_instance" "main" {
  name             = "${local.name_prefix}-pg"
  database_version = "POSTGRES_18"
  region           = var.region

  deletion_protection = true

  settings {
    tier              = var.db_tier
    edition           = "ENTERPRISE" # required for db-f1-micro; PG16+ defaults to ENTERPRISE_PLUS
    availability_type = "ZONAL"
    disk_size         = var.db_disk_gb
    disk_type         = "PD_SSD"
    disk_autoresize   = false

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.main.id
      enable_private_path_for_google_cloud_services = true
    }

    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"
      point_in_time_recovery_enabled = false
      backup_retention_settings {
        retained_backups = 7
        retention_unit   = "COUNT"
      }
    }

    user_labels = local.labels
  }

  depends_on = [
    google_service_networking_connection.private_vpc,
    google_project_service.services,
  ]
}

resource "google_sql_database" "app" {
  name     = "matchmaker"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "app" {
  name     = "matchmaker"
  instance = google_sql_database_instance.main.name
  password = var.db_password
}
