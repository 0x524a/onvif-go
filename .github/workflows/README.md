# GitHub Actions Workflows

This directory contains all CI/CD workflows for the ONVIF Go library.

## Workflows

### 🔄 CI (`ci.yml`) - Main Pipeline
**Unified continuous integration workflow with fail-fast behavior.**

The CI pipeline runs sequentially - if any stage fails, subsequent stages are skipped:

```
fmt → lint → test → sonarcloud
                  ↘ build
```

**Stages:**

| Stage | Description | Depends On |
|-------|-------------|------------|
| **fmt** | Format check using `gofmt -s` | - |
| **lint** | Static analysis with `go vet` and `golangci-lint` | fmt |
| **test** | Unit tests with race detector + coverage | lint |
| **sonarcloud** | Code quality & security analysis (push to master only) | test |
| **build** | Build verification for all packages | test |
| **ci-success** | Final status check | all |

**Features:**
- ✅ Fail-fast: stops immediately if any check fails
- ✅ Codecov integration for coverage reporting
- ✅ SonarCloud integration for code quality
- ✅ Go module caching for faster builds
- ✅ Concurrency control (cancels in-progress runs)

**Triggers:**
- Push to `master`, `main`
- All pull requests targeting `master`, `main`

**Required for PR Merge:**
All stages must pass before a PR can be merged. Configure branch protection rules in GitHub:
1. Go to **Settings → Branches → Branch protection rules**
2. Add rule for `master`
3. Enable **Require status checks to pass before merging**
4. Select these required checks:
   - `Format Check`
   - `Lint`
   - `Test & Coverage`
   - `SonarCloud Analysis`
   - `Build Verification`
   - `CI Success`

---

### 🧪 Extended Tests (`test.yml`)
Extended testing workflow for comprehensive test coverage.

**Jobs:**
- **test-older-versions** - Test on older Go versions (1.19, 1.20)
- **benchmark** - Run benchmark tests
- **race-detector** - Extended race detector tests

**Triggers:**
- Manual dispatch
- Weekly schedule (Sunday 2 AM UTC)
- Push to `master`/`main` when Go files change

---

### 🚀 Release (`release.yml`)
Automated release workflow for creating GitHub releases.

**Jobs:**
- **build** - Build binaries for all platforms (Linux, Windows, macOS, multiple architectures)
- **release** - Create GitHub release with artifacts
- **docker** - Build and push Docker images to GHCR

**Triggers:**
- Push tags matching `v*.*.*`
- Manual dispatch with version input

---

### 🔒 Security (`security.yml`)
Security scanning workflow.

**Jobs:**
- **gosec** - Security scanner
- **govulncheck** - Vulnerability checker

**Triggers:**
- Push to `master`/`main`
- Pull requests
- Weekly schedule

---

### 📚 Documentation (`docs.yml`)
Documentation validation workflow.

**Triggers:**
- Push to `master`/`main` when docs change
- Manual dispatch

---

### 🔐 Dependency Review (`dependency-review.yml`)
Dependency vulnerability review.

**Triggers:**
- Pull requests

---

## CI Pipeline Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         CI PIPELINE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐     ┌─────────┐     ┌─────────────────────────┐   │
│  │   FMT   │────▶│  LINT   │────▶│  TEST + COVERAGE        │   │
│  └─────────┘     └─────────┘     └───────────┬─────────────┘   │
│                                              │                  │
│                                    ┌─────────┴─────────┐       │
│                                    ▼                   ▼       │
│                            ┌────────────┐      ┌───────────┐   │
│                            │ SONARCLOUD │      │   BUILD   │   │
│                            │ (push only)│      └───────────┘   │
│                            └────────────┘              │       │
│                                    │                   │       │
│                                    └─────────┬─────────┘       │
│                                              ▼                 │
│                                    ┌─────────────────┐         │
│                                    │   CI SUCCESS    │         │
│                                    └─────────────────┘         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

❌ If any stage fails, the pipeline stops immediately (fail-fast)
ℹ️ SonarCloud only runs on push to master/main (skipped for PRs)
```

---

## SonarCloud Configuration

Security Hotspot analysis excludes:
- Test files (`**/*_test.go`)
- CI configuration (`**/.github/**`)
- Test utilities (`**/testing/**`)
- Example code (`**/examples/**`)
- CLI tools (`**/cmd/**`)

This ensures security analysis focuses on production library code.

---

## Required Secrets

| Secret | Required | Description |
|--------|----------|-------------|
| `CODECOV_TOKEN` | Yes | Coverage reporting to Codecov |
| `SONAR_TOKEN` | Yes | SonarCloud code analysis |
| `DOCKERHUB_USERNAME` | No | Docker Hub releases |
| `DOCKERHUB_TOKEN` | No | Docker Hub releases |

---

## Workflow Status

- ✅ Go 1.24 as primary version
- ✅ Unified fail-fast CI pipeline
- ✅ Go module caching for faster builds
- ✅ Artifact uploads for coverage and releases
- ✅ Concurrency control

---

*Last Updated: December 3, 2025*
