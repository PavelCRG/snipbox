package syncer

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/snipbox/cli/internal/storage"
)

func Push() error {
	dir := filepath.Dir(storage.Path())
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "sync"}, {"push"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if b, e := c.CombinedOutput(); e != nil {
			return fmt.Errorf("git %v: %s: %w", args, b, e)
		}
	}
	return nil
}
