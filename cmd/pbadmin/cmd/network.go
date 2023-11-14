// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	iampb "go.packetbroker.org/api/iam"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	"go.packetbroker.org/pb/cmd/internal/pbflag"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	networkCmd = &cobra.Command{
		Use:               "network",
		Aliases:           []string{"networks", "nwk", "nwks", "n"},
		Short:             "Manage Packet Broker networks",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	networkListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				offset          = uint32(0)
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			fmt.Fprintln(tabout, "NetID\tAuthority\tName\tDevAddr Blocks\tListed\tTarget\tDelegated NetID\t")
			for {
				res, err := iampb.NewNetworkRegistryClient(conn).ListNetworks(ctx, &iampb.ListNetworksRequest{
					Offset:       offset,
					NameContains: nameContains,
				})
				if err != nil {
					return err
				}
				for _, t := range res.Networks {
					var delegatedNetID *uint32
					if val := t.GetDelegatedNetId(); val != nil {
						delegatedNetID = &val.Value
					}
					fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(t.GetNetId()),
						t.Authority,
						t.GetName(),
						column.DevAddrBlocks(t.GetDevAddrBlocks()),
						column.YesNo(t.GetListed()),
						(*packetbroker.NetID)(delegatedNetID),
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
	networkCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a network",
		Example: `
  Create:
    $ pbadmin network create --net-id 000013

  Create with name and listed in the catalog:
    $ pbadmin network create --net-id 000013 --name "The Things Network" \
      --listed

  Define DevAddr blocks to named clusters:
    $ pbadmin network create --net-id 000013 \
      --dev-addr-blocks 26011000/20=eu1.cloud.thethings.industries,26012000=eu2.cloud.thethings.industries \
      --cluster-ns-ids eu1.cloud.thethings.industries=EC656E0000000001,eu2.cloud.thethings.industries=EC656E0000000002`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID, _ := pbflag.GetNetID(cmd.Flags(), "")
			name, _ := cmd.Flags().GetString("name")
			devAddrBlocks, _, _ := pbflag.GetDevAddrBlocks(cmd.Flags())
			clusterNSIDs := pbflag.GetClusterNSIDs(cmd.Flags())
			adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin")
			techContact := pbflag.GetContactInfo(cmd.Flags(), "tech")
			listed, _ := cmd.Flags().GetBool("listed")
			var delegatedNetID *wrapperspb.UInt32Value
			if netID, ok := pbflag.GetNetID(cmd.Flags(), "delegated"); ok {
				delegatedNetID = wrapperspb.UInt32(uint32(netID))
			}
			res, err := iampb.NewNetworkRegistryClient(conn).CreateNetwork(ctx, &iampb.CreateNetworkRequest{
				Network: &packetbroker.Network{
					NetId:                 uint32(netID),
					Name:                  name,
					DevAddrBlocks:         devAddrBlocks,
					NsIds:                 clusterNSIDs,
					AdministrativeContact: adminContact,
					TechnicalContact:      techContact,
					Listed:                listed,
					DelegatedNetId:        delegatedNetID,
				},
			})
			if err != nil {
				return err
			}
			return column.WriteNetwork(tabout, res.Network, false)
		},
	}
	networkGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a network",
		Example: `
  Get:
    $ pbadmin network get --net-id 000013`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID, _ := pbflag.GetNetID(cmd.Flags(), "")
			res, err := iampb.NewNetworkRegistryClient(conn).GetNetwork(ctx, &iampb.NetworkRequest{
				NetId: uint32(netID),
			})
			if err != nil {
				return err
			}
			verbose, _ := cmd.Flags().GetBool("verbose")
			return column.WriteNetwork(tabout, res.Network, verbose)
		},
	}
	networkUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update a network",
		Example: `
  Update name:
    $ pbadmin network update --net-id 000013 --name "The Things Network"

  Define DevAddr blocks to named clusters:
    $ pbadmin network update --net-id 000013 \
      --dev-addr-blocks 26011000/20=eu1.cloud.thethings.industries,26012000=eu2.cloud.thethings.industries \
      --cluster-ns-ids eu1.cloud.thethings.industries=EC656E0000000001,eu2.cloud.thethings.industries=EC656E0000000002`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID, _ := pbflag.GetNetID(cmd.Flags(), "")
			client := iampb.NewNetworkRegistryClient(conn)
			nwk, err := client.GetNetwork(ctx, &iampb.NetworkRequest{
				NetId: uint32(netID),
			})
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("listed") {
				listed, _ := cmd.Flags().GetBool("listed")
				_, err := client.UpdateNetworkListed(ctx, &iampb.UpdateNetworkListedRequest{
					NetId:  uint32(netID),
					Listed: listed,
				})
				if err != nil {
					return err
				}
			}
			var any bool
			req := &iampb.UpdateNetworkRequest{
				NetId: uint32(netID),
			}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
				any = true
			}
			devAddrBlocksAll, devAddrBlocksAllAdd, devAddrBlocksAllRemove := pbflag.GetDevAddrBlocks(cmd.Flags())
			if cmd.Flags().Changed("dev-addr-blocks") {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocksAll,
				}
				any = true
			} else if len(devAddrBlocksAllAdd) > 0 || len(devAddrBlocksAllRemove) > 0 {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: mergeDevAddrBlocks(nwk.Network.DevAddrBlocks, devAddrBlocksAllAdd, devAddrBlocksAllRemove),
				}
				any = true
			}
			if pbflag.ClusterNSIDsChanged(cmd.Flags()) {
				clusterNSIDs := pbflag.GetClusterNSIDs(cmd.Flags())
				req.NsIds = &iampb.NSIDsValue{
					Value: clusterNSIDs,
				}
				any = true
			}
			if adminContact := pbflag.GetContactInfo(cmd.Flags(), "admin"); adminContact != nil {
				req.AdministrativeContact = &packetbroker.ContactInfoValue{
					Value: adminContact,
				}
				any = true
			}
			if techContact := pbflag.GetContactInfo(cmd.Flags(), "tech"); techContact != nil {
				req.TechnicalContact = &packetbroker.ContactInfoValue{
					Value: techContact,
				}
				any = true
			}
			if delegatedNetID, ok := pbflag.GetNetID(cmd.Flags(), "delegated"); ok {
				req.DelegatedNetId = &iampb.UpdateNetworkRequest_DelegatedNetID{
					Value: wrapperspb.UInt32(uint32(delegatedNetID)),
				}
				any = true
			} else if unset, _ := cmd.Flags().GetBool("unset-delegated-net-id"); cmd.Flags().Changed("unset-delegated-net-id") && unset {
				req.DelegatedNetId = new(iampb.UpdateNetworkRequest_DelegatedNetID)
				any = true
			}
			if any {
				_, err = client.UpdateNetwork(ctx, req)
				return err
			}
			return nil
		},
	}
	networkDeleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete a network",
		Example: `
  Delete:
    $ pbadmin network delete --net-id 000013`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID, _ := pbflag.GetNetID(cmd.Flags(), "")
			_, err := iampb.NewNetworkRegistryClient(conn).DeleteNetwork(ctx, &iampb.NetworkRequest{
				NetId: uint32(netID),
			})
			return err
		},
	}
	networkInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize network configuration",
		Long: `Initialize network configuration

This stores the router address and a newly requested network API key in a local
configuration file (.pb.yaml). This configuration can be used by Packet Broker
command-line interfaces.`,
		Example: `
  Initialize configuration for a network:
    $ pbadmin network init --net-id 000013 \
        --router-address eu.packetbroker.io

  Initialize configuration for a tenant:
    $ pbadmin network init --net-id 000013 --tenant-id ttn \
        --router-address eu.packetbroker.io

  Initialize configuration for a named cluster in a tenant:
    $ pbadmin network init --net-id 000013 --tenant-id ttn --cluster-id eu1 \
        --router-address eu.packetbroker.io

Router addresses:
  apac.packetbroker.io  Asia Pacific
  eu.packetbroker.io    Europe, Middle East and Africa
  nam.packetbroker.io   Americas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, _ := pbflag.GetEndpoint(cmd.Flags(), "")
			req := &iampbv2.CreateNetworkAPIKeyRequest{
				NetId:     uint32(endpoint.NetID),
				TenantId:  endpoint.TenantID.ID,
				ClusterId: endpoint.ClusterID,
			}
			res, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).CreateAPIKey(ctx, req)
			if err != nil {
				return err
			}
			iamAddress, _ := cmd.Flags().GetString("iam-address")
			controlPlaneAddress, _ := cmd.Flags().GetString("controlplane-address")
			reportsAddress, _ := cmd.Flags().GetString("reports-address")
			routerAddress, _ := cmd.Flags().GetString("router-address")
			viper.Set("controlplane-address", controlPlaneAddress)
			viper.Set("reports-address", reportsAddress)
			viper.Set("router-address", routerAddress)
			viper.Set("client-id", res.Key.GetKeyId())
			viper.Set("client-secret", res.Key.GetKey())
			if err := viper.WriteConfigAs(".pb.yaml"); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "Saved configuration to .pb.yaml")
			return column.WriteKV(tabout,
				"NetID", endpoint.NetID.String(),
				"Tenant ID", endpoint.ID,
				"Cluster ID", endpoint.ClusterID,
				"IAM Address", iamAddress,
				"Control Plane Address", controlPlaneAddress,
				"Router Address", routerAddress,
				"API Key ID", res.Key.GetKeyId(),
				"API Secret Key", res.Key.GetKey(),
				"API Key State", res.Key.GetState().String(),
			)
		},
	}
)

func networkSettingsFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("name", "", "network name")
	flags.Bool("listed", false, "list network in catalog")
	flags.AddFlagSet(pbflag.NetID("delegated"))
	return flags
}

func init() {
	rootCmd.AddCommand(networkCmd)

	networkListCmd.Flags().String("name-contains", "", "filter networks by name")
	networkCmd.AddCommand(networkListCmd)

	networkCreateCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkCreateCmd.Flags().AddFlagSet(networkSettingsFlags())
	networkCreateCmd.Flags().AddFlagSet(pbflag.DevAddrBlocks(false))
	networkCreateCmd.Flags().AddFlagSet(pbflag.ClusterNSIDs())
	networkCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	networkCreateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	networkCmd.AddCommand(networkCreateCmd)

	networkGetCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkGetCmd.Flags().Bool("verbose", false, "verbose output")
	networkCmd.AddCommand(networkGetCmd)

	networkUpdateCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkUpdateCmd.Flags().AddFlagSet(networkSettingsFlags())
	networkUpdateCmd.Flags().AddFlagSet(pbflag.DevAddrBlocks(true))
	networkUpdateCmd.Flags().AddFlagSet(pbflag.ClusterNSIDs())
	networkUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("admin"))
	networkUpdateCmd.Flags().AddFlagSet(pbflag.ContactInfo("tech"))
	networkUpdateCmd.Flags().Bool("unset-delegated-net-id", false, "unset the delegated NetID")
	networkCmd.AddCommand(networkUpdateCmd)

	networkDeleteCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkCmd.AddCommand(networkDeleteCmd)

	networkInitCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	networkInitCmd.Flags().String("controlplane-address", "cp.packetbroker.net:443", `Packet Broker Control Plane address "host[:port]"`)
	networkInitCmd.Flags().String("reports-address", "reports.packetbroker.net:443", `Packet Broker Reporter address "host[:port]"`)
	networkInitCmd.Flags().String("router-address", "", `Packet Broker Router address "host[:port]"`)
	networkCmd.AddCommand(networkInitCmd)
}
