variable "shodan_api_key" {
  description = "Shodan API key for authentication"
  type        = string
  sensitive   = true
}

variable "slack_notifications" {
  description = "List of Slack notifier IDs from your Shodan account settings"
  type        = list(string)
  default     = []
  sensitive   = true
}

variable "rate_limit" {
  description = "Rate limit for API requests in requests per second"
  type        = number
  default     = 2
}
