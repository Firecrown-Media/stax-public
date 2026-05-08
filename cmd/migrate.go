package cmd

import (
	"path/filepath"

	"github.com/firecrown-media/stax/pkg/migration"
	_ "github.com/firecrown-media/stax/pkg/migration/providers/vip"
	_ "github.com/firecrown-media/stax/pkg/migration/providers/wpengine"
	"github.com/firecrown-media/stax/pkg/provider"
	"github.com/firecrown-media/stax/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	migDestination   string
	migLocalPath     string
	migVIPRepoPath   string
	migExportPath    string // --output on export
	migSQLPath       string // --sql on import and report
	migOutputPath    string // --output on report
	migPullDryRun    bool
	migImportDryRun  bool
	migSlug          string
	migThemesOnly    bool
	migPluginsOnly   bool
	migMuPluginsOnly bool
	migSeverity      int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate WordPress sites between hosting providers",
	Long: `Orchestrate migration from one hosting provider to another.

Source is determined by the provider: field in .stax.yml.
Destination is set via migration.destination in .stax.yml or the --destination flag.

Requires migration.destination to be set:

  migration:
    destination: vip`,
}

var migratePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull files from source provider to local wp-content/",
	Long:  `Download wp-content from the source provider (uploads excluded by default).`,
	Example: `  stax migrate pull
  stax migrate pull --themes-only
  stax migrate pull --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		p, err := provider.NewAuthenticatedProvider(cfg)
		if err != nil {
			return err
		}
		return migration.Pull(p, cfg, migration.PullOptions{
			ThemesOnly:    migThemesOnly,
			PluginsOnly:   migPluginsOnly,
			MuPluginsOnly: migMuPluginsOnly,
			DryRun:        migPullDryRun,
			ProjectDir:    getProjectDir(),
		})
	},
}

var migrateExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database from source with VIP-compatible flags",
	Long:  `Export the source database with mysqldump flags required by WordPress VIP.`,
	Example: `  stax migrate export
  stax migrate export --output=mysite-export.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		p, err := provider.NewAuthenticatedProvider(cfg)
		if err != nil {
			return err
		}
		return migration.Export(p, cfg, migration.ExportOptions{
			OutputPath: migExportPath,
		})
	},
}

var migrateAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run phpcs compatibility audit against the VIP ruleset",
	Long: `Scan plugins, themes, and client-mu-plugins in the local wp-content
directory against the WordPress-VIP-Go phpcs ruleset.`,
	Example: `  stax migrate audit
  stax migrate audit --path=../astronomy/wp-content
  stax migrate audit --severity=5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		localPath := migLocalPath
		if localPath == "" {
			localPath = filepath.Join(getProjectDir(), "wp-content")
		}
		report, err := migration.Audit(cfg, localPath, migration.AuditOptions{
			Severity: migSeverity,
		})
		if err != nil {
			return err
		}
		printAuditSummary(report)
		return nil
	},
}

var migrateCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Diff local files against the destination VIP repo",
	Long: `Compare plugins, themes, and client-mu-plugins between the downloaded
WPEngine wp-content and the local VIP repo checkout.`,
	Example: `  stax migrate compare --vip-repo=../vip-repo
  stax migrate compare --path=../wpe/wp-content --vip-repo=../vip-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		localPath := migLocalPath
		if localPath == "" {
			localPath = filepath.Join(getProjectDir(), "wp-content")
		}
		result, err := migration.Compare(cfg, localPath, migVIPRepoPath)
		if err != nil {
			return err
		}
		printCompareResult(result)
		return nil
	},
}

var migrateImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Push DB/media into destination (wraps VIP CLI)",
	Long:  `Validate and import the SQL dump into the VIP destination using the VIP CLI.`,
	Example: `  stax migrate import --sql=.stax/mysite-export.sql
  stax migrate import --sql=export.sql --slug=my-vip-env`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		return migration.Import(cfg, migSQLPath, migration.ImportOptions{
			DryRun: migImportDryRun,
			Slug:   migSlug,
		})
	},
}

var migrateReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate combined audit + compare report as markdown",
	Long:  `Run audit and compare, then write a combined migration report to .stax/migration-report.md.`,
	Example: `  stax migrate report --vip-repo=../vip-repo
  stax migrate report --path=../wpe/wp-content --vip-repo=../vip-repo --sql=.stax/mysite-export.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		p, err := provider.NewAuthenticatedProvider(cfg)
		if err != nil {
			return err
		}
		localPath := migLocalPath
		if localPath == "" {
			localPath = filepath.Join(getProjectDir(), "wp-content")
		}
		return migration.Report(p, cfg, migration.ReportOptions{
			LocalPath:   localPath,
			VIPRepoPath: migVIPRepoPath,
			SQLPath:     migSQLPath,
			OutputPath:  migOutputPath,
		})
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration state for this project",
	Long:  `Print migration configuration and the presence of key artifacts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigForCommand()
		if err != nil {
			return err
		}
		if migDestination != "" {
			cfg.Migration.Destination = migDestination
		}
		return migration.Status(cfg)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migratePullCmd)
	migrateCmd.AddCommand(migrateExportCmd)
	migrateCmd.AddCommand(migrateAuditCmd)
	migrateCmd.AddCommand(migrateCompareCmd)
	migrateCmd.AddCommand(migrateImportCmd)
	migrateCmd.AddCommand(migrateReportCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	migrateCmd.PersistentFlags().StringVar(&migDestination, "destination", "", "override migration.destination from config")

	migratePullCmd.Flags().BoolVar(&migThemesOnly, "themes-only", false, "pull only themes directory")
	migratePullCmd.Flags().BoolVar(&migPluginsOnly, "plugins-only", false, "pull only plugins directory")
	migratePullCmd.Flags().BoolVar(&migMuPluginsOnly, "mu-plugins-only", false, "pull only mu-plugins directory")
	migratePullCmd.Flags().BoolVar(&migPullDryRun, "dry-run", false, "show what would be transferred without pulling")

	migrateExportCmd.Flags().StringVar(&migExportPath, "output", "", "path for the SQL dump file (default: .stax/<install>-export.sql)")

	migrateAuditCmd.Flags().StringVar(&migLocalPath, "path", "", "path to wp-content directory (default: <project>/wp-content)")
	migrateAuditCmd.Flags().IntVar(&migSeverity, "severity", 1, "minimum phpcs severity level (1-5)")

	migrateCompareCmd.Flags().StringVar(&migLocalPath, "path", "", "path to downloaded wp-content (default: <project>/wp-content)")
	migrateCompareCmd.Flags().StringVar(&migVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
	_ = migrateCompareCmd.MarkFlagRequired("vip-repo")

	migrateImportCmd.Flags().StringVar(&migSQLPath, "sql", "", "path to the SQL dump file")
	migrateImportCmd.Flags().StringVar(&migSlug, "slug", "", "VIP environment slug (passed to --slug)")
	migrateImportCmd.Flags().BoolVar(&migImportDryRun, "dry-run", false, "dry run (no changes)")
	_ = migrateImportCmd.MarkFlagRequired("sql")

	migrateReportCmd.Flags().StringVar(&migLocalPath, "path", "", "path to wp-content (default: <project>/wp-content)")
	migrateReportCmd.Flags().StringVar(&migVIPRepoPath, "vip-repo", "", "path to local VIP repo checkout")
	migrateReportCmd.Flags().StringVar(&migSQLPath, "sql", "", "path to SQL dump file (optional)")
	migrateReportCmd.Flags().StringVar(&migOutputPath, "output", "", "output path for report (default: .stax/migration-report.md)")
	_ = migrateReportCmd.MarkFlagRequired("vip-repo")
}

func printAuditSummary(report *migration.AuditReport) {
	ui.Section("phpcs Audit Summary")
	ui.Info("Files scanned:  %d", len(report.Files))
	ui.Info("Total errors:   %d", report.TotalErrors)
	ui.Info("Total warnings: %d", report.TotalWarnings)
	for _, f := range report.Files {
		if f.Errors > 0 || f.Warnings > 0 {
			ui.Warning("  %s — %d errors, %d warnings", f.FilePath, f.Errors, f.Warnings)
		}
	}
}

func printCompareResult(result *migration.CompareResult) {
	ui.Section("File Comparison")
	if len(result.MissingFromVIP) == 0 && len(result.MissingFromWPE) == 0 {
		ui.Success("No differences found")
		return
	}
	if len(result.MissingFromVIP) > 0 {
		ui.Warning("Present in WPEngine, missing from VIP repo:")
		for _, p := range result.MissingFromVIP {
			ui.Info("  - %s", p)
		}
	}
	if len(result.MissingFromWPE) > 0 {
		ui.Info("Present in VIP repo, missing from WPEngine:")
		for _, p := range result.MissingFromWPE {
			ui.Info("  - %s", p)
		}
	}
}
