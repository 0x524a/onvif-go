# 📚 Documentation Index

Welcome to onvif-go! This index helps you navigate all available documentation.

## 🚀 Start Here

**New to onvif-go?**
1. Read: [`README.md`](README.md) - Project overview
2. Read: [`QUICKSTART.md`](QUICKSTART.md) - Get started in 5 minutes
3. Try: `./cmd/onvif-cli/onvif-cli` - Run the CLI

## 📖 Core Documentation

### User Guides

| Document | Purpose | Length | Audience |
|----------|---------|--------|----------|
| [README.md](README.md) | Project overview | Short | Everyone |
| [QUICKSTART.md](QUICKSTART.md) | Getting started | Medium | New users |
| [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md) | CLI automation guide | 800+ lines | Automation engineers |
| [NETWORK_INTERFACE_DISCOVERY.md](NETWORK_INTERFACE_DISCOVERY.md) | Discovery API guide | 400+ lines | Developers |

### Implementation Details

| Document | Purpose | Audience |
|----------|---------|----------|
| [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md) | Status & metrics | Project managers |
| [PROJECT_COMPLETION_SUMMARY.md](PROJECT_COMPLETION_SUMMARY.md) | What was built | Stakeholders |
| [BUILDING.md](BUILDING.md) | Build instructions | Developers |

## 🎯 By Use Case

### I want to...

#### Discover cameras on my network
```bash
./onvif-cli discover -interface eth0
```
→ See [QUICKSTART.md](QUICKSTART.md) or [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md)

#### Use the CLI in a script
```bash
./onvif-cli -op discover -interface eth0 -timeout 5
```
→ Read [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md)

#### Integrate discovery into my Go code
```go
import "github.com/0x524a/onvif-go/discovery"
```
→ Read [NETWORK_INTERFACE_DISCOVERY.md](NETWORK_INTERFACE_DISCOVERY.md)

#### Build the project
```bash
make build-cli
```
→ See [BUILDING.md](BUILDING.md)

#### Run tests
```bash
go test ./discovery -v
```
→ See [BUILDING.md](BUILDING.md)

#### Modernize the CLI with urfave/cli
→ Follow [SAFE_MIGRATION_GUIDE.md](SAFE_MIGRATION_GUIDE.md)

## 📁 Code Structure

```
onvif-go/
├── cmd/onvif-cli/          Main CLI tool (1,195 lines)
├── cmd/onvif-quick/        Quick discovery tool
├── discovery/              Discovery library + tests
├── examples/               5 working example programs
└── docs/                   Additional documentation
```

## 🔍 Quick Reference

### Common Commands

| Command | Purpose |
|---------|---------|
| `./onvif-cli` | Launch interactive menu |
| `./onvif-cli discover -interface eth0` | Discover on specific interface |
| `./onvif-cli -op discover -interface eth0` | Non-interactive discover |
| `go test ./discovery -v` | Run tests |
| `go build ./cmd/onvif-cli` | Build CLI |

### Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/onvif-cli/main.go` | Main CLI implementation | 1,195 |
| `discovery/discovery.go` | Discovery API | ~300 |
| `discovery/discovery_test.go` | Discovery tests | ~400 |

## 📊 Statistics

| Metric | Value |
|--------|-------|
| Total documentation | 1,200+ lines |
| CLI code | 1,195 lines |
| Test code | ~400 lines |
| Code examples | 10+ |
| Working examples | 5 |
| Tests passing | 8/8 ✅ |

## 🎓 Learning Path

### Beginner
1. [README.md](README.md) - Understand what it does
2. [QUICKSTART.md](QUICKSTART.md) - Try it out
3. `./onvif-cli` - Run interactive mode

### Intermediate
1. [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md) - Learn automation
2. [NETWORK_INTERFACE_DISCOVERY.md](NETWORK_INTERFACE_DISCOVERY.md) - Understand API
3. Review examples in `examples/` directory

### Advanced
1. Study `cmd/onvif-cli/main.go` (implementation)
2. Study `discovery/discovery.go` (library)
3. Review `discovery/discovery_test.go` (testing)

### Expert
1. [SAFE_MIGRATION_GUIDE.md](SAFE_MIGRATION_GUIDE.md) - Extend the CLI
2. [URFAVE_CLI_MIGRATION_GUIDE.md](URFAVE_CLI_MIGRATION_GUIDE.md) - Modernize
3. Build custom features

## 🔗 Related Files

### Examples
- `examples/discovery/` - Network discovery example
- `examples/device-info/` - Get device info
- `examples/ptz-control/` - Pan/tilt/zoom
- `examples/imaging-settings/` - Camera imaging
- `examples/complete-demo/` - Full integration

### Other Docs
- [CHANGELOG.md](CHANGELOG.md) - Version history
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [LICENSE](LICENSE) - Project license

## ❓ FAQ

**Q: Where do I start?**
A: Read [README.md](README.md) and [QUICKSTART.md](QUICKSTART.md)

**Q: How do I use the CLI for automation?**
A: See [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md)

**Q: How do I use the discovery API?**
A: See [NETWORK_INTERFACE_DISCOVERY.md](NETWORK_INTERFACE_DISCOVERY.md)

**Q: How do I upgrade the CLI framework?**
A: Follow [SAFE_MIGRATION_GUIDE.md](SAFE_MIGRATION_GUIDE.md)

**Q: Are there examples?**
A: Yes! Check `examples/` directory (5 working programs)

**Q: How do I run tests?**
A: `go test ./discovery -v` (all 8 tests pass)

**Q: Is this production ready?**
A: Yes! See [PROJECT_COMPLETION_SUMMARY.md](PROJECT_COMPLETION_SUMMARY.md)

## 📞 Support

- **General questions:** See [README.md](README.md)
- **Usage questions:** See [QUICKSTART.md](QUICKSTART.md)
- **CLI questions:** See [CLI_NON_INTERACTIVE_MODE.md](CLI_NON_INTERACTIVE_MODE.md)
- **API questions:** See [NETWORK_INTERFACE_DISCOVERY.md](NETWORK_INTERFACE_DISCOVERY.md)
- **Build questions:** See [BUILDING.md](BUILDING.md)
- **Upgrade questions:** See [SAFE_MIGRATION_GUIDE.md](SAFE_MIGRATION_GUIDE.md)

## ✅ Project Status

- ✅ Core features: Complete
- ✅ CLI tool: Production ready
- ✅ Documentation: Comprehensive
- ✅ Tests: All passing
- ✅ Examples: 5 working programs

**Status: PRODUCTION READY** 🚀

---

*Last Updated: 2024*
*Go Version: 1.21+*
*urfave/cli: v2.27.7 (installed)*
