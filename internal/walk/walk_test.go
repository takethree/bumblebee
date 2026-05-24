package walk

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultExcludesCoverProtectedMacOSLibraryPaths ensures the macOS
// Library subtrees that routinely produce TCC denials under broad
// $HOME scans are matched by the default suffix-component excludes.
// Adding new paths to DefaultExcludes is cheap; regressing one of
// these silently is what makes the diagnostics output scary.
func TestDefaultExcludesCoverProtectedMacOSLibraryPaths(t *testing.T) {
	want := []string{
		"Library/ContainerManager",
		"Library/Daemon Containers",
		"Library/DoNotDisturb",
		"Library/DuetExpertCenter",
		"Library/IntelligencePlatform",
		"Library/Photos",
		"Library/Sharing",
		"Library/Shortcuts",
		"Library/StatusKit",
	}
	have := make(map[string]bool, len(DefaultExcludes))
	for _, x := range DefaultExcludes {
		have[x] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("DefaultExcludes missing %q", w)
		}
	}
}

func TestDefaultExcludesCoverWindowsSensitivePaths(t *testing.T) {
	want := []string{
		"AppData/Local/Google/Chrome/User Data",
		"AppData/Local/Microsoft/Edge/User Data",
		"AppData/Roaming/Microsoft/Credentials",
		"AppData/Local/Microsoft/Credentials",
		"AppData/Roaming/Microsoft/Protect",
		"AppData/Local/Microsoft/Vault",
		"AppData/Local/Microsoft/Windows/WebCache",
		"AppData/Local/Packages",
		"Google Drive",
		"Dropbox",
	}
	have := make(map[string]bool, len(DefaultExcludes))
	for _, x := range DefaultExcludes {
		have[x] = true
	}
	if runtime.GOOS != "windows" {
		for _, w := range want {
			if have[w] {
				t.Errorf("non-Windows DefaultExcludes unexpectedly include Windows-only path %q", w)
			}
		}
		return
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("DefaultExcludes missing %q", w)
		}
	}
}

// TestWalkSkipsExcludedLibrarySubtrees verifies that an exclude with
// a "/"-separated suffix (e.g. "Library/ContainerManager") prunes a
// matching directory anywhere under any root, while a sibling
// directory that does not match continues to be walked.
func TestWalkSkipsExcludedLibrarySubtrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path-separator semantics differ on Windows")
	}
	root := t.TempDir()
	// Simulate a $HOME-shaped tree.
	mustMkdir(t, filepath.Join(root, "Library", "ContainerManager", "deep"))
	mustMkdir(t, filepath.Join(root, "Library", "StatusKit"))
	mustMkdir(t, filepath.Join(root, "code", "proj"))

	// Drop sentinel files we can detect from the visitor.
	mustWrite(t, filepath.Join(root, "Library", "ContainerManager", "deep", "secret.json"), "{}")
	mustWrite(t, filepath.Join(root, "Library", "StatusKit", "x"), "{}")
	mustWrite(t, filepath.Join(root, "code", "proj", "package-lock.json"), "{}")

	excludes := append([]string{}, DefaultExcludes...)

	var seen []string
	err := Walk(Options{
		Roots:    []string{root},
		Excludes: excludes,
	}, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			seen = append(seen, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	for _, p := range seen {
		if filepath.Base(filepath.Dir(p)) == "deep" || filepath.Base(filepath.Dir(p)) == "StatusKit" {
			t.Errorf("excluded path was visited: %s", p)
		}
	}
	want := filepath.Join(root, "code", "proj", "package-lock.json")
	found := false
	for _, p := range seen {
		if p == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to visit %q; saw %v", want, seen)
	}
}

func TestWalkDoesNotExcludeCuratedBrowserExtensionRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Extensions")
	want := filepath.Join(root, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "1.0.0_0", "manifest.json")
	mustWrite(t, want, `{"manifest_version":3}`)

	var seen []string
	err := Walk(Options{
		Roots:    []string{root},
		Excludes: append([]string{}, DefaultExcludes...),
	}, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			seen = append(seen, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if !containsPath(seen, want) {
		t.Fatalf("curated browser extension root was pruned; saw %v", seen)
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}
