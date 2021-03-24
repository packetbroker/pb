// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	iampb "go.packetbroker.org/api/iam/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
)

type network interface {
	GetNetId() uint32
	GetName() string
	GetDevAddrBlocks() []*packetbroker.DevAddrBlock
}

type homeNetwork struct {
	network
	tenantID string
}

var (
	catalogCmd = &cobra.Command{
		Use:     "catalog",
		Aliases: []string{"cat"},
		Short:   "Packet Broker catalog",
	}
	catalogHomeNetworksCmd = &cobra.Command{
		Use:          "home-networks",
		Aliases:      []string{"home-network", "hns"},
		Short:        "Show listed Home Networks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			offset := uint32(0)
			fmt.Fprintln(tabout, "NetID\tTenant ID\tName\tDevAddr Blocks\t")
			for {
				res, err := iampb.NewCatalogClient(iamConn).ListHomeNetworks(ctx, &iampb.ListHomeNetworksRequest{
					Offset: offset,
				})
				if err != nil {
					return err
				}
				for _, hn := range res.HomeNetworks {
					var row homeNetwork
					if nwk := hn.GetNetwork(); nwk != nil {
						row.network = nwk
						row.tenantID = "-"
					} else if tnt := hn.GetTenant(); tnt != nil {
						row.network = tnt
						row.tenantID = tnt.GetTenantId()
					}
					fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(row.GetNetId()),
						row.tenantID,
						row.GetName(),
						column.DevAddrBlocks(row.GetDevAddrBlocks()),
					)
				}
				offset += uint32(len(res.HomeNetworks))
				if len(res.HomeNetworks) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(catalogCmd)

	catalogCmd.AddCommand(catalogHomeNetworksCmd)
}
