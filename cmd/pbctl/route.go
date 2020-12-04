// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	routingpb "go.packetbroker.org/api/routing"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"htdvisser.dev/exp/clicontext"
)

func parseRouteFlags() bool {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Missing command")
		return false
	}
	switch os.Args[2] {
	case "list":
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		return false
	}

	flag.CommandLine.Parse(os.Args[3:])

	return true
}

func runRoute(ctx context.Context) {
	client := routingpb.NewRoutesClient(conn)
	switch os.Args[2] {
	case "list":
		var lastCreatedAt time.Time
		for {
			res, err := client.ListRoutes(ctx, &routingpb.ListRoutesRequest{
				CreatedSince: timestamppb.New(lastCreatedAt),
			})
			if err != nil {
				logger.Error("Failed to list routes", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
				return
			}
			for _, p := range res.Routes {
				if err := protojson.Write(os.Stdout, p); err != nil {
					logger.Error("Failed to convert route to JSON", zap.Error(err))
					clicontext.SetExitCode(ctx, 1)
					return
				}
				lastCreatedAt = p.CreatedAt.AsTime()
			}
			if len(res.Routes) == 0 {
				break
			}
		}
	}
}
