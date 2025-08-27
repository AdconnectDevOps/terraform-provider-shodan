---
page_title: "Getting Started with Shodan Provider"
description: "Learn how to get started with the Terraform Provider for Shodan"
---

# Getting Started with Shodan Provider

This guide will walk you through setting up and using the Terraform Provider for Shodan to monitor your networks for security threats.

## Prerequisites

Before you begin, ensure you have:

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 installed
- A [Shodan account](https://shodan.io) with an API key
- Access to the networks you want to monitor

## Step 1: Get Your Shodan API Key

1. **Visit [Shodan](https://shodan.io) and create an account**
2. **Log into your account**
3. **Go to Account Settings → API Key**
4. **Copy your API key** (it will look like: `xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`)

## Step 2: Install the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    shodan = {
      source = "AdconnectDevOps/shodan"
      version = "~> 0.1"
    }
  }
}
```

## Step 3: Configure the Provider

Create a `main.tf` file with your provider configuration:

```hcl
provider "shodan" {
  api_key = var.shodan_api_key
}

variable "shodan_api_key" {
  description = "Your Shodan API key"
  type        = string
  sensitive   = true
}
```

## Step 4: Create Your First Alert

Add a resource to monitor your network:

```hcl
resource "shodan_alert" "my_first_alert" {
  name        = "my-first-shodan-alert"
  network     = ["192.168.1.0/24"]
  description = "Monitoring my home network for security threats"
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  notifiers = ["default"]
}
```

## Step 5: Deploy Your Configuration

1. **Initialize Terraform:**
   ```bash
   terraform init
   ```

2. **Create a variables file:**
   ```bash
   # terraform.tfvars
   shodan_api_key = "your-actual-api-key-here"
   ```

3. **Plan your deployment:**
   ```bash
   terraform plan
   ```

4. **Apply your configuration:**
   ```bash
   terraform apply
   ```

## Step 6: Verify Your Alert

After successful deployment:

1. **Check the Terraform output** for your alert ID
2. **Visit [Shodan Dashboard](https://shodan.io)** to see your alert
3. **Test notifications** by checking your email

## Next Steps

Now that you have your first alert running:

- **Add more networks** to monitor
- **Configure additional triggers** for comprehensive monitoring
- **Set up Slack notifications** for immediate alerts
- **Explore advanced configurations** in the examples

## Common Issues and Solutions

### Provider not found
- Ensure you've run `terraform init`
- Check that the source path is correct: `AdconnectDevOps/shodan`

### Authentication errors
- Verify your Shodan API key is correct
- Check that your Shodan account is active

### Rate limiting
- Increase the `request_interval` in your provider configuration
- Check your Shodan API plan limits

## Additional Resources

- [Resource Documentation](../resources/shodan_alert.md) - Complete resource reference
- [Data Source Documentation](../data-sources/shodan_alert.md) - Query existing alerts
- [Examples](../../examples/) - Real-world configurations
- [GitHub Repository](https://github.com/AdconnectDevOps/terraform-provider-shodan) - Source code and issues

## Support

If you encounter issues:

1. **Check the [FAQ](../../faq.md)** for common solutions
2. **Search existing issues** on [GitHub](https://github.com/AdconnectDevOps/terraform-provider-shodan/issues)
3. **Create a new issue** with detailed error information
4. **Join discussions** on [GitHub Discussions](https://github.com/AdconnectDevOps/terraform-provider-shodan/discussions)

---

**Happy monitoring!** 🚀 Your networks are now protected with Shodan's powerful threat detection capabilities.
