package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/firecrown-media/stax/pkg/config"
	"github.com/firecrown-media/stax/pkg/git"
	"github.com/firecrown-media/stax/pkg/prompts"
	"github.com/firecrown-media/stax/pkg/ui"
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
	ui.PrintHeader("Setting Up GitHub Actions")

	projectDir := getProjectDir()

	// Check if this is a Git repository
	if !git.IsGitRepository(projectDir) {
		return fmt.Errorf("not a git repository - run 'git init' or 'stax repo init' first")
	}

	// Try to load existing Stax config for install names
	cfg, _ := config.Load("", projectDir)

	// Determine install names
	prodInstall := actionsProdInstall
	stageInstall := actionsStageInstall

	if prodInstall == "" && cfg != nil && cfg.WPEngine.Install != "" {
		prodInstall = cfg.WPEngine.Install
	}

	if prodInstall == "" {
		if !prompts.IsInteractive() {
			return fmt.Errorf("--prod-install is required in non-interactive mode")
		}
		var err error
		prodInstall, err = prompts.WPEngineInstallPrompt("WPEngine production install name")
		if err != nil {
			return err
		}
	}

	// Ask about staging if not provided
	if actionsStagingBranch != "" && stageInstall == "" {
		if prompts.IsInteractive() {
			var err error
			stageInstall, err = prompts.SafePromptInput("WPEngine staging install name (optional)", "", true)
			if err != nil {
				return err
			}
		}
	}

	// Create .github/workflows directory
	workflowDir := filepath.Join(projectDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	// Generate workflow file
	ui.Section("Generating Deployment Workflow")
	workflowPath := filepath.Join(workflowDir, "deploy.yml")

	if !actionsForce {
		if _, err := os.Stat(workflowPath); err == nil {
			ui.Warning("Workflow file already exists at .github/workflows/deploy.yml")
			ui.Info("Use --force to overwrite")
			return nil
		}
	}

	workflowData := WorkflowTemplateData{
		ProductionBranch:  actionsProductionBranch,
		StagingBranch:     actionsStagingBranch,
		ProductionInstall: prodInstall,
		StagingInstall:    stageInstall,
		HasStaging:        actionsStagingBranch != "" && stageInstall != "",
	}

	if err := generateWorkflowFile(workflowPath, workflowData); err != nil {
		return fmt.Errorf("failed to generate workflow: %w", err)
	}
	ui.Success("Created .github/workflows/deploy.yml")

	// Generate CODEOWNERS if it doesn't exist
	codeownersPath := filepath.Join(projectDir, ".github", "CODEOWNERS")
	if _, err := os.Stat(codeownersPath); os.IsNotExist(err) {
		if err := generateCodeowners(codeownersPath); err != nil {
			ui.Warning(fmt.Sprintf("Failed to create CODEOWNERS: %v", err))
		} else {
			ui.Success("Created .github/CODEOWNERS template")
		}
	}

	// Display setup instructions
	displayActionsInstructions(workflowData)

	return nil
}

// WorkflowTemplateData contains data for the workflow template
type WorkflowTemplateData struct {
	ProductionBranch  string
	StagingBranch     string
	ProductionInstall string
	StagingInstall    string
	HasStaging        bool
}

