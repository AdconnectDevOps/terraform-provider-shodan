# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup
- GitHub Actions workflows for CI/CD
- Comprehensive documentation
- Contributing guidelines

## [0.1.0] - 2025-08-26

### Added
- **Initial Release**: First public release of the Terraform Provider for Shodan
- **Core Functionality**: Complete CRUD operations for Shodan network alerts
- **Network Monitoring**: Support for IP range monitoring with CIDR notation
- **Trigger Rules**: All 12 Shodan security trigger rules supported:
  - `end_of_life` - End of life software detection
  - `industrial_control_system` - Industrial control system detection
  - `internet_scanner` - Internet scanner detection
  - `iot` - Internet of Things device detection
  - `malware` - Malware detection
  - `new_service` - New service detection
  - `open_database` - Open database detection
  - `ssl_expired` - SSL certificate expiration
  - `uncommon` - Uncommon service detection
  - `uncommon_plus` - Extended uncommon service detection
  - `vulnerable` - Vulnerable service detection
  - `vulnerable_unverified` - Unverified vulnerable service detection

- **Dual Notifications**: Support for both email and Slack notifications
  - Email notifications via default notifier
  - Slack notifications for configured channels
- **Resource Management**: Full Terraform resource lifecycle management
  - Create, Read, Update, Delete operations
  - State management and tracking
  - Import existing resources
- **Data Source**: Read-only access to existing Shodan alerts
- **Multi-Platform Support**: Binary releases for:
  - Linux (AMD64, ARM64)
  - macOS (AMD64, ARM64)
  - Windows (AMD64)
- **Comprehensive Examples**: Working examples for common use cases
- **Professional Documentation**: Complete README with installation and usage instructions

### Technical Features
- **Go Implementation**: Built with Go 1.21+ and Terraform Plugin Framework
- **API Integration**: Full Shodan API integration for alert management
- **Error Handling**: Comprehensive error handling and user feedback
- **Logging**: Structured logging for debugging and monitoring
- **Testing**: Unit tests and integration tests
- **Build Automation**: Makefile for common development tasks

### Documentation
- **README.md**: Comprehensive usage guide with examples
- **Examples Directory**: Working Terraform configurations
- **API Documentation**: Complete endpoint documentation
- **Installation Guide**: Step-by-step installation instructions
- **Contributing Guide**: Guidelines for contributors
- **Changelog**: Version history and changes

---

## Release Notes

### Breaking Changes
- None in this initial release

### Known Issues
- Network filter updates require resource recreation (Shodan API limitation)
- Trigger/notifier removal not supported (Shodan API limitation)

### Dependencies
- Go >= 1.21
- Terraform >= 1.0
- Terraform Plugin Framework

### Supported Platforms
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)  
- Windows (AMD64)
