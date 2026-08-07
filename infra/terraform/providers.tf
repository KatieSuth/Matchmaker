provider "google" {
  project = var.project_id
  region  = var.region

  # User ADC does not set a quota project by default; APIs like billingbudgets
  # then hit Google's shared ADC project (764086051850) and fail with 403.
  # These force X-Goog-User-Project to our GCP project on every request.
  user_project_override = true
  billing_project       = var.project_id
}

provider "google-beta" {
  project = var.project_id
  region  = var.region

  user_project_override = true
  billing_project       = var.project_id
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
