package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a keybinding in $EDITOR",
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	s, err := store.Load()
	if err != nil {
		return err
	}

	if len(s.Bindings) == 0 {
		fmt.Println("No bindings to edit.")
		return nil
	}

	lines := make([]string, len(s.Bindings))
	for i, b := range s.Bindings {
		lines[i] = formatLine(b)
	}

	selected, err := runFzf(lines, []string{
		"--header=Select binding to edit",
	})
	if err != nil {
		return err
	}
	if selected == "" {
		return nil
	}

	id := strings.Split(selected, "\t")[0]
	binding := s.FindByID(id)
	if binding == nil {
		return fmt.Errorf("binding %q not found", id)
	}

	data, err := json.MarshalIndent(binding, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal binding: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "bmgr-edit-*.json")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("could not write temp file: %w", err)
	}
	tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, tmpFile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("could not read edited file: %w", err)
	}

	var updated store.Binding
	if err := json.Unmarshal(edited, &updated); err != nil {
		return fmt.Errorf("could not parse edited JSON: %w", err)
	}

	if updated.ID != id {
		return fmt.Errorf("ID must not be changed (was %q, got %q)", id, updated.ID)
	}
	if updated.App == "" {
		return fmt.Errorf("App must not be empty")
	}
	if updated.Description == "" {
		return fmt.Errorf("Description must not be empty")
	}

	if err := s.Replace(updated); err != nil {
		return err
	}

	if err := s.Save(); err != nil {
		return err
	}

	fmt.Printf("Updated binding %s: %s — %s\n", updated.ID, updated.App, updated.Description)
	return nil
}
