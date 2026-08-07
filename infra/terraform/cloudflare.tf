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
