resource "cloudflare_workers_script" "edge" {
  account_id = var.cloudflare_account_id
  name       = "${local.name_prefix}-edge"
  content    = file("${path.module}/../cloudflare/worker.js")
  module     = true

  plain_text_binding {
    name = "API_ORIGIN"
    text = trimsuffix(google_cloud_run_v2_service.api.uri, "/")
  }

  plain_text_binding {
    name = "FRONTEND_ORIGIN"
    text = trimsuffix(google_cloud_run_v2_service.frontend.uri, "/")
  }

  secret_text_binding {
    name = "ORIGIN_VERIFY_SECRET"
    text = var.origin_verify_secret
  }

  # Base64 service-account key JSON; the Worker mints Cloud Run ID tokens from it.
  secret_text_binding {
    name = "GCP_INVOKER_KEY"
    text = google_service_account_key.edge.private_key
  }
}

# Per-IP rate limit on the API. Runs in the WAF phase, before the Worker, so blocked
# requests cost no Worker invocations, Cloud Run instances, or Discord API calls.
# Free plan allows one rule with a fixed 10s period/timeout and IP + colo counting.
# If an http_ratelimit ruleset already exists on the zone (made in the dashboard),
# import it first: terraform import cloudflare_ruleset.api_rate_limit zone/<zone_id>/<ruleset_id>
resource "cloudflare_ruleset" "api_rate_limit" {
  zone_id     = var.cloudflare_zone_id
  name        = "${local.name_prefix}-rate-limit"
  description = "Matchmaker per-IP API rate limit (Terraform)"
  kind        = "zone"
  phase       = "http_ratelimit"

  rules {
    action      = "block"
    expression  = "(starts_with(http.request.uri.path, \"/api/\"))"
    description = "Per-IP /api/* limit"
    enabled     = true

    ratelimit {
      characteristics     = ["ip.src", "cf.colo.id"]
      period              = 10
      requests_per_period = var.api_rate_limit_per_10s
      mitigation_timeout  = 10
    }
  }
}

resource "cloudflare_workers_route" "edge" {
  zone_id     = var.cloudflare_zone_id
  pattern     = "${var.domain}/*"
  script_name = cloudflare_workers_script.edge.name
}

# Proxied apex record so the Workers route can serve HTTPS on the zone.
resource "cloudflare_record" "apex" {
  zone_id = var.cloudflare_zone_id
  name    = "@"
  type    = "AAAA"
  content = "100::"
  proxied = true
  comment = "Matchmaker Worker (Terraform)"
}
