//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/perplexityai/bumblebee/internal/model"
	"github.com/perplexityai/bumblebee/internal/scanner"
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

func platformBaselineHomeCandidates(home string) []scanner.Root {
	var roots []scanner.Root
	appData := platformRoamingAppDataDir(home)
	if appData != "" {
		roots = append(roots, scanner.Root{Path: filepath.Join(appData, "Claude"), Kind: model.RootKindMCPConfig})
	}
	localAppData := platformLocalAppDataDir(home)
	if localAppData != "" {
		roots = append(roots, scanner.Root{
			Path: filepath.Join(localAppData, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude"),
			Kind: model.RootKindMCPConfig,
		})
	}
	return roots
}

func platformBrowserExtensionCandidateRoots(home string) []string {
	var roots []string
	chromiumProfiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3", "Profile 4", "Profile 5", "Profile 6", "Profile 7", "Profile 8", "Profile 9"}
	localAppData := platformLocalAppDataDir(home)
	if localAppData != "" {
		for _, base := range []string{
			filepath.Join(localAppData, "Google", "Chrome", "User Data"),
			filepath.Join(localAppData, "Microsoft", "Edge", "User Data"),
		} {
			for _, prof := range chromiumProfiles {
				roots = append(roots, filepath.Join(base, prof, "Extensions"))
			}
		}
	}
	appData := platformRoamingAppDataDir(home)
	if appData != "" {
		roots = append(roots, filepath.Join(appData, "Mozilla", "Firefox", "Profiles"))
	}
	return roots
}
