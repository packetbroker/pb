// Copyright © 2021 The Things Industries B.V.

package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	mappingpb "go.packetbroker.org/api/mapping/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
)

var (
	gatewayVisibilityCmd = &cobra.Command{
		Use:               "gateway-visibility",
		Aliases:           []string{"gwvis"},
		Short:             "Manage Packet Broker gateway visibilities",
		PersistentPreRunE: prerunConnect,
		PersistentPostRun: postrunConnect,
	}
	gatewayVisibilitySetCmd = &cobra.Command{
		Use:   "set",
		Short: "Set gateway visibility",
		Long: `Set default gateway visibility of Forwarder, or a specific visibility
between a Forwarder and a Home Network.

You can specify visibilities with the following symbols:
  Lo: Location (GPS coordinates and altitude)
  Ap: Antenna placement (indoor or outdoor)
  Ac: Antenna count
  Ft: Whether the gateway produces fine timestamps for geolocation
  Ci: Contact information
  St: Online/offline status
  Fp: Frequency plan
  Pr: Packet rates (receive and transmit)

Visibilities are defined by the Forwarding network or tenant, as they control
who may see their infrastructure.`,
		Example: `
  Set default gateway visibility of Forwarder network to see everything:
    $ pbctl gateway-visibility set --forwarder-net-id 000013 \
      --defaults --set LoApAcFtCiStFpPr

  Set default gateway visibility of Forwarder tenant to see everything:
    $ pbctl gateway-visibility set --forwarder-net-id 000013 \
      --forwarder-tenant-id tti --defaults --set LoApAcFtCiStFpPr

  Configure only location and frequency plan between The Things Network
  (NetID 000013) and Senet (NetID 000009):
    $ pbctl gateway-visibility set --forwarder-net-id 000013 \
      --home-network-net-id 000009 --set LoFp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := mappingpb.NewGatewayVisibilityManagerClient(cpConn)
			forwarderTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			visibility := pbflag.GetGatewayVisibility(cmd.Flags())
			visibility.ForwarderNetId = uint32(forwarderTenantID.NetID)
			visibility.ForwarderTenantId = forwarderTenantID.ID
			defaults, _ := cmd.Flags().GetBool("defaults")
			var err error
			if defaults {
				_, err = client.SetDefaultVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
					Visibility: visibility,
				})
			} else {
				homeNetworkTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "home-network")
				visibility.HomeNetworkNetId = uint32(homeNetworkTenantID.NetID)
				visibility.HomeNetworkTenantId = homeNetworkTenantID.ID
				_, err = client.SetHomeNetworkVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
					Visibility: visibility,
				})
			}
			if err != nil {
				return err
			}
			return column.WriteVisibilities(tabout, defaults, visibility)
		},
	}
	gatewayVisibilityGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get gateway visibility",
		Example: `
  Get default gateway visibility of Forwarder network:
    $ pbctl gateway-visibility get --forwarder-net-id 000013 --defaults

  Get default gateway visibility of Forwarder tenant:
    $ pbctl gateway-visibility get --forwarder-net-id 000013 \
      --forwarder-tenant-id tti --defaults

  Get visibility between The Things Network (NetID 000013) and Senet (000009):
    $ pbctl gateway-visibility get --forwarder-net-id 000013 \
      --home-network-net-id 000009`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				client = mappingpb.NewGatewayVisibilityManagerClient(cpConn)
				res    *mappingpb.GetGatewayVisibilityResponse
				err    error
			)
			forwarderTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			defaults, _ := cmd.Flags().GetBool("defaults")
			if defaults {
				res, err = client.GetDefaultVisibility(ctx, &mappingpb.GetDefaultGatewayVisibilityRequest{
					ForwarderNetId:    uint32(forwarderTenantID.NetID),
					ForwarderTenantId: forwarderTenantID.ID,
				})
			} else {
				homeNetworkTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "home-network")
				res, err = client.GetHomeNetworkVisibility(ctx, &mappingpb.GetHomeNetworkGatewayVisibilityRequest{
					ForwarderNetId:      uint32(forwarderTenantID.NetID),
					ForwarderTenantId:   forwarderTenantID.ID,
					HomeNetworkNetId:    uint32(homeNetworkTenantID.NetID),
					HomeNetworkTenantId: homeNetworkTenantID.ID,
				})
			}
			if err != nil {
				return err
			}
			return column.WriteVisibilities(tabout, defaults, res.Visibility)
		},
	}
	gatewayVisibilityDeleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete a visibility",
		Example: `
  Delete default gateway visibility of Forwarder network:
    $ pbctl gateway-visibility delete --forwarder-net-id 000013 --defaults

  Delete default gateway visibility of Forwarder tenant:
    $ pbctl gateway-visibility delete --forwarder-net-id 000013 \
      --forwarder-tenant-id tti --defaults

  Delete visibility between The Things Network (NetID 000013) and Senet (000009):
    $ pbctl gateway-visibility delete --forwarder-net-id 000013 \
      --home-network-net-id 000009`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := mappingpb.NewGatewayVisibilityManagerClient(cpConn)
			forwarderTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			visibility := &packetbroker.GatewayVisibility{
				ForwarderNetId:    uint32(forwarderTenantID.NetID),
				ForwarderTenantId: forwarderTenantID.ID,
			}
			var err error
			if defaults, _ := cmd.Flags().GetBool("defaults"); defaults {
				_, err = client.SetDefaultVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
					Visibility: visibility,
				})
			} else {
				homeNetworkTenantID, _ := pbflag.GetTenantID(cmd.Flags(), "home-network")
				visibility.HomeNetworkNetId = uint32(homeNetworkTenantID.NetID)
				visibility.HomeNetworkTenantId = homeNetworkTenantID.ID
				_, err = client.SetHomeNetworkVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
					Visibility: visibility,
				})
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func gatewayVisibilitySourceFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TenantID("forwarder"))
	return flags
}

func gatewayVisibilityTargetFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TenantID("home-network"))
	flags.Bool("defaults", false, "default gateway visibility")
	return flags
}

func init() {
	gatewayVisibilityCmd.PersistentFlags().AddFlagSet(gatewayVisibilitySourceFlags())
	rootCmd.AddCommand(gatewayVisibilityCmd)

	gatewayVisibilitySetCmd.Flags().AddFlagSet(gatewayVisibilityTargetFlags())
	gatewayVisibilitySetCmd.Flags().AddFlagSet(pbflag.GatewayVisibility())
	gatewayVisibilityCmd.AddCommand(gatewayVisibilitySetCmd)

	gatewayVisibilityGetCmd.Flags().AddFlagSet(gatewayVisibilityTargetFlags())
	gatewayVisibilityCmd.AddCommand(gatewayVisibilityGetCmd)

	gatewayVisibilityDeleteCmd.Flags().AddFlagSet(gatewayVisibilityTargetFlags())
	gatewayVisibilityCmd.AddCommand(gatewayVisibilityDeleteCmd)
}
