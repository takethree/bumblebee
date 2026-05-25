//go:build windows

package scanner

import (
	"path/filepath"
	"strings"
)

func isSensitiveBrowserProfileFile(path, base string) bool {
	p := filepath.ToSlash(filepath.Clean(path))
	if !isBrowserProfilePath(p) {
		return false
	}
	switch base {
	case "Cookies", "Cookies-journal", "Login Data", "Login Data For Account", "History", "Web Data":
		return true
	case "cookies.sqlite", "places.sqlite", "favicons.sqlite", "formhistory.sqlite", "key4.db", "logins.json", "sessionstore.jsonlz4":
		return true
	}
	return strings.HasPrefix(base, "cookies.sqlite-") ||
		strings.HasPrefix(base, "places.sqlite-") ||
		strings.HasPrefix(base, "favicons.sqlite-") ||
		strings.HasPrefix(base, "formhistory.sqlite-") ||
		strings.HasPrefix(base, "sessionstore")
}

func isBrowserProfilePath(path string) bool {
	for _, marker := range []string{
		"/AppData/Local/Google/Chrome/User Data/",
		"/AppData/Local/Microsoft/Edge/User Data/",
		"/AppData/Local/Chromium/User Data/",
		"/AppData/Local/BraveSoftware/Brave-Browser/User Data/",
		"/AppData/Local/Vivaldi/User Data/",
		"/AppData/Roaming/Mozilla/Firefox/Profiles/",
		"/AppData/Roaming/LibreWolf/Profiles/",
		"/AppData/Roaming/Waterfox/Waterfox/Profiles/",
		"/AppData/Roaming/Waterfox/Profiles/",
		"/Mozilla/Firefox/Profiles/",
		"/LibreWolf/Profiles/",
		"/Waterfox/Waterfox/Profiles/",
		"/Waterfox/Profiles/",
		"/.mozilla/firefox/",
	} {
		if strings.Contains(path, marker) {
			return true
		}
	}
	return false
}
