//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/perplexityai/bumblebee/internal/scanner"
)

func userHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func isPlatformBroadHomeRoot(_ string) bool {
	return false
}

func classifyPlatformRoot(_ string) (string, bool) {
	return "", false
}

func allUsersExpansionSupported() bool {
	return runtime.GOOS == "darwin"
}

func defaultUsersDir() string {
	return "/Users"
}

func isPlatformServiceUserHomeName(_ string) bool {
	return false
}

func platformRoamingAppDataDir(_ string) string {
	return ""
}

func platformLocalAppDataDir(_ string) string {
	return ""
}

func platformBaselineHomeCandidates(_ string) []scanner.Root {
	return nil
}

func platformBrowserExtensionCandidateRoots(_ string) []string {
	return nil
}
