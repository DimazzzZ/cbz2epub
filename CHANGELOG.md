# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-01-16

### Added
- Streaming EPUB converter (`ConvertStreaming`) for memory-efficient processing of large CBZ files
- Streaming CBZ reader (`IterateImages`) with callback-based image processing
- Comprehensive test suite with 88.5% code coverage
- CI/CD pipeline: automated testing, linting, and release builds
- Release workflow with cross-platform binary builds (Linux/macOS/Windows × amd64/arm64)
- `-version` flag to display application version

### Changed
- **Breaking**: Moved main entry point from root `main.go` into `cmd/cbz2epub/main.go`
  - Users building from source must now run: `go build -o cbz2epub ./cmd/cbz2epub`
- Refactored internal package structure: collapsed `util/` package into `epub/`
- Improved error handling in CBZ reader with dedicated `readCBZEntries` helper
- Enhanced CLI with testable `execute()` function for better integration testing

### Fixed
- Proper handling of image MIME types in EPUB generation
- Deterministic UUID generation in EPUB metadata

### Removed
- `util/util.go` and `util/util_test.go` (functionality moved to `epub/`)
- Root-level `main.go` (moved to `cmd/cbz2epub/`)

### Technical Notes
- **Test Coverage**: 88.5% across all packages. Uncovered lines are primarily error-handling branches that would require filesystem mocking.
- **Go Version**: Requires Go 1.24 or later
- **Performance**: Streaming reader reduces memory usage for large CBZ files (>500MB)

---

## [0.1.0] - 2025-01-01

### Added
- Initial release
- Basic CBZ file reading and merging
- CBZ to EPUB conversion
- Command-line interface with merge and convert modes
- Recursive directory processing
- Verbose output option
