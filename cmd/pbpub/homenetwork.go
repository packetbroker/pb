// Copyright © 2020 The Things Industries B.V.

package main

import (
	"io"

	"github.com/gogo/protobuf/jsonpb"
	packetbroker "go.packetbroker.org/api/v1beta3"
	"go.uber.org/zap"
)

func runHomeNetwork() {
	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg := new(packetbroker.DownlinkMessage)
		if err := jsonpb.UnmarshalNext(decoder, msg); err != nil {
			if err == io.EOF {
				return
			}
			logger.Fatal("Failed to decode downlink message", zap.Error(err))
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
				if err != io.EOF {
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
