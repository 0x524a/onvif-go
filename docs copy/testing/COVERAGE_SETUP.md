# Code Quality & Coverage Setup Guide

This guide explains how to set up CodeCov and SonarCloud integration for the onvif-go project.

## Overview

The project uses two code quality platforms:
- **CodeCov** - Code coverage tracking and visualization
- **SonarCloud** - Code quality, security vulnerabilities, and technical debt analysis

## CodeCov Integration

### What is CodeCov?

CodeCov provides code coverage reports and metrics to help ensure your tests cover your codebase effectively.

### Setup Steps

1. **Sign up for CodeCov**
   - Go to https://codecov.io/
   - Sign in with your GitHub account
   - Authorize CodeCov to access your repositories

2. **Add Repository**
   - Navigate to https://codecov.io/gh/0x524a
   - Click "Add new repository"
   - Select `onvif-go` from the list

3. **Get Upload Token**
   - In the repository settings on CodeCov, find your upload token
   - Copy the token (format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)

4. **Add Secret to GitHub**
   - Go to https://github.com/0x524a/onvif-go/settings/secrets/actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: Paste your CodeCov upload token
   - Click "Add secret"

### Configuration Files

The following files configure CodeCov:

**`.codecov.yml`** - CodeCov configuration
```yaml
codecov:
  require_ci_to_pass: yes

coverage:
  precision: 2
  round: down
  range: "70...100"
  status:
    project:
      default:
        target: 45%        # Current coverage target
        threshold: 1%      # Allow 1% decrease
    patch:
      default:
        target: 80%        # New code should have 80% coverage
        threshold: 5%
```

**Key Settings:**
- **Project target**: 45% (matches current coverage)
- **Patch target**: 80% (new code should be well-tested)
- **Threshold**: 1% decrease allowed to prevent flaky failures
- **Excluded**: Examples, commands, test files

### Viewing Reports

After setup, coverage reports will be available at:
- Main dashboard: https://codecov.io/gh/0x524a/onvif-go
- Pull request comments will show coverage changes
- Commit-level coverage available in GitHub checks

### Coverage Badges

The README includes a CodeCov badge:
```markdown
[![codecov](https://codecov.io/gh/0x524a/onvif-go/branch/master/graph/badge.svg)](https://codecov.io/gh/0x524a/onvif-go)
```

## SonarCloud Integration

### What is SonarCloud?

SonarCloud provides continuous code quality analysis, detecting bugs, vulnerabilities, code smells, and security hotspots.

### Setup Steps

1. **Sign up for SonarCloud**
   - Go to https://sonarcloud.io/
   - Click "Log in" and sign in with GitHub
   - Authorize SonarCloud to access your repositories

2. **Import Repository**
   - Click the "+" button in the top right
   - Select "Analyze new project"
   - Choose `0x524a/onvif-go`
   - Click "Set Up"

3. **Configure Organization**
   - Organization key: `0x524a`
   - Project key: `0x524a_onvif-go`
   - These are already set in `sonar-project.properties`

4. **Get Authentication Token**
   - Go to https://sonarcloud.io/account/security
   - Generate a new token
   - Name it "GitHub Actions - onvif-go"
   - Copy the token

5. **Add Secret to GitHub**
   - Go to https://github.com/0x524a/onvif-go/settings/secrets/actions
   - Click "New repository secret"
   - Name: `SONAR_TOKEN`
   - Value: Paste your SonarCloud token
   - Click "Add secret"

### Configuration Files

**`sonar-project.properties`** - SonarCloud configuration
```properties
sonar.projectKey=0x524a_onvif-go
sonar.organization=0x524a
sonar.projectName=onvif-go

# Source and test locations
sonar.sources=.
sonar.tests=.
sonar.test.inclusions=**/*_test.go

# Coverage report
sonar.go.coverage.reportPaths=coverage.out

# Exclusions
sonar.exclusions=**/vendor/**,**/*_test.go,**/examples/**,**/cmd/**
sonar.coverage.exclusions=**/cmd/**,**/examples/**,**/*_test.go
```

