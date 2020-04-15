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
		res, err := client.Publish(ctx, &packetbroker.PublishUplinkMessageRequest{
			ForwarderNetId: uint32(*input.forwarderNetID),
			ForwarderId:    input.forwarderID,
			Message:        msg,
		})
		if err != nil {
			logger.Error("Failed to publish uplink message", zap.Error(err))
		} else {
			logger.Info("Published uplink message", zap.String("id", res.Id))
		}
	}
}
