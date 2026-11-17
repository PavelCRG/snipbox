package history

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Recent() []string {
	h, _ := os.UserHomeDir()
	paths := []string{filepath.Join(h, ".zsh_history"), filepath.Join(h, ".bash_history")}
	if runtime.GOOS == "windows" {
		paths = append(paths, filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"))
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		f, e := os.Open(p)
		if e != nil {
			continue
		}
		var lines []string
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 64*1024), 1024*1024)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if i := strings.Index(line, "; "); strings.HasPrefix(line, ": ") && i >= 0 {
				line = line[i+2:]
			}
			if line != "" {
				for i, previous := range lines {
					if previous == line {
						lines = append(lines[:i], lines[i+1:]...)
						break
					}
				}
				lines = append(lines, line)
				if len(lines) > 20 {
					lines = lines[1:]
				}
			}
		}
		f.Close()
		for i := len(lines) - 1; i >= 0 && len(out) < 20; i-- {
			if !seen[lines[i]] {
				seen[lines[i]] = true
				out = append(out, lines[i])
			}
		}
	}
	return out
}
