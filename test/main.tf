terraform {
  required_providers {
    shodan = {
      source = "AdconnectDevOps/shodan"
      version = "~> 0.1"
    }
  }
}

provider "shodan" {
  api_key = var.shodan_api_key
}

# Test Shodan alert with Slack notifications
resource "shodan_alert" "test_alert" {
  name        = "test-network-5-6-7-8"
  network     = "5.6.7.8/32"
  description = "Test network for IP 5.6.7.8 - Provider test with Slack"
  enabled     = true
  
  tags = [
    "test",
    "provider",
    "monitoring",
    "slack"
  ]
  
  triggers = [
    "end_of_life",
    "industrial_control_system",
    "internet_scanner",
    "iot",
    "malware",
    "new_service",
    "open_database",
    "ssl_expired",
    "uncommon",
    "uncommon_plus",
    "vulnerable",
    "vulnerable_unverified"
  ]
  
  notifiers = [
    "default"
  ]
  
  slack_notifications = var.slack_notifications
}

# Outputs
output "alert_id" {
  description = "ID of the created Shodan alert"
  value       = shodan_alert.test_alert.id
}

output "alert_name" {
  description = "Name of the created Shodan alert"
  value       = shodan_alert.test_alert.name
}

output "alert_network" {
  description = "Network being monitored"
  value       = shodan_alert.test_alert.network
}

output "slack_channels" {
  description = "Slack channels configured for notifications"
  value       = shodan_alert.test_alert.slack_notifications
  sensitive   = true
}
