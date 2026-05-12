package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ChecklistData holds all data passed to the migration checklist template.
type ChecklistData struct {
	Install      string
	Destination  string
	Domain       string
	GeneratedAt  string
	ReportPath   string
	ReportExists bool
	SQLPath      string
	SQLExists    bool
	VIPCommitSHA string // empty if not yet published to VIP repo
	PullDone     bool   // true if wp-content/ dir detected
	ExportDone   bool   // true if SQL dump detected
	ReportDone   bool   // true if migration report detected
	PublishDone  bool   // true if VIP commit SHA found
}

var checklistTmpl = template.Must(template.New("checklist").Funcs(template.FuncMap{
	"check": func(done bool) string {
		if done {
			return "x"
		}
		return " "
	},
}).Parse(checklistTemplate))

// WriteChecklist writes the checklist markdown to outputPath, creating parent dirs as needed.
func WriteChecklist(data ChecklistData, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create checklist file: %w", err)
	}
	defer f.Close()
	if err := checklistTmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to render checklist template: %w", err)
	}
	return nil
}

const checklistTemplate = `# Migration Checklist: {{ .Install }}

**Site:** {{ .Install }}
**Destination:** {{ .Destination }}
**Domain:** {{ .Domain }}
**Generated:** {{ .GeneratedAt }}

---

## Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Migration report | {{ .ReportPath }} | {{ if .ReportExists }}Present{{ else }}Not found{{ end }} |
| SQL dump | {{ .SQLPath }} | {{ if .SQLExists }}Present{{ else }}Not found{{ end }} |
| VIP repo commit | {{ if .VIPCommitSHA }}{{ .VIPCommitSHA }}{{ else }}—{{ end }} | {{ if .VIPCommitSHA }}Present{{ else }}Not found{{ end }} |

---

## Pre-migration steps

- [{{ check .PullDone }}] ` + "`stax migrate pull`" + ` — files downloaded to ` + "`wp-content/`" + `
- [{{ check .ExportDone }}] ` + "`stax migrate export`" + ` — SQL dump at ` + "`{{ .SQLPath }}`" + `
- [{{ check .ReportDone }}] ` + "`stax migrate audit`" + ` — phpcs results in migration report
- [{{ check .ReportDone }}] ` + "`stax migrate compare`" + ` — file comparison in migration report
- [ ] ` + "`stax migrate import`" + ` — database imported to VIP
- [{{ check .ReportDone }}] ` + "`stax migrate report`" + ` — report at ` + "`{{ .ReportPath }}`" + `
- [{{ check .PublishDone }}] ` + "`stax migrate publish`" + ` — artifacts uploaded to S3, committed to VIP repo{{ if .VIPCommitSHA }} ({{ .VIPCommitSHA }}){{ end }}

---

## QA Checklist

Run on the VIP staging environment before DNS cutover.

- [ ] Front page loads
- [ ] Key pages load (about, contact, category pages)
- [ ] Forms work
- [ ] No broken images

---

## DNS Cutover

- [ ] Reduce TTL on ` + "`{{ .Domain }}`" + ` to 300s (do this 24–48h before cutover)
- [ ] Confirm TTL reduction has propagated
- [ ] Swap A/CNAME record for ` + "`{{ .Domain }}`" + ` to VIP
- [ ] Verify DNS resolution has updated

---

## Post-launch Validation

Run on ` + "`{{ .Domain }}`" + ` after DNS cutover.

- [ ] Front page loads on ` + "`{{ .Domain }}`" + `
- [ ] Key pages load on ` + "`{{ .Domain }}`" + `
- [ ] Forms work on ` + "`{{ .Domain }}`" + `
- [ ] No broken images on ` + "`{{ .Domain }}`" + `

---

## Sign-off

**Operator:** ___________________________
**Date:** ___________________________
`
