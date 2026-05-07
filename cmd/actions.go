package cmd

import (
	"github.com/firecrown-media/stax/pkg/actions"
	"github.com/spf13/cobra"
)

var (
	actionsProductionBranch string
	actionsStagingBranch    string
	actionsProdInstall      string
	actionsStageInstall     string
	actionsForce            bool
)

// actionsCmd represents the actions command group
var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "GitHub Actions workflow management",
	Long:  `Commands for managing GitHub Actions workflows for WordPress deployments.`,
}

// actionsSetupCmd sets up GitHub Actions for WPEngine deployment
var actionsSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up GitHub Actions for WPEngine deployment",
	Long: `Generate GitHub Actions workflow files for deploying to WPEngine.

This command creates:
- .github/workflows/deploy.yml - Deployment workflow
- .github/CODEOWNERS template (if it doesn't exist)
- Provides instructions for branch protection

Example:
  # Basic setup (uses main for production)
  stax actions setup

  # With staging branch
  stax actions setup --production main --staging develop

  # Specify WPEngine install names
  stax actions setup --prod-install mysite-prod --stage-install mysite-staging`,
	RunE: runActionsSetup,
}

func init() {
	rootCmd.AddCommand(actionsCmd)
	actionsCmd.AddCommand(actionsSetupCmd)

	actionsSetupCmd.Flags().StringVar(&actionsProductionBranch, "production", "main", "Branch for production deployment")
	actionsSetupCmd.Flags().StringVar(&actionsStagingBranch, "staging", "", "Branch for staging deployment (optional)")
	actionsSetupCmd.Flags().StringVar(&actionsProdInstall, "prod-install", "", "WPEngine production install name")
	actionsSetupCmd.Flags().StringVar(&actionsStageInstall, "stage-install", "", "WPEngine staging install name")
	actionsSetupCmd.Flags().BoolVar(&actionsForce, "force", false, "Overwrite existing workflow files")
}

func runActionsSetup(cmd *cobra.Command, args []string) error {
	return actions.Setup(actions.SetupOptions{
		ProductionBranch: actionsProductionBranch,
		StagingBranch:    actionsStagingBranch,
		ProdInstall:      actionsProdInstall,
		StageInstall:     actionsStageInstall,
		Force:            actionsForce,
		ProjectDir:       getProjectDir(),
	})
}
