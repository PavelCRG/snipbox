package history

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRecentReturnsUniqueCommandsNewestFirst(t *testing.T) {
	home := t.TempDir()
	appData := filepath.Join(home, "appdata")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", appData)
	path := filepath.Join(home, ".bash_history")
	if runtime.GOOS == "windows" {
		path = filepath.Join(appData, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("first\nsecond\nfirst\nthird\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got := Recent()
	if strings.Join(got, ",") != "third,first,second" {
		t.Fatalf("Recent() = %v; want newest unique commands first", got)
	}
}
