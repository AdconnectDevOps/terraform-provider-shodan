---
page_title: "Data Source: shodan_domain"
description: |-
  Retrieve domain information from Shodan including subdomains and DNS records.
---

# Data Source: shodan_domain

The `shodan_domain` data source allows you to retrieve comprehensive information about a domain from Shodan, including subdomains, DNS records, and associated metadata.

## Example Usage

```hcl
# Get information about a domain
data "shodan_domain" "example" {
  domain = "example.com"
}

# Use the domain information
output "domain_info" {
  value = {
    domain     = data.shodan_domain.example.domain
    tags       = data.shodan_domain.example.tags
    subdomains = data.shodan_domain.example.subdomains
    dns_records_count = length(data.shodan_domain.example.data)
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain name to lookup (e.g., 'example.com').

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `domain` - The domain name that was looked up.
* `tags` - Tags associated with the domain (e.g., "dmarc", "ipv6", "spf").
* `subdomains` - List of subdomains found for the domain.
* `data` - DNS records and other data for the domain. Each record contains:
  * `subdomain` - The subdomain name (empty string for root domain).
  * `type` - The DNS record type (A, AAAA, MX, NS, SOA, TXT, etc.).
  * `value` - The value of the DNS record.
  * `last_seen` - When this record was last seen by Shodan.
* `more` - Whether there are more results available.

## Example Output

```hcl
data "shodan_domain" "example" {
  domain = "example.com"
}

output "domain_details" {
  value = {
    domain     = data.shodan_domain.example.domain
    tags       = data.shodan_domain.example.tags
    subdomains = data.shodan_domain.example.subdomains
    dns_records = data.shodan_domain.example.data
  }
}
```

Example output:
```json
{
  "domain": "example.com",
  "tags": ["dmarc", "ipv6", "spf"],
  "subdomains": ["_dmarc", "k2._domainkey", "www"],
  "dns_records": [
    {
      "subdomain": "",
      "type": "A",
      "value": "93.184.216.34",
      "last_seen": "2025-08-28T12:00:00.000000"
    },
    {
      "subdomain": "",
      "type": "AAAA",
      "value": "2606:2800:220:1:248:1893:25c8:1946",
      "last_seen": "2025-08-28T12:00:00.000000"
    }
  ]
}
```

## Use Cases

### Security Research
```hcl
# Analyze domain infrastructure for security assessment
data "shodan_domain" "target" {
  domain = "company.com"
}

output "security_analysis" {
  value = {
    has_ipv6 = contains(data.shodan_domain.target.tags, "ipv6")
    has_dmarc = contains(data.shodan_domain.target.tags, "dmarc")
    subdomain_count = length(data.shodan_domain.target.subdomains)
    dns_record_count = length(data.shodan_domain.target.data)
  }
}
```

### Infrastructure Monitoring
```hcl
# Monitor domain changes over time
data "shodan_domain" "monitored" {
  domain = "critical-service.com"
}

output "infrastructure_status" {
  value = {
    domain = data.shodan_domain.monitored.domain
    active_subdomains = data.shodan_domain.monitored.subdomains
    total_records = length(data.shodan_domain.monitored.data)
  }
}
```

### Compliance Checking
```hcl
# Check domain compliance requirements
data "shodan_domain" "compliance" {
  domain = "regulated-company.com"
}

output "compliance_status" {
  value = {
    has_dmarc = contains(data.shodan_domain.compliance.tags, "dmarc")
    has_spf = contains(data.shodan_domain.compliance.tags, "spf")
    has_ipv6 = contains(data.shodan_domain.compliance.tags, "ipv6")
  }
}
```

## Notes

- **API Credits**: Each domain lookup consumes 1 Shodan API query credit.
- **Rate Limiting**: The provider automatically implements rate limiting to comply with Shodan's API requirements.
- **Data Freshness**: The data represents Shodan's current view of the domain infrastructure.
- **Comprehensive Coverage**: Includes all DNS record types and subdomains discovered by Shodan.

## Related Resources

- [`shodan_domain` resource](../resources/shodan_domain.md) - Monitor a domain for security threats
- [`shodan_alert` resource](../resources/shodan_alert.md) - Create network monitoring alerts
- [`shodan_alert` data source](../data-sources/shodan_alert.md) - Retrieve alert information

