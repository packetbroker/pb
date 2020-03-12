// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"

	packetbroker "go.packetbroker.org/api/v2beta1"
	"go.packetbroker.org/pb/cmd/internal/console"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"htdvisser.dev/exp/clicontext"
)

func runForwarder(ctx context.Context) error {
	client := packetbroker.NewRouterForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeForwarderRequest{
		ForwarderNetId: uint32(*input.forwarderNetID),
		ForwarderId:    input.forwarderID,
		Group:          input.group,
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
