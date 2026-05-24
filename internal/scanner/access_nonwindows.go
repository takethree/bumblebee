//go:build !windows

package scanner

func isPlatformExpectedAccessError(err error) bool {
	return false
}
