// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package gen provides commands to generate documentation and command trees.
package gen

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	cobradoc "github.com/spf13/cobra/doc"
)

const hugoDocFrontmatterTemplate = `---
title: "%s"
slug: %s
---
`

var hugoDocCmd = &cobra.Command{
	Use:   "hugodoc",
	Short: "Generate documentation for Hugo",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmd.VisitParents(func(c *cobra.Command) {
			c.DisableAutoGenTag = true
		})

		out, _ := cmd.Flags().GetString("out")

		prepender := func(filename string) string {
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			title := strings.ReplaceAll(base, "_", " ")
			fmt.Printf(`Write "%s" to %s`+"\n", title, filename)
			return fmt.Sprintf(hugoDocFrontmatterTemplate, title, base)
		}

		linkHandler := func(name string) string {
			base := strings.TrimSuffix(name, path.Ext(name))
			return fmt.Sprintf(`{{< relref "%s" >}}`, strings.ToLower(base))
		}

		return cobradoc.GenMarkdownTreeCustom(cmd.Root(), out, prepender, linkHandler)
	},
}

func init() {
	hugoDocCmd.Flags().String("out", ".", "output directory")
	Cmd.AddCommand(hugoDocCmd)
}
