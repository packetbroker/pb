// Copyright © 2020 The Things Industries B.V.

package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/column"
	pbflag "go.packetbroker.org/pb/cmd/internal/pbflag"
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
    $ pbctl policy list

  List effective Home Network policies for a Home Network:
    $ pbctl policy list --home-network-net-id 000013

  List effective Home Network policies for a Home Network tenant:
    $ pbctl policy list --home-network-net-id 000013 \
      --home-network-tenant-id tti
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				client              = routingpb.NewPolicyManagerClient(cpConn)
				policies            []*packetbroker.RoutingPolicy
				defaults            bool
				homeNetworkTenantID = pbflag.GetTenantID(cmd.Flags(), "home-network")
			)
			if homeNetworkTenantID.IsEmpty() {
				var (
					lastUpdatedAt     *timestamppb.Timestamp
					forwarderTenantID = pbflag.GetTenantID(cmd.Flags(), "forwarder")
				)
				defaults, _ = cmd.Flags().GetBool("defaults")
				for {
					var page []*packetbroker.RoutingPolicy
					if defaults {
						res, err := client.ListDefaultPolicies(ctx, &routingpb.ListDefaultPoliciesRequest{
							UpdatedSince: lastUpdatedAt,
						})
						if err != nil {
							return err
						}
						page = res.Policies
					} else {
						req := &routingpb.ListHomeNetworkPoliciesRequest{
							UpdatedSince: lastUpdatedAt,
						}
						if !forwarderTenantID.IsEmpty() {
							req.ForwarderNetId = uint32(forwarderTenantID.NetID)
							req.ForwarderTenantId = forwarderTenantID.ID
						}
						res, err := client.ListHomeNetworkPolicies(ctx, req)
						if err != nil {
							return err
						}
						page = res.Policies
					}
					if len(page) == 0 {
						break
					}
					policies = append(policies, page...)
					lastUpdatedAt = page[len(page)-1].GetUpdatedAt()
				}
			} else {
				offset := uint32(0)
				for {
					res, err := client.ListEffectivePolicies(ctx, &routingpb.ListEffectivePoliciesRequest{
						HomeNetworkNetId:    uint32(homeNetworkTenantID.NetID),
						HomeNetworkTenantId: homeNetworkTenantID.ID,
						Offset:              offset,
					})
					if err != nil {
						return err
					}
					policies = append(policies, res.Policies...)
					offset += uint32(len(res.Policies))
					if len(res.Policies) == 0 || offset >= res.Total {
						break
					}
				}
			}
			column.WritePolicies(tabout, defaults, policies...)
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
			client := routingpb.NewPolicyManagerClient(cpConn)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			uplink, downlink := pbflag.GetRoutingPolicy(cmd.Flags())
			policy := &packetbroker.RoutingPolicy{
				ForwarderNetId:    uint32(forwarderTenantID.NetID),
				ForwarderTenantId: forwarderTenantID.ID,
				Uplink:            uplink,
				Downlink:          downlink,
			}
			defaults, _ := cmd.Flags().GetBool("defaults")
			var err error
			if defaults {
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
			return column.WritePolicies(tabout, defaults, policy)
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
				client = routingpb.NewPolicyManagerClient(cpConn)
				res    *routingpb.GetPolicyResponse
				err    error
			)
			forwarderTenantID := pbflag.GetTenantID(cmd.Flags(), "forwarder")
			defaults, _ := cmd.Flags().GetBool("defaults")
			if defaults {
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
			return column.WritePolicies(tabout, defaults, res.Policy)
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
			client := routingpb.NewPolicyManagerClient(cpConn)
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

	policyListCmd.Flags().AddFlagSet(policyTargetFlags())
	policyCmd.AddCommand(policyListCmd)

	policySetCmd.Flags().AddFlagSet(policyTargetFlags())
	policySetCmd.Flags().AddFlagSet(pbflag.RoutingPolicy())
	policyCmd.AddCommand(policySetCmd)

	policyGetCmd.Flags().AddFlagSet(policyTargetFlags())
	policyCmd.AddCommand(policyGetCmd)

	policyDeleteCmd.Flags().AddFlagSet(policyTargetFlags())
	policyCmd.AddCommand(policyDeleteCmd)
}
