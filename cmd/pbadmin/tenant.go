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
	"google.golang.org/protobuf/types/known/wrapperspb"
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
		flag.StringVar(&input.tenant.name, "name", "", "tenant name")
		flag.StringSliceVar(&input.devAddrBlocksHex, "dev-addr-blocks", nil, "DevAddr blocks")
	case "get":
	case "delete":
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.StringVar(&input.tenantID, "tenant-id", "", "Tenant ID")

	flag.CommandLine.Parse(os.Args[3:])

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			input.tenant.hasName = true
		case "dev-addr-blocks":
			input.hasDevAddrBlocks = true
		}
	})

	if !input.help {
		if input.netIDHex == "" {
			fmt.Fprintln(os.Stderr, "Must set net-id")
			return false
		}
		switch os.Args[2] {
		case "create", "update":
			input.devAddrBlocks = make([]*packetbroker.DevAddrBlock, len(input.devAddrBlocksHex))
			for i, s := range input.devAddrBlocksHex {
				block, err := parseDevAddrBlock(s)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Invalid DevAddr block:", err)
					return false
				}
				input.devAddrBlocks[i] = block
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
		offset := uint32(0)
		for {
			res, err := client.ListTenants(ctx, &iampb.ListTenantsRequest{
				NetId:  uint32(*input.netID),
				Offset: offset,
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
			offset += uint32(len(res.Tenants))
			if len(res.Tenants) == 0 || offset >= res.Total {
				break
			}
		}

	case "create":
		_, err := client.CreateTenant(ctx, &iampb.CreateTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:         uint32(*input.netID),
				TenantId:      input.tenantID,
				Name:          input.tenant.name,
				DevAddrBlocks: input.devAddrBlocks,
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
		req := &iampb.UpdateTenantRequest{
			NetId:    uint32(*input.netID),
			TenantId: input.tenantID,
		}
		if input.tenant.hasName {
			req.Name = wrapperspb.String(input.tenant.name)
		}
		if input.hasDevAddrBlocks {
			req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
				Value: input.devAddrBlocks,
			}
		}
		if _, err := client.UpdateTenant(ctx, req); err != nil {
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
