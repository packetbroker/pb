// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"htdvisser.dev/exp/clicontext"
)

func parsePolicyFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
		flag.BoolVar(&input.policy.defaults, "defaults", false, "default policies")
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

	flag.StringVar(&input.forwarderNetIDHex, "forwarder-net-id", "", "NetID of the Forwarder (hex)")
	flag.StringVar(&input.forwarderTenantID, "forwarder-tenant-id", "", "Tenant ID of the Forwarder")
	flag.StringVar(&input.homeNetworkNetIDHex, "home-network-net-id", "", "NetID of the Home Network (hex)")
	flag.StringVar(&input.homeNetworkTenantID, "home-network-tenant-id", "", "Tenant ID of the Home Network")

	flag.CommandLine.Parse(os.Args[3:])

	if !input.help {
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
		if input.forwarderNetIDHex != "" {
			input.forwarderNetID = new(packetbroker.NetID)
			if err := input.forwarderNetID.UnmarshalText([]byte(input.forwarderNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid forwarder-net-id:", err)
				return false
			}
		}
		if input.homeNetworkNetIDHex != "" {
			input.homeNetworkNetID = new(packetbroker.NetID)
			if err := input.homeNetworkNetID.UnmarshalText([]byte(input.homeNetworkNetIDHex)); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid home-network-net-id:", err)
				return false
			}
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
	client := routingpb.NewPolicyManagerClient(conn)
	switch os.Args[2] {
	case "list":
		var lastUpdatedAt time.Time
		for {
			var policies []*packetbroker.RoutingPolicy
			if input.policy.defaults {
				res, err := client.ListDefaultPolicies(ctx, &routingpb.ListDefaultPoliciesRequest{
					UpdatedSince: timestamppb.New(lastUpdatedAt),
				})
				if err != nil {
					logger.Error("Failed to list default policies", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
				policies = res.Policies
			} else {
				req := &routingpb.ListHomeNetworkPoliciesRequest{
					UpdatedSince: timestamppb.New(lastUpdatedAt),
				}
				if input.forwarderNetID != nil {
					req.ForwarderNetId = uint32(*input.forwarderNetID)
					req.ForwarderTenantId = input.forwarderTenantID
				}
				res, err := client.ListHomeNetworkPolicies(ctx, req)
				if err != nil {
					logger.Error("Failed to list Home Network policies", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
				policies = res.Policies
			}
			for _, p := range policies {
				if err := protojson.Write(os.Stdout, p); err != nil {
					logger.Error("Failed to convert policy to JSON", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
				lastUpdatedAt = p.UpdatedAt.AsTime()
			}
			if len(policies) == 0 {
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
			_, err = client.SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
				Policy: policy,
			})
		} else {
			policy.HomeNetworkNetId = uint32(*input.homeNetworkNetID)
			policy.HomeNetworkTenantId = input.homeNetworkTenantID
			_, err = client.SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
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
			res *routingpb.GetPolicyResponse
			err error
		)
		if input.policy.defaults {
			res, err = client.GetDefaultPolicy(ctx, &routingpb.GetDefaultPolicyRequest{
				ForwarderNetId:    uint32(*input.forwarderNetID),
				ForwarderTenantId: input.forwarderTenantID,
			})
		} else {
			res, err = client.GetHomeNetworkPolicy(ctx, &routingpb.GetHomeNetworkPolicyRequest{
				ForwarderNetId:      uint32(*input.forwarderNetID),
				ForwarderTenantId:   input.forwarderTenantID,
				HomeNetworkNetId:    uint32(*input.homeNetworkNetID),
				HomeNetworkTenantId: input.homeNetworkTenantID,
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
