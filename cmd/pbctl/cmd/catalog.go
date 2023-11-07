// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	iampb "go.packetbroker.org/api/iam/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
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
		Use:               "catalog",
		Aliases:           []string{"cat"},
		Short:             "Packet Broker catalog",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	catalogNetworksCmd = &cobra.Command{
		Use:     "networks",
		Aliases: []string{"network", "ns"},
		Short:   "Show Forwarders and Home Networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				tenantID, _       = pbflag.GetTenantID(cmd.Flags(), "")
				offset            = uint32(0)
				idContains, _     = cmd.Flags().GetString("id-contains")
				nameContains, _   = cmd.Flags().GetString("name-contains")
				policyTenantID, _ = pbflag.GetTenantID(cmd.Flags(), "policy")
			)
			var policyRef *iampb.ListNetworksRequest_PolicyReference
			if !policyTenantID.IsEmpty() {
				policyRef = &iampb.ListNetworksRequest_PolicyReference{
					NetId:    uint32(policyTenantID.NetID),
					TenantId: policyTenantID.ID,
				}
			}
			fmt.Fprintln(tabout, "NetID\tTenant ID\tName\tDevAddr Blocks\t")
			for {
				res, err := iampb.NewCatalogClient(iamConn).ListNetworks(ctx, &iampb.ListNetworksRequest{
					NetId:            uint32(tenantID.NetID),
					TenantId:         tenantID.ID,
					Offset:           offset,
					TenantIdContains: idContains,
					NameContains:     nameContains,
					PolicyReference:  policyRef,
				})
				if err != nil {
					return err
				}
				for _, hn := range res.Networks {
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
				offset += uint32(len(res.Networks))
				if len(res.Networks) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
	catalogHomeNetworksCmd = &cobra.Command{
		Use:     "home-networks",
		Aliases: []string{"home-network", "hns"},
		Short:   "Show Home Networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				tenantID, _       = pbflag.GetTenantID(cmd.Flags(), "")
				offset            = uint32(0)
				idContains, _     = cmd.Flags().GetString("id-contains")
				nameContains, _   = cmd.Flags().GetString("name-contains")
				policyTenantID, _ = pbflag.GetTenantID(cmd.Flags(), "policy")
			)
			var policyRef *iampb.ListNetworksRequest_PolicyReference
			if !policyTenantID.IsEmpty() {
				policyRef = &iampb.ListNetworksRequest_PolicyReference{
					NetId:    uint32(policyTenantID.NetID),
					TenantId: policyTenantID.ID,
				}
			}
			fmt.Fprintln(tabout, "NetID\tTenant ID\tName\tDevAddr Blocks\t")
			for {
				res, err := iampb.NewCatalogClient(iamConn).ListHomeNetworks(ctx, &iampb.ListNetworksRequest{
					NetId:            uint32(tenantID.NetID),
					TenantId:         tenantID.ID,
					Offset:           offset,
					TenantIdContains: idContains,
					NameContains:     nameContains,
					PolicyReference:  policyRef,
				})
				if err != nil {
					return err
				}
				for _, hn := range res.Networks {
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
				offset += uint32(len(res.Networks))
				if len(res.Networks) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
	catalogJoinServersCmd = &cobra.Command{
		Use:     "join-servers",
		Aliases: []string{"join-server", "js"},
		Short:   "Show Join Servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				offset          = uint32(0)
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			fmt.Fprintln(tabout, "  ID\tName\tJoinEUI Prefixes\t")
			for {
				res, err := iampb.NewCatalogClient(iamConn).ListJoinServers(ctx, &iampb.ListJoinServersRequest{
					Offset:       offset,
					NameContains: nameContains,
				})
				if err != nil {
					return err
				}
				for _, js := range res.JoinServers {
					fmt.Fprintf(tabout, "%4d\t%s\t%s\t\n",
						js.GetId(),
						js.GetName(),
						column.JoinEUIPrefixes(js.GetJoinEuiPrefixes()),
					)
				}
				offset += uint32(len(res.JoinServers))
				if len(res.JoinServers) == 0 || offset >= res.Total {
					break
				}
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(catalogCmd)

	catalogNetworksCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	catalogNetworksCmd.Flags().String("id-contains", "", "filter tenants by ID")
	catalogNetworksCmd.Flags().String("name-contains", "", "filter networks or tenants by name")
	catalogNetworksCmd.Flags().AddFlagSet(pbflag.TenantID("policy"))
	catalogCmd.AddCommand(catalogNetworksCmd)

	catalogHomeNetworksCmd.Flags().AddFlagSet(pbflag.TenantID(""))
	catalogHomeNetworksCmd.Flags().String("id-contains", "", "filter tenants by ID")
	catalogHomeNetworksCmd.Flags().String("name-contains", "", "filter networks or tenants by name")
	catalogHomeNetworksCmd.Flags().AddFlagSet(pbflag.TenantID("policy"))
	catalogCmd.AddCommand(catalogHomeNetworksCmd)

	catalogJoinServersCmd.Flags().String("name-contains", "", "filter Join Servers by name")
	catalogCmd.AddCommand(catalogJoinServersCmd)
}
