# Contributing to go-onvif

Thank you for your interest in contributing to go-onvif! 🎉

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please be respectful and considerate in all interactions.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Code samples
- Your environment (Go version, OS, camera model)
- Error messages or logs

### Suggesting Features

Feature requests are welcome! Please:

- Use a clear, descriptive title
- Provide detailed description of the proposed feature
- Explain the use case and benefits
- Consider if the feature fits the project scope

### Camera Compatibility Reports

Help us maintain compatibility information:

- Report both working and non-working cameras
- Include manufacturer, model, and firmware version
- Run `onvif-diagnostics` and share the output
- Note any special configuration needed

### Pull Requests

#### Before Submitting

1. Check if there's an existing PR for the same change
2. For major changes, open an issue first to discuss
3. Ensure your code follows the project style
4. Add tests for new functionality
5. Update documentation as needed

#### Submission Process

1. **Fork** the repository
2. **Create** a feature branch from `main`:
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make** your changes:
   - Write clear, descriptive commit messages
   - Follow Go best practices and idioms
   - Add comments for complex logic
   - Include tests

4. **Test** your changes:
   ```bash
   make test
   make lint
   ```

5. **Commit** using conventional commits:
   ```bash
   git commit -m "feat: add GetAnalyticsConfigurations support"
   git commit -m "fix: correct PTZ coordinate calculation"
   git commit -m "docs: update README with new examples"
   ```

6. **Push** to your fork:
   ```bash
   git push origin feature/amazing-feature
   ```

7. **Open** a Pull Request with:
   - Clear title and description
   - Reference related issues
   - List of changes made
   - Testing performed

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, for Makefile targets)
- golangci-lint for linting

### Clone and Build

```bash
git clone https://github.com/0x524a/onvif-go.git
cd go-onvif
go build ./...
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./discovery/...
```

### Linting

```bash
make lint
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions focused and small
- Write self-documenting code

### Naming Conventions

- Use descriptive variable names
- Follow Go naming conventions (camelCase for private, PascalCase for public)
- Avoid abbreviations unless widely understood

### Error Handling

- Always check errors
- Provide context in error messages
- Use `fmt.Errorf` with `%w` for error wrapping

### Documentation

- Add GoDoc comments for all exported types and functions
- Include usage examples for complex features
- Update README.md when adding new features

### Testing

- Write table-driven tests when applicable
- Test both success and failure cases
- Mock external dependencies
- Aim for >80% coverage for new code

### Example Test

```go
func TestGetDeviceInformation(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*testing.T) *Client
        want    *DeviceInformation
        wantErr bool
    }{
        {
            name: "success",
            setup: func(t *testing.T) *Client {
                // Setup mock
            },
            want: &DeviceInformation{
                Manufacturer: "Test",
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := tt.setup(t)
            got, err := client.GetDeviceInformation(context.Background())
            
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Commit Message Guidelines

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions or modifications
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

Examples:
```
feat: add support for Event service
fix: correct PTZ velocity calculation in ContinuousMove
docs: add examples for imaging settings
test: add integration tests for Hikvision cameras
```

## Project Structure

```
go-onvif/
├── client.go           # Main ONVIF client
├── types.go            # ONVIF type definitions
├── device.go           # Device service
├── media.go            # Media service
├── ptz.go              # PTZ service
├── imaging.go          # Imaging service
├── soap/               # SOAP client
├── discovery/          # WS-Discovery
├── server/             # ONVIF server implementation
├── testing/            # Test utilities
├── testdata/           # Test fixtures
├── cmd/                # Command-line tools
└── examples/           # Usage examples
```

## Adding New Features

### Client Features

1. Add method to appropriate service file (device.go, media.go, etc.)
2. Define request/response types in types.go
3. Add tests
4. Update documentation
5. Add example if useful

### Server Features

1. Add handler to server service file
2. Define request/response types
3. Register handler in server.go
4. Add tests
5. Update server documentation

## Review Process

1. Automated checks run on all PRs (tests, linting)
2. Maintainers review code and provide feedback
3. Address review comments
4. Once approved, PR will be merged

## Getting Help

- 💬 [GitHub Discussions](https://github.com/0x524a/onvif-go/discussions) - Ask questions
- 🐛 [GitHub Issues](https://github.com/0x524a/onvif-go/issues) - Report bugs
- 📖 [Documentation](https://pkg.go.dev/github.com/0x524a/onvif-go) - Read the docs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to go-onvif! Your efforts help make ONVIF integration better for everyone. 🚀
