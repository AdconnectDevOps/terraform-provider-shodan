---
page_title: "shodan_alert"
description: "Manages a Shodan network security alert"
---

# shodan_alert Resource

The `shodan_alert` resource allows you to manage Shodan network alerts and monitoring configurations. This resource enables you to create, update, and delete security monitoring alerts for your networks.

## Example Usage

### Basic Alert

```hcl
resource "shodan_alert" "basic_monitoring" {
  name        = "basic-security-monitoring"
  network     = ["192.168.1.0/24"]
  description = "Basic security monitoring for home network"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

### Advanced Alert with Multiple Networks

```hcl
resource "shodan_alert" "advanced_monitoring" {
  name        = "advanced-security-monitoring"
  network     = ["192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"]
  description = "Comprehensive security monitoring for multiple networks"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "open_database",
    "ssl_expired",
    "iot",
    "industrial_control_system"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["slack-notifier-id"]
  
  tags = ["production", "security", "monitoring"]
}
```

### Critical Infrastructure Monitoring

```hcl
resource "shodan_alert" "critical_infrastructure" {
  name        = "critical-infrastructure-monitoring"
  network     = ["203.0.113.1/32", "198.51.100.1/32"]
  description = "High-priority monitoring for critical systems"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["critical-alerts", "incident-response"]
  
  tags = ["critical", "production", "infrastructure"]
}
```

## Argument Reference

The following arguments are supported:

*   `name` (Required, String) - The name of the Shodan alert. Must be unique within your account.

*   `network` (Required, List of String) - The IP network range(s) to monitor. Can be:
    - Single IP: `["192.168.1.1/32"]`
    - Subnet: `["192.168.1.0/24"]`
    - Multiple networks: `["192.168.1.0/24", "10.0.0.0/8"]`
    - Mixed types: `["192.168.1.0/24", "203.0.113.1/32"]`

*   `description` (Optional, String) - A description of the alert and what it monitors.

*   `tags` (Optional, List of String) - Tags to associate with the alert for organization and filtering.

*   `enabled` (Optional, Bool) - Whether the alert is enabled and actively monitoring. Defaults to `true`.

*   `triggers` (Optional, List of String) - List of trigger rules to enable. Available triggers include:
    - `malware` - Malware detected
    - `vulnerable` - Vulnerable service detected
    - `new_service` - New service detected
    - `open_database` - Open database detected
    - `ssl_expired` - SSL certificate expired
    - `iot` - Internet of Things device
    - `industrial_control_system` - Industrial control system
    - `end_of_life` - End of life software
    - `internet_scanner` - Internet scanner detected
    - `uncommon` - Uncommon service
    - `uncommon_plus` - Extended uncommon detection
    - `vulnerable_unverified` - Unverified vulnerable service

*   `notifiers` (Optional, List of String) - List of notifier IDs to associate. Use `["default"]` for email notifications.

*   `slack_notifications` (Optional, List of String) - List of Slack notifier IDs to send notifications to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

*   `id` (String) - The unique identifier for the Shodan alert.

*   `created_at` (String) - The timestamp when the alert was created.

## Import

Shodan alerts can be imported using their ID:

```bash
terraform import shodan_alert.example_alert alert-id-here
```

## Available Trigger Rules

The following trigger rules are available for Shodan alerts:

| Trigger | Description |
|---------|-------------|
| `end_of_life` | End of life software detected |
| `industrial_control_system` | Industrial control system detected |
| `internet_scanner` | Internet scanner detected |
| `iot` | Internet of Things device detected |
| `malware` | Malware detected |
| `new_service` | New service detected |
| `open_database` | Open database detected |
| `ssl_expired` | SSL certificate expired |
| `uncommon` | Uncommon service detected |
| `uncommon_plus` | Extended uncommon service detection |
| `vulnerable` | Vulnerable service detected |
| `vulnerable_unverified` | Unverified vulnerable service |

## Available Notifiers

- `default` - Default email notifier (configured in Shodan account)
- `slack` - Slack integration (configured in Shodan account)

### Getting Slack Notifier IDs

To use Slack notifications, you need to get your Slack notifier ID from your Shodan account:

1. **Go to your Shodan account settings**
2. **Navigate to "Notifications" or "Integrations"**
3. **Find your Slack integration** and note the notifier ID
4. **Use this ID in your Terraform configuration**:

```hcl
variable "slack_notifier_ids" {
  description = "List of Slack notifier IDs from your Shodan account"
  type        = list(string)
  default     = ["xxxxxxxxxxx"]  # Your actual Slack notifier ID
  sensitive   = true
}

resource "shodan_alert" "example" {
  # ... other configuration ...
  
  slack_notifications = var.slack_notifier_ids
}
```

## Best Practices

### Network Configuration

- **Use specific IPs** for critical systems: `["203.0.113.1/32"]`
- **Use subnets** for general monitoring: `["192.168.1.0/24"]`
- **Combine multiple networks** in single alerts for efficiency
- **Avoid overly broad ranges** unless necessary

### Trigger Selection

- **Start with essential triggers**: `malware`, `vulnerable`, `new_service`
- **Add specialized triggers** based on your infrastructure
- **Use `industrial_control_system`** for OT environments
- **Enable `ssl_expired`** for web-facing services

### Notification Strategy

- **Use email** for general monitoring
- **Add Slack** for immediate team awareness
- **Create escalation paths** for critical alerts
- **Test notifications** after setup

## Common Use Cases

### Home Network Security

```hcl
resource "shodan_alert" "home_security" {
  name        = "home-network-monitoring"
  network     = ["192.168.1.0/24"]
  description = "Monitor home network for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

### Business Infrastructure

```hcl
resource "shodan_alert" "business_infra" {
  name        = "business-infrastructure"
  network     = ["10.0.0.0/8", "172.16.0.0/12"]
  description = "Monitor business infrastructure for threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "industrial_control_system",
    "open_database"
  ]
  
  notifiers = ["default"]
}
```

### Critical Systems

```hcl
resource "shodan_alert" "critical_systems" {
  name        = "critical-systems-monitoring"
  network     = ["203.0.113.1/32", "198.51.100.1/32"]
  description = "High-priority monitoring for critical systems"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
}
```
