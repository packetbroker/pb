// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	packetbroker "go.packetbroker.org/api/v2beta1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"htdvisser.dev/exp/clicontext"
)

func runForwarder(ctx context.Context) error {
	client := packetbroker.NewRouterForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg := new(packetbroker.UplinkMessage)
		if err := jsonpb.UnmarshalNext(decoder, msg); err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				logger.Error("Failed to decode uplink message", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
			}
			return err
		}
		stream, err := client.Publish(ctx, &packetbroker.PublishUplinkMessageRequest{
			ForwarderNetId: uint32(*input.forwarderNetID),
			ForwarderId:    input.forwarderID,
			Message:        msg,
		})
		if err != nil {
			logger.Error("Failed to publish uplink message", zap.Error(err))
			continue
		}
		for {
			res, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
					logger.Error("Failed to receive publish uplink message progress", zap.Error(err))
				}
				break
			}
			logger.Info("Publish uplink message state change",
				zap.String("id", res.Id),
				zap.String("state", res.State.String()),
			)
		}
	}
}
