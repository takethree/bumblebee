package nuget

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/perplexityai/bumblebee/internal/model"
)

func TestScanPackagesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packages.config")
	writeFile(t, path, `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Newtonsoft.Json" version="13.0.3" targetFramework="net472" />
  <package id="Serilog" version="3.1.1" />
  <package id="MissingVersion" />
</packages>`)

	var records []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { records = append(records, r) }}
	if err := s.ScanPackagesConfig(path, model.Record{Profile: model.ProfileProject}); err != nil {
		t.Fatalf("ScanPackagesConfig: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records len=%d, want 2: %#v", len(records), records)
	}
	assertRecord(t, records[0], "Newtonsoft.Json", "newtonsoft.json", "13.0.3", "", "nuget-packages-config", "")
	assertRecord(t, records[1], "Serilog", "serilog", "3.1.1", "", "nuget-packages-config", "")
	for _, r := range records {
		if r.ProjectPath != dir {
			t.Errorf("ProjectPath=%q, want %q", r.ProjectPath, dir)
		}
		if r.DirectDependency != nil {
			t.Errorf("%s direct_dependency=%v, want empty for packages.config", r.PackageName, *r.DirectDependency)
		}
	}
}

func TestScanLockfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packages.lock.json")
	writeFile(t, path, `{
  "version": 1,
  "dependencies": {
    ".NETFramework,Version=v4.7.2": {
      "Newtonsoft.Json": {
        "type": "Direct",
        "requested": "[13.0.3, )",
        "resolved": "13.0.3",
        "contentHash": "abc"
      },
      "Serilog": {
        "type": "Transitive",
        "requested": "[3.0.0, )",
        "resolved": "3.1.1",
        "contentHash": "def"
      },
      "Local.Project": {
        "type": "Project"
      },
      "NoResolved": {
        "type": "Direct"
      }
    },
    "net8.0": {
      "Newtonsoft.Json": {
        "type": "Direct",
        "requested": "[13.0.3, )",
        "resolved": "13.0.3",
        "contentHash": "abc"
      },
      "Different.Version": {
        "type": "Transitive",
        "requested": "[2.0.0, )",
        "resolved": "2.0.0",
        "contentHash": "ghi"
      }
    }
  }
}`)

	var records []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { records = append(records, r) }}
	if err := s.ScanLockfile(path, model.Record{Profile: model.ProfileProject}); err != nil {
		t.Fatalf("ScanLockfile: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records len=%d, want 3: %#v", len(records), records)
	}

	byName := map[string]model.Record{}
	for _, r := range records {
		byName[r.PackageName] = r
		if r.ProjectPath != dir {
			t.Errorf("%s ProjectPath=%q, want %q", r.PackageName, r.ProjectPath, dir)
		}
	}
	assertRecord(t, byName["Newtonsoft.Json"], "Newtonsoft.Json", "newtonsoft.json", "13.0.3", "[13.0.3, )", "nuget-lockfile", "")
	if byName["Newtonsoft.Json"].DirectDependency == nil || !*byName["Newtonsoft.Json"].DirectDependency {
		t.Errorf("Newtonsoft.Json direct_dependency=%v, want true", byName["Newtonsoft.Json"].DirectDependency)
	}
	assertRecord(t, byName["Serilog"], "Serilog", "serilog", "3.1.1", "[3.0.0, )", "nuget-lockfile", "transitive")
	if byName["Serilog"].DirectDependency == nil || *byName["Serilog"].DirectDependency {
		t.Errorf("Serilog direct_dependency=%v, want false", byName["Serilog"].DirectDependency)
	}
	assertRecord(t, byName["Different.Version"], "Different.Version", "different.version", "2.0.0", "[2.0.0, )", "nuget-lockfile", "transitive")
	if _, ok := byName["Local.Project"]; ok {
		t.Fatal("project reference should be skipped")
	}
	if _, ok := byName["NoResolved"]; ok {
		t.Fatal("lockfile entry without resolved version should be skipped")
	}
}

func assertRecord(t *testing.T, r model.Record, packageName, normalized, version, requested, sourceType, scope string) {
	t.Helper()
	if r.Ecosystem != model.EcosystemNuGet {
		t.Errorf("%s ecosystem=%q, want %q", packageName, r.Ecosystem, model.EcosystemNuGet)
	}
	if r.PackageManager != "nuget" {
		t.Errorf("%s package_manager=%q, want nuget", packageName, r.PackageManager)
	}
	if r.PackageName != packageName {
		t.Errorf("PackageName=%q, want %q", r.PackageName, packageName)
	}
	if r.NormalizedName != normalized {
		t.Errorf("%s NormalizedName=%q, want %q", packageName, r.NormalizedName, normalized)
	}
	if r.Version != version {
		t.Errorf("%s Version=%q, want %q", packageName, r.Version, version)
	}
	if r.RequestedSpec != requested {
		t.Errorf("%s RequestedSpec=%q, want %q", packageName, r.RequestedSpec, requested)
	}
	if r.SourceType != sourceType {
		t.Errorf("%s SourceType=%q, want %q", packageName, r.SourceType, sourceType)
	}
	if r.SourceFile == "" {
		t.Errorf("%s SourceFile is empty", packageName)
	}
	if r.InstallScope != scope {
		t.Errorf("%s InstallScope=%q, want %q", packageName, r.InstallScope, scope)
	}
	if r.Confidence != "high" {
		t.Errorf("%s Confidence=%q, want high", packageName, r.Confidence)
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
