//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplexityai/bumblebee/internal/model"
	"github.com/perplexityai/bumblebee/internal/scanner"
)

func setTestCurrentUserDocumentsDir(t *testing.T, dir string) {
	t.Helper()
	old := currentUserDocumentsDir
	currentUserDocumentsDir = func() string { return dir }
	t.Cleanup(func() { currentUserDocumentsDir = old })
}

func testWSLLikePath(p string) bool {
	p = strings.ToLower(filepath.ToSlash(filepath.Clean(p)))
	return strings.HasPrefix(p, "//wsl$/") ||
		strings.HasPrefix(p, "//wsl.localhost/") ||
		strings.Contains(p, "/localstate/rootfs/")
}

func assertNoWSLLikeRoots(t *testing.T, roots []scanner.Root) {
	t.Helper()
	for _, r := range roots {
		if testWSLLikePath(r.Path) {
			t.Fatalf("default roots included WSL-looking path %q", r.Path)
		}
	}
}

func TestResolveRootsBaselineIncludesWindowsCurrentUserRoots(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	redirectedDocuments := filepath.Join(t.TempDir(), "OneDrive - Contoso", "Documents")
	setTestCurrentUserDocumentsDir(t, redirectedDocuments)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	programFilesX86 := filepath.Join(t.TempDir(), "Program Files (x86)")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	t.Setenv("ProgramFiles(x86)", programFilesX86)
	msixClaude := filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	psUserModules := filepath.Join(home, "Documents", "PowerShell", "Modules")
	winPsUserModules := filepath.Join(home, "Documents", "WindowsPowerShell", "Modules")
	redirectedPsUserModules := filepath.Join(redirectedDocuments, "PowerShell", "Modules")
	redirectedWinPsUserModules := filepath.Join(redirectedDocuments, "WindowsPowerShell", "Modules")
	psAllUsersModules := filepath.Join(programFiles, "PowerShell", "Modules")
	winPsAllUsersModules := filepath.Join(programFiles, "WindowsPowerShell", "Modules")
	npmGlobalModules := filepath.Join(appData, "npm", "node_modules")
	pythonUserSite := filepath.Join(appData, "Python", "Python311", "site-packages")
	pythonInstallUserSite := filepath.Join(localAppData, "Programs", "Python", "Python314", "Lib", "site-packages")
	pythonInstallGlobalSite := filepath.Join(programFiles, "Python 3.14", "Lib", "site-packages")
	pythonInstallGlobalX86Site := filepath.Join(programFilesX86, "Python 3.14", "Lib", "site-packages")
	customPythonPrefixSite := filepath.Join(t.TempDir(), "Python314", "Lib", "site-packages")
	pipxHomeVenvs := filepath.Join(home, "pipx", "venvs")
	pipxLocalAppDataVenvs := filepath.Join(localAppData, "pipx", "venvs")
	pipxLegacyVenvs := filepath.Join(home, ".local", "pipx", "venvs")
	chromeDefaultExt := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions")
	chromeProfile1Ext := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Profile 1", "Extensions")
	braveDefaultExt := filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	braveProfile9Ext := filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Profile 9", "Extensions")
	chromiumDefaultExt := filepath.Join(localAppData, "Chromium", "User Data", "Default", "Extensions")
	chromiumProfile1Ext := filepath.Join(localAppData, "Chromium", "User Data", "Profile 1", "Extensions")
	edgeDefaultExt := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions")
	edgeProfile1Ext := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions")
	vivaldiDefaultExt := filepath.Join(localAppData, "Vivaldi", "User Data", "Default", "Extensions")
	vivaldiProfile9Ext := filepath.Join(localAppData, "Vivaldi", "User Data", "Profile 9", "Extensions")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	librewolfProfiles := filepath.Join(appData, "LibreWolf", "Profiles")
	waterfoxProfiles := filepath.Join(appData, "Waterfox", "Waterfox", "Profiles")
	waterfoxLegacyProfiles := filepath.Join(appData, "Waterfox", "Profiles")

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
		redirectedPsUserModules:          model.RootKindUserPackage,
		redirectedWinPsUserModules:       model.RootKindUserPackage,
		psAllUsersModules:                model.RootKindGlobalPackage,
		winPsAllUsersModules:             model.RootKindGlobalPackage,
		npmGlobalModules:                 model.RootKindUserPackage,
		pythonUserSite:                   model.RootKindUserPackage,
		pythonInstallUserSite:            model.RootKindUserPackage,
		pythonInstallGlobalSite:          model.RootKindGlobalPackage,
		pythonInstallGlobalX86Site:       model.RootKindGlobalPackage,
		pipxHomeVenvs:                    model.RootKindUserPackage,
		pipxLocalAppDataVenvs:            model.RootKindUserPackage,
		pipxLegacyVenvs:                  model.RootKindUserPackage,
		filepath.Join(appData, "Claude"): model.RootKindMCPConfig,
		msixClaude:                       model.RootKindMCPConfig,
		chromeDefaultExt:                 model.RootKindBrowserExtension,
		chromeProfile1Ext:                model.RootKindBrowserExtension,
		braveDefaultExt:                  model.RootKindBrowserExtension,
		braveProfile9Ext:                 model.RootKindBrowserExtension,
		chromiumDefaultExt:               model.RootKindBrowserExtension,
		chromiumProfile1Ext:              model.RootKindBrowserExtension,
		edgeDefaultExt:                   model.RootKindBrowserExtension,
		edgeProfile1Ext:                  model.RootKindBrowserExtension,
		vivaldiDefaultExt:                model.RootKindBrowserExtension,
		vivaldiProfile9Ext:               model.RootKindBrowserExtension,
		firefoxProfiles:                  model.RootKindBrowserExtension,
		librewolfProfiles:                model.RootKindBrowserExtension,
		waterfoxProfiles:                 model.RootKindBrowserExtension,
		waterfoxLegacyProfiles:           model.RootKindBrowserExtension,
	}
	for p := range want {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(customPythonPrefixSite, 0o755); err != nil {
		t.Fatal(err)
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
	if _, ok := got[customPythonPrefixSite]; ok {
		t.Errorf("baseline unexpectedly included custom Python prefix root %q", customPythonPrefixSite)
	}
}

func TestResolveRootsWindowsDefaultsDoNotIncludeWSLPaths(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestCurrentUserDocumentsDir(t, filepath.Join(home, "Documents"))
	appData := filepath.Join(home, "AppData", "Roaming")
	localAppData := filepath.Join(home, "AppData", "Local")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)

	for _, p := range []string{
		filepath.Join(home, "go"),
		filepath.Join(appData, "npm", "node_modules"),
		filepath.Join(localAppData, "Packages", "CanonicalGroupLimited.Ubuntu_79rhkp1fndgsc", "LocalState", "rootfs", "home", "alice", "go"),
		filepath.Join(localAppData, "Packages", "CanonicalGroupLimited.Ubuntu_79rhkp1fndgsc", "LocalState", "rootfs", "home", "alice", ".vscode", "extensions"),
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}
	assertNoWSLLikeRoots(t, roots)

	usersDir, realHomes := fakeUsersDir(t, []string{"alice"}, nil)
	t.Setenv("BUMBLEBEE_USERS_DIR", usersDir)
	setTestHome(t, realHomes[0])
	wslRootfs := filepath.Join(realHomes[0], "AppData", "Local", "Packages", "CanonicalGroupLimited.Ubuntu_79rhkp1fndgsc", "LocalState", "rootfs", "home", "alice", "go")
	if err := os.MkdirAll(wslRootfs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(realHomes[0], "go"), 0o755); err != nil {
		t.Fatal(err)
	}

	allUserRoots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{AllUsers: true})
	if err != nil {
		t.Fatalf("resolveRoots baseline --all-users: %v", err)
	}
	assertNoWSLLikeRoots(t, allUserRoots)
}

