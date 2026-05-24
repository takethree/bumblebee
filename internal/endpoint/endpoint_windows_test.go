package endpoint

import (
	"regexp"
	"strings"
	"testing"
)

var windowsSIDPattern = regexp.MustCompile(`^S-\d-\d+(?:-\d+)+$`)

func TestCurrentWindowsUserIdentity(t *testing.T) {
	ep := Current("")
	if strings.TrimSpace(ep.Username) == "" {
		t.Fatal("Username is empty")
	}
	if !windowsSIDPattern.MatchString(ep.UID) {
		t.Fatal("UID is not a Windows SID")
	}
	if ep.UID == "-1" {
		t.Fatal("UID used the Windows os.Getuid fallback value")
	}
}

func TestFallbackUIDOnWindowsDoesNotInventPOSIXUID(t *testing.T) {
	if got := fallbackUID(); got != "" {
		t.Fatal("fallbackUID should be empty on Windows")
	}
}
