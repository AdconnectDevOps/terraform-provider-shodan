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
