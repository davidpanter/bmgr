package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importKeybCmd = &cobra.Command{
	Use:   "import-keyb [file]",
	Short: "Import bindings from a keyb YAML file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runImportKeyb,
}

func init() {
	rootCmd.AddCommand(importKeybCmd)
}

type keybGroup struct {
	Name     string       `yaml:"name"`
	Prefix   string       `yaml:"prefix"`
	Keybinds []keybEntry  `yaml:"keybinds"`
}

type keybEntry struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

func runImportKeyb(cmd *cobra.Command, args []string) error {
	path := "~/.config/keyb/keyb.yml"
	if len(args) > 0 {
		path = args[0]
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = home + path[1:]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", path, err)
	}

	var groups []keybGroup
	if err := yaml.Unmarshal(data, &groups); err != nil {
		return fmt.Errorf("could not parse YAML: %w", err)
	}

	s, err := store.Load()
	if err != nil {
		return err
	}

	imported := 0
	for _, group := range groups {
		// Parse group name: "App: Category" → app="App", tag="Category"
		app, tag := parseGroupName(group.Name)

		for _, entry := range group.Keybinds {
			desc := parseEntryDescription(entry.Name)

			id, err := store.GenerateID()
			if err != nil {
				return err
			}

			tags := []string{}
			if tag != "" {
				tags = []string{strings.ToLower(tag)}
			}

			notes := ""
			if group.Prefix != "" {
				notes = "prefix: " + group.Prefix
			}

			b := store.Binding{
				ID:          id,
				App:         sanitizeTab(app),
				Description: sanitizeTab(desc),
				Keys:        []string{sanitizeTab(entry.Key)},
				Alternates:  []string{},
				Tags:        tags,
				Notes:       notes,
			}

			if err := s.Add(b); err != nil {
				return err
			}
			imported++
		}
	}

	if err := s.Save(); err != nil {
		return err
	}

	fmt.Printf("Imported %d bindings from %s\n", imported, path)
	return nil
}

// parseGroupName splits "App: Category" into ("App", "Category").
// If no colon, returns (full name, "").
func parseGroupName(name string) (app, category string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(name), ""
}

// parseEntryDescription strips the "app: " prefix from a keyb entry name.
// e.g. "cosmic: Open launcher" → "Open launcher"
// If no colon, returns as-is.
func parseEntryDescription(name string) string {
	parts := strings.SplitN(name, ": ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(name)
}
