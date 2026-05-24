//go:build !windows

package scanner

func isSensitiveBrowserProfileFile(_, _ string) bool {
	return false
}
