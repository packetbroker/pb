// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v2beta1"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
	"htdvisser.dev/exp/clicontext"
)

func parsePolicyFlags() bool {
	flag.BoolVar(&input.policy.defaults, "defaults", false, "default policy")
	flag.StringVar(&input.policy.setUplink, "set-uplink", "", "concatenated J (join-request), M (MAC data), A (application data), S (signal quality), L (localization), D (allow downlink)")
	flag.BoolVar(&input.policy.unsetUplink, "unset-uplink", false, "unset uplink policy")
	flag.StringVar(&input.policy.setDownlink, "set-downlink", "", "concatenated J (join-accept), M (MAC data), A (application data)")
	flag.BoolVar(&input.policy.unsetDownlink, "unset-downlink", false, "unset downlink policy")

	flag.CommandLine.Parse(os.Args[2:])

	if input.forwarderNetIDHex == "" {
		fmt.Fprintln(os.Stderr, "Must set forwarder-net-id")
		return false
	}
	if input.policy.defaults == (input.homeNetworkNetIDHex != "") {
		fmt.Fprintln(os.Stderr, "Must set either home-network-net-id or defaults")
		return false
	}
	if input.policy.setUplink != "" && input.policy.unsetUplink || input.policy.setDownlink != "" && input.policy.unsetDownlink {
		fmt.Fprintln(os.Stderr, "Cannot set and unset policies")
		return false
	}

	return true
}

func parseUplinkPolicy() *packetbroker.RoutingPolicy_Uplink {
	if input.policy.setUplink == "" {
		return nil
	}
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
		case 'D':
			res.AllowDownlink = true
		default:
			logger.Warn("Invalid uplink policy", zap.String("designator", string(l)))
		}
	}
	return &res
}

func parseDownlinkPolicy() *packetbroker.RoutingPolicy_Downlink {
	if input.policy.setDownlink == "" {
		return nil
	}
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
	if input.policy.setUplink != "" || input.policy.unsetUplink ||
		input.policy.setDownlink != "" || input.policy.unsetDownlink {
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
			_, err = client.SetHomeNetworkPolicy(ctx, &packetbroker.SetRoutingPolicyRequest{
				Policy: policy,
			})
		}
		if err != nil {
			logger.Error("Failed to set routing policy", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		console.WriteProto(policy)
	} else {
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
		console.WriteProto(res.Policy)
	}
}
