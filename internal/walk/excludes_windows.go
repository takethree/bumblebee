//go:build windows

package walk

import (
	"path/filepath"
	"strings"
)

func platformDefaultExcludes() []string {
	return []string{
		// Windows browser, credential, cloud-sync, and cache trees. These
		// are noisy under explicit broad roots and contain private user data
		// the metadata scanners do not need. Curated baseline roots point at
		// the precise extension/config locations that remain in scope.
		"AppData/Local/Google/Chrome/User Data",
		"AppData/Local/Microsoft/Edge/User Data",
		"AppData/Local/Chromium/User Data",
		"AppData/Local/BraveSoftware/Brave-Browser/User Data",
		"AppData/Local/Vivaldi/User Data",
		"AppData/Local/Perplexity/Comet/User Data",
		"AppData/Local/Packages/TheBrowserCompany.Arc_ttt1ap7aakyb4/LocalCache/Local/Arc/User Data",
		"AppData/Roaming/Microsoft/Credentials",
		"AppData/Local/Microsoft/Credentials",
		"AppData/Roaming/Microsoft/Protect",
		"AppData/Local/Microsoft/Protect",
		"AppData/Roaming/Microsoft/Vault",
		"AppData/Local/Microsoft/Vault",
		"AppData/Roaming/Microsoft/Crypto/RSA",
		"AppData/Local/Microsoft/Crypto/RSA",
		"AppData/Local/Temp",
		"AppData/Local/Microsoft/Windows/INetCache",
		"AppData/Local/Microsoft/Windows/WebCache",
		"AppData/Local/Packages",
		"Google Drive",
		"Dropbox",
	}
}

func isPlatformExcludedDir(fullPath, base string) bool {
	return isCloudSyncDir(base) || isFirefoxProfileSensitiveDir(fullPath, base)
}

func isCloudSyncDir(base string) bool {
	return base == "OneDrive" || strings.HasPrefix(base, "OneDrive - ")
}

func isFirefoxProfileSensitiveDir(fullPath, base string) bool {
	switch base {
	case "cache2", "storage", "sessionstore-backups", "extensions":
	default:
		return false
	}
	p := filepath.ToSlash(filepath.Clean(fullPath))
	return strings.Contains(p, "/AppData/Roaming/Mozilla/Firefox/Profiles/") ||
		strings.Contains(p, "/AppData/Roaming/LibreWolf/Profiles/") ||
		strings.Contains(p, "/AppData/Roaming/Waterfox/Waterfox/Profiles/") ||
		strings.Contains(p, "/AppData/Roaming/Waterfox/Profiles/") ||
		strings.Contains(p, "/Mozilla/Firefox/Profiles/") ||
		strings.Contains(p, "/LibreWolf/Profiles/") ||
		strings.Contains(p, "/Waterfox/Waterfox/Profiles/") ||
		strings.Contains(p, "/Waterfox/Profiles/") ||
		strings.Contains(p, "/.mozilla/firefox/")
}
