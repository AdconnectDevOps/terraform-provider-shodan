# Contributing to Terraform Provider for Shodan

Thank you for your interest in contributing to the Terraform Provider for Shodan! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Git
- Make (optional)

### Development Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/terraform-provider-shodan.git
   cd terraform-provider-shodan
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/AdconnectDevOps/terraform-provider-shodan.git
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```

## 🔧 Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. **Make your changes** following the coding standards below
3. **Test your changes**:
   ```bash
   make test
   make build
   ```
4. **Commit your changes** with a descriptive message:
   ```bash
   git commit -m "feat: add new feature description"
   ```
5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
6. **Create a Pull Request** on GitHub

### Coding Standards

- **Go Code**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **Terraform**: Follow [HashiCorp's Terraform style conventions](https://www.terraform.io/docs/language/syntax/style.html)
- **Documentation**: Update README.md and add examples for new features
- **Tests**: Add tests for new functionality and ensure existing tests pass

### Commit Message Format

Use conventional commit format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(alert): add support for custom trigger rules
fix(client): resolve API timeout issues
docs(readme): update installation instructions
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestFunctionName
```

### Testing with Terraform

```bash
# Go to test directory
cd test

# Initialize Terraform
terraform init

# Plan changes
terraform plan

# Apply changes (use with caution)
terraform apply

# Destroy resources
terraform destroy
```

### Manual Testing

1. Set your Shodan API key in `test/terraform.tfvars`
2. Run the test configuration
3. Verify the alert is created in Shodan UI
4. Check that notifications are properly configured

## 📚 Documentation

### Updating Documentation

- **README.md**: Main documentation and usage examples
- **Examples**: Add working examples in the `examples/` directory
- **Code Comments**: Add clear comments for complex logic
- **API Documentation**: Document any new API endpoints or changes

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Keep examples up-to-date with code changes
- Use proper markdown formatting

## 🐛 Bug Reports

### Before Reporting

1. Check existing issues for similar problems
2. Ensure you're using the latest version
3. Verify your configuration is correct
4. Test with a minimal configuration

### Bug Report Template

```markdown
**Description**
Brief description of the issue

**Steps to Reproduce**
1. Step 1
2. Step 2
3. Step 3

**Expected Behavior**
What you expected to happen

**Actual Behavior**
What actually happened

**Environment**
- OS: [e.g., macOS 12.0]
- Go version: [e.g., 1.21.0]
- Terraform version: [e.g., 1.5.0]
- Provider version: [e.g., 0.1.0]

**Configuration**
```hcl
# Your Terraform configuration here
```

**Error Messages**
Any error messages or logs

**Additional Information**
Any other relevant information
```

## 💡 Feature Requests

### Before Requesting

1. Check if the feature already exists
2. Consider if it fits the project's scope
3. Think about implementation complexity
4. Consider backward compatibility

### Feature Request Template

```markdown
**Feature Description**
Clear description of the requested feature

**Use Case**
Why this feature is needed and how it would be used

**Proposed Implementation**
Optional: suggestions for how to implement

**Alternatives Considered**
Other approaches you've considered

**Additional Context**
Any other relevant information
```

## 🔄 Pull Request Process

### Before Submitting

1. **Ensure tests pass** locally
2. **Update documentation** if needed
3. **Add examples** for new features
4. **Follow coding standards**
5. **Test with real Shodan API** if possible

### Pull Request Template

```markdown
**Description**
Brief description of changes

**Type of Change**
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Other (please describe)

**Testing**
- [ ] Tests pass locally
- [ ] Manual testing completed
- [ ] Documentation updated

**Checklist**
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Code commented where needed
- [ ] Corresponding changes to documentation
- [ ] No breaking changes
```

## 📋 Review Process

### What We Look For

- **Functionality**: Does the code work as intended?
- **Quality**: Is the code well-written and maintainable?
- **Testing**: Are there adequate tests?
- **Documentation**: Is the documentation updated?
- **Compatibility**: Are there any breaking changes?

### Review Timeline

- Initial review: Within 1-2 business days
- Follow-up reviews: Within 1 business day
- Final approval: When all concerns are addressed

## 🏷️ Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Create release branch** from main
2. **Update version** in code and documentation
3. **Create release tag** on GitHub
4. **GitHub Actions** automatically build and release
5. **Update documentation** with new version

## 📞 Getting Help

### Communication Channels

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Pull Requests**: For code contributions

### Code of Conduct

- Be respectful and inclusive
- Focus on technical discussions
- Help others learn and contribute
- Follow the project's coding standards

## 🙏 Acknowledgments

Thank you for contributing to the Terraform Provider for Shodan! Your contributions help make this project better for everyone in the community.

---

**Note**: This contributing guide is a living document. Feel free to suggest improvements or ask questions about any part of the contribution process.
