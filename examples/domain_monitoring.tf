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

variable "shodan_api_key" {
  description = "Shodan API key for authentication"
  type        = string
  sensitive   = true
}

# Get domain information
data "shodan_domain" "example_domain" {
  domain = "example.com"
}

# Basic domain monitoring
resource "shodan_domain" "basic_monitoring" {
  domain = "example.com"
}

# Custom domain monitoring with specific triggers
resource "shodan_domain" "custom_monitoring" {
  domain      = "google.com"
  name        = "Google Domain Security Monitoring"
  description = "Monitor Google's domain for security threats"
  enabled     = true
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}

# Comprehensive security monitoring
resource "shodan_domain" "comprehensive_monitoring" {
  domain      = "github.com"
  name        = "GitHub Comprehensive Security"
  description = "Comprehensive security monitoring for GitHub"
  enabled     = true
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "vulnerable_unverified",
    "new_service",
    "open_database",
    "ssl_expired",
    "iot",
    "end_of_life",
    "industrial_control_system",
    "internet_scanner",
    "uncommon",
    "uncommon_plus"
  ]
  
  notifiers = ["default"]
}

# Monitor multiple domains using for_each
resource "shodan_domain" "multiple_domains" {
  for_each = toset([
    "cloudflare.com",
    "amazon.com",
    "microsoft.com"
  ])
  
  domain      = each.value
  name        = "Enterprise Domain: ${each.value}"
  description = "Monitor enterprise domain for security threats"
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}

# Domain monitoring with Slack notifications
resource "shodan_domain" "slack_monitoring" {
  domain      = "github.com"
  name        = "GitHub with Slack Alerts"
  description = "Monitor GitHub domain with Slack notifications"
  enabled     = true
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired",
    "iot"
  ]
  
  # Use your actual Slack notifier ID from Shodan account settings
  slack_notifications = ["slack_12345", "slack_67890"]  # Replace with actual IDs
  
  # You can also use regular notifiers alongside Slack
  notifiers = ["default"]
}

# Outputs
output "example_domain_info" {
  description = "Information about example.com domain"
  value = {
    domain     = data.shodan_domain.example_domain.domain
    tags       = data.shodan_domain.example_domain.tags
    subdomains = data.shodan_domain.example_domain.subdomains
    dns_records_count = length(data.shodan_domain.example_domain.data)
  }
}

output "basic_monitoring_id" {
  description = "ID of the basic domain monitoring alert"
  value       = shodan_domain.basic_monitoring.id
}

output "custom_monitoring_id" {
  description = "ID of the custom domain monitoring alert"
  value       = shodan_domain.custom_monitoring.id
}

output "comprehensive_monitoring_id" {
  description = "ID of the comprehensive domain monitoring alert"
  value       = shodan_domain.comprehensive_monitoring.id
}

output "multiple_domains_ids" {
  description = "IDs of all multiple domain monitoring alerts"
  value = {
    for domain, resource in shodan_domain.multiple_domains : domain => resource.id
  }
}
