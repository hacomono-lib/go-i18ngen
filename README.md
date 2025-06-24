# i18ngen - Type-Safe Internationalization Code Generator

[![CI](https://github.com/hacomono-lib/go-i18ngen/workflows/CI/badge.svg)](https://github.com/hacomono-lib/go-i18ngen/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

i18ngen is a CLI tool that automatically generates type-safe Go code for internationalization (i18n) from YAML configuration files. It creates strongly-typed structs and functions that ensure compile-time safety for localized messages and placeholders.

## Features

- **Type-safe message construction** - No more string concatenation or runtime errors
- **Automatic template rendering** - Built-in Go template processing with placeholders
- **Template functions** - Essential string manipulation functions (title, upper, lower)
- **Suffix notation** - Meaningful parameter names with `:suffix` syntax (e.g., `{{.entity:from}}`, `{{.field:input}}`)
- **Locale fallback handling** - Graceful degradation when translations are missing
- **Compile-time validation** - Catch missing parameters at build time, not runtime
- **Utility access patterns** - Pre-defined instances for common use cases
- **Standard Go patterns** - Uses fmt.Errorf and standard error handling (no custom error systems)
- **Sequential processing** - Simple, reliable processing suitable for typical i18n project sizes

## Installation

```bash
go install github.com/hacomono-lib/go-i18ngen@latest
```

## Quick Start

1. Create a configuration file `config.yaml`:

```yaml
compound: true
locales: [ja, en]
messages: "./messages/*.yaml"
placeholders: "./placeholders/*.yaml"
output_dir: "./"
output_package: "i18n"
```

2. Create message files `messages/messages.yaml`:

```yaml
EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"
UserAlreadyExists:
  ja: "{{.entity}}はすでに存在します: {{.user_id}}"
  en: "{{.entity}} already exists: {{.user_id}}"
```

3. Create placeholder files `placeholders/entity.yaml`:

```yaml
user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
```

4. Generate type-safe Go code:

```bash
go-i18ngen generate --config config.yaml
```

5. Use the generated code:

```go
// Create a localized error message
msg := NewEntityNotFound(EntityTexts.User, ReasonTexts.AlreadyDeleted)
fmt.Println(msg.Localize("ja")) // "ユーザーが見つかりません: すでに削除されています"
fmt.Println(msg.Localize("en")) // "User not found: already deleted"
```

## Usage Examples

### Basic Message Localization

```go
msg := NewEntityNotFound(EntityTexts.User, ReasonTexts.AlreadyDeleted)
fmt.Println(msg.Localize("ja")) // "ユーザーが見つかりません: すでに削除されています"
fmt.Println(msg.Localize("en")) // "User not found: already deleted"
```

### Suffix Notation in Practice

```go
// File copy operation with meaningful parameter names
fileCopy := NewFileCopyMessage(
    NewFileValue("document.pdf"),     // Source file name
    NewFileValue("backup.pdf"),       // Destination file name
)
fmt.Println(fileCopy.Localize("ja")) // "document.pdfからbackup.pdfへコピーしました"
fmt.Println(fileCopy.Localize("en")) // "Copied from document.pdf to backup.pdf"

// Value comparison with old and new values
comparison := NewComparisonMessage(
    NewValueValue("100"),        // Old value
    NewValueValue("200"),        // New value
)
fmt.Println(comparison.Localize("ja")) // "100から200に変更されました"
fmt.Println(comparison.Localize("en")) // "Changed from 100 to 200"

// Person introduction message
introduction := NewIntroductionMessage(
    NewPersonValue("Alice"),     // Introducer name
    NewPersonValue("Bob"),       // Introducee name
)
fmt.Println(introduction.Localize("ja")) // "AliceがBobを紹介しました"
fmt.Println(introduction.Localize("en")) // "Alice introduced Bob"
```

### Template Functions in Action

Assuming you have placeholders defined in lowercase:

```yaml
# placeholders/entity.yaml
member:
  ja: "会員"
  en: "member"

# messages/messages.yaml
CMN_000004:
  ja: "データが見つかりません。[{{.entity}}]"
  en: "Data not found.[{{.entity | title}}]"
```

The generated code will produce:

```go
memberEntity := NewEntityText("member")
msg := NewCMN000004(memberEntity)

fmt.Println(msg.Localize("ja")) // "データが見つかりません。[会員]"
fmt.Println(msg.Localize("en")) // "Data not found.[Member]"  // ← Capitalized by title function
```

### With Dynamic Values

```go
msg := NewUserAlreadyExists(EntityTexts.User, NewUserIdValue("user123"))
fmt.Println(msg.Localize("ja")) // "ユーザーはすでに存在します: user123"
```

### Error Messages in Application Code

```go
func GetUser(id string) (*User, error) {
    user, err := repo.FindByID(id)
    if err != nil {
        msg := NewEntityNotFound(EntityTexts.User, NewUserIdValue(id))
        return nil, errors.New(msg.Localize("ja"))
    }
    return user, nil
}
```

## Configuration

### config.yaml

| Field | Type | Description |
|-------|------|-------------|
| `compound` | bool | Use compound format (supports multiple locales per file) |
| `locales` | []string | Supported locales (first is primary/fallback) |
| `messages` | string | Glob pattern for message files |
| `placeholders` | string | Glob pattern for placeholder files |
| `output_dir` | string | Output directory |
| `output_package` | string | Generated package name |

### Message Format

Messages support Go template syntax with placeholders and template functions:

```yaml
MessageName:
  ja: "{{.placeholder}}のメッセージ"
  en: "Message with {{.placeholder}}"

# Using template functions for string manipulation
DataNotFound:
  ja: "データが見つかりません。[{{.entity}}]"
  en: "Data not found.[{{.entity | title}}]"  # Capitalizes first letter

ValidationError:
  ja: "入力エラー: {{.field}}"
  en: "Validation error: {{.field | upper}}"  # Converts to uppercase
```

#### Suffix Notation (Recommended)

Use suffix notation to create meaningful parameter names for multiple instances of the same placeholder type:

```yaml
FileCopyMessage:
  ja: "{{.file:source}}から{{.file:destination}}へコピーしました"
  en: "Copied from {{.file:source}} to {{.file:destination}}"

ComparisonMessage:
  ja: "{{.value:old}}から{{.value:new}}に変更されました"
  en: "Changed from {{.value:old}} to {{.value:new}}"

IntroductionMessage:
  ja: "{{.person:introducer}}が{{.person:introducee}}を紹介しました"
  en: "{{.person:introducer}} introduced {{.person:introducee}}"
```

This generates meaningful parameter names and struct fields:

```go
type FileCopyMessage struct {
    FileSource      FileValue  // {{.file:source}}
    FileDestination FileValue  // {{.file:destination}}
}

func NewFileCopyMessage(fileSource FileValue, fileDestination FileValue) FileCopyMessage

type ComparisonMessage struct {
    ValueOld ValueValue  // {{.value:old}}
    ValueNew ValueValue  // {{.value:new}}
}

func NewComparisonMessage(valueOld ValueValue, valueNew ValueValue) ComparisonMessage

// Usage examples:
fileCopy := NewFileCopyMessage(NewFileValue("source.txt"), NewFileValue("backup.txt"))
comparison := NewComparisonMessage(NewValueValue("old_value"), NewValueValue("new_value"))
```

**Suffix notation supports:**
- **Descriptive suffixes**: `:from`, `:to`, `:input`, `:display`, `:user`, `:admin`
- **Numeric suffixes**: `:1`, `:2`, `:3`
- **Complex suffixes**: `:input_value`, `:display_name`
- **Template functions**: `{{.entity:from | title}}`, `{{.field:input | upper}}`

**Important**: When using the same placeholder type multiple times in a message, you must use suffix notation to distinguish between different instances. Using duplicate placeholders without suffixes will result in a validation error.

#### Available Template Functions

The following template functions are available for string manipulation:

| Function | Description | Example Input | Example Output |
|----------|-------------|---------------|----------------|
| `title` | Capitalizes the first letter | `"member"` | `"Member"` |
| `upper` | Converts to uppercase | `"member"` | `"MEMBER"` |
| `lower` | Converts to lowercase | `"Member"` | `"member"` |

**Usage examples:**

```yaml
# Different formatting for different locales
ErrorMessage:
  ja: "エラー: {{.entity}}"           # Japanese uses entity as-is
  en: "Error: {{.entity | title}}"    # English capitalizes first letter

DebugMessage:
  ja: "デバッグ: {{.component}}"
  en: "DEBUG: {{.component | upper}}"  # English uses all caps for debug

StatusMessage:
  ja: "状態: {{.status}}"
  en: "Status: {{.status | lower}}"    # English uses lowercase
```

This allows you to define base values in lowercase in your placeholder files and format them appropriately for each locale:

```yaml
# placeholders/entity.yaml
member:
  ja: "会員"
  en: "member"    # lowercase base form

facility:
  ja: "施設"
  en: "facility"  # lowercase base form
```

### Placeholder Types

**Text Placeholders** (localized strings):
```yaml
# placeholders/entity.yaml
user:
  ja: "ユーザー"
  en: "User"
```

**Value Placeholders** (non-localized values):
```yaml
# Generated automatically from message templates
# For {{.user_id}} in templates
```

### File Formats

#### Compound Format (Recommended)

Single file contains all locales for each message/placeholder:

```yaml
EntityNotFound:
  ja: "{{.entity}}が見つかりません"
  en: "{{.entity}} not found"
```

#### Simple Format

Separate files for each locale:
- `entity.ja.yaml`
- `entity.en.yaml`

Files must follow `name.locale.ext` pattern.

## CLI Usage

```bash
go-i18ngen generate [flags]
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `-c, --config` | string | Path to config file (default "i18ngen.yaml") |
| `--locales` | []string | List of locales (e.g. ja,en) |
| `--compound` | bool | Use compound format |
| `--messages` | string | Messages glob pattern |
| `--placeholders` | string | Placeholders glob pattern |
| `--output` | string | Output directory |
| `--package` | string | Output package name |

## Generated Code

### Message Structs

```go
type EntityNotFound struct {
    Entity EntityText
    Reason ReasonText
}

func NewEntityNotFound(entity EntityText, reason ReasonText) EntityNotFound
func (m EntityNotFound) Localize(locale string) string
```

### Placeholder Types

```go
// Text placeholders (localized)
type EntityText struct { id string }

// Value placeholders (non-localized)
type UserIdValue struct { Value string }
func NewUserIdValue(value string) UserIdValue
```

### Utility Collections

Pre-defined instances for common usage (Text types only):

```go
var EntityTexts = struct {
    User    EntityText
    Product EntityText
}{
    User:    EntityText{id: "user"},
    Product: EntityText{id: "product"},
}
```

## Type Safety Features

1. **Compile-time Parameter Validation**: Missing or incorrect parameters cause compilation errors
2. **Structured Message Building**: Cannot accidentally mix incompatible placeholder types
3. **Locale-aware Rendering**: Automatic fallback prevents runtime errors
4. **Template Syntax Validation**: Go template parsing errors are caught during generation

## Error Handling

The generated code includes robust error handling:

- **Missing Templates**: Returns formatted error message instead of crashing
- **Template Parse Errors**: Returns descriptive error with details
- **Missing Placeholders**: Returns formatted error message with context
- **Locale Fallback**: Gracefully falls back to primary locale or any available locale

## Integration

The tool is designed to integrate seamlessly with existing Go projects:

- Generates standard Go code following common conventions
- No external runtime dependencies beyond Go standard library
- Compatible with existing build processes and CI/CD pipelines
- Can be integrated into `go generate` workflows

Add to your Go file for automatic generation:

```go
//go:generate go-i18ngen generate --config config.yaml
```

## Limitations

1. **Go Template Syntax**: Must use valid Go template syntax in message templates
2. **Identifier Names**: Placeholder kinds and IDs must be valid Go identifiers
3. **File Naming**: For simple format, files must follow `name.locale.ext` pattern
4. **Template Fields**: Fields in templates must match available placeholder definitions
5. **Static Generation**: Changes require regeneration; no runtime template loading
6. **Suffix Characters**: Suffix names must be valid Go identifier characters (alphanumeric and underscore)

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)
- Docker (optional, for containerized development)

### Local Development

#### Using Make (Recommended)

```bash
# Setup development environment
make dev-setup      # Install required tools (golangci-lint, etc.)

# Development workflow
make fmt            # Format code
make lint           # Run linter
make test           # Run tests with coverage
make build          # Build binary

# All-in-one checks
make check          # Run fmt + lint + test
make ci             # Run full CI pipeline locally
make pre-commit     # Quick checks before committing
```

#### Manual Commands

```bash
# Install dependencies
go mod tidy

# Run tests
go test -v -race ./...

# Build
go build -o build/i18ngen .

# Install locally
go install .
```

### Available Make Targets

Run `make help` to see all available targets:

```bash
make help
```

Key targets:
- `make test` - Run tests with race detector and coverage
- `make test-short` - Run tests without race detector (faster)
- `make coverage` - Generate HTML coverage report
- `make lint` - Run golangci-lint
- `make security` - Run security scan with gosec
- `make build-all` - Build for all platforms
- `make clean` - Clean build artifacts

### Docker Development

#### Using Docker Compose (Recommended)

```bash
# Start development environment
docker-compose run dev

# Run tests in Docker
docker-compose run test

# Run linter in Docker
docker-compose run lint
```

#### Using Docker directly

```bash
# Build Docker image
make docker-build

# Run the built image
docker run --rm go-i18ngen:latest --help

# Development with volume mount
docker run --rm -v $(pwd):/workspace -w /workspace golang:1.23 make test
```

### Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Run tests**: `make check`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

#### Code Style

- Follow standard Go conventions
- Run `make fmt` to format code
- Ensure `make lint` passes without warnings
- Add tests for new functionality
- Update documentation as needed

#### Running Tests

```bash
# Run all tests
make test

# Run specific test
go test -v ./internal/parser -run TestParseMessages

# Run tests with verbose output
go test -v ./...

# Generate coverage report
make coverage
# Open coverage.html in browser
```

#### Testing with Different Go Versions

```bash
# Test with Go 1.21
docker run --rm -v $(pwd):/app -w /app golang:1.21 go test ./...

# Test with Go 1.22
docker run --rm -v $(pwd):/app -w /app golang:1.22 go test ./...

# Test with Go 1.23
docker run --rm -v $(pwd):/app -w /app golang:1.23 go test ./...
```

### Project Structure

```
go-i18ngen/
├── main.go             # CLI application entry point
├── internal/           # Internal packages (unit tests are *_test.go files within each package)
│   ├── cmd/           # CLI commands
│   ├── config/        # Configuration handling
│   ├── generator/     # Code generation logic
│   ├── model/         # Data models
│   ├── parser/        # File parsing
│   ├── templatex/     # Template rendering
│   └── utils/         # Utility functions
├── tests/             # Integration tests
├── testdata/          # Test data and examples
├── .github/           # GitHub Actions CI/CD
├── Dockerfile         # Production Docker image
├── Dockerfile.dev     # Development Docker image
├── docker-compose.yml # Development environment
├── Makefile          # Development commands
└── .golangci.yml     # Linter configuration
```

### Release Process

Releases are automated via GitHub Actions:

1. **Create a tag**: `git tag v1.0.0`
2. **Push the tag**: `git push origin v1.0.0`
3. **GitHub Actions will**:
   - Run all tests
   - Build binaries for multiple platforms
   - Create a GitHub release
   - Attach binaries to the release

### Installation Methods

#### From Source (Development)

```bash
git clone https://github.com/hacomono-lib/go-i18ngen.git
cd go-i18ngen
make install
```

#### Using Go Install (Recommended)

```bash
go install github.com/hacomono-lib/go-i18ngen@latest
```

#### Using Docker

```bash
# Pull and run
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/hacomono-lib/go-i18ngen:latest generate --config config.yaml
```

### Troubleshooting

#### Common Issues

1. **`golangci-lint` not found**
   ```bash
   make dev-setup  # Installs required tools
   ```

2. **Tests failing on different OS**
   ```bash
   docker-compose run test  # Run in standardized environment
   ```

3. **Build fails with Go version mismatch**
   ```bash
   # Check Go version
   go version
   # Should be 1.21 or later
   ```

4. **Permission denied on Docker**
   ```bash
   # Add user to docker group (Linux)
   sudo usermod -aG docker $USER
   ```

## Examples

The `testdata/` directory contains working examples of message and placeholder definitions that demonstrate the various features and usage patterns of i18ngen.
