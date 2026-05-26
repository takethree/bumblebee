//go:build windows

package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perplexityai/bumblebee/internal/model"
	"github.com/perplexityai/bumblebee/internal/output"
)

func TestSensitiveBrowserProfileFilesSkippedBeforeConsidered(t *testing.T) {
	root := t.TempDir()
	firefoxProfile := filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default-release")
	librewolfProfile := filepath.Join(root, "AppData", "Roaming", "LibreWolf", "Profiles", "abcd.default-release")
	waterfoxProfile := filepath.Join(root, "AppData", "Roaming", "Waterfox", "Waterfox", "Profiles", "abcd.default-release")
	waterfoxLegacyProfile := filepath.Join(root, "AppData", "Roaming", "Waterfox", "Profiles", "abcd.default-release")
	writeFile(t, filepath.Join(firefoxProfile, "cookies.sqlite"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "places.sqlite"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "favicons.sqlite-wal"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "formhistory.sqlite-shm"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "key4.db"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "logins.json"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "sessionstore.jsonlz4"), "private")
	writeFile(t, filepath.Join(firefoxProfile, "extensions.json"), `{
  "addons": [
    {"id":"safe@example.com","version":"1.0.0","type":"extension","active":true,"defaultLocale":{"name":"Safe Extension"}}
  ]
}`)
	writeFile(t, filepath.Join(librewolfProfile, "cookies.sqlite"), "private")
	writeFile(t, filepath.Join(librewolfProfile, "extensions.json"), `{
  "addons": [
    {"id":"librewolf-safe@example.com","version":"1.0.0","type":"extension","active":true,"defaultLocale":{"name":"LibreWolf Safe Extension"}}
  ]
}`)
	writeFile(t, filepath.Join(waterfoxProfile, "cookies.sqlite"), "private")
	writeFile(t, filepath.Join(waterfoxProfile, "extensions.json"), `{
  "addons": [
    {"id":"waterfox-nested-safe@example.com","version":"1.0.0","type":"extension","active":true,"defaultLocale":{"name":"Waterfox Nested Safe Extension"}}
  ]
}`)
	writeFile(t, filepath.Join(waterfoxLegacyProfile, "cookies.sqlite"), "private")
	writeFile(t, filepath.Join(waterfoxLegacyProfile, "extensions.json"), `{
  "addons": [
    {"id":"waterfox-legacy-safe@example.com","version":"1.0.0","type":"extension","active":true,"defaultLocale":{"name":"Waterfox Legacy Safe Extension"}}
  ]
}`)
	writeFile(t, filepath.Join(firefoxProfile, "cache2", "sentinel.txt"), "blocked")
	writeFile(t, filepath.Join(root, "proj", "package-lock.json"), `{
  "lockfileVersion": 3,
  "packages": {
    "node_modules/lodash": {"version":"4.17.21"}
  }
}`)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	em := output.New(stdout, stderr, "r")
	res, err := Run(context.Background(), Config{
		Roots:       []Root{{Path: root, Kind: model.RootKindDeepHome}},
		Profile:     model.ProfileDeep,
		MaxFileSize: 1 << 20,
		Concurrency: 1,
		BaseRecord: model.Record{
			SchemaVersion:  model.SchemaVersion,
			ScannerName:    model.ScannerName,
			ScannerVersion: "test",
			RunID:          "r",
			ScanTime:       time.Now().UTC().Format(time.RFC3339Nano),
		},
		Emitter: em,
	})
	if err != nil {
		t.Fatalf("Run: %v; stderr=%s", err, stderr.String())
	}
	if res.FilesConsidered != 5 {
		t.Fatalf("FilesConsidered = %d, want 5 for package-lock.json and Firefox-family extensions.json files only; stdout=%s stderr=%s", res.FilesConsidered, stdout.String(), stderr.String())
	}

	var sawFirefox, sawLibreWolf, sawWaterfoxNested, sawWaterfoxLegacy, sawNPM bool
	for _, line := range bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var r model.Record
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatalf("bad ndjson line: %v: %s", err, line)
		}
		if strings.Contains(filepath.ToSlash(r.SourceFile), "cookies.sqlite") ||
			strings.Contains(filepath.ToSlash(r.SourceFile), "places.sqlite") ||
			strings.Contains(filepath.ToSlash(r.SourceFile), "logins.json") {
			t.Fatalf("sensitive browser profile file emitted a record: %+v", r)
		}
		if r.SourceType == "browser-extension" && r.PackageName == "Safe Extension" {
			sawFirefox = true
		}
		if r.SourceType == "browser-extension" && r.PackageName == "LibreWolf Safe Extension" {
			sawLibreWolf = true
		}
		if r.SourceType == "browser-extension" && r.PackageName == "Waterfox Nested Safe Extension" {
			sawWaterfoxNested = true
		}
		if r.SourceType == "browser-extension" && r.PackageName == "Waterfox Legacy Safe Extension" {
			sawWaterfoxLegacy = true
		}
		if r.SourceType == "npm-lockfile" && r.PackageName == "lodash" {
			sawNPM = true
		}
	}
	if !sawFirefox || !sawLibreWolf || !sawWaterfoxNested || !sawWaterfoxLegacy || !sawNPM {
		t.Fatalf("required metadata discovery regressed: firefox=%v librewolf=%v waterfox_nested=%v waterfox_legacy=%v npm=%v stdout=%s", sawFirefox, sawLibreWolf, sawWaterfoxNested, sawWaterfoxLegacy, sawNPM, stdout.String())
	}
}

