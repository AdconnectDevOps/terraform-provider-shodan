---
page_title: "Shodan Provider"
description: |-
  The Shodan provider is used to interact with Shodan's network security monitoring API.
  The provider needs to be configured with the proper credentials before it can be used.
---

# Shodan Provider

The Terraform Provider for Shodan allows you to manage Shodan network alerts and monitoring configurations using Terraform. This provider enables infrastructure as code for security monitoring, letting you programmatically create, update, and manage Shodan alerts.

Use the navigation to the left to read about the available resources and data sources.

## Example Usage

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

resource "shodan_alert" "security_monitoring" {
  name        = "comprehensive-security-monitoring"
  network     = ["172.16.0.0/12"]
  description = "Comprehensive security monitoring for internal network"
  
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
```

## Authentication

The Shodan provider requires an API key to authenticate with Shodan's services. You can provide the API key via the `api_key` argument in the provider configuration block, or via the `SHODAN_API_KEY` environment variable.

```hcl
provider "shodan" {
  api_key = "your-shodan-api-key"
}
```

## Rate Limiting

The provider automatically implements request spacing to ensure compliance with Shodan's API requirements. You can configure the request interval via the `request_interval` provider attribute:

```hcl
provider "shodan" {
  api_key = var.shodan_api_key
  request_interval = 5  # 5 seconds between requests
}
```

## Features

- **Multiple IP Support**: Monitor multiple networks with single alerts
- **Built-in Rate Limiting**: Automatic API rate limiting to prevent hitting Shodan limits
- **Slack Integration**: Direct Slack notifications for security alerts
- **Comprehensive Triggers**: Support for all Shodan alert trigger types
- **Terraform 1.0+ Compatible**: Built with the latest Terraform plugin framework

## Getting Started

1. **Install the provider** by adding it to your Terraform configuration
2. **Configure authentication** with your Shodan API key
3. **Create your first alert** using the `shodan_alert` resource
4. **Monitor your networks** for security threats

## Support

- **Documentation**: [GitHub Repository](https://github.com/AdconnectDevOps/terraform-provider-shodan)
- **Issues**: [GitHub Issues](https://github.com/AdconnectDevOps/terraform-provider-shodan/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AdconnectDevOps/terraform-provider-shodan/discussions)

## License

This project is licensed under the MIT License.
