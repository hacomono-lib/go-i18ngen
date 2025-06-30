# go-i18ngen - Type-Safe Internationalization Code Generator

[![CI](https://github.com/hacomono-lib/go-i18ngen/workflows/CI/badge.svg)](https://github.com/hacomono-lib/go-i18ngen/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

go-i18ngen is a CLI tool that automatically generates type-safe Go code for internationalization (i18n) from YAML configuration files. It creates strongly-typed structs and functions that ensure compile-time safety for localized messages and placeholders.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Message Format](#message-format)
- [CLI Usage](#cli-usage)
- [Generated Code](#generated-code)
- [Advanced Features](#advanced-features)
- [Integration & Development](#integration--development)
- [Examples](#examples)

## Features

- ✅ **Type-safe message construction** - No more string concatenation or runtime errors
- ✅ **CLDR pluralization support** - Full Unicode CLDR plural rules powered by go-i18n
- ✅ **Automatic template rendering** - Built-in Go template processing with placeholders
- ✅ **Template functions** - Essential string manipulation (title, upper, lower)
- ✅ **Suffix notation** - Meaningful parameter names with `:suffix` syntax
- ✅ **Locale fallback handling** - Graceful degradation when translations are missing
- ✅ **Compile-time validation** - Catch missing parameters at build time
- ✅ **Utility access patterns** - Pre-defined instances for common use cases
- ✅ **Standard Go patterns** - Uses standard error handling, no custom error systems
- ✅ **Sequential processing** - Simple, reliable processing for typical i18n projects

## Installation

```bash
go install github.com/hacomono-lib/go-i18ngen@latest
```

Alternative methods:
- [From source](#from-source-development)
- [Using Docker](#using-docker)

## Quick Start

### 1. Create Configuration

Create `config.yaml`:

```yaml
compound: true
locales: [ja, en]              # ja is default language, en is secondary
messages: "./messages/*.yaml"
placeholders: "./placeholders/*.yaml"
output_dir: "./"
output_package: "i18n"
```

### 2. Define Messages

Create `messages/messages.yaml`:

```yaml
EntityNotFound:
  ja: "{{.entity}}が見つかりません: {{.reason}}"
  en: "{{.entity}} not found: {{.reason}}"

UserAlreadyExists:
  ja: "{{.entity}}はすでに存在します: {{.user_id}}"
  en: "{{.entity}} already exists: {{.user_id}}"

# Pluralization support
ItemCount:
  ja: "{{.entity}} アイテム ({{.Count}}個)"
  en:
    one: "{{.entity}} item"
    other: "{{.entity}} items ({{.Count}})"
```

### 3. Define Placeholders

Create `placeholders/entity.yaml`:

```yaml
user:
  ja: "ユーザー"
  en: "User"
product:
  ja: "製品"
  en: "Product"
```

Create `placeholders/reason.yaml`:

```yaml
not_found:
  ja: "見つかりません"
  en: "not found"
already_deleted:
  ja: "すでに削除されています"
  en: "already deleted"
```

### 4. Generate Code

```bash
go-i18ngen generate --config config.yaml
```

### 5. Use Generated Code

```go
package main

import "fmt"

func main() {
    // Type-safe message construction
    msg := NewEntityNotFound(EntityTexts.User, ReasonTexts.AlreadyDeleted)
    
    fmt.Println(msg.Localize("ja")) // "ユーザーが見つかりません: すでに削除されています"
    fmt.Println(msg.Localize("en")) // "User not found: already deleted"
    
    // Pluralization support
    items := NewItemCount(EntityTexts.Product).WithPluralCount(5)
    fmt.Println(items.Localize("en")) // "Product items (5)"
    
    // Dynamic values
    userError := NewUserAlreadyExists(EntityTexts.User, NewUserIdValue("user123"))
    fmt.Println(userError.Localize("ja")) // "ユーザーはすでに存在します: user123"
}
```

## Configuration

### Configuration File Format

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `compound` | bool | Yes | Use compound format (multiple locales per file) |
| `locales` | []string | Yes | Supported locales (first is default language for go-i18n bundle) |
| `messages` | string | Yes | Glob pattern for message files |
| `placeholders` | string | Yes | Glob pattern for placeholder files |
| `output_dir` | string | Yes | Output directory for generated code |
| `output_package` | string | Yes | Generated package name |
| `plural_placeholders` | []string | No | Custom plural placeholder names |

### Example Configuration

```yaml
# Japanese-first project (日本語がデフォルト言語)
compound: true
locales: [ja, en, fr]        # ja is default language for go-i18n bundle
messages: "./locales/messages/*.yaml"
placeholders: "./locales/placeholders/*.yaml"
output_dir: "./internal/i18n"
output_package: "i18n"

# English-first project (English is default language)
compound: true
locales: [en, ja, fr]        # en is default language for go-i18n bundle
messages: "./locales/messages/*.yaml"
placeholders: "./locales/placeholders/*.yaml"
output_dir: "./internal/i18n"
output_package: "i18n"

# Optional: Custom plural placeholders
plural_placeholders: ["Count", "Quantity", "Amount", "Total"]
```

### File Formats

#### Compound Format (Recommended)

Single file contains all locales for each message/placeholder:

```yaml
# messages/errors.yaml
EntityNotFound:
  ja: "{{.entity}}が見つかりません"
  en: "{{.entity}} not found"
  fr: "{{.entity}} introuvable"

# placeholders/entity.yaml
user:
  ja: "ユーザー"
  en: "User"
  fr: "Utilisateur"
```

#### Simple Format

Separate files for each locale:
- `entity.ja.yaml`
- `entity.en.yaml`
- `entity.fr.yaml`

Files must follow `name.locale.ext` pattern.

## Message Format

### Basic Template Syntax

Messages support Go template syntax with placeholders:

```yaml
WelcomeMessage:
  ja: "{{.name}}さん、ようこそ！"
  en: "Welcome, {{.name}}!"

ErrorMessage:
  ja: "エラー: {{.entity}}が{{.reason}}"
  en: "Error: {{.entity}} {{.reason}}"
```

### Template Functions

| Function | Description | Input Example | Output Example |
|----------|-------------|---------------|----------------|
| `title` | Capitalizes first letter | `"member"` | `"Member"` |
| `upper` | Converts to uppercase | `"error"` | `"ERROR"` |
| `lower` | Converts to lowercase | `"WARNING"` | `"warning"` |

```yaml
FormattedMessage:
  ja: "エラー: {{.entity}}"
  en: "Error: {{.entity | title}}"  # Capitalizes entity
  
DebugMessage:
  ja: "デバッグ: {{.component}}"
  en: "DEBUG: {{.component | upper}}"  # Uppercase for debug
```

### Suffix Notation (Advanced)

Use suffix notation for multiple instances of the same placeholder type:

```yaml
# Without suffix (ERROR - duplicate placeholders)
FileCopyMessage:
  ja: "{{.file}}から{{.file}}へコピー"  # ❌ Ambiguous
  en: "Copy from {{.file}} to {{.file}}"  # ❌ Ambiguous

# With suffix notation (CORRECT)
FileCopyMessage:
  ja: "{{.file:source}}から{{.file:destination}}へコピー"
  en: "Copy from {{.file:source}} to {{.file:destination}}"

ComparisonMessage:
  ja: "{{.value:old}}から{{.value:new}}に変更"
  en: "Changed from {{.value:old}} to {{.value:new}}"

IntroductionMessage:
  ja: "{{.person:introducer}}が{{.person:introducee}}を紹介"
  en: "{{.person:introducer}} introduced {{.person:introducee}}"
```

This generates meaningful struct fields:

```go
type FileCopyMessage struct {
    FileSource      FileValue    // {{.file:source}}
    FileDestination FileValue    // {{.file:destination}}
}

type ComparisonMessage struct {
    ValueOld ValueValue  // {{.value:old}}
    ValueNew ValueValue  // {{.value:new}}
}
```

### Pluralization

Certain placeholder names trigger pluralization support:

**Default plural placeholder**: `Count` (case-insensitive, also includes `count`)

```yaml
ItemCount:
  ja: "{{.entity}} アイテム ({{.Count}}個)"
  en:
    one: "{{.Count}} {{.entity}} item"
    other: "{{.Count}} {{.entity}} items"

UserCount:
  ja: "{{.Count}}人のユーザー"
  en:
    one: "{{.Count}} user"
    other: "{{.Count}} users"
```

Generated code supports automatic plural form selection:

```go
// Automatically selects correct plural form based on count
msg := NewItemCount(EntityTexts.Product).WithPluralCount(1)
fmt.Println(msg.Localize("en")) // "1 Product item"

msg = NewItemCount(EntityTexts.Product).WithPluralCount(5)
fmt.Println(msg.Localize("en")) // "5 Product items"
```

## CLI Usage

### Basic Command

```bash
go-i18ngen generate [flags]
```

### Available Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `-c, --config` | string | Path to config file | `--config ./i18n.yaml` |
| `--locales` | []string | List of locales | `--locales ja,en,fr` |
| `--compound` | bool | Use compound format | `--compound` |
| `--messages` | string | Messages glob pattern | `--messages "./msg/*.yaml"` |
| `--placeholders` | string | Placeholders glob pattern | `--placeholders "./ph/*.yaml"` |
| `--output` | string | Output directory | `--output ./internal/i18n` |
| `--package` | string | Output package name | `--package i18n` |

### Examples

```bash
# Using config file (recommended)
go-i18ngen generate --config config.yaml

# Using command line flags
go-i18ngen generate \
  --locales ja,en,fr \
  --compound \
  --messages "./messages/*.yaml" \
  --placeholders "./placeholders/*.yaml" \
  --output ./internal/i18n \
  --package i18n

# Generate with custom config location
go-i18ngen generate --config ./configs/i18n-production.yaml
```

## Generated Code

### Message Structs

For each message, go-i18ngen generates:

```go
// Type-safe message struct
type EntityNotFound struct {
    Entity EntityText  // Localized placeholder
    Reason ReasonText  // Localized placeholder  
}

// Constructor function
func NewEntityNotFound(entity EntityText, reason ReasonText) EntityNotFound {
    return EntityNotFound{Entity: entity, Reason: reason}
}

// Localization method
func (m EntityNotFound) Localize(locale string) string {
    // Uses go-i18n for CLDR-compliant localization
    // Automatic fallback to default locale (first in config) if requested locale not found
}

// Interface compliance
func (m EntityNotFound) ID() string { return "EntityNotFound" }
```

### Placeholder Types

#### Text Placeholders (Localized)

```go
// Localized text placeholder
type EntityText struct {
    id string
}

func NewEntityText(id string) EntityText {
    return EntityText{id: id}
}

func (p EntityText) Localize(locale string) string {
    // Returns localized text from embedded placeholder data
}

// Utility collection for easy access
var EntityTexts = struct {
    User    EntityText
    Product EntityText
}{
    User:    EntityText{id: "user"},
    Product: EntityText{id: "product"},
}
```

#### Value Placeholders (Non-localized)

```go
// Dynamic value placeholder
type UserIdValue struct {
    Value string
}

func NewUserIdValue(value string) UserIdValue {
    return UserIdValue{Value: value}
}

func (p UserIdValue) Localize(locale string) string {
    return p.Value  // Returns value as-is
}
```

### Pluralization Support

```go
type ItemCount struct {
    Entity EntityText
    count  *int  // Internal field for pluralization
}

func NewItemCount(entity EntityText) ItemCount {
    return ItemCount{Entity: entity}
}

// Enables pluralization
func (m ItemCount) WithPluralCount(count int) ItemCount {
    m.count = &count
    return m
}

func (m ItemCount) Localize(locale string) string {
    // go-i18n automatically selects correct plural form
    // based on CLDR rules and count value
}
```

### Common Interface

All generated types implement the `Localizable` interface:

```go
type Localizable interface {
    Localize(locale string) string
    ID() string
}
```

## Advanced Features

### Type Safety Features

1. **Compile-time Parameter Validation**: Missing or incorrect parameters cause compilation errors
2. **Structured Message Building**: Cannot accidentally mix incompatible placeholder types  
3. **Locale-aware Rendering**: Automatic fallback prevents runtime errors
4. **Template Syntax Validation**: Go template parsing errors caught during generation

### Error Handling

The generated code includes robust error handling:

```go
// Example of generated error handling
func (m EntityNotFound) Localize(locale string) string {
    localizer := getLocalizer(locale)
    
    result, err := localizer.Localize(&i18n.LocalizeConfig{
        MessageID:    "EntityNotFound",
        TemplateData: templateData,
    })
    
    if err != nil {
        // Returns formatted error instead of crashing
        return fmt.Sprintf("[Localization error for %s.%s: %s]", 
                          "EntityNotFound", locale, err.Error())
    }
    
    return result
}
```

Error scenarios handled:
- **Missing Templates**: Returns formatted error message with context
- **Template Parse Errors**: Returns descriptive error details
- **Missing Placeholders**: Returns error message with missing placeholder info
- **Locale Fallback**: Gracefully falls back to primary or available locale

### Practical Usage Examples

#### Error Messages in Application Code

```go
func GetUser(id string) (*User, error) {
    user, err := repo.FindByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            msg := NewEntityNotFound(EntityTexts.User, NewUserIdValue(id))
            return nil, fmt.Errorf(msg.Localize("ja"))
        }
        return nil, err
    }
    return user, nil
}
```

#### HTTP API Response Messages

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        msg := NewValidationError(NewFieldValue("request_body"))
        http.Error(w, msg.Localize("en"), http.StatusBadRequest)
        return
    }
    
    if userExists(req.Email) {
        msg := NewUserAlreadyExists(EntityTexts.User, NewUserIdValue(req.Email))
        http.Error(w, msg.Localize("en"), http.StatusConflict)
        return
    }
    
    // ... create user
}
```

#### Logging with Localized Messages

```go
func ProcessOrder(orderID string) error {
    logger := log.With("order_id", orderID)
    
    order, err := orderRepo.FindByID(orderID)
    if err != nil {
        msg := NewEntityNotFound(EntityTexts.Order, NewOrderIdValue(orderID))
        logger.Error(msg.Localize("en"))
        return err
    }
    
    logger.Info(NewOrderProcessed(NewOrderIdValue(orderID)).Localize("en"))
    return nil
}
```

### Custom Plural Placeholders

Configure custom placeholder names for pluralization:

```yaml
# config.yaml
plural_placeholders: ["Count", "Quantity", "Amount", "Total", "ItemCount"]

# messages/inventory.yaml
StockMessage:
  ja: "在庫: {{.Quantity}}個"
  en:
    one: "Stock: {{.Quantity}} item"
    other: "Stock: {{.Quantity}} items"
```

## Integration & Development

### Build Integration

#### Using go generate

Add to your Go file:

```go
//go:generate go-i18ngen generate --config config.yaml
```

Then run:
```bash
go generate ./...
```

#### Using Makefile

```makefile
# Makefile
.PHONY: i18n
i18n:
	go-i18ngen generate --config config.yaml

.PHONY: build
build: i18n
	go build ./...
```

#### CI/CD Integration

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install i18ngen
      run: go install github.com/hacomono-lib/go-i18ngen@latest
    
    - name: Generate i18n code
      run: go-i18ngen generate --config config.yaml
    
    - name: Verify no changes
      run: git diff --exit-code
    
    - name: Run tests
      run: go test ./...
```

### Development Setup

#### Prerequisites

- Go 1.21 or later
- Make (optional)
- Docker (optional)

#### Local Development

```bash
# Clone repository
git clone https://github.com/hacomono-lib/go-i18ngen.git
cd go-i18ngen

# Setup development environment
make dev-setup      # Install required tools

# Development workflow
make fmt           # Format code
make lint          # Run linter  
make test          # Run tests with coverage
make build         # Build binary

# All-in-one checks
make check         # Run fmt + lint + test
make ci            # Run full CI pipeline locally
make pre-commit    # Quick checks before committing
```

#### Available Make Targets

| Target | Description |
|--------|-------------|
| `make help` | Show all available targets |
| `make test` | Run tests with race detector and coverage |
| `make test-short` | Run tests without race detector (faster) |
| `make coverage` | Generate HTML coverage report |
| `make lint` | Run golangci-lint |
| `make security` | Run security scan with gosec |
| `make build-all` | Build for all platforms |
| `make clean` | Clean build artifacts |

#### Docker Development

```bash
# Using Docker Compose (recommended)
docker-compose run dev     # Start development environment
docker-compose run test    # Run tests in Docker
docker-compose run lint    # Run linter in Docker

# Using Docker directly
make docker-build          # Build Docker image
docker run --rm go-i18ngen:latest --help
```

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

### Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Run tests**: `make check`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

#### Code Style Guidelines

- Follow standard Go conventions
- Run `make fmt` to format code
- Ensure `make lint` passes without warnings
- Add tests for new functionality
- Update documentation as needed

#### Tool Version Management

The project uses specific versions of development tools to ensure consistency between local development and CI:

- **golangci-lint version**: Defined in `Makefile` as `GOLANGCI_LINT_VERSION`
- **CI automatically reads**: The GitHub Actions workflow reads this version from the Makefile
- **To update versions**: Only modify the version in `Makefile` - CI will automatically use the new version

```bash
# To update golangci-lint version, edit this line in Makefile:
GOLANGCI_LINT_VERSION=v1.61.0

# Then run to update local tools:
make install-tools
```

### Project Structure

```
go-i18ngen/
├── main.go                 # CLI application entry point
├── internal/               # Internal packages
│   ├── cmd/               # CLI commands and flags
│   ├── config/            # Configuration loading and validation
│   ├── generator/         # Main code generation logic
│   ├── model/             # Data models and structures
│   ├── parser/            # YAML file parsing
│   ├── templatex/         # Template rendering and functions
│   └── utils/             # Utility functions
├── tests/                 # Integration and comprehensive tests
│   ├── comprehensive_test.go  # Main integration test suite
│   ├── integration_test.go    # Basic integration tests
│   └── pluralization_test.go  # Pluralization feature tests
├── testdata/              # Test data and working examples
├── .github/               # GitHub Actions CI/CD workflows
├── Dockerfile             # Production Docker image
├── Dockerfile.dev         # Development Docker image  
├── docker-compose.yml     # Development environment
├── Makefile              # Development commands and targets
└── .golangci.yml         # Linter configuration
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
   go version  # Should be 1.21 or later
   ```

4. **Permission denied on Docker**
   ```bash
   # Add user to docker group (Linux)
   sudo usermod -aG docker $USER
   ```

5. **Generated code has compilation errors**
   - Check message templates for valid Go template syntax
   - Ensure placeholder names are valid Go identifiers
   - Verify all referenced placeholders are defined

### Limitations

1. **Go Template Syntax**: Must use valid Go template syntax in message templates
2. **Identifier Names**: Placeholder kinds and IDs must be valid Go identifiers
3. **File Naming**: For simple format, files must follow `name.locale.ext` pattern
4. **Template Fields**: Fields in templates must match available placeholder definitions
5. **Static Generation**: Changes require regeneration; no runtime template loading
6. **Suffix Characters**: Suffix names must be valid Go identifier characters

## Examples

The `testdata/` directory contains working examples that demonstrate various features:

- **Basic message localization**
- **Pluralization with CLDR rules**
- **Template functions usage**
- **Suffix notation for complex messages**
- **Multi-language support**
- **Error handling patterns**

### Real-world Example Structure

```
project/
├── i18n/
│   ├── config.yaml
│   ├── messages/
│   │   ├── errors.yaml      # Error messages
│   │   ├── ui.yaml          # UI labels and text
│   │   └── notifications.yaml # Notification messages
│   └── placeholders/
│       ├── entities.yaml    # User, Product, Order, etc.
│       ├── actions.yaml     # Create, Update, Delete, etc.
│       └── statuses.yaml    # Active, Inactive, Pending, etc.
├── generated/
│   └── i18n.gen.go         # Generated type-safe code
└── main.go
```

This structure provides a clear separation of concerns and makes it easy to maintain localized content as your application grows.

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](#contributing) for details on how to get started.