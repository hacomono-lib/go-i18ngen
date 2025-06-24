package cmd

import (
	"github.com/hacomono-lib/go-i18ngen/internal/config"
	"github.com/hacomono-lib/go-i18ngen/internal/generator"

	"github.com/spf13/cobra"
)

var (
	configPath string
	flags      Flags
)

// NewGenerateCommand creates and returns the generate command
func NewGenerateCommand() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate i18n message and placeholder code",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}
			merged := MergeConfig(cfg, &flags)
			return generator.Run(merged)
		},
	}

	genCmd.Flags().StringVarP(&configPath, "config", "c", "i18ngen.yaml", "path to config file")
	genCmd.Flags().StringSliceVar(&flags.Locales, "locales", nil, "list of locales (e.g. ja,en)")
	genCmd.Flags().BoolVar(&flags.Compound, "compound", false, "use compound format")
	genCmd.Flags().StringVar(&flags.MessagesGlob, "messages", "", "messages glob pattern")
	genCmd.Flags().StringVar(&flags.PlaceholdersGlob, "placeholders", "", "placeholders glob pattern")
	genCmd.Flags().StringVar(&flags.OutputDir, "output", "", "output directory")
	genCmd.Flags().StringVar(&flags.OutputPackage, "package", "", "output package name")

	return genCmd
}

// MergeConfig merges CLI flags with config file, prioritizing flags
func MergeConfig(cfg *config.Config, flags *Flags) *config.Config {
	if len(flags.Locales) > 0 {
		cfg.Locales = flags.Locales
	}
	if flags.Compound {
		cfg.Compound = flags.Compound
	}
	if flags.MessagesGlob != "" {
		cfg.MessagesGlob = flags.MessagesGlob
	}
	if flags.PlaceholdersGlob != "" {
		cfg.PlaceholdersGlob = flags.PlaceholdersGlob
	}
	if flags.OutputDir != "" {
		cfg.OutputDir = flags.OutputDir
	}
	if flags.OutputPackage != "" {
		cfg.OutputPackage = flags.OutputPackage
	}
	return cfg
}
