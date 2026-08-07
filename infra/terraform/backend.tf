# Configure via: terraform init -backend-config=backend.hcl
# See backend.hcl.example and infra/README.md (Phase 1 bootstrap).
terraform {
  backend "gcs" {}
}
