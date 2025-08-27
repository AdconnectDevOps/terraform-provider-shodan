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

variable "request_interval" {
  description = "Interval between API requests in seconds"
  type        = number
  default     = 2
}
