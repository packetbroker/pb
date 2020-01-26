// Copyright © 2020 The Things Industries B.V.

package main

import (
	"io"

	"github.com/gogo/protobuf/jsonpb"
	packetbroker "go.packetbroker.org/api/v1beta3"
	"go.uber.org/zap"
)

func runForwarder() {
	client := packetbroker.NewRouterForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg := new(packetbroker.UplinkMessage)
		if err := jsonpb.UnmarshalNext(decoder, msg); err != nil {
			if err == io.EOF {
				return
			}
			logger.Fatal("Failed to decode uplink message", zap.Error(err))
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
				if err != io.EOF {
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
