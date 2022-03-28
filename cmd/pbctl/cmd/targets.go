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
	Use:               "targets",
	Short:             "List Packet Broker targets",
	SilenceUsage:      true,
	PersistentPreRunE: prerunConnect,
	PersistentPostRun: postrunConnect,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			client  = routingpb.NewRoutesClient(cpConn)
			offset  = uint32(0)
			targets []*packetbroker.NetworkTarget
		)
		for {
			res, err := client.ListNetworkTargets(ctx, &routingpb.ListNetworkTargetsRequest{
				Offset: offset,
			})
			if err != nil {
				return err
			}
			targets = append(targets, res.Targets...)
			offset += uint32(len(res.Targets))
			if len(res.Targets) == 0 || offset >= res.Total {
				break
			}
		}
		fmt.Fprintln(tabout, "NetID\tTenant ID\tTarget\t")
		for _, t := range targets {
			fmt.Fprintf(tabout,
				"%s\t%s\t%s\t\n",
				packetbroker.NetID(t.GetNetId()),
				t.GetTenantId(),
				(*column.Target)(t.Target),
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
}
