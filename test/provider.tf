terraform {
  required_providers {
    shodan = {
      source = "AdconnectDevOps/shodan"
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

variable "slack_notifications" {
  description = "List of Slack notifier IDs from your Shodan account settings"
  type        = list(string)
  default     = []
}

variable "request_interval" {
  description = "Request interval in seconds between API calls"
  type        = number
  default     = 1
}