func TestResolveRootsWindowsExplicitWSLRootRemainsGeneric(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	explicit := `\\wsl$\Ubuntu\home\alice\repo`

	roots, _, err := resolveRoots(model.ProfileProject, []string{explicit}, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots explicit WSL-like root: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("roots length = %d, want 1", len(roots))
	}
	if roots[0].Path != explicit {
		t.Errorf("explicit root path = %q, want %q", roots[0].Path, explicit)
	}
	if roots[0].Kind != model.RootKindProject {
		t.Errorf("explicit WSL-like root kind = %q, want %q", roots[0].Kind, model.RootKindProject)
	}
}

func TestResolveRootsBaselineSkipsAbsentWindowsCandidates(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	redirectedDocuments := filepath.Join(t.TempDir(), "OneDrive - Contoso", "Documents")
	setTestCurrentUserDocumentsDir(t, redirectedDocuments)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	programFilesX86 := filepath.Join(t.TempDir(), "Program Files (x86)")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	t.Setenv("ProgramFiles(x86)", programFilesX86)
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
		filepath.Join(redirectedDocuments, "PowerShell", "Modules"),
		filepath.Join(redirectedDocuments, "WindowsPowerShell", "Modules"),
		filepath.Join(programFiles, "PowerShell", "Modules"),
		filepath.Join(programFiles, "WindowsPowerShell", "Modules"),
		filepath.Join(appData, "npm", "node_modules"),
		filepath.Join(appData, "Python", "Python311", "site-packages"),
		filepath.Join(home, "pipx", "venvs"),
		filepath.Join(localAppData, "pipx", "venvs"),
		filepath.Join(home, ".local", "pipx", "venvs"),
		filepath.Join(appData, "Claude"),
		msixClaude,
		filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Profile 9", "Extensions"),
		filepath.Join(localAppData, "Chromium", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "Chromium", "User Data", "Profile 1", "Extensions"),
		filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "Vivaldi", "User Data", "Default", "Extensions"),
		filepath.Join(localAppData, "Vivaldi", "User Data", "Profile 9", "Extensions"),
		filepath.Join(appData, "Mozilla", "Firefox", "Profiles"),
		filepath.Join(appData, "LibreWolf", "Profiles"),
		filepath.Join(appData, "Waterfox", "Waterfox", "Profiles"),
		filepath.Join(appData, "Waterfox", "Profiles"),
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

