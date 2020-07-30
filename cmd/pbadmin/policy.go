// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"htdvisser.dev/exp/clicontext"
)

func parsePolicyFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
	case "set":
		flag.StringVar(&input.policy.setUplink, "set-uplink", "", "concatenated J (join-request), M (MAC data), A (application data), S (signal quality), L (localization)")
		flag.StringVar(&input.policy.setDownlink, "set-downlink", "", "concatenated J (join-accept), M (MAC data), A (application data)")
		flag.BoolVar(&input.policy.unset, "unset", false, "unset policy")
		fallthrough
	case "get":
		flag.BoolVar(&input.policy.defaults, "defaults", false, "default policy")
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.Parse(os.Args[3:])

	if input.forwarderNetIDHex == "" {
		fmt.Fprintln(os.Stderr, "Must set forwarder-net-id")
		return false
	}
	switch os.Args[2] {
	case "set":
		if (input.policy.setUplink != "" || input.policy.setDownlink != "") == input.policy.unset {
			fmt.Fprintln(os.Stderr, "Must set or unset policies")
			return false
		}
		fallthrough
	case "get":
		if input.policy.defaults == (input.homeNetworkNetIDHex != "") {
			fmt.Fprintln(os.Stderr, "Must set either home-network-net-id or defaults")
			return false
		}
	}

	return true
}

func parseUplinkPolicy() *packetbroker.RoutingPolicy_Uplink {
	var res packetbroker.RoutingPolicy_Uplink
	for _, l := range input.policy.setUplink {
		switch l {
		case 'J':
			res.JoinRequest = true
		case 'M':
			res.MacData = true
		case 'A':
			res.ApplicationData = true
		case 'S':
			res.SignalQuality = true
		case 'L':
			res.Localization = true
		default:
			logger.Warn("Invalid uplink policy", zap.String("designator", string(l)))
		}
	}
	return &res
}

func parseDownlinkPolicy() *packetbroker.RoutingPolicy_Downlink {
	var res packetbroker.RoutingPolicy_Downlink
	for _, l := range input.policy.setDownlink {
		switch l {
		case 'J':
			res.JoinAccept = true
		case 'M':
			res.MacData = true
		case 'A':
			res.ApplicationData = true
		default:
			logger.Warn("Invalid downlink policy", zap.String("designator", string(l)))
		}
	}
	return &res
}

func runPolicy(ctx context.Context) {
	client := packetbroker.NewRoutingPolicyManagerClient(conn)
	switch os.Args[2] {
	case "list":
		pageSize := 50
		for i := 0; ; i += pageSize {
			res, err := client.ListHomeNetworkPolicies(ctx, &packetbroker.ListHomeNetworkRoutingPoliciesRequest{
				ForwarderNetId:    uint32(*input.forwarderNetID),
				ForwarderTenantId: input.forwarderTenantID,
				Offset:            uint32(i),
				Limit:             uint32(pageSize),
			})
			if err != nil {
				logger.Error("Failed to list home network policies", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
			for _, p := range res.Policies {
				if err = protojson.Write(os.Stdout, p); err != nil {
					logger.Error("Failed to convert policy to JSON", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
			}
			if i+len(res.Policies) >= int(res.Total) {
				break
			}
		}

	case "set":
		var (
			policy = &packetbroker.RoutingPolicy{
				ForwarderNetId:    uint32(*input.forwarderNetID),
				ForwarderTenantId: input.forwarderTenantID,
				Uplink:            parseUplinkPolicy(),
				Downlink:          parseDownlinkPolicy(),
			}
			err error
		)
		if input.policy.defaults {
			_, err = client.SetDefaultPolicy(ctx, &packetbroker.SetRoutingPolicyRequest{
				Policy: policy,
			})
		} else {
			policy.HomeNetworkNetId = uint32(*input.homeNetworkNetID)
			policy.HomeNetworkTenantId = input.homeNetworkTenantID
			_, err = client.SetHomeNetworkPolicy(ctx, &packetbroker.SetRoutingPolicyRequest{
				Policy: policy,
			})
		}
		if err != nil {
			logger.Error("Failed to set routing policy", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		if err = protojson.Write(os.Stdout, policy); err != nil {
			logger.Error("Failed to convert policy to JSON", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "get":
		var (
			res *packetbroker.GetRoutingPolicyResponse
			err error
		)
		if input.policy.defaults {
			res, err = client.GetDefaultPolicy(ctx, &packetbroker.GetDefaultRoutingPolicyRequest{
				ForwarderNetId:    uint32(*input.forwarderNetID),
				ForwarderTenantId: input.forwarderTenantID,
			})
		} else {
			res, err = client.GetHomeNetworkPolicy(ctx, &packetbroker.GetHomeNetworkRoutingPolicyRequest{
				ForwarderNetId:    uint32(*input.forwarderNetID),
				ForwarderTenantId: input.forwarderTenantID,
				HomeNetworkNetId:  uint32(*input.homeNetworkNetID),
			})
		}
		if err != nil {
			logger.Error("Failed to get routing policy", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		protojson.Write(os.Stdout, res.Policy)
	}
}
