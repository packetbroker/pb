// Copyright © 2020 The Things Industries B.V.

package main

import (
	"io"

	packetbroker "go.packetbroker.org/api/v1beta3"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
)

func runHomeNetwork() {
	var forwardersFilter *packetbroker.RoutingFilter_ForwarderWhitelist
	if input.homeNetworkFilters.forwarderNetID != nil {
		forwarder := &packetbroker.ForwarderIdentifier{
			NetId: uint32(*input.homeNetworkFilters.forwarderNetID),
		}
		logger := logger
		if id := input.homeNetworkFilters.forwarderID; id != "" {
			forwarder.ForwarderId = id
			logger = logger.With(zap.Any("id", id))
		}
		forwardersFilter = &packetbroker.RoutingFilter_ForwarderWhitelist{
			ForwarderWhitelist: &packetbroker.ForwarderIdentifiers{
				List: []*packetbroker.ForwarderIdentifier{forwarder},
			},
		}
		logger.Debug("Filter Forwarder", zap.Any("net_id", forwarder.NetId))
	}

	// Subscribe to all MAC payload and join-requests.
	filters := []*packetbroker.RoutingFilter{
		{
			Forwarders: forwardersFilter,
			Message: &packetbroker.RoutingFilter_Mac{
				Mac: &packetbroker.RoutingFilter_MACPayload{},
			},
		},
		{
			Forwarders: forwardersFilter,
			Message: &packetbroker.RoutingFilter_JoinRequest_{
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{},
			},
		},
	}

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId: uint32(*input.homeNetworkNetID),
		Group:            input.group,
		Filters:          filters,
	})
	if err != nil {
		logger.Fatal("Failed to subscribe", zap.Error(err))
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			logger.Fatal("Stream failed", zap.Error(err))
		}
		console.WriteProto(msg)
	}
}
