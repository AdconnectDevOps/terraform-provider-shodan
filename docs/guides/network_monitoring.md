---
page_title: "Network Monitoring Guide"
description: |-
  Learn how to use the Shodan provider to monitor networks for security threats.
---

# Network Monitoring Guide

This guide explains how to use the Shodan provider to monitor networks for security threats using Terraform. Network monitoring allows you to create comprehensive security alerts for IP ranges, subnets, and individual hosts.

## Overview

Network monitoring with Shodan provides several key benefits:

- **Comprehensive Coverage**: Monitor entire networks, subnets, or specific IP addresses
- **Real-time Alerts**: Get notified immediately when security threats are detected
- **Flexible Configuration**: Support for various network ranges and alert types
- **Integration Ready**: Built-in support for email and Slack notifications
- **Scalable**: Monitor from single hosts to large enterprise networks

## How Network Monitoring Works

### 1. Network Definition
The provider accepts various network formats:
- **Single IP**: `["192.168.1.1/32"]` - Monitor a specific host
- **Subnet**: `["192.168.1.0/24"]` - Monitor an entire subnet
- **Multiple Networks**: `["192.168.1.0/24", "10.0.0.0/8"]` - Monitor multiple ranges
- **Mixed Types**: Combine single IPs and subnets in one alert

### 2. Alert Creation
Creates a Shodan alert for the specified networks:
- Monitors all IP addresses within the defined ranges
- Applies configured security triggers
- Associates notification channels (email, Slack)

### 3. Threat Detection
Continuously monitors for various security threats:
- **Malware infections** and suspicious activity
- **Vulnerable services** and exploitable systems
- **Infrastructure changes** like new services or open databases
- **Compliance issues** such as expired SSL certificates
- **Unusual devices** including IoT and industrial control systems

## Basic Network Monitoring

### Simple Network Alert
```hcl
# Monitor a single network with basic triggers
resource "shodan_alert" "basic" {
  name        = "home-network-security"
  network     = ["192.168.1.0/24"]
  description = "Basic security monitoring for home network"
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

This creates a network monitoring alert with:
- Default email notifications
- Essential security triggers
- Simple network range monitoring

### Custom Network Alert
```hcl
# Monitor multiple networks with custom configuration
resource "shodan_alert" "custom" {
  name        = "company-networks"
  network     = ["192.168.1.0/24", "10.0.0.0/8"]
  description = "Monitor company networks for security threats"
  enabled     = true
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired",
    "open_database"
  ]
  
  notifiers = ["default"]
  tags = ["production", "security"]
}
```

## Advanced Network Monitoring

### Comprehensive Security Monitoring
```hcl
# Monitor critical infrastructure with all available triggers
resource "shodan_alert" "comprehensive" {
  name        = "critical-infrastructure"
  network     = ["203.0.113.0/24", "198.51.100.0/24"]
  description = "Comprehensive security monitoring for critical systems"
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
  slack_notifications = ["critical-alerts", "incident-response"]
  tags = ["critical", "production", "infrastructure"]
}
```

### Multi-Environment Monitoring
```hcl
# Monitor different environments with appropriate triggers
resource "shodan_alert" "production" {
  name        = "production-network"
  network     = ["10.1.0.0/16"]
  description = "Production environment monitoring"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["prod-alerts"]
  tags = ["production", "high-priority"]
}

resource "shodan_alert" "development" {
  name        = "development-network"
  network     = ["10.2.0.0/16"]
  description = "Development environment monitoring"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
  tags = ["development", "medium-priority"]
}
```

## Slack Notifications

### Basic Slack Integration
```hcl
# Enable Slack notifications for immediate team awareness
resource "shodan_alert" "slack_enabled" {
  name        = "slack-monitoring"
  network     = ["192.168.1.0/24"]
  description = "Network monitoring with Slack alerts"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  # Use your actual Slack notifier ID from Shodan
  slack_notifications = ["slack_12345"]
  
  # You can combine with regular notifiers
  notifiers = ["default"]
}
```

### Multiple Slack Channels
```hcl
# Send alerts to multiple Slack channels for different purposes
resource "shodan_alert" "multi_slack" {
  name        = "multi-channel-monitoring"
  network     = ["10.0.0.0/8"]
  description = "Network monitoring with multiple Slack channels"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  # Different channels for different alert types
  slack_notifications = [
    "general-alerts",      # General security alerts
    "critical-incidents",  # High-priority incidents
    "compliance-team"      # Compliance-related alerts
  ]
  
  notifiers = ["default"]
}
```

## Network Configuration Strategies

### Single Host Monitoring
```hcl
# Monitor critical individual hosts
resource "shodan_alert" "critical_hosts" {
  name        = "critical-hosts"
  network     = [
    "203.0.113.1/32",    # Web server
    "198.51.100.1/32",   # Database server
    "203.0.113.10/32"    # Mail server
  ]
  description = "Monitor critical production hosts"
  
  triggers = [
    "ai",
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["critical-hosts"]
  tags = ["critical", "single-host"]
}
```

### Subnet Monitoring
```hcl
# Monitor entire subnets for general security
resource "shodan_alert" "subnet_monitoring" {
  name        = "office-subnet"
  network     = ["192.168.100.0/24"]
  description = "Monitor office network subnet"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "iot"
  ]
  
  notifiers = ["default"]
  tags = ["office", "general"]
}
```

### Large Network Monitoring
```hcl
# Monitor large enterprise networks
resource "shodan_alert" "enterprise" {
  name        = "enterprise-network"
  network     = [
    "10.0.0.0/8",        # Main corporate network
    "172.16.0.0/12",     # Secondary network
    "192.168.0.0/16"     # Branch offices
  ]
  description = "Enterprise-wide network security monitoring"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "open_database",
    "ssl_expired",
    "industrial_control_system"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["enterprise-alerts", "security-team"]
  tags = ["enterprise", "large-scale"]
}
```

## Trigger Configuration

### Essential Triggers
Start with these core security triggers:
```hcl
triggers = [
  "ai",             # Detect AI-related services
  "malware",        # Detect malware infections
  "vulnerable",     # Find vulnerable services
  "new_service"     # Alert on new services
]
```

### Extended Security Monitoring
Add these for comprehensive coverage:
```hcl
triggers = [
  "malware",
  "vulnerable",
  "new_service",
  "open_database",      # Detect exposed databases
  "ssl_expired",        # SSL certificate issues
  "iot",                # Internet of Things devices
  "end_of_life"         # End-of-life software
]
```

### Specialized Monitoring
For specific environments, consider:
```hcl
# Industrial/OT environments
triggers = [
  "malware",
  "vulnerable",
  "industrial_control_system",  # SCADA/ICS systems
  "end_of_life"
]

