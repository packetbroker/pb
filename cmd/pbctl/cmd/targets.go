// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	routingpb "go.packetbroker.org/api/routing/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
)

var targetsCmd = &cobra.Command{
	Use:               "network-server-clusters",
	Aliases:           []string{"nsc"},
	Short:             "List Packet Broker Network Server clusters",
	SilenceUsage:      true,
	PersistentPreRunE: prerunConnect,
	PersistentPostRun: postrunConnect,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			client   = routingpb.NewRoutesClient(cpConn)
			offset   = uint32(0)
			clusters []*packetbroker.NetworkServerCluster
		)
		for {
			res, err := client.ListNetworkServerClusters(ctx, &routingpb.ListNetworkServerClustersRequest{
				Offset: offset,
			})
			if err != nil {
				return err
			}
			clusters = append(clusters, res.Clusters...)
			offset += uint32(len(res.Clusters))
			if len(res.Clusters) == 0 || offset >= res.Total {
				break
			}
		}
		fmt.Fprintln(tabout, "Authority\tCluster ID\tNet IDs\tTarget\t")
		for _, t := range clusters {
			fmt.Fprintf(tabout,
				"%s\t%s\t%s\t%s\t\n",
				t.GetAuthority(),
				t.GetClusterId(),
				(column.NetIDs)(t.GetNetIds()),
				(*column.Target)(t.Target),
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
}
