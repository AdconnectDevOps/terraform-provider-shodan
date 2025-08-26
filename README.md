# Terraform Provider for Shodan

[![Go Report Card](https://goreportcard.com/badge/github.com/AdconnectDevOps/terraform-provider-shodan)](https://goreportcard.com/report/github.com/AdconnectDevOps/terraform-provider-shodan)
[![Release](https://img.shields.io/github/v/release/AdconnectDevOps/terraform-provider-shodan)](https://github.com/AdconnectDevOps/terraform-provider-shodan/releases)
[![License](https://img.shields.io/github/license/AdconnectDevOps/terraform-provider-shodan)](LICENSE)

A Terraform provider for managing Shodan network alerts and monitoring configurations. This provider allows you to programmatically create, manage, and monitor network security alerts using Shodan's powerful threat detection capabilities.

## 🚀 Features

- **Network Alert Management**: Create and manage Shodan network monitoring alerts
- **Dual Notifications**: Support for both email and Slack notifications
- **Comprehensive Trigger Rules**: All 12 Shodan security trigger rules supported
- **IP Range Monitoring**: Monitor specific IP ranges (CIDR notation supported)
- **State Management**: Full Terraform state tracking and lifecycle management
- **Multi-Platform**: Supports Linux, macOS, and Windows (AMD64 and ARM64)

## 📋 Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/dl/) >= 1.21 (for development)
- [Shodan API Key](https://account.shodan.io/)

## 🔧 Installation

### From GitHub Releases (Recommended)

1. **Download the appropriate binary** for your platform from the [releases page](https://github.com/AdconnectDevOps/terraform-provider-shodan/releases)

2. **Create the provider directory structure**:
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/AdconnectDevOps/shodan/0.1.0/darwin_arm64/
   ```

3. **Install the provider**:
   ```bash
   # For macOS ARM64 (Apple Silicon)
   cp terraform-provider-shodan_0.1.0_darwin_arm64 ~/.terraform.d/plugins/registry.terraform.io/AdconnectDevOps/shodan/0.1.0/darwin_arm64/terraform-provider-shodan
   
   # For macOS AMD64 (Intel)
   cp terraform-provider-shodan_0.1.0_darwin_amd64 ~/.terraform.d/plugins/registry.terraform.io/AdconnectDevOps/shodan/0.1.0/darwin_amd64/terraform-provider-shodan
   
   # For Linux AMD64
   cp terraform-provider-shodan_0.1.0_linux_amd64 ~/.terraform.d/plugins/registry.terraform.io/AdconnectDevOps/shodan/0.1.0/linux_amd64/terraform-provider-shodan
   
   # For Linux ARM64
   cp terraform-provider-shodan_0.1.0_linux_arm64 ~/.terraform.d/plugins/registry.terraform.io/AdconnectDevOps/shodan/0.1.0/linux_arm64/terraform-provider-shodan
   ```

### From Source

```bash
git clone https://github.com/AdconnectDevOps/terraform-provider-shodan.git
cd terraform-provider-shodan
make install
```

## 📖 Usage

### Provider Configuration

```hcl
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
```

### Finding Your Slack Notifier IDs

To configure Slack notifications, you need to get your Slack notifier IDs from your Shodan account:

1. **Log into your Shodan account** at [shodan.io](https://shodan.io)
2. **Go to Account Settings** → **Notifications** or **Integrations**
3. **Find your Slack integration** and note the notifier ID
4. **Use this ID in your Terraform configuration**:

```hcl
variable "slack_notifier_ids" {
  description = "List of Slack notifier IDs from your Shodan account"
  type        = list(string)
  default     = ["xxxxxxxxxxx"]  # Your actual Slack notifier ID
  sensitive   = true
}

# Example: Using your actual Slack notifier ID
slack_notifier_ids = ["xxxxxxxxxxx"]  # Replace with your actual ID
```

**💡 Pro Tip**: You can have multiple Slack notifiers for different channels or teams. Just add more IDs to the list:
```hcl
slack_notifier_ids = [
  "xxxxxxxxxxx",  # Main alerts channel
  "another-id-here",    # Secondary channel
  "third-id-here"       # Team-specific channel
]
```
```

### Basic Network Monitoring

```hcl
resource "shodan_alert" "basic_monitoring" {
  name    = "basic-network-monitoring"
  network = "10.0.0.0/8"
}
```

### Comprehensive Security Monitoring

```hcl
resource "shodan_alert" "security_monitoring" {
  name        = "comprehensive-security-monitoring"
  network     = "172.16.0.0/12"
  description = "Comprehensive security monitoring for internal network"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service",
    "open_database",
    "ssl_expired",
    "iot"
  ]
  
  notifiers = ["default"]  # Email notifications
}
```

### Slack Integration with Email Notifications

```hcl
resource "shodan_alert" "slack_monitoring" {
  name        = "slack-enabled-monitoring"
  network     = "192.168.1.0/24"
  description = "Network monitoring with Slack and email notifications"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]  # Email notifications
  
  slack_notifications = var.slack_notifier_ids  # Uses your Slack notifier IDs
}
```

**Example with actual Slack notifier ID:**
```hcl
resource "shodan_alert" "production_monitoring" {
  name        = "production-network-monitoring"
  network     = "10.0.0.0/8"
  description = "Production network monitoring with Slack alerts"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  slack_notifications = ["xxxxxxxxxxx"]  # Your actual Slack notifier ID
}
```

### Multiple Network Monitoring

```hcl
locals {
  networks = [
    {
      name        = "production"
      network     = "192.168.1.0/24"
      description = "Production network"
    },
    {
      name        = "staging"
      network     = "192.168.2.0/24"
      description = "Staging network"
    }
  ]
}

resource "shodan_alert" "network_alerts" {
  for_each = { for net in local.networks : net.name => net }
  
  name        = each.value.name
  network     = each.value.network
  description = each.value.description
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

## 📚 Resources

### `shodan_alert`

Manages a Shodan network alert for monitoring specific IP ranges.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | `string` | Yes | The name of the Shodan alert |
| `network` | `string` | Yes | The IP network range to monitor (e.g., '192.168.1.0/24') |
| `description` | `string` | No | A description of the alert |
| `tags` | `list(string)` | No | Tags to associate with the alert |
| `enabled` | `bool` | No | Whether the alert is enabled (default: true) |
| `triggers` | `list(string)` | No | List of trigger rules to enable |
| `notifiers` | `list(string)` | No | List of notifier IDs to associate |
| `slack_notifications` | `list(string)` | No | List of Slack channels to send notifications to |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | `string` | The unique identifier for the Shodan alert |
| `created_at` | `string` | The timestamp when the alert was created |

## 🔔 Available Trigger Rules

The following trigger rules are available for Shodan alerts:

- `end_of_life` - End of life software detected
- `industrial_control_system` - Industrial control system detected
- `internet_scanner` - Internet scanner detected
- `iot` - Internet of Things device detected
- `malware` - Malware detected
- `new_service` - New service detected
- `open_database` - Open database detected
- `ssl_expired` - SSL certificate expired
- `uncommon` - Uncommon service detected
- `uncommon_plus` - Extended uncommon service detection
- `vulnerable` - Vulnerable service detected
- `vulnerable_unverified` - Unverified vulnerable service

## 📧 Available Notifiers

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

## 🚀 Development

### Prerequisites

- Go >= 1.21
- Make (optional, for using Makefile)

### Building

```bash
# Build for current platform
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Install locally
make install

# Run tests
make test

# Format code
make fmt

# Run linter
make vet
```

### Project Structure

```
terraform-provider-shodan/
├── .github/              # GitHub Actions workflows
├── examples/             # Usage examples
├── test/                 # Test configurations
├── client.go             # Shodan API client
├── main.go               # Provider entry point
├── provider.go           # Provider configuration
├── resource_shodan_alert.go  # Alert resource
├── datasource_shodan_alert.go # Alert data source
├── go.mod                # Go module file
├── go.sum                # Go module checksums
├── Makefile              # Build automation
└── README.md             # This file
```

## 📝 Examples

See the [examples](./examples/) directory for complete working examples.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Shodan](https://shodan.io/) for providing the security monitoring API
- [HashiCorp](https://www.hashicorp.com/) for the Terraform plugin framework
- The open-source community for inspiration and support

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/AdconnectDevOps/terraform-provider-shodan/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AdconnectDevOps/terraform-provider-shodan/discussions)
- **Releases**: [GitHub Releases](https://github.com/AdconnectDevOps/terraform-provider-shodan/releases)

---

**Note**: This provider is not officially affiliated with Shodan. It's a community-driven project to bring Shodan's powerful security monitoring capabilities to Terraform workflows.
