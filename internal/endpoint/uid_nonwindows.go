//go:build !windows

package endpoint

import (
	"os"
	"strconv"
)

func fallbackUID() string {
	return strconv.Itoa(os.Getuid())
}
