---
page_title: "shodan_alert"
description: "Retrieves information about an existing Shodan alert"
---

# shodan_alert Data Source

The `shodan_alert` data source allows you to retrieve information about an existing Shodan alert. This is useful for referencing existing alerts in your Terraform configuration or for data analysis purposes.

## Example Usage

### Basic Data Source Usage

```hcl
data "shodan_alert" "existing_alert" {
  id = "existing-alert-id-here"
}

# Use the data in other resources
resource "shodan_alert" "related_alert" {
  name        = "related-to-${data.shodan_alert.existing_alert.name}"
  network     = data.shodan_alert.existing_alert.network
  description = "Related alert based on existing configuration"
  
  triggers = [
    "ai",
    "malware",
    "vulnerable"
  ]
  
  notifiers = ["default"]
}
```

### Reference Existing Alert Configuration

```hcl
data "shodan_alert" "production_alert" {
  id = "prod-security-monitoring-id"
}

# Create a staging alert with similar configuration
resource "shodan_alert" "staging_alert" {
  name        = "staging-${data.shodan_alert.production_alert.name}"
  network     = ["10.1.0.0/16"]  # Staging network
  description = "Staging version of production monitoring"
  
  triggers = data.shodan_alert.production_alert.triggers
  notifiers = data.shodan_alert.production_alert.notifiers
  
  tags = concat(data.shodan_alert.production_alert.tags, ["staging"])
}
```

### Output Alert Information

```hcl
data "shodan_alert" "monitoring_alert" {
  id = "network-monitoring-id"
}

output "alert_details" {
  description = "Details of the monitoring alert"
  value = {
    name        = data.shodan_alert.monitoring_alert.name
    description = data.shodan_alert.monitoring_alert.description
    networks    = data.shodan_alert.monitoring_alert.network
    triggers    = data.shodan_alert.monitoring_alert.triggers
    created_at  = data.shodan_alert.monitoring_alert.created_at
  }
}
```

## Argument Reference

The following argument is supported:

*   `id` (Required, String) - The unique identifier of the Shodan alert to retrieve.

## Attribute Reference

The following attributes are exported:

*   `id` (String) - The unique identifier for the Shodan alert.

*   `name` (String) - The name of the Shodan alert.

*   `network` (List of String) - The IP network range(s) being monitored.

*   `description` (String) - The description of the alert.

*   `tags` (List of String) - Tags associated with the alert.

*   `enabled` (Bool) - Whether the alert is enabled and actively monitoring.

*   `triggers` (List of String) - List of trigger rules that are enabled.

*   `notifiers` (List of String) - List of notifier IDs associated with the alert.

*   `slack_notifications` (List of String) - List of Slack notifier IDs for notifications.

*   `created_at` (String) - The timestamp when the alert was created.

## Use Cases

### Configuration Replication

Use existing alerts as templates for new configurations:

```hcl
data "shodan_alert" "template" {
  id = "template-alert-id"
}

resource "shodan_alert" "new_monitoring" {
  name        = "new-${data.shodan_alert.template.name}"
  network     = ["192.168.2.0/24"]  # New network
  description = data.shodan_alert.template.description
  
  triggers = data.shodan_alert.template.triggers
  notifiers = data.shodan_alert.template.notifiers
  
  tags = concat(data.shodan_alert.template.tags, ["new"])
}
```

### Environment Parity

Ensure staging and production have similar monitoring:

```hcl
data "shodan_alert" "production" {
  id = "prod-security-id"
}

resource "shodan_alert" "staging" {
  name        = "staging-${data.shodan_alert.production.name}"
  network     = ["10.1.0.0/16"]  # Staging network
  description = data.shodan_alert.production.description
  
  triggers = data.shodan_alert.production.triggers
  notifiers = data.shodan_alert.production.notifiers
  
  tags = concat(data.shodan_alert.production.tags, ["staging"])
}
```

### Monitoring and Reporting

Use data sources for monitoring and reporting purposes:

```hcl
data "shodan_alert" "all_alerts" {
  for_each = toset(["alert-1", "alert-2", "alert-3"])
  id       = each.value
}

output "monitoring_summary" {
  description = "Summary of all monitoring alerts"
  value = {
    total_alerts = length(data.shodan_alert.all_alerts)
    alert_names  = [for alert in data.shodan_alert.all_alerts : alert.name]
    networks     = flatten([for alert in data.shodan_alert.all_alerts : alert.network])
    triggers     = distinct(flatten([for alert in data.shodan_alert.all_alerts : alert.triggers]))
  }
}
```

## Best Practices

### Error Handling

Always handle cases where the alert might not exist:

```hcl
data "shodan_alert" "existing" {
  id = var.alert_id
}

locals {
  alert_exists = data.shodan_alert.existing.id != ""
  
  # Use default values if alert doesn't exist
  default_triggers = ["malware", "vulnerable"]
  alert_triggers   = alert_exists ? data.shodan_alert.existing.triggers : default_triggers
}
```

### Conditional Usage

Use data sources conditionally based on your configuration:

```hcl
variable "use_existing_alert" {
  description = "Whether to use an existing alert as template"
  type        = bool
  default     = false
}

variable "existing_alert_id" {
  description = "ID of existing alert to use as template"
  type        = string
  default     = ""
}

data "shodan_alert" "template" {
  count = var.use_existing_alert ? 1 : 0
  id    = var.existing_alert_id
}

resource "shodan_alert" "monitoring" {
  name        = "network-monitoring"
  network     = ["192.168.1.0/24"]
  description = var.use_existing_alert ? data.shodan_alert.template[0].description : "Default monitoring"
  
  triggers = var.use_existing_alert ? data.shodan_alert.template[0].triggers : ["malware", "vulnerable"]
  notifiers = ["default"]
}
```

## Limitations

- **Read-only**: Data sources are read-only and cannot modify existing alerts
- **Dependency**: The data source depends on the alert existing before Terraform runs
- **No validation**: Terraform cannot validate that the alert ID exists until runtime

## Related Resources

- [`shodan_alert` resource](../resources/shodan_alert.md) - Create and manage Shodan alerts
- [Provider configuration](../index.md) - Configure the Shodan provider
- [Examples](../../examples/) - Additional usage examples
