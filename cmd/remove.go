package cmd

import (
	"fmt"
	"strings"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a keybinding",
	RunE:    runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	s, err := store.Load()
	if err != nil {
		return err
	}

	if len(s.Bindings) == 0 {
		fmt.Println("No bindings to remove.")
		return nil
	}

	lines := make([]string, len(s.Bindings))
	for i, b := range s.Bindings {
		lines[i] = formatLine(b)
	}

	selected, err := runFzf(lines, []string{
		"--header=Select binding to remove",
	})
	if err != nil {
		return err
	}
	if selected == "" {
		return nil
	}

	id := strings.Split(selected, "\t")[0]

	if err := s.Remove(id); err != nil {
		return err
	}

	if err := s.Save(); err != nil {
		return err
	}

	fmt.Printf("Removed binding %s\n", id)
	return nil
}
