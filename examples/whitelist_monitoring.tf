terraform {
  required_providers {
    shodan = {
      source  = "AdconnectDevOps/shodan"
      version = "~> 0.1"
    }
  }
}

provider "shodan" {
  api_key = var.shodan_api_key
}

# Some hosts have a known service on a non-standard port (e.g. a managed
# database engine that exposes both 3306 and a secondary port) that fires
# new_service every rescan. Whitelisting the specific (ip, port) silences
# the false positive without disabling new_service for the rest of the
# asset group — a genuinely new service on the same host (port 22, 80, …)
# would still alert.
resource "shodan_alert" "known_service_host" {
  name    = "known-service-host"
  network = ["203.0.113.10/32"]

  triggers = [
    "new_service",
    "open_database",
    "ssl_expired",
    "vulnerable",
  ]

  notifiers = ["default"]

  whitelist = {
    new_service = [
      "203.0.113.10:3307",
    ]
  }
}
