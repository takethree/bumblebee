//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/perplexityai/bumblebee/internal/model"
	"github.com/perplexityai/bumblebee/internal/scanner"
)

type knownFolderID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	folderIDDocuments = knownFolderID{
		Data1: 0xFDD39AD0,
		Data2: 0x238F,
		Data3: 0x46AF,
		Data4: [8]byte{0xAD, 0xB4, 0x6C, 0x85, 0x48, 0x03, 0x69, 0xC7},
	}
	shell32                  = syscall.NewLazyDLL("shell32.dll")
	procSHGetKnownFolderPath = shell32.NewProc("SHGetKnownFolderPath")
	ole32                    = syscall.NewLazyDLL("ole32.dll")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
	currentUserDocumentsDir  = windowsCurrentUserDocumentsDir
)

func userHomeDir() string {
	if home := strings.TrimSpace(os.Getenv("USERPROFILE")); home != "" {
		return home
	}
	drive := strings.TrimSpace(os.Getenv("HOMEDRIVE"))
	path := strings.TrimSpace(os.Getenv("HOMEPATH"))
	if drive != "" && path != "" {
		return drive + path
	}
	home, _ := os.UserHomeDir()
	return home
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func isPlatformBroadHomeRoot(abs string) bool {
	vol := filepath.VolumeName(abs)
	if vol != "" && samePath(abs, vol+string(filepath.Separator)) {
		return true
	}
	p := filepath.ToSlash(abs)
	if vol != "" {
		p = strings.TrimPrefix(p, filepath.ToSlash(vol))
	}
	p = strings.TrimPrefix(p, "/")
	parts := strings.Split(p, "/")
	if len(parts) == 1 && strings.EqualFold(parts[0], "Users") {
		return true
	}
	return len(parts) == 2 && strings.EqualFold(parts[0], "Users") && parts[1] != ""
}

func classifyPlatformRoot(p string) (string, bool) {
	switch {
	case (strings.HasSuffix(p, "/Extensions") || strings.HasSuffix(p, "/extensions")) &&
		strings.Contains(p, "Microsoft/Edge"):
		return model.RootKindBrowserExtension, true
	case strings.HasSuffix(p, "/AppData/Roaming/Claude") ||
		strings.HasSuffix(p, "/Packages/Claude_pzs8sxrjxfjjc/LocalCache/Roaming/Claude"):
		return model.RootKindMCPConfig, true
	case strings.HasSuffix(p, "/PowerShell/Modules") ||
		strings.HasSuffix(p, "/WindowsPowerShell/Modules"):
		if strings.Contains(strings.ToLower(p), "/program files/") {
			return model.RootKindGlobalPackage, true
		}
		return model.RootKindUserPackage, true
	case strings.HasSuffix(strings.ToLower(p), "/lib/site-packages") &&
		strings.Contains(strings.ToLower(p), "/programs/python/python"):
		return model.RootKindUserPackage, true
	case strings.HasSuffix(strings.ToLower(p), "/lib/site-packages") &&
		strings.Contains(strings.ToLower(p), "/program files") &&
		strings.Contains(strings.ToLower(p), "/python"):
		return model.RootKindGlobalPackage, true
	}
	return "", false
}

func allUsersExpansionSupported() bool {
	return true
}

func defaultUsersDir() string {
	if drive := strings.TrimSpace(os.Getenv("SystemDrive")); drive != "" {
		return filepath.Join(windowsDriveRoot(drive), "Users")
	}
	if home := userHomeDir(); home != "" {
		if vol := filepath.VolumeName(home); vol != "" {
			return filepath.Join(windowsDriveRoot(vol), "Users")
		}
	}
	return `C:\Users`
}

func windowsDriveRoot(drive string) string {
	drive = strings.TrimSpace(drive)
	if strings.HasSuffix(drive, `\`) || strings.HasSuffix(drive, `/`) {
		return drive
	}
	return drive + `\`
}

func isPlatformServiceUserHomeName(name string) bool {
	switch strings.ToLower(name) {
	case "public", "default", "default user", "all users", "desktop.ini", "defaultaccount", "defaultuser0", "wdagutilityaccount":
		return true
	}
	return false
}

func platformRoamingAppDataDir(home string) string {
	if home == "" {
		return ""
	}
	if current := userHomeDir(); current != "" && samePath(home, current) {
		if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
			return appData
		}
	}
	return filepath.Join(home, "AppData", "Roaming")
}

func platformLocalAppDataDir(home string) string {
	if home == "" {
		return ""
	}
	if current := userHomeDir(); current != "" && samePath(home, current) {
		if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
			return localAppData
		}
	}
	return filepath.Join(home, "AppData", "Local")
}

func windowsCurrentUserDocumentsDir() string {
	var pathPtr *uint16
	ret, _, _ := procSHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(&folderIDDocuments)),
		0,
		0,
		uintptr(unsafe.Pointer(&pathPtr)),
	)
	if pathPtr != nil {
		defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(pathPtr)))
	}
	if ret != 0 || pathPtr == nil {
		return ""
	}
	return strings.TrimSpace(utf16PtrToString(pathPtr))
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var s []uint16
	for *p != 0 {
		s = append(s, *p)
		p = (*uint16)(unsafe.Add(unsafe.Pointer(p), unsafe.Sizeof(*p)))
	}
	return string(utf16.Decode(s))
}

func platformDocumentsDirs(home string) []string {
	var dirs []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		for _, existing := range dirs {
			if samePath(existing, p) {
				return
			}
		}
		dirs = append(dirs, p)
	}
	add(filepath.Join(home, "Documents"))
	if current := userHomeDir(); current != "" && samePath(home, current) {
		add(currentUserDocumentsDir())
	}
	return dirs
}

func platformBaselineHomeCandidates(home string) []scanner.Root {
	var roots []scanner.Root
	for _, documents := range platformDocumentsDirs(home) {
		roots = append(roots,
			scanner.Root{Path: filepath.Join(documents, "PowerShell", "Modules"), Kind: model.RootKindUserPackage},
			scanner.Root{Path: filepath.Join(documents, "WindowsPowerShell", "Modules"), Kind: model.RootKindUserPackage},
		)
	}
	appData := platformRoamingAppDataDir(home)
	if appData != "" {
		roots = append(roots,
			scanner.Root{Path: filepath.Join(appData, "npm", "node_modules"), Kind: model.RootKindUserPackage},
			scanner.Root{Path: filepath.Join(appData, "Claude"), Kind: model.RootKindMCPConfig},
		)
		for _, p := range globExisting(filepath.Join(appData, "Python", "Python*", "site-packages")) {
			roots = append(roots, scanner.Root{Path: p, Kind: model.RootKindUserPackage})
		}
	}
	localAppData := platformLocalAppDataDir(home)
	if localAppData != "" {
		roots = append(roots, scanner.Root{
			Path: filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"),
			Kind: model.RootKindMCPConfig,
		})
		for _, p := range globExisting(filepath.Join(localAppData, "Programs", "Python", "Python*", "Lib", "site-packages")) {
			roots = append(roots, scanner.Root{Path: p, Kind: model.RootKindUserPackage})
		}
		roots = append(roots, scanner.Root{Path: filepath.Join(localAppData, "pipx", "venvs"), Kind: model.RootKindUserPackage})
	}
	roots = append(roots,
		scanner.Root{Path: filepath.Join(home, "pipx", "venvs"), Kind: model.RootKindUserPackage},
		scanner.Root{Path: filepath.Join(home, ".local", "pipx", "venvs"), Kind: model.RootKindUserPackage},
	)
	return roots
}

func platformBrowserExtensionCandidateRoots(home string) []string {
	var roots []string
	chromiumProfiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3", "Profile 4", "Profile 5", "Profile 6", "Profile 7", "Profile 8", "Profile 9"}
	localAppData := platformLocalAppDataDir(home)
	if localAppData != "" {
		for _, base := range []string{
			filepath.Join(localAppData, "Google", "Chrome", "User Data"),
			filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"),
			filepath.Join(localAppData, "Chromium", "User Data"),
			filepath.Join(localAppData, "Microsoft", "Edge", "User Data"),
			filepath.Join(localAppData, "Vivaldi", "User Data"),
			filepath.Join(localAppData, "Perplexity", "Comet", "User Data"),
			filepath.Join(localAppData, "Packages", "TheBrowserCompany.Arc_ttt1ap7aakyb4", "LocalCache", "Local", "Arc", "User Data"),
		} {
			for _, prof := range chromiumProfiles {
				roots = append(roots, filepath.Join(base, prof, "Extensions"))
			}
		}
	}
	appData := platformRoamingAppDataDir(home)
	if appData != "" {
		roots = append(roots,
			filepath.Join(appData, "Mozilla", "Firefox", "Profiles"),
			filepath.Join(appData, "LibreWolf", "Profiles"),
			filepath.Join(appData, "Waterfox", "Waterfox", "Profiles"),
			filepath.Join(appData, "Waterfox", "Profiles"),
		)
	}
	return roots
}
