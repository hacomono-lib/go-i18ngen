// Package cmd provides command line interface for the i18n code generator.
package cmd

type Flags struct {
	Locales          []string
	Compound         bool
	MessagesGlob     string
	PlaceholdersGlob string
	OutputDir        string
	OutputPackage    string
	Backend          string
}
