// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	policyCmd = &cobra.Command{
		Use:     "policy",
		Aliases: []string{"policies", "po"},
		Short:   "Manage Packet Broker routing policies",
	}
	policyListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List policies",
		Example: `
  List Home Network policies of a Forwarder network:
    $ pbctl policy list --forwarder-net-id 000013

  List Home Network policies of a Forwarder tenant:
    $ pbctl policy list --forwarder-net-id 000013 --forwarder-tenant-id tti

  List default policies of all Forwarders:
    $ pbctl policy list --defaults

  List Home Network policies of all Forwarders:
    $ pbctl policy list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				client        = routingpb.NewPolicyManagerClient(conn)
				lastUpdatedAt time.Time
			)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			defaults, _ := cmd.Flags().GetBool("defaults")
			for {
				var policies []*packetbroker.RoutingPolicy
				if defaults {
					res, err := client.ListDefaultPolicies(ctx, &routingpb.ListDefaultPoliciesRequest{
						UpdatedSince: timestamppb.New(lastUpdatedAt),
					})
					if err != nil {
						return err
					}
					policies = res.Policies
				} else {
					req := &routingpb.ListHomeNetworkPoliciesRequest{
						UpdatedSince: timestamppb.New(lastUpdatedAt),
					}
					if !forwarderTenantID.IsEmpty() {
						req.ForwarderNetId = uint32(forwarderTenantID.NetID)
						req.ForwarderTenantId = forwarderTenantID.ID
					}
					res, err := client.ListHomeNetworkPolicies(ctx, req)
					if err != nil {
						return err
					}
					policies = res.Policies
				}
				for _, p := range policies {
					if err := protojson.Write(os.Stdout, p); err != nil {
						return err
					}
					lastUpdatedAt = p.UpdatedAt.AsTime()
				}
				if len(policies) == 0 {
					break
				}
			}
			return nil
		},
	}
	policySetCmd = &cobra.Command{
		Use:   "set",
		Short: "Set a policy",
		Long: `Set default routing policy of Forwarder, or a specific policy between a
Forwarder and a Home Network.

You can specify uplink policies with the following letters:
  J: Join-request messages
  M: MAC layer uplink (FPort 0)
  A: Application layer uplink (FPort > 0)
  S: Signal quality
  L: Gateway location

You can specify downlink policies with the following letters:
  J: Join-accept messages
  M: MAC layer downlink (FPort 0)
  A: Application layer downlink (FPort > 0)

