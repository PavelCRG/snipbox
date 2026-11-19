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
	if len(commands) != 60 {
		t.Fatalf("starter library has %d commands; want 60", len(commands))
	}
	counts := map[string]int{}
	names := map[string]bool{}
	commandText := map[string]string{}
	subcategories := map[string]map[string]int{}
	for _, command := range commands {
		counts[command.Category]++
		if names[command.Name] {
			t.Errorf("duplicate starter command name %q", command.Name)
		}
		names[command.Name] = true
		if previous, ok := commandText[command.Command]; ok {
			t.Errorf("duplicate starter command text %q used by %q and %q", command.Command, previous, command.Name)
		}
		commandText[command.Command] = command.Name
		if subcategories[command.Category] == nil {
			subcategories[command.Category] = map[string]int{}
		}
		subcategories[command.Category][command.Subcategory]++
		if command.Name == "" || command.Description == "" || command.Command == "" || len(command.Tags) == 0 {
			t.Errorf("incomplete starter command: %+v", command)
		}
	}
	for _, name := range []string{"Stop a process", "Inspect file permissions", "Create a ZIP archive", "Inspect listening ports", "Search file contents", "List files recursively", "Restart a service", "Publish the current branch", "Inspect a container", "Measure directory sizes", "Check a TLS endpoint"} {
		if !names[name] {
			t.Errorf("starter library is missing %q", name)
		}
	}
	for category, want := range map[string]int{"Git": 16, "PowerShell": 14, "Docker": 10, "Linux/Bash": 10, "Network": 10} {
		if counts[category] != want {
			t.Errorf("%s has %d starter commands; want %d", category, counts[category], want)
		}
		if len(subcategories[category]) == 0 {
			t.Errorf("%s has no subcategories", category)
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

func TestMigrationKeepsCustomCommandsAndSeedsNewLibrary(t *testing.T) {
	commands, changed := migrate([]Command{{Name: "My command", Category: "Tools", Command: "echo custom", Description: "Custom", Tags: []string{"mine"}}})
	if !changed {
		t.Fatal("migration should update legacy categories and seed missing starter commands")
	}
	if len(commands) != 61 {
		t.Fatalf("migrated command count = %d, want 61", len(commands))
	}
	if commands[0].Category != "Linux/Bash" || commands[0].Subcategory != "Files & Search" {
		t.Fatalf("legacy command was not migrated: %+v", commands[0])
	}
	if commands[0].Command != "echo custom" {
		t.Fatal("migration must preserve custom command content")
	}
}
