// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	packetbroker "go.packetbroker.org/api/v1unary"
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
		res, err := client.Publish(ctx, &packetbroker.PublishDownlinkMessageRequest{
			HomeNetworkNetId: uint32(*input.homeNetworkNetID),
			ForwarderNetId:   uint32(*input.forwarderNetID),
			ForwarderId:      input.forwarderID,
			Message:          msg,
		})
		if err != nil {
			logger.Error("Failed to publish downlink message", zap.Error(err))
		} else {
			logger.Info("Published downlink message", zap.String("id", res.Id))
		}
	}
}
