// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"errors"
	"io"

	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.packetbroker.org/pb/cmd/internal/protojson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"htdvisser.dev/exp/clicontext"
)

func runHomeNetwork(ctx context.Context) error {
	client := routingpb.NewHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg := new(packetbroker.DownlinkMessage)
		if err := protojson.Decode(decoder, msg); err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				logger.Error("Failed to decode downlink message", zap.Error(err))
				clicontext.SetExitCode(ctx, 1)
			}
			return err
		}
		res, err := client.Publish(ctx, &routingpb.PublishDownlinkMessageRequest{
			HomeNetworkNetId:     uint32(*input.homeNetworkNetID),
			HomeNetworkClusterId: input.homeNetworkClusterID,
			HomeNetworkTenantId:  input.homeNetworkTenantID,
			ForwarderNetId:       uint32(*input.forwarderNetID),
			ForwarderClusterId:   input.forwarderClusterID,
			ForwarderTenantId:    input.forwarderTenantID,
			Message:              msg,
		})
		if err != nil {
			logger.Error("Failed to publish downlink message", zap.Error(err))
			continue
		}
		logger.Info("Published downlink message", zap.String("id", res.Id))
	}
}
