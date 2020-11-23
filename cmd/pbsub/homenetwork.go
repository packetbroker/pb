// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"
	"os"

	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"htdvisser.dev/exp/clicontext"
)

func runHomeNetwork(ctx context.Context) error {
	// Subscribe to all MAC payload and join-requests.
	macFilter := &packetbroker.RoutingFilter_MACPayload{}
	if len(input.homeNetworkFilters.devAddrPrefixes) > 0 {
		macFilter.DevAddrPrefixes = input.homeNetworkFilters.devAddrPrefixes
	}
	filters := []*packetbroker.RoutingFilter{
		{
			Message: &packetbroker.RoutingFilter_Mac{
				Mac: macFilter,
			},
		},
		{
			Message: &packetbroker.RoutingFilter_JoinRequest_{
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{},
			},
		},
	}

	client := routingpb.NewHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeHomeNetworkRequest{
		HomeNetworkNetId:     uint32(*input.homeNetworkNetID),
		HomeNetworkClusterId: input.homeNetworkClusterID,
		HomeNetworkTenantId:  input.homeNetworkTenantID,
		Group:                input.group,
		Filters:              filters,
	})
	if err != nil {
		logger.Error("Failed to subscribe", zap.Error(err))
		clicontext.SetExitCode(ctx, 1)
		return err
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				logger.Error("Stream failed", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
			}
			return err
		}
		if err = protojson.Write(os.Stdout, msg); err != nil {
			return err
		}
	}
}
