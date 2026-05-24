//go:build !windows

package walk

func platformDefaultExcludes() []string {
	return nil
}

func isPlatformExcludedDir(_, _ string) bool {
	return false
}
