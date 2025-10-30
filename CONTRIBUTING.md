# Contributing to go-onvif

First off, thank you for considering contributing to go-onvif! It's people like you that make go-onvif such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps to reproduce the problem**
* **Provide specific examples to demonstrate the steps**
* **Describe the behavior you observed and what behavior you expected**
* **Include camera model and firmware version if relevant**
* **Include Go version and OS information**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a detailed description of the suggested enhancement**
* **Provide specific examples to demonstrate the enhancement**
* **Explain why this enhancement would be useful**

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the existing style
6. Issue that pull request!

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/go-onvif.git
cd go-onvif

# Add upstream remote
git remote add upstream https://github.com/0x524A/go-onvif.git

# Create a branch
git checkout -b feature/my-new-feature

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter (if installed)
golangci-lint run
```

## Coding Standards

* Follow standard Go conventions and idioms
* Use `gofmt` to format your code
* Write clear, self-documenting code with comments where necessary
* Add tests for new functionality
* Keep functions focused and modular
* Use meaningful variable and function names

## Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

Example:
```
Add support for Analytics service

- Implement GetAnalyticsConfiguration
- Add rule engine support
- Update documentation

Closes #123
```

## Testing

* Write unit tests for new functionality
* Ensure all tests pass before submitting PR
* Add integration tests for new ONVIF services
* Test with real cameras when possible

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestGetDeviceInformation
```

## Documentation

* Update README.md for user-facing changes
* Add godoc comments for exported types and functions
* Update examples if API changes
* Add changelog entry for significant changes

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

Thank you for contributing! 🎉
