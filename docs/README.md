# onvif-go documentation

The package reference lives on [pkg.go.dev](https://pkg.go.dev/github.com/0x524a/onvif-go),
and a searchable API reference is published on the
[project site](https://0x524a.github.io/onvif-go/api.html). These guides cover
what the reference cannot.

## Getting started

- [Quick start](QUICKSTART.md): install, connect, and make your first calls.
- [Architecture](ARCHITECTURE.md): how the client, SOAP layer, discovery and server fit together.

## Command-line tools

- [Non-interactive mode](CLI_NON_INTERACTIVE_MODE.md): using `onvif-cli` in scripts and CI.
- [Network interface selection](CLI_NETWORK_INTERFACE_USAGE.md): choosing an interface for discovery on multi-homed hosts.

## Device API

- [Device API status](api/DEVICE_API_STATUS.md): every Device Management operation and whether it is implemented.
- [Device API quick reference](api/DEVICE_API_QUICKREF.md): common calls with examples.

## Testing

- [Camera tests](CAMERA_TESTS.md): how camera-specific regression tests work.
- [Adding your camera](testing/CAMERA_TESTING_FLOW.md): capture a camera and turn it into tests.
- [Coverage setup](testing/COVERAGE_SETUP.md): coverage, Codecov and SonarCloud.
- [XML debugging](XML_DEBUGGING_SOLUTION.md): capturing raw SOAP to diagnose a camera.
- [RTSP stream inspection](RTSP_STREAM_INSPECTION.md): checking that a stream URI is reachable.

## Project

- [CI/CD](CI_CD.md): the workflows and what each one checks.
- [Changelog](../CHANGELOG.md) and [contributing](../.github/CONTRIBUTING.md).