func TestResolveRootsBaselineDeduplicatesKnownDocumentsWindows(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestCurrentUserDocumentsDir(t, filepath.Join(home, "Documents"))
	psUserModules := filepath.Join(home, "Documents", "PowerShell", "Modules")
	if err := os.MkdirAll(psUserModules, 0o755); err != nil {
		t.Fatal(err)
	}

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}
	count := 0
	for _, r := range roots {
		if r.Path == psUserModules {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("PowerShell module root count = %d, want 1 (roots=%v)", count, roots)
	}
}

func TestClassifyRootWindowsBrowserExtensions(t *testing.T) {
	cases := []string{
		`C:\Users\alice\AppData\Local\Google\Chrome\User Data\Default\Extensions`,
		`C:\Users\alice\AppData\Local\BraveSoftware\Brave-Browser\User Data\Default\Extensions`,
		`C:\Users\alice\AppData\Local\BraveSoftware\Brave-Browser\User Data\Profile 9\Extensions`,
		`C:\Users\alice\AppData\Local\Chromium\User Data\Default\Extensions`,
		`C:\Users\alice\AppData\Local\Chromium\User Data\Profile 1\Extensions`,
		`C:\Users\alice\AppData\Local\Microsoft\Edge\User Data\Profile 1\Extensions`,
		`C:\Users\alice\AppData\Local\Vivaldi\User Data\Default\Extensions`,
		`C:\Users\alice\AppData\Local\Vivaldi\User Data\Profile 9\Extensions`,
		`C:\Users\alice\AppData\Roaming\Mozilla\Firefox\Profiles`,
		`C:\Users\alice\AppData\Roaming\LibreWolf\Profiles`,
		`C:\Users\alice\AppData\Roaming\Waterfox\Waterfox\Profiles`,
		`C:\Users\alice\AppData\Roaming\Waterfox\Profiles`,
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

func TestClassifyRootWindowsPythonInstallSites(t *testing.T) {
	userCases := []string{
		`C:\Users\alice\AppData\Local\Programs\Python\Python314\Lib\site-packages`,
		`C:\Users\alice\AppData\Local\Programs\Python\Python314-64\Lib\site-packages`,
	}
	for _, p := range userCases {
		if got := classifyRoot(p, model.ProfileBaseline); got != model.RootKindUserPackage {
			t.Errorf("classifyRoot(%q) = %q, want %q", p, got, model.RootKindUserPackage)
		}
	}
	globalCases := []string{
		`C:\Program Files\Python 3.14\Lib\site-packages`,
		`C:\Program Files (x86)\Python 3.14\Lib\site-packages`,
	}
	for _, p := range globalCases {
		if got := classifyRoot(p, model.ProfileBaseline); got != model.RootKindGlobalPackage {
			t.Errorf("classifyRoot(%q) = %q, want %q", p, got, model.RootKindGlobalPackage)
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
	braveDefault := filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default")
	chromiumDefault := filepath.Join(localAppData, "Chromium", "User Data", "Default")
	edgeDefault := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default")
	vivaldiDefault := filepath.Join(localAppData, "Vivaldi", "User Data", "Default")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	firefoxProfile := filepath.Join(firefoxProfiles, "abcd.default-release")
	librewolfProfiles := filepath.Join(appData, "LibreWolf", "Profiles")
	waterfoxProfiles := filepath.Join(appData, "Waterfox", "Waterfox", "Profiles")
	waterfoxLegacyProfiles := filepath.Join(appData, "Waterfox", "Profiles")
	wantRoots := []string{
		filepath.Join(chromeDefault, "Extensions"),
		filepath.Join(braveDefault, "Extensions"),
		filepath.Join(chromiumDefault, "Extensions"),
		filepath.Join(edgeDefault, "Extensions"),
		filepath.Join(vivaldiDefault, "Extensions"),
		firefoxProfiles,
		librewolfProfiles,
		waterfoxProfiles,
		waterfoxLegacyProfiles,
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
		filepath.Dir(braveDefault),
		braveDefault,
		filepath.Join(braveDefault, "Cookies"),
		filepath.Join(braveDefault, "Login Data"),
		filepath.Join(braveDefault, "History"),
		filepath.Dir(chromiumDefault),
		chromiumDefault,
		filepath.Join(chromiumDefault, "Cookies"),
		filepath.Join(chromiumDefault, "Login Data"),
		filepath.Join(chromiumDefault, "History"),
		filepath.Dir(edgeDefault),
		edgeDefault,
		filepath.Join(edgeDefault, "Cookies"),
		filepath.Join(edgeDefault, "Login Data"),
		filepath.Join(edgeDefault, "History"),
		filepath.Join(edgeDefault, "Local Storage"),
		filepath.Join(edgeDefault, "IndexedDB"),
		filepath.Join(edgeDefault, "Cache"),
		filepath.Dir(vivaldiDefault),
		vivaldiDefault,
		filepath.Join(vivaldiDefault, "Cookies"),
		filepath.Join(vivaldiDefault, "Login Data"),
		filepath.Join(vivaldiDefault, "History"),
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
	redirectedDocuments := filepath.Join(t.TempDir(), "OneDrive - Contoso", "Documents")
	setTestCurrentUserDocumentsDir(t, redirectedDocuments)
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
			filepath.Join(roaming, "npm", "node_modules"),
			filepath.Join(roaming, "Python", "Python311", "site-packages"),
			filepath.Join(local, "Programs", "Python", "Python314", "Lib", "site-packages"),
			filepath.Join(h, "pipx", "venvs"),
			filepath.Join(local, "pipx", "venvs"),
			filepath.Join(h, ".local", "pipx", "venvs"),
			filepath.Join(h, ".vscode", "extensions"),
			filepath.Join(roaming, "Claude"),
			filepath.Join(local, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"),
			filepath.Join(local, "Google", "Chrome", "User Data", "Default", "Extensions"),
			filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions"),
			filepath.Join(local, "Chromium", "User Data", "Default", "Extensions"),
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions"),
			filepath.Join(local, "Vivaldi", "User Data", "Default", "Extensions"),
			filepath.Join(roaming, "Mozilla", "Firefox", "Profiles"),
			filepath.Join(roaming, "LibreWolf", "Profiles"),
			filepath.Join(roaming, "Waterfox", "Waterfox", "Profiles"),
			filepath.Join(roaming, "Waterfox", "Profiles"),
		} {
			mustMkdir(p)
		}
	}
	redirectedPsUserModules := filepath.Join(redirectedDocuments, "PowerShell", "Modules")
	redirectedWinPsUserModules := filepath.Join(redirectedDocuments, "WindowsPowerShell", "Modules")
	mustMkdir(redirectedPsUserModules)
	mustMkdir(redirectedWinPsUserModules)
	otherUserOneDrivePsModules := filepath.Join(realHomes[1], "OneDrive - Contoso", "Documents", "PowerShell", "Modules")
	otherUserOneDriveWinPsModules := filepath.Join(realHomes[1], "OneDrive - Contoso", "Documents", "WindowsPowerShell", "Modules")
	mustMkdir(otherUserOneDrivePsModules)
	mustMkdir(otherUserOneDriveWinPsModules)

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
			filepath.Join(h, "go"):                                                                       model.RootKindUserPackage,
			filepath.Join(h, "Documents", "PowerShell", "Modules"):                                       model.RootKindUserPackage,
			filepath.Join(h, "Documents", "WindowsPowerShell", "Modules"):                                model.RootKindUserPackage,
			filepath.Join(roaming, "npm", "node_modules"):                                                model.RootKindUserPackage,
			filepath.Join(roaming, "Python", "Python311", "site-packages"):                               model.RootKindUserPackage,
			filepath.Join(local, "Programs", "Python", "Python314", "Lib", "site-packages"):              model.RootKindUserPackage,
			filepath.Join(h, "pipx", "venvs"):                                                            model.RootKindUserPackage,
			filepath.Join(local, "pipx", "venvs"):                                                        model.RootKindUserPackage,
			filepath.Join(h, ".local", "pipx", "venvs"):                                                  model.RootKindUserPackage,
			filepath.Join(h, ".vscode", "extensions"):                                                    model.RootKindEditorExtension,
			filepath.Join(roaming, "Claude"):                                                             model.RootKindMCPConfig,
			filepath.Join(local, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"):  model.RootKindMCPConfig,
			filepath.Join(local, "Google", "Chrome", "User Data", "Default", "Extensions"):               model.RootKindBrowserExtension,
			filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions"): model.RootKindBrowserExtension,
			filepath.Join(local, "Chromium", "User Data", "Default", "Extensions"):                       model.RootKindBrowserExtension,
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Profile 1", "Extensions"):            model.RootKindBrowserExtension,
			filepath.Join(local, "Vivaldi", "User Data", "Default", "Extensions"):                        model.RootKindBrowserExtension,
			filepath.Join(roaming, "Mozilla", "Firefox", "Profiles"):                                     model.RootKindBrowserExtension,
			filepath.Join(roaming, "LibreWolf", "Profiles"):                                              model.RootKindBrowserExtension,
			filepath.Join(roaming, "Waterfox", "Waterfox", "Profiles"):                                   model.RootKindBrowserExtension,
			filepath.Join(roaming, "Waterfox", "Profiles"):                                               model.RootKindBrowserExtension,
		}
		if samePath(h, realHomes[0]) {
			want[redirectedPsUserModules] = model.RootKindUserPackage
			want[redirectedWinPsUserModules] = model.RootKindUserPackage
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
	if gotPaths[redirectedPsUserModules] != model.RootKindUserPackage {
		t.Errorf("current user's redirected PowerShell module root missing (roots=%v)", roots)
	}
	redirectedPsCount := 0
	for _, r := range roots {
		if r.Path == redirectedPsUserModules {
			redirectedPsCount++
		}
	}
	if redirectedPsCount != 1 {
		t.Errorf("redirected PowerShell module root count = %d, want 1 for current user only (roots=%v)", redirectedPsCount, roots)
	}
	for _, p := range []string{otherUserOneDrivePsModules, otherUserOneDriveWinPsModules} {
		if _, ok := gotPaths[p]; ok {
			t.Errorf("baseline --all-users discovered other user's OneDrive-redirected Documents root %q; only literal per-home Documents roots are supported", p)
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
	redirectedDocuments := filepath.Join(t.TempDir(), "OneDrive - Contoso", "Documents")
	setTestCurrentUserDocumentsDir(t, redirectedDocuments)
	appData := filepath.Join(t.TempDir(), "Roaming")
	localAppData := filepath.Join(t.TempDir(), "Local")
	programFiles := filepath.Join(t.TempDir(), "Program Files")
	programFilesX86 := filepath.Join(t.TempDir(), "Program Files (x86)")
	t.Setenv("APPDATA", appData)
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("ProgramFiles", programFiles)
	t.Setenv("ProgramFiles(x86)", programFilesX86)
	msixClaude := filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	psUserModules := filepath.Join(home, "Documents", "PowerShell", "Modules")
	redirectedPsUserModules := filepath.Join(redirectedDocuments, "PowerShell", "Modules")
	psAllUsersModules := filepath.Join(programFiles, "PowerShell", "Modules")
	npmGlobalModules := filepath.Join(appData, "npm", "node_modules")
	pythonUserSite := filepath.Join(appData, "Python", "Python311", "site-packages")
	pythonInstallUserSite := filepath.Join(localAppData, "Programs", "Python", "Python314", "Lib", "site-packages")
	pythonInstallGlobalSite := filepath.Join(programFiles, "Python 3.14", "Lib", "site-packages")
	pythonInstallGlobalX86Site := filepath.Join(programFilesX86, "Python 3.14", "Lib", "site-packages")
	pipxHomeVenvs := filepath.Join(home, "pipx", "venvs")
	pipxLocalAppDataVenvs := filepath.Join(localAppData, "pipx", "venvs")
	pipxLegacyVenvs := filepath.Join(home, ".local", "pipx", "venvs")
	chromeDefaultExt := filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions")
	braveDefaultExt := filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	chromiumDefaultExt := filepath.Join(localAppData, "Chromium", "User Data", "Default", "Extensions")
	edgeDefaultExt := filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Extensions")
	vivaldiDefaultExt := filepath.Join(localAppData, "Vivaldi", "User Data", "Default", "Extensions")
	firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	librewolfProfiles := filepath.Join(appData, "LibreWolf", "Profiles")
	waterfoxProfiles := filepath.Join(appData, "Waterfox", "Waterfox", "Profiles")
	waterfoxLegacyProfiles := filepath.Join(appData, "Waterfox", "Profiles")
	want := map[string]string{
		filepath.Join(home, "go"):        model.RootKindUserPackage,
		psUserModules:                    model.RootKindUserPackage,
		redirectedPsUserModules:          model.RootKindUserPackage,
		psAllUsersModules:                model.RootKindGlobalPackage,
		npmGlobalModules:                 model.RootKindUserPackage,
		pythonUserSite:                   model.RootKindUserPackage,
		pythonInstallUserSite:            model.RootKindUserPackage,
		pythonInstallGlobalSite:          model.RootKindGlobalPackage,
		pythonInstallGlobalX86Site:       model.RootKindGlobalPackage,
		pipxHomeVenvs:                    model.RootKindUserPackage,
		pipxLocalAppDataVenvs:            model.RootKindUserPackage,
		pipxLegacyVenvs:                  model.RootKindUserPackage,
		filepath.Join(appData, "Claude"): model.RootKindMCPConfig,
		msixClaude:                       model.RootKindMCPConfig,
		chromeDefaultExt:                 model.RootKindBrowserExtension,
		braveDefaultExt:                  model.RootKindBrowserExtension,
		chromiumDefaultExt:               model.RootKindBrowserExtension,
		edgeDefaultExt:                   model.RootKindBrowserExtension,
		vivaldiDefaultExt:                model.RootKindBrowserExtension,
		firefoxProfiles:                  model.RootKindBrowserExtension,
		librewolfProfiles:                model.RootKindBrowserExtension,
		waterfoxProfiles:                 model.RootKindBrowserExtension,
		waterfoxLegacyProfiles:           model.RootKindBrowserExtension,
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