func generateWorkflowFile(path string, data WorkflowTemplateData) error {
	tmpl := `name: Deploy to WPEngine

on:
  push:
    branches:
      - {{ .ProductionBranch }}
{{- if .HasStaging }}
      - {{ .StagingBranch }}
{{- end }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci
        continue-on-error: true

      - name: Build assets
        run: npm run build
        continue-on-error: true

      - name: Deploy to WPEngine
        uses: wpengine/github-action-wpe-site-deploy@v3
        with:
          WPE_SSHG_KEY_PRIVATE: ${{ "{{" }} secrets.WPE_SSHG_KEY_PRIVATE {{ "}}" }}
{{- if .HasStaging }}
          WPE_ENV: ${{ "{{" }} github.ref == 'refs/heads/{{ .ProductionBranch }}' && '{{ .ProductionInstall }}' || '{{ .StagingInstall }}' {{ "}}" }}
{{- else }}
          WPE_ENV: {{ .ProductionInstall }}
{{- end }}
          SRC_PATH: "wp-content/"
          REMOTE_PATH: "wp-content/"
          PHP_LINT: true
          FLAGS: -azvr --inplace --delete --exclude=".*" --exclude="node_modules/" --exclude="*.sql"
`

	t, err := template.New("workflow").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func generateCodeowners(path string) error {
	content := `# CODEOWNERS - Define code owners for pull request reviews
# See: https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners

# Default owners for everything in the repo
# * @your-org/maintainers

# Theme owners
# /wp-content/themes/ @your-org/frontend-team

# Plugin owners
# /wp-content/plugins/ @your-org/backend-team

# MU-plugins typically need senior review
# /wp-content/mu-plugins/ @your-org/senior-devs
`

	return os.WriteFile(path, []byte(content), 0644)
}

func displayActionsInstructions(data WorkflowTemplateData) {
	ui.Section("Next Steps")

	fmt.Println()
	ui.Info("1. Add the WPEngine SSH private key as a GitHub secret:")
	fmt.Println()
	fmt.Println("   a. Go to your GitHub repository")
	fmt.Println("   b. Navigate to Settings → Secrets and variables → Actions")
	fmt.Println("   c. Click 'New repository secret'")
	fmt.Println("   d. Name: WPE_SSHG_KEY_PRIVATE")
	fmt.Println("   e. Value: Your WPEngine SSH private key (entire contents)")
	fmt.Println()
	ui.Info("   Get your key from: https://my.wpengine.com/ → SSH Gateway")
	fmt.Println()

	ui.Info("2. Configure branch protection (recommended):")
	fmt.Println()
	fmt.Printf("   For '%s' branch:\n", data.ProductionBranch)
	fmt.Println("   - Require pull request before merging")
	fmt.Println("   - Require approvals: 1+")
	fmt.Println("   - Require status checks to pass")
	fmt.Println("   - Require conversation resolution")
	fmt.Println()
	if data.HasStaging {
		fmt.Printf("   For '%s' branch:\n", data.StagingBranch)
		fmt.Println("   - Require status checks to pass")
		fmt.Println()
	}

	ui.Info("3. Update CODEOWNERS file:")
	fmt.Println()
	fmt.Println("   Edit .github/CODEOWNERS to set appropriate code owners")
	fmt.Println()

	ui.Info("4. Commit and push the workflow:")
	fmt.Println()
	fmt.Println("   git add .github/")
	fmt.Println("   git commit -m \"chore: add GitHub Actions deployment workflow\"")
	fmt.Println("   git push")
	fmt.Println()

	// Build environment info
	envInfo := []string{fmt.Sprintf("%s → %s", data.ProductionBranch, data.ProductionInstall)}
	if data.HasStaging {
		envInfo = append(envInfo, fmt.Sprintf("%s → %s", data.StagingBranch, data.StagingInstall))
	}

	ui.Section("Deployment Configuration")
	fmt.Println()
	for _, env := range envInfo {
		fmt.Printf("   %s\n", env)
	}
	fmt.Println()
	fmt.Println("   Pushes to these branches will automatically deploy to WPEngine.")
	fmt.Println()

	// Documentation links
	ui.Section("Documentation")
	fmt.Println()
	fmt.Println("   WPEngine GitHub Action: https://github.com/wpengine/github-action-wpe-site-deploy")
	fmt.Println("   Branch Protection: " + git.GetBranchProtectionGuideURL())
	fmt.Println("   CODEOWNERS: https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners")
	fmt.Println()
}

// Helper function to escape strings for YAML
func escapeYAML(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
