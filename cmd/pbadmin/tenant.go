// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	iampb "go.packetbroker.org/api/iam"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/protojson"
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
	case "create", "update":
		flag.CommandLine.StringVar(&input.tenant.name, "name", "", "tenant name")
		flag.CommandLine.StringSliceVar(&input.tenant.devAddrPrefixesHex, "dev-addr-prefixes", nil, "DevAddr prefixes")
	case "get":
	case "delete":
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.Parse(os.Args[3:])

	if !input.help {
		if input.netIDHex == "" {
			fmt.Fprintln(os.Stderr, "Must set net-id")
			return false
		}
		switch os.Args[2] {
		case "create", "update":
			input.tenant.devAddrPrefixes = make([]*packetbroker.DevAddrPrefix, len(input.tenant.devAddrPrefixesHex))
			for i, s := range input.tenant.devAddrPrefixesHex {
				prefix, err := parseDevAddrPrefixClusterID(s)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Invalid DevAddr prefixes:", err)
					return false
				}
				input.tenant.devAddrPrefixes[i] = prefix
			}
			fallthrough
		case "get", "delete":
			if input.tenantID == "" {
				fmt.Fprintln(os.Stderr, "Must set tenant-id")
				return false
			}
		}
	}

	return true
}

func runTenant(ctx context.Context) {
	client := iampb.NewTenantRegistryClient(conn)
	switch os.Args[2] {
	case "list":
		pageSize := 50
		for i := 0; ; i += pageSize {
			res, err := client.ListTenants(ctx, &iampb.ListTenantsRequest{
				NetId:  uint32(*input.netID),
				Offset: uint32(i),
				Limit:  uint32(pageSize),
			})
			if err != nil {
				logger.Error("Failed to list tenants", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
			for _, t := range res.Tenants {
				if err = protojson.Write(os.Stdout, t); err != nil {
					logger.Error("Failed to convert tenant to JSON", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
			}
			if i+len(res.Tenants) >= int(res.Total) {
				break
			}
		}

	case "create":
		_, err := client.CreateTenant(ctx, &iampb.CreateTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:           uint32(*input.netID),
				TenantId:        input.tenantID,
				Name:            input.tenant.name,
				DevAddrPrefixes: input.tenant.devAddrPrefixes,
				// TODO: Contact info
			},
		})
		if err != nil {
			logger.Error("Failed to create tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "get":
		tenant, err := client.GetTenant(ctx, &iampb.TenantRequest{
			NetId:    uint32(*input.netID),
			TenantId: input.tenantID,
		})
		if err != nil {
			logger.Error("Failed to get tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		if err = protojson.Write(os.Stdout, tenant); err != nil {
			logger.Error("Failed to convert tenant to JSON", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "update":
		_, err := client.UpdateTenant(ctx, &iampb.UpdateTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:           uint32(*input.netID),
				TenantId:        input.tenantID,
				Name:            input.tenant.name,
				DevAddrPrefixes: input.tenant.devAddrPrefixes,
				// TODO: Contact info
			},
		})
		if err != nil {
			logger.Error("Failed to update tenant", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "delete":
		_, err := client.DeleteTenant(ctx, &iampb.TenantRequest{
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
