variable "shodan_api_key" {
  description = "Shodan API key for authentication"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name (e.g., production, staging, development)"
  type        = string
  default     = "production"
}

variable "team" {
  description = "Team name responsible for these assets"
  type        = string
  default     = "security"
}

variable "slack_notifications" {
  description = "List of Slack notifier IDs from your Shodan account settings"
  type        = list(string)
  default     = []
  sensitive   = true
}
