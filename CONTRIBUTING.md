# Contributing to onvif-go

Thanks for your interest in contributing! This document covers the workflow,
the quality bar, and what makes a good bug report for an ONVIF library.

## Workflow: pull requests only

The `main` branch is protected: all three CI jobs (lint, test, build) are
required status checks, **direct pushes are rejected — everything merges
through a pull request**, including changes by the maintainer.

1. Fork the repo and create a branch from `main`
2. Make your change with tests
3. Ensure the full check suite passes locally:

```bash
make check   # lint + test
make fmt     # gofumpt + goimports via golangci-lint fmt — run before committing
```

4. Open a pull request; CI must be green before merge

Commit messages: imperative mood, first line ≤ 72 characters, reference
issues after the first line (`Fixes #123`).

Example:
```
Add GetAnalyticsConfiguration

- Implement the Analytics service facade
- Add mock-device tests
- Update media docs (en/zh)

Closes #123
```

## Quality bar

- **Zero lint findings.** `golangci-lint run` (v2) must report nothing;
  exemptions require a written rationale in `.golangci.yml`.
- **Formatting** is gofumpt + goimports (enforced by `make fmt` and checked
  in CI).
- **Tests are written first (TDD)**: a bug fix lands with the reproducing
  test, a new service operation with the wire-contract test. Behavioral
  tests run against `httptest` mock devices or the in-memory
  `internal/testutil.FakeCaller` — see
  the testing guide in the `docs/README.md` for the layers, conventions,
  bounded-wait/deadlock rules, and CI timing budgets. Real-camera tests
  are environment-gated and never run in CI.
- **Zero third-party dependencies** in the library module (stdlib only).
  This is a deliberate constraint; propose changes accordingly.
- Go 1.26; standard Go idioms, self-documenting names, comments where the
  code cannot speak for itself.

## Reporting bugs

Beyond the basics (reproduction steps, Go version, OS), the two things that
matter most for this library:

- **Camera model and firmware version** — ONVIF quirks are almost always
  firmware-specific.
- **Raw SOAP exchange** when possible — run `onvif diagnostics` (cmd/onvif) with
  `-capture-xml` (see the CLI guide in the `docs/README.md`) and attach the
  output **after redacting credentials**. Captured responses can become
  `testdata/captures/` fixtures so the regression outlives the camera.

Never include real credentials in issues, examples, or test fixtures.

## Documentation

User-facing behavior changes should update the topic docs in the
`docs/README.md` — both languages
(the hub's `mibeelibs/` collection, `zh-CN` + `en-US` pages in one PR).
API additions go into the
[README](README.md) quick-start only when they are primary-surface features,
and significant changes get a [CHANGELOG](CHANGELOG.md) entry.

## Attribution

This project continues [0x524a/onvif-go](https://github.com/0x524a/onvif-go)
(MIT). Keep the existing copyright lines intact; contributions land under
the same MIT license.
