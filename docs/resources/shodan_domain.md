---
page_title: "Resource: shodan_domain"
description: |-
  Monitor a domain for security threats using Shodan alerts with automatic IP resolution.
---

# Resource: shodan_domain

The `shodan_domain` resource allows you to monitor a domain for security threats by automatically resolving the domain to IP addresses and creating Shodan alerts for those IPs.

## Example Usage

### Basic Domain Monitoring
```hcl
# Monitor a domain with default settings
resource "shodan_domain" "basic_monitoring" {
  domain = "example.com"
}
```

### Custom Domain Monitoring
```hcl
# Monitor a domain with custom configuration
resource "shodan_domain" "custom_monitoring" {
  domain      = "company.com"
  name        = "Company Domain Security Monitoring"
  description = "Monitor company.com for security threats and vulnerabilities"
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

### Comprehensive Security Monitoring
```hcl
# Monitor a domain with all available security triggers
resource "shodan_domain" "comprehensive" {
  domain      = "critical-service.com"
  name        = "Critical Service Security Monitoring"
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

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain name to monitor (e.g., 'example.com').
* `name` - (Optional) Optional custom name for the alert. If not provided, will use `__domain: {domain}` format.
* `description` - (Optional) Optional description of the domain monitoring alert.
* `enabled` - (Optional) Whether the domain monitoring alert is enabled. Defaults to `true`.
* `triggers` - (Optional) List of trigger rules to enable for domain monitoring.
* `notifiers` - (Optional) List of notifier IDs to associate with the domain alert.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The unique identifier for the Shodan domain alert.
* `created_at` - The timestamp when the domain alert was created.

## How It Works

### 1. Domain Resolution
The provider automatically resolves the domain to IP addresses using Shodan's DNS API:
- Queries Shodan for all DNS records (A, AAAA) associated with the domain
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

## Use Cases

### Company Domain Monitoring
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
  for_each = toset(["customer1.com", "customer2.com", "customer3.com"])
  
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

## Available Triggers

The following trigger rules are available for domain monitoring:

- `malware` - Malware detection
- `vulnerable` - Vulnerable services
- `vulnerable_unverified` - Unverified vulnerabilities
- `new_service` - New service detection
- `open_database` - Open database detection
- `ssl_expired` - Expired SSL certificates
- `iot` - Internet of Things devices
- `end_of_life` - End-of-life software
- `industrial_control_system` - Industrial control systems
- `internet_scanner` - Internet scanning activity
- `uncommon` - Uncommon services
- `uncommon_plus` - Extended uncommon service detection

## Naming Convention

Domain alerts use a special naming convention:
- **Default**: `__domain: {domain}` (e.g., `__domain: example.com`)
- **Custom**: `__domain: {domain} ({custom_name})` (e.g., `__domain: example.com (Custom Name)`)

This naming convention helps identify domain-based alerts in your Shodan dashboard.

## State Management

The provider handles domain changes automatically:
- **Domain Change**: If the domain changes, the provider recreates the alert with new IP addresses
- **IP Updates**: Automatically adapts to IP address changes for the same domain
- **Resource Cleanup**: Properly deletes old alerts when recreating

## Notes

- **API Credits**: Domain resolution consumes 1 Shodan API query credit per domain.
- **Rate Limiting**: The provider automatically implements rate limiting to comply with Shodan's API requirements.
- **IP Discovery**: Automatically discovers all IP addresses associated with a domain.
- **Dynamic Monitoring**: Adapts to changes in domain infrastructure automatically.
- **Comprehensive Coverage**: Monitors all IP addresses for comprehensive security coverage.

## Related Resources

- [`shodan_domain` data source](../data-sources/shodan_domain.md) - Retrieve domain information
- [`shodan_alert` resource](../resources/shodan_alert.md) - Create network monitoring alerts
- [`shodan_alert` data source](../data-sources/shodan_alert.md) - Retrieve alert information

## Import

Domain monitoring alerts can be imported using their alert ID:

```bash
terraform import shodan_domain.example BVJ6BXDDODSKP9WZ
```
