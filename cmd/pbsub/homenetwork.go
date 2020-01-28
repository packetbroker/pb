// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"

	packetbroker "go.packetbroker.org/api/v1"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"htdvisser.dev/exp/clicontext"
)

func runHomeNetwork(ctx context.Context) error {
	// Subscribe to all MAC payload and join-requests.
	filters := []*packetbroker.RoutingFilter{
		{
			Message: &packetbroker.RoutingFilter_Mac{
				Mac: &packetbroker.RoutingFilter_MACPayload{},
			},
		},
		{
			Message: &packetbroker.RoutingFilter_JoinRequest_{
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{},
			},
		},
	}

	if input.homeNetworkFilters.forwarderNetID != nil {
		forwarder := &packetbroker.ForwarderIdentifier{
			NetId: uint32(*input.homeNetworkFilters.forwarderNetID),
		}
		logger := logger
		if id := input.homeNetworkFilters.forwarderID; id != "" {
			forwarder.ForwarderId = id
			logger = logger.With(zap.Any("id", id))
		}
		logger.Debug("Filter Forwarder", zap.Any("net_id", forwarder.NetId))
		for _, f := range filters {
			f.Forwarders = &packetbroker.RoutingFilter_ForwarderWhitelist{
				ForwarderWhitelist: &packetbroker.ForwarderIdentifiers{
					List: []*packetbroker.ForwarderIdentifier{forwarder},
				},
			}
		}
	}

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId: uint32(*input.homeNetworkNetID),
		Group:            input.group,
		Filters:          filters,
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
		console.WriteProto(msg)
	}
}
