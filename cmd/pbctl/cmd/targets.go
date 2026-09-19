// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

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
	RunE: func(_ *cobra.Command, _ []string) error {
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
				return fmt.Errorf("list network targets: %w", err)
			}
			targets = append(targets, res.GetTargets()...)
			offset += uint32(len(res.GetTargets()))
			if len(res.GetTargets()) == 0 || offset >= res.GetTotal() {
				break
			}
		}
		tabout.Println("NetID\tTenant ID\tTarget\t")
		for _, t := range targets {
			tabout.Printf("%s\t%s\t%s\t\n",
				packetbroker.NetID(t.GetNetId()),
				t.GetTenantId(),
				(*column.Target)(t.GetTarget()),
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
}
