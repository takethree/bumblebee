//go:build windows

package walk

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestWalkSkipsWindowsSensitiveSubtrees(t *testing.T) {
	root := t.TempDir()
	blocked := []string{
		filepath.Join(root, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Cookies"),
		filepath.Join(root, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Login Data"),
		filepath.Join(root, "AppData", "Roaming", "Microsoft", "Credentials", "token"),
		filepath.Join(root, "AppData", "Local", "Microsoft", "Protect", "secret"),
		filepath.Join(root, "AppData", "Local", "Microsoft", "Windows", "WebCache", "db"),
		filepath.Join(root, "AppData", "Local", "Packages", "app", "LocalState"),
		filepath.Join(root, "OneDrive - Contoso", "private"),
		filepath.Join(root, "Google Drive", "private"),
		filepath.Join(root, "Dropbox", "private"),
		filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default", "cache2"),
		filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default", "storage"),
		filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default", "sessionstore-backups"),
		filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default", "extensions"),
	}
	for _, dir := range blocked {
		mustWrite(t, filepath.Join(dir, "sentinel.txt"), "blocked")
	}
	wantProject := filepath.Join(root, "code", "proj", "package-lock.json")
	wantFirefox := filepath.Join(root, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles", "abcd.default", "extensions.json")
	mustWrite(t, wantProject, "{}")
	mustWrite(t, wantFirefox, `{"addons":[]}`)

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
	for _, p := range seen {
		if filepath.Base(p) == "sentinel.txt" {
			t.Errorf("sensitive Windows path was visited: %s", p)
		}
	}
	if !containsPath(seen, wantProject) {
		t.Errorf("expected project metadata to remain visitable; saw %v", seen)
	}
	if !containsPath(seen, wantFirefox) {
		t.Errorf("expected Firefox extensions.json to remain visitable; saw %v", seen)
	}
}