func TestWindowsStrictParityPackageRootsEmitRecords(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AppData", "Roaming", "npm", "node_modules", "smoke-npm-root", "package.json"), `{
  "name": "smoke-npm-root",
  "version": "1.0.0"
}`)
	writeFile(t, filepath.Join(root, "AppData", "Roaming", "Python", "Python311", "site-packages", "SmokePythonRoot-1.0.0.dist-info", "METADATA"), "Metadata-Version: 2.1\nName: SmokePythonRoot\nVersion: 1.0.0\n\n")
	writeFile(t, filepath.Join(root, "pipx", "venvs", "smoke-pipx", "Lib", "site-packages", "SmokePipxRoot-2.0.0.dist-info", "METADATA"), "Metadata-Version: 2.1\nName: SmokePipxRoot\nVersion: 2.0.0\n\n")
	writeFile(t, filepath.Join(root, "AppData", "Local", "pipx", "venvs", "smoke-pipx-local", "Lib", "site-packages", "SmokePipxLocalRoot-3.0.0.dist-info", "METADATA"), "Metadata-Version: 2.1\nName: SmokePipxLocalRoot\nVersion: 3.0.0\n\n")
	writeFile(t, filepath.Join(root, ".local", "pipx", "venvs", "smoke-pipx-legacy", "Lib", "site-packages", "SmokePipxLegacyRoot-4.0.0.dist-info", "METADATA"), "Metadata-Version: 2.1\nName: SmokePipxLegacyRoot\nVersion: 4.0.0\n\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	em := output.New(stdout, stderr, "r")
	res, err := Run(context.Background(), Config{
		Roots:       []Root{{Path: root, Kind: model.RootKindDeepHome}},
		Profile:     model.ProfileDeep,
		MaxFileSize: 1 << 20,
		Concurrency: 1,
		BaseRecord: model.Record{
			SchemaVersion:  model.SchemaVersion,
			ScannerName:    model.ScannerName,
			ScannerVersion: "test",
			RunID:          "r",
			ScanTime:       time.Now().UTC().Format(time.RFC3339Nano),
		},
		Emitter: em,
	})
	if err != nil {
		t.Fatalf("Run: %v; stderr=%s", err, stderr.String())
	}
	if res.FilesConsidered != 5 {
		t.Fatalf("FilesConsidered = %d, want 5 Windows strict-parity metadata files; stdout=%s stderr=%s", res.FilesConsidered, stdout.String(), stderr.String())
	}

	want := map[string]string{
		"npm:smoke-npm-root:1.0.0":       "npm-node_modules",
		"pypi:SmokePythonRoot:1.0.0":     "pypi-dist-info",
		"pypi:SmokePipxRoot:2.0.0":       "pypi-dist-info",
		"pypi:SmokePipxLocalRoot:3.0.0":  "pypi-dist-info",
		"pypi:SmokePipxLegacyRoot:4.0.0": "pypi-dist-info",
	}
	for _, line := range bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var r model.Record
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatalf("bad ndjson line: %v: %s", err, line)
		}
		key := r.Ecosystem + ":" + r.PackageName + ":" + r.Version
		if sourceType, ok := want[key]; ok {
			if r.SourceType != sourceType {
				t.Errorf("%s source_type = %q, want %q", key, r.SourceType, sourceType)
			}
			delete(want, key)
		}
	}
	if len(want) > 0 {
		t.Fatalf("missing Windows strict-parity package records: %v; stdout=%s stderr=%s", want, stdout.String(), stderr.String())
	}
}

func TestWindowsArcAndCometExplicitProfileRootsSkipSensitiveFiles(t *testing.T) {
	root := t.TempDir()
	cometProfile := filepath.Join(root, "AppData", "Local", "Perplexity", "Comet", "User Data", "Default")
	arcProfile := filepath.Join(root, "AppData", "Local", "Packages", "TheBrowserCompany.Arc_ttt1ap7aakyb4", "LocalCache", "Local", "Arc", "User Data", "Default")
	writeFile(t, filepath.Join(cometProfile, "Cookies"), "private")
	writeFile(t, filepath.Join(cometProfile, "Login Data"), "private")
	writeFile(t, filepath.Join(cometProfile, "Extensions", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "1.0.0", "manifest.json"), `{"name":"Comet Safe Extension","version":"1.0.0","manifest_version":3}`)
	writeFile(t, filepath.Join(arcProfile, "History"), "private")
	writeFile(t, filepath.Join(arcProfile, "Web Data"), "private")
	writeFile(t, filepath.Join(arcProfile, "Extensions", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "2.0.0", "manifest.json"), `{"name":"Arc Safe Extension","version":"2.0.0","manifest_version":3}`)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	em := output.New(stdout, stderr, "r")
	res, err := Run(context.Background(), Config{
		Roots: []Root{
			{Path: cometProfile, Kind: model.RootKindBrowserExtension},
			{Path: arcProfile, Kind: model.RootKindBrowserExtension},
		},
		Profile:     model.ProfileBaseline,
		MaxFileSize: 1 << 20,
		Concurrency: 1,
		BaseRecord: model.Record{
			SchemaVersion:  model.SchemaVersion,
			ScannerName:    model.ScannerName,
			ScannerVersion: "test",
			RunID:          "r",
			ScanTime:       time.Now().UTC().Format(time.RFC3339Nano),
		},
		Emitter: em,
	})
	if err != nil {
		t.Fatalf("Run: %v; stderr=%s", err, stderr.String())
	}
	if res.FilesConsidered != 2 {
		t.Fatalf("FilesConsidered = %d, want only the two extension manifests; stdout=%s stderr=%s", res.FilesConsidered, stdout.String(), stderr.String())
	}

	var sawComet, sawArc bool
	for _, line := range bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var r model.Record
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatalf("bad ndjson line: %v: %s", err, line)
		}
		if strings.Contains(filepath.ToSlash(r.SourceFile), "Cookies") ||
			strings.Contains(filepath.ToSlash(r.SourceFile), "Login Data") ||
			strings.Contains(filepath.ToSlash(r.SourceFile), "History") ||
			strings.Contains(filepath.ToSlash(r.SourceFile), "Web Data") {
			t.Fatalf("sensitive Arc/Comet profile file emitted a record: %+v", r)
		}
		if r.SourceType == "browser-extension" && r.PackageName == "Comet Safe Extension" {
			sawComet = true
		}
		if r.SourceType == "browser-extension" && r.PackageName == "Arc Safe Extension" {
			sawArc = true
		}
	}
	if !sawComet || !sawArc {
		t.Fatalf("required Arc/Comet extension metadata missing: comet=%v arc=%v stdout=%s", sawComet, sawArc, stdout.String())
	}
}

func TestWindowsAccessDeniedIsDebugLevelDiagnostic(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "proj", "package-lock.json"), `{
  "lockfileVersion": 3,
  "packages": {
    "node_modules/lodash": {"version":"4.17.21"}
  }
}`)
	denied := filepath.Join(root, "denied")
	writeFile(t, filepath.Join(denied, "sentinel.txt"), "private")

	current, err := user.Current()
	if err != nil || current.Username == "" {
		t.Skipf("cannot determine current user for ACL fixture: %v", err)
	}
	denySpec := current.Username + ":(OI)(CI)(RX)"
	out, err := exec.Command("icacls", denied, "/deny", denySpec).CombinedOutput()
	if err != nil {
		t.Skipf("cannot apply deny ACL: %v: %s", err, string(out))
	}
	t.Cleanup(func() {
		_ = exec.Command("icacls", denied, "/remove:d", current.Username).Run()
		_ = os.Chmod(denied, 0o755)
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	em := output.New(stdout, stderr, "r")
	_, err = Run(context.Background(), Config{
		Roots:       []Root{{Path: root, Kind: model.RootKindProject}},
		Profile:     model.ProfileProject,
		MaxFileSize: 1 << 20,
		Concurrency: 1,
		BaseRecord: model.Record{
			SchemaVersion:  model.SchemaVersion,
			ScannerName:    model.ScannerName,
			ScannerVersion: "test",
			RunID:          "r",
			ScanTime:       time.Now().UTC().Format(time.RFC3339Nano),
		},
		Emitter: em,
	})
	if err != nil {
		t.Fatalf("Run: %v; stderr=%s", err, stderr.String())
	}
	if em.RecordsEmitted < 1 {
		t.Fatalf("expected healthy scan to emit at least one record; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	var diagsForDenied []model.Diagnostic
	for _, line := range bytes.Split(bytes.TrimSpace(stderr.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var d model.Diagnostic
		if err := json.Unmarshal(line, &d); err != nil {
			t.Fatalf("bad diagnostic line: %v: %s", err, line)
		}
		if strings.Contains(d.Path, "denied") {
			diagsForDenied = append(diagsForDenied, d)
		}
	}
	if len(diagsForDenied) == 0 {
		t.Fatalf("expected at least one diagnostic for the denied path; stderr=%s", stderr.String())
	}
	for _, d := range diagsForDenied {
		if d.Level != "debug" {
			t.Errorf("permission-denied diagnostic level = %q, want %q (path=%q msg=%q)",
				d.Level, "debug", d.Path, d.Message)
		}
	}
}
