# Test domain monitoring with Slack notifications
resource "shodan_domain" "slack_test" {
  domain      = "example.com"
  name        = "Slack Test Domain Monitoring"
  description = "Test Slack notifications for domain monitoring"
  enabled     = true
  
  triggers = [
    "malware",
    "vulnerable",
    "new_service"
  ]
  
  # Replace with your actual Slack notifier ID from Shodan
  slack_notifications = ["slack_12345"]
}

# Output the Slack notifications configuration
output "slack_notifications" {
  description = "Slack notifications configured for the domain"
  value       = shodan_domain.slack_test.slack_notifications
}
