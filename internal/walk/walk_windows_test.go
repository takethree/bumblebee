//go:build windows

package walk

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestWalkSkipsWindowsJunctionReparseDirs(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "linked")
	mustJunction(t, link, outside)
	if !shouldSkipLinkedDir(link) {
		t.Fatal("junction was not recognized as a linked/reparse directory")
	}

	want := filepath.Join(root, "code", "proj", "package-lock.json")
	mustWrite(t, want, "{}")
	mustWrite(t, filepath.Join(outside, "secret", "package-lock.json"), "{}")

	var seen []string
	err := Walk(Options{
		Roots:    []string{root},
		Excludes: append([]string{}, DefaultExcludes...),
	}, func(path string, d fs.DirEntry) error {
		seen = append(seen, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if !containsPath(seen, want) {
		t.Fatalf("normal metadata was not visited; saw %v", seen)
	}
	linkPrefix := link + string(filepath.Separator)
	for _, p := range seen {
		if strings.HasPrefix(p, linkPrefix) {
			t.Fatalf("junction reparse directory was visited: %s; saw %v", p, seen)
		}
	}
}

func TestWalkDoesNotLoopThroughWindowsJunction(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(a, "b")
	mustWrite(t, filepath.Join(a, "package-lock.json"), "{}")
	mustMkdir(t, b)
	loop := filepath.Join(b, "loop")
	mustJunction(t, loop, a)
	if !shouldSkipLinkedDir(loop) {
		t.Fatal("junction loop was not recognized as a linked/reparse directory")
	}

	done := make(chan error, 1)
	var seen []string
	go func() {
		done <- Walk(Options{
			Roots:    []string{root},
			Excludes: append([]string{}, DefaultExcludes...),
		}, func(path string, d fs.DirEntry) error {
			seen = append(seen, path)
			return nil
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("walk did not terminate; junction loop suspected")
	}

	loopPrefix := loop + string(filepath.Separator)
	for _, p := range seen {
		if strings.HasPrefix(p, loopPrefix) {
			t.Fatalf("junction loop was visited: %s; saw %v", p, seen)
		}
	}
}

func mustJunction(t *testing.T, link, target string) {
	t.Helper()
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Skipf("cannot create junction: %v: %s", err, string(out))
	}
}
