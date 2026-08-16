variable "project_id" {
  description = "GCP project ID (created manually in Phase 0)"
  type        = string
}

variable "region" {
  description = "GCP region for Cloud Run, Cloud SQL, and VPC"
  type        = string
  default     = "us-central1"
}

variable "domain" {
  description = "Public apex domain (Cloudflare zone), e.g. matchmaker.games"
  type        = string
}

variable "billing_account_id" {
  description = "GCP billing account ID for budget alerts (XXXXXX-XXXXXX-XXXXXX)"
  type        = string
}

variable "alert_email" {
  description = "Email for budget and uptime notifications"
  type        = string
}

variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for var.domain"
  type        = string
}

variable "cloudflare_api_token" {
  description = "Cloudflare custom API token (Account Workers Scripts Edit + Zone Workers Routes Edit + Zone DNS Edit)"
  type        = string
  sensitive   = true
}

variable "discord_client_id" {
  description = "Discord OAuth2 client ID"
  type        = string
}

variable "discord_client_secret" {
  description = "Discord OAuth2 client secret"
  type        = string
  sensitive   = true
}

variable "db_admin_password" {
  description = "Cloud SQL postgres (admin) password — Studio / Auth Proxy / db-bootstrap only"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.db_admin_password) >= 16
    error_message = "db_admin_password must be at least 16 characters."
  }
}

variable "db_migrator_password" {
  description = "Password for SQL role matchmaker_migrator (goose Job; DDL+DML)"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.db_migrator_password) >= 16
    error_message = "db_migrator_password must be at least 16 characters."
  }
}

variable "db_app_password" {
  description = "Password for SQL role matchmaker_app (API DML-only). Also used for legacy matchmaker user until db_roles_bootstrapped."
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.db_app_password) >= 16
    error_message = "db_app_password must be at least 16 characters."
  }
}

variable "db_roles_bootstrapped" {
  description = "Set true after make gcp-db-bootstrap so API/migrate use least-privilege roles and legacy matchmaker user is removed"
  type        = bool
  default     = false
}

variable "jwt_secret" {
  description = "JWT signing secret as 64 hex chars (openssl rand -hex 32 or make gen-keys)"
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^[0-9a-fA-F]{64}$", var.jwt_secret))
    error_message = "jwt_secret must be exactly 64 hex characters (openssl rand -hex 32)."
  }
}

variable "cookie_hash_key" {
  description = "Securecookie hash key as 64 hex chars (openssl rand -hex 32 or make gen-keys)"
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^[0-9a-fA-F]{64}$", var.cookie_hash_key))
    error_message = "cookie_hash_key must be exactly 64 hex characters (openssl rand -hex 32)."
  }
}

variable "cookie_encrypt_key" {
  description = "Securecookie encrypt key as 64 hex chars (openssl rand -hex 32 or make gen-keys)"
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^[0-9a-fA-F]{64}$", var.cookie_encrypt_key))
    error_message = "cookie_encrypt_key must be exactly 64 hex characters (openssl rand -hex 32)."
  }
}

variable "origin_verify_secret" {
  description = "Shared X-Origin-Verify secret (Worker + Cloud Run); use a long random string"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.origin_verify_secret) >= 16
    error_message = "origin_verify_secret must be at least 16 characters."
  }
}

variable "sentry_dsn_api" {
  description = "Sentry DSN for the Go API (empty disables)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "sentry_dsn_frontend" {
  description = "Sentry DSN for the Next.js frontend (empty disables)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "db_tier" {
  description = "Cloud SQL machine tier"
  type        = string
  default     = "db-f1-micro"
}

variable "db_disk_gb" {
  description = "Cloud SQL SSD size in GB (autoresize off)"
  type        = number
  default     = 10

  validation {
    condition     = var.db_disk_gb >= 10 && var.db_disk_gb <= 100
    error_message = "db_disk_gb must be between 10 and 100."
  }
}

variable "cloud_run_max_instances" {
  description = "Max Cloud Run instances per service"
  type        = number
  default     = 2
}

variable "cloud_run_cpu" {
  type    = string
  default = "1"
}

variable "cloud_run_memory" {
  type    = string
  default = "512Mi"
}

variable "budget_amounts" {
  description = "Billing budget alert thresholds (USD)"
  type        = list(number)
  default     = [15, 30, 50]
}

variable "labels" {
  description = "Common resource labels"
  type        = map(string)
  default = {
    app        = "matchmaker"
    env        = "prod"
    managed_by = "terraform"
  }
}
