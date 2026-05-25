package powershell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/perplexityai/bumblebee/internal/model"
)

func TestScanManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Pester.psd1")
	writeFile(t, path, `@{
  RootModule = 'Pester.psm1'
  ModuleVersion = '5.7.1'
  GUID = 'a699dea5-2c73-4616-a270-1f7abb777e71'
  Author = 'PowerShell Team'
}`)

	var records []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { records = append(records, r) }}
	if err := s.ScanManifest(path, model.Record{Profile: model.ProfileProject}); err != nil {
		t.Fatalf("ScanManifest: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records len=%d, want 1: %#v", len(records), records)
	}
	r := records[0]
	if r.Ecosystem != model.EcosystemPowerShellModule {
		t.Errorf("ecosystem=%q, want %q", r.Ecosystem, model.EcosystemPowerShellModule)
	}
	if r.PackageName != "Pester" || r.NormalizedName != "pester" {
		t.Errorf("package identity=(%q,%q), want Pester/pester", r.PackageName, r.NormalizedName)
	}
	if r.Version != "5.7.1" {
		t.Errorf("version=%q, want 5.7.1", r.Version)
	}
	if r.PackageManager != "powershell" {
		t.Errorf("package_manager=%q, want powershell", r.PackageManager)
	}
	if r.SourceType != "powershell-module-manifest" {
		t.Errorf("source_type=%q, want powershell-module-manifest", r.SourceType)
	}
	if r.SourceFile != path {
		t.Errorf("source_file=%q, want %q", r.SourceFile, path)
	}
	if r.ProjectPath != dir {
		t.Errorf("project_path=%q, want %q", r.ProjectPath, dir)
	}
	if r.Confidence != "high" {
		t.Errorf("confidence=%q, want high", r.Confidence)
	}
}

func TestScanManifestSkipsMissingModuleVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "NoVersion.psd1")
	writeFile(t, path, `@{
  RootModule = 'NoVersion.psm1'
}`)

	var records []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { records = append(records, r) }}
	if err := s.ScanManifest(path, model.Record{}); err != nil {
		t.Fatalf("ScanManifest: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("records len=%d, want 0: %#v", len(records), records)
	}
}

func TestParseManifestFields(t *testing.T) {
	fields := parseManifestFields(`@{
  # comments are ignored
  ModuleVersion = "1.2.3"
  PrivateData = @{
    PSData = @{
      ModuleVersion = '9.9.9'
    }
  }
  Author = 'O''Brien'
}`)
	if got := fields["moduleversion"]; got != "1.2.3" {
		t.Fatalf("moduleversion=%q, want 1.2.3", got)
	}
	if got := fields["author"]; got != "O'Brien" {
		t.Fatalf("author=%q, want O'Brien", got)
	}
}

func TestParseManifestFieldsDoesNotEvaluateExpressions(t *testing.T) {
	fields := parseManifestFields(`@{
  ModuleVersion = (Get-Date)
  RootModule = 'Example.psm1'
}`)
	if _, ok := fields["moduleversion"]; ok {
		t.Fatalf("expression ModuleVersion should not be captured: %#v", fields)
	}
}

func TestIsManifest(t *testing.T) {
	for _, base := range []string{"Module.psd1", "Module.PSD1"} {
		if !IsManifest(base) {
			t.Errorf("IsManifest(%q)=false, want true", base)
		}
	}
	if IsManifest("Module.psm1") {
		t.Fatal("IsManifest accepted .psm1")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
