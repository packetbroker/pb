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
		Use:     "network",
		Aliases: []string{"networks", "nwk", "nwks", "n"},
		Short:   "Manage Packet Broker networks",
	}
	networkListCmd = &cobra.Command{
		Use:          "list",
		Aliases:      []string{"ls"},
		Short:        "List networks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				offset          = uint32(0)
				nameContains, _ = cmd.Flags().GetString("name-contains")
			)
			fmt.Fprintln(tabout, "NetID\tName\tDevAddr Blocks\tListed\tTarget\t")
			for {
				res, err := iampb.NewNetworkRegistryClient(conn).ListNetworks(ctx, &iampb.ListNetworksRequest{
					Offset:       offset,
					NameContains: nameContains,
				})
				if err != nil {
					return err
				}
				for _, t := range res.Networks {
					fmt.Fprintf(tabout, "%s\t%s\t%s\t%s\t%s\t\n",
						packetbroker.NetID(t.GetNetId()),
						t.GetName(),
						column.DevAddrBlocks(t.GetDevAddrBlocks()),
						column.YesNo(t.GetListed()),
						(*column.Target)(t.GetTarget()),
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
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2

  Configure a LoRaWAN Backend Interfaces 1.1.0 target with HTTP basic auth:
    $ pbadmin network create --net-id 000013 --target-protocol TS002_V1_1_0 \
      --target-address https://user:pass@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
			name, _ := cmd.Flags().GetString("name")
			devAddrBlocks := pbflag.GetDevAddrBlocks(cmd.Flags())
			listed, _ := cmd.Flags().GetBool("listed")
			target, err := target(cmd.Flags(), "target")
			if err != nil {
				return err
			}
			res, err := iampb.NewNetworkRegistryClient(conn).CreateNetwork(ctx, &iampb.CreateNetworkRequest{
				Network: &packetbroker.Network{
					NetId:         uint32(netID),
					Name:          name,
					DevAddrBlocks: devAddrBlocks,
					// TODO: Contact info (https://github.com/packetbroker/pb/issues/5)
					Listed: listed,
					Target: target,
				},
			})
			if err != nil {
				return err
			}
			return column.WriteNetwork(tabout, res.Network)
		},
	}
	networkGetCmd = &cobra.Command{
		Use:          "get",
		Short:        "Get a network",
		SilenceUsage: true,
		Example: `
  Get:
    $ pbadmin network get --net-id 000013`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
			res, err := iampb.NewNetworkRegistryClient(conn).GetNetwork(ctx, &iampb.NetworkRequest{
				NetId: uint32(netID),
			})
			if err != nil {
				return err
			}
			return column.WriteNetwork(tabout, res.Network)
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
      --dev-addr-blocks 26011000/20=eu1,26012000=eu2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
			req := &iampb.UpdateNetworkRequest{
				NetId: uint32(netID),
				// TODO: Contact info (https://github.com/packetbroker/pb/issues/5)
			}
			if cmd.Flags().Lookup("name").Changed {
				name, _ := cmd.Flags().GetString("name")
				req.Name = wrapperspb.String(name)
			}
			if cmd.Flags().Lookup("dev-addr-blocks").Changed {
				devAddrBlocks := pbflag.GetDevAddrBlocks(cmd.Flags())
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocks,
				}
			}
			if cmd.Flags().Lookup("listed").Changed {
				listed, _ := cmd.Flags().GetBool("listed")
				req.Listed = wrapperspb.Bool(listed)
			}
			_, err := iampb.NewNetworkRegistryClient(conn).UpdateNetwork(ctx, req)
			return err
		},
	}
	networkUpdateTargetCmd = &cobra.Command{
		Use:   "target",
		Short: "Update a network target",
		Example: `
  Configure a LoRaWAN Backend Interfaces 1.0 target with HTTP basic auth:
    $ pbadmin network update target --net-id 000013 --protocol TS002_V1_0 \
      --address https://user:pass@example.com

  Configure a LoRaWAN Backend Interfaces 1.0 target with TLS:
    $ pbadmin network update target --net-id 000013 --protocol TS002_V1_0 \
      --address https://example.com --root-cas-file ca.pem \
      --tls-cert-file key.pem --tls-key-file key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
			req := &iampb.UpdateNetworkRequest{
				NetId: uint32(netID),
			}
			target, err := target(cmd.Flags(), "")
			if err != nil {
				return err
			}
			req.Target = &iampb.TargetValue{
				Value: target,
			}
			_, err = iampb.NewNetworkRegistryClient(conn).UpdateNetwork(ctx, req)
			return err
		},
	}
	networkDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete a network",
		SilenceUsage: true,
		Example: `
  Delete:
    $ pbadmin network delete --net-id 000013`,
		RunE: func(cmd *cobra.Command, args []string) error {
			netID := pbflag.GetNetID(cmd.Flags(), "")
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
		SilenceUsage: true,
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

  Initialize configuration for network with rights to read networks:
    $ pbadmin network init --net-id 000013 --rights READ_NETWORK \
        --router-address eu.packetbroker.io

Rights:
  READ_NETWORK          Read networks
  READ_NETWORK_CONTACT  Read network contact information

Router addresses:
  apac.packetbroker.io  Asia Pacific
  eu.packetbroker.io    Europe, Middle East and Africa
  nam.packetbroker.io   Americas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := pbflag.GetEndpoint(cmd.Flags(), "")
			req := &iampbv2.CreateNetworkAPIKeyRequest{
				NetId:     uint32(endpoint.NetID),
				TenantId:  endpoint.TenantID.ID,
				ClusterId: endpoint.ClusterID,
				Rights:    pbflag.GetAPIKeyRights(cmd.Flags()),
			}
			res, err := iampbv2.NewNetworkAPIKeyVaultClient(conn).CreateAPIKey(ctx, req)
			if err != nil {
				return err
			}
			routerAddress, _ := cmd.Flags().GetString("router-address")
			viper.Set("router-address", routerAddress)
			viper.Set("client-id", res.Key.GetKeyId())
			viper.Set("client-secret", res.Key.GetKey())
			if err := viper.WriteConfigAs(".pb.yaml"); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "Saved configuration to .pb.yaml")
			return column.WriteKV(tabout,
				"NetID", endpoint.NetID,
				"Tenant ID", endpoint.ID,
				"Cluster ID", endpoint.ClusterID,
				"Router Address", routerAddress,
				"API Key ID", res.Key.GetKeyId(),
				"API Secret Key", res.Key.GetKey(),
				"API Key Rights", column.Rights(res.Key.GetRights()),
				"API Key State", res.Key.GetState(),
			)
		},
	}
)

func networkSettingsFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String("name", "", "network name")
	flags.Bool("listed", false, "list network in catalog")
	flags.AddFlagSet(pbflag.DevAddrBlocks())
	return flags
}

func init() {
	rootCmd.AddCommand(networkCmd)

	networkListCmd.Flags().String("name-contains", "", "filter networks by name")
	networkCmd.AddCommand(networkListCmd)

	networkCreateCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkCreateCmd.Flags().AddFlagSet(networkSettingsFlags())
	networkCreateCmd.Flags().AddFlagSet(targetFlags("target"))
	networkCmd.AddCommand(networkCreateCmd)

	networkGetCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkCmd.AddCommand(networkGetCmd)

	networkUpdateCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkUpdateCmd.Flags().AddFlagSet(networkSettingsFlags())
	networkUpdateTargetCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkUpdateTargetCmd.Flags().AddFlagSet(targetFlags(""))
	networkUpdateCmd.AddCommand(networkUpdateTargetCmd)
	networkCmd.AddCommand(networkUpdateCmd)

	networkDeleteCmd.Flags().AddFlagSet(pbflag.NetID(""))
	networkCmd.AddCommand(networkDeleteCmd)

	networkInitCmd.Flags().AddFlagSet(pbflag.Endpoint(""))
	networkInitCmd.Flags().AddFlagSet(pbflag.APIKeyRights())
	networkInitCmd.Flags().String("router-address", "", "Packet Broker router address")
	networkCmd.AddCommand(networkInitCmd)
}
