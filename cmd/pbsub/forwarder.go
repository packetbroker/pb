// Copyright © 2020 The Things Industries B.V.

package main

import (
	"io"

	packetbroker "go.packetbroker.org/api/v1"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
)

func runForwarder() {
	client := packetbroker.NewRouterForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeForwarderRequest{
		ForwarderNetId: uint32(*input.forwarderNetID),
		ForwarderId:    input.forwarderID,
		Group:          input.group,
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
