//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplexityai/bumblebee/internal/model"
)

func TestResolveRootsBaselineIncludesWindowsCurrentUserRoots(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	msixClaude := filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	psUserModules := filepath.Join(home, "Documents", "PowerShell", "Modules")
	winPsUserModules := filepath.Join(home, "Documents", "WindowsPowerShell", "Modules")
	psAllUsersModules := filepath.Join(programFiles, "PowerShell", "Modules")
	winPsAllUsersModules := filepath.Join(programFiles, "WindowsPowerShell", "Modules")
	chromeDefaultExt := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions")
	chromeProfile1Ext := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Profile 1", "Extensions")
	edgeDefaultExt := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions")
	edgeProfile1Ext := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")

	want := map[string]string{
		filepath.Join(home, "go"):                             model.RootKindUserPackage,
		filepath.Join(home, ".vscode", "extensions"):          model.RootKindEditorExtension,
		filepath.Join(home, ".vscode-insiders", "extensions"): model.RootKindEditorExtension,
		filepath.Join(home, ".cursor", "extensions"):          model.RootKindEditorExtension,
		filepath.Join(home, ".cursor-server", "extensions"):   model.RootKindEditorExtension,
		filepath.Join(home, ".windsurf", "extensions"):        model.RootKindEditorExtension,
		filepath.Join(home, ".windsurf-server", "extensions"): model.RootKindEditorExtension,
		filepath.Join(home, ".vscodium", "extensions"):        model.RootKindEditorExtension,
		psUserModules:                    model.RootKindUserPackage,
		winPsUserModules:                 model.RootKindUserPackage,
		psAllUsersModules:                model.RootKindGlobalPackage,
		winPsAllUsersModules:             model.RootKindGlobalPackage,
		filepath.Join(appData, "Claude"): model.RootKindMCPConfig,
		msixClaude:                       model.RootKindMCPConfig,
		chromeDefaultExt:                 model.RootKindBrowserExtension,
		chromeProfile1Ext:                model.RootKindBrowserExtension,
		edgeDefaultExt:                   model.RootKindBrowserExtension,
		edgeProfile1Ext:                  model.RootKindBrowserExtension,
		firefoxProfiles:                  model.RootKindBrowserExtension,
	}
	for p := range want {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}
	got := map[string]string{}
	for _, r := range roots {
		got[r.Path] = r.Kind
	}
	for p, kind := range want {
		gotKind, ok := got[p]
		if !ok {
			t.Errorf("baseline missing Windows root %q (got %v)", p, roots)
			continue
		}
		if gotKind != kind {
			t.Errorf("baseline root %q kind = %q, want %q", p, gotKind, kind)
		}
	}
}

func TestResolveRootsBaselineSkipsAbsentWindowsCandidates(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	msixClaude := filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	if err := os.MkdirAll(filepath.Join(home, "go"), 0o755); err != nil {
		t.Fatal(err)
	}

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}
	absent := []string{
		filepath.Join(home, ".vscode", "extensions"),
		filepath.Join(home, ".cursor", "extensions"),
		filepath.Join(home, ".windsurf", "extensions"),
		filepath.Join(home, ".vscodium", "extensions"),
		filepath.Join(home, "Documents", "PowerShell", "Modules"),
		filepath.Join(home, "Documents", "WindowsPowerShell", "Modules"),
		filepath.Join(programFiles, "PowerShell", "Modules"),
		filepath.Join(programFiles, "WindowsPowerShell", "Modules"),
		filepath.Join(appData, "Claude"),
		msixClaude,
		filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions"),
		filepath.Join(appData, "Mozilla", "Firefox", "Profiles"),
	}
	gotPaths := map[string]bool{}
	for _, r := range roots {
		gotPaths[r.Path] = true
	}
	for _, p := range absent {
		if gotPaths[p] {
			t.Errorf("baseline emitted absent Windows candidate %q (roots=%v)", p, roots)
		}
	}
}

func TestClassifyRootWindowsBrowserExtensions(t *testing.T) {
	cases := []string{
		`C:\Users\alice\AppData\Local\Google\Chrome\User Data\Default\Extensions`,
		`C:\Users\alice\AppData\Local\Microsoft\Edge\User Data\Profile 1\Extensions`,
		`C:\Users\alice\AppData\Roaming\Mozilla\Firefox\Profiles`,
	}
	for _, p := range cases {
		if got := classifyRoot(p, model.ProfileBaseline); got != model.RootKindBrowserExtension {
			t.Errorf("classifyRoot(%q) = %q, want %q", p, got, model.RootKindBrowserExtension)
		}
	}
}

func TestClassifyRootWindowsClaudeMCP(t *testing.T) {
	cases := []string{
		`C:\Users\alice\AppData\Roaming\Claude`,
		`C:\Users\alice\AppData\Local\Packages\Claude_pzs8sxrjxfjjc\LocalCache\Roaming\Claude`,
	}
	for _, p := range cases {
		if got := classifyRoot(p, model.ProfileBaseline); got != model.RootKindMCPConfig {
			t.Errorf("classifyRoot(%q) = %q, want %q", p, got, model.RootKindMCPConfig)
		}
	}
}

