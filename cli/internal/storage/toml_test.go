package storage

import (
	"os"
	"testing"
)

func TestStarterLibraryIsCompleteAndCanBeSaved(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	commands := defaults()
	if len(commands) != 25 {
		t.Fatalf("starter library has %d commands; want 25", len(commands))
	}
	counts := map[string]int{}
	for _, command := range commands {
		counts[command.Category]++
		if command.Name == "" || command.Description == "" || command.Command == "" || len(command.Tags) == 0 {
			t.Errorf("incomplete starter command: %+v", command)
		}
	}
	for _, category := range []string{"Git", "Docker", "System", "Network", "Tools"} {
		if counts[category] != 5 {
			t.Errorf("%s has %d starter commands; want 5", category, counts[category])
		}
	}
	if err := Save(commands); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != len(commands) {
		t.Fatalf("loaded %d commands; want %d", len(loaded), len(commands))
	}
	if _, err := os.Stat(Path()); err != nil {
		t.Fatalf("saved database is unavailable: %v", err)
	}
}
