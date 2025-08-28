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
  description = "Your Shodan API key"
  type        = string
  sensitive   = true
}

# Data source to get domain information
data "shodan_domain" "example" {
  domain = "example.com"
}

# Output domain information
output "domain_info" {
  description = "Information about the example.com domain"
  value = {
    domain     = data.shodan_domain.example.domain
    tags       = data.shodan_domain.example.tags
    subdomains = data.shodan_domain.example.subdomains
    dns_records = data.shodan_domain.example.data
  }
}

# Resource to monitor a domain for security threats
resource "shodan_domain" "example_monitoring" {
  domain      = "example.com"
  name        = "Example Domain Security Monitoring"
  description = "Monitor example.com for security threats and vulnerabilities"
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

# Output the created domain alert
output "domain_alert" {
  description = "Information about the created domain monitoring alert"
  value = {
    id         = shodan_domain.example_monitoring.id
    domain     = shodan_domain.example_monitoring.domain
    name       = shodan_domain.example_monitoring.name
    created_at = shodan_domain.example_monitoring.created_at
  }
}
