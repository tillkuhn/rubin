# Agent Guidelines for Rubin Repository

## Build/Lint/Test Commands
- `make test` - Run all tests with coverage (uses gotest if available)
- `go test -v ./path/to/package` - Run single test package
- `go test -v -run TestSpecificFunction ./path/to/package` - Run single test function
- `make lint` - Run golangci-lint with auto-fix
- `make fmt` - Format code with go fmt
- `make build` - Build binaries to out/bin
- `make test-int` - Run integration tests (//go:build integration)

## Code Style Guidelines
- **Imports**: Group stdlib, third-party, and internal imports with blank lines
- **Formatting**: Use tabs for Go files (configured in .editorconfig)
- **Error handling**: Use wrapped static errors with fmt.Errorf("%w: ...") pattern
- **Naming**: Follow Go conventions - PascalCase for exported, camelCase for unexported
- **Testing**: Use testify/assert, create mocks in internal/testutil/
- **Logging**: Use zerolog, avoid commented-out logger code
- **Constants**: Define meaningful constants for timeouts and error messages
- **Linting**: golangci-lint v2.2.1 with strict settings (no globals, max 100 lines per function)