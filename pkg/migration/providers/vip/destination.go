package vip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecrown-media/stax/pkg/migration"
)

func init() {
	migration.RegisterDestination("vip", func(repoPath string) migration.Destination {
		return NewVIPDestination(repoPath)
	})
}

// VIPDestination implements migration.Destination for WordPress VIP.
type VIPDestination struct {
	repoPath string // local path to VIP repo checkout (for CompareFiles)
}

// NewVIPDestination constructs a VIPDestination.
// repoPath is the local VIP repo checkout used by CompareFiles.
func NewVIPDestination(repoPath string) *VIPDestination {
	return &VIPDestination{repoPath: repoPath}
}

// Audit runs phpcs with the WordPress-VIP-Go ruleset against plugins, themes,
// and client-mu-plugins under localPath.
func (d *VIPDestination) Audit(localPath string, opts migration.AuditOptions) (*migration.AuditReport, error) {
	if _, err := exec.LookPath("phpcs"); err != nil {
		return nil, fmt.Errorf("phpcs not found: install with 'composer global require automattic/vip-coding-standards'")
	}

	severity := opts.Severity
	if severity == 0 {
		severity = 1
	}

	targets := []string{
		filepath.Join(localPath, "plugins"),
		filepath.Join(localPath, "themes"),
		filepath.Join(localPath, "client-mu-plugins"),
	}

	report := &migration.AuditReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	for _, target := range targets {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		}

		args := []string{
			"--standard=WordPress-VIP-Go",
			fmt.Sprintf("--severity=%d", severity),
			"--report=json",
			target,
		}

		// phpcs exits non-zero when violations are found but still produces valid
		// JSON on stdout. Capture stdout and stderr separately so we can tell
		// "found violations" from "phpcs itself blew up" (missing standards,
		// memory exhaustion, autoload failure). .Output() drops stderr and
		// silently treats a fatal as zero results — see the related stax bug
		// memory note.
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd := exec.Command("phpcs", args...)
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		runErr := cmd.Run()
		if runErr != nil {
			if _, ok := runErr.(*exec.ExitError); !ok {
				return nil, fmt.Errorf("phpcs failed on %s: %w (stderr: %s)",
					target, runErr, strings.TrimSpace(stderrBuf.String()))
			}
		}

		out := stdoutBuf.Bytes()
		// If phpcs produced no JSON output but wrote to stderr, surface the
		// stderr content instead of silently treating it as zero violations.
		// JSON output always starts with '{' once leading whitespace is stripped.
		trimmed := bytes.TrimSpace(out)
		if len(trimmed) == 0 || trimmed[0] != '{' {
			stderrStr := strings.TrimSpace(stderrBuf.String())
			if stderrStr != "" {
				return nil, fmt.Errorf("phpcs produced no JSON output for %s; stderr: %s",
					target, stderrStr)
			}
		}

		results, err := parsePHPCSReport(out)
		if err != nil {
			return nil, fmt.Errorf("failed to parse phpcs output for %s: %w", target, err)
		}
		report.Files = append(report.Files, results...)
	}

	for _, f := range report.Files {
		report.TotalErrors += f.Errors
		report.TotalWarnings += f.Warnings
	}

	return report, nil
}

// ValidateDatabase runs vip import validate-sql against the SQL file at path.
func (d *VIPDestination) ValidateDatabase(path string) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	out, err := exec.Command("vip", "import", "validate-sql", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("database validation failed: %w\n%s", err, string(out))
	}
	return nil
}

// ImportDatabase runs vip import sql on the SQL file at path.
func (d *VIPDestination) ImportDatabase(path string, opts migration.ImportOptions) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	args := []string{"import", "sql", path}
	if opts.Slug != "" {
		args = append(args, "--slug="+opts.Slug)
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	out, err := exec.Command("vip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("database import failed: %w\n%s", err, string(out))
	}
	return nil
}

// ImportMedia runs vip import media.
func (d *VIPDestination) ImportMedia(opts migration.ImportOptions) error {
	if _, err := exec.LookPath("vip"); err != nil {
		return fmt.Errorf("VIP CLI not found: install with 'npm install -g @automattic/vip'")
	}
	args := []string{"import", "media"}
	if opts.Slug != "" {
		args = append(args, "--slug="+opts.Slug)
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	out, err := exec.Command("vip", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("media import failed: %w\n%s", err, string(out))
	}
	return nil
}

// CompareFiles diffs top-level directories between the WPEngine wp-content
// (localPath) and the VIP repo checkout (d.repoPath) across plugins, themes,
// and client-mu-plugins.
func (d *VIPDestination) CompareFiles(localPath string) (*migration.CompareResult, error) {
	result := &migration.CompareResult{
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	for _, dir := range []string{"plugins", "themes", "client-mu-plugins"} {
		wpePath := filepath.Join(localPath, dir)
		vipPath := filepath.Join(d.repoPath, dir)

		wpeItems, err := listTopLevel(wpePath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", wpePath, err)
		}
		vipItems, err := listTopLevel(vipPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", vipPath, err)
		}

		wpeSet := toSet(wpeItems)
		vipSet := toSet(vipItems)

		for item := range wpeSet {
			if !vipSet[item] {
				result.MissingFromVIP = append(result.MissingFromVIP, dir+"/"+item)
			}
		}
		for item := range vipSet {
			if !wpeSet[item] {
				result.MissingFromWPE = append(result.MissingFromWPE, dir+"/"+item)
			}
		}
	}

	return result, nil
}

// ---- phpcs JSON parsing ----

type phpcsOutput struct {
	Files map[string]phpcsFile `json:"files"`
}

type phpcsFile struct {
	Errors   int            `json:"errors"`
	Warnings int            `json:"warnings"`
	Messages []phpcsMessage `json:"messages"`
}

type phpcsMessage struct {
	Message  string `json:"message"`
	Source   string `json:"source"`
	Severity int    `json:"severity"`
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func parsePHPCSReport(data []byte) ([]migration.FileAuditResult, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var out phpcsOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	results := make([]migration.FileAuditResult, 0, len(out.Files))
	for filePath, file := range out.Files {
		result := migration.FileAuditResult{
			FilePath: filePath,
			Errors:   file.Errors,
			Warnings: file.Warnings,
		}
		for _, msg := range file.Messages {
			result.Messages = append(result.Messages, migration.AuditMessage{
				Line:     msg.Line,
				Column:   msg.Column,
				Severity: msg.Severity,
				Type:     msg.Type,
				Message:  msg.Message,
				Source:   msg.Source,
			})
		}
		results = append(results, result)
	}
	return results, nil
}

// ---- file comparison helpers ----

func listTopLevel(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
