resource "google_billing_budget" "alerts" {
  for_each = { for amt in var.budget_amounts : tostring(amt) => amt }

  billing_account = var.billing_account_id
  display_name    = "Matchmaker ${each.value} USD"

  budget_filter {
    projects = ["projects/${data.google_project.current.number}"]
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units         = tostring(each.value)
    }
  }

  threshold_rules {
    threshold_percent = 1.0
  }

  all_updates_rule {
    monitoring_notification_channels = [
      google_monitoring_notification_channel.email.id,
    ]
    disable_default_iam_recipients = true
  }

  depends_on = [google_project_service.services]
}

data "google_project" "current" {
  project_id = var.project_id
}