# Web-facing services
triggers = [
  "malware",
  "vulnerable",
  "ssl_expired",
  "open_database"
]

# IoT environments
triggers = [
  "malware",
  "vulnerable",
  "iot",
  "uncommon"
]
```

## Best Practices

### Network Range Selection
- **Be specific**: Use `/32` for single hosts, `/24` for small subnets
- **Avoid overly broad ranges**: Large ranges can generate many alerts
- **Group logically**: Combine related networks in single alerts
- **Consider growth**: Plan for network expansion

### Trigger Selection
- **Start minimal**: Begin with essential triggers and expand
- **Match environment**: Use appropriate triggers for your infrastructure
- **Monitor compliance**: Enable SSL and database triggers for web services
- **Industrial focus**: Use ICS triggers for operational technology

### Notification Strategy
- **Email for general alerts**: Reliable but may be delayed
- **Slack for immediate response**: Real-time team awareness
- **Escalation paths**: Different channels for different severity levels
- **Testing**: Verify notifications work after setup

### Resource Organization
- **Use descriptive names**: Clear, meaningful alert names
- **Add descriptions**: Explain what each alert monitors
- **Apply tags**: Organize alerts by environment, priority, or purpose
- **Documentation**: Keep track of what each alert covers

## Common Use Cases

### Home Network Security
```hcl
resource "shodan_alert" "home" {
  name        = "home-network"
  network     = ["192.168.1.0/24"]
  description = "Monitor home network for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "iot"
  ]
  
  notifiers = ["default"]
  tags = ["home", "personal"]
}
```

### Small Business Monitoring
```hcl
resource "shodan_alert" "small_business" {
  name        = "business-network"
  network     = ["10.0.1.0/24", "10.0.2.0/24"]
  description = "Monitor business networks for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "ssl_expired",
    "open_database"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["business-alerts"]
  tags = ["business", "production"]
}
```

### Enterprise Security
```hcl
resource "shodan_alert" "enterprise_core" {
  name        = "enterprise-core"
  network     = ["10.1.0.0/16"]
  description = "Core enterprise network monitoring"
  
  triggers = [
    "malware",
    "vulnerable",
    "vulnerable_unverified",
    "new_service",
    "open_database",
    "ssl_expired",
    "industrial_control_system"
  ]
  
  notifiers = ["default"]
  slack_notifications = ["enterprise-core", "security-ops"]
  tags = ["enterprise", "core", "critical"]
}
```

## Troubleshooting

### Common Issues

**Alert Creation Fails**
- Verify API key is valid and has sufficient permissions
- Check network range format (CIDR notation)
- Ensure alert name is unique within your account

**No Notifications Received**
- Verify notifier IDs are correct
- Check Shodan account notification settings
- Test with simple configuration first

**Too Many Alerts**
- Review trigger selection - start with fewer triggers
- Check network range size - smaller ranges generate fewer alerts
- Use tags to organize and filter alerts

### Testing Your Configuration

1. **Start Simple**: Begin with basic configuration and minimal triggers
2. **Test Notifications**: Verify email and Slack notifications work
3. **Monitor Alerts**: Check Shodan dashboard for created alerts
4. **Iterate**: Add more triggers and networks gradually

### Getting Help

- **Check Shodan API documentation** for endpoint details
- **Verify account settings** for notification configuration
- **Review Terraform logs** for detailed error messages
- **Test with curl** to verify API connectivity

## Next Steps

After setting up network monitoring:

1. **Review and refine** your trigger configuration
2. **Set up escalation procedures** for different alert types
3. **Integrate with other tools** like SIEM or ticketing systems
4. **Monitor and adjust** based on alert volume and relevance
5. **Expand coverage** to additional networks as needed

Network monitoring with Shodan provides a solid foundation for security awareness. Start with essential monitoring and gradually expand based on your security requirements and infrastructure needs.
