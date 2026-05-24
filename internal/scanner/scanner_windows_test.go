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
	if res.FilesConsidered != 2 {
		t.Fatalf("FilesConsidered = %d, want 2 for package-lock.json and extensions.json only; stdout=%s stderr=%s", res.FilesConsidered, stdout.String(), stderr.String())
	}

	var sawFirefox, sawNPM bool
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
		if r.SourceType == "npm-lockfile" && r.PackageName == "lodash" {
			sawNPM = true
		}
	}
	if !sawFirefox || !sawNPM {
		t.Fatalf("required metadata discovery regressed: firefox=%v npm=%v stdout=%s", sawFirefox, sawNPM, stdout.String())
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
