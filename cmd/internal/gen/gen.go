// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package gen

import "github.com/spf13/cobra"

// Cmd contains sub-commands to generate things.
var Cmd = &cobra.Command{
	Use:   "gen",
	Short: "Generation commands",
}
