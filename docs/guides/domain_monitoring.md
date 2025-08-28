---
page_title: "Domain Monitoring Guide"
description: |-
  Learn how to use the Shodan provider to monitor domains for security threats.
---

# Domain Monitoring Guide

This guide explains how to use the Shodan provider to monitor domains for security threats using Terraform. Domain monitoring automatically resolves domains to IP addresses and creates Shodan alerts for comprehensive security coverage.

## Overview

Domain monitoring provides several advantages over traditional IP-based monitoring:

- **Automatic IP Discovery**: No need to manually track IP addresses for domains
- **Dynamic Monitoring**: Automatically adapts to IP address changes
- **Comprehensive Coverage**: Monitors all IP addresses associated with a domain
- **Easy Management**: Monitor domains instead of individual IP ranges

## How Domain Monitoring Works

### 1. Domain Resolution
The provider automatically resolves domains to IP addresses:
- Queries Shodan's DNS API for all DNS records (A, AAAA)
- Extracts unique IP addresses from the domain data
- Handles both IPv4 and IPv6 addresses

### 2. Alert Creation
Creates a Shodan alert with the resolved IP addresses:
- Uses the domain name in the alert name (e.g., `__domain: example.com`)
- Applies the specified security triggers
- Associates the configured notifiers

### 3. IP Monitoring
Monitors all IP addresses associated with the domain for:
- Security threats (malware, vulnerabilities)
- Infrastructure changes (new services, open databases)
- Compliance issues (SSL expiration, end-of-life software)
- Unusual activity (IoT devices, industrial control systems)

## Basic Domain Monitoring

### Simple Domain Alert
```hcl
# Monitor a domain with default settings
resource "shodan_domain" "basic" {
  domain = "example.com"
}
```

This creates a domain monitoring alert with:
- Default triggers (malware, vulnerable, new_service)
- Default notifier (email)
- Automatic naming (`__domain: example.com`)

### Custom Domain Alert
```hcl
# Monitor a domain with custom configuration
resource "shodan_domain" "custom" {
  domain      = "company.com"
  name        = "Company Domain Security"
  description = "Monitor company domain for security threats"
  enabled     = true
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}
```

## Advanced Domain Monitoring

### Comprehensive Security Monitoring
```hcl
# Monitor a domain with all available security triggers
resource "shodan_domain" "comprehensive" {
  domain      = "critical-service.com"
  name        = "Critical Service Security"
  description = "Comprehensive security monitoring for critical service"
  enabled     = true
  
  triggers = [
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
```

### Multiple Domain Monitoring
```hcl
# Monitor multiple domains using for_each
resource "shodan_domain" "enterprise_domains" {
  for_each = toset([
    "company.com",
    "subsidiary.com",
    "partner.com"
  ])
  
  domain      = each.value
  name        = "Enterprise Domain: ${each.value}"
  description = "Monitor enterprise domain for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}
```

## Advanced Configuration

### Custom Triggers

You can customize which security events trigger alerts:

```hcl
resource "shodan_domain" "custom_triggers" {
  domain = "example.com"
  name   = "Custom Trigger Monitoring"
  
  triggers = [
    "malware",           # Detect malware infections
    "vulnerable",        # Find vulnerable services
    "new_service",       # Alert on new services
    "ssl_expired",       # SSL certificate expiration
    "iot",               # Internet of Things devices
    "open_database"      # Open database instances
  ]
}
```

### Slack Notifications

Enable Slack notifications by specifying your Slack notifier IDs:

```hcl
resource "shodan_domain" "slack_monitoring" {
  domain      = "example.com"
  name        = "Slack-Enabled Monitoring"
  description = "Domain monitoring with Slack alerts"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  # Use your actual Slack notifier ID from Shodan account settings
  slack_notifications = ["slack_12345"]
  
  # You can combine with regular notifiers
  notifiers = ["default"]
}
```

**Note:** To get your Slack notifier ID, go to your Shodan account settings and configure Slack integrations.

## Domain Information Retrieval

### Get Domain Details
```hcl
# Retrieve domain information
data "shodan_domain" "info" {
  domain = "example.com"
}

# Use domain information
output "domain_details" {
  value = {
    domain     = data.shodan_domain.info.domain
    tags       = data.shodan_domain.info.tags
    subdomains = data.shodan_domain.info.subdomains
    dns_records_count = length(data.shodan_domain.info.data)
  }
}
```

