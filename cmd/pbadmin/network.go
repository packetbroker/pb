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

func parseNetworkFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
	case "create", "update":
		flag.StringVar(&input.network.name, "name", "", "network name")
	case "get":
	case "delete":
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.Parse(os.Args[3:])

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			input.network.hasName = true
		}
	})

	if !input.help {
		switch os.Args[2] {
		case "create", "get", "update", "delete":
			if input.netIDHex == "" {
				fmt.Fprintln(os.Stderr, "Must set net-id")
				return false
			}
		}
	}

	return true
}

func runNetwork(ctx context.Context) {
	client := iampb.NewNetworkRegistryClient(conn)
	switch os.Args[2] {
	case "list":
		pageSize := 50
		for i := 0; ; i += pageSize {
			res, err := client.ListNetworks(ctx, &iampb.ListNetworksRequest{
				Offset: uint32(i),
				Limit:  uint32(pageSize),
			})
			if err != nil {
				logger.Error("Failed to list networks", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
			for _, t := range res.Networks {
				if err = protojson.Write(os.Stdout, t); err != nil {
					logger.Error("Failed to convert network to JSON", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
			}
			if i+len(res.Networks) >= int(res.Total) {
				break
			}
		}

	case "create":
		_, err := client.CreateNetwork(ctx, &iampb.CreateNetworkRequest{
			Network: &packetbroker.Network{
				NetId: uint32(*input.netID),
				Name:  input.network.name,
				// TODO: Contact info
			},
		})
		if err != nil {
			logger.Error("Failed to create network", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "get":
		network, err := client.GetNetwork(ctx, &iampb.NetworkRequest{
			NetId: uint32(*input.netID),
		})
		if err != nil {
			logger.Error("Failed to get network", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
		if err = protojson.Write(os.Stdout, network); err != nil {
			logger.Error("Failed to convert network to JSON", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "update":
		req := &iampb.UpdateNetworkRequest{
			NetId: uint32(*input.netID),
		}
		if input.network.hasName {
			req.Name = wrapperspb.String(input.network.name)
		}
		if _, err := client.UpdateNetwork(ctx, req); err != nil {
			logger.Error("Failed to update network", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}

	case "delete":
		_, err := client.DeleteNetwork(ctx, &iampb.NetworkRequest{
			NetId: uint32(*input.netID),
		})
		if err != nil {
			logger.Error("Failed to delete network", zap.Error(err))
			clicontext.SetExitCode(ctx, 1)
			return
		}
	}
}
