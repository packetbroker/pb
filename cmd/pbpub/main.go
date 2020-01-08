// Copyright © 2020 The Things Industries B.V.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"go.packetbroker.org/pb/cmd/internal/logging"
	"go.packetbroker.org/pb/internal/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const usage = `Usage:

  Publish as Forwarder:
  $ cat message.json | pbpub -forwarder-net-id NETID [-forwarder-id ID]

  Publish as Home Network:
  $ cat message.json | pbpub -home-network-net-id NETID -forwarder-net-id NETID [-forwarder-id ID]

Flags:`

var (
	ctx     = context.Background()
	logger  *zap.Logger
	conn    *grpc.ClientConn
	decoder *json.Decoder
)

func main() {
	parseInput()
	if input.help {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
		return
	}

	logger = logging.GetLogger(input.debug)
	defer logger.Sync()

	var err error
	conn, err = client.DialContext(ctx, logger, input.client)
	if err != nil {
		logger.Fatal("Failed to connect", zap.String("address", input.client.Address), zap.Error(err))
	}
	defer conn.Close()

	decoder = json.NewDecoder(os.Stdin)

	switch {
	case input.forwarderNetID != nil && input.homeNetworkNetID == nil:
		runForwarder()
	case input.homeNetworkNetID != nil:
		runHomeNetwork()
	}
}