### Domain Analysis
```hcl
# Analyze domain infrastructure
data "shodan_domain" "analysis" {
  domain = "target.com"
}

output "infrastructure_analysis" {
  value = {
    has_ipv6 = contains(data.shodan_domain.analysis.tags, "ipv6")
    has_dmarc = contains(data.shodan_domain.analysis.tags, "dmarc")
    has_spf = contains(data.shodan_domain.analysis.tags, "spf")
    subdomain_count = length(data.shodan_domain.analysis.subdomains)
    dns_record_count = length(data.shodan_domain.analysis.data)
  }
}
```

## Use Cases

### Company Domain Security
```hcl
# Monitor your company's main domain
resource "shodan_domain" "company_main" {
  domain      = "mycompany.com"
  name        = "Company Main Domain Security"
  description = "Monitor main company domain for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}
```

### Customer Domain Monitoring
```hcl
# Monitor customer domains for security issues
resource "shodan_domain" "customer_domains" {
  for_each = var.customer_domains
  
  domain      = each.value
  name        = "Customer Domain: ${each.value}"
  description = "Monitor customer domain for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}

variable "customer_domains" {
  description = "List of customer domains to monitor"
  type        = list(string)
  default     = ["customer1.com", "customer2.com"]
}
```

### Third-Party Service Monitoring
```hcl
# Monitor critical third-party services
resource "shodan_domain" "third_party" {
  domain      = "payment-gateway.com"
  name        = "Payment Gateway Security"
  description = "Monitor payment gateway for security issues"
  
  triggers = [
    "malware",
    "vulnerable",
    "ssl_expired",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

## Best Practices

### 1. Trigger Selection
Choose triggers based on your security requirements:
- **Essential**: `malware`, `vulnerable`, `new_service`
- **Compliance**: `ssl_expired`, `dmarc`, `spf`
- **Advanced**: `iot`, `industrial_control_system`, `uncommon`

### 2. Naming Conventions
Use descriptive names for easy identification:
```hcl
resource "shodan_domain" "production" {
  domain = "prod.company.com"
  name   = "Production Domain Security"
  # ...
}
```

### 3. Environment Separation
Separate monitoring by environment:
```hcl
resource "shodan_domain" "production" {
  domain = "prod.company.com"
  name   = "[PROD] Company Domain Security"
  # ...
}

resource "shodan_domain" "staging" {
  domain = "staging.company.com"
  name   = "[STAGING] Company Domain Security"
  # ...
}
```

### 4. Notifier Configuration
Configure appropriate notifiers for different domains:
```hcl
# Critical domains with immediate notifications
resource "shodan_domain" "critical" {
  domain = "critical.company.com"
  notifiers = ["default", "slack_urgent"]
  # ...
}

# Standard domains with regular notifications
resource "shodan_domain" "standard" {
  domain = "standard.company.com"
  notifiers = ["default"]
  # ...
}
```

## Troubleshooting

### Common Issues

#### 1. Domain Not Found
```hcl
# Ensure the domain exists and is accessible
data "shodan_domain" "test" {
  domain = "nonexistent.com"
}
```

**Solution**: Verify the domain exists and is accessible from the internet.

#### 2. No IP Addresses Found
```hcl
# Check if the domain has DNS records
data "shodan_domain" "check" {
  domain = "example.com"
}

output "dns_check" {
  value = length(data.shodan_domain.check.data)
}
```

**Solution**: Ensure the domain has active DNS records (A or AAAA).

#### 3. Alert Creation Fails
```hcl
# Verify domain resolution before creating alert
data "shodan_domain" "verify" {
  domain = "example.com"
}

resource "shodan_domain" "monitoring" {
  domain = data.shodan_domain.verify.domain
  # ...
}
```

**Solution**: Use the data source to verify domain information before creating alerts.

### Debugging

#### Enable Debug Logging
```hcl
provider "shodan" {
  api_key = var.shodan_api_key
  # Enable debug logging
  # debug = true  # Uncomment for debugging
}
```

#### Check Domain Information
```hcl
# Always check domain information first
data "shodan_domain" "debug" {
  domain = "example.com"
}

output "debug_info" {
  value = data.shodan_domain.debug
}
```

## Monitoring and Maintenance

### Regular Checks
- Monitor alert status in Shodan dashboard
- Review trigger effectiveness
- Update domain lists as needed
- Check for new security threats

### Alert Review
- Review false positives
- Adjust trigger sensitivity
- Update notifier configurations
- Archive resolved alerts

### Performance Optimization
- Use appropriate request intervals
- Batch domain operations
- Monitor API usage
- Optimize trigger combinations

## Related Resources

- [`shodan_domain` resource](../resources/shodan_domain.md) - Complete resource documentation
- [`shodan_domain` data source](../data-sources/shodan_domain.md) - Data source documentation
- [`shodan_alert` resource](../resources/shodan_alert.md) - Network monitoring alerts
- [Examples](../../examples/domain_monitoring.tf) - Complete working examples

