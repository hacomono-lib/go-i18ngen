package cmd

type Flags struct {
	Locales          []string
	Compound         bool
	MessagesGlob     string
	PlaceholdersGlob string
	OutputDir        string
	OutputPackage    string
}
