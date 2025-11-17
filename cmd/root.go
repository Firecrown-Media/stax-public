package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	verbose    bool
	debug      bool
	quiet      bool
	noColor    bool
	projectDir string
	cfg        *config.Config
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stax",
	Short: "A CLI tool for WordPress development with WPEngine integration",
	Long: `Stax is a powerful CLI tool that streamlines WordPress development workflows.
It supports both single WordPress sites and multisite networks, leverages DDEV
for container orchestration, and provides seamless integration with multiple
hosting providers.

Features:
  - Automated WordPress setup (single site or multisite)
  - WPEngine database sync and file management
  - Remote media proxying (BunnyCDN + WPEngine)
  - DDEV container management
  - Secure credential storage via macOS Keychain
  - Team-friendly configuration management`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Commands that don't require .stax.yml config
		skipConfigCommands := []string{"setup", "version", "completion", "man", "list", "doctor", "init", "start", "stop", "restart", "status", "wpengine", "config"}
		for _, skipCmd := range skipConfigCommands {
			if cmd.Name() == skipCmd {
				// Still initialize UI
				ui.SetVerbose(verbose)
				ui.SetDebug(debug)
				ui.SetQuiet(quiet)
				ui.SetNoColor(noColor)
				return nil
			}
		}

		// Initialize UI based on flags
		ui.SetVerbose(verbose)
		ui.SetDebug(debug)
		ui.SetQuiet(quiet)
		ui.SetNoColor(noColor)

		// Load configuration
		var err error
		cfg, err = config.Load(cfgFile, projectDir)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .stax.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVar(&projectDir, "project-dir", "", "project directory (default is current directory)")

	// Bind flags to viper
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic(fmt.Sprintf("failed to bind verbose flag: %v", err))
	}
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(fmt.Sprintf("failed to bind debug flag: %v", err))
	}
	if err := viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		panic(fmt.Sprintf("failed to bind quiet flag: %v", err))
	}
	if err := viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")); err != nil {
		panic(fmt.Sprintf("failed to bind no-color flag: %v", err))
	}

	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s " .}}{{end}}{{printf "version %%s" .Version}}
Git Commit: %s
Build Date: %s
`, GitCommit, BuildDate))

	// Customize usage template with status indicators
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Status Indicators:
  ✓ Fully implemented and tested
  ⚠ Partial implementation or workaround available
  🚧 Under development - not yet implemented

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

// getConfig returns the loaded configuration
// This is a helper function for subcommands to access the config
func getConfig() *config.Config {
	return cfg
}

// getProjectDir returns the absolute project directory
func getProjectDir() string {
	dir := projectDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "."
		}
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		// Fallback to original if absolute path conversion fails
		return dir
	}
	return absDir
}
