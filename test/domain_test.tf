
# Test domain data source
data "shodan_domain" "test_domain" {
  domain = "example.com"
}

# Test domain monitoring resource
resource "shodan_domain" "test_domain_monitoring" {
  domain      = "example.com"
  name        = "Test Domain Security Monitoring"
  description = "Test domain monitoring for example.com"
  enabled     = true
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "open_database",
    "ssl_expired",
    "iot"
  ]
  
  notifiers = ["default"]
}

# Test domain monitoring with Slack notifications
resource "shodan_domain" "test_domain_slack" {
  domain      = "google.com"
  name        = "Google Domain with Slack Notifications"
  description = "Monitor Google domain with Slack alerts"
  enabled     = true
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  # Use your actual Slack notifier ID from Shodan account settings
  slack_notifications = ["slack_12345"]  # Replace with your actual Slack notifier ID
}

# Outputs
output "domain_info" {
  description = "Information about the test domain"
  value = {
    domain     = data.shodan_domain.test_domain.domain
    tags       = data.shodan_domain.test_domain.tags
    subdomains = data.shodan_domain.test_domain.subdomains
    dns_records_count = length(data.shodan_domain.test_domain.data)
  }
}

output "domain_alert_id" {
  description = "ID of the created domain monitoring alert"
  value       = shodan_domain.test_domain_monitoring.id
}

output "domain_alert_name" {
  description = "Name of the created domain monitoring alert"
  value       = shodan_domain.test_domain_monitoring.name
}

output "domain_alert_domain" {
  description = "Domain being monitored"
  value       = shodan_domain.test_domain_monitoring.domain
}

output "domain_alert_created" {
  description = "When the domain alert was created"
  value       = shodan_domain.test_domain_monitoring.created_at
}