Policies are defined by the Forwarding network or tenant, as they control who
may use their infrastructure.`,
		Example: `
  Set default policy of Forwarder network to allow all peering:
    $ pbctl policy set --forwarder-net-id 000013 \
      --defaults --set-uplink JMASL --set-downlink JMA

  Set default policy of Forwarder tenant to allow all peering:
    $ pbctl policy set --forwarder-net-id 000013 --forwarder-tenant-id tti \
      --defaults --set-uplink JMASL --set-downlink JMA

  Configure only device activations and MAC-layer traffic between The Things
  Network (NetID 000013) and Senet (NetID 000009):
    $ pbctl policy set --forwarder-net-id 000013 --home-network-net-id 000009 \
      --set-uplink JM --set-downlink JM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := routingpb.NewPolicyManagerClient(conn)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			uplink, downlink := pbflag.GetRoutingPolicy(cmd.Flags())
			policy := &packetbroker.RoutingPolicy{
				ForwarderNetId:    uint32(forwarderTenantID.NetID),
				ForwarderTenantId: forwarderTenantID.ID,
				Uplink:            uplink,
				Downlink:          downlink,
			}
			var err error
			if defaults, _ := cmd.Flags().GetBool("defaults"); defaults {
				_, err = client.SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
					Policy: policy,
				})
			} else {
				homeNetworkTenantID := pbflag.GetTenantID(cmd.Flags(), "home-network")
				policy.HomeNetworkNetId = uint32(homeNetworkTenantID.NetID)
				policy.HomeNetworkTenantId = homeNetworkTenantID.ID
				_, err = client.SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
					Policy: policy,
				})
			}
			if err != nil {
				return err
			}
			return protojson.Write(os.Stdout, policy)
		},
	}
	policyGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a policy",
		Example: `
  Get default policy of Forwarder network:
    $ pbctl policy get --forwarder-net-id 000013 --defaults

  Get default policy of Forwarder tenant:
    $ pbctl policy get --forwarder-net-id 000013 --forwarder-tenant-id tti \
      --defaults

  Get policy between The Things Network (NetID 000013) and Senet (000009):
    $ pbctl policy get --forwarder-net-id 000013 --home-network-net-id 000009`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				client = routingpb.NewPolicyManagerClient(conn)
				res    *routingpb.GetPolicyResponse
				err    error
			)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			if defaults, _ := cmd.Flags().GetBool("defaults"); defaults {
				res, err = client.GetDefaultPolicy(ctx, &routingpb.GetDefaultPolicyRequest{
					ForwarderNetId:    uint32(forwarderTenantID.NetID),
					ForwarderTenantId: forwarderTenantID.ID,
				})
			} else {
				homeNetworkTenantID := pbflag.GetTenantID(cmd.Flags(), "home-network")
				res, err = client.GetHomeNetworkPolicy(ctx, &routingpb.GetHomeNetworkPolicyRequest{
					ForwarderNetId:      uint32(forwarderTenantID.NetID),
					ForwarderTenantId:   forwarderTenantID.ID,
					HomeNetworkNetId:    uint32(homeNetworkTenantID.NetID),
					HomeNetworkTenantId: homeNetworkTenantID.ID,
				})
			}
			if err != nil {
				return err
			}
			return protojson.Write(os.Stdout, res.Policy)
		},
	}
	policyDeleteCmd = &cobra.Command{
		Use:          "delete",
		Aliases:      []string{"rm"},
		Short:        "Delete a policy",
		SilenceUsage: true,
		Example: `
  Delete default policy of Forwarder network:
    $ pbctl policy delete --forwarder-net-id 000013 --defaults

  Delete default policy of Forwarder tenant:
    $ pbctl policy delete --forwarder-net-id 000013 --forwarder-tenant-id tti \
      --defaults

  Delete policy between The Things Network (NetID 000013) and Senet (000009):
    $ pbctl policy delete --forwarder-net-id 000013 --home-network-net-id 000009`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := routingpb.NewPolicyManagerClient(conn)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			policy := &packetbroker.RoutingPolicy{
				ForwarderNetId:    uint32(forwarderTenantID.NetID),
				ForwarderTenantId: forwarderTenantID.ID,
			}
			var err error
			if defaults, _ := cmd.Flags().GetBool("defaults"); defaults {
				_, err = client.SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
					Policy: policy,
				})
			} else {
				homeNetworkTenantID := pbflag.GetTenantID(cmd.Flags(), "home-network")
				policy.HomeNetworkNetId = uint32(homeNetworkTenantID.NetID)
				policy.HomeNetworkTenantId = homeNetworkTenantID.ID
				_, err = client.SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
					Policy: policy,
				})
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func policySourceFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TenantID("forwarder"))
	return flags
}

func policyTargetFlags() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(pbflag.TenantID("home-network"))
	flags.Bool("defaults", false, "default policy")
	return flags
}

func init() {
	policyCmd.PersistentFlags().AddFlagSet(policySourceFlags())
	rootCmd.AddCommand(policyCmd)

	policyListCmd.Flags().Bool("defaults", false, "list all default policies")
	policyCmd.AddCommand(policyListCmd)

	policySetCmd.Flags().AddFlagSet(policyTargetFlags())
	policySetCmd.Flags().AddFlagSet(pbflag.RoutingPolicy())
	policyCmd.AddCommand(policySetCmd)

	policyGetCmd.Flags().AddFlagSet(policyTargetFlags())
	policyCmd.AddCommand(policyGetCmd)

	policyDeleteCmd.Flags().AddFlagSet(policyTargetFlags())
	policyCmd.AddCommand(policyDeleteCmd)
}
