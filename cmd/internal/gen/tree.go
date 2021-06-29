// Copyright © 2021 The Things Industries B.V.

package gen

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

type command struct {
	Short       string             `json:"short,omitempty"`
	Path        string             `json:"path,omitempty"`
	SubCommands map[string]command `json:"subCommands,omitempty"`
}

func commandTree(cmd *cobra.Command) (res command) {
	res.Path = cmd.CommandPath()
	res.Short = cmd.Short
	if len(cmd.Commands()) == 0 {
		return
	}
	res.SubCommands = make(map[string]command, len(cmd.Commands()))
	for _, cmd := range cmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		res.SubCommands[cmd.Name()] = commandTree(cmd)
	}
	return
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Generate command tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("out")

		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]command{
			cmd.Root().Name(): commandTree(cmd.Root()),
		})
	},
}

func init() {
	treeCmd.Flags().String("out", "tree.json", "output file")
	Cmd.AddCommand(treeCmd)
}
