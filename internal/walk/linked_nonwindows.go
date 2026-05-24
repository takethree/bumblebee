//go:build !windows

package walk

import "os"

func isPlatformLinkedDir(info os.FileInfo) bool {
	return false
}
