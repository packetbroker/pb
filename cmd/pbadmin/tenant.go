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

func parseTenantFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
	case "get":
	case "set":
		flag.CommandLine.StringSliceVar(&input.tenant.devAddrPrefixes, "dev-addr-prefixes", nil, "DevAddr prefixes")
	case "delete":
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.Parse(os.Args[3:])

	if input.netIDHex == "" {
		fmt.Fprintln(os.Stderr, "Must set net-id")
		return false
	}

	return true
}

func runTenant(ctx context.Context) {
	client := packetbroker.NewTenantManagerClient(conn)
	switch os.Args[2] {
	case "list":
		res, err := client.ListTenants(ctx, &packetbroker.ListTenantsRequest{
			NetId: uint32(*input.netID),
		})
		if err != nil {
			logger.Error("Failed to list tenants", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		console.WriteProto(res)
	case "get":
		tenant, err := client.GetTenant(ctx, &packetbroker.GetTenantRequest{
			NetId:    uint32(*input.netID),
			TenantId: input.tenantID,
		})
		if err != nil {
			logger.Error("Failed to get tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		console.WriteProto(tenant)
	case "set":
		prefixes := make([]*packetbroker.DevAddrPrefix, len(input.tenant.devAddrPrefixes))
		for i, s := range input.tenant.devAddrPrefixes {
			prefix, err := parseDevAddrPrefix(s)
			if err != nil {
				logger.Error("Failed to parse DevAddr prefix", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
			prefixes[i] = prefix
		}
		_, err := client.SetTenant(ctx, &packetbroker.SetTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:           uint32(*input.netID),
				TenantId:        input.tenantID,
				DevAddrPrefixes: prefixes,
			},
		})
		if err != nil {
			logger.Error("Failed to set tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
	case "delete":
		_, err := client.DeleteTenant(ctx, &packetbroker.DeleteTenantRequest{
			NetId:    uint32(*input.netID),
			TenantId: input.tenantID,
		})
		if err != nil {
			logger.Error("Failed to delete tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
	}
}
