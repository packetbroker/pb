// Copyright © 2021 The Things Industries B.V.

package gen

import "github.com/spf13/cobra"

// Cmd contains sub-commands to generate things.
var Cmd = &cobra.Command{
	Use:   "gen",
	Short: "Generation commands",
}