func TestResolveRootsBaselineWindowsBrowserRootsAvoidSensitiveParents(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)

	chromeDefault := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default")
	edgeDefault := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	firefoxProfile := filepath.Join(firefoxProfiles, "abcd.default-release")
	wantRoots := []string{
		filepath.Join(chromeDefault, "Extensions"),
		filepath.Join(edgeDefault, "Extensions"),
		firefoxProfiles,
	}
	notRoots := []string{
		filepath.Dir(chromeDefault),
		chromeDefault,
		filepath.Join(chromeDefault, "Cookies"),
		filepath.Join(chromeDefault, "Login Data"),
		filepath.Join(chromeDefault, "History"),
		filepath.Join(chromeDefault, "Local Storage"),
		filepath.Join(chromeDefault, "IndexedDB"),
		filepath.Join(chromeDefault, "Cache"),
		filepath.Dir(edgeDefault),
		edgeDefault,
		filepath.Join(edgeDefault, "Cookies"),
		filepath.Join(edgeDefault, "Login Data"),
		filepath.Join(edgeDefault, "History"),
		filepath.Join(edgeDefault, "Local Storage"),
		filepath.Join(edgeDefault, "IndexedDB"),
		filepath.Join(edgeDefault, "Cache"),
		firefoxProfile,
		filepath.Join(firefoxProfile, "cache2"),
		filepath.Join(firefoxProfile, "cookies.sqlite"),
		filepath.Join(firefoxProfile, "places.sqlite"),
		filepath.Join(firefoxProfile, "storage"),
		filepath.Join(firefoxProfile, "sessionstore-backups"),
		filepath.Join(firefoxProfile, "extensions"),
	}
	for _, p := range append(append([]string{}, wantRoots...), notRoots...) {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(firefoxProfile, "extensions.json"), []byte(`{"addons":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}
	gotPaths := map[string]string{}
	for _, r := range roots {
		gotPaths[r.Path] = r.Kind
	}
	for _, p := range wantRoots {
		if gotPaths[p] != model.RootKindBrowserExtension {
			t.Errorf("browser root %q kind = %q, want %q (roots=%v)", p, gotPaths[p], model.RootKindBrowserExtension, roots)
		}
	}
	for _, p := range notRoots {
		if _, ok := gotPaths[p]; ok {
			t.Errorf("sensitive browser path %q was emitted as a root (roots=%v)", p, roots)
		}
	}
}

func TestAllUsersHomesFiltersWindowsServiceEntries(t *testing.T) {
	usersDir, realHomes := fakeUsersDir(t,
		[]string{"alice", "bob"},
		[]string{"Public", "Default", "Default User", "All Users", "desktop.ini", "defaultuser0", "WDAGUtilityAccount"})
	got := allUsersHomes(usersDir)
	if len(got) != len(realHomes) {
		t.Fatalf("allUsersHomes returned %d entries, want %d (got=%v real=%v)", len(got), len(realHomes), got, realHomes)
	}
	want := map[string]bool{}
	for _, h := range realHomes {
		want[h] = true
	}
	for _, g := range got {
		if !want[g] {
			t.Errorf("unexpected home %q", g)
		}
	}
}

func TestResolveRootsBaselineAllUsersExpansionWindows(t *testing.T) {
	usersDir, realHomes := fakeUsersDir(t,
		[]string{"alice", "bob"},
		[]string{"Public", "Default", "Default User", "All Users", "desktop.ini"})
	t.Setenv("BUMBLEBEE_USERS_DIR", usersDir)
	setTestHome(t, realHomes[0])
	t.Setenv("APPDATA", filepath.Join(realHomes[0], "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(realHomes[0], "AppData", "Local"))

	mustMkdir := func(p string) {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, h := range realHomes {
		roaming := filepath.Join(h, "AppData", "Roaming")
		local := filepath.Join(h, "AppData", "Local")
		for _, p := range []string{
			filepath.Join(h, "go"),
			filepath.Join(h, "Documents", "PowerShell", "Modules"),
			filepath.Join(h, "Documents", "WindowsPowerShell", "Modules"),
			filepath.Join(h, ".vscode", "extensions"),
			filepath.Join(roaming, "Claude"),
			filepath.Join(local, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"),
			filepath.Join(local, "Google", "Chrome", "User Data", "Default", "Extensions"),
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions"),
			filepath.Join(roaming, "Mozilla", "Firefox", "Profiles"),
		} {
			mustMkdir(p)
		}
	}

	roots, notes, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{AllUsers: true})
	if err != nil {
		t.Fatalf("resolveRoots baseline --all-users: %v", err)
	}
	gotPaths := map[string]string{}
	for _, r := range roots {
		gotPaths[r.Path] = r.Kind
	}
	for _, h := range realHomes {
		roaming := filepath.Join(h, "AppData", "Roaming")
		local := filepath.Join(h, "AppData", "Local")
		want := map[string]string{
			filepath.Join(h, "go"):                                                                      model.RootKindUserPackage,
			filepath.Join(h, "Documents", "PowerShell", "Modules"):                                      model.RootKindUserPackage,
			filepath.Join(h, "Documents", "WindowsPowerShell", "Modules"):                               model.RootKindUserPackage,
			filepath.Join(h, ".vscode", "extensions"):                                                   model.RootKindEditorExtension,
			filepath.Join(roaming, "Claude"):                                                            model.RootKindMCPConfig,
			filepath.Join(local, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"): model.RootKindMCPConfig,
			filepath.Join(local, "Google", "Chrome", "User Data", "Default", "Extensions"):              model.RootKindBrowserExtension,
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions"):           model.RootKindBrowserExtension,
			filepath.Join(roaming, "Mozilla", "Firefox", "Profiles"):                                    model.RootKindBrowserExtension,
		}
		for p, kind := range want {
			if gotPaths[p] != kind {
				t.Errorf("root %q kind = %q, want %q (roots=%v)", p, gotPaths[p], kind, roots)
			}
		}
		if _, ok := gotPaths[h]; ok {
			t.Errorf("bare home %q was added as a root", h)
		}
	}
	for _, name := range []string{"Public", "Default", "Default User", "All Users"} {
		bad := filepath.Join(usersDir, name)
		for p := range gotPaths {
			if strings.HasPrefix(p, bad+string(filepath.Separator)) {
				t.Errorf("service profile %q produced root %q", name, p)
			}
		}
	}
	var sawExpansion, sawUnsupported bool
	for _, n := range notes {
		if strings.Contains(n, "--all-users") && strings.Contains(n, "home(s)") {
			sawExpansion = true
		}
		if strings.Contains(n, "--all-users") && strings.Contains(n, "not supported") {
			sawUnsupported = true
		}
	}
	if !sawExpansion || sawUnsupported {
		t.Fatalf("expected Windows expansion note and no unsupported note, got %v", notes)
	}
}

func TestResolveRootsProjectAllUsersExpansionWindows(t *testing.T) {
	usersDir, realHomes := fakeUsersDir(t, []string{"alice", "bob"}, []string{"Public", "Default"})
	t.Setenv("BUMBLEBEE_USERS_DIR", usersDir)
	setTestHome(t, realHomes[0])

	for _, h := range realHomes {
		if err := os.MkdirAll(filepath.Join(h, "code"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	roots, _, err := resolveRoots(model.ProfileProject, nil, rootsOpts{AllUsers: true})
	if err != nil {
		t.Fatalf("resolveRoots project --all-users: %v", err)
	}
	want := map[string]bool{}
	for _, h := range realHomes {
		want[filepath.Join(h, "code")] = true
	}
	for _, r := range roots {
		delete(want, r.Path)
		for _, h := range realHomes {
			if r.Path == h {
				t.Errorf("project --all-users emitted bare home %q", h)
			}
		}
	}
	if len(want) > 0 {
		t.Errorf("project --all-users missed: %v", want)
	}
}

func TestRunRootsBaselinePrintsWindowsCurrentUserRoots(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	msixClaude := filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	psUserModules := filepath.Join(home, "Documents", "PowerShell", "Modules")
	psAllUsersModules := filepath.Join(programFiles, "PowerShell", "Modules")
	chromeDefaultExt := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions")
	edgeDefaultExt := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	want := map[string]string{
		filepath.Join(home, "go"):        model.RootKindUserPackage,
		psUserModules:                    model.RootKindUserPackage,
		psAllUsersModules:                model.RootKindGlobalPackage,
		filepath.Join(appData, "Claude"): model.RootKindMCPConfig,
		msixClaude:                       model.RootKindMCPConfig,
		chromeDefaultExt:                 model.RootKindBrowserExtension,
		edgeDefaultExt:                   model.RootKindBrowserExtension,
		firefoxProfiles:                  model.RootKindBrowserExtension,
	}
	for p := range want {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	stdout, stderr, code := captureStdoutStderr(t, func() int {
		return runRoots([]string{"--profile", "baseline"})
	})
	if code != 0 {
		t.Fatalf("runRoots exit code = %d, want 0; stderr=%s", code, stderr)
	}
	got := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			t.Fatalf("bad roots line %q in stdout:\n%s", line, stdout)
		}
		got[parts[1]] = parts[0]
	}
	for p, kind := range want {
		gotKind, ok := got[p]
		if !ok {
			t.Errorf("roots output missing %q (stdout=%s)", p, stdout)
			continue
		}
		if gotKind != kind {
			t.Errorf("roots output %q kind = %q, want %q", p, gotKind, kind)
		}
	}
}
