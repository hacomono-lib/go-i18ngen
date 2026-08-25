package templatex

// The generated code (tests/i18n.gen.go, produced by `go generate` and not
// committed) imports these packages, but no committed .go source does — the
// import lines live inside go-i18n.gotmpl, which `go mod tidy` can't see.
// Without this file, `go mod tidy` treats them as unused and drops them from
// go.mod/go.sum, breaking the next `go generate`.
import (
	_ "github.com/nicksnyder/go-i18n/v2/i18n"
	_ "golang.org/x/text/language"
)
