// Copyright © 2021 The Things Industries B.V.

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
	Short: "Generate documentation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.VisitParents(func(c *cobra.Command) {
			c.DisableAutoGenTag = true
		})

		out, _ := cmd.Flags().GetString("out")
		if out == "" {
			out = "."
		}

		prepender := func(filename string) string {
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			return fmt.Sprintf(hugoDocFrontmatterTemplate, strings.Replace(base, "_", " ", -1), base)
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
