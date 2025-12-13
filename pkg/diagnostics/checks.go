package diagnostics

// CheckResult represents the result of a diagnostic check
type CheckResult struct {
	Name       string
	Status     CheckStatus
	Message    string
	Suggestion string
	Details    map[string]string
	Category   string // Category for grouping checks
	CanAutoFix bool   // Whether this check can be auto-fixed
	FixApplied bool   // Whether a fix was applied
}

// CheckStatus represents the status of a check
type CheckStatus string

const (
	StatusPass    CheckStatus = "pass"
	StatusWarning CheckStatus = "warning"
	StatusFail    CheckStatus = "fail"
	StatusSkip    CheckStatus = "skip"
)

// DiagnosticReport contains all diagnostic check results
type DiagnosticReport struct {
	Checks      []CheckResult
	Summary     Summary
	ProjectPath string
	Verbose     bool
	AutoFix     bool
	Categories  map[string][]CheckResult // Checks grouped by category
}

// Summary provides a summary of check results
type Summary struct {
	Total    int
	Passed   int
	Warnings int
	Failed   int
	Skipped  int
	Fixed    int
}

// RunAllChecks runs all diagnostic checks
func RunAllChecks(projectPath string, verbose bool, autoFix bool) (*DiagnosticReport, error) {
	report := &DiagnosticReport{
		ProjectPath: projectPath,
		Checks:      []CheckResult{},
		Verbose:     verbose,
		AutoFix:     autoFix,
		Categories:  make(map[string][]CheckResult),
	}

	// System requirements checks
	report.Checks = append(report.Checks, CheckGit())
	report.Checks = append(report.Checks, CheckDocker())
	report.Checks = append(report.Checks, CheckDDEV())
	report.Checks = append(report.Checks, CheckMemory())
	report.Checks = append(report.Checks, CheckRequiredCommands())
	report.Checks = append(report.Checks, CheckGo())

	// Project configuration checks
	report.Checks = append(report.Checks, CheckStaxConfig(projectPath))
	report.Checks = append(report.Checks, CheckDDEVConfig(projectPath))

	// Credential checks
	report.Checks = append(report.Checks, CheckCredentials(projectPath))
	report.Checks = append(report.Checks, CheckSSHKey())
	report.Checks = append(report.Checks, CheckGitHubToken())

	// Network checks
	report.Checks = append(report.Checks, CheckPorts())
	report.Checks = append(report.Checks, CheckWPEngineAPI())
	report.Checks = append(report.Checks, CheckWPEngineSSH())
	report.Checks = append(report.Checks, CheckGitHubAPI())
	report.Checks = append(report.Checks, CheckInternetConnectivity())

	// Environment checks
	report.Checks = append(report.Checks, CheckDiskSpace(projectPath))

	// Service health checks - with auto-fix support
	ddevStatus := CheckDDEVStatus(projectPath)
	if autoFix && ddevStatus.CanAutoFix && (ddevStatus.Status == StatusWarning || ddevStatus.Status == StatusFail) {
		ddevStatus = FixDDEVStatus(projectPath, ddevStatus)
	}
	report.Checks = append(report.Checks, ddevStatus)

	report.Checks = append(report.Checks, CheckDatabaseConnectivity(projectPath))
	report.Checks = append(report.Checks, CheckWordPressInstallation(projectPath))

	// Group checks by category
	report.groupByCategory()

	// Calculate summary
	report.Summary = calculateSummary(report.Checks)

	return report, nil
}

// groupByCategory groups check results by category
func (r *DiagnosticReport) groupByCategory() {
	r.Categories = make(map[string][]CheckResult)
	for _, check := range r.Checks {
		category := check.Category
		if category == "" {
			category = "Other"
		}
		r.Categories[category] = append(r.Categories[category], check)
	}
}

// calculateSummary calculates the summary of check results
func calculateSummary(checks []CheckResult) Summary {
	summary := Summary{
		Total: len(checks),
	}

	for _, check := range checks {
		if check.FixApplied {
			summary.Fixed++
		}
		switch check.Status {
		case StatusPass:
			summary.Passed++
		case StatusWarning:
			summary.Warnings++
		case StatusFail:
			summary.Failed++
		case StatusSkip:
			summary.Skipped++
		}
	}

	return summary
}

// HasCriticalFailures returns true if there are any critical failures
func (r *DiagnosticReport) HasCriticalFailures() bool {
	return r.Summary.Failed > 0
}

// HasWarnings returns true if there are any warnings
func (r *DiagnosticReport) HasWarnings() bool {
	return r.Summary.Warnings > 0
}

// IsHealthy returns true if all checks passed
func (r *DiagnosticReport) IsHealthy() bool {
	return r.Summary.Failed == 0 && r.Summary.Warnings == 0
}
