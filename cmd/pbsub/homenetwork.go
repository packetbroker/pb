// Copyright © 2020 The Things Industries B.V.

package main

import (
	"io"

	packetbroker "go.packetbroker.org/api/v1alpha1"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
)

func runHomeNetwork() {
	filter := new(packetbroker.RoutingFilter)
	if input.homeNetworkFilters.forwarderNetID != nil {
		forwarder := &packetbroker.RoutingFilter_Forwarder{
			NetId: uint32(*input.homeNetworkFilters.forwarderNetID),
		}
		logger.Debug("Filter Forwarder by NetID", zap.Any("net_id", forwarder.NetId))
		if id := input.homeNetworkFilters.forwarderID; id != "" {
			forwarder.ForwarderIds = append(forwarder.ForwarderIds, id)
			logger.Debug("Filter Forwarder by ID", zap.Any("id", id))
		}
		filter.Forwarders = append(filter.Forwarders)
	}

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId: uint32(*input.homeNetworkNetID),
		Group:            input.group,
		Filters:          []*packetbroker.RoutingFilter{filter},
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
