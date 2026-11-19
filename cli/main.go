package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/snipbox/cli/internal/storage"
	"github.com/snipbox/cli/internal/sync"
	"github.com/snipbox/cli/internal/ui"
)

func main() {
	var list, path, push bool
	root := &cobra.Command{Use: "snipbox [query]", Short: "Search and run your command snippets", Args: cobra.ArbitraryArgs, RunE: func(cmd *cobra.Command, args []string) error {
		if path {
			fmt.Println(storage.Path())
			return nil
		}
		commands, err := storage.Load()
		if err != nil {
			return err
		}
		if push {
			return syncer.Push()
		}
		if list {
			for _, c := range commands {
				fmt.Printf("[%s] %s — %s\n", c.Category, c.Name, c.Command)
			}
			return nil
		}
		_, err = tea.NewProgram(ui.New(commands, strings.Join(args, " ")), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
		return err
	}}
	root.Flags().BoolVar(&list, "list", false, "List commands")
	root.Flags().BoolVar(&path, "path", false, "Print the command database path")
	root.Flags().BoolVar(&push, "sync", false, "Commit and push the local Snipbox database")
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
