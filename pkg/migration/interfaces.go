package migration

// PullOptions configures a file pull from the source provider.
type PullOptions struct {
	ExcludeUploads bool
	ThemesOnly     bool
	PluginsOnly    bool
	MuPluginsOnly  bool
	DryRun         bool
	ProjectDir     string
}

// ExportOptions configures a database export from the source provider.
type ExportOptions struct {
	OutputPath string // local file path where the SQL dump is written
}

// AuditOptions configures a phpcs compatibility audit.
type AuditOptions struct {
	Severity int // minimum phpcs severity (1–5); 0 defaults to 1
}

// AuditMessage is a single phpcs finding.
type AuditMessage struct {
	Line     int
	Column   int
	Severity int
	Type     string // "ERROR" or "WARNING"
	Message  string
	Source   string // phpcs sniff identifier
}

// FileAuditResult holds phpcs findings for a single file.
type FileAuditResult struct {
	FilePath string
	Errors   int
	Warnings int
	Messages []AuditMessage
}

// AuditReport aggregates phpcs findings across all scanned paths.
type AuditReport struct {
	Files         []FileAuditResult
	TotalErrors   int
	TotalWarnings int
	GeneratedAt   string
}

// ImportOptions configures a VIP import operation.
type ImportOptions struct {
	DryRun bool
	Slug   string // VIP environment slug (passed to --slug flag)
}

// CompareResult holds the file diff between WPEngine and a VIP repo.
type CompareResult struct {
	MissingFromVIP []string // present in WPEngine wp-content, absent from VIP repo
	MissingFromWPE []string // present in VIP repo, absent from WPEngine wp-content
	GeneratedAt    string
}

// Source is the platform being migrated away from.
type Source interface {
	PullFiles(opts PullOptions) error
	ExportDatabase(opts ExportOptions) error
}

// Destination is the platform being migrated to.
type Destination interface {
	Audit(localPath string, opts AuditOptions) (*AuditReport, error)
	ValidateDatabase(path string) error
	ImportDatabase(path string, opts ImportOptions) error
	ImportMedia(opts ImportOptions) error
	CompareFiles(localPath string) (*CompareResult, error)
}
