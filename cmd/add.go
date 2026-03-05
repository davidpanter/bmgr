package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new keybinding",
	RunE:  runAdd,
}

var addJSON string

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addJSON, "json", "", "Add binding from JSON (skips interactive prompts)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if addJSON != "" {
		var b store.Binding
		if err := json.Unmarshal([]byte(addJSON), &b); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		if b.App == "" {
			return fmt.Errorf("JSON must include 'app'")
		}
		if b.Description == "" {
			return fmt.Errorf("JSON must include 'description'")
		}
		if b.ID == "" {
			id, err := store.GenerateID()
			if err != nil {
				return err
			}
			b.ID = id
		}
		if b.Keys == nil {
			b.Keys = []string{}
		}
		if b.Alternates == nil {
			b.Alternates = []string{}
		}
		if b.Tags == nil {
			b.Tags = []string{}
		}
		s, err := store.Load()
		if err != nil {
			return err
		}
		if err := s.Add(b); err != nil {
			return err
		}
		if err := s.Save(); err != nil {
			return err
		}
		fmt.Printf("Added binding %s: %s — %s\n", b.ID, b.App, b.Description)
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	app := promptRequired(scanner, "App")
	if app == "" {
		return fmt.Errorf("aborted")
	}

	desc := promptRequired(scanner, "Description")
	if desc == "" {
		return fmt.Errorf("aborted")
	}

	keysRaw := prompt(scanner, "Keys (comma-separated, optional)")
	altsRaw := prompt(scanner, "Alternates (comma-separated, optional)")
	tagsRaw := prompt(scanner, "Tags (comma-separated, optional)")
	notes := prompt(scanner, "Notes (optional)")

	id, err := store.GenerateID()
	if err != nil {
		return err
	}

	b := store.Binding{
		ID:          id,
		App:         sanitizeTab(app),
		Description: sanitizeTab(desc),
		Keys:        splitAndTrim(keysRaw),
		Alternates:  splitAndTrim(altsRaw),
		Tags:        splitAndTrim(tagsRaw),
		Notes:       sanitizeTab(notes),
	}

	s, err := store.Load()
	if err != nil {
		return err
	}

	if err := s.Add(b); err != nil {
		return err
	}

	if err := s.Save(); err != nil {
		return err
	}

	fmt.Printf("Added binding %s: %s — %s\n", b.ID, b.App, b.Description)
	return nil
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("%s: ", label)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func promptRequired(scanner *bufio.Scanner, label string) string {
	for {
		fmt.Printf("%s: ", label)
		if !scanner.Scan() {
			return ""
		}
		val := strings.TrimSpace(scanner.Text())
		if val != "" {
			return val
		}
		fmt.Printf("%s is required.\n", label)
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
