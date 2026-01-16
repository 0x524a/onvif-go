# CI/CD Documentation

## Overview

The ONVIF Go library uses GitHub Actions for continuous integration and deployment. All workflows are located in `.github/workflows/`.

## Workflow Summary

| Workflow | Purpose | Triggers | Status |
|----------|---------|----------|--------|
| **CI** | Main CI pipeline | Push/PR to main branches | ✅ Active |
| **Test** | Extended testing | Manual/Weekly/Code changes | ✅ Active |
| **Coverage** | Coverage analysis | After CI success | ✅ Active |
| **Release** | Create releases | Tags/Manual | ✅ Active |
| **Lint** | Code linting | Push/PR | ✅ Active |
| **Security** | Security scanning | Push/PR/Weekly | ✅ Active |
| **Docs** | Documentation checks | Docs changes | ✅ Active |
| **Dependency Review** | Dependency security | PRs | ✅ Active |

## Main CI Workflow

The **CI** workflow (`ci.yml`) is the primary workflow that runs on every push and pull request.

### Jobs

1. **validate** - Quick validation (5-10 minutes)
   - Code formatting check
   - `go vet`
   - Linting with golangci-lint

2. **test** - Primary testing (10-15 minutes)
   - Runs on Go 1.23
   - Race detector enabled
   - Coverage report generation
   - Uploads to Codecov

3. **test-matrix** - Multi-platform testing (20-30 minutes)
   - Tests on Go 1.21, 1.22, 1.23
   - Tests on Linux, macOS, Windows
   - Parallel execution

4. **build** - Build verification (5-10 minutes)
   - Builds all packages
   - Builds all examples
   - Builds all CLI tools

5. **sonarcloud** - Code quality (10-15 minutes)
   - Only on master/main
   - Requires SONAR_TOKEN secret

### Performance

- **Total CI time**: ~40-60 minutes (parallel jobs)
- **Fast feedback**: Validation job fails fast on formatting/lint issues
- **Caching**: Go modules and build cache for faster runs

## Release Workflow

The **Release** workflow (`release.yml`) creates GitHub releases with binaries for all platforms.

### Supported Platforms

- **Linux**: amd64, arm64, arm (v7)
- **Windows**: amd64, arm64
- **macOS**: amd64, arm64

### Release Process

1. **Tag creation**: Push a tag like `v1.2.3`
2. **Build**: Automatically builds for all platforms
3. **Archive**: Creates `.tar.gz` (Linux/macOS) and `.zip` (Windows)
4. **Checksums**: Generates SHA256 checksums
5. **Release**: Creates GitHub release with all artifacts
6. **Docker**: Builds and pushes multi-arch Docker image to GHCR

### Manual Release

You can also trigger a release manually:
1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Enter version (e.g., `v1.2.3`)

## Security Workflow

The **Security** workflow (`security.yml`) scans for vulnerabilities.

### Tools

- **gosec**: Security scanner for Go code
- **govulncheck**: Vulnerability checker for dependencies

### Schedule

Runs weekly on Sundays to catch new vulnerabilities.

## Coverage

Coverage is tracked and reported to Codecov. The coverage workflow provides detailed analysis:

- Total coverage percentage
- Coverage by package
- Coverage trends over time

### Coverage Threshold

Minimum coverage threshold: **50%**

## Required Secrets

### Optional Secrets

- `CODECOV_TOKEN` - For Codecov integration
- `SONAR_TOKEN` - For SonarCloud integration
- `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` - For Docker Hub

## Workflow Status Badges

Add these badges to your README:

```markdown
![CI](https://github.com/0x524a/onvif-go/workflows/CI/badge.svg)
![Test](https://github.com/0x524a/onvif-go/workflows/Extended%20Tests/badge.svg)
![Release](https://github.com/0x524a/onvif-go/workflows/Release/badge.svg)
```

## Best Practices

1. **Always run CI locally first**: `make check test`
2. **Keep workflows fast**: Use caching and parallel jobs
3. **Fail fast**: Validation job catches issues early
4. **Test before release**: All tests must pass before tagging
5. **Review security scans**: Check security workflow results

## Troubleshooting

### CI Fails on Formatting

```bash
# Fix formatting
make fmt

# Or manually
gofmt -w .
```

### CI Fails on Linting

```bash
# Run linter locally
make lint

# Or manually
golangci-lint run ./...
```

### Tests Fail Locally but Pass in CI

- Check Go version: CI uses Go 1.23
- Check race detector: CI runs with `-race`
- Check environment differences

### Release Fails

- Ensure tag format: `v1.2.3` (not `1.2.3`)
- Check permissions: Need `contents: write`
- Verify all tests pass before tagging

## Workflow Files

All workflow files are in `.github/workflows/`:

- `ci.yml` - Main CI pipeline
- `test.yml` - Extended tests
- `coverage.yml` - Coverage analysis
- `release.yml` - Release automation
- `lint.yml` - Linting
- `security.yml` - Security scanning
- `docs.yml` - Documentation checks
- `dependency-review.yml` - Dependency review

## See Also

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow README](../.github/workflows/README.md)
- [Makefile](../Makefile) - Local development commands

---

*Last Updated: December 2, 2025*

