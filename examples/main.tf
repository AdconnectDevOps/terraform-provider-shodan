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

# Create a Shodan network alert
resource "shodan_alert" "production_network" {
  name        = "production-network-monitoring"
  network     = "192.168.1.0/24"
  description = "Monitor production network for security threats"
  enabled     = true
  
  tags = [
    "production",
    "security",
    "monitoring"
  ]
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "open_database",
    "ssl_expired"
  ]
  
  notifiers = [
    "default"  # Default email notifier
  ]
  
  slack_notifications = var.slack_notifications
}

# Create another alert for DMZ
resource "shodan_alert" "dmz_network" {
  name        = "dmz-network-monitoring"
  network     = "10.0.0.0/24"
  description = "Monitor DMZ network for external threats"
  enabled     = true
  
  tags = [
    "dmz",
    "external",
    "security"
  ]
  
  triggers = [
    "malware",
    "vulnerable",
    "internet_scanner",
    "iot"
  ]
  
  notifiers = [
    "default"
  ]
  
  slack_notifications = var.slack_notifications
}

# Data source to read an existing alert
data "shodan_alert" "existing_alert" {
  id = "EXISTING_ALERT_ID"  # Replace with actual alert ID
}

# Outputs
output "production_alert_id" {
  description = "ID of the production network alert"
  value       = shodan_alert.production_network.id
}

output "dmz_alert_id" {
  description = "ID of the DMZ network alert"
  value       = shodan_alert.dmz_network.id
}

output "existing_alert_name" {
  description = "Name of the existing alert"
  value       = data.shodan_alert.existing_alert.name
}
