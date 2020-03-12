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

func runHomeNetwork(ctx context.Context) error {
	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg := new(packetbroker.DownlinkMessage)
		if err := jsonpb.UnmarshalNext(decoder, msg); err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				logger.Error("Failed to decode downlink message", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
			}
			return err
		}
		stream, err := client.Publish(ctx, &packetbroker.PublishDownlinkMessageRequest{
			HomeNetworkNetId: uint32(*input.homeNetworkNetID),
			ForwarderNetId:   uint32(*input.forwarderNetID),
			ForwarderId:      input.forwarderID,
			Message:          msg,
		})
		if err != nil {
			logger.Error("Failed to publish downlink message", zap.Error(err))
			continue
		}
		for {
			res, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
					logger.Error("Failed to receive publish downlink message progress", zap.Error(err))
				}
				break
			}
			logger.Info("Publish downlink message state change",
				zap.String("id", res.Id),
				zap.String("state", res.State.String()),
			)
		}
	}
}
