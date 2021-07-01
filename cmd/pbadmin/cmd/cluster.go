// Copyright © 2021 The Things Industries B.V.

package cmd

import "github.com/spf13/cobra"

var clusterCmd = &cobra.Command{
	Use:               "cluster",
	Aliases:           []string{"cluster", "c"},
	Short:             "Manage Packet Broker clusters",
	PersistentPreRunE: prerunConnect,
	PersistentPostRun: postrunConnect,
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
