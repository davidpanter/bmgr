package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a keybinding from JSON",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var updateJSON string
var updateFile string

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateJSON, "json", "", "Replacement binding as JSON")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Target file in config directory (e.g. tmux.json)")
	updateCmd.MarkFlagRequired("json")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	var b store.Binding
	if err := json.Unmarshal([]byte(updateJSON), &b); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if b.App == "" {
		return fmt.Errorf("JSON must include 'app'")
	}
	if b.Description == "" {
		return fmt.Errorf("JSON must include 'description'")
	}
	b.ID = id

	s, err := store.Load()
	if err != nil {
		return err
	}

	if updateFile != "" {
		b.SetSourceFile(updateFile)
	} else if existing := s.FindByID(id); existing != nil {
		b.SetSourceFile(existing.SourceFile())
	}

	if err := s.Replace(b); err != nil {
		return err
	}

	if err := s.Save(); err != nil {
		return err
	}

	fmt.Printf("Updated binding %s: %s — %s\n", b.ID, b.App, b.Description)
	return nil
}
