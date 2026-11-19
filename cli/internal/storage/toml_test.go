package storage

import (
	"os"
	"path/filepath"
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
		if len(command.Shells) == 0 || len(command.Platforms) == 0 {
			t.Errorf("starter command %q has no runtime metadata", command.Name)
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
	if len(loaded[0].Shells) == 0 || len(loaded[0].Platforms) == 0 {
		t.Fatalf("runtime metadata was not preserved by TOML round-trip: %+v", loaded[0])
	}
	if _, err := os.Stat(Path()); err != nil {
		t.Fatalf("saved database is unavailable: %v", err)
	}
}

func TestStarterMetadataMarksLinuxOnlyUtilities(t *testing.T) {
	for _, command := range defaults() {
		if command.Name == "Show memory usage" && (len(command.Platforms) != 1 || command.Platforms[0] != "Linux" || len(command.Requirements) != 1 || command.Requirements[0] != "procps") {
			t.Errorf("Linux memory command metadata = %+v", command)
		}
		if command.Name == "Show listening sockets" && (len(command.Platforms) != 1 || command.Platforms[0] != "Linux" || len(command.Requirements) != 1 || command.Requirements[0] != "iproute2") {
			t.Errorf("Linux socket command metadata = %+v", command)
		}
		if command.Name == "Measure directory sizes" && (len(command.Platforms) != 1 || command.Platforms[0] != "Linux") {
			t.Errorf("GNU du command metadata = %+v", command)
		}
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

func TestMigrationAddsRuntimeMetadataToUnchangedStarterCommands(t *testing.T) {
	legacy := defaults()[0]
	legacy.Shells, legacy.Platforms, legacy.Requirements = nil, nil, nil
	legacy.Description = "My notes for this command"
	legacy.Favorite, legacy.Usage = true, 14
	commands, changed := migrate([]Command{legacy})
	if !changed {
		t.Fatal("migration should add runtime metadata to an unchanged starter command")
	}
	if len(commands[0].Shells) == 0 || len(commands[0].Platforms) == 0 {
		t.Fatalf("starter command metadata was not added: %+v", commands[0])
	}
	if commands[0].Description != legacy.Description || !commands[0].Favorite || commands[0].Usage != legacy.Usage {
		t.Fatalf("migration changed user data: %+v", commands[0])
	}
}

func TestLoadAcceptsLegacyTOMLWithoutRuntimeMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	legacy := "[[commands]]\nname = \"My status\"\ncategory = \"Git\"\nsubcategory = \"Basics & Staging\"\ndescription = \"My saved command\"\ntags = [\"git\"]\ncommand = \"git status\"\n"
	if err := os.MkdirAll(filepath.Dir(Path()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(), []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	commands, err := Load()
	if err != nil {
		t.Fatalf("Load() rejected a legacy TOML entry: %v", err)
	}
	if len(commands) != 61 || commands[0].Name != "My status" || commands[0].Shells != nil || commands[0].Platforms != nil {
		t.Fatalf("legacy entry was not preserved with optional metadata: %+v (count %d)", commands[0], len(commands))
	}
}

func TestWorkingDirectoryIsOptionalAndSurvivesTOMLRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	commands := []Command{{Name: "Project status", Category: "Git", Description: "Inspect this project", Command: "git status", WorkingDirectory: `D:\Projects\demo`}}
	if err := Save(commands); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded[0].WorkingDirectory != commands[0].WorkingDirectory {
		t.Fatalf("working directory = %q, want %q", loaded[0].WorkingDirectory, commands[0].WorkingDirectory)
	}
}
