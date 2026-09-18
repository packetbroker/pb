// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"encoding/json"
	"fmt"
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
		return res
	}
	res.SubCommands = make(map[string]command, len(cmd.Commands()))
	for _, cmd := range cmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		res.SubCommands[cmd.Name()] = commandTree(cmd)
	}
	return res
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Generate command tree",
	RunE: func(cmd *cobra.Command, _ []string) error {
		out, _ := cmd.Flags().GetString("out")

		file, err := os.Create(out) //nolint:gosec // the output file path is provided by the user of this CLI
		if err != nil {
			return fmt.Errorf("create output file %q: %w", out, err)
		}

		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]command{
			cmd.Root().Name(): commandTree(cmd.Root()),
		}); err != nil {
			_ = file.Close()
			return fmt.Errorf("write command tree: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close output file %q: %w", out, err)
		}
		return nil
	},
}

func init() {
	treeCmd.Flags().String("out", "tree.json", "output file")
	Cmd.AddCommand(treeCmd)
}