**Key Settings:**
- **Language**: Go
- **Coverage**: Uses Go's native coverage.out format
- **Exclusions**: Examples, commands, and test files excluded from analysis
- **Source encoding**: UTF-8

### Quality Gates

SonarCloud will check:
- **Bugs**: Serious coding errors
- **Vulnerabilities**: Security issues
- **Code Smells**: Maintainability issues
- **Coverage**: Test coverage percentage
- **Duplications**: Copy-pasted code
- **Security Hotspots**: Potential security risks

### Viewing Reports

After setup, reports will be available at:
- Main dashboard: https://sonarcloud.io/project/overview?id=0x524a_onvif-go
- Pull request decoration shows issues inline
- Quality gate status in GitHub checks

### SonarCloud Badges

The README includes SonarCloud badges:
```markdown
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
```

Additional badges available:
```markdown
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=bugs)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
```

## GitHub Actions Workflows

### Coverage Workflow

**File**: `.github/workflows/coverage.yml`

Runs on:
- Push to master/main/develop branches
- Pull requests to master/main/develop

Steps:
1. Checkout code with full history (required for SonarCloud)
2. Set up Go 1.21
3. Install dependencies
4. Run tests with race detector and coverage
5. Upload coverage to CodeCov
6. Run SonarCloud analysis
7. Generate HTML coverage report
8. Archive coverage artifacts

### Test Workflow

**File**: `.github/workflows/test.yml`

Runs on:
- Push to master/main/develop branches
- Pull requests to master/main/develop

Matrix testing:
- **Operating Systems**: Ubuntu, macOS, Windows
- **Go Versions**: 1.21, 1.22, 1.23

Includes:
- Unit tests with race detector
- Build verification
- golangci-lint code quality checks

## Required GitHub Secrets

Set up these secrets in your GitHub repository:

| Secret Name | Source | Purpose |
|------------|--------|---------|
| `CODECOV_TOKEN` | CodeCov dashboard | Upload coverage reports |
| `SONAR_TOKEN` | SonarCloud account security | Run code quality analysis |

### How to Add Secrets

1. Go to repository settings: https://github.com/0x524a/onvif-go/settings/secrets/actions
2. Click "New repository secret"
3. Enter name and value
4. Click "Add secret"

**Note**: `GITHUB_TOKEN` is automatically provided by GitHub Actions and doesn't need to be added manually.

## Local Testing

### Run Coverage Locally

```bash
# Generate coverage report
go test -v -race -covermode=atomic -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Test CodeCov Upload (requires token)

```bash
# Install codecov CLI
go install github.com/codecov/codecov-cli@latest

# Upload coverage
codecov upload-process --file coverage.out --token YOUR_CODECOV_TOKEN
```

### Run SonarCloud Locally (requires Docker)

```bash
# Using sonar-scanner Docker image
docker run --rm \
  -e SONAR_HOST_URL="https://sonarcloud.io" \
  -e SONAR_TOKEN="YOUR_SONAR_TOKEN" \
  -v "$(pwd):/usr/src" \
  sonarsource/sonar-scanner-cli
```

## Troubleshooting

### CodeCov Issues

**Problem**: Coverage upload fails
```
Error: No coverage reports found
```

**Solution**:
- Ensure `coverage.out` is generated: `go test -coverprofile=coverage.out ./...`
- Check the file exists: `ls -la coverage.out`
- Verify the workflow has the correct path

**Problem**: Coverage percentage is 0%
```
Coverage: 0.00%
```

**Solution**:
- Ensure tests are actually running: `go test -v ./...`
- Check coverage mode is set: `-covermode=atomic`
- Verify exclusions in `.codecov.yml` aren't too broad

### SonarCloud Issues

**Problem**: Analysis fails with authentication error
```
Error: Invalid authentication token
```

**Solution**:
- Regenerate token in SonarCloud account security
- Update `SONAR_TOKEN` secret in GitHub
- Ensure token has project analysis permissions

**Problem**: No coverage data in SonarCloud
```
Warning: No coverage information
```

**Solution**:
- Verify `coverage.out` exists before SonarCloud scan
- Check `sonar.go.coverage.reportPaths=coverage.out` in properties
- Ensure coverage file is in Go format (not HTML)

### GitHub Actions Issues

**Problem**: Workflow doesn't run
```
No checks ran on this commit
```

**Solution**:
- Check workflow triggers match your branch name
- Verify YAML syntax is valid
- Look at Actions tab for error messages

**Problem**: Secrets not found
```
Error: CODECOV_TOKEN is not set
```

**Solution**:
- Add secret in repository settings
- Check secret name matches exactly (case-sensitive)
- Verify you have repository admin permissions

## Coverage Goals

### Current Status
- **Overall Coverage**: 44.6%
- **Device Management**: 100% API implementation
- **New Code**: 88-100% per file

### Improvement Plan

1. **Short-term** (Target: 50%)
   - Add integration tests for Media service
   - Expand PTZ control testing
   - Test error scenarios more thoroughly

2. **Medium-term** (Target: 60%)
   - Add end-to-end tests with mock camera
   - Test concurrent operations
   - Expand discovery testing

3. **Long-term** (Target: 70%+)
   - Integration tests with real devices
   - Stress testing and edge cases
   - Performance benchmarks

### Coverage Exclusions

The following are excluded from coverage metrics:
- **Examples** (`examples/`) - Demonstration code
- **Commands** (`cmd/`) - CLI tools
- **Server** (`server/`) - Mock server implementation
- **Test utilities** (`testing/`) - Test helpers
- **Test files** (`*_test.go`) - Test code itself

## Best Practices

### Writing Testable Code

1. **Use interfaces** for dependencies
2. **Inject dependencies** via constructors
3. **Keep functions focused** - single responsibility
4. **Avoid global state** - use struct methods
5. **Mock external services** - don't rely on real cameras for unit tests

### Maintaining Coverage

1. **Write tests first** (TDD) when adding features
2. **Test happy path and errors** for each function
3. **Use table-driven tests** for multiple scenarios
4. **Mock HTTP clients** with httptest
5. **Check coverage locally** before pushing

### Code Quality

1. **Fix issues early** - address SonarCloud findings promptly
2. **Keep functions small** - easier to test and maintain
3. **Document public APIs** - helps maintain quality
4. **Use golangci-lint** - catches issues before they reach SonarCloud
5. **Review coverage reports** - identify untested code paths

## Monitoring & Reporting

### Regular Checks

- **Weekly**: Review coverage trends on CodeCov
- **Per PR**: Check coverage changes and SonarCloud findings
- **Monthly**: Review quality gate trends on SonarCloud
- **Quarterly**: Update coverage targets based on progress

### Metrics to Track

| Metric | Tool | Target | Current |
|--------|------|--------|---------|
| Overall Coverage | CodeCov | 45% | 44.6% |
| New Code Coverage | CodeCov | 80% | 88-100% |
| Quality Gate | SonarCloud | Pass | TBD |
| Code Smells | SonarCloud | <50 | TBD |
| Security Rating | SonarCloud | A | TBD |
| Maintainability | SonarCloud | A | TBD |

## References

- **CodeCov Documentation**: https://docs.codecov.com/
- **SonarCloud Documentation**: https://docs.sonarcloud.io/
- **GitHub Actions**: https://docs.github.com/en/actions
- **Go Testing**: https://pkg.go.dev/testing
- **Go Coverage**: https://go.dev/blog/cover

## Support

If you encounter issues with the coverage setup:

1. Check the [troubleshooting section](#troubleshooting) above
2. Review GitHub Actions logs in the repository
3. Check CodeCov/SonarCloud status pages
4. Open an issue on GitHub with:
   - Error message
   - Workflow run link
   - Steps to reproduce

---

**Setup Status**: ⚠️ Requires manual configuration

**Next Steps**:
1. ✅ Configuration files created
2. ⏳ Sign up for CodeCov and SonarCloud
3. ⏳ Add repository secrets to GitHub
4. ⏳ Push changes to trigger first workflow run
5. ⏳ Verify badges appear in README

Once setup is complete, coverage and quality metrics will be automatically tracked for all commits and pull requests!
